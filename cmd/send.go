/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"
	"rtt/constants"
	"rtt/data"
	"rtt/rabbit"
	"rtt/rttio"

	"github.com/spf13/cobra"
)

// sendCmd represents the send command
var sendCmd = &cobra.Command{
	Use:   "send",
	Short: "Sends a message to a queue",
	Long: `Sends a message to a queue.
	Use -f to specify a setup.json or use the default namespace
	Example: rtt send queueName '{"someKey":"someValue"}'`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 2 {
			fmt.Println("This command requires the name of a queue as parameter and a payload")
			os.Exit(1)
		}

		setupFile := VerifyFile(setupLocation)

		queueName := args[0]
		payload := args[1]
		namespace := rttio.LoadSetupFile(setupFile)
		if namespace.ContainsQueue(queueName) {
			send(payload, queueName, namespace)
		} else {
			fmt.Printf("Namespace '%v' does not contain a queue with name '%v'", setupFile, queueName)
			os.Exit(1)
		}
	},
}

func send(payload string, queueName string, namespace data.SetupFile) {
	connection := rabbit.Connect(namespace.Connection.Host, namespace.Connection.Port, namespace.Connection.User, namespace.Connection.Password)
	defer connection.Close()
	ch := rabbit.GetChannel(connection)
	defer ch.Close()

	queueInfo := namespace.GetQueueInfo(queueName)

	rabbit.SendTextMessage(payload, queueInfo, ch)

	rttio.PrintlnInColor(fmt.Sprintf("Successfully sent message to queue '%v'!", queueName), constants.GREEN)
}

func init() {
	rootCmd.AddCommand(sendCmd)
}
