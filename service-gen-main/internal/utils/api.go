package utils

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strconv"

	"github.com/getkin/kin-openapi/openapi3"
)

func ContainsString(col []string, s string) bool {
	for i := range col {
		if col[i] == s {
			return true
		}
	}
	return false
}

func GetXGroup(component *openapi3.SchemaRef, name string) string {
	if component.Value.Extensions["x-group"] == "" {
		slog.Error("missing x-group in components/schemas", "name", name)
		os.Exit(1)
	}
	switch t := component.Value.Extensions["x-group"].(type) {
	case json.RawMessage:
		unquote, err := strconv.Unquote(string(t))
		if err != nil {
			slog.Error(err.Error())
			os.Exit(1)
		}
		return unquote
	case string:
		return component.Value.Extensions["x-group"].(string)
	default:
		return ""
	}
}

func GetExtensionMap(i map[string]any, path string) (map[string]any, error) {
	data := make(map[string]any)
	d := i[path]
	switch t := d.(type) {
	case json.RawMessage:
		err := json.Unmarshal(t, &data)
		if err != nil {
			return nil, fmt.Errorf("cannot unmarshal raw json message %w", err)
		}
	case map[string]any:
		return d.(map[string]any), nil
	}
	return data, nil
}

func IsPage(component *openapi3.SchemaRef) (bool, string) {
	var isPage bool
	switch t := component.Value.Extensions["x-page"].(type) {
	case json.RawMessage:
		if string(t) == "true" {
			return true, ""
		}
	case bool:
		return component.Value.Extensions["x-page"].(bool), ""
	}
	// find page
	for i := range component.Value.AllOf {
		if component.Value.AllOf[i].Value != nil {
			isPage, _ = IsPage(component.Value.AllOf[i])
			break
		}
	}
	// find content type
	if isPage {
		for i := range component.Value.AllOf {
			value := component.Value.AllOf[i].Value
			if value != nil {
				content := value.Properties["content"]
				if content != nil {
					cv := content.Value
					if cv != nil {
						items := cv.Items
						if items != nil {
							ref := items.Ref
							if ref != "" {
								return isPage, ref[len("#/components/schemas/"):]
							}
						}
					}
				}
			}
		}
	}
	return false, ""
}
