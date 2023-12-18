package cmd

import (
	"encoding/csv"
	"fmt"
	"errors"
	"io"
	"log"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var Organization string
var WorkflowFile string
var LogFile string

var Branch string
var CsvFile string
var Force bool
var Errors = make(map[string]error)

func init() {
	codeScanningCmd.PersistentFlags().StringVarP(&Organization, "organization", "o", "", "specify Organisation to implement code scanning")
	codeScanningCmd.PersistentFlags().StringVarP(&WorkflowFile, "workflow", "w", "", "specify the path to the code scanning workflow file")
	codeScanningCmd.MarkPersistentFlagRequired("workflow")
	codeScanningCmd.PersistentFlags().StringVarP(&LogFile, "log", "l", "gh-add-files.log", "specify the path where the log file will be saved")
	codeScanningCmd.PersistentFlags().StringVarP(&CsvFile, "csv", "c", "", "specify the location of csv file")
	// MarkFlagsOneRequired is only available in cobra v1.8.0 that still isn't released yet (https://github.com/spf13/cobra/issues/1936#issuecomment-1669126066)
	// codeScanningCmd.MarkFlagsOneRequired("csv", "organization")
	codeScanningCmd.MarkFlagsMutuallyExclusive("csv", "organization")
	codeScanningCmd.PersistentFlags().BoolVarP(&Force, "force", "f", false, "force enable code scanning advanced setup or update the existing code scanning workflow file")
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
		if len(Organization) <= 0 && len(CsvFile) <= 0 && len(args) <= 0 {
			log.Fatalln("ERROR: Either organization flag or csv flag must be provided")
		} else if len(Organization) > 0 && len(args) > 0 {
			log.Fatalln("ERROR: You cannot provide both organization flag and repository names as arguments")
		} else if len(CsvFile) > 0 && len(args) > 0 {
			log.Fatalln("ERROR: You cannot provide both csv flag and repository names as arguments")
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
				log.Printf("Retrieving Repository: %s \n", repository)
				repo, err := getRepo(repository)
				if err != nil {
					Errors[repository] = err
				}
				repos = append(repos, repo)
			}
		} else if len(args) > 0 {
			for _, repository := range args {
				log.Printf("Retrieving Repository: %s \n", repository)
				repo, err := getRepo(repository)
				if err != nil {
					Errors[repository] = err
				} else {
				repos = append(repos, repo)
				}
			}
		} else {
			log.Printf("Retrieving Repositories for the Organization: %s \n", Organization)

			if repos, err = getRepos(Organization); err != nil {
				log.Fatalln(err)
			}
		}

		var pullRequests []string
		var defaultScan []string
		var noLanguage []string
		var advancedSetup []string

		for _, repo := range repos {

			log.Printf("Details for Repository: Full Name: %s; Name: %s; Default Branch: %s\n", repo.FullName, repo.Name, repo.DefaultBranch)
			//check that repo has at least one codeql supported language
			coverage, err := repo.GetCodeqlLanguages()
			if err != nil {
				log.Printf("ERROR: Unable to get repo languages, skipping repository \"%s\"\n Error Message: %s\n", repo.FullName, err)
				continue
			}

			if len(coverage) <= 0 {
				log.Printf("No CodeQL supported language found for repository: %s", repo.FullName)
				noLanguage = append(noLanguage, repo.FullName)
				continue
			}

			//check that default setup is not enabled
			isDefaultSetupEnabled, err := repo.checkDefaultSetupEnabled()
			if err != nil {
				log.Println(err)
				Errors[repo.FullName] = err
				continue
			}
			if isDefaultSetupEnabled == true && Force == false {
				log.Printf("Default setup already enabled for this repository: %s, skipping enablement.", repo.FullName)
				defaultScan = append(defaultScan, repo.FullName)
				continue
			} else if isDefaultSetupEnabled == true && Force == true {
				log.Printf("Default setup already enabled for this repository: %s, but force flag is set, converting repo to advanced setup", repo.FullName)

				result, err := repo.disableDefaultSetup()
				if err != nil {
					Errors[repo.FullName] = err
					continue
				}

				if result == true {
					log.Printf("Default setup disabled for repository: %s", repo.FullName)
				}
			}

			//check that codeql workflow file doesn't already exist
			isCodeQLEnabled, sha, err := repo.doesCodeqlWorkflowExist()
			if err != nil {
				log.Println(err)
				Errors[repo.FullName] = err
				continue
			}
			if isCodeQLEnabled == true && Force == false {
				log.Printf("CodeQL workflow file already exists for this repository: %s, skipping enablement.", repo.FullName)
				advancedSetup = append(advancedSetup, repo.FullName)
				continue
			} else if isCodeQLEnabled == true && Force == true {
				log.Printf("CodeQL workflow file already exists for this repository: %s, but force flag is set, updating workflow file", repo.FullName)
			}

			newbranchref, err := repo.createBranchForRepo()
			if err != nil {
				// log.Println(err)
				if strings.Contains(err.Error(), "already exists") && Force == true {
					log.Printf("Force flag is set, removing existing branch for repository: %s\n", repo.FullName)
					err := repo.deleteBranch()
					if err != nil {
						Errors[repo.FullName] = err
						continue
					}
					log.Printf("Successfully removed branch for repository: %s\n", repo.FullName)
					newbranchref, err = repo.createBranchForRepo()
					if err != nil {
						Errors[repo.FullName] = err
						continue
					}
				} else {
					Errors[repo.FullName] = err
					continue
				}
			}
			if len(newbranchref) <= 0 {
				log.Println("ERROR: Unable to create new branch")
				Errors[repo.FullName] = errors.New("Something went wrong when creating new branch")
			}
			log.Printf("Ref created succesfully at : %s\n", newbranchref)
			
			createdFile, err := repo.createWorkflowFile(WorkflowFile, sha)
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
			pullRequests = append(pullRequests, createdPR)

		}
		log.Printf("Number of repos processed: %d\n", len(repos))
		if len(Errors) == 0 {
			log.Println("No errors where found when enabling code scanning")
		}

		if len(noLanguage) > 0 {
			log.Printf("Repositories with no CodeQL supported language: %d\n", len(noLanguage))
			for _, repo := range noLanguage {
				log.Printf("Repository: %s\n", repo)
			}
		}

		if len(defaultScan) > 0 {
			log.Printf("Repositories with default setup already enabled: %d\n", len(defaultScan))
			for _, repo := range defaultScan {
				log.Printf("Repository: %s\n", repo)
			}
		}

		if len(advancedSetup) > 0 {
			log.Printf("Repositories with advanced setup already enabled: %d\n", len(advancedSetup))
			for _, repo := range advancedSetup {
				log.Printf("Repository: %s\n", repo)
			}
		}

		if len(pullRequests) > 0 {
			log.Printf("Pull requests raised: %d\n", len(pullRequests))
			for _, pr := range pullRequests {
				log.Printf("PR URL: %s\n", pr)
			}
		}

		if len(Errors) > 0 {
			log.Printf("Repositories with errors: %d\n", len(Errors))
			for k, v := range Errors {

				log.Printf("Repository: %s Message: [%s]\n", k, v)
			}
		}

		log.Printf("Finished enable code scanning! \n")
		return
	},
}
