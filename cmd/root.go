package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "atl",
	Short: "CLI tool for Atlassian Jira and Confluence",
	Long: `A command-line interface for interacting with Atlassian products.
Supports Jira and Confluence with 1:1 mapping to their REST APIs.`,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Global flags can be added here if needed
}
