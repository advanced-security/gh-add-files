package cmd

import (
	"encoding/csv"
	"encoding/json"
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
	codescanningrepoCmd.PersistentFlags().StringVarP(&CsvFile, "csv-file", "c", "", "scope of enablement, options are: enable-all enable-repo. If enable-repo is chosen then csv needs to be passed in")
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

			}
			repositories = append(repositories, fmt.Sprint(row[0]))
		}

		for _, repository := range repositories {

			//get repo
			requestPath := fmt.Sprintf("repos/%s", repository)
			println(requestPath)

			response, err := client.Request(http.MethodGet, requestPath, nil)
			if err != nil {
				log.Fatal(err)
			}

			var repo Repository
			body, err := io.ReadAll(response.Body)
			if err := json.Unmarshal(body, &repo); err != nil {
				log.Fatalln("Cannot unmarshal JSON")
			}
			log.Printf("Details for Repository: Full Name: %s; Name: %s; Default Branch: %s\n", repo.FullName, repo.Name, repo.DefaultBranch)
			//check that repo has at least one codeql supported language
			coverage, err := repo.GetCodeqlLanguages(client)
			if err != nil {
				log.Fatalln(err)
			}
			if len(coverage) <= 0 {
				log.Printf("No CodeQL supported language found for repository: %s", repo.FullName)
				continue
			}
			//check that codeql workflow file doesn't already exist
			isCodeQLEnabled, err := repo.doesCodeqlWorkflowExist(client)
			if err != nil {
				log.Fatal(err)
			}

			if isCodeQLEnabled == true {
				log.Printf("CodeQL workflow file already exists for this repository: %s, skipping enablement.", repo.FullName)
				continue
			}
			newbranchref, err := repo.createBranchForRepo(client)
			if err != nil {
				log.Fatalln(err)
			}
			log.Printf("Ref created succesfully at : %s\n", newbranchref)

			createdFile, err := repo.createWorkflowFile(client)
			if err != nil {
				log.Fatalln(err)
			}
			log.Printf("Successfully created file %s on branch %s in repository %s\n", createdFile, newbranchref, repo.FullName)
			createdPR := repo.raisePullRequest()
			log.Printf("Successfully raised pull request %s on branch %s in repository %s\n", createdPR, newbranchref, repo.FullName)

		}

		log.Printf("Number of repos in csv is %d\n", len(repositories))
		log.Printf("Successfully raised commited CodeQL workflow files and raised pull requests for organisations %s\n", Organization)
		return
	},
}
