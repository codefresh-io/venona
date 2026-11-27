FROM octopusdeploy/dhi-golang:1.27.4-alpine3.24-dev@sha256:a76849b53b06b671dff57ebea6a83206c54a74e5fbd27ba365b9e2dc94a18195 AS build
RUN apk -U add --no-cache git make ca-certificates && update-ca-certificates
WORKDIR /venona
COPY . .
RUN go mod download -x
RUN go mod verify
RUN make build


FROM octopusdeploy/dhi-alpine-base:3.24@sha256:7d9384f025a398c7a60c4a978bf419f3b4bd86bde0577d352d902224d4dc7d87 AS prod
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
