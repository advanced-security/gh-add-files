package cmd

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/cli/go-gh/v2/pkg/api"

	"github.com/spf13/cobra"
)

func init() {
	codescanningrepoCmd.PersistentFlags().StringVarP(&Organization, "organization", "o", "", "specify Organisation to implement code scanning")
	codescanningrepoCmd.MarkPersistentFlagRequired("organization")
	codescanningrepoCmd.PersistentFlags().StringVarP(&WorkflowFile, "workflow-file", "f", "", "specify the path to the code scanning workflow file")
	codescanningrepoCmd.MarkPersistentFlagRequired("workflow-file")
	codescanningrepoCmd.PersistentFlags().StringVarP(&LogFile, "log-file", "l", "", "specify the path where the log file will be saved")
	codescanningrepoCmd.MarkPersistentFlagRequired("log-file")
	codescanningrepoCmd.PersistentFlags().StringVarP(&CsvFile, "csv-file", "c", "", "specify the location of csv file")
	codescanningrepoCmd.MarkPersistentFlagRequired("csv-file")
}

var codescanningrepoCmd = &cobra.Command{
	Use:   "code-scanning-enable-repo",
	Short: "Add workflow files to enable code scanning",
	Long:  "Creates branch `code-scanning-automated` on each csv specified repo in organisation and checks in workflow file defined in `--workflow` flag",
	Run: func(cmd *cobra.Command, args []string) {

		//set up logging
		logFile, err := os.OpenFile(LogFile, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
		if err != nil {
			panic(err)
		}
		mw := io.MultiWriter(os.Stdout, logFile)
		log.SetOutput(mw)
		defer logFile.Close()

		log.Println("Set up REST API Client for GitHub interactions")
		client, err := api.DefaultRESTClient()
		if err != nil {
			log.Fatalln(err)
		}

		//open csv file
		csvFile, err := os.OpenFile(CsvFile, os.O_RDONLY, 0666)
		if err != nil {
			panic(err)
		}

		defer csvFile.Close()
		csvr := csv.NewReader(csvFile)
		var repositories []string

		for {
			row, err := csvr.Read()
			if err != nil {
				if err == io.EOF {
					break
				}
				log.Fatalln(err)

			}
			repositories = append(repositories, fmt.Sprint(row[0]))
		}

		for _, repository := range repositories {

			//get repo
			requestPath := fmt.Sprintf("repos/%s", repository)

			response, err := client.Request(http.MethodGet, requestPath, nil)
			if err != nil {
				log.Println(err)
				Errors[repository] = err
				continue
			}

			var repo Repository
			body, err := io.ReadAll(response.Body)
			if err := json.Unmarshal(body, &repo); err != nil {
				log.Println("Cannot unmarshal JSON")
				Errors[repository] = err
				continue
			}

			log.Printf("Details for Repository: Full Name: %s; Name: %s; Default Branch: %s\n", repo.FullName, repo.Name, repo.DefaultBranch)
			//check that repo has at least one codeql supported language
			coverage, err := repo.GetCodeqlLanguages(client)
			if err != nil {
				log.Println(err)
				continue
			}
			if len(coverage) <= 0 {
				log.Printf("No CodeQL supported language found for repository: %s", repo.FullName)
				continue
			}
			//check that codeql workflow file doesn't already exist
			isCodeQLEnabled, err := repo.doesCodeqlWorkflowExist(client)
			if err != nil {
				log.Println(err)
				Errors[repo.FullName] = err
				continue
			}

			if isCodeQLEnabled == true {
				log.Printf("CodeQL workflow file already exists for this repository: %s, skipping enablement.\n", repo.FullName)
				continue
			}
			newbranchref, err := repo.createBranchForRepo(client)
			if err != nil {
				log.Panicln(err)
				continue
			}
			if len(newbranchref) <= 0 {
				log.Println("Unable to create new branch")
				Errors[repo.FullName] = errors.New("Something went wrong when creating new branch")
			}

			log.Printf("Ref created succesfully at : %s\n", newbranchref)

			createdFile, err := repo.createWorkflowFile(client)
			if err != nil {
				log.Panicln(err)
				continue
			}
			if len(createdFile) <= 0 {
				log.Println("Unable to create commit new file")
				Errors[repo.FullName] = errors.New("Something went wrong when creating new file")
				continue
			}

			log.Printf("Successfully created file %s on branch %s in repository %s\n", createdFile, newbranchref, repo.FullName)
			createdPR, err := repo.raisePullRequest()
			if len(createdFile) <= 0 {
				log.Println("Unable to create new file")
				Errors[repo.FullName] = errors.New("Something went wrong when creating new file")
				continue
			}
			log.Printf("Successfully raised pull request %s on branch %s in repository %s\n", createdPR, newbranchref, repo.FullName)
		}

		log.Printf("Number of repos in csv is %d\n", len(repositories))
		if len(Errors) == 0 {
			println("No Errors where found when processing enable-all job.")
		}
		for k, v := range Errors {
			log.Printf("ERROR: Repository: [%s] Message: [%s]\n", k, v)
		}
		log.Printf("Finishing enable repo job for Organisation: %s\n", Organization)
		return
	},
}
