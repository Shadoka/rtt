package cmd

import (
	"fmt"
	"log"
	"os"
	"rtt/rttio"
	"slices"

	"github.com/spf13/cobra"
)

// deleteNsCmd represents the deleteNs command
var deleteNsCmd = &cobra.Command{
	Use:   "delete",
	Short: "Deletes a namespace",
	Long:  `Deletes a namespace from the rtt home dir`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			log.Fatal("namespace delete needs the name of the namespace to delete as argument")
		}

		nsToDelete := args[0]
		namespaces := getListOfNamespaces()

		if slices.Contains(namespaces, nsToDelete) {
			pathToDelete := fmt.Sprintf("%v/%v", rttio.RttConfigDir(), nsToDelete)
			err := os.RemoveAll(pathToDelete)
			if err != nil {
				fmt.Printf("Successfully deleted namespace %v", args[0])
			}
		} else {
			log.Fatal("namespace delete only deletes previously created namespaces")
		}

		handleDefaultNamespaceDeletion(nsToDelete)
	},
}

func handleDefaultNamespaceDeletion(deletedNs string) {
	nsConfig := rttio.LoadConfigFile()
	if nsConfig.DefaultNamespaceAlias == deletedNs {
		os.RemoveAll(rttio.RttConfigFile())
	}
}

func init() {
	namespaceCmd.AddCommand(deleteNsCmd)
}
