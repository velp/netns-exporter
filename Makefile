build:
	go build

build-in-docker:
	docker run -it --rm -v `pwd`:/go/netns-exporter -w /go/netns-exporter -e CGO_ENABLED=0 -e GOOS=linux golang:1.12 go build
