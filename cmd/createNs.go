package cmd

import (
	"fmt"
	"log"
	"os"
	"rtt/constants"

	"github.com/spf13/cobra"
)

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Creates a rabbit namespace with connection info for that environment",
	Long: `Creates a rabbit namespace with connection info for that environment.
	An example call might be 'rtt namespace create dev dev/setup.json'
	The first parameter is the alias of the newly created namespace.
	The second parameter is the path to the setup.json for that rabbit environment`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 2 {
			panic("create subcommand needs two parameters for alias and setup.json file")
		}
		alias := args[0]
		setupFile := args[1]

		createNamespace(alias, setupFile)
	},
}

func createNamespace(alias string, setupFile string) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	rttDir := fmt.Sprintf("%v/%v", homeDir, constants.RTT_DIR)
	_, err = os.Stat(rttDir)
	if err != nil {
		os.Mkdir(rttDir, 0777)
	}

	namespaceDir := fmt.Sprintf("%v/%v", rttDir, alias)
	err = os.MkdirAll(namespaceDir, 0777)
	if err != nil {
		log.Fatal(err)
	}

	setupContent, err := os.ReadFile(setupFile)
	if err != nil {
		log.Fatal(err)
	}

	nsConnectionFileName := fmt.Sprintf("%v/%v", namespaceDir, constants.SETUP_FILE_NAME)
	err = os.WriteFile(nsConnectionFileName, setupContent, 0777)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Namespace %v successfully created!\n", alias)
}

func init() {
	namespaceCmd.AddCommand(createCmd)
}
