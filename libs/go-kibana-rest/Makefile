TEST?=./...
PKG_NAME=kbapi
KIBANA_URL ?= http://127.0.0.1:5601
KIBANA_USERNAME ?= elastic
KIBANA_PASSWORD ?= changeme

all: help


test: fmt
	KIBANA_URL=${KIBANA_URL} KIBANA_USERNAME=${KIBANA_USERNAME} KIBANA_PASSWORD=${KIBANA_PASSWORD} go test $(TEST) -v -count 1 -parallel 1 -race -coverprofile=coverage.out -covermode=atomic $(TESTARGS) -timeout 120m

fmt:
	@echo "==> Fixing source code with gofmt..."
	gofmt -s -w ./



