package cmd

import (
	"io"
	"log"
	"os"

	"github.com/cli/go-gh/v2/pkg/api"

	"github.com/spf13/cobra"
)

func init() {
	codeScanningCmd.PersistentFlags().StringVarP(&Organization, "organization", "o", "", "specify Organisation to implement code scanning")
	codeScanningCmd.MarkPersistentFlagRequired("organization")
	codeScanningCmd.PersistentFlags().StringVarP(&WorkflowFile, "workflow-file", "f", "", "specify the path to the code scanning workflow file")
	codeScanningCmd.MarkPersistentFlagRequired("workflow-file")
	codeScanningCmd.PersistentFlags().StringVarP(&LogFile, "log-file", "l", "", "specify the path where the log file will be saved")
	codeScanningCmd.MarkPersistentFlagRequired("log-file")

}

var codeScanningCmd = &cobra.Command{
	Use:   "code-scanning-enable-all",
	Short: "Add workflow files to enable code scanning",
	Long:  "Creates branch `code-scanning-automated` on each repo in organisation and checks in workflow file defined in `--workflow` flag",
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

		log.Printf("Retrieving Repositories for the Organization: %s .\n", Organization)
		var repos []Repository

		if repos, err = getRepos(client); err != nil {
			log.Fatalln(err)
		}
		for _, repo := range repos {

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
		log.Printf("Number of repos in organziation is %d\n", len(repos))
		log.Printf("Successfully raised commited CodeQL workflow files and raised pull requests for organisations %s\n", Organization)
		return
	},
}
