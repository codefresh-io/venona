ifndef GOBIN
ifndef GOPATH
$(error GOPATH is not set, please make sure you set your GOPATH correctly!)
endif
GOBIN=$(GOPATH)/bin
ifndef GOBIN
$(error GOBIN is not set, please make sure you set your GOBIN correctly!)
endif
endif

.PHONY: build
build:
	@sh ./scripts/build.sh

.PHONY: run
run: build
	./venona start

.PHONY: debug
debug: build
	./venona start --verbose

.PHONY: test-all
test-all: test test-fmt spellcheck gocyclo lint security-check license

.PHONY: test
test:
	@sh ./scripts/test.sh

.PHONY: test-fmt
test-fmt:
	@sh ./scripts/test-fmt.sh

# spellcheck Finds commonly misspelled English words
.PHONY: spellcheck
spellcheck:
	@misspell -error .

# Gocyclo calculates cyclomatic complexities of functions in Go source code.
# The cyclomatic complexity of a function is calculated according to the following rules: 
# 1 is the base complexity of a function +1 for each 'if', 'for', 'case', '&&' or '||'
# Go Report Card warns on functions with cyclomatic complexity > 15.
.PHONY: gocyclo
gocyclo:
	@gocyclo -over 15 .

.PHONY: lint
lint: $(GOBIN)/golangci-lint
	@echo linting go code...
	@$(GOBIN)/golangci-lint run --fix --timeout 10m

.PHONY: security-check
security-check:
	@gosec ./... -nosec

.PHONY: docker-security-scan
docker-security-scan:
	@trivy image --clear-cache
	@trivy image codefresh/venona:$(TAG)

## License check all the golang files to have the license
.PHONY: license
license:
	@addlicense -check -f License **/**/*.go

# Fix fmt errors in file
.PHONY: fmt
fmt:
	go fmt ./...

# Generate mock struct from interface
# example: make mock PKG=./pkg/runtime NAME=Runtime
.PHONY: mock
mock: $(GOBIN)/mockery
	@mockery

# Runs cript to upload codecov coverage data
.PHONY: upload-coverage
upload-coverage:
	@./scripts/codecov.sh -t $(CODECOV_TOKEN)

$(GOBIN)/mockery:
	@go install github.com/vektra/mockery/v2@v2.32.0
	@mockery --version

$(GOBIN)/golangci-lint:
	@echo installing: golangci-lint
	@curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(GOBIN) v1.54.2
