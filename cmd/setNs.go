package cmd

import (
	"fmt"
	"log"
	"os"
	"rtt/data"
	"rtt/rttio"

	"github.com/spf13/cobra"
)

// setCmd represents the set command
var setCmd = &cobra.Command{
	Use:   "set",
	Short: "Sets the default namespace",
	Long: `Sets a default namespace.
	Usage: rtt namespace set <alias>
	
	The connection and topology information will be used for all rtt commands
	that have no setup.json as parameter defined.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			log.Fatal("namespace set requires an alias of the namespace as parameter")
		}

		namespaceSetupFile := fmt.Sprintf("%v/%v/setup.json", rttio.RttConfigDir(), args[0])
		_, err := os.Stat(namespaceSetupFile)
		if err != nil {
			log.Fatalf("cannot find setup file in '%v'", namespaceSetupFile)
		}

		rttConfig := data.ConfigFile{
			DefaultNamespaceSetup: namespaceSetupFile,
		}

		rttio.WriteConfigFile(rttConfig)
	},
}

func init() {
	namespaceCmd.AddCommand(setCmd)
}
