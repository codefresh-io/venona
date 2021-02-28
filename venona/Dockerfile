FROM golang:1.14.4-alpine3.12 as build

RUN apk -U add git make 

WORKDIR /venona

COPY go.mod .
RUN go mod download

RUN chgrp -R 0 /venona /root \
  && chmod -R g+rwX /venona /root

COPY . .

RUN make build

FROM alpine:3.12

RUN apk update && apk add --no-cache ca-certificates && apk upgrade

COPY --from=build /venona/venona /usr/local/bin/venona

ENTRYPOINT [ "venona" ]

CMD [ "start" ]