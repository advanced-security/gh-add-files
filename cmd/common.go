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

	"github.com/cli/go-gh"
	"github.com/cli/go-gh/v2/pkg/api"
	"github.com/thedevsaddam/gojsonq/v2"
)

var Organization string
var WorkflowFile string
var LogFile string
var Scope string
var Branch string
var CsvFile string

type Repository struct {
	FullName      string `json:"full_name"`
	Name          string `json:"name"`
	DefaultBranch string `json:"default_branch"`
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
