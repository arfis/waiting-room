install:
	go install ./cmd/service-generator

lint:
	golangci-lint run -v

gosec:
	gosec -confidence high -severity high ./...

test:
	go test -v ./...

download-linter:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

download-gosec:
	go install github.com/securego/gosec/v2/cmd/gosec@latest

.PHONY: install lint gosec test download-linter download-gosec
