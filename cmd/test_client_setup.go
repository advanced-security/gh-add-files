package cmd

import (
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

type Client interface {
	Request(method string, path string, body io.Reader) (*http.Response, error)
}

type TestClient struct {
	http *http.Client
}

// Request implements Client.
func (*TestClient) Request(method string, path string, body io.Reader) (*http.Response, error) {
	var jsonResponse string
	var statusCode int
	var err error

	switch {
	case method == "GET":
		jsonResponse, statusCode, err = MockGetResponse(path)
	case method == "PATCH":
		jsonResponse, statusCode, err = MockPatchResponse(path)
	case method == "POST":
		jsonResponse, statusCode, err = MockPostResponse(path)
	case method == "PUT":
		jsonResponse, statusCode, err = MockPutResponse(path)
	case method == "DELETE":
		jsonResponse, statusCode, err = MockDeleteResponse(path)
	}

	//add error handling
	if statusCode == 0 {
		log.Fatalln("Something went wrong with the mock response", err)
	}

	// Create a new response
	response := &http.Response{
		StatusCode: statusCode,
		Body:       ioutil.NopCloser(strings.NewReader(jsonResponse)),
		Header:     make(http.Header),
	}
	response.Header.Set("Content-Type", "application/json")

	//Mock Next Page Header
	MockNextPageHeader(response, path)

	// Return the response
	return response, err

	// If the method and path don't match, return an error
	//return nil, fmt.Errorf("unsupported method or path")

}
