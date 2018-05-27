
all: help

help: ## Display help text.
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

clean: ## Clean up build artifacts
	@echo "cleanup"
	@go clean ./...
	@-rm ./authsvc-cli ./authserver ./userserver

dot: ## Generate DOT graph of internal dependencies
	@echo "Generate deps.svg"
	@.bin/dependencies | dot -Tsvg >| deps.svg

deps: ## Ensure dependencies
	@echo "ensure dependencies"
	@dep ensure

build: ## Build binaries
	@echo "build artifacts"
	@.bin/build

format: ## Format Go code
	@echo "format code"
	@.bin/format

lint: ## Run linting tools on Go code
	@echo "run linting tools"
	@.bin/lint

verify: deps lint ## Verify all code quality checks pass
	@echo "verify code quality"
	@.bin/verify

install-hooks: ## Install git pre-push hooks to run verification checks before pushing
	@echo "install git hooks"
	@.bin/install-pre-push-hook

images: ## Build Docker images
	@echo "build docker images"
	@cd docker/authserver && $(MAKE) image
	@cd docker/userserver && $(MAKE) image

push-images: ## Push Docker images
	@echo "push docker images"
	@cd docker/authserver && $(MAKE) push
	@cd docker/userserver && $(MAKE) push


.PHONY: all help clean deps dot build format lint verify install-hooks images push-images
