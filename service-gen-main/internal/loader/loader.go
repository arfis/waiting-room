package loader

import (
	"context"
	"fmt"

	"github.com/getkin/kin-openapi/openapi3"
)

func LoadOpenApi(openapiFile string) (*openapi3.T, error) {
	loader := openapi3.NewLoader()

	file, err := loader.LoadFromFile(openapiFile)
	if err != nil {
		return nil, fmt.Errorf("problem reading open-api.yaml %w", err)
	}

	// we check the file for possible errors before we continue
	err = file.Validate(context.Background())
	if err != nil {
		return nil, fmt.Errorf("invalid openAPI: %w", err)
	}

	return file, nil
}
