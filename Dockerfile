# Stage 1: Build netns exporter
FROM golang:1.15 AS builder

ENV CGO_ENABLED=0
ENV GOOS=linux
COPY . /go/netns-exporter
WORKDIR /go/netns-exporter
RUN go build
RUN strip netns-exporter
RUN mkdir -p /etc_tmp/netns-exporter
RUN cp /go/netns-exporter/config.docker.yaml /etc_tmp/netns-exporter/config.yaml

# Stage 2: Prepare final image
FROM ubuntu:20.04

# Copy binary from Stage 1
COPY --from=builder /go/netns-exporter/netns-exporter .
COPY --from=builder /etc_tmp/netns-exporter/config.yaml /etc/netns-exporter/config.yaml

ENTRYPOINT [ "/netns-exporter" ]
