
COMMIT=`git rev-parse HEAD`
BUILD=`date +%FT%T%z`
LDFLAG_LOCATION=github.com/eytan-avisror/ws-simple-pubsub/pubsub

LDFLAGS=-ldflags "-X ${LDFLAG_LOCATION}.buildDate=${BUILD} -X ${LDFLAG_LOCATION}.gitCommit=${COMMIT}"

GIT_TAG=$(shell git rev-parse --short HEAD)

build:
	CGO_ENABLED=0 go build ${LDFLAGS} -o bin/ws-simple-pubsub github.com/eytan-avisror/ws-simple-pubsub
	chmod +x bin/ws-simple-pubsub

test:
	go test -v ./... -coverprofile coverage.txt
	go tool cover -html=coverage.txt -o coverage.html
