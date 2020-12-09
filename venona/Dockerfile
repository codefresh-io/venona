FROM golang:alpine AS venona
COPY ./ /venona
WORKDIR /venona
RUN apk add make
RUN make build

FROM alpine:3.12

RUN apk update && apk add --no-cache ca-certificates && apk upgrade

COPY --from=venona /venona/venona /usr/local/bin/venona

ENTRYPOINT [ "venona" ]

CMD [ "start" ]
