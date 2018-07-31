.PHONY: force

all: dependencies polling

polling: polling-test polling-build

polling-build: force
	go build -i -o ${GOPATH}/bin/fridgelethics-polling-service github.com/nafcollective/fridgelethics-service/v0/cmd/polling

polling-test: force
	go test -race -v ./v0/pkg/polling
	go test -race -v ./v0/cmd/polling

dependencies: force
	go get -t ./...