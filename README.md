# gh add-files : A GitHub CLI Extension

The `gh add files` is a GitHub CLI Extension that allows you to add files to your GitHub repositories directly from the command line.
`v1.x.x` of this tool exculsively accomodates `codeql.yml` files, that are committed to the `.github/workflows/codeql.yml` path of the repository. 

This tool streamlines the process of rolling out Code Scanning to your Organization when using centralised workflows. 

### Prerequisites 

1. Install gh-cli. For further instructions please see [here]https://github.com/cli/cli#installation 

2. This extension modifies files in the `.github/workflows` directory. Therefore you must authenticate with the `workflow` and `project` scope. You will also need write access to all required repositories. For example run command `gh auth login -s "workflow project". Alternatively, you can authenticate with a PAT that has the required scope.

## Installation 

Once prerequisites are met, run the following command:
```bash
gh extension install add-files
```

## Features

### Code Scanning Enable All 

You can add a code scanning workflow file to every repository in the organisation by running the following command:
```bash
gh add-files code-scanning-enable-all -o ORG_NAME -w WORKFLOW_FILE -l LOG_FILE
```
The following flags are mandatory: 
- `-o` - specifies the organisation you want to roll out code scanning to
- `-w` - specify the path to the code scanning file
- `-l` - specify the path where the log file will be saved

This command operates by traversing all the repositories within the organization. For each repository, it performs the following steps:

1. Creates a new branch, naming it gh-cli/codescanningworkflow, branching off the default branch.

2. Commits the workflow file specified by the user using the `-w` flag. 

3. Initiates a pull request to the default branch.

In case of any errors during this process, it logs the error but continues to the next repository.

After the command completes its execution, it is strongly recommended to review the log file for any potential errors. Once any identified issues are rectified, you can rerun the command.

### Code Scanning Enable Repository

You have the option to incorporate a code scanning workflow file into multiple repositories within an organization, as defined by a CSV file. The CSV file's format is straightforward, consisting of a single column where each row specifies a repository in the format `{OWNER}/{REPO}`. No heading is required for this csv. You can refer to the examples/test.csv file in this repository for an illustration.

You can run the following command:
```bash
gh add-files code-scanning-enable-repo -o ORGANISATION -w WORKFLOW_FILE -l LOG_FILE -c CSV_FILE
```
The following flags are mandatory: 
- `-o` - specifies the organisation you want to roll out code scanning to
- `-w` - specify the path to the code scanning file
- `-l` - specify the path where the log file will be saved
- `-c` - specify the location of the csv file


This command operates by traversing all the repositories specified in the csv within the organization. For each repository, it performs the following steps:

1. Creates a new branch, naming it gh-cli/codescanningworkflow, branching off the default branch.

2. Commits the workflow file specified by the user using the `-w` flag. 

3. Initiates a pull request to the default branch.

In case of any errors during this process, it logs the error but continues to the next repository.

After the command completes its execution, it is strongly recommended to review the log file for any potential errors. Once any identified issues are rectified, you can rerun the command.

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





