ROOT_DIR := $(realpath $(firstword $(MAKEFILE_LIST)))
LOCALBIN := $(ROOT_DIR)/bin

OUTDIR := $(ROOT_DIR)/out

GO ?= go

GOLANG_CI ?= $(GO) run -modfile $(shell dirname $(ROOT_DIR))/hack/tools/golang-ci/go.mod github.com/golangci/golangci-lint/cmd/golangci-lint

## Local Folders
$(LOCALBIN):
	mkdir $(LOCALBIN)
$(OUTDIR):
	@mkdir $(OUTDIR)

.PHONY: clean
clean: ## Delete local folders.
	@-rm -r $(LOCALBIN)
	@-rm -r $(OUTDIR)

.PHONY: lint
lint: ## Run go linter.
	$(GOLANG_CI) run ./...

.PHONY: vet
vet: ## Run go vet against code.
	$(GO) vet ./...

.PHONY: tidy
tidy: ## Run go tidy against code.
	$(GO) mod tidy

.PHONY: fmt
fmt: ## Run go fmt against code.
	$(GO) fmt ./...
