COMMIT		?= $(shell basename `git rev-list -1 HEAD`)
DATE		?= $(shell basename `date +%m-%d-%Y`)
SERVICE		?= $(shell basename `go list`)
VERSION		?= $(shell git describe --tags --always --dirty --match=v* 2> /dev/null || cat $(PWD)/.version 2> /dev/null || echo v0)
PACKAGE		?= $(shell go list)
PACKAGES	?= $(shell go list ./...)
LOCAL_PATH	?= $(shell dirname "${BASH_SOURCE[0]}")
GROUP		?= 'prometheusserver'
API_VERSION	?= 'v1alpha1'
LD_FLAGS=-ldflags " \
    -X github.com/marcosQuesada/prometheus-operator/pkg/config.Commit=$(COMMIT) \
    -X github.com/marcosQuesada/prometheus-operator/pkg/config.Date=$(DATE) \
    "

.PHONY: help fmt vet vendor tidy env test test-cover test-all generate-crd build build-docker external all

default: help

help:   ## show this help
	@echo 'usage: make [target] ...'
	@echo ''
	@echo 'targets:'
	@egrep '^(.+)\:\ .*##\ (.+)' ${MAKEFILE_LIST} | sed 's/:.*##/#/' | column -t -c 2 -s '#'

all: fmt build test    ## format, build and unit test

vendor:    ## install dependencies
	go mod vendor

tidy:    ## tidy vendors
	go mod tidy

env:    ## Print useful environment variables to stdout
	echo $(COMMIT)
	echo $(DATE)
	echo $(LOCAL_PATH)
	echo $(GROUP)
	echo $(API_VERSION)
	echo $(SERVICE)
	echo $(VERSION)
	echo $(PACKAGE)
	echo $(FILES)

fmt:    ## format the go source files
	gofmt -w -s .
	goimports -w .

vet:    ## run go vet on the source files
	go vet ./...

generate-crd: vendor   ## generate API CRDs clientset, listers, informers and deepCopy. Note it will be generated on github folder
	chmod 755 vendor/k8s.io/code-generator/generate-groups.sh
	./vendor/k8s.io/code-generator/generate-groups.sh all \
		github.com/marcosQuesada/prometheus-operator/pkg/crd/generated \
		github.com/marcosQuesada/prometheus-operator/pkg/crd/apis \
		${GROUP}:${API_VERSION} \
		--go-header-file ./hack/boilerplate.go.txt \
		--output-base "$(LOCAL_PATH)/" \
		-v 10

build: vendor    ## Build binary
	CGO_ENABLED=0 go build $(LD_FLAGS) .

external:    ## Run controller locally on an external k8s client
	go run main.go external

test: vendor    ## Run test suite
	go test --race -v ./...

test-cover:     ## Run test coverage and generate html report
	rm -fr coverage
	mkdir coverage
	go list -f '{{if gt (len .TestGoFiles) 0}}"go test -covermode count -coverprofile {{.Name}}.coverprofile -coverpkg ./... {{.ImportPath}}"{{end}}' ./... | xargs -I {} bash -c {}
	echo "mode: count" > coverage/cover.out
	grep -h -v "^mode:" *.coverprofile >> "coverage/cover.out"
	rm *.coverprofile
	go tool cover -html=coverage/cover.out -o=coverage/cover.html

test-all: test test-cover    ## Run test suite and generate test coverage

# Build Docker image
build-docker: vendor    ## Build docker image
	docker build -t prometheus-operator . --build-arg COMMIT=$(COMMIT) --build-arg DATE=$(DATE)
