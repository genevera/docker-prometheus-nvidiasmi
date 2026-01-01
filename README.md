# Docker Prometheus Nvidia SMI Exporter

Dockerized Prometheus exporter for GPU statistics from [nvidia-smi](https://developer.nvidia.com/nvidia-system-management-interface), written in Go.
Supports multiple GPUs.

# How-To

Run with a Docker command:
`docker run --gpus=all -p 9202:9202/tcp quay.io/genevera/prometheus-nvidiasmi`

Run through a docker-compose file:
```
services:
  prometheus-nvidiasmi:
    image: quay.io/genevera/prometheus-nvidiasmi
    gpus: all
    ports:
      - "9202:9202/tcp"
```

Check result at: [http://localhost:9202/metrics](http://localhost:9202/metrics)

# Grafana dashboard

[Nvidia SMI Metrics dashboard](https://grafana.com/grafana/dashboards/12357) on Grafana Labs


[![Badge showing the build status of the Docker repository for genevera prometheus nvidiasmi on Quay. The badge displays the current status such as ready or building in a rectangular format with a neutral and informative tone. The badge links to the Docker repository page on Quay.](https://quay.io/repository/genevera/prometheus-nvidiasmi/status "Docker Repository on Quay")](https://quay.io/repository/genevera/prometheus-nvidiasmi)
