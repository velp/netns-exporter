# Stage 1: Build netns exporter
FROM golang:1.15 AS builder

ENV CGO_ENABLED=0
ENV GOOS=linux
COPY . /go/netns-exporter
WORKDIR /go/netns-exporter
RUN go build
RUN strip netns-exporter

# Stage 2: Prepare final image
FROM alpine:3.14.1


# Copy binary from Stage 1
COPY --from=builder /go/netns-exporter/netns-exporter .
COPY config.docker.yaml /etc/netns-exporter/config.yaml

ENTRYPOINT [ "/netns-exporter" ]
