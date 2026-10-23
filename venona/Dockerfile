FROM octopusdeploy/dhi-golang:1.26-alpine3.24-dev@sha256:c48766f47ce584f72b290949310376b0c6dfdc7565aed51f219507d42807501e AS build
RUN apk -U add --no-cache git make ca-certificates && update-ca-certificates
WORKDIR /venona
COPY . .
RUN go mod download -x
RUN go mod verify
RUN make build


FROM octopusdeploy/dhi-alpine-base:3.24@sha256:648a78db84b3e98a95104e2e8d246f37f03a3bee3fb2fb7afb3ed91998b17483 AS prod
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /etc/passwd /etc/passwd
COPY --from=build /etc/group /etc/group
WORKDIR /home/venona
RUN chown -R nonroot:nonroot /home/venona \
    && chmod 755 /home/venona
COPY --from=build /venona/venona /usr/local/bin/venona
USER nonroot

ENTRYPOINT [ "venona" ]
CMD [ "start" ]
