GLOLANGCI_LINT_VERSION = 1.30.0


install-deps:
	go mod download

install-test:
	@(cd; GO111MODULE=on go get github.com/golangci/golangci-lint/cmd/golangci-lint@v$(GLOLANGCI_LINT_VERSION))

build:
	go build

build-in-docker:
	docker run -it --rm -v `pwd`:/go/netns-exporter -w /go/netns-exporter -e CGO_ENABLED=0 -e GOOS=linux golang:1.15 go build

docker-image:
	docker build . -t netns-exporter

test: .FORCE
	go test -v -race $(call get_go_packages)

.PHONY: .FORCE

lint:
	GOGC=20 golangci-lint run --deadline 3m0s --enable-all --disable gochecknoglobals,lll,funlen

define get_go_packages
	$(shell go list ./... $(foreach pattern,$1,| grep -v /$(pattern)/))
endef
