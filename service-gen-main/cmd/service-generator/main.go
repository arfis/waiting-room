package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"gitlab.com/soluqa/bookio/service-generator/internal/generate"
)

var (
	spec = flag.String("open-api", "open-api.yaml", "open-api file specification, service will be generated using this file")
	dir  = flag.String("wd", "", "root directory for the generated service")
	help = flag.String("help", "", "information about generator")
)

func Usage() {
	fmt.Fprintf(os.Stdout, "Usage of service-generator:\n")
	fmt.Fprintf(os.Stdout, "\tservice-generator -open-api ./path/to/file/spec.yaml\n")
	fmt.Fprintf(os.Stdout, "\tservice-generator -open-api ./path/to/file/spec.yaml -wd ./servicefolder\n")
	flag.PrintDefaults()
}

func main() {
	flag.Usage = Usage
	flag.Parse()
	if len(*help) != 0 {
		fmt.Println("to print help of modules type: -help <module1,module2,...>")
		return
	}
	var workingDir string
	if len(*dir) != 0 {
		workingDir = *dir
	} else {
		d, err := os.Getwd()
		if err != nil {
			slog.Error("cannot get working directory")
			return
		}
		workingDir = d
	}
	workingDir, err := filepath.Abs(workingDir)
	if err != nil {
		slog.Error("cannot get working directory")
		return
	}
	openApi, err := filepath.Abs(*spec)
	if err != nil {
		slog.Error("cannot get working directory")
		return
	}
	slog.Info("service generator", "working directory", workingDir, "spec", openApi)
	generate.GenerateProject(workingDir, openApi)
}
