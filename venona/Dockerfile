FROM octopusdeploy/dhi-golang:1.26-alpine3.24-dev@sha256:8baca6975280a6b32bd947ea981230b587599ae6c0985105163a0a5d2d6eb84b AS build
RUN apk -U add --no-cache git make ca-certificates && update-ca-certificates
WORKDIR /venona
COPY . .
RUN go mod download -x
RUN go mod verify
RUN make build


FROM octopusdeploy/dhi-alpine-base:3.24@sha256:32e94d5fda692447a1ca48ff8910e5be4657053ba74744cd6b0cf6ed816c6cbf AS prod
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
