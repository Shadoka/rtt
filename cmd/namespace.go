package cmd

import (
	"fmt"
	"os"
	"rtt/data"
	"rtt/rttio"

	"github.com/spf13/cobra"
)

var setupLocation string

// namespaceCmd represents the namespace command
var namespaceCmd = &cobra.Command{
	Use:   "namespace",
	Short: "The parent command for managing rabbit namespaces",
	Long: `There are several available subcommands for managing namespaces.
	Those are as follows:
	- create
	- set
	- update
	- list`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Calling namespace by itself is not useful. Please use a subcommand")
	},
}

func GetDefaultSetupFile() *data.SetupFile {
	rttConfig := rttio.LoadConfigFile()

	_, err := os.Stat(rttConfig.DefaultNamespaceSetup)
	if err != nil {
		return nil
	}

	setupFile := rttio.LoadSetupFile(rttConfig.DefaultNamespaceSetup)
	return &setupFile
}

func init() {
	rootCmd.AddCommand(namespaceCmd)

	peekCmd.Flags().StringVarP(&setupLocation, "setup", "s", "", "Setup file to use for rabbit connection")
}
