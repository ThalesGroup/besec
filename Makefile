# Adapted from https://github.com/genuinetools/weather, MIT

SHELL = /bin/bash
# Ensure that errors in pipes in recipes cause the overall result to be a fail, and turn on extglob support
.SHELLFLAGS = -o pipefail -O extglob -c

SWAGGER = goswagger             # go-swagger binary
MD5SUM = md5sum | cut -d' ' -f1 # md5 checksum command, removing the filename

UNAME_S := $(shell uname -s)
ifeq ($(UNAME_S),Darwin)
    SWAGGER = swagger
    MD5SUM = md5 -r
endif

NAME = besec
PKG := github.com/ThalesGroup/$(NAME)

# Populate version variables
# Add to compile time flags
VERSION := $(shell cat VERSION.txt)
GITCOMMIT := $(shell git rev-parse --short HEAD)
GITUNTRACKEDCHANGES := $(shell git status --porcelain --untracked-files=no)
ifneq ($(GITUNTRACKEDCHANGES),)
	GITCOMMIT := $(GITCOMMIT)-dirty
endif
CTIMEVAR=-X $(PKG)/cmd.GITCOMMIT=$(GITCOMMIT) -X $(PKG)/cmd.VERSION=$(VERSION)
GO_FLAGS=-tags "${GO_TAGS}"

# Build the binary with statically linked C libraries, for running in an otherwise empty container
ifdef SCRATCH
    CGO=CGO_ENABLED=0
    GO_FLAGS+= -a
else
    CGO=
endif

.PHONY: debug
debug: GO_FLAGS += -gcflags="all=-N -l"
debug: GO_LDFLAGS=-ldflags "$(CTIMEVAR)"
debug: $(NAME)
.PHONY: release
release: GO_LDFLAGS=-ldflags "-w $(CTIMEVAR)"
release: $(NAME)

GO_FILES = $(shell find ./ -type f -name '*.go')
# ui/build is embedded into the Go binary
$(NAME): $(GO_FILES) ui/build api/generated_checksum
	@echo "+ $@"
	@$(CGO) go build ${GO_FLAGS} ${GO_LDFLAGS} -o $(NAME) .

.PHONY: all
all: release golangci-lint lint test
test: testgo testgo-integration testui

.PHONY: clean
clean:
	@echo "+ $@"
	@rm -f $(NAME) cmd/*_vfsdata.go
	@rm -rf ui/build

.PHONY: fmt
fmt: ## Verifies all files have been `gofmt`ed. Not included in the `all` target, as it's covered by golangci-lint
	@echo "+ $@"
	@! gofmt -s -l . | grep -vE 'api/(restapi|models)'

.phony: lint
lint: ## outputs any linter suggestions. suggestions don't fail the build.
	@echo "+ $@"
	@golint ./... > /dev/stderr

.phony: golangci-lint
# aggregates the output from many linters.
golangci-lint: $(NAME) # Doesn't strictly need the binary, but does need its dependent generated code
	@echo "+ $@"
	@golangci-lint --build-tags "${GO_TAGS}" run > /dev/stderr

.PHONY: testgo
testgo: $(NAME)
	@echo "+ $@"
	@go test ${GO_FLAGS} -v  $(shell go list ./...)

.PHONY: testgo-integration
testgo-integration: GO_TAGS+= integration
testgo-integration: $(NAME)
	@echo "+ $@"
	@go test ${GO_FLAGS} -v  $(shell go list ./...)

.PHONY: check
check: $(NAME)
	./$(NAME) check

ui/build: $(shell find ui/src -type f) ui/package.json $(shell find ui/public -type f) ui/src/client.ts
	@echo "+ $@"
	@cd ui && npm run build > /tmp/npm-build.out || cat /tmp/npm-build.out

.PHONY: testui
testui: ## Runs the React tests.
	@echo "+ $@"
	@cd ui && REACT_APP_API_HOST="http://notused/" npm test -- --all --watchAll=false

API_DEF_FILES = ./api/swagger.yaml ./practices/schema.json

# We generate a checksum over all of the generated files, to ensure that any change in the dependencies is reflected as a change in the make target.
# The go files in the prerequisites define some of the data-structures and serialization format.
# requires go-swagger to be installed locally
# Because this depends on modification times, a fresh checkout may lead Make to think this needs rebuilding. In this case, run ./set_modification_time.sh first.
api/generated_checksum: $(API_DEF_FILES) ./lib/practices.go ./lib/plan.go
	@if [[ -n "$(CI)" ]]; then echo -e "Error: it looks like we're running in CI but the generated go files aren't up to date.\nPlease re-run make locally, check in any generated files, and try again." > /dev/stderr && exit 1; fi
	@echo "+ generate API server"
	@$(SWAGGER) generate server --name=$(NAME) --exclude-main --principal github.com/ThalesGroup/besec/api/models.User --target api -f api/swagger.yaml > /dev/null 2>&1
	@echo "+ generate Go API client"
	@$(SWAGGER) generate client --name=$(NAME) --target api -f api/swagger.yaml > /dev/null 2>&1
	@find api/restapi api/models api/client $^ -type f  | sort | xargs cat | $(MD5SUM) > $@

# requires dotnet installed locally
ui/src/client.ts: $(API_DEF_FILES) ./ui/nswag.json ./ui/src/clientExtensions.ts
	@echo "+ generate UI API client"
	@rm -f $@ # nswag doesn't overwrite the client if there are no changes to be made
	@cd ui && npm run generate-client > /dev/null

.PHONY: generate-api
generate-api: ./api/generated_checksum ./ui/src/client.ts


.PHONY: clean-api
clean-api:
	rm -f api/generated_checksum
	rm -rf api/restapi/!(configure_*.go)
	rm -f api/models/!(user.go)
	rm -rf api/client
	rm -f ui/src/client.ts

.PHONY: cloc
cloc:
	cloc --vcs=git --exclude-list-file=.clocignore .

# Rebuild whenever a file changes in this directory tree
.PHONY: watch
watch:
	while true; do \
		make $(WATCHMAKE); \
        inotifywait -qre close_write .; \
	done
