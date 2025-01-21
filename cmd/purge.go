package cmd

import (
	"fmt"
	"os"
	"rtt/rabbit"
	"rtt/rttio"

	"github.com/rabbitmq/amqp091-go"
	"github.com/spf13/cobra"
)

var queueToPurge string

// TODO: duplicated from rttio
const RESET = "\033[0m"
const RED = "\033[31m"
const GREEN = "\033[32m"

// purgeCmd represents the purge command
var purgeCmd = &cobra.Command{
	Use:   "purge",
	Short: "Purges all messages of one or all queues defined in a setup.json",
	Long: `Purges all messages of one or all queues defined in a setup.json
	You can select a specific queue with 'rtt purge setup.json -q <queueName>'.
	If you don't select a specific queue all messages from all queues defined in the setup.json will be purged`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			_, _ = fmt.Fprintf(os.Stderr, "purge command expects a setup.json file as argument\n")
			os.Exit(1)
		}
		setupFile := args[0]

		_, err := os.Stat(setupFile)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "no file found at '%v'\n", setupFile)
			os.Exit(1)
		}

		purgeMessages(setupFile)
	},
}

func purgeMessages(setupFile string) {
	setupData := rttio.LoadSetupFile(setupFile)

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
				RED,
				err,
				RESET)
		} else {
			fmt.Printf("%v: %v%v message(s) purged%v\n",
				q,
				GREEN,
				purgedMessagesAmount,
				RESET)
		}
	}
}

func init() {
	rootCmd.AddCommand(purgeCmd)
	purgeCmd.Flags().StringVarP(&queueToPurge, "queue", "q", "", "Queue to purge all messages from")
}
