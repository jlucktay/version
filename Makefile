# Inspiration:
# - https://devhints.io/makefile
# - https://tech.davis-hansson.com/p/make/
# - https://marmelab.com/blog/2016/02/29/auto-documented-makefile.html

# Default - top level rule is what gets run when you run just 'make' without specifying a goal/target.
.DEFAULT_GOAL := build

# Make will delete the target of a rule if it has changed and its recipe exits with a nonzero exit status, just as it
# does when it receives a signal.
.DELETE_ON_ERROR:

# When a target is built, all lines of the recipe will be given to a single invocation of the shell rather than each
# line being invoked separately.
.ONESHELL:

# If this variable is not set, the program '/bin/sh' is used as the shell.
SHELL := bash

# The default value of .SHELLFLAGS is -c normally, or -ec in POSIX-conforming mode.
# Extra options are set for Bash:
#   -e             Exit immediately if a command exits with a non-zero status.
#   -u             Treat unset variables as an error when substituting.
#   -o pipefail    The return value of a pipeline is the status of the last command to exit with a non-zero status,
#                  or zero if no command exited with a non-zero status.
.SHELLFLAGS := -euo pipefail -c

# Eliminate use of Make's built-in implicit rules.
MAKEFLAGS += --no-builtin-rules

# Issue a warning message whenever Make sees a reference to an undefined variable.
MAKEFLAGS += --warn-undefined-variables

# Bring in variables from the '.env' file, ignoring errors if it does not exist.
-include .env

# Export all variables to child processes by default.
# This is used to bring forward all of the values that have been set in the '.env' file included above.
.EXPORT_ALL_VARIABLES:

# Check that the version of Make running this file supports the .RECIPEPREFIX special variable.
# We set it to '>' to clarify inlined scripts and disambiguate whitespace prefixes.
# All script lines start with "> " which is the angle bracket and one space, with no tabs.
ifeq ($(origin .RECIPEPREFIX), undefined)
  $(error This Make does not support .RECIPEPREFIX. Please use GNU Make 4.0 or later.)
endif

.RECIPEPREFIX = >

# GNU make knows how to execute several recipes at once.
# Normally, make will execute only one recipe at a time, waiting for it to finish before executing the next.
# However, the '-j' or '--jobs' option tells make to execute many recipes simultaneously.
# With no argument, make runs as many recipes simultaneously as possible.
MAKEFLAGS += --jobs

# Configure an 'all' target to cover the bases.
all: test lint build ## Test and lint and build.
.PHONY: all

# Adjust the width of the first column by changing the '-20s' value in the printf pattern.
help:
> @grep -E '^[a-zA-Z0-9_-]+:.*? ## .*$$' $(filter-out .env, $(MAKEFILE_LIST)) | sort \
  | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'
.PHONY: help

# Set up some lazy initialisation functions to find code files, so that targets using the output of '$(shell ...)' only
# execute their respective shell commands when they need to, rather than every single instance of '$(shell ...)' being
# executed every single time 'make' is run for any target and wasting a lot of time.
# Further reading at https://www.oreilly.com/library/view/managing-projects-with/0596006101/ch10.html under the 'Lazy
# Initialization' heading.
find-go-files = $(shell find $1 -name vendor -prune -or -type f \( -iname '*.go' -or -name go.mod -or -name go.sum \))

GO_FILES = $(redefine-go-files) $(GO_FILES)

redefine-go-files = $(eval GO_FILES := $(call find-go-files, .))

# Tests look for sentinel files to determine whether or not they need to be run again.
# If any Go code file has been changed since the sentinel file was last touched, it will trigger a retest.
test: tmp/.tests-passed.sentinel ## Run tests.
test-cover: tmp/.cover-tests-passed.sentinel ## Run all tests with the race detector and output a coverage profile.
bench: tmp/.benchmarks-ran.sentinel ## Run enough iterations of each benchmark to take ten seconds each.
.PHONY: test test-cover bench

# Linter checks look for sentinel files to determine whether or not they need to check again.
# If any Go code file has been changed since the sentinel file was last touched, it will trigger a rerun.
lint: tmp/.linted.sentinel ## Lint all of the Go code. Will also test.
.PHONY: lint

# Builds look for image ID files to determine whether or not they need to build again.
# If any Go code file has been changed since the image ID file was last touched, it will trigger a rebuild.
build: tmp/.built.sentinel ## [DEFAULT] Build the library. Will also test and lint.
.PHONY: build

clean: ## Clean up any build output, test coverage, and the temp and output sub-directories.
> go clean -x -v
> rm -rf cover.out tmp out
.PHONY: clean

clean-hack: ## Deletes all binaries under 'hack'.
> rm -rf hack/bin
.PHONY: clean-hack

clean-all: clean clean-hack ## Clean all of the things.
.PHONY: clean-all

# Tests - re-run if any Go files have changes since 'tmp/.tests-passed.sentinel' was last touched.
tmp/.tests-passed.sentinel: $(GO_FILES)
> mkdir -p $(@D)
> go test -v ./...
> touch $@

tmp/.cover-tests-passed.sentinel: $(GO_FILES)
> mkdir -p $(@D)
> go test -count=1 -covermode=atomic -coverprofile=cover.out -race -v ./...
> touch $@

tmp/.benchmarks-ran.sentinel: $(GO_FILES)
> mkdir -p $(@D)
> go test -bench=. -benchmem -benchtime=10s -run='^DoNotRunTests$$' -v ./...
> touch $@

hack/bin/golangci-lint:
> mkdir -p $(@D)
> curl --fail --location --show-error --silent \
  https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh \
  | sh -s -- -b $(CURDIR)/hack/bin

# Lint - re-run if the tests have been re-run (and so, by proxy, whenever the source files have changed).
# These checks are all read-only and will not make any changes.
tmp/.linted.sentinel: tmp/.linted.gofmt.sentinel tmp/.linted.go.vet.sentinel tmp/.linted.golangci-lint.sentinel
> mkdir -p $(@D)
> touch $@

tmp/.linted.gofmt.sentinel: tmp/.tests-passed.sentinel
> mkdir -p $(@D)
> find . -type f -iname "*.go" -exec gofmt -d -e -l -s "{}" + \
  | awk '{ print } END { if (NR != 0) { print "Please run \"make gofmt\" to fix these issues!"; exit 1 } }'
> touch $@

tmp/.linted.go.vet.sentinel: tmp/.tests-passed.sentinel
> mkdir -p $(@D)
> go vet ./...
> touch $@

tmp/.linted.golangci-lint.sentinel: .golangci.yaml hack/bin/golangci-lint tmp/.tests-passed.sentinel
> mkdir -p $(@D)
> hack/bin/golangci-lint run --verbose
> touch $@

tmp/.built.sentinel: tmp/.linted.sentinel
> mkdir -p $(@D)
> go build -ldflags="-buildid= -w" -trimpath -v
> touch $@

gofmt: ## Runs 'gofmt -s' to format and simplify all Go code.
> find . -type f -iname "*.go" -exec gofmt -s -w "{}" +
.PHONY: gofmt
