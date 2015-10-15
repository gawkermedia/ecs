
sdk_dest = github.com/aws/aws-sdk-go
sdk_source = github.com/gawkermedia/aws-sdk-go

help:
	@echo "Please use \`make <target>' where <target> is one of"
	@echo "  install                 to go install"
	@echo "  build                   to go build"
	@echo "  deps                    to go get dependencies"

install:
	@make build
	@go install

build:
	@go build

deps:
	@if [ ! -d "$(GOPATH)/src/$(sdk_dest)" ]; then git clone https://$(sdk_source) $(GOPATH)/src/$(sdk_dest); fi
	@cd $(GOPATH)/src/$(sdk_dest) && git pull && go build && go install
