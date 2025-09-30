package errors

import (
	"bytes"
	"errors"
	"fmt"
	"go/format"
	"log/slog"
	"regexp"
	"sort"
	"strconv"

	"github.com/getkin/kin-openapi/openapi3"
	"gitlab.com/soluqa/bookio/service-generator/internal/generate/internal/data"
	"gitlab.com/soluqa/bookio/service-generator/internal/utils"
)

var (
	// these should match codesT
	requiredApplicationErrorCodes = []string{
		"BUSINESS_ERROR",
		"VALIDATION_ERROR",
		"MISSING_REQUIRED_FIELD_ERROR",
		"OBJECT_VERSION_MISMATCH_ERROR",
		"NOT_FOUND_ERROR",
		"SERVICE_CALL_ERROR",
		"FORBIDDEN_ERROR",
		"INTERNAL_SERVER_ERROR",
		"AUTH_HEADER",
		"UNPROCESSABLE_ENTITY_FOREIGN_KEY",
		"UNPROCESSABLE_ENTITY_UNIQUE",
	}
)

type customError struct {
	Code          string
	CodeCamelCase string
	HasValues     bool
	FormatString  bool
	Message       string
	HttpCode      string // used as a string in generated template
	Description   string
}

// GenerateCustomErrors creates prefabricated errors from the 'x-errors' definition in base part of the openAPI.
// See README.md for additional instructions about how this work.
func GenerateCustomErrors(api *openapi3.T) ([]byte, error) {
	// no API and path defined -> no need to bother with this
	if len(api.Paths) == 0 {
		slog.Info("no API and paths defined, no need to generate errors structs")
		return nil, nil
	}

	// validate that ApplicationError is present and is in correct format
	codes, err := validateApplicationError(api)
	if err != nil {
		return nil, err
	}

	// validate codes of ApplicationError
	codesPresent, err := validateApplicationErrorCodes(codes)
	if err != nil {
		return nil, err
	}

	// check if x-errors are set
	customErrors, err := utils.GetExtensionMap(api.Extensions, "x-errors")
	if err != nil {
		return nil, err
	}
	if len(customErrors) == 0 {
		return nil, nil
	}

	// and finally extract custom errors and render the template
	errorsToRender, formatted, err := extractAndValidateCustomErrors(customErrors, codesPresent)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	err = customErrorsT.Execute(&buf, map[string]interface{}{
		"Errors":    errorsToRender,
		"Formatted": formatted,
	})
	if err != nil {
		return nil, err
	}

	// format source
	formattedOutput, err := format.Source(buf.Bytes())
	if err != nil {
		return nil, err
	}

	return formattedOutput, nil
}

func validateApplicationError(api *openapi3.T) (*openapi3.SchemaRef, error) {
	var appErr *openapi3.SchemaRef
	for name, schema := range api.Components.Schemas {
		if name == data.ComponentNameApplicationError {
			appErr = schema
		}
	}
	if appErr == nil {
		// no ApplicationError found -> the developer should fix it
		return nil, errors.New("API and paths defined but no ApplicationError defined. See README.md for further instructions")
	}
	codes, codeExists := appErr.Value.Properties["code"]
	if !codeExists {
		return nil, errors.New("ApplicationError should contain field 'code' but does not. See README.md for further instructions")
	}
	_, textExists := appErr.Value.Properties["text"]
	if !textExists {
		return nil, errors.New("ApplicationError should contain field 'text' but does not. See README.md for further instructions")
	}
	if len(codes.Value.Enum) == 0 {
		return nil, errors.New("no codes defined in ApplicationError. See README.md for further instructions")
	}

	return codes, nil
}

func validateApplicationErrorCodes(codes *openapi3.SchemaRef) (map[string]struct{}, error) {
	codesPresent := make(map[string]struct{}, len(codes.Value.Enum))
	for _, code := range codes.Value.Enum {
		codeString, ok := code.(string)
		if !ok {
			return nil, errors.New("some of ApplicationError code is not of type string")
		}
		codesPresent[codeString] = struct{}{}
	}
	for _, requiredCode := range requiredApplicationErrorCodes {
		if _, ok := codesPresent[requiredCode]; !ok {
			return nil, fmt.Errorf("ApplicationError - missing required code '%s' in enum defioition, please add it. See README.md for further instructions", requiredCode)
		}
	}

	return codesPresent, nil
}

func extractAndValidateCustomErrors(customErrors map[string]any, codesPresent map[string]struct{}) ([]customError, bool, error) {
	keys := make([]string, 0, len(customErrors))
	for k := range customErrors {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	formatted := false
	errorData := make([]customError, 0, len(customErrors))
	for _, code := range keys {
		// check if code is defined in ApplicationError... if not, it should be added to keep the consistency
		if _, ok := codesPresent[code]; !ok {
			return nil, false, fmt.Errorf("the code '%s' defined in x-errors is missing in ApplicationError code enums, please add it. See README.md for further instructions", code)
		}

		values, ok := customErrors[code]
		if !ok {
			// this should never happen
			slog.Error("data inconsistency error", "errors", customErrors, "keys", keys)
			return nil, false, errors.New("data inconsistency error")
		}
		if values == nil {
			slog.Error("code does not have any parameters set", "code", code)
			return nil, false, errors.New("code does not have any parameters set")
		}
		errorRow := customError{
			Code:          code,
			CodeCamelCase: utils.SnakeCaseToCamelCase(code),
		}

		if message, ok := values.(map[string]any)["message"]; ok {
			if errorRow.Message, ok = message.(string); !ok {
				slog.Error("x-errors: message must be string", "code", code)
				return nil, false, errors.New("x-errors: message must be string")
			}
		}
		if hasValues, ok := values.(map[string]any)["hasValues"]; ok {
			if errorRow.HasValues, ok = hasValues.(bool); !ok {
				slog.Error("x-errors: hasValues must be bool", "code", code)
				return nil, false, errors.New("x-errors: hasValues must be bool")
			}
		}
		if httpCode, ok := values.(map[string]any)["httpCode"]; ok {
			httpCodeFloat, ok := httpCode.(float64)
			if !ok {
				slog.Error("x-errors: httpCode must be int", "code", code)
				return nil, false, errors.New("x-errors: httpCode must be int")
			}
			errorRow.HttpCode = strconv.FormatFloat(httpCodeFloat, 'f', -1, 64)
		}
		if description, ok := values.(map[string]any)["description"]; ok {
			if errorRow.Description, ok = description.(string); !ok {
				slog.Error("x-errors: description must be string", "code", code)
				return nil, false, errors.New("x-errors: description must be string")
			}
		}

		if errorRow.Message == "" {
			slog.Error("x-errors: message field is missing for custom error", "code", code)
			return nil, false, errors.New("x-errors: message field is missing for custom error")
		}
		if errorRow.HttpCode == "" {
			slog.Error("x-errors: HTTP code is missing for custom error", "code", code)
			return nil, false, errors.New("x-errors: HTTP code is missing for custom error")
		}

		// messages containing placeholders %s are put into Sprintf in the template
		regex, err := regexp.Compile(`%+\w`)
		if err != nil {
			slog.Error("unable to compile regex", "error", err)
			return nil, false, errors.New("unable to compile regex")
		}
		if regex.MatchString(errorRow.Message) {
			errorRow.FormatString = true
			formatted = true
		}
		errorData = append(errorData, errorRow)
	}

	return errorData, formatted, nil
}
