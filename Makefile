
sdk_dep = github.com/aws/aws-sdk-go

help:
	@echo "Please use \`make <target>' where <target> is one of"
	@echo "  install                 to go install"
	@echo "  build                   to go build"
	@echo "  deps                    to go get dependencies"

install:
	@make build
	@go install

build:
	@make deps
	@go build

deps:
	@if [ ! -d "$(GOPATH)/src/$(sdk_dep)" ]; then git clone https://$(sdk_dep) $(GOPATH)/src/$(sdk_dep); fi
	@cd $(GOPATH)/src/$(sdk_dep) && git pull && go get ./... && go build && go install
