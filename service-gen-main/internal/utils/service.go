package utils

import "github.com/getkin/kin-openapi/openapi3"

var PageQueryParams = []string{"page", "size", "sort"}

type ParamType string

const (
	Query ParamType = "query"
	Path  ParamType = "path"
)

type Param struct {
	ParamType ParamType
	ValType   string
	Name      string
	Required  bool
}

func ContainsParam(slice []Param, p Param) bool {
	for _, val := range slice {
		if val.Name == p.Name {
			return true
		}
	}
	return false
}

func GetParams(parameters openapi3.Parameters) []Param {
	var params []Param
	for i := range parameters {
		param := parameters[i]
		if param.Value.In == "query" && !ContainsString(PageQueryParams, param.Value.Name) {
			p := Param{
				ParamType: Query,
				ValType:   getParamType(param),
				Name:      param.Value.Name,
				Required:  param.Value.Required,
			}
			params = append(params, p)
		}
		if param.Value.In == "path" && !ContainsString(PageQueryParams, param.Value.Name) {
			p := Param{
				ParamType: Path,
				ValType:   getParamType(param),
				Name:      param.Value.Name,
				Required:  param.Value.Required,
			}
			params = append(params, p)
		}
	}
	return params
}

func getParamType(param *openapi3.ParameterRef) string {
	if param.Value.Schema.Value.Type != "array" {
		return param.Value.Schema.Value.Type + ":" + param.Value.Schema.Value.Format
	} else {
		return "array:" + param.Value.Schema.Value.Items.Value.Type + ":" + param.Value.Schema.Value.Items.Value.Format
	}
}
