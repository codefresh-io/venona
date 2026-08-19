FROM octopusdeploy/dhi-golang:1.26-debian13-dev@sha256:e7f194366f8993d4c33eeeb24d96eee1b615c695127606917e93a9a81c3bd8ea AS build
RUN apt update && apt install -y git make ca-certificates && update-ca-certificates
ENV USER=venona
ENV UID=10001
USER root
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


FROM octopusdeploy/dhi-debian-base:trixie-debian13@sha256:014bacad6433af5bc9bd098ac713c4f104f9b395dd1d9d58bbcafd51ad2ce582 AS prod
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
