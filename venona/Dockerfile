FROM golang:1.14.4-alpine3.12 AS godev

RUN apk update && apk add --no-cache ca-certificates && apk upgrade && apk add git make

WORKDIR /venona

COPY . .

ENV GO111MODULE=on
ENV GOSUMDB=off
ENV GOPROXY=direct

RUN make build

FROM alpine:3.12

COPY VERSION .

RUN apk update && apk add --no-cache ca-certificates && apk upgrade

COPY --from=godev ./venona/venona /venona

ENTRYPOINT [ "/venona" ]

CMD [ "start" ]