package data

import (
	"bytes"
	"errors"
	"fmt"
	"go/format"
	"log/slog"
	"os"
	"sort"
	"strings"
	"unicode"

	"github.com/getkin/kin-openapi/openapi3"
	"gitlab.com/soluqa/bookio/service-generator/internal/utils"
)

const (
	ComponentNamePage                   = "Page"
	ComponentNameSort                   = "Sort"
	ComponentNamePageable               = "Pageable"
	ComponentNameApplicationError       = "ApplicationError"
	ComponentNameApplicationErrorValues = "ApplicationErrorValues"
)

var (
	// skipComponentNames we want not to generate some specific component for various reasons.
	// See description for each field.
	skipComponentNames = map[string]struct{}{
		// skip pageable that has been generated before
		ComponentNamePage: {},

		// skip pageable that has been generated before
		ComponentNameSort: {},

		// skip pageable that has been generated before
		ComponentNamePageable: {},

		// this is a representation of our application error. In the apps, it is generated as an error type
		// that matches the structure defined in openAPI for 'ApplicationError'. So we do not want to create
		// duplicate as regular DTO that is not used anywhere. See README.md for further information.
		ComponentNameApplicationError:       {},
		ComponentNameApplicationErrorValues: {},
	}
)

type FileData struct {
	buf     *bytes.Buffer
	Imports []string
}

func GenerateDTO(wd string, api *openapi3.T, module string) {
	// generate pageable
	page := utils.ExecuteAndFormat(pageT, nil)
	utils.ForceSave(fmt.Sprintf("%s/internal/data/dto/page.go", wd), page)
	// generate custom types
	types := utils.ExecuteAndFormat(typesT, nil)
	utils.ForceSave(fmt.Sprintf("%s/internal/data/dto/generated.go", wd), types)

	// one fileData per file determined by x-group attribute
	fileData := make(map[string]*FileData)

	// sort keys to preserve order
	components := api.Components.Schemas
	sortedComponents := make([]string, len(components))
	i := 0
	for s := range components {
		sortedComponents[i] = s
		i++
	}
	sort.Strings(sortedComponents)
	for _, s := range sortedComponents {
		_, skip := skipComponentNames[s]
		if skip {
			continue
		}
		component := components[s]
		isPage, _ := utils.IsPage(component)
		if isPage {
			// skip page dto generation
			continue
		}
		group := utils.GetXGroup(component, s)
		if component.Value.Enum == nil && group == "" {
			slog.Error("x-group missing for DTO", "component", s)
			os.Exit(1)
		}
		if _, nok := fileData[group]; !nok {
			var b bytes.Buffer
			fileData[group] = &FileData{buf: &b}
		}
		generateComponent(wd, fileData[group], component, s, module)
	}
	for s := range fileData {
		// When enum has x-group empty (unset) then it would create ".go" file as the key is empty string
		if s != "" {
			// handle file header with Imports
			var headBuf bytes.Buffer
			// Sort imports
			sort.Strings(fileData[s].Imports)
			err := dtoHeaderT.Execute(&headBuf, fileData[s])
			if err != nil {
				slog.Error(err.Error())
				os.Exit(1)
			}
			// append file body
			headBuf.Write(fileData[s].buf.Bytes())
			headBufFormatted, err := format.Source(headBuf.Bytes())
			if err != nil {
				slog.Error(err.Error())
				os.Exit(1)
			}
			utils.ForceSave(fmt.Sprintf("%s/internal/data/dto/%s.go", wd, s), headBufFormatted)
		}
	}
}

func generateComponent(wd string, fileData *FileData, component *openapi3.SchemaRef, name string, module string) {
	if len(component.Value.Enum) > 0 {
		data := make(map[string]interface{})
		data["module"] = module
		data["packageName"] = strings.ToLower(name)
		var enumBuff bytes.Buffer
		err := dtoEnumHeaderT.Execute(&enumBuff, data)
		if err != nil {
			slog.Error(err.Error())
			os.Exit(1)
		}
		generateEnum(&enumBuff, component, name)
		enumBufferFormatted, err := format.Source(enumBuff.Bytes())
		if err != nil {
			slog.Error(err.Error())
			os.Exit(1)
		}
		utils.CreateDir(wd, fmt.Sprintf("internal/data/dto/%s", data["packageName"]))
		utils.ForceSave(fmt.Sprintf("%s/internal/data/dto/%s/%s.go", wd, data["packageName"], data["packageName"]), enumBufferFormatted)
	} else {
		generateDTO(fileData, component, name, module)
	}
}

func generateDTO(fileData *FileData, component *openapi3.SchemaRef, name string, module string) {
	data := make(map[string]interface{})
	data["name"] = name

	if component.Value.Type == "array" {
		if len(component.Value.Items.Value.AnyOf) > 0 {
			generateArrOfInt(fileData.buf, data)
			return
		}
	}

	xml := component.Value.XML
	if xml != nil {
		data["xmlName"] = name

		if xml.Name != "" {
			data["xmlName"] = xml.Name
		}

		if xml.Prefix != "" {
			data["xmlPrefix"] = xml.Prefix
		}
	}

	fields := make(map[string]map[string]interface{})
	for i := range component.Value.AllOf {
		field := make(map[string]interface{})
		ref := component.Value.AllOf[i].Ref
		if ref != "" {
			compositionType := ref[len("#/component/schemas/_"):]
			field["type"] = compositionType
			field["isComposition"] = true
			fields[compositionType] = field
		} else if component.Value.AllOf[i].Value.Type == "object" {
			getFieldsForProperties(fileData, component.Value.AllOf[i], name, fields, module, data)
		}
	}
	getFieldsForProperties(fileData, component, name, fields, module, data)
	data["fields"] = fields

	err := dtoT.Execute(fileData.buf, data)
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}

func generateArrOfInt(buffer *bytes.Buffer, data map[string]interface{}) {
	err := arrayOfIntT.Execute(buffer, data)
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}

func getFieldsForProperties(fileData *FileData, component *openapi3.SchemaRef, name string, fields map[string]map[string]interface{}, module string, data map[string]interface{}) {
	for p := range component.Value.Properties {
		prop := component.Value.Properties[p]
		field := make(map[string]interface{})
		isRequired := utils.ContainsString(component.Value.Required, p)
		var needsDive bool
		var isPrimitive bool
		if prop.Ref != "" {
			if prop.Value.Enum != nil && len(prop.Value.Enum) > 0 {
				packageName := strings.ToLower(prop.Ref[len("#/component/schemas/_"):])
				field["type"] = fmt.Sprintf("%s.%s", packageName, prop.Ref[len("#/component/schemas/_"):])
				field["isPointer"] = !isRequired
				if !contains(fileData.Imports, fmt.Sprintf("%s/internal/data/dto/%s", module, packageName)) {
					fileData.Imports = append(fileData.Imports, fmt.Sprintf("%s/internal/data/dto/%s", module, packageName))
				}
			} else {
				field["type"] = prop.Ref[len("#/component/schemas/_"):]
				field["isPointer"] = true
			}
		} else if prop.Value.Type != "" {
			field["type"], field["isPointer"], needsDive, isPrimitive = getType(prop.Value, fileData, module, isRequired)
		} else {
			slog.Error("one of type or ref must be specified", "name", name, "property", p)
		}
		validations := make([]interface{}, 0)
		omitEmpty := false
		if arrValidations, ok := prop.Value.Extensions["x-validate"]; ok {
			validations = arrValidations.([]interface{})
		}
		if !isPrimitive && isRequired {
			validations = append(validations, "required")
		} else if !isRequired {
			omitEmpty = true
		}

		if needsDive {
			validations = append(validations, "dive")
		}
		validationArr := make([]string, len(validations))
		for i, v := range validations {
			validationArr[i] = v.(string)
		}

		field["validate"] = strings.Join(validationArr[:], ",")
		field["jsonTag"] = p
		field["omitEmpty"] = omitEmpty

		if _, enableXml := data["xmlName"]; enableXml {
			field["xmlTag"] = p
		}

		xml := prop.Value.XML
		if xml != nil {
			if xml.Name != "" {
				field["xmlTag"] = xml.Name
			}

			if xml.Prefix != "" {
				field["xmlPrefix"] = xml.Prefix
			}

			if xml.Attribute {
				field["xmlAttribute"] = xml.Attribute
			}
		}

		fields[getFieldName(p)] = field
	}
}

func getFieldName(p string) string {
	if strings.HasPrefix(p, "_") {
		return "Underscore" + strings.ToUpper(p[1:2]) + p[2:]
	}
	return p
}

func getType(value *openapi3.Schema, data *FileData, module string, isRequired bool) (name string, pointer bool, needsDive bool, primitive bool) {
	// When additionalProperties are set then the field is a type of map
	if value.Type == "object" && value.AdditionalProperties.Schema != nil {
		return getMapType(value, data, module, isRequired)
	} else {
		return getSimpleFieldType(value, data, module, isRequired)
	}
}

func getSimpleFieldType(value *openapi3.Schema, data *FileData, module string, isRequired bool) (name string, pointer bool, needsDive bool, primitive bool) {
	switch value.Type {
	case "array":
		if value.Items.Ref != "" {
			if value.Items.Value.Enum != nil {
				packageName := strings.ToLower(value.Items.Ref[len("#/component/schemas/_"):])
				if !contains(data.Imports, fmt.Sprintf("%s/internal/data/dto/%s", module, packageName)) {
					data.Imports = append(data.Imports, fmt.Sprintf("%s/internal/data/dto/%s", module, packageName))
				}
				return fmt.Sprintf("[]%s.%s", packageName, value.Items.Ref[len("#/component/schemas/_"):]), false, true, false
			}
			return "[]" + value.Items.Ref[len("#/component/schemas/_"):], false, true, false
		}
		var arrayType string
		switch value.Items.Value.Type {
		case "number":
			fallthrough
		case "integer":
			arrayType = "[]" + value.Items.Value.Format
		case "object":
			arrayType = "[]map[string]interface{}"
		default:
			arrayType = "[]" + value.Items.Value.Type
		}
		return arrayType, false, true, false
	case "number":
		fallthrough
	case "integer":
		return value.Format, !isRequired, false, true
	case "boolean":
		return "bool", !isRequired, false, true
	case "string":
		switch value.Format {
		case "time":
			return "LocalTime", !isRequired, false, false
		case "date":
			return "LocalDate", !isRequired, false, false
		case "local-date-time":
			return "LocalDateTime", !isRequired, false, false
		case "date-time":
			if !contains(data.Imports, "time") {
				data.Imports = append(data.Imports, "time")
			}
			return "time.Time", !isRequired, false, false
		default:
			return value.Type, !isRequired, false, false
		}
	case "object":
		if value.Format == "interface" {
			return "interface{}", false, false, false
		}
		if value.Format == "map" {
			return "map[string]interface{}", !isRequired, false, false
		}
		return "map[string]interface{}", !isRequired, false, false
	default:
		return value.Type, !isRequired, false, false
	}
}

func getMapType(value *openapi3.Schema, data *FileData, module string, isRequired bool) (name string, pointer bool, needsDive bool, primitive bool) {
	var fieldBuilder bytes.Buffer
	var property *openapi3.SchemaRef
	for value.AdditionalProperties.Schema != nil {
		fieldBuilder.WriteString("map[string]")
		property = value.AdditionalProperties.Schema
		value = value.AdditionalProperties.Schema.Value
	}
	fieldType, _, _, _ := getSimpleFieldType(value, data, module, isRequired)
	if property != nil && property.Ref != "" {
		// property which has ref set and the reference is Enum then package has to be handled
		if value.Enum != nil && len(value.Enum) > 0 {
			packageName := strings.ToLower(property.Ref[len("#/component/schemas/_"):])
			if !contains(data.Imports, fmt.Sprintf("%s/internal/data/dto/%s", module, packageName)) {
				data.Imports = append(data.Imports, fmt.Sprintf("%s/internal/data/dto/%s", module, packageName))
			}
			fieldBuilder.WriteString(fmt.Sprintf("%s.%s", packageName, property.Ref[len("#/component/schemas/_"):]))
		} else {
			fieldBuilder.WriteString(property.Ref[len("#/component/schemas/_"):])
		}
	} else if fieldType == "object" {
		fieldBuilder.WriteString("map[string]interface{}")
	} else {
		fieldBuilder.WriteString(fieldType)
	}
	return fieldBuilder.String(), false, false, false
}

func generateEnum(buf *bytes.Buffer, component *openapi3.SchemaRef, name string) {
	data := make(map[string]interface{})
	data["type"] = component.Value.Type
	if data["type"] == "string" {
		data["isString"] = true
	}
	data["name"] = name

	// check enums for some name normalization if needed
	enumsNormalized, err := processEnums(component.Value.Enum)
	if err != nil {
		os.Exit(1)
	}
	data["fields"] = enumsNormalized

	err = enumT.Execute(buf, data)
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}

var dtoHeaderT = utils.CreateTemplate(`
// Code generated by go generate; DO NOT EDIT.
package dto
{{range .Imports}}
import "{{.}}"{{end}}
`)

var dtoEnumHeaderT = utils.CreateTemplate(`
// Code generated by go generate; DO NOT EDIT.
package {{.packageName}}

import (
	"fmt"
	"{{.module}}/internal/errors"
)
`)

var enumT = utils.CreateTemplate(`
type {{.name}} {{.type}}

var (
	UNKNOWN_VALUE       {{$.name}} = "UNKNOWN_VALUE"
{{- range .fields }}
	{{ .NameNormalized }}  {{$.name}} = {{if $.isString}}"{{ .Value }}"{{else}}{{ .Value }}{{end}}{{end}}
)

// String gets the string representation of the {{$.name}}
func (c {{$.name}}) String() string {
	return string(c)
}

func StringTo{{$.name}}(source string) ({{$.name}}, error) {
    switch source {
	{{- range .fields }}
    	case string({{ .NameNormalized }}):
			return {{ .NameNormalized }},nil{{end}}
    default:
        return "UNKNOWN_VALUE", errors.Validation(fmt.Errorf("invalid value ('%s') passed to {{$.name}}", source), nil)
    }
}
`)

var arrayOfIntT = utils.CreateTemplate(`
type {{.name}} []interface{} 
`)

var dtoT = utils.CreateTemplate(`
type {{.name}} struct {
{{if .xmlName}}    XMLName          xml.Name ` + "`" + `xml:"{{if .xmlPrefix}}{{.xmlPrefix}}:{{end}}{{.xmlName}}"` + "`" + `{{end}}
{{range $name, $value := .fields }}{{if .isComposition}}
    {{ $value.type }}{{else}}
    {{ $name | ToPublic }} {{if $value.isPointer}}*{{end}}{{$value.type}} ` + "`" + `json:"{{$value.jsonTag}}{{if $value.omitEmpty}},omitempty{{end}}"{{if $value.validate}} validate:"{{$value.validate}}"{{end}}{{if $.xmlName}} xml:"{{if $value.xmlPrefix}}{{$value.xmlPrefix}}:{{end}}{{$value.xmlTag}}{{if $value.xmlAttribute}},attr{{end}}"{{end}}` + "`" + `{{end}}{{end}}
}

{{range $name, $value := .fields }}
func ({{$.name | ToPrivate }} {{ $.name }})Get{{ $name | ToPublic }}() {{$value.type}}{
	{{- if $value.isPointer}}
	var v {{$value.type}}
	if {{$.name | ToPrivate }}.{{ $name | ToPublic }}	!= nil{
		return {{if $value.isPointer}}*{{end}}{{$.name | ToPrivate }}.{{ $name | ToPublic }}
	}
	return v
	{{- else}}	
	return {{$.name | ToPrivate }}.{{ $name | ToPublic }}
	{{- end}}	
}
{{end}}	
`)

var pageT = utils.CreateTemplate(`
// Code generated by go generate; DO NOT EDIT.
package dto

type Page struct {
	Content  interface{} ` + "`" + `json:"content"` + "`" + `
    Pageable         Pageable ` + "`" + `json:"pageable"` + "`" + `
	Empty            bool     ` + "`" + `json:"empty"` + "`" + `
	First            bool     ` + "`" + `json:"first"` + "`" + `
	Last             bool     ` + "`" + `json:"last"` + "`" + `
	NumberOfElements int32    ` + "`" + `json:"numberOfElements"` + "`" + `
	Size             int32    ` + "`" + `json:"size"` + "`" + `
	TotalElements    int64    ` + "`" + `json:"totalElements"` + "`" + `
	TotalPages       int32    ` + "`" + `json:"totalPages"` + "`" + `
	Sort             Sort     ` + "`" + `json:"sort"` + "`" + `
}

type Pageable struct {
	Page   int32 ` + "`" + `json:"page"` + "`" + `
	Size   int32 ` + "`" + `json:"size"` + "`" + `
	Offset int32 ` + "`" + `json:"offset"` + "`" + `
}

type Sort struct {
	Empty    bool ` + "`" + `json:"empty"` + "`" + `
	Sorted   bool ` + "`" + `json:"sorted"` + "`" + `
	Unsorted bool ` + "`" + `json:"unsorted"` + "`" + `
}
`)

var typesT = utils.CreateTemplate(`
// Code generated by go generate; DO NOT EDIT.
package dto

import (
	"encoding/xml"
	"strconv"
	"time"
    "errors"
)

const (
	timeOnly            = "15:04:05"
	dateOnly            = "2006-01-02"
	dateTimeWithoutZone = "2006-01-02T15:04:05"
)

type LocalTime struct {
	time.Time
}

type LocalDate struct {
	time.Time
}

type LocalDateTime struct {
	time.Time
}

func (lt *LocalTime) UnmarshalJSON(b []byte) (err error) {
	s := string(b)

	s, _ = strconv.Unquote(s)
	t, err := time.Parse(timeOnly, s)
	if err != nil {
		return errors.New("unable to unmarshal JSON")
	}
	lt.Time = t
	return nil
}

func (lt *LocalTime) UnmarshalXML(decoder *xml.Decoder, start xml.StartElement) (err error) {
	var value string

	err = decoder.DecodeElement(&value, &start)
	if err != nil {
		return errors.New("unable to decode XML")
	}
	parsedTime, err := time.Parse(timeOnly, value)
	if err != nil {
		return errors.New("unable to unmarshal XML")
	}
	lt.Time = parsedTime
	return nil
}

func (lt LocalTime) MarshalJSON() (array []byte, err error) {
	s := lt.Format(timeOnly)
	array = []byte(strconv.Quote(s))
	return array, nil
}

func (lt LocalTime) MarshalXML(encoder *xml.Encoder, start xml.StartElement) error {
	return encoder.EncodeElement(lt.Format(timeOnly), start)
}

func (ld *LocalDate) UnmarshalJSON(b []byte) (err error) {
	s := string(b)

	s, _ = strconv.Unquote(s)
	t, err := time.Parse(dateOnly, s)
	if err != nil {
		return errors.New("unable to unmarshal JSON")
	}
	ld.Time = t
	return nil
}

func (ld *LocalDate) UnmarshalXML(decoder *xml.Decoder, start xml.StartElement) (err error) {
	var value string

	err = decoder.DecodeElement(&value, &start)
	if err != nil {
		return errors.New("unable to decode XML")
	}
	parsedTime, err := time.Parse(dateOnly, value)
	if err != nil {
		return errors.New("unable to unmarshal XML")
	}
	ld.Time = parsedTime
	return nil
}

func (ld LocalDate) MarshalJSON() (array []byte, err error) {
	s := ld.Format(dateOnly)
	array = []byte(strconv.Quote(s))
	return array, nil
}

func (ld LocalDate) MarshalXML(encoder *xml.Encoder, start xml.StartElement) error {
	return encoder.EncodeElement(ld.Format(dateOnly), start)
}

func (ldt *LocalDateTime) UnmarshalJSON(b []byte) (err error) {
	s := string(b)

	s, _ = strconv.Unquote(s)
	t, err := time.Parse(dateTimeWithoutZone, s)
	if err != nil {
		return errors.New("unable to unmarshal JSON")
	}
	ldt.Time = t
	return nil
}

func (ldt *LocalDateTime) UnmarshalXML(decoder *xml.Decoder, start xml.StartElement) (err error) {
	var value string

	err = decoder.DecodeElement(&value, &start)
	if err != nil {
		return errors.New("unable to decode XML")
	}
	parsedTime, err := time.Parse(dateTimeWithoutZone, value)
	if err != nil {
		return errors.New("unable to unmarshal XML")
	}
	ldt.Time = parsedTime
	return nil
}

func (ldt LocalDateTime) MarshalJSON() (array []byte, err error) {
	s := ldt.Format(dateTimeWithoutZone)
	array = []byte(strconv.Quote(s))
	return array, nil
}

func (ldt LocalDateTime) MarshalXML(encoder *xml.Encoder, start xml.StartElement) error {
	return encoder.EncodeElement(ldt.Format(dateTimeWithoutZone), start)
}

`)

func contains(array []string, s string) bool {
	for i := range array {
		if array[i] == s {
			return true
		}
	}
	return false
}

type enumNormalized struct {
	// Value original value of enum
	Value string
	// NameNormalized normalized name to be able to generate some enum names like 'ENUM_1_3_158_00165387_100_40_50' in Golang.
	// In 99 % will be the same as value.
	NameNormalized string
}

func processEnums(enums []interface{}) ([]enumNormalized, error) {
	enumsNormalized := make([]enumNormalized, 0, len(enums))
	for _, enumItem := range enums {
		enumName, ok := enumItem.(string)
		if !ok {
			slog.Error("unable to retype enum name to string", "enum", enumItem)
			return nil, errors.New("unable to retype enum name to string")
		}
		originalName := enumName

		// enum starts with digit, we need to add prefix to be able to use it in Golang
		if unicode.IsDigit(rune(enumName[0])) {
			enumName = "ENUM_" + enumName
		}
		enumsNormalized = append(enumsNormalized, enumNormalized{
			Value:          originalName,
			NameNormalized: strings.ReplaceAll(enumName, ".", "_"), // enum contains dots, we need to replace it to be able to use it in Golang
		})
	}

	return enumsNormalized, nil
}
