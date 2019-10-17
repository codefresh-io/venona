FROM alpine:3.8

RUN apk add --update ca-certificates

COPY dist/venona_linux_386/venona /

ENTRYPOINT [ "/venona" ]

CMD [ "--help" ]