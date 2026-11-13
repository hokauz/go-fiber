package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/getkin/kin-openapi/openapi2"
	"github.com/getkin/kin-openapi/openapi2conv"
)

func main() {
	in := flag.String("in", "docs/swagger.json", "path to swagger 2.0 json input")
	out := flag.String("out", "docs/openapi.json", "path to openapi v3 json output")
	flag.Parse()

	inputData, err := os.ReadFile(*in)
	if err != nil {
		exitf("read input %q: %v", *in, err)
	}

	var swagger2 openapi2.T
	if err := json.Unmarshal(inputData, &swagger2); err != nil {
		exitf("parse swagger input %q: %v", *in, err)
	}

	openapi3doc, err := openapi2conv.ToV3(&swagger2)
	if err != nil {
		exitf("convert swagger2 to openapi3: %v", err)
	}

	outputData, err := json.MarshalIndent(openapi3doc, "", "  ")
	if err != nil {
		exitf("marshal openapi3 output: %v", err)
	}

	if err := os.MkdirAll(dirOf(*out), 0o755); err != nil {
		exitf("create output dir for %q: %v", *out, err)
	}

	if err := os.WriteFile(*out, append(outputData, '\n'), 0o644); err != nil {
		exitf("write output %q: %v", *out, err)
	}
}

func dirOf(path string) string {
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' {
			if i == 0 {
				return "/"
			}
			return path[:i]
		}
	}
	return "."
}

func exitf(format string, a ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", a...)
	os.Exit(1)
}
