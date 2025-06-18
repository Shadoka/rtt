/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"log"

	"github.com/spf13/cobra"
)

// updateCmd represents the update command
var updateNsCmd = &cobra.Command{
	Use:   "update",
	Short: "Updates a namespace with a new setup.json",
	Long: `This command can be used to update the connection settings
	of an already existing namespace.
	Example: rtt namespace update <namespace> <setup.json>`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 2 {
			log.Fatal("namespace update needs two arguments: namespace and setup.json")
		}

		createNamespace(args[0], args[1])
	},
}

func init() {
	namespaceCmd.AddCommand(updateNsCmd)
}
