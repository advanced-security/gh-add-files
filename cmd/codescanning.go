package cmd

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"

	"github.com/cli/go-gh/v2"
	"github.com/cli/go-gh/v2/pkg/api"

	"github.com/spf13/cobra"
	"github.com/thedevsaddam/gojsonq/v2"
)

var Organization string
var WorkflowFile string
var LogFile string
var Scope string

func init() {
	codeScanningCmd.PersistentFlags().StringVarP(&Organization, "organization", "o", "", "specify Organisation to implement code scanning")
	codeScanningCmd.MarkPersistentFlagRequired("organization")
	codeScanningCmd.PersistentFlags().StringVarP(&WorkflowFile, "workflow-file", "f", "", "specify the path to the code scanning workflow file")
	codeScanningCmd.MarkPersistentFlagRequired("workflow-file")
	codeScanningCmd.PersistentFlags().StringVarP(&LogFile, "log-file", "l", "", "specify the path where the log file will be saved")
	codeScanningCmd.MarkPersistentFlagRequired("log-file")
	codeScanningCmd.PersistentFlags().StringVarP(&Scope, "scope", "s", "", "scope of enablement, options are: enable-all enable-repo. If enable-repo is chosen then csv needs to be passed in")
	codeScanningCmd.MarkPersistentFlagRequired("scope")
}

type Repository struct {
	FullName      string
	Name          string
	DefaultBranch string
}

var codeScanningCmd = &cobra.Command{
	Use:   "code-scanning",
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

		//validate flags
		if err := validateFlags(); err != nil {
			log.Fatalf("Error: %v\n", err)
		}
		log.Printf("Selected scope is %s\n", Scope)

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

func validateFlags() error {
	validScope := []string{"enable-all", "enable-repo"}
	for _, scope := range validScope {
		if Scope == scope {
			// scope is valid
			return nil
		}
	}
	return fmt.Errorf("Value '%s' is invalid for flag 'scope'. Valid values "+
		"come from the set %v", Scope, validScope)
}

func (repo *Repository) GetCodeqlLanguages(client *api.RESTClient) ([]string, error) {
	var response map[string]int
	err := client.Get(fmt.Sprintf("repos/%s/languages", repo.FullName), &response)
	if err != nil {
		log.Fatal(err)
	}

	validLanguages := []string{"Go", "Swift", "Csharp", "Cpp", "C", "Java", "JavaScript", "Python", "Kotlin", "Ruby"}
	var codeqlLanguages []string
	for _, validLanguage := range validLanguages {
		if _, ok := response[validLanguage]; ok {
			codeqlLanguages = append(codeqlLanguages, validLanguage)
		}
	}
	return codeqlLanguages, nil
}

func findNextPage(response *http.Response) (string, bool) {
	var linkRE = regexp.MustCompile(`<([^>]+)>;\s*rel="([^"]+)"`)
	for _, m := range linkRE.FindAllStringSubmatch(response.Header.Get("Link"), -1) {
		if len(m) > 2 && m[2] == "next" {
			return m[1], true
		}
	}
	return "", false

}

func getRepos(client *api.RESTClient) ([]Repository, error) {

	requestPath := fmt.Sprintf("orgs/%s/repos", Organization)
	page := 1
	var allrepos []Repository

	for {
		response, err := client.Request(http.MethodGet, requestPath, nil)
		if err != nil {
			log.Fatal(err)
		}
		data := []struct{ Full_Name, Name, Default_Branch string }{}
		decoder := json.NewDecoder(response.Body)
		err = decoder.Decode(&data)
		if err != nil {
			log.Fatal(err)
		}
		if err := response.Body.Close(); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Page: %d\n", page)
		for _, repoResponse := range data {
			//add value in data to allrepos map
			allrepos = append(allrepos, Repository{repoResponse.Full_Name, repoResponse.Name, repoResponse.Default_Branch})
		}

		var hasNextPage bool
		if requestPath, hasNextPage = findNextPage(response); !hasNextPage {
			break
		}
		page++
	}

	log.Printf("Number of repos is %d\n", len(allrepos))
	return allrepos, nil
}

func (repo *Repository) createBranchForRepo(client *api.RESTClient) (string, error) {
	//get sha for default
	response := map[string]interface{}{}
	err := client.Get(fmt.Sprintf("repos/%s/branches/%s", repo.FullName, repo.DefaultBranch), &response)
	sha := gojsonq.New().FromInterface(response).Find("commit.sha")

	if err != nil {
		log.Fatal(err)
	}

	type RequestBody struct {
		Ref string `json:"ref"`
		Sha string `json:"sha"`
	}
	request := RequestBody{
		Ref: "refs/heads/gh-cli/codescanningworkflow",
		Sha: fmt.Sprint(sha),
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		log.Fatalf("Error converting POST Ref body to json: %s", err)
	}

	var postresp interface{}
	err = client.Post(fmt.Sprintf("repos/%s/git/refs", repo.FullName), bytes.NewReader(jsonData), &postresp)
	if err != nil {
		log.Fatal(err)
	}
	ref := gojsonq.New().FromInterface(postresp).Find("ref")

	return fmt.Sprint(ref), nil

}

func (repo *Repository) doesCodeqlWorkflowExist(client *api.RESTClient) (bool, error) {
	// skipped repos - continue on error and return out if there is a response because it means the file already exists
	var response interface{}
	err := client.Get(fmt.Sprintf("repos/%s/contents/.github/workflows/codeql.yml", repo.FullName), &response)
	if err != nil {
		var httpError *api.HTTPError
		errors.As(err, &httpError)

		if httpError.StatusCode == 404 {
			log.Printf("Checked for CodeQL workflow file for the repository %s and received 404 status code, file does not exist\n", repo.FullName)
			log.Println(err)
			return false, nil
		}
		//if not 404 log fatal and exit
		log.Fatalln(err)
	}
	if response != nil {
		log.Panicln("CodeQL workflow file already exists for this repository.")
		return true, nil
	}
	err = errors.New(fmt.Sprintf("Something went wrong when checking for existence of CodeQL workflow for repository: %s\n", repo.FullName))
	return true, err

}

func (repo *Repository) createWorkflowFile(client *api.RESTClient) (string, error) {

	//Open file on disk
	f, _ := os.Open(WorkflowFile)
	reader := bufio.NewReader(f)
	content, _ := io.ReadAll(reader)

	encoded := base64.StdEncoding.EncodeToString((content))

	type Commiter struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	type RequestBody struct {
		Message   string   `json:"message"`
		Committer Commiter `json:"commiter"`
		Branch    string   `json:"branch"`
		Content   string   `json:"content"`
	}

	request := RequestBody{
		Message: "AUTOMATED: commited CodeQL file",
		Committer: Commiter{
			Name:  "gh-cli add-files",
			Email: "security@clsa",
		},
		Branch:  "gh-cli/codescanningworkflow",
		Content: encoded,
	}

	jsonData, err := json.Marshal(request)

	//create workflow file
	var createresponse interface{}
	err = client.Put(fmt.Sprintf("repos/%s/contents/.github/workflows/codeql.yml", repo.FullName), bytes.NewReader(jsonData), &createresponse)
	if err != nil {
		log.Fatal(err)
	}
	createdFile := gojsonq.New().FromInterface(createresponse).Find("content.name")
	return fmt.Sprint(createdFile), nil

}

func (repo *Repository) raisePullRequest() string {

	// Shell out to a gh command and read its output.
	pr, _, err := gh.Exec("pr", "create", "-R", repo.FullName, "-B", repo.DefaultBranch, "-H", "gh-cli/codescanningworkflow", "-t", "Automated PR: CodeQL workflow added", "-b", "This is an automated pull request adding a codeql workflow")
	if err != nil {
		log.Fatal(err)
	}
	return pr.String()

}
