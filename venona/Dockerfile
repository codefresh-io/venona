FROM golang:1.25-alpine3.23 AS build
RUN apk -U add --no-cache git make ca-certificates && update-ca-certificates
ENV USER=venona
ENV UID=10001 
RUN adduser \    
    --disabled-password \    
    --gecos "" \    
    --home "/nonexistent" \    
    --shell "/sbin/nologin" \    
    --no-create-home \    
    --uid "${UID}" \    
    "${USER}"
WORKDIR /venona
COPY . .
RUN go mod download -x
RUN go mod verify
RUN make build


FROM alpine:3.23.3 AS prod
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /etc/passwd /etc/passwd
COPY --from=build /etc/group /etc/group
WORKDIR /home/venona
RUN chown -R venona:venona /home/venona \
    && chmod 755 /home/venona
COPY --from=build /venona/venona /usr/local/bin/venona
USER venona:venona

ENTRYPOINT [ "venona" ]
CMD [ "start" ]
