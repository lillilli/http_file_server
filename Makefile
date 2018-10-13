SERVICE_NAME   := http_file_server
PKG            := github.com/lillilli/http_file_server
PKG_LIST       := $(shell go list ${PKG}/... | grep -v /vendor/)
CONFIG         := $(wildcard local.yml)
NAMESPACE	   := "default"

.PHONY: clean test

.PHONY: all
all: setup test build

.PHONY: setup
setup: ## Installing all service dependencies.
	@echo "Setup..."
	vgo mod vendor

.PHONY: config
config: ## Creating the local config yml.
	@echo "Creating local config yml ..."
	cp config.example.yml local.yml

.PHONY: build
build: ## Build the executable file of service.
	@echo "Building..."
	cd cmd/$(SERVICE_NAME) && go build

.PHONY: run
run: build ## Run service with local config.
	@echo "Running..."
	cd cmd/$(SERVICE_NAME) && ./$(SERVICE_NAME) -config=../../local.yml

image: ## Build a docker image.
	@echo "Docker image building..."
	$Q docker build -t $(SERVICE_NAME) .

run\:image: ## Run a docker image.
	@echo "Running docker image..."
	docker run -it -p 8080:8081 $(SERVICE_NAME) ./http_file_server -config local.yml

.PHONY: clean
clean: ## Cleans the temp files and etc.
	@echo "Clean..."
	rm -f cmd/$(SERVICE_NAME)/$(SERVICE_NAME)

help: ## Display this help screen
	@grep -E '^[a-zA-Z_-\:]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ": .*?## "}; {gsub(/[\\]*/,""); printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
