package middleware

func Generate(wd string, module string) {
	generateCors(wd)
	generatePaging(wd, module)
	generateLogging(wd)
	generateRequestId(wd, module)
}
