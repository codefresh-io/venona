FROM golang:1.14.4-alpine3.12 as build

RUN apk -U add git make 

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
RUN go mod download
RUN go mod verify

COPY . .

RUN make build

FROM alpine:3.12

RUN apk -U upgrade

RUN apk add --no-cache ca-certificates

COPY --from=build /etc/passwd /etc/passwd
COPY --from=build /etc/group /etc/group

# copy binary
COPY --from=build /venona/venona /usr/local/bin/venona

USER venona:venona

ENTRYPOINT [ "venona" ]

CMD [ "start" ]