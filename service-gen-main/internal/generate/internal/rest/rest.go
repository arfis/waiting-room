package rest

import (
	"bytes"
	"fmt"
	"log/slog"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"gitlab.com/soluqa/bookio/service-generator/internal/utils"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type Path struct {
	Method   string
	Path     string
	Handler  string
	Package  string
	IsPaged  bool
	isPublic bool
}

type Handler struct {
	buf       *bytes.Buffer
	hasParams bool
	hasDTO    bool
	hasReq    bool
}

func Generate(wd string, api *openapi3.T, module string) {
	generateRestUtils(wd, module)
	data := make(map[string]any)
	data["module"] = module
	var handlerImports []string
	var publicPaths []Path
	var protectedPaths []Path
	handlers := make(map[string]*Handler)
	sortedPaths := make([]string, len(api.Paths))
	i := 0
	for s := range api.Paths {
		sortedPaths[i] = s
		i++
	}
	sort.Strings(sortedPaths)
	for _, p := range sortedPaths {
		path := api.Paths[p]
		operations := path.Operations()
		sortedOperations := make([]string, len(operations))
		io := 0
		for s := range operations {
			sortedOperations[io] = s
			io++
		}
		pathParams := utils.GetParams(path.Parameters)
		sort.Strings(sortedOperations)
		for _, o := range sortedOperations {
			extensionMap, err := utils.GetExtensionMap(operations[o].Extensions, "x-generated")
			if err != nil {
				slog.Error(err.Error())
				os.Exit(1)
			}
			if extensionMap["package"] == nil {
				slog.Error("package missing in", "endpoint", p)
				os.Exit(1)
			}
			newPath := Path{
				Method: fmt.Sprintf("%s%s", strings.ToUpper(o[:1]), strings.ToLower(o[1:])), // format to Get/Post/...
				Path:   p,
			}
			var handler bool
			var pack string
			var isPublic bool

			if extensionMap["customHandler"] != nil {
				handler = extensionMap["customHandler"].(bool)
			}
			if extensionMap["package"] != nil {
				pack = extensionMap["package"].(string)
			}

			if extensionMap["isPublic"] != nil {
				isPublic = extensionMap["isPublic"].(bool)
			}

			if pack == "" {
				slog.Error("package in paths.operation.x-generated is required")
				os.Exit(1)
			}

			newPath.Package = pack
			newPath.isPublic = isPublic

			var responses []int
			for key := range operations[o].Responses {
				intVar, err := strconv.Atoi(key)
				if err == nil {
					responses = append(responses, intVar)
				}
			}
			sort.Ints(responses[:])

			if responses == nil || len(responses) == 0 {
				slog.Error("no response code")
				os.Exit(1)
			} else if len(responses) > 1 {
				slog.Info("found more response codes, we will generate", "num", responses[0])
			}

			// handle custom handlers defined by x-generated.handler
			mediaTypeJson := operations[o].Responses.Get(responses[0]).Value.Content["application/json"]
			var requestMediaTypeJson *openapi3.MediaType
			if operations[o].RequestBody != nil && operations[o].RequestBody.Value != nil {
				requestMediaTypeJson = operations[o].RequestBody.Value.Content["application/json"]
			}
			if mediaTypeJson == nil {
				mediaTypeJson = operations[o].Responses.Get(responses[0]).Value.Content["application/json;charset=UTF-8"]
			}
			if operations[o].OperationID != "" {
				newPath.Handler = fmt.Sprintf("%sHandler.%s", pack, operations[o].OperationID)
			} else {
				slog.Error("missing operation-id property", "operation", o, "property", p)
			}
			if handler == true {
				if !utils.ContainsString(handlerImports, pack) {
					handlerImports = append(handlerImports, pack)
				}
				if mediaTypeJson != nil {
					ref := mediaTypeJson.Schema

					isPage, _ := utils.IsPage(ref)
					if isPage {
						newPath.IsPaged = true
					}
				}

				if isPublic {
					publicPaths = append(publicPaths, newPath)
				} else {
					protectedPaths = append(protectedPaths, newPath)
				}
				newPath.Handler = fmt.Sprintf("%sHandler.%s", pack, operations[o].OperationID)
			} else {
				req := operations[o].RequestBody
				// generate handler
				if _, nok := handlers[pack]; !nok {
					handlers[pack] = &Handler{
						buf:       &bytes.Buffer{},
						hasParams: false,
						hasDTO:    false,
					}
				}
				if req != nil {
					handlers[pack].hasReq = true
				}
				hndlr := handlers[pack]
				options := Options{
					paged:       false,
					pack:        pack,
					status:      responses[0],
					operationId: utils.ToPublic(operations[o].OperationID),
				}
				if requestMediaTypeJson != nil {
					hndlr.hasDTO = true
				}
				if mediaTypeJson != nil {
					ref := mediaTypeJson.Schema

					isPage, pageOfDTO := utils.IsPage(ref)
					if isPage {
						options.paged = true
						newPath.IsPaged = true
						options.pageOf = pageOfDTO
						hndlr.hasDTO = true
					} else if ref.Value.Type == "array" {
						if ref.Value.Items != nil {
							options.responseArray = true
							if ref.Value.Items.Ref != "" {
								options.response = ref.Value.Items.Ref[len("#/component/schemas/_"):]
								options.responseDto = true
								hndlr.hasDTO = true
							} else {
								if ref.Value.Items.Value.Format != "" {
									options.response = ref.Value.Items.Value.Format
								} else {
									options.response = ref.Value.Items.Value.Type
								}
								options.responseDto = false
							}

						} else {
							slog.Error("response is type of array but has no items.")
							os.Exit(1)
						}
					} else {
						if ref.Ref != "" {
							options.response = ref.Ref[len("#/component/schemas/_"):]
							options.responseDto = true
							hndlr.hasDTO = true
						} else {
							if ref.Value.Type == "object" && ref.Value.AdditionalProperties.Schema != nil {
								options.response = getMapType(ref.Value)
							} else if ref.Value.Format != "" {
								options.response = ref.Value.Format
							} else {
								options.response = ref.Value.Type
							}
							options.responseDto = false
						}
					}
				}
				if req != nil {
					if req.Value.Content["application/json"] == nil {
						slog.Error("unsupported content type", "endpoint", p)
						os.Exit(1)
					}
					schema := req.Value.Content["application/json"].Schema
					if schema.Value.Type == "array" {
						options.requestArray = true
						if schema.Value.Items.Ref != "" {
							options.request = schema.Value.Items.Ref[len("#/component/schemas/_"):]
						} else { // It's a primitive type
							options.requestPrimitive = true
							if schema.Value.Items.Value.Format != "" {
								options.request = schema.Value.Items.Value.Format
							} else {
								options.response = schema.Value.Items.Value.Type
							}
						}
					} else {
						options.request = schema.Ref[len("#/component/schemas/_"):]
					}
				}

				options.params = utils.GetParams(operations[o].Parameters)
				if len(options.params) > 0 {
					hndlr.hasParams = true
				}
				for _, param := range pathParams {
					if !utils.ContainsParam(options.params, param) {
						options.params = append(options.params, param)
					} else {
						slog.Error("multiple occurences", "parameter", param.Name, "operation", o, "path", p)
						os.Exit(1)
					}
				}
				if !utils.ContainsString(handlerImports, pack) {
					handlerImports = append(handlerImports, pack)
				}

				if isPublic {
					publicPaths = append(publicPaths, newPath)
				} else {
					protectedPaths = append(protectedPaths, newPath)
				}
				generateHandler(hndlr.buf, options)
			}
		}
	}

	data["publicPaths"] = publicPaths
	data["protectedPaths"] = protectedPaths
	data["handlerImports"] = handlerImports
	data["needMiddleware"] = false

	var needMiddleware bool
	allPaths := append(publicPaths, protectedPaths...)
	for _, pp := range allPaths {
		if pp.IsPaged {
			needMiddleware = true
			break
		}
	}

	if len(protectedPaths) > 0 {
		needMiddleware = true
	}

	data["needMiddleware"] = needMiddleware

	var buf bytes.Buffer
	err := registerGenerated.Execute(&buf, data)
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
	utils.ForceSave(fmt.Sprintf("%s/internal/rest/register/generated.go", wd), buf.Bytes())

	// save handlers
	for p := range handlers {
		var buf bytes.Buffer
		generateHandlerImports(&buf, handlers[p], p, module)
		buffer := append(buf.Bytes(), handlers[p].buf.Bytes()...)
		utils.CreateDir(wd, fmt.Sprintf("internal/rest/handler/%s", p))
		utils.ForceSave(fmt.Sprintf("%s/internal/rest/handler/%s/generated.go", wd, p), buffer)
	}
}

func getMapType(value *openapi3.Schema) (name string) {
	var fieldBuilder bytes.Buffer
	for value.AdditionalProperties.Schema != nil {
		fieldBuilder.WriteString("map[string]")
		value = value.AdditionalProperties.Schema.Value
	}
	fieldBuilder.WriteString(getPrimitiveType(value))

	return fieldBuilder.String()
}

func getPrimitiveType(value *openapi3.Schema) (name string) {
	switch value.Type {
	case "array":
		switch value.Items.Value.Type {
		case "number":
			fallthrough
		case "integer":
			return "[]" + value.Items.Value.Format
		case "boolean":
			return "[]bool"
		case "string":
			return "[]string"
		default:
			return "[]any"
		}
	case "number":
		fallthrough
	case "integer":
		return value.Format
	case "boolean":
		return "bool"
	case "string":
		return "string"
	default:
		return "any"
	}
}

func generateRestUtils(wd string, module string) {
	data := map[string]string{"Module": module}
	var buf bytes.Buffer
	err := utilsT.Execute(&buf, data)
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
	utils.ForceSave(fmt.Sprintf("%s/internal/rest/handler/utils.go", wd), buf.Bytes())
}

func generateHandlerImports(buf *bytes.Buffer, handler *Handler, pack string, module string) {
	data := make(map[string]any)
	data["package"] = pack
	data["handlerName"] = fmt.Sprintf("%sHandler", cases.Title(language.English).String(pack))
	data["module"] = module
	data["hasDTO"] = handler.hasDTO
	data["hasParams"] = handler.hasParams
	data["hasRequest"] = handler.hasReq
	err := handlerImports.Execute(buf, data)
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}

var registerGenerated = utils.CreateTemplate(`
// Code generated by go generate; DO NOT EDIT.
package register

import (
	"github.com/go-chi/chi/v5"
	"{{.module}}/internal/middleware"
{{- range .handlerImports}}
	"{{$.module}}/internal/rest/handler/{{.}}"
{{- end}}
	"go.uber.org/dig"
)

func Generated(r chi.Router, diContainer *dig.Container) {
	err := diContainer.Invoke(func(
{{- range .handlerImports}}
		{{.}}Handler *{{.}}.Handler,
{{- end}}
{{- if .needMiddleware }}
		authorizationMiddleware *middleware.AuthorizationMiddleware,
{{- end }}
	) error {
{{- if .publicPaths}}
		// Public routes (no JWT needed)
{{- range .publicPaths}}
		r{{if .IsPaged}}.With(pagingMiddleware.Paging){{end}}.{{.Method}}("{{.Path}}", {{.Handler}})
{{- end}}
{{- end}}

{{- if .protectedPaths}}

		// Protected routes (require JWT)
		r.With(authorizationMiddleware.Middleware()).Group(func(protected chi.Router) {
{{- range .protectedPaths}}
			protected{{if .IsPaged}}.With(pagingMiddleware.Paging){{end}}.{{.Method}}("{{.Path}}", {{.Handler}})
{{- end}}

		})
{{- end}}

		return nil
	})

	if err != nil {
		panic(err)
	}
}`)

var handlerImports = utils.CreateTemplate(`
// Code generated by go generate; DO NOT EDIT.
package {{.package}}

import (
	"net/http"
{{- if .hasDTO }}
	"{{.module}}/internal/data/dto"
{{- end }}
{{- if .hasRequest }}
	"encoding/json"
{{- end }}
{{- if or .hasDTO .hasParams }}
	"{{.module}}/internal/rest/handler"
{{- end }}
	ngErrors "{{.module}}/internal/errors"
	"{{.module}}/internal/service/{{.package}}"
)

type Handler struct {
    svc *{{.package}}.Service
    responseErrorHandler *ngErrors.ResponseErrorHandler
}

func New(
	svc *{{.package}}.Service,
	responseErrorHandler *ngErrors.ResponseErrorHandler,
) *Handler {
	return &Handler{
		svc: svc,
		responseErrorHandler: responseErrorHandler,
	}
}

`)

var utilsT = utils.CreateTemplate(`
// Code generated by go generate; DO NOT EDIT.
package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"reflect"
	"strconv"
	"time"

	"{{.Module}}/internal/data/dto"
	"{{.Module}}/internal/errors"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
)

var (
	validate = validator.New()
)

func GetValidator() *validator.Validate {
	return validate
}

func WriteJson(c context.Context, w http.ResponseWriter, status int, dto any) error {
	b, err := json.Marshal(dto)
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_, err = w.Write(b)
	return err
}

func PathParamToString(r *http.Request, param string) string {
	s := chi.URLParam(r, param)
	return s
}

func PathParamToInt64(r *http.Request, param string) (int64, error) {
	parseInt, err := strconv.ParseInt(chi.URLParam(r, param), 10, 64)
	if err != nil {
		return 0, errors.New(errors.BusinessErrorCode, fmt.Sprintf("%s could not be parsed", param), http.StatusBadRequest, nil)
	}
	return parseInt, nil
}

func PathParamToBool(r *http.Request, param string) (bool, error) {
	parseBool, err := strconv.ParseBool(chi.URLParam(r, param))
	if err != nil {
		return false, errors.New(errors.BusinessErrorCode, fmt.Sprintf("%s could not be parsed", param), http.StatusBadRequest, nil)
	}
	return parseBool, nil
}

func PathParamToDateTime(r *http.Request, param string) (time.Time, error) {
	parsedDateTime, err := time.Parse(time.RFC3339, param)
	if err != nil {
		return time.Time{}, errors.New(errors.BusinessErrorCode, fmt.Sprintf("%s could not be parsed", param), http.StatusBadRequest, nil)
	}
	return parsedDateTime, nil
}

func PathParamToInt32(r *http.Request, param string) (int32, error) {
	parseInt, err := strconv.ParseInt(chi.URLParam(r, param), 10, 32)
	if err != nil {
		return 0, errors.New(errors.BusinessErrorCode, fmt.Sprintf("%s could not be parsed", param), http.StatusBadRequest, nil)
	}
	return int32(parseInt), nil
}

func QueryParamToString(r *http.Request, param string) string {
	return r.URL.Query().Get(param)
}

func QueryOptionalParamToString(r *http.Request, param string) *string {
	val := r.URL.Query().Get(param)
    if val == "" {
		return nil
	}
	return &val
}

func QueryParamToArrayString(r *http.Request, param string) []string {
	return r.URL.Query()[param]
}

func QueryParamToBool(r *http.Request, param string) (bool, error) {
	parseBool, err := strconv.ParseBool(r.URL.Query().Get(param))
	if err != nil {
		return false, errors.New(errors.BusinessErrorCode, fmt.Sprintf("%s could not be parsed", param), http.StatusBadRequest, nil)
	}
	return parseBool, nil
}

func QueryOptionalParamToBool(r *http.Request, param string) (*bool, error) {
	s := r.URL.Query().Get(param)
	if s != "" {
		parseBool, err := strconv.ParseBool(s)
		if err != nil {
			return nil, errors.New(errors.BusinessErrorCode, fmt.Sprintf("%s could not be parsed", param), http.StatusBadRequest, nil)
		}
		return &parseBool, nil
	}
	return nil, nil
}

func QueryOptionalParamToDateTime(r *http.Request, param string) (*time.Time, error) {
	s := r.URL.Query().Get(param)
	if s != "" {
		date, err := time.Parse(time.RFC3339, s)
		if err != nil {
			return nil, errors.New(errors.BusinessErrorCode, fmt.Sprintf("%s could not be parsed", param), http.StatusBadRequest, nil)
		}
		return &date, nil
	}
	return nil, nil
}

func QueryParamToDateTime(r *http.Request, param string) (time.Time, error) {
	date, err := time.Parse(time.RFC3339, r.URL.Query().Get(param))
	if err != nil {
		return time.Time{}, errors.New(errors.BusinessErrorCode, fmt.Sprintf("%s could not be parsed", param), http.StatusBadRequest, nil)
	}
	return date, nil
}

func QueryOptionalParamToLocalDate(r *http.Request, param string) (*dto.LocalDate, error) {
	s := r.URL.Query().Get(param)
	if s != "" {
		var localDate dto.LocalDate
		date, err := time.Parse("2006-01-02", s)
		if err != nil {
			return nil, errors.New(errors.BusinessErrorCode, fmt.Sprintf("%s could not be parsed", param), http.StatusBadRequest, nil)
		}
		localDate.Time = date
		return &localDate, nil
	}
	return nil, nil
}

func QueryParamToLocalDate(r *http.Request, param string) (dto.LocalDate, error) {
	var localDate dto.LocalDate
	date, err := time.Parse("2006-01-02", r.URL.Query().Get(param))
	if err != nil {
		return localDate, errors.New(errors.BusinessErrorCode, fmt.Sprintf("%s could not be parsed", param), http.StatusBadRequest, nil)
	}
    localDate.Time = date
	return localDate, nil
}

func QueryOptionalParamToInt32(r *http.Request, param string) (*int32, error) {
	s := r.URL.Query().Get(param)
	if s != "" {
		parseInt, err := strconv.ParseInt(s, 10, 32)
		if err != nil {
			return nil, errors.New(errors.BusinessErrorCode, fmt.Sprintf("%s could not be parsed", param), http.StatusBadRequest, nil)
		}
		result := int32(parseInt)
		return &result, nil
	}
	return nil, nil
}

func QueryParamToInt32(r *http.Request, param string) (int32, error) {
	s := r.URL.Query().Get(param)
	parseInt, err := strconv.ParseInt(s, 10, 32)
	if err != nil {
		return 0, errors.New(errors.BusinessErrorCode, fmt.Sprintf("%s could not be parsed", param), http.StatusBadRequest, nil)
	}
	return int32(parseInt), nil
}

func QueryParamToArrayInt32(r *http.Request, param string) ([]int32, error) {
	arr := r.URL.Query()[param]
	res := make([]int32, len(arr))
	for i := range arr {
		parseInt, err := strconv.ParseInt(arr[i], 10, 32)
		if err != nil {
			return nil, errors.New(errors.BusinessErrorCode, fmt.Sprintf("%s could not be parsed", param), http.StatusBadRequest, nil)
		}
		res[i] = int32(parseInt)
	}
	return res, nil
}

func QueryParamToInt64(r *http.Request, param string) (int64, error) {
	s := r.URL.Query().Get(param)
	parseInt, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, errors.New(errors.BusinessErrorCode, fmt.Sprintf("%s could not be parsed", param), http.StatusBadRequest, nil)
	}
	return parseInt, nil
}

func QueryOptionalParamToInt64(r *http.Request, param string) (*int64, error) {
	s := r.URL.Query().Get(param)
	if s != "" {
		parseInt, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return nil, errors.New(errors.BusinessErrorCode, fmt.Sprintf("%s could not be parsed", param), http.StatusBadRequest, nil)
		}
		return &parseInt, nil
	}
	return nil, nil
}

func QueryParamToArrayInt64(r *http.Request, param string) ([]int64, error) {
	arr := r.URL.Query()[param]
	res := make([]int64, len(arr))
	for i := range arr {
		parseInt, err := strconv.ParseInt(arr[i], 10, 64)
		if err != nil {
			return nil, errors.New(errors.BusinessErrorCode, fmt.Sprintf("%s could not be parsed", param), http.StatusBadRequest, nil)
		}
		res[i] = parseInt
	}
	return res, nil
}

// GetVersion returns version from requests query parameter version
func GetVersion(r *http.Request) sql.NullInt64 {
	v, err := strconv.ParseInt(QueryParamToString(r, "version"), 10, 64)
	version := sql.NullInt64{
		Int64: v,
		Valid: err == nil,
	}
	return version
}

func GetPageParams(r *http.Request) (page int32, size int32) {
	p, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil || p < 0 {
		p = 0
	}
	s, err := strconv.Atoi(r.URL.Query().Get("size"))
	if err != nil || s < 0 {
		s = 20
	}
	return int32(p), int32(s)
}

func CreatePage(content any, page int32, size int32, totalElements int64, sort []string) *dto.Page {
	contentType := reflect.ValueOf(content)
	var numberOfElements int
	if contentType.Kind() == reflect.Pointer {
		contentType = contentType.Elem()
	}
	if contentType.Kind() == reflect.Slice || contentType.Kind() == reflect.Array {
		numberOfElements = contentType.Len()
	}
	totalPages := int32(math.Ceil(float64(totalElements) / float64(size)))
	offset := page * size
	return &dto.Page{
		Pageable: dto.Pageable{
			Page:   page,
			Size:   size,
			Offset: offset,
		},
		Sort: dto.Sort{
			Empty:    len(sort) == 0,
			Sorted:   len(sort) != 0,
			Unsorted: len(sort) == 0,
		},
		Content:          content,
		Empty:            numberOfElements == 0,
		First:            page == 0,
		Last:             int32(page)+1 >= totalPages,
		NumberOfElements: int32(numberOfElements),
		Size:             int32(size),
		TotalElements:    totalElements,
		TotalPages:       totalPages,
	}
}
`)
