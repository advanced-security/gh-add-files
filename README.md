# gh add-files : A GitHub CLI Extension

The `gh add files` is a GitHub CLI Extension that allows you to add files to your GitHub repositories directly from the command line.
`v1.x.x` of this tool exculsively accomodates `codeql.yml` files, that are committed to the `.github/workflows/codeql.yml` path of the repository. 

This tool streamlines the process of rolling out Code Scanning to your Organization when using centralised workflows. 




### Prerequisites 

1. Install gh-cli. For further instructions please see [here]https://github.com/cli/cli#installation 

2. This extension modifies files in the `.github/workflows` directory. Therefore you must authenticate with the `workflow` and `project` scope. For example run command `gh auth login -s "workflow project"


## Installation 

Once prerequisites are met, run the following:
