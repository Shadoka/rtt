package cmd

import (
	"fmt"
	"rtt/constants"
	"rtt/rttio"

	"github.com/spf13/cobra"
)

// showNsCmd represents the showNs command
var showNsCmd = &cobra.Command{
	Use:   "show",
	Short: "Shows exchanges and queues in a namespace",
	Long:  `Shows a hierarchical view of exchanges and queues`,
	Run: func(cmd *cobra.Command, args []string) {
		setupFile := VerifyFile(setupLocation)

		namespace := rttio.LoadSetupFile(setupFile)

		for _, v := range namespace.Exchanges {
			exchangeQueues := namespace.GetQueuesInExchange(v.Name)
			fmt.Printf("%v%v%v\n", constants.PURPLE, v.Name, constants.RESET)
			for _, bqp := range exchangeQueues {
				fmt.Printf("\t%v%v => %v%v\n", constants.CYAN, bqp.Binding, bqp.QueueName, constants.RESET)
			}
		}
	},
}

func init() {
	namespaceCmd.AddCommand(showNsCmd)
}
