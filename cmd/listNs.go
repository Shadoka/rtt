package cmd

import (
	"fmt"
	"log"
	"os"
	"rtt/rttio"

	"github.com/spf13/cobra"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Lists all available namespaces",
	Long:  `This command lists all available namespaces.`,
	Run: func(cmd *cobra.Command, args []string) {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			log.Fatal(err)
		}
		rttDir := fmt.Sprintf("%v/%v", homeDir, RTT_DIR)
		_, err = os.Stat(rttDir)
		if err != nil {
			log.Fatal("no namespaces available")
		}

		namespaces := getListOfNamespaces()

		amountDirs := 0
		for _, ns := range namespaces {
			amountDirs++
			fmt.Printf("%v%v%v\n", PURPLE, ns, RESET)
		}
		fmt.Printf("%v%v%v namespace(s) available\n", PURPLE, amountDirs, RESET)
	},
}

func getListOfNamespaces() []string {
	namespaces := make([]string, 0)

	files, err := os.ReadDir(rttio.RttConfigDir())
	if err != nil {
		fmt.Println("no rtt directory in home directory - you need to create a namespace first")
	}

	for _, file := range files {
		if file.IsDir() {
			namespaces = append(namespaces, file.Name())
		}
	}

	return namespaces
}

func init() {
	namespaceCmd.AddCommand(listCmd)
}
