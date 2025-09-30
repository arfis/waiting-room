package rest

import (
	"bytes"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"gitlab.com/soluqa/bookio/service-generator/internal/utils"
)

type Options struct {
	response         string
	responseArray    bool
	responseDto      bool
	request          string
	requestArray     bool
	requestPrimitive bool
	paged            bool
	pageOf           string
	pack             string
	operationId      string
	params           []utils.Param
	status           int
}

func generateHandler(buf *bytes.Buffer, options Options) {
	custom(buf, options)
}

func custom(buf *bytes.Buffer, options Options) {
	data := make(map[string]interface{})
	data["request"] = options.request
	data["requestArray"] = options.requestArray
	data["requestPrimitive"] = options.requestPrimitive
	data["response"] = options.response
	data["responseDto"] = options.responseDto
	data["responseArray"] = options.responseArray
	data["pageOf"] = options.pageOf
	data["service"] = options.pack
	data["status"] = options.status
	queryParams := make(map[string]string)
	pathParams := make(map[string]string)
	// check for query param duplicate names in path
	for _, val := range options.params {
		if val.ParamType == utils.Query {
			queryParams[val.Name] = getParamMappingMethod(val)
		} else if val.ParamType == utils.Path {
			pathParams[val.Name] = getParamMappingMethod(val)
		}
	}
	data["queryParams"] = queryParams
	data["pathParams"] = pathParams
	data["paged"] = options.paged
	if options.operationId != "" {
		data["operationId"] = options.operationId
	} else {
		slog.Error("missing operation id for request")
		os.Exit(1)
	}
	if err := customHandlerT.Execute(buf, data); err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}

func getParamMappingMethod(param utils.Param) string {
	if param.ParamType == utils.Path && !param.Required {
		slog.Error("path parameter must be set as required", "param", param.Name)
		os.Exit(1)
	}
	methodPrefix := strings.ToUpper(string(param.ParamType[:1])) + string(param.ParamType[1:])
	optionalPrefix := ""
	if !param.Required {
		optionalPrefix = "Optional"
	}

	paramTpl := `
	%s, applicationErr := %s(r, "%s")
	if applicationErr != nil {
		h.responseErrorHandler.HandleAndWriteError(w, r, applicationErr)
		return
	}`

	switch param.ValType {
	case "array:integer:int":
		fallthrough
	case "array:integer:int32":
		return fmt.Sprintf(paramTpl, strings.ToLower(param.Name[:1])+param.Name[1:], "handler."+methodPrefix+"ParamToArrayInt32", param.Name)
	case "array:integer:int64":
		return fmt.Sprintf(paramTpl, strings.ToLower(param.Name[:1])+param.Name[1:], "handler."+methodPrefix+"ParamToArrayInt64", param.Name)
	case "array:string:":
		return fmt.Sprintf("%s := %s(r, \"%s\")", strings.ToLower(param.Name[:1])+param.Name[1:], "handler."+methodPrefix+"ParamToArrayString", param.Name)
	case "integer:int":
		fallthrough
	case "integer:int32":
		return fmt.Sprintf(paramTpl, strings.ToLower(param.Name[:1])+param.Name[1:], "handler."+methodPrefix+optionalPrefix+"ParamToInt32", param.Name)
	case "integer:int64":
		return fmt.Sprintf(paramTpl, strings.ToLower(param.Name[:1])+param.Name[1:], "handler."+methodPrefix+optionalPrefix+"ParamToInt64", param.Name)
	case "boolean:":
		return fmt.Sprintf(paramTpl, strings.ToLower(param.Name[:1])+param.Name[1:], "handler."+methodPrefix+optionalPrefix+"ParamToBool", param.Name)
	case "string:date":
		return fmt.Sprintf(paramTpl, strings.ToLower(param.Name[:1])+param.Name[1:], "handler."+methodPrefix+optionalPrefix+"ParamToLocalDate", param.Name)
	case "string:date-time":
		return fmt.Sprintf(paramTpl, strings.ToLower(param.Name[:1])+param.Name[1:], "handler."+methodPrefix+optionalPrefix+"ParamToDateTime", param.Name)
	case "string:":
		fallthrough
	default:
		return fmt.Sprintf("%s := %s(r, \"%s\")", strings.ToLower(param.Name[:1])+param.Name[1:], "handler."+methodPrefix+optionalPrefix+"ParamToString", param.Name)
	}
}

var customHandlerT = utils.CreateTemplate(`
func (h *Handler) {{.operationId}}(w http.ResponseWriter, r *http.Request) {
	var applicationErr error
{{- range .pathParams }}
	{{.}}
{{- end}}
{{- range .queryParams }}
	{{.}}
{{- end}}
{{- if .paged }}
	{{- if .request }}
		{{- if .requestArray }}
			req := []dto.{{.request}}{}
		{{- else }}
			req := dto.{{.request}}{}
	    {{- end }}
	    applicationErr = json.NewDecoder(r.Body).Decode(&req)
        if applicationErr != nil {
            h.responseErrorHandler.HandleAndWriteError(w, r, ngErrors.New(ngErrors.InternalServerErrorCode, "problem decoding request body", http.StatusInternalServerError, nil))
            return
        }
	    {{- if.requestArray }}
		    for _, item := range req {
			   applicationErr = handler.GetValidator().Struct(item)
               if applicationErr != nil {
                   h.responseErrorHandler.HandleAndWriteError(w, r, ngErrors.RequestValidation(applicationErr))
                   return
               }
		    }
	    {{- else }}
		    applicationErr = handler.GetValidator().Struct(req)
            if applicationErr != nil {
                h.responseErrorHandler.HandleAndWriteError(w, r, ngErrors.RequestValidation(applicationErr))
                return
            }
	    {{- end }}
	{{- end }}
	page, size := handler.GetPageParams(r)
	sort := handler.QueryParamToArrayString(r, "sort")
	var content []dto.{{.pageOf}}
	var total int64
	content, total, applicationErr = h.svc.{{.operationId}}(
		r.Context(),
		{{- if .request}} {{if .requestArray}} req, {{else}} &req, {{end}}{{end}}
	    {{- range $k,$v := .pathParams}}
	    {{$k}},
		{{- end}}
		{{- range $k,$v := .queryParams}}
		{{$k}},
		{{- end}}
		page,
		size,
		sort,
	)
	if applicationErr != nil {
		h.responseErrorHandler.HandleAndWriteError(w, r, applicationErr)
		return
	}
	resp := handler.CreatePage(content, page, size, total, sort)
	handler.WriteJson(r.Context(), w, {{.status}}, resp)
{{- else if and .responseArray .response }}
	var resp []{{if .responseDto}} dto. {{end}}{{.response}}
	{{- if .request }}
		{{- if .requestArray }}
			req := []dto.{{.request}}{}
		{{- else }}
			req := dto.{{.request}}{}
		{{- end }}
		applicationErr = json.NewDecoder(r.Body).Decode(&req)
        if applicationErr != nil {
            h.responseErrorHandler.HandleAndWriteError(w, r, ngErrors.New(ngErrors.InternalServerErrorCode, "problem decoding request body", http.StatusInternalServerError, nil))
            return
        }
		{{- if.requestArray }}
			for _, item := range req {
				applicationErr = handler.GetValidator().Struct(item)
                if applicationErr != nil {
                    h.responseErrorHandler.HandleAndWriteError(w, r, ngErrors.RequestValidation(applicationErr))
                    return
                }
			}
		{{- else }}
		    applicationErr = handler.GetValidator().Struct(req)
            if applicationErr != nil {
                h.responseErrorHandler.HandleAndWriteError(w, r, ngErrors.RequestValidation(applicationErr))
                return
            }
		{{- end }}
		resp, applicationErr = h.svc.{{.operationId}}( 
			r.Context(),
			{{- range $k,$v := .pathParams}}
			{{$k}},
			{{- end}}
			{{- range $k,$v := .queryParams}}
			{{$k}},
			{{- end}}
			{{if .request}} {{if .requestArray}} req, {{else}} &req, {{end}}{{end}}
		)
	{{- else }}
		resp, applicationErr = h.svc.{{.operationId}}( 
			r.Context(),
			{{- if .isProtected }}
			r.Header.Get("Authorization"),
			{{- end }}
			{{- range $k,$v := .pathParams}}
			{{$k}},
			{{- end}}
			{{- range $k,$v := .queryParams}}
			{{$k}},
			{{- end}}
		)
	{{- end }}
	if applicationErr != nil {
		h.responseErrorHandler.HandleAndWriteError(w, r, applicationErr)
		return
	}
	handler.WriteJson(r.Context(), w, {{.status}}, resp)
{{- else}}
	{{- if .request }}
		{{- if .requestPrimitive }}
			{{- if .requestArray }}
				var req []{{.request}}
			{{- else }}
				var req {{.request}}
			{{- end }}
			applicationErr = json.NewDecoder(r.Body).Decode(&req)
            if applicationErr != nil {
                h.responseErrorHandler.HandleAndWriteError(w, r, ngErrors.New(ngErrors.InternalServerErrorCode, "problem decoding request body", http.StatusInternalServerError, nil))
                return
            }
		{{- else }}	
			{{- if .requestArray }}
				req := []dto.{{.request}}{}
			{{- else }}
				req := dto.{{.request}}{}
			{{- end }}
			applicationErr = json.NewDecoder(r.Body).Decode(&req)
            if applicationErr != nil {
                h.responseErrorHandler.HandleAndWriteError(w, r, ngErrors.New(ngErrors.InternalServerErrorCode, "problem decoding request body", http.StatusInternalServerError, nil))
                return
            }
			{{- if.requestArray }}
				for _, item := range req {
					applicationErr = handler.GetValidator().Struct(item)
                    if applicationErr != nil {
                        h.responseErrorHandler.HandleAndWriteError(w, r, ngErrors.RequestValidation(applicationErr))
                        return
                    }
				}
			{{- else }}
			    applicationErr = handler.GetValidator().Struct(req)
                if applicationErr != nil {
                    h.responseErrorHandler.HandleAndWriteError(w, r, ngErrors.RequestValidation(applicationErr))
                    return
                }
			{{- end }}
		{{- end }}
		{{- if .response }}
		  	var resp *{{if .responseDto}} dto. {{end}}{{.response}}
			resp, applicationErr = h.svc.{{.operationId}}(
				r.Context(),
				{{- range $k,$v := .pathParams}}
				{{$k}},
				{{- end}}
				{{- range $k,$v := .queryParams}}
				{{$k}},
				{{- end}}
				{{- if .request}}{{if .requestArray }}req{{else}}&req{{end}},{{end}}
			)
			if applicationErr != nil {
				h.responseErrorHandler.HandleAndWriteError(w, r, applicationErr)
			return
			}
			handler.WriteJson(r.Context(), w, {{.status}}, resp)
		{{- else }}
			applicationErr = h.svc.{{.operationId}}(
        r.Context(),
				{{- range $k,$v := .pathParams}}
				{{$k}},
				{{- end}}
				{{- range $k,$v := .queryParams}}
				{{$k}},
				{{- end}}
				{{if .request}}{{if .requestArray }}req{{else}}&req{{end}}, {{end}}
			)
			if applicationErr != nil {
				h.responseErrorHandler.HandleAndWriteError(w, r, applicationErr)
			return
			}
			w.WriteHeader({{.status}})
		{{- end }}
	{{- else }}
		{{- if .response }}
		  	var resp *{{if .responseDto}} dto. {{end}}{{.response}}
			resp, applicationErr = h.svc.{{.operationId}}(
				r.Context(),
				{{- range $k,$v := .pathParams}}
				{{$k}},
				{{- end}}
				{{- range $k,$v := .queryParams}}
				{{$k}},
				{{- end}}
			)
			if applicationErr != nil {
				h.responseErrorHandler.HandleAndWriteError(w, r, applicationErr)
			return
			}
			handler.WriteJson(r.Context(), w, {{.status}}, resp)
		{{- else }}
			applicationErr = h.svc.{{.operationId}}(
				r.Context(),
				{{- range $k,$v := .pathParams}}
				{{$k}},
				{{- end}}
				{{- range $k,$v := .queryParams}}
				{{$k}},
				{{- end}}
			)
			if applicationErr != nil {
				h.responseErrorHandler.HandleAndWriteError(w, r, applicationErr)
				return
			}
			w.WriteHeader({{.status}})
		{{- end }}
	{{- end }}
{{- end}}
}
`)
