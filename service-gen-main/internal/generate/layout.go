package generate

import (
	"gitlab.com/soluqa/bookio/service-generator/internal/utils"
)

func createBaseDirLayout(wd string) {
	for _, dir := range []string{
		"internal/context",
		"internal/data/dto",
		"internal/errors",
		"internal/rest/handler",
		"internal/rest/register",
		"internal/middleware",
		"internal/service",
		"migrations",
	} {
		utils.CreateDir(wd, dir)
	}
}
