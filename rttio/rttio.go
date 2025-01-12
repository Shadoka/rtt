package rttio

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"rtt/data"
	"strings"
)

// TODO: Those will probably only work in MingW on Windows
const RESET = "\033[0m"
const RED = "\033[31m"
const GREEN = "\033[32m"
const CYAN = "\033[36m"

var valResults []data.ValidationResult

func LoadFile(filename string) data.RttFile {
	fileContent, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer fileContent.Close()

	byteResult, _ := io.ReadAll(fileContent)

	var rttFile data.RttFile
	if err = json.Unmarshal(byteResult, &rttFile); err != nil {
		log.Fatal(err)
	}

	return rttFile
}

func LoadSetupFile(filename string) data.SetupFile {
	fileContent, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer fileContent.Close()

	byteResult, _ := io.ReadAll(fileContent)

	var setupFile data.SetupFile
	if err = json.Unmarshal(byteResult, &setupFile); err != nil {
		log.Fatal(err)
	}

	return setupFile
}

func GetConnectionDataFromRttFile(path string) data.Connection {
	rttFile := LoadFile(path)

	rttFilePath := filepath.Dir(path)
	setupFilePath := fmt.Sprintf("%v/%v", rttFilePath, rttFile.ConnectionFile)
	setupFile := LoadSetupFile(setupFilePath)

	return setupFile.Connection
}

func WalkDirectory[T HandleFileResult](dirname string, walker FileWalker[T]) ([]T, error) {
	handleFileResults := make([]T, 0)
	walkDirErr := filepath.WalkDir(dirname, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if IsPotentialTestFile(path, d) {
			handleResult := walker.HandleFile(path)
			handleFileResults = append(handleFileResults, handleResult)
		}

		return nil
	})

	return handleFileResults, walkDirErr
}

type FileWalker[T HandleFileResult] interface {
	HandleFile(filepath string) T
}

type HandleFileResult interface {
	data.ApplicationResult | data.ValidationResult
}

func IsPotentialTestFile(path string, d fs.DirEntry) bool {
	return !d.IsDir() &&
		d.Name() != "setup.json" &&
		!strings.HasSuffix(path, ".schema.json") &&
		strings.HasSuffix(path, ".json")
}

func PrintValidationResults() {
	fmt.Println("======= SCHEMA VALIDATION RESULTS  =======")

	for _, result := range valResults {
		fmt.Printf("%v%v%v:", CYAN, result.TestName, RESET)
		if result.ValidationError == nil {
			fmt.Printf("\t%vSUCCESS%v\n", GREEN, RESET)
		} else {
			fmt.Printf("\t%vFAILURE%v\n", RED, RESET)
			fmt.Printf("%v%v%v\n", RED, result.ValidationError, RESET)
		}
	}
}

func AppendValidationResult(result data.ValidationResult) {
	valResults = append(valResults, result)
}

func AppendValidationResults(results []data.ValidationResult) {
	valResults = append(valResults, results...)
}

func Init() {
	valResults = make([]data.ValidationResult, 0)
}
