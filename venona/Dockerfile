FROM octopusdeploy/dhi-golang:1.27.4-alpine3.24-dev@sha256:72cac473a0b802b521e23a2aed591ad166645729533d6d3db4ce96ae22c697ff AS build
RUN apk -U add --no-cache git make ca-certificates && update-ca-certificates
WORKDIR /venona
COPY . .
RUN go mod download -x
RUN go mod verify
RUN make build


FROM octopusdeploy/dhi-alpine-base:3.24@sha256:7b06d747fcb428a0699679955e903b13322c92c79d94ce94c108634f13a714f1 AS prod
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
