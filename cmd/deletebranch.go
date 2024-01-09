package cmd

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/cli/go-gh/v2/pkg/api"
	"github.com/spf13/cobra"
)

func init() {
	deleteBranchCmd.PersistentFlags().StringVarP(&Organization, "organization", "o", "", "specify Organisation to implement code scanning")
	deleteBranchCmd.MarkPersistentFlagRequired("organization")
	deleteBranchCmd.PersistentFlags().StringVarP(&LogFile, "log-file", "l", "", "specify the path where the log file will be saved")
	deleteBranchCmd.MarkPersistentFlagRequired("log-file")
	deleteBranchCmd.PersistentFlags().StringVarP(&Branch, "branch", "b", "", "specify the branch to delete")
	deleteBranchCmd.MarkPersistentFlagRequired("branch")
}

var deleteBranchCmd = &cobra.Command{
	Use:   "delete-branch",
	Short: "Deletes branch",
	Long:  "Deletes named branch on each repo in organisation",
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

		if repos, err = getRepos(Organization, client); err != nil {
			log.Fatalln(err)
		}

		for _, repo := range repos {

			log.Printf("Details for Repository: Full Name: %s; Name: %s; Default Branch: %s\n", repo.FullName, repo.Name, repo.DefaultBranch)
			var resp interface{}
			err := client.Delete(fmt.Sprintf("repos/%s/git/refs/heads/%s", repo.FullName, Branch), &resp)
			if err != nil {
				log.Println(err)
			}
			if resp != nil {
				log.Println(resp)
			}

			log.Printf("Successfully deleted branch %s from repository %s\n", Branch, repo.FullName)
		}
	},
}
