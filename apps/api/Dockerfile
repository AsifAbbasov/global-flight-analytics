ARG GO_IMAGE=golang:1.26.5-alpine3.24
ARG APP_VERSION=1.0.0
ARG VCS_REF=unknown
ARG BUILD_DATE=unknown

FROM ${GO_IMAGE} AS builder

ARG APP_VERSION
ARG VCS_REF
ARG BUILD_DATE

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
        materialize-flight-features \
        migrate \
        reconcile \
        server \
        verify-airports; \
    do \
        ldflags="-s -w -buildid="; \
        if [ "$command" = "server" ]; then \
            ldflags="$ldflags \
                -X github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/buildinfo.version=${APP_VERSION} \
                -X github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/buildinfo.revision=${VCS_REF} \
                -X github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/buildinfo.builtAt=${BUILD_DATE}"; \
        fi; \
        CGO_ENABLED=0 go build \
            -mod=readonly \
            -buildvcs=false \
            -trimpath \
            -ldflags="$ldflags" \
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

ARG APP_VERSION
ARG VCS_REF
ARG BUILD_DATE

LABEL org.opencontainers.image.title="Global Flight Analytics API"
LABEL org.opencontainers.image.description="Container image for the Global Flight Analytics Go backend"
LABEL org.opencontainers.image.source="https://github.com/AsifAbbasov/global-flight-analytics"
LABEL org.opencontainers.image.version="${APP_VERSION}"
LABEL org.opencontainers.image.revision="${VCS_REF}"
LABEL org.opencontainers.image.created="${BUILD_DATE}"

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

# STAGE-14-4-FEATURE-MATERIALIZATION
