package cmd

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/cli/go-gh/v2/pkg/api"
)

// MockDeleteResponse simulates a DELETE HTTP response for a given path.
// It returns a JSON string, a status code, and an error if the operation fails.
func MockDeleteResponse(path string) (string, int, error) {
	switch path {
	case "repos/paradisisland/maria/git/refs/heads/gh-cli/codescanningworkflow",
		"repos/paradisisland/shiganshima/git/refs/heads/gh-cli/codescanningworkflow":
		return `{}`, 204, nil
	case "repos/paradisisland/rose/git/refs/heads/gh-cli/codescanningworkflow",
		"repos/paradisisland/marley/git/refs/heads/gh-cli/codescanningworkflow":
		return `{}`, 422, &api.HTTPError{Message: "Reference does not exist", StatusCode: 422}
	default:
		return "", 0, fmt.Errorf("MockDeleteResponse: Unhandled path: %s", path)
	}
}

// MockPutResponse mocks the response for a PUT request to a specific path.
// It returns the response body, status code, and an error.
// The response body is a JSON string representing the content of the file at the specified path.
// The status code indicates the success or failure of the request.
// An error is returned if the path is not handled by the mock.
func MockPutResponse(path string) (string, int, error) {
	switch path {

	case "repos/paradisisland/maria/contents/.github/workflows/codeql.yml":
		return `{
			"content": {
			  "name": "codeql.yml",
			  "path": ".github/workflows/codeql.yml",
			  "sha": "95b966ae1c166bd92f8ae7d1c313e738c731dfc3",
			  "size": 9,
			  "url": "https://api.github.com/repos/paradiisland/maria/contents/.github/workflows/codeql.yml?ref=main",
			  "html_url": "https://github.com/paradisisland/maria/blob/main/.github/workflows/codeql.yml",
			  "git_url": "https://api.github.com/repos/octocat/Hello-World/git/blobs/95b966ae1c166bd92f8ae7d1c313e738c731dfc3",
			  "download_url": "https://raw.githubusercontent.com/octocat/HelloWorld/master/notes/hello.txt",
			  "type": "file",
			  "_links": {
				"self": "https://api.github.com/repos/octocat/Hello-World/contents/notes/hello.txt",
				"git": "https://api.github.com/repos/octocat/Hello-World/git/blobs/95b966ae1c166bd92f8ae7d1c313e738c731dfc3",
				"html": "https://github.com/octocat/Hello-World/blob/master/notes/hello.txt"
			  }
			},
			"commit": {
			  "sha": "7638417db6d59f3c431d3e1f261cc637155684cd",
			  "node_id": "MDY6Q29tbWl0NzYzODQxN2RiNmQ1OWYzYzQzMWQzZTFmMjYxY2M2MzcxNTU2ODRjZA==",
			  "url": "https://api.github.com/repos/octocat/Hello-World/git/commits/7638417db6d59f3c431d3e1f261cc637155684cd",
			  "html_url": "https://github.com/octocat/Hello-World/git/commit/7638417db6d59f3c431d3e1f261cc637155684cd",
			  "author": {
				"date": "2014-11-07T22:01:45Z",
				"name": "Monalisa Octocat",
				"email": "octocat@github.com"
			  },
			  "committer": {
				"date": "2014-11-07T22:01:45Z",
				"name": "Monalisa Octocat",
				"email": "octocat@github.com"
			  },
			  "message": "my commit message",
			  "tree": {
				"url": "https://api.github.com/repos/octocat/Hello-World/git/trees/691272480426f78a0138979dd3ce63b77f706feb",
				"sha": "691272480426f78a0138979dd3ce63b77f706feb"
			  },
			  "parents": [
				{
				  "url": "https://api.github.com/repos/octocat/Hello-World/git/commits/1acc419d4d6a9ce985db7be48c6349a0475975b5",
				  "html_url": "https://github.com/octocat/Hello-World/git/commit/1acc419d4d6a9ce985db7be48c6349a0475975b5",
				  "sha": "1acc419d4d6a9ce985db7be48c6349a0475975b5"
				}
			  ],
			  "verification": {
				"verified": false,
				"reason": "unsigned",
				"signature": null,
				"payload": null
			  }
			}
		  }`, 200, nil

	case "repos/paradisisland/rose/contents/.github/workflows/codeql.yml":
		return `{}`, 422, &api.HTTPError{Message: "Reference already exists", StatusCode: 422}

	case "repos/paradisisland/marley/contents/.github/workflows/codeql.yml":
		return `{}`, 404, &api.HTTPError{Message: "Not Found", StatusCode: 404}

	case "repos/paradisisland/shiganshima/contents/.github/workflows/codeql.yml":
		return `{
			"content": {
				"name": "codeql.yml",
				"path": ".github/workflows/codeql.yml",
				"sha": "95b966ae1c166bd92f8ae7d1c313e738c731dfc3",
				"size": 9,
				"url": "https://api.github.com/repos/paradiisland/shiganshima/contents/.github/workflows/codeql.yml?ref=main",
				"html_url": "https://github.com/paradisisland/shiganshima/blob/main/.github/workflows/codeql.yml",
				"git_url": "https://api.github.com/repos/octocat/Hello-World/git/blobs/95b966ae1c166bd92f8ae7d1c313e738c731dfc3",
				"download_url": "https://raw.githubusercontent.com/octocat/HelloWorld/master/notes/hello.txt",
				"type": "file",
				"_links": {
				  "self": "https://api.github.com/repos/octocat/Hello-World/contents/notes/hello.txt",
				  "git": "https://api.github.com/repos/octocat/Hello-World/git/blobs/95b966ae1c166bd92f8ae7d1c313e738c731dfc3",
				  "html": "https://github.com/octocat/Hello-World/blob/master/notes/hello.txt"
				}
			  },
			  "commit": {
				"sha": "7638417db6d59f3c431d3e1f261cc637155684cd",
				"node_id": "MDY6Q29tbWl0NzYzODQxN2RiNmQ1OWYzYzQzMWQzZTFmMjYxY2M2MzcxNTU2ODRjZA==",
				"url": "https://api.github.com/repos/octocat/Hello-World/git/commits/7638417db6d59f3c431d3e1f261cc637155684cd",
				"html_url": "https://github.com/octocat/Hello-World/git/commit/7638417db6d59f3c431d3e1f261cc637155684cd",
				"author": {
				  "date": "2014-11-07T22:01:45Z",
				  "name": "Monalisa Octocat",
				  "email": "octocat@github.com"
				},
				"committer": {
				  "date": "2014-11-07T22:01:45Z",
				  "name": "Monalisa Octocat",
				  "email": "octocat@github.com"
				},
				"message": "my commit message",
				"tree": {
				  "url": "https://api.github.com/repos/octocat/Hello-World/git/trees/691272480426f78a0138979dd3ce63b77f706feb",
				  "sha": "691272480426f78a0138979dd3ce63b77f706feb"
				},
				"parents": [
				  {
					"url": "https://api.github.com/repos/octocat/Hello-World/git/commits/1acc419d4d6a9ce985db7be48c6349a0475975b5",
					"html_url": "https://github.com/octocat/Hello-World/git/commit/1acc419d4d6a9ce985db7be48c6349a0475975b5",
					"sha": "1acc419d4d6a9ce985db7be48c6349a0475975b5"
				  }
				],
				"verification": {
				  "verified": false,
				  "reason": "unsigned",
				  "signature": null,
				  "payload": null
				}
			  }

		}`, 200, nil
	default:
		return "", 0, fmt.Errorf("MockPutResponse: Unhandled path: %s", path)
	}

}

// MockPostResponse simulates a POST HTTP response for a given path.
// It returns a JSON string, a status code, and an error if the operation fails.
func MockPostResponse(path string) (string, int, error) {
	switch path {
	case "repos/paradisisland/maria/git/refs", "repos/paradisisland/shiganshima/git/refs":
		return `{
            "ref": "refs/heads/gh-cli/codescanningworkflow",
            "node_id": "MDM6UmVmcmVmcy9oZWFkcy9mZWF0dXJlQQ==",
            "url": "https://api.github.com/repos/paradisisland/maria/git/refs/heads/featureA",
            "object": {
              "type": "commit",
              "sha": "aa218f56b14c9653891f9e74264a383fa43fefbd",
              "url": "https://api.github.com/repos/paradisisland/maria/commits/aa218f56b14c9653891f9e74264a383fa43fefbd"
            }
          }`, 200, nil
	case "repos/paradisisland/rose/git/refs":
		return `{}`, 422, &api.HTTPError{Message: "Reference already exists", StatusCode: 422}
	case "repos/paradisisland/maria/pulls":
		return `{
            "url": "https://api.github.com/repos/paradisisland/maria/pulls/1347",
            "id": 1,
            "node_id": "MDExOlB1bGxSZXF1ZXN0MQ==",
            "html_url": "https://github.com/paradisisland/maria/pull/"
        }`, 201, nil
	case "repos/paradisisland/rose/pulls":
		return `{}`, 422, &api.HTTPError{Message: "Validation Failed", StatusCode: 422}
	case "repos/paradisisland/marley/pulls":
		return `{}`, 404, &api.HTTPError{Message: "Not Found", StatusCode: 404}
	default:
		return "", 0, fmt.Errorf("MockPostResponse: Unexpected path: %s", path)
	}
}

// MockPatchResponse simulates a PATCH HTTP response for a given path.
// It returns a JSON string, a status code, and an error if the operation fails.
func MockPatchResponse(path string) (string, int, error) {
	switch path {
	case "repos/paradisisland/shiganshima/code-scanning/default-setup",
		"repos/paradisisland/maria/code-scanning/default-setup":
		return `{}`, 200, nil
	case "repos/paradisisland/rose/code-scanning/default-setup":
		return `{}`, 403, &api.HTTPError{Message: "GHAS Not Enabled", StatusCode: 403}
	case "repos/paradisisland/marley/code-scanning/default-setup":
		return `{}`, 404, &api.HTTPError{Message: "Not Found", StatusCode: 404}
	default:
		return "", 0, fmt.Errorf("MockPatchResponse: Unexpected path: %s", path)
	}
}

// MockOrgGetResponses simulates a GET HTTP response for a given path.
// It returns a JSON string, a status code, and an error if the operation fails.
func MockOrgGetResponses(path string) (string, int, error) {
	switch path {
	case "orgs/paradisisland/repos":
		return `[
            {
                "full_name": "paradisisland/maria",
                "name": "maria",
                "default_branch": "main"
            },
            {
                "full_name": "paradisisland/rose",
                "name": "rose",
                "default_branch": "main"
            },
            {
                "full_name": "paradisisland/sheena",
                "name": "sheena",
                "default_branch": "main"
            },
            {
                "full_name": "paradisisland/titanforest",
                "name": "titanforest",
                "default_branch": "main"
            },
            {
                "full_name": "paradisisland/shiganshima",
                "name": "shiganshima",
                "default_branch": "main"
            }
        ]`, 200, nil
	case "orgs/sandora-desert/repos", "orgs/ansible/repos":
		return `[]`, 200, nil
	case "orgs/atotallyrealorgname/repos":
		return `[]`, 404, &api.HTTPError{Message: "Not Found", StatusCode: 404}
	default:
		return "", 0, fmt.Errorf("MockOrgGetResponses: Unexpected path: %s", path)
	}
}

// MockRepoGetResponses simulates a GET HTTP response for a given path.
// It returns a JSON string, a status code, and an error if the operation fails.
func MockRepoGetResponses(path string) (string, int, error) {
	switch path {
	case "repos/paradisisland/maria":
		return `{"full_name":"paradisisland/maria","name":"maria","default_branch":"main"}`, 200, nil
	case "repos/paradisisland/marley":
		return `[]`, 404, &api.HTTPError{Message: "Not Found", StatusCode: 404}
	case "repos/paradisisland/maria/languages":
		return `{
            "Go": 100,
            "Java": 200,
            "JavaScript": 300,
            "Python": 400
        }`, 200, nil
	case "repos/paradisisland/titanforest/languages":
		return `{}`, 200, nil
	case "repos/paradisisland/marley/languages":
		return `[]`, 404, &api.HTTPError{Message: "Not Found", StatusCode: 404}
	case "repos/paradisisland/sheena/code-scanning/default-setup":
		return `{
            "state": "configured",
            "languages": [
              "ruby",
              "python"
            ],
            "query_suite": "default",
            "updated_at": "2023-01-19T11:21:34Z",
            "schedule": "weekly"
          }`, 200, nil
	case "repos/paradisisland/maria/code-scanning/default-setup":
		return `{"state": "not-configured"}`, 200, nil
	case "repos/paradisisland/rose/code-scanning/default-setup":
		return `{}`, 403, &api.HTTPError{Message: "GHAS Not Enabled", StatusCode: 403}
	case "repos/paradisisland/marley/code-scanning/default-setup":
		return `[]`, 404, &api.HTTPError{Message: "Not Found", StatusCode: 404}
		//Removed most of the following response for brevity, but full response is here: https://docs.github.com/en/rest/branches/branches?apiVersion=2022-11-28#get-a-branch
	case "repos/paradisisland/maria/branches/main", "repos/paradisisland/rose/branches/main", "repos/paradisisland/shiganshima/branches/main":
		return `{
				"name": "main",
				"commit": {
				  "sha": "7fd1a60b01f91b314f59955a4e4d4e80d8edf11d",
				  "node_id": "MDY6Q29tbWl0MTI5NjI2OTo3ZmQxYTYwYjAxZjkxYjMxNGY1OTk1NWE0ZTRkNGU4MGQ4ZWRmMTFk",
				  "commit": {
					"author": {
					  "name": "The Octocat",
					  "email": "octocat@nowhere.com",
					  "date": "2012-03-06T23:06:50Z"
					},
					"committer": {
					  "name": "The Octocat",
					  "email": "octocat@nowhere.com",
					  "date": "2012-03-06T23:06:50Z"
					},
					"message": "Merge pull request #6 from Spaceghost/patch-1\n\nNew line at end of file.",
					"tree": {
					  "sha": "b4eecafa9be2f2006ce1b709d6857b07069b4608",
					  "url": "https://api.github.com/repos/octocat/Hello-World/git/trees/b4eecafa9be2f2006ce1b709d6857b07069b4608"
					}
				}
			}
			  }`, 200, nil
	case "repos/paradisisland/marley/branches/main":
		return `{}`, 500, &api.HTTPError{Message: "Internal Server Error", StatusCode: 500}
	case "repos/paradisisland/maria/contents/.github/workflows/codeql.yml":
		return `{}`, 404, &api.HTTPError{Message: "Not Found", StatusCode: 404}
	case "repos/paradisisland/sheena/contents/.github/workflows/codeql.yml":
		return `{
			"type": "file",
			"encoding": "base64",
			"size": 5362,
			"name": "codeql.yml",
			"path": "codeql.yml",
			"content": "IyBZb2dhIEJvmsgaW4gcHJvZ3Jlc3MhIEZlZWwgdAoKOndhcm5pbmc6IFdvc\\nZnJlZSBmUgdG8gY0byBjaGVjayBvdXQgdGhlIGFwcCwgYnV0IGJlIHN1c29t\\nZSBiYWNrIG9uY2UgaXQgaXMgY29tcGxldGUuCgpBIHdlYiBhcHAgdGhhdCBs\\nZWFkcyB5b3UgdGhyb3VnaCBhIHlvZ2Egc2Vzc2lvbi4KCltXb3Jrb3V0IG5v\\ndyFdKGh0dHBzOi8vc2tlZHdhcmRzODguZ2l0aHViLmlvL3lvZ2EvKQoKPGlt\\nZyBzcmM9InNyYy9pbWFnZXMvbWFza2FibGVfaWNvbl81MTIucG5nIiBhbHQ9\\nImJvdCBsaWZ0aW5nIHdlaWdodHMiIHdpZHRoPSIxMDAiLz4KCkRvIHlvdSBo\\nYXZlIGZlZWRiYWNrIG9yIGlkZWFzIGZvciBpbXByb3ZlbWVudD8gW09wZW4g\\nYW4gaXNzdWVdKGh0dHBzOi8vZ2l0aHViLmNvbS9za2Vkd2FyZHM4OC95b2dh\\nL2lzc3Vlcy9uZXcpLgoKV2FudCBtb3JlIGdhbWVzPyBWaXNpdCBbQ25TIEdh\\nbWVzXShodHRwczovL3NrZWR3YXJkczg4LmdpdGh1Yi5pby9wb3J0Zm9saW8v\\nKS4KCiMjIERldmVsb3BtZW50CgpUbyBhZGQgYSBuZXcgcG9zZSwgYWRkIGFu\\nIGVudHJ5IHRvIHRoZSByZWxldmFudCBmaWxlIGluIGBzcmMvYXNhbmFzYC4K\\nClRvIGJ1aWxkLCBydW4gYG5wbSBydW4gYnVpbGRgLgoKVG8gcnVuIGxvY2Fs\\nbHkgd2l0aCBsaXZlIHJlbG9hZGluZyBhbmQgbm8gc2VydmljZSB3b3JrZXIs\\nIHJ1biBgbnBtIHJ1biBkZXZgLiAoSWYgYSBzZXJ2aWNlIHdvcmtlciB3YXMg\\ncHJldmlvdXNseSByZWdpc3RlcmVkLCB5b3UgY2FuIHVucmVnaXN0ZXIgaXQg\\naW4gY2hyb21lIGRldmVsb3BlciB0b29sczogYEFwcGxpY2F0aW9uYCA+IGBT\\nZXJ2aWNlIHdvcmtlcnNgID4gYFVucmVnaXN0ZXJgLikKClRvIHJ1biBsb2Nh\\nbGx5IGFuZCByZWdpc3RlciB0aGUgc2VydmljZSB3b3JrZXIsIHJ1biBgbnBt\\nIHN0YXJ0YC4KClRvIGRlcGxveSwgcHVzaCB0byBgbWFpbmAgb3IgbWFudWFs\\nbHkgdHJpZ2dlciB0aGUgYC5naXRodWIvd29ya2Zsb3dzL2RlcGxveS55bWxg\\nIHdvcmtmbG93Lgo=\\n",
			"sha": "8d1c8b69c3fce7bea45c73efd06983e3c419a92f",
			"url": "https://api.github.com/repos/paradisisland/sheena/contents/codeql.yml",
			"git_url": "https://api.github.com/repos/paradisisland/sheena/git/blobs/3d21ec53a331a6f037a91c368710b99387d012c1",
			"html_url": "https://github.com/paradisisland/sheena/blob/master/codeql.yml",
			"download_url": "https://raw.githubusercontent.com/paradisisland/sheena/main/codeql.yml",
			"_links": {
			  "git": "https://api.github.com/repos/paradisisland/sheena/git/blobs/3d21ec53a331a6f037a91c368710b99387d012c1",
			  "self": "https://api.github.com/repos/paradisisland/sheena/contents/codeql.yml",
			  "html": "https://github.com/paradisisland/sheena/blob/main/codeql.yml"
			}
		  }`, 200, nil
	case "repos/paradisisland/shiganshima/contents/.github/workflows/codeql.yml":
		return `{
				"type": "file",
				"encoding": "base64",
				"size": 5362,
				"name": "codeql.yml",
				"path": "codeql.yml",
				"content": "IyBZb2dhIEJvmsgaW4gcHJvZ3Jlc3MhIEZlZWwgdAoKOndhcm5pbmc6IFdvc\\nZnJlZSBmUgdG8gY0byBjaGVjayBvdXQgdGhlIGFwcCwgYnV0IGJlIHN1c29t\\nZSBiYWNrIG9uY2UgaXQgaXMgY29tcGxldGUuCgpBIHdlYiBhcHAgdGhhdCBs\\nZWFkcyB5b3UgdGhyb3VnaCBhIHlvZ2Egc2Vzc2lvbi4KCltXb3Jrb3V0IG5v\\ndyFdKGh0dHBzOi8vc2tlZHdhcmRzODguZ2l0aHViLmlvL3lvZ2EvKQoKPGlt\\nZyBzcmM9InNyYy9pbWFnZXMvbWFza2FibGVfaWNvbl81MTIucG5nIiBhbHQ9\\nImJvdCBsaWZ0aW5nIHdlaWdodHMiIHdpZHRoPSIxMDAiLz4KCkRvIHlvdSBo\\nYXZlIGZlZWRiYWNrIG9yIGlkZWFzIGZvciBpbXByb3ZlbWVudD8gW09wZW4g\\nYW4gaXNzdWVdKGh0dHBzOi8vZ2l0aHViLmNvbS9za2Vkd2FyZHM4OC95b2dh\\nL2lzc3Vlcy9uZXcpLgoKV2FudCBtb3JlIGdhbWVzPyBWaXNpdCBbQ25TIEdh\\nbWVzXShodHRwczovL3NrZWR3YXJkczg4LmdpdGh1Yi5pby9wb3J0Zm9saW8v\\nKS4KCiMjIERldmVsb3BtZW50CgpUbyBhZGQgYSBuZXcgcG9zZSwgYWRkIGFu\\nIGVudHJ5IHRvIHRoZSByZWxldmFudCBmaWxlIGluIGBzcmMvYXNhbmFzYC4K\\nClRvIGJ1aWxkLCBydW4gYG5wbSBydW4gYnVpbGRgLgoKVG8gcnVuIGxvY2Fs\\nbHkgd2l0aCBsaXZlIHJlbG9hZGluZyBhbmQgbm8gc2VydmljZSB3b3JrZXIs\\nIHJ1biBgbnBtIHJ1biBkZXZgLiAoSWYgYSBzZXJ2aWNlIHdvcmtlciB3YXMg\\ncHJldmlvdXNseSByZWdpc3RlcmVkLCB5b3UgY2FuIHVucmVnaXN0ZXIgaXQg\\naW4gY2hyb21lIGRldmVsb3BlciB0b29sczogYEFwcGxpY2F0aW9uYCA+IGBT\\nZXJ2aWNlIHdvcmtlcnNgID4gYFVucmVnaXN0ZXJgLikKClRvIHJ1biBsb2Nh\\nbGx5IGFuZCByZWdpc3RlciB0aGUgc2VydmljZSB3b3JrZXIsIHJ1biBgbnBt\\nIHN0YXJ0YC4KClRvIGRlcGxveSwgcHVzaCB0byBgbWFpbmAgb3IgbWFudWFs\\nbHkgdHJpZ2dlciB0aGUgYC5naXRodWIvd29ya2Zsb3dzL2RlcGxveS55bWxg\\nIHdvcmtmbG93Lgo=\\n",
				"sha": "0ae040b692ec3e927163db2b984135aa3c088cba",
				"url": "https://api.github.com/repos/paradisisland/shiganshima/contents/codeql.yml",
				"git_url": "https://api.github.com/repos/paradisisland/shiganshima/git/blobs/3d21ec53a331a6f037a91c368710b99387d012c1",
				"html_url": "https://github.com/paradisisland/shiganshima/blob/master/codeql.yml",
				"download_url": "https://raw.githubusercontent.com/paradisisland/shiganshima/main/codeql.yml",
				"_links": {
				  "git": "https://api.github.com/repos/paradisisland/shiganshima/git/blobs/3d21ec53a331a6f037a91c368710b99387d012c1",
				  "self": "https://api.github.com/repos/paradisisland/shiganshima/contents/codeql.yml",
				  "html": "https://github.com/paradisisland/shiganshima/blob/main/codeql.yml"
				}
			  }`, 200, nil
	case "repos/paradisisland/marley/contents/.github/workflows/codeql.yml":
		return `{}`, 404, nil
	default:
		return "", 0, fmt.Errorf("MockRepoGetResponses: Unexpected path: %s", path)
	}
}

func MockGetResponse(path string) (string, int, error) {

	//Orgs
	if strings.HasPrefix(path, "orgs/") {
		return MockOrgGetResponses(path)
	}

	//Get Branches
	if strings.HasPrefix(path, "repos/") {
		return MockRepoGetResponses(path)
	}
	// Default return if no condition is met
	return "", 0, fmt.Errorf("Invalid path: %s", path)

}

func MockNextPageHeader(response *http.Response, path string) {
	if path == "orgs/ansible/repos" {
		response.Header.Set("Link", "<https://api.github.com/organizations/1507452/repos?page=2>; rel=\"next\", <https://api.github.com/organizations/1507452/repos?page=9>; rel=\"last\"")
	}
}
