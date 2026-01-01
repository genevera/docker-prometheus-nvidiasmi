FROM golang:1.21 AS builder
WORKDIR /go
COPY src/app.go app.go
ENV CGO_ENABLED=0
RUN go build -trimpath -ldflags="-s -w" -o /out/app app.go
LABEL org.opencontainers.image.title="Nvidia SMI exporter for your Prometheus-like scraper"
LABEL org.opencontainers.image.authors='<genevera@users.noreply.github.com>'
LABEL org.opencontainers.image.source="https://github.com/genevera/docker-prometheus-nvidiasmi"
FROM gcr.io/distroless/base-debian12
COPY --from=builder /out/app /app
EXPOSE 9202
ENTRYPOINT ["/app"]