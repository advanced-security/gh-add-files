package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(codeScanningCmd)
	rootCmd.AddCommand(deleteBranchCmd)
}

var rootCmd = &cobra.Command{
	Use:   "add-files",
	Short: "Add code scanning workflows to your organisation in GitHub",
	Long:  "A GH-CLI extension that allows you to add code scanning workflows to your organisation in GitHub",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
