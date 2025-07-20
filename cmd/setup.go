package cmd

import (
	"fmt"
	"os"
	"rtt/data"
	"rtt/rabbit"
	"rtt/rttio"

	"github.com/spf13/cobra"
)

// setupCmd represents the setup command
var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Takes a setup file and creates exchanges, queues and bindings if necessary",
	Long:  `Takes a setup file and creates exchanges, queues and bindings if necessary`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			_, _ = fmt.Fprintf(os.Stderr, "run command expects a file or directory as argument\n")
			os.Exit(1)
		}
		var fileName = args[0]

		_, err := os.Stat(fileName)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "no file found at '%v'\n", fileName)
			os.Exit(1)
		}

		rabbit.Verbose = Verbose
		SetupRabbitFromFile(fileName)
	},
}

// SetupRabbitFromFile
// Creates rabbit infrastructure from given file. File is expected to exist.
func SetupRabbitFromFile(fileName string) data.Connection {
	VerbosePrintln(fmt.Sprintf("Loading setup data from '%v'", fileName))
	setupData := rttio.LoadSetupFile(fileName)

	if setupData.Protected {
		if !rttio.ConfirmAction() {
			os.Exit(0)
		}
	}

	conn := rabbit.Connect(setupData.Connection.Host, setupData.Connection.Port, setupData.Connection.User, setupData.Connection.Password)
	defer conn.Close()
	ch := rabbit.GetChannel(conn)
	defer ch.Close()

	for _, exchange := range setupData.Exchanges {
		rabbit.DeclareExchange(exchange, ch)
	}

	for _, queue := range setupData.Queues {
		rabbit.DeclareQueue(&queue, ch)
		rabbit.BindQueue(&queue, ch)
	}

	return setupData.Connection
}

func init() {
	rootCmd.AddCommand(setupCmd)
}
