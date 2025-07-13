/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"
	"os"
	"rtt/constants"
	"rtt/data"
	"rtt/rabbit"
	"rtt/rttio"
	"time"

	"github.com/spf13/cobra"
)

var peekAmount int

// peekCmd represents the peek command
var peekCmd = &cobra.Command{
	Use:   "peek",
	Short: "Show messages of a queue without consuming them",
	Long: `With this command a user can look into messages of a queue without consuming them.
	By default only the first message of the queue is shown but that behaviour can be changed
	by the -n parameter.
	Additionally the command uses the default namespace but that can be overridden by specifying
	a setup.json with -s.
	Example: rtt peek <queue> (-n 10 -s examples/setup.json)`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			fmt.Println("This command requires the name of a queue as parameter")
			os.Exit(1)
		}

		setupFile := VerifyFile(setupLocation)

		queueName := args[0]
		namespace := rttio.LoadSetupFile(setupFile)
		if namespace.ContainsQueue(queueName) {
			peek(queueName, namespace)
		} else {
			fmt.Printf("Namespace '%v' does not contain a queue with name '%v'", setupFile, queueName)
			os.Exit(1)
		}
	},
}

func peek(queueName string, namespace data.SetupFile) {
	// TODO: Create function that returns both connection and channel
	connection := rabbit.Connect(namespace.Connection.Host, namespace.Connection.Port, namespace.Connection.User, namespace.Connection.Password)
	defer connection.Close()
	ch := rabbit.GetChannel(connection)
	defer ch.Close()

	queueInfo := namespace.GetQueueInfo(queueName)
	peekedMsgs := 0
	msPerMsg := 500
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(peekAmount*msPerMsg*int(time.Millisecond)))
	defer cancel()

	consumer := rabbit.CreateConsumerWithConfiguration(ch, queueInfo, false, ctx)
	for msg := range consumer {
		payload := string(msg.Body)
		msgId := msg.MessageId

		fmt.Printf("%vMessageId%v: %v\n", constants.PURPLE, constants.RESET, msgId)
		fmt.Printf("%vPayload%v:\n%v\n", constants.PURPLE, constants.RESET, payload)
		peekedMsgs++

		if peekedMsgs == peekAmount {
			break
		}
	}
}

func init() {
	rootCmd.AddCommand(peekCmd)

	peekCmd.Flags().IntVarP(&peekAmount, "amount", "n", 1, "Amount of messages to peek into")
}
