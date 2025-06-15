package cmd

import (
	"fmt"
	"log"
	"os"
	"rtt/data"
	"rtt/rttio"

	"github.com/spf13/cobra"
)

const PURPLE = "\033[35m"

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
			DefaultNamespaceAlias: args[0],
		}

		var previousAlias string
		if _, err = os.Stat(rttio.RttConfigFile()); err == nil {
			previousAlias = rttio.LoadConfigFile().DefaultNamespaceAlias
		}

		rttio.WriteConfigFile(rttConfig)

		if previousAlias != "" {
			fmt.Printf("Switched from %v%v%v to %v%v%v", PURPLE, previousAlias, RESET, PURPLE, args[0], RESET)
		}
	},
}

func init() {
	namespaceCmd.AddCommand(setCmd)
}
