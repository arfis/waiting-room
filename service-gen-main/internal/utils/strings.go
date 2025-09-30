package utils

import (
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"strings"
)

func SnakeCaseToCamelCase(snakeCase string) string {
	arr := strings.Split(strings.ToLower(snakeCase), "_")
	res := ""
	for i := range arr {
		capitalized := cases.Title(language.English)
		res = res + capitalized.String(arr[i])
	}
	return res
}
