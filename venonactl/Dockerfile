FROM golang:alpine3.7 as gobuild

RUN wget -O /go/bin/dep https://github.com/golang/dep/releases/download/v0.5.0/dep-linux-amd64 && \
    chmod +x /go/bin/dep

COPY Gopkg* /go/src/github.com/codefresh-io/venona/venonactl/
WORKDIR /go/src/github.com/codefresh-io/venona/venonactl/

RUN mkdir -p /go/src/github.com/codefresh-io/venona/venonactl/{cmd,pkg}
RUN dep ensure --vendor-only

COPY cmd ./cmd/
# COPY lib ./lib/
COPY pkg ./pkg/

RUN go generate pkg/runtimectl/types.go
RUN go build -o /usr/local/bin/venonactl ./cmd && \
    chmod +x /usr/local/bin/venonactl && \
    rm -rf /go/*

FROM alpine:3.7
COPY --from=gobuild /usr/local/bin/venonactl /usr/local/bin/