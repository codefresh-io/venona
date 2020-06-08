.PHONY: build	
build:
	@sh ./scripts/build.sh

.PHONY: test
test:
	@sh ./scripts/test.sh

.PHONY: test-fmt
test-fmt:
	@sh ./scripts/test-fmt.sh

# spellcheck Finds commonly misspelled English words
.PHONY: spellcheck
spellcheck:
	@go get github.com/client9/misspell/cmd/misspell
	@misspell -error .

# Gocyclo calculates cyclomatic complexities of functions in Go source code.
# The cyclomatic complexity of a function is calculated according to the following rules: 
# 1 is the base complexity of a function +1 for each 'if', 'for', 'case', '&&' or '||'
# Go Report Card warns on functions with cyclomatic complexity > 15.
.PHONY: gocyclo
gocyclo:
	@go get github.com/fzipp/gocyclo
	@gocyclo -over 15 .

.PHONY: lint
lint:
	@go get golang.org/x/lint/golint
	@golint -set_exit_status ./...

.PHONY: security-check
lint:
	@go get github.com/securego/gosec/cmd/gosec
	@gosec ./...

## License check all the golang files to have the license
.PHONY: license
license:
	@go get github.com/google/addlicense
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
