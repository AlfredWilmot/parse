
all: fmt build

# Build the application into a static binary
build:
	# https://stackoverflow.com/a/61324538/22415851
	CGO_ENABLED=0 go build -a -ldflags '-extldflags "-static"' .

fmt:
	go mod tidy
	find -iname "*.go" | xargs gofmt -w -s

.PHONY: build fmt
