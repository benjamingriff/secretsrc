.PHONY: help run build test install clean

BINARY := secretsrc
CMD := ./cmd/secretsrc

help:
	@echo "Available targets:"
	@echo "  make run      Run the app with go run"
	@echo "  make build    Build a local $(BINARY) binary in the repo root"
	@echo "  make test     Run go test ./..."
	@echo "  make install  Install $(BINARY) to GOBIN or GOPATH/bin"
	@echo "  make clean    Remove the local repo binary"

run:
	go run $(CMD)

build:
	go build -o $(BINARY) $(CMD)

test:
	go test ./...

install:
	go install $(CMD)

clean:
	rm -f $(BINARY)
