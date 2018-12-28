FROM golang:alpine3.7 as gobuild

RUN wget -O /go/bin/dep https://github.com/golang/dep/releases/download/v0.5.0/dep-linux-amd64 && \
    chmod +x /go/bin/dep

COPY Gopkg* /go/src/github.com/codefresh-io/Isser/installer/
WORKDIR /go/src/github.com/codefresh-io/Isser/installer/

RUN mkdir -p /go/src/github.com/codefresh-io/Isser/installer/{cmd,pkg}
RUN dep ensure --vendor-only

COPY cmd ./cmd/
# COPY lib ./lib/
COPY pkg ./pkg/

RUN go generate pkg/runtime/types.go
RUN go build -o /usr/local/bin/isserctl ./cmd && \
    chmod +x /usr/local/bin/isserctl && \
    rm -rf /go/*

FROM alpine:3.7
COPY --from=gobuild /usr/local/bin/isserctl /usr/local/bin/