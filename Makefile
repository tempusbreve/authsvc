
all: help

help: ## Display help text.
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

clean: ## Clean up build artifacts
	@echo "cleanup"
	@go clean ./...
	@-rm -vf ./authsvc-cli ./authsvc

dot: ## Generate DOT graph of internal dependencies
	@echo "Generate package_dependencies.svg"
	@.bin/dependencies | dot -Tsvg >| docs/package_dependencies.svg

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
	@cd docker/authsvc && $(MAKE) image

push-images: ## Push Docker images
	@echo "push docker images"
	@cd docker/authsvc && $(MAKE) push


.PHONY: all help clean deps dot build format lint verify install-hooks images push-images
