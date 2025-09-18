OPENAPI=openapi/waiting-room.yaml
PKG=api
OUT=internal/api

.PHONY: gen
gen:
	@oapi-codegen -generate types      -package $(PKG) -o $(OUT)/types.gen.go $(OPENAPI)
	@oapi-codegen -generate chi-server -package $(PKG) -o $(OUT)/server.gen.go $(OPENAPI)

.PHONY: run
run:
	go run ./cmd/api

.PHONY: tidy
tidy:
	go mod tidy

