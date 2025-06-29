package cmd

import (
	"fmt"
	"os"
	"rtt/rttio"
)

// Returns the string of the namespace to use and checks for existence.
// If cmdFile is non null and the setup file for it exists, cmdFile will be returned.
// Otherwise the default namespace will be returned.
func VerifyFile(cmdFile string) string {
	var setupFile string
	if cmdFile != "" {
		setupFile = cmdFile
	} else {
		setupFile = rttio.LoadConfigFile().DefaultNamespaceSetup
	}

	_, err := os.Stat(setupFile)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "no file found at '%v'\n", setupFile)
		os.Exit(1)
	}

	return setupFile
}
