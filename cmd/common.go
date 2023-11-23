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

	// "github.com/cli/go-gh"
	"github.com/cli/go-gh/v2/pkg/api"
	"github.com/thedevsaddam/gojsonq/v2"
)

// var Organization string
// var WorkflowFile string
// var LogFile string

// var Branch string
// var CsvFile string
// var Errors = make(map[string]error)

type Repository struct {
	FullName      string `json:"full_name"`
	Name          string `json:"name"`
	DefaultBranch string `json:"default_branch"`
}

type HttpMethod int

const (
	GET HttpMethod = iota
	POST
	PUT
	DELETE
)

func getRepos(Organization string) ([]Repository, error) {

	requestPath := fmt.Sprintf("orgs/%s/repos", Organization)
	page := 1
	var allrepos []Repository
	

	for {
		log.Printf("Getting all repositories for organization: %s\n", Organization)
		data := []Repository{}

		statusCode, nextPage, err := callApi(requestPath, &data, GET)
		if err != nil {
			// check if the error is a 404

			if statusCode == 404 {
				log.Printf("ERROR: The organization %s does not exist\n", Organization)
				return allrepos, err
			} else {

			log.Printf("ERROR: Unable to get repositories for organization %s\n", Organization)
			return allrepos, err
			}
		}

		// data := []Repository{}
		// err = decodeJSONResponse(payload, &data)
		// if err != nil {
		// 	return allrepos, err
		// }

		log.Printf("Processing page: %d\n", page)
		for _, repoResponse := range data {
			//add value in data to allrepos map
			allrepos = append(allrepos, Repository{repoResponse.FullName, repoResponse.Name, repoResponse.DefaultBranch})
		}

		var hasNextPage bool
		if requestPath, hasNextPage = findNextPage(nextPage); !hasNextPage {
			break
		}
		page++
	}

	log.Printf("Number of repos in %s is %d\n", Organization, len(allrepos))
	return allrepos, nil
}

func getRepo(RepositoryName string) (Repository, error) {
	requestPath := fmt.Sprintf("repos/%s", RepositoryName)
	var repo Repository

	log.Printf("Getting repo: %s\n", RepositoryName)

	statusCode, _, err := callApi(requestPath, &repo, GET)
	if err != nil {
		// check if the error is a 404
		// var httpError *api.HTTPError
		// errors.As(err, &httpError)

		if statusCode == 404 {
			log.Printf("ERROR: The repository %s does not exist\n", RepositoryName)
			return repo, err
		} else {

		log.Printf("ERROR: Unable to get repository %s\n", RepositoryName)
		return repo, err
		}
	}

	// err = decodeJSONResponse(payload, &repo)
	// if err != nil {
	// 	return repo, err
	// }

	return repo, nil
}

func callApi(requestPath string, parseType interface{}, method HttpMethod, postBody ...[]byte) (int, string, error) {
    client, err := api.DefaultRESTClient()
    if err != nil {
        log.Println("ERROR: Unable to create REST client")
        return -1, "", err
    }

	var httpMethod string
	switch method {
	case POST:
		httpMethod = http.MethodPost
	case PUT:
		httpMethod = http.MethodPut
	case DELETE:
		httpMethod = http.MethodDelete
	default:
		httpMethod = http.MethodGet
	}

	var body io.Reader
	if len(postBody) > 0 {
		body = bytes.NewReader(postBody[0])
	} else {
		body = nil
	}


    response, err := client.Request(httpMethod, requestPath, body)
    if err != nil {
		var httpError *api.HTTPError
		errors.As(err, &httpError)

		return httpError.StatusCode, "", err
    }

    defer response.Body.Close()
    nextPage := response.Header.Get("Link")
    responseBody, err := io.ReadAll(response.Body)
    if err != nil {
        log.Println("ERROR: Unable to read next page link")
        return response.StatusCode, nextPage, err
    }

	err = decodeJSONResponse(responseBody, &parseType)
	if err != nil {
		log.Println("ERROR: Unable to decode JSON response")
		return response.StatusCode, nextPage, err
	}

    return response.StatusCode, nextPage, nil
}

func decodeJSONResponse(body []byte, parseType interface{}) error {
	decoder := json.NewDecoder(bytes.NewReader(body))
	err := decoder.Decode(&parseType)
	if err != nil {
		log.Println("ERROR: Unable to decode JSON response")
		return err
	}

	return nil
}

func (repo *Repository) GetCodeqlLanguages() ([]string, error) {
	var repoLanguages map[string]int
	requestPath := fmt.Sprintf("repos/%s/languages", repo.FullName)

	//get languages for repo
	_, _, err := callApi(requestPath, &repoLanguages, GET)
	// err := client.Get(fmt.Sprintf("repos/%s/languages", repo.FullName), &response)
	if err != nil {
		log.Printf("ERROR: Unable to get languages for repository %s\n", repo.FullName)
		// Errors[repo.FullName] = err
		return nil, err
	}

	validLanguages := []string{"Go", "Swift", "Csharp", "Cpp", "C", "Java", "JavaScript", "Python", "Kotlin", "Ruby"}
	var codeqlLanguages []string
	for _, validLanguage := range validLanguages {
		if _, ok := repoLanguages[validLanguage]; ok {
			codeqlLanguages = append(codeqlLanguages, validLanguage)
		}
	}
	return codeqlLanguages, nil
}

func findNextPage(nextPageLink string) (string, bool) {
	var linkRE = regexp.MustCompile(`<([^>]+)>;\s*rel="([^"]+)"`)
	for _, m := range linkRE.FindAllStringSubmatch(nextPageLink, -1) {
		if len(m) > 2 && m[2] == "next" {
			return m[1], true
		}
	}
	return "", false

}

func (repo *Repository) createBranchForRepo() (string, error) {
	//get sha for default
	repoBranches := map[string]interface{}{}
	requestPath := fmt.Sprintf("repos/%s/branches/%s", repo.FullName, repo.DefaultBranch)
	statusCode, _, err := callApi(requestPath, &repoBranches, GET)
	if statusCode == 404 {
		log.Printf("ERROR: The branch \"%s\" in %s does not exist\n", repo.DefaultBranch, repo.FullName)
		return "", err
	}
	if err != nil {
		log.Printf("ERROR: Unable to get branch %s for repository %s\n", repo.DefaultBranch, repo.FullName)
		// Errors[repo.FullName] = err
		return "", err
	}
	// err := client.Get(fmt.Sprintf("repos/%s/branches/%s", repo.FullName, repo.DefaultBranch), &repoBranches)
	sha := gojsonq.New().FromInterface(repoBranches).Find("commit.sha")

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
		log.Printf("ERROR: Unable to convert POST Ref body to json: %s\n", err)
		// Errors[repo.FullName] = err
		return "", err
	}

	var postresp interface{}
	requestPath = fmt.Sprintf("repos/%s/git/refs", repo.FullName)
	statusCode, _, err = callApi(requestPath, &postresp, POST, jsonData)
	// err = client.Post(fmt.Sprintf("repos/%s/git/refs", repo.FullName), bytes.NewReader(jsonData), &postresp)
	if statusCode == 422 {
		log.Printf("ERROR: The branch \"%s\" already exists in repo %s\n", request.Ref, repo.FullName)
		return "", err
	}
	if err != nil {
		log.Printf("ERROR: Unable to create branch for repository %s\n", repo.FullName)
		// Errors[repo.FullName] = err
		return "", err
	}
	ref := gojsonq.New().FromInterface(postresp).Find("ref")

	return fmt.Sprint(ref), nil

}

func (repo *Repository) doesCodeqlWorkflowExist() (bool, error) {
	// skipped repos - continue on error and return out if there is a response because it means the file already exists
	var response interface{}
	requestPath := fmt.Sprintf("repos/%s/contents/.github/workflows/codeql.yml", repo.FullName)
	statusCode, _, err := callApi(requestPath, &response, GET)
	if statusCode == 200 {
		log.Printf("CodeQL workflow file already exists for repo: %s\n", repo.FullName)
		return true, nil
	} else if statusCode == 404 {
		log.Printf("CodeQL workflow file does not exist for repo: %s\n", repo.FullName)
		return false, nil
	} else {
		log.Printf("ERROR: Unable to check for existence of CodeQL workflow for repository: %s\n", repo.FullName)
		// Errors[repo.FullName] = err
		return true, err
	}
	// err := client.Get(fmt.Sprintf("repos/%s/contents/.github/workflows/codeql.yml", repo.FullName), &response)
	// if err != nil {
	// 	var httpError *api.HTTPError
	// 	errors.As(err, &httpError)

	// 	if httpError.StatusCode == 404 {
	// 		log.Printf("Checked for CodeQL workflow file for the repository %s and received 404 status code, file does not exist\n", repo.FullName)
	// 		log.Println(err)
	// 		return false, nil
	// 	}
	// 	//if not 404 log and exit
	// 	return false, err
	// }
	// if response != nil {
	// 	log.Println("CodeQL workflow file already exists for this repository.")
	// 	return true, nil
	// }
	// err = errors.New(fmt.Sprintf("Something went wrong when checking for existence of CodeQL workflow for repository: %s\n", repo.FullName))
	// return true, err

}

func (repo *Repository) createWorkflowFile(WorkflowFile string) (string, error) {

	//Open file on disk
	f, err := os.Open(WorkflowFile)
	if err != nil {
		log.Println(err)
		// Errors[repo.FullName] = err
		return "", err
	}
	reader := bufio.NewReader(f)
	content, err := io.ReadAll(reader)
	if err != nil {
		log.Println(err)
		// Errors[repo.FullName] = err
		return "", err
	}

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
	if err != nil {
		log.Println(err)
		// Errors[repo.FullName] = err
		return "", err
	}

	//create workflow file
	var createresponse interface{}
	requestPath := fmt.Sprintf("repos/%s/contents/.github/workflows/codeql.yml", repo.FullName)
	statusCode, _, err := callApi(requestPath, &createresponse, PUT, jsonData)
	if statusCode == 404 {
		log.Printf("ERROR: The branch \"gh-cli/codescanningworkflow\" does not exist in repo %s\n", repo.FullName)
		return "", err
	} else if statusCode == 422 {
		log.Printf("ERROR: The file \".github/workflows/codeql.yml\" already exists in repo %s\n", repo.FullName)
		return "", err
	} else if statusCode == 201 {
		log.Printf("Successfully created CodeQL workflow file for repo %s\n", repo.FullName)
	} else {
		log.Printf("ERROR: Unable to create CodeQL workflow file for repository %s\n", repo.FullName)
		// Errors[repo.FullName] = err
		return "", err
	}

	// err = client.Put(fmt.Sprintf("repos/%s/contents/.github/workflows/codeql.yml", repo.FullName), bytes.NewReader(jsonData), &createresponse)
	// if err != nil {
	// 	log.Println(err)
		Errors[repo.FullName] = err
	// 	return "", err
	// }
	createdFile := gojsonq.New().FromInterface(createresponse).Find("content.name")
	return fmt.Sprint(createdFile), nil

}

func (repo *Repository) raisePullRequest() (string, error) {

	type PullRequestBody struct {
		Title string `json:"title"`
		Head  string `json:"head"`
		Base  string `json:"base"`
		Body  string `json:"body"`
	}

	request := PullRequestBody{
		Title: "Automated PR: CodeQL workflow added",
		Head:  "gh-cli/codescanningworkflow",
		Base:  repo.DefaultBranch,
		Body:  "This is an automated pull request adding a codeql workflow",
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		log.Println(err)
		// Errors[repo.FullName] = err
		return "", err
	}

	//create pull request
	var createPullRequest interface{}
	requestPath := fmt.Sprintf("repos/%s/pulls", repo.FullName)
	statusCode, _, err := callApi(requestPath, &createPullRequest, POST, jsonData)
	if statusCode == 201 {
		log.Printf("Successfully created pull request for repo %s\n", repo.FullName)
	} else if statusCode == 422 {
		log.Printf("ERROR: Failed to create a pull request for repository %s\n", repo.FullName)
		return "", err
	} else {
		log.Printf("ERROR: Unable to create pull request for repository %s\n", repo.FullName)
		// Errors[repo.FullName] = err
		return "", err
	}

	createdPullRequest := gojsonq.New().FromInterface(createPullRequest).Find("html_url")
	return fmt.Sprint(createdPullRequest), nil


	// // Shell out to a gh command and read its output.
	// pr, _, err := gh.Exec("pr", "create", "-R", repo.FullName, "-B", repo.DefaultBranch, "-H", "gh-cli/codescanningworkflow", "-t", "Automated PR: CodeQL workflow added", "-b", "This is an automated pull request adding a codeql workflow")
	// if err != nil {
	// 	log.Println(err)
	// 	// Errors[repo.FullName] = err
	// 	return "", err
	// }
	// return pr.String(), nil

}

func (repo *Repository) deleteBranch() error {

	//delete branch
	var deleteBranch interface{}
	requestPath := fmt.Sprintf("repos/%s/git/refs/heads/gh-cli/codescanningworkflow", repo.FullName)
	statusCode, _, err := callApi(requestPath, &deleteBranch, DELETE, nil)
	if statusCode == 204 {
		log.Printf("Successfully deleted branch for repo %s\n", repo.FullName)
	} else if statusCode == 422 {
		log.Printf("ERROR: Failed to delete branch for repository %s\n", repo.FullName)
		return err
	} else {
		log.Printf("ERROR: Unable to delete branch for repository %s\n", repo.FullName)
		// Errors[repo.FullName] = err
		return err
	}
	return nil

	// // Shell out to a gh command and read its output.
	// _, _, err := gh.Exec("api", "-X", "DELETE", fmt.Sprintf("repos/%s/git/refs/heads/gh-cli/codescanningworkflow", repo.FullName))
	// if err != nil {
	// 	log.Println(err)
	// 	// Errors[repo.FullName] = err
	// 	return err
	// }
	// return nil

}
