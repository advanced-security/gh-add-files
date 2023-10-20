package cmd

import (
	"encoding/csv"
	"fmt"
	"errors"
	"io"
	"log"
	"os"

	"github.com/cli/go-gh/v2/pkg/api"

	"github.com/spf13/cobra"
)

func init() {
	codeScanningCmd.PersistentFlags().StringVarP(&Organization, "organization", "o", "", "specify Organisation to implement code scanning")
	codeScanningCmd.PersistentFlags().StringVarP(&WorkflowFile, "workflow", "w", "", "specify the path to the code scanning workflow file")
	codeScanningCmd.MarkPersistentFlagRequired("workflow")
	codeScanningCmd.PersistentFlags().StringVarP(&LogFile, "log", "l", "gh-add-files.log", "specify the path where the log file will be saved")
	codeScanningCmd.PersistentFlags().StringVarP(&CsvFile, "csv", "c", "", "specify the location of csv file")
	// MarkFlagsOneRequired is only available in cobra v1.8.0 that still isn't released yet (https://github.com/spf13/cobra/issues/1936#issuecomment-1669126066)
	// codeScanningCmd.MarkFlagsOneRequired("csv", "organization")
	codeScanningCmd.MarkFlagsMutuallyExclusive("csv", "organization")
}

var codeScanningCmd = &cobra.Command{
	Use:   "code-scanning",
	Short: "Add workflow files to enable code scanning",
	Long:  "Creates branch `code-scanning-automated` on each repo in organisation and checks in workflow file defined in `--workflow` flag",
	Run: func(cmd *cobra.Command, args []string) {

		//set up logging
		if len(LogFile) <= 0 {
			LogFile = "gh-add-files.log"
		}

		logFile, err := os.OpenFile(LogFile, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
		if err != nil {
			panic(err)
		}
		mw := io.MultiWriter(os.Stdout, logFile)
		log.SetOutput(mw)
		defer logFile.Close()

		log.Printf("Logging all output to %s\n", LogFile)

		// check if organization or csv file is provided
		if len(Organization) <= 0 && len(CsvFile) <= 0 {
			log.Fatalln("ERROR: Either organization flag or csv flag must be provided")
		}

		log.Println("Set up REST API Client for GitHub interactions")
		client, err := api.DefaultRESTClient()
		if err != nil {
			log.Fatalln(err)
		}

		var repos []Repository

		if len(CsvFile) > 0 {
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
				log.Printf("Retrieving Repository: %s .\n", repository)
				repo, err := getRepo(client, repository)
				if err != nil {
					log.Fatalln(err)
				}
				repos = append(repos, repo)
			}
		} else {
			log.Printf("Retrieving Repositories for the Organization: %s \n", Organization)

			if repos, err = getRepos(client); err != nil {
				log.Fatalln(err)
			}
		}

		for _, repo := range repos {

			log.Printf("Details for Repository: Full Name: %s; Name: %s; Default Branch: %s\n", repo.FullName, repo.Name, repo.DefaultBranch)
			//check that repo has at least one codeql supported language
			coverage, err := repo.GetCodeqlLanguages(client)
			if err != nil {
				log.Printf("Unable to get repo languages, skipping this repository %s error: %s\n", repo.FullName, err)
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
				log.Printf("CodeQL workflow file already exists for this repository: %s, skipping enablement.", repo.FullName)
				continue
			}

			newbranchref, err := repo.createBranchForRepo(client)
			if err != nil {
				log.Println(err)
				continue
			}
			if len(newbranchref) <= 0 {
				log.Println("ERROR: Unable to create new branch")
				Errors[repo.FullName] = errors.New("Something went wrong when creating new branch")
			}
			log.Printf("Ref created succesfully at : %s\n", newbranchref)
			
			createdFile, err := repo.createWorkflowFile(client)
			if err != nil {
				log.Println(err)
				continue
			}
			if len(createdFile) <= 0 {
				log.Println("ERROR: Unable to create commit new file")
				Errors[repo.FullName] = errors.New("Something went wrong when creating new file")
				continue
			}
			log.Printf("Successfully created file %s on branch %s in repository %s\n", createdFile, newbranchref, repo.FullName)

			createdPR, err := repo.raisePullRequest()
			if err != nil {
				log.Println(err)
				continue
			}
			if len(createdPR) <= 0 {
				log.Println("ERROR: Unable to create new pull request")
				Errors[repo.FullName] = errors.New("Something went wrong when creating new pull request")
				continue
			}
			log.Printf("Successfully raised pull request %s on branch %s in repository %s\n", createdPR, newbranchref, repo.FullName)

		}
		log.Printf("Number of repos processed: %d\n", len(repos))
		if len(Errors) == 0 {
			log.Println("No errors where found when enabling code scanning")
		}
		for k, v := range Errors {

			log.Printf("ERROR: Repository: [%s] Message: [%s]\n", k, v)
		}

		log.Printf("Finished enable code scanning! \n")
		return
	},
}
