FROM golang:1.21 AS builder

WORKDIR /go

COPY src/app.go app.go

ENV CGO_ENABLED=0
RUN go build -v -ldflags '-w -extldflags "-static"' -o bin/app app.go

FROM nvidia/cuda:12.8.0-base-ubuntu24.04

LABEL org.opencontainers.image.title="Nvidia SMI exporter for Prometheus-like scrappers"
LABEL org.opencontainers.image.authors='Psycle Research <tech@psycle.io>'
LABEL org.opencontainers.image.source="https://github.com/PsycleResearch/docker-prometheus-nvidiasmi"

COPY --from=builder /go/bin/app /app

EXPOSE 9202

CMD [ "/app" ]
