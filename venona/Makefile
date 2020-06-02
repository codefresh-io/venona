.PHONY: build	
build:
	go build -o venona *.go

.PHONY: test
test:
	@sh ./scripts/test.sh

test-fmt:
	@sh ./scripts/test-fmt.sh

.PHONY: fmt
fmt:
	go fmt ./...