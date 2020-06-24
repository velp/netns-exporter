build:
	go build

build-in-docker:
	docker run -it --rm -v `pwd`:/go/netns-exporter -w /go/netns-exporter -e CGO_ENABLED=0 -e GOOS=linux golang:1.14 go build

lint: ${GOPATH}/bin/golangci-lint
	GOGC=20 golangci-lint run --deadline 3m0s --enable-all --disable gochecknoglobals,lll,funlen

${GOPATH}/bin/golangci-lint:
	curl -sfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b ${GOPATH}/bin v1.18.0
