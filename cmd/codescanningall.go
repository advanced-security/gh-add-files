package cmd

import (
	"errors"
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
			log.Printf("Ref created succesfully at : %s\n", newbranchref)
			if len(newbranchref) <= 0 {
				log.Println("Unable to create new branch")
				Errors[repo.FullName] = errors.New("Something went wrong when creating new branch")
			}

			createdFile, err := repo.createWorkflowFile(client)
			if err != nil {
				log.Println(err)
				continue
			}
			if len(createdFile) <= 0 {
				log.Println("Unable to create commit new file")
				Errors[repo.FullName] = errors.New("Something went wrong when creating new file")
				continue
			}
			log.Printf("Successfully created file %s on branch %s in repository %s\n", createdFile, newbranchref, repo.FullName)

			createdPR, err := repo.raisePullRequest()
			if err != nil {
				log.Println(err)
				continue
			}

			if len(createdFile) <= 0 {
				log.Println("Unable to create new file")
				Errors[repo.FullName] = errors.New("Something went wrong when creating new file")
				continue
			}
			log.Printf("Successfully raised pull request %s on branch %s in repository %s\n", createdPR, newbranchref, repo.FullName)

		}
		log.Printf("Number of repos in organziation is %d\n", len(repos))
		if len(Errors) == 0 {
			println("No Errors where found when processing enable-all job.")
		}
		for k, v := range Errors {

			log.Printf("ERROR: Repository: [%s] Message: [%s]\n", k, v)
		}

		log.Printf("Finishing enable-all job for organisation %s\n", Organization)
		return
	},
}
