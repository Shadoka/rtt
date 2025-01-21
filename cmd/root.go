package cmd

import (
	"fmt"
	"os"
	"rtt/rttio"
	"rtt/schemas"

	"github.com/spf13/cobra"
)

var Verbose bool

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "rtt",
	Short: "Tooling for testing/managing queues and queue processing applications",
	Long: `rtt is designed to be a handy tool for working with RabbitMQ infrastructure and applications
	processing messages out of RabbitMQ.
	It can connect to rabbit instances, create rabbit queues/exchanges/topologies, purge messages from queues
	and listen to response queues and match those responses to expectations.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&Verbose, "verbose", "v", false, "Generate verbose output")
	rttio.Init()
	schemas.Init()
}

func VerbosePrintln(text string) {
	if Verbose {
		fmt.Println(text)
	}
}
