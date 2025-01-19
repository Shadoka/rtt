package schemas

import (
	_ "embed"
	"fmt"
	"strings"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

const SETUP_SCHEMA_URL string = "http://www.rtt.fake/internal/setup"
const RTT_FILE_SCHEMA_URL string = "http://www.rtt.fake/internal/rttfile"

//go:embed setup.schema.json
var setupSchema string

//go:embed rttfile.schema.json
var rttFileSchema string

var RttValidator *jsonschema.Compiler

func Init() {
	RttValidator = jsonschema.NewCompiler()
	RttValidator.UseLoader(EmbeddedFileLoader{})
}

type EmbeddedFileLoader struct{}

func (fl EmbeddedFileLoader) Load(url string) (any, error) {
	if url == SETUP_SCHEMA_URL {
		return jsonschema.UnmarshalJSON(strings.NewReader(setupSchema))
	} else if url == RTT_FILE_SCHEMA_URL {
		return jsonschema.UnmarshalJSON(strings.NewReader(rttFileSchema))
	}
	return nil, fmt.Errorf("unknown internal schema url: %v", url)
}
