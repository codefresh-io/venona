FROM golang:1.15.8-alpine3.12 as build

RUN apk -U add --no-cache git make ca-certificates && update-ca-certificates

ENV USER=venona
ENV UID=10001 

RUN which curl

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

# compile
RUN make build

FROM alpine:3.12

# copy ca-certs and user details
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /etc/passwd /etc/passwd
COPY --from=build /etc/group /etc/group

WORKDIR /home/venona

# copy binary
COPY --from=build /venona/venona /usr/local/bin/venona

USER venona:venona

ENTRYPOINT [ "venona" ]

CMD [ "start" ]