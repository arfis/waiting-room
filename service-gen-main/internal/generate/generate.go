package generate

import (
	"bytes"
	"fmt"
	"log/slog"
	"os"

	"gitlab.com/soluqa/bookio/service-generator/internal/loader"
)

func GenerateProject(wd string, openapiFile string) {
	var output bytes.Buffer
	output.WriteString("\n")
	openApi, err := loader.LoadOpenApi(openapiFile)
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	module, err := readModule(openApi.Extensions)
	if err != nil {
		slog.Error("problem reading module from the open-api.yaml", "err", err)
		os.Exit(1)
	}

	createBaseDirLayout(wd)
	generateCode(wd, openApi, module)

	// show final output of the generator - issues and todo tasks
	if len(output.Bytes()) > 0 {
		slog.Info(output.String())
	}
}

// readModule tries to get the main module from the open-api.yaml definition
// x-configuration.module
func readModule(extensions map[string]any) (string, error) {
	xConfigurationIfc, ok := extensions["x-configuration"]
	if !ok {
		return "", fmt.Errorf("x-configuration not defined")
	}

	xConfiguration, ok := xConfigurationIfc.(map[string]any)
	if !ok {
		return "", fmt.Errorf("x-configuration wrong type, expected map, got %T", xConfigurationIfc)
	}

	module, ok := xConfiguration["module"]
	if !ok {
		return "", fmt.Errorf("module not defined")
	}

	return module.(string), nil
}
