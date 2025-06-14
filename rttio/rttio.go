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
	"rtt/schemas"
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

	validateInternalFile(byteResult, schemas.RTT_FILE_SCHEMA_URL)

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

	validateInternalFile(byteResult, schemas.SETUP_SCHEMA_URL)

	var setupFile data.SetupFile
	if err = json.Unmarshal(byteResult, &setupFile); err != nil {
		log.Fatal(err)
	}

	return setupFile
}

func LoadConfigFile() data.ConfigFile {
	fileName := RttConfigFile()
	_, err := os.Stat(fileName)
	if err != nil {
		log.Fatal(err)
	}

	fileContent, err := os.Open(fileName)
	if err != nil {
		log.Fatal(err)
	}
	defer fileContent.Close()

	byteResult, _ := io.ReadAll(fileContent)

	var configFile data.ConfigFile
	if err = json.Unmarshal(byteResult, &configFile); err != nil {
		log.Fatal(err)
	}

	return configFile
}

func WriteConfigFile(config data.ConfigFile) {
	fileName := RttConfigFile()
	fileData, err := json.Marshal(config)
	if err != nil {
		log.Fatal(err)
	}
	if err = os.WriteFile(fileName, fileData, 0777); err != nil {
		log.Fatal(err)
	}
}

func RttConfigDir() string {
	homeDir, _ := os.UserHomeDir()
	// TODO: create file with constants
	return fmt.Sprintf("%v/%v", homeDir, ".rtt")
}

func RttConfigFile() string {
	rttDir := RttConfigDir()
	return fmt.Sprintf("%v/%v", rttDir, ".rttconf")
}

// See schemas.go for valid schemas
func validateInternalFile(fileContent []byte, validationSchema string) {
	schema, err := schemas.RttValidator.Compile(validationSchema)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Unable to parse json schema file: %v\n", err)
		os.Exit(1)
	}

	var v interface{}
	err = json.Unmarshal(fileContent, &v)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Unable to unmarshal json data: %v\n", err)
		os.Exit(1)
	}

	err = schema.Validate(v)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error validating file: %v\n", err)
		os.Exit(1)
	}
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
	data.ConsumerResult | data.ValidationResult
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

func PrintAssertionResult(assertionMessage string, successful bool) {
	if successful {
		fmt.Printf("%v%v%v\n", GREEN, assertionMessage, RESET)
	} else {
		fmt.Printf("%v%v%v\n", RED, assertionMessage, RESET)
	}
}

func PrintConsumerResult(consumerResult data.ConsumerResult) {
	if consumerResult.AssertionError == nil {
		fmt.Printf("%v%v%v%v\n", GREEN, "Successfully received all messages of queue ", consumerResult.ConsumerQueue, RESET)
	} else {
		fmt.Printf("%v%v%v%v%v\n", RED, "Queue '", consumerResult.ConsumerQueue, "' has either not received all expected messages or had assertion errors", RESET)
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
