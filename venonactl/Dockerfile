FROM golang:1.21-alpine3.20 AS build
WORKDIR /venona
COPY go.mod .
RUN go mod download
RUN apk add git 
COPY . .
ARG COMMIT
RUN VERSION=$(cat VERSION) \
    DATE=$(date -u "+%Y-%m-%dT%TZ") && \
    env CGO_ENABLED=0 \
    go build -ldflags="-w -X github.com/codefresh-io/venona/venonactl/cmd.version=${VERSION} \ 
    -X github.com/codefresh-io/venona/venonactl/cmd.commit=${COMMIT} -X github.com/codefresh-io/venona/venonactl/cmd.date=${DATE}" \
    -o venona

FROM alpine:3.20
RUN apk add --update ca-certificates
COPY --from=build /venona/venona /usr/local/bin/venona

ENTRYPOINT [ "venona" ]
CMD [ "--help" ]
