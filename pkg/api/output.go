package api

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"gopkg.in/yaml.v2"
)

var padChar = byte(' ')

// DisplayOutput prints the output based on desired variables
func DisplayOutput(outputs []*Output, outputFormat string, showNonDeprecated bool, targetVersion string) error {
	if len(outputs) == 0 {
		fmt.Println("There were no apiVersions found that match our records.")
		return nil
	}
	var err error
	var outData []byte
	switch outputFormat {
	case "normal":
		t, err := tabOut(outputs, showNonDeprecated, targetVersion, "normal")
		if err != nil {
			return err
		}
		err = t.Flush()
		if err != nil {
			return err
		}
		return nil
	case "wide":
		t, err := tabOut(outputs, showNonDeprecated, targetVersion, "wide")
		if err != nil {
			return err
		}
		err = t.Flush()
		if err != nil {
			return err
		}
		return nil
	case "json":
		outData, err = json.Marshal(outputs)
		if err != nil {
			return err
		}
		fmt.Println(string(outData))
	case "yaml":
		outData, err = yaml.Marshal(outputs)
		if err != nil {
			return err
		}
		fmt.Println(string(outData))
	default:
		fmt.Println("output format should be one of (json,yaml,normal,wide)")
	}
	return nil
}

func tabOut(outputs []*Output, showNonDeprecated bool, targetVersion string, format string) (*tabwriter.Writer, error) {
	var usableOutputs []*Output
	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 0, 15, 2, padChar, 0)

	if showNonDeprecated {
		usableOutputs = outputs
	} else {
		for _, output := range outputs {
			if output.APIVersion.IsDeprecatedIn(targetVersion) {
				usableOutputs = append(usableOutputs, output)
			}
		}
	}
	if len(usableOutputs) == 0 {
		_, err := fmt.Fprintln(w, "APIVersions were found, but none were deprecated. Try --show-all.")
		if err != nil {
			return nil, err
		}
		return w, nil
	}

	if format == "normal" {
		_, err := fmt.Fprintln(w, "NAME\t KIND\t VERSION\t REPLACEMENT\t REMOVED\t")
		if err != nil {
			return nil, err
		}
		for _, output := range usableOutputs {
			kind := output.APIVersion.Kind
			removed := fmt.Sprintf("%t", output.APIVersion.IsRemovedIn(targetVersion))
			version := output.APIVersion.Name
			name := output.Name
			replacement := output.APIVersion.ReplacementAPI

			_, err = fmt.Fprintf(w, "%s\t %s\t %s\t %s\t %s\t\n", name, kind, version, replacement, removed)
			if err != nil {
				return nil, err
			}
		}
	}

	if format == "wide" {
		_, err := fmt.Fprintln(w, "NAME\t KIND\t VERSION\t REPLACEMENT\t DEPRECATED\t DEPRECATED IN\t REMOVED\t REMOVED IN\t")
		if err != nil {
			return nil, err
		}
		for _, output := range usableOutputs {
			kind := output.APIVersion.Kind
			deprecated := fmt.Sprintf("%t", output.APIVersion.IsDeprecatedIn(targetVersion))
			removed := fmt.Sprintf("%t", output.APIVersion.IsRemovedIn(targetVersion))
			version := output.APIVersion.Name
			name := output.Name
			replacement := output.APIVersion.ReplacementAPI
			deprecatedIn := output.APIVersion.DeprecatedIn
			removedIn := output.APIVersion.RemovedIn

			_, err = fmt.Fprintf(w, "%s\t %s\t %s\t %s\t %s\t %s\t %s\t %s\t\n", name, kind, version, replacement, deprecated, deprecatedIn, removed, removedIn)
			if err != nil {
				return nil, err
			}
		}

	}
	return w, nil
}

// GetReturnCode checks for deprecated versions and returns a code.
// takes a boolean to ignore any errors.
func GetReturnCode(outputs []*Output, ignoreErrors bool, targetVersion string) int {
	if ignoreErrors {
		return 0
	}
	for _, output := range outputs {
		if output.APIVersion.IsDeprecatedIn(targetVersion) {
			return 1
		}
	}
	return 0
}
