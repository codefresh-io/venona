FROM alpine:3.12

RUN apk update && apk add --no-cache ca-certificates && apk upgrade

COPY venona /usr/local/venona

ENTRYPOINT [ "venona" ]

CMD [ "start" ]