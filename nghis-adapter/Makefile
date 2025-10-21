include conf/nghis-adapter/.env
export


ifeq ("$(shell go env GOOS)","windows")
	GO_PATH := $(subst \,/,$(shell go env GOPATH))
else
	GO_PATH := $(GOPATH)
endif

RUN_ARGS=nghis-adapter

.DEFAULT_GOAL := generate
OUTPUT ?= __output/

$(info "GOPATH: $(shell go env GOPATH)")

generate:
	@service-gen -wd ./ -open-api ./open-api.yaml

run-dev:
	go run cmd/$(RUN_ARGS)/*.go

build:
	go build \
		-o ${OUTPUT} \
		./cmd/$(RUN_ARGS)

build-for-image:
	CGO_ENABLED=0 go build ./cmd/$(RUN_ARGS)

lint:
	golangci-lint run --timeout 5m

sec: 
	gosec -quiet -confidence high -severity high ./...

download-linter:
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest

download-gosec: 
	go install github.com/securego/gosec/v2/cmd/gosec@latest

download-all:
	go mod download
	make download-linter
	make download-gosec

.PHONY: test
test:
	go test -v ./... > test.log

test-all:
	go test -v ./...

bench-all:
	go test -run=XXX -bench=. -v ./...

test-report:
	gotestsum --junitfile report.xml --format testname 

.PHONY: build download-all download-linter download-migrator lint run-dev test-all bench-all test-report generate sec 
