ARG GO_IMAGE=golang:1.26.5-alpine3.24

FROM ${GO_IMAGE} AS builder

WORKDIR /workspace/apps/api

COPY apps/api/go.mod apps/api/go.sum ./

RUN go mod download && \
    go mod verify

COPY apps/api/ ./

RUN set -eux; \
    mkdir -p /out; \
    for command in \
        healthcheck \
        import-airports \
        ingest \
        migrate \
        reconcile \
        server \
        verify-airports; \
    do \
        CGO_ENABLED=0 go build \
            -mod=readonly \
            -buildvcs=false \
            -trimpath \
            -ldflags="-s -w -buildid=" \
            -o "/out/${command}" \
            "./cmd/${command}"; \
    done

RUN set -eux; \
    test -s /etc/ssl/certs/ca-certificates.crt; \
    mkdir -p \
        /rootfs/etc/ssl/certs \
        /rootfs/tmp; \
    cp \
        /etc/ssl/certs/ca-certificates.crt \
        /rootfs/etc/ssl/certs/ca-certificates.crt; \
    printf '%s\n' \
        'app:x:10001:10001:Global Flight Analytics:/nonexistent:/sbin/nologin' \
        > /rootfs/etc/passwd; \
    printf '%s\n' \
        'app:x:10001:' \
        > /rootfs/etc/group; \
    chmod 1777 /rootfs/tmp

FROM scratch AS runtime

LABEL org.opencontainers.image.title="Global Flight Analytics API"
LABEL org.opencontainers.image.description="Container image for the Global Flight Analytics Go backend"
LABEL org.opencontainers.image.source="https://github.com/AsifAbbasov/global-flight-analytics"

WORKDIR /app

COPY --from=builder /rootfs/ /
COPY --from=builder --chown=10001:10001 /out/ /app/
COPY --chown=10001:10001 database/migrations/ /app/migrations/

ENV API_PORT=8080
ENV MIGRATIONS_DIR=/app/migrations
ENV HEALTHCHECK_TIMEOUT=2s

USER 10001:10001

EXPOSE 8080

STOPSIGNAL SIGTERM

HEALTHCHECK \
    --interval=10s \
    --timeout=3s \
    --start-period=5s \
    --retries=3 \
    CMD ["/app/healthcheck"]

CMD ["/app/server"]
