package cmd

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"rtt/consumer"
	"rtt/data"
	"rtt/rabbit"
	"rtt/rttio"
	"strings"

	"github.com/spf13/cobra"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Executes either a single rtt file or traverses a directory",
	Long:  `Executes either a single rtt file or traverses a directory`,
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: Purge all messages in response queues?
		if len(args) == 0 {
			_, _ = fmt.Fprintf(os.Stderr, "run command expects a file or directory as argument\n")
			os.Exit(1)
		}
		var filename = args[0]

		fileInfo, err := os.Stat(filename)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "no file or directory found at '%v'\n", filename)
			os.Exit(1)
		}

		rabbit.Verbose = Verbose
		consumer.Init()

		var connectionData data.Connection
		if fileInfo.IsDir() {
			connectionData = SetupRabbit(filename)
			_ = runDirectory(filename)
		} else {
			valResult := runFile(filename)
			rttio.AppendValidationResult(valResult)
			connectionData = rttio.GetConnectionDataFromRttFile(filename)
		}
		rttio.PrintValidationResults()

		if len(consumer.ReplyConsumers) > 0 {
			awaitReplies(connectionData)
		}
	},
}

func awaitReplies(connectionData data.Connection) {
	responseChannel := make(chan data.ApplicationResult)
	expectedResponseCount := consumer.GetMaxReplyCount()
	currentResponseCount := 0
	successfulResponses := 0
	failedResponses := 0

	con := rabbit.Connect(connectionData.Host, connectionData.Port, connectionData.User, connectionData.Password)
	defer con.Close()
	channel := rabbit.GetChannel(con)
	defer channel.Close()

	for _, replyConsumer := range consumer.ReplyConsumers {
		go replyConsumer.ListenForReplies(channel, responseChannel)
	}

	for currentResponseCount < expectedResponseCount {
		currentResponse := <-responseChannel
		if currentResponse.AssertionError == nil {
			successfulResponses++
			currentResponseCount++
			// TODO: Print success message
		} else {
			fmt.Printf("%v\n", currentResponse.AssertionError)
			failedResponses++
			// TODO: Print failure message
		}

		fmt.Printf("Currently at %v of %v messages received\n", currentResponseCount, expectedResponseCount)
	}
}

func runDirectory(dirName string) error {
	return filepath.WalkDir(dirName, traverseDirectory)
}

func traverseDirectory(path string, d fs.DirEntry, err error) error {
	if err != nil {
		return err
	}

	if isPotentialTestFile(path, d) {
		VerbosePrintln(fmt.Sprintf("Executing test '%v'", path))
		validationResult := runFile(path)
		rttio.AppendValidationResult(validationResult)
	}

	return nil
}

func isPotentialTestFile(path string, d fs.DirEntry) bool {
	return !d.IsDir() &&
		d.Name() != "setup.json" &&
		!strings.HasSuffix(path, ".schema.json") &&
		strings.HasSuffix(path, ".json")
}

func runFile(filename string) data.ValidationResult {
	rttFile := rttio.LoadFile(filename)
	VerbosePrintln("Successfully loaded rtt file")
	validationResult := data.ValidationResult{
		TestName:        rttFile.Name,
		ValidationError: nil,
	}

	rttFilePath := filepath.Dir(filename)
	schemaLocation := fmt.Sprintf("%v/%v", rttFilePath, rttFile.InputQueue.Queue.Schema)
	err := rttio.ValidateJson(rttFile.InputQueue.Data, schemaLocation)
	if err != nil {
		validationResult.ValidationError = err
		return validationResult
	}
	VerbosePrintln("Successfully validated input message")

	connectionFileLocation := fmt.Sprintf("%v/%v", rttFilePath, rttFile.ConnectionFile)
	connectionFile := rttio.LoadSetupFile(connectionFileLocation)
	rabbitConnection := rabbit.Connect(connectionFile.Connection.Host,
		connectionFile.Connection.Port,
		connectionFile.Connection.User,
		connectionFile.Connection.Password)
	defer rabbitConnection.Close()
	VerbosePrintln("Successfully connected to rabbitmq")

	channel := rabbit.GetChannel(rabbitConnection)
	defer channel.Close()
	VerbosePrintln("Successfully created rabbitmq channel")

	inputQueue := rabbit.DeclareQueue(&rttFile.InputQueue.Queue, channel)
	VerbosePrintln(fmt.Sprintf("Connected to rabbit queue '%v'", inputQueue.Name))

	msgId := rabbit.SendMessage(rttFile.InputQueue, inputQueue, channel)
	VerbosePrintln("Successfully sent message to rabbit mq")

	if rttFile.ResponseQueue.Queue.Name != "" {
		VerbosePrintln(fmt.Sprintf("response queue: '%v'\n", rttFile.ResponseQueue.Queue.Name))
		consumer.AddResponse(msgId, rttFile.ResponseQueue)
	}

	return validationResult
}

func SetupRabbit(dirName string) data.Connection {
	dirPath := filepath.Dir(dirName)
	setupFileName := fmt.Sprintf("%v/setup.json", dirPath)
	_, err := os.Stat(setupFileName)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "no setup.json detected, aborting execution")
		os.Exit(1)
	}
	return SetupRabbitFromFile(setupFileName)
}

func init() {
	rootCmd.AddCommand(runCmd)
}
