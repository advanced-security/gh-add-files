# gh add-files : A GitHub CLI Extension

The `gh add files` is a GitHub CLI Extension that allows you to add files to your GitHub repositories directly from the command line.

This tool currently streamlies the process of enabling advanced setup for Code Scanning to your repositories.

### Prerequisites 

1. Install gh-cli. For further instructions please see [here]https://github.com/cli/cli#installation 

2. This extension modifies files in the `.github/workflows` directory. Therefore you must authenticate with the `workflow` and `project` scope. You will also need write access to all required repositories. For example run command `gh auth login -s "workflow project". Alternatively, you can authenticate with a PAT that has the required scope.

## Installation 

Once prerequisites are met, run the following command:
```bash
gh extension install add-files
```

## Features

### Code Scanning

To enable advanced setup for code scanning, you can use the following command with the following usage:

```bash
gh add-files code-scanning
Add / Update the codeql.yml file in a repository via a PR

Usage:
  add-files code-scanning [flags]

Flags:
  -c, --csv string            specify the location of csv file
  -f, --force                 force enable code scanning advanced setup or update the existing code scanning workflow file
  -h, --help                  help for code-scanning
  -l, --log string            specify the path where the log file will be saved (default "gh-add-files.log")
  -o, --organization string   specify Organisation to implement code scanning
  -t, --template string       specify the path to the code scanning workflow template file
  -w, --workflow string       specify the path to the code scanning workflow file 
```

The code-scanning command accepts the following three input sources:

- `c` - A CSV file containing a list of repositories to enable code scanning for. The CSV file's format is straightforward, consisting of a single column where each row specifies a repository in the format `{OWNER}/{REPO}`. No heading is required for this csv. You can refer to the examples/test.csv file in this repository for an illustration.
- `o` - An organization to enable code scanning for. This will enable code scanning for all repositories within the organization.
- standard input - A space separated list of repositories to enable code scanning for.

You cannot specify more than one of these input sources.

#### codeql.yml

There are two ways to push a `codeql.yml` file to your repository:

- You can specify the path to a `codeql.yml` file using the `-w` flag. This file will be pushed to the repository as is.
- You can specify the path to a `codeql.yml` template file using the `-t` flag. This template file will be used to generate a `codeql.yml` file, which will then be pushed to the repository. The template file is used if you want to dynamically generate a `codeql.yml` where the default branch will be different for every repo. The tool will determine the default branch for the repository and update the template file for the repository.

#### Force Flag

The `-f` flag allows you to force enable code scanning advanced setup or update the existing code scanning workflow file. If default setup is currently enabled or if advanced setup is already enabled in the repository, this flag will disable default setup. If advanced setup is already enabled, this flag will open a PR to update the file. repository.

#### Usage Examples

To enable code scanning for all repositories within an organization, run the following command:
```bash
gh add-files code-scanning -o ORG_NAME -w WORKFLOW_FILE
```

To enable code scanning for a list of repositories specified in a CSV file, run the following command:
```bash
gh add-files code-scanning -c CSV_FILE -w WORKFLOW_FILE
```

To enable code scanning for a list of repositories specified in standard input, run the following command:
```bash
gh add-files code-scanning -w WORKFLOW_FILE ORG/REPO1 ORG/REPO2
```

To enable code scanning for all repositories within an organization using a template file, run the following command:
```bash
gh add-files code-scanning -o ORG_NAME -t TEMPLATE_FILE
```

### Delete Branch 

This feature provides the capability to remove a branch across all repositories within an organization, based on its branch name. This functionality is designed for convenient branch cleanup, allowing you to execute a single command to achieve this goal.

The command to accomplish this is:

```bash
gh add-files delete-branch -o ORGANISATION -l LOG_FILE -b BRANCH_NAME
```
The following flags are mandatory: 
- `-o` - specifies the organisation you want to roll out code scanning to
- `-b` - branch to be deleted e.g `gh-cli/codescanningworkflow`
- `-l` - specify the path where the log file will be saved





