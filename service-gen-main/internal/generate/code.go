package generate

import (
	"gitlab.com/soluqa/bookio/service-generator/internal/generate/internal/service"
	"gitlab.com/soluqa/bookio/service-generator/internal/generate/internal/utils"

	"github.com/getkin/kin-openapi/openapi3"
	"gitlab.com/soluqa/bookio/service-generator/internal/generate/internal/context"
	"gitlab.com/soluqa/bookio/service-generator/internal/generate/internal/data"
	"gitlab.com/soluqa/bookio/service-generator/internal/generate/internal/errors"
	"gitlab.com/soluqa/bookio/service-generator/internal/generate/internal/linter"
	"gitlab.com/soluqa/bookio/service-generator/internal/generate/internal/middleware"
	"gitlab.com/soluqa/bookio/service-generator/internal/generate/internal/rest"
)

func generateCode(wd string, api *openapi3.T, module string) {
	linter.Generate(wd)
	context.Generate(wd)
	data.GenerateDTO(wd, api, module)
	errors.Generate(wd, api)
	middleware.Generate(wd, module)
	rest.Generate(wd, api, module)
	service.Generate(wd, module)
	utils.Generate(wd, module)
}
