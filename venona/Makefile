.PHONY: build	
build:
	@sh ./scripts/build.sh

.PHONY: run
run: build
	./venona start

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
lint:
	@golint -set_exit_status ./...

.PHONY: security-check
security-check:
	@gosec ./...

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
mock:
	@sh ./scripts/mock.sh $(PKG) $(NAME)

# Generate mock struct from interface
# example: make mock PKG=./pkg/runtime NAME=Runtime
.PHONY: upload-coverage
upload-coverage:
	@./scripts/codecov.sh -t $(CODECOV_TOKEN)
