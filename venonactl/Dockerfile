FROM alpine:3.9

RUN apk add --update ca-certificates

COPY dist/venona_linux_386/venona /

ENTRYPOINT [ "/venona" ]

CMD [ "--help" ]