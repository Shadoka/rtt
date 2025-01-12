package rttio

import (
	"encoding/json"
	"fmt"
	"github.com/santhosh-tekuri/jsonschema/v6"
	"os"
)

func ValidateJson(jsonData map[string]interface{}, schemaLocation string) error {
	c := jsonschema.NewCompiler()
	schema, err := c.Compile(schemaLocation)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Unable to parse json schema file: %v\n", err)
		os.Exit(1)
	}

	marshalledJson, err := json.Marshal(jsonData)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Unable to marshal json data: %v\n", err)
		os.Exit(1)
	}

	var v interface{}
	err = json.Unmarshal(marshalledJson, &v)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Unable to unmarshal json data: %v\n", err)
		os.Exit(1)
	}

	err = schema.Validate(v)
	return err
}
