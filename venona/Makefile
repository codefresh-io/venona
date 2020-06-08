.PHONY: build	
build:
	@sh ./scripts/build.sh

.PHONY: test
test:
	@sh ./scripts/test.sh

test-fmt:
	@sh ./scripts/test-fmt.sh

.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: mock
# example: make mock PKG=./pkg/runtime NAME=Runtime
mock:
	@sh ./scripts/mock.sh $(PKG) $(NAME)

.PHONY: license
license:
	@go get -u github.com/google/addlicense
	@addlicense -check -f License **/**/*.go