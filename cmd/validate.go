package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"rtt/data"
	"rtt/rttio"

	"github.com/spf13/cobra"
)

// validateCmd represents the validate command
var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validates either a single rtt file or traverses a directory",
	Long:  `Validates either a single rtt file or traverses a directory`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			_, _ = fmt.Fprintf(os.Stderr, "validate command expects a file or directory as argument\n")
			os.Exit(1)
		}
		var filename = args[0]

		fileInfo, err := os.Stat(filename)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "no file or directory found at '%v'\n", filename)
			os.Exit(1)
		}

		fileHandler := ValidationFileWalker{}
		if fileInfo.IsDir() {
			valResults, err := rttio.WalkDirectory(filename, &fileHandler)
			if err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "error traversing the file tree: '%v'\n", err)
				os.Exit(1)
			}
			rttio.AppendValidationResults(valResults)
		} else {
			valResult := fileHandler.HandleFile(filename)
			rttio.AppendValidationResult(valResult)
		}

		rttio.PrintValidationResults()
	},
}

type ValidationFileWalker struct {
}

func (fw *ValidationFileWalker) HandleFile(path string) data.ValidationResult {
	rttFile := rttio.LoadFile(path)

	rttDirectoryPath := filepath.Dir(path)
	schemaLocation := fmt.Sprintf("%v/%v", rttDirectoryPath, rttFile.InputQueue.Queue.Schema)
	valErr := rttio.ValidateJson(rttFile.InputQueue.Data, schemaLocation)

	return data.ValidationResult{
		TestName:        rttFile.Name,
		ValidationError: valErr,
	}
}

func init() {
	rootCmd.AddCommand(validateCmd)
}
