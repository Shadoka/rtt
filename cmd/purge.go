package cmd

import (
	"fmt"
	"os"
	"rtt/constants"
	"rtt/rabbit"
	"rtt/rttio"

	"github.com/rabbitmq/amqp091-go"
	"github.com/spf13/cobra"
)

var setupToUse string

// purgeCmd represents the purge command
var purgeCmd = &cobra.Command{
	Use:   "purge",
	Short: "Purges all messages of one or all queues defined in a setup.json (or default namespace)",
	Long: `Purges all messages of one or all queues defined in a setup.json (or default namespace)
	You can select a specific queue with 'rtt purge <queueName>'.
	If you don't select a specific queue all messages from all queues defined in the namespace will be purged.
	You can select a specific environment by using the -s parameter to give a path to a setup.json file`,
	Run: func(cmd *cobra.Command, args []string) {
		setupFile := VerifyFile(setupToUse)

		var queueToPurge string
		if len(args) != 0 {
			queueToPurge = args[0]
		}

		purgeMessages(setupFile, queueToPurge)
	},
}

func purgeMessages(setupFile string, queueToPurge string) {
	setupData := rttio.LoadSetupFile(setupFile)

	if setupData.Protected {
		if !rttio.ConfirmAction() {
			os.Exit(0)
		}
	}

	conn := rabbit.Connect(setupData.Connection.Host, setupData.Connection.Port, setupData.Connection.User, setupData.Connection.Password)
	defer conn.Close()
	ch := rabbit.GetChannel(conn)
	defer ch.Close()

	var listOfQueues = make([]string, 0)
	if queueToPurge != "" {
		listOfQueues = append(listOfQueues, queueToPurge)
	} else {
		for _, q := range setupData.Queues {
			listOfQueues = append(listOfQueues, q.Name)
		}
	}

	purgeMessagesFromQueues(listOfQueues, ch)
}

func purgeMessagesFromQueues(queues []string, channel *amqp091.Channel) {
	fmt.Printf("Purging messages from %v queue(s)\n", len(queues))

	for _, q := range queues {
		purgedMessagesAmount, err := channel.QueuePurge(q, false)
		if err != nil {
			fmt.Printf("%v: %v%v%v\n",
				q,
				constants.RED,
				err,
				constants.RESET)
		} else {
			fmt.Printf("%v: %v%v message(s) purged%v\n",
				q,
				constants.GREEN,
				purgedMessagesAmount,
				constants.RESET)
		}
	}
}

func init() {
	rootCmd.AddCommand(purgeCmd)
	purgeCmd.Flags().StringVarP(&setupToUse, "setup", "s", "", "Setup file to use for the connection")
}
