BINARY := gcm
CMD_DIR := ./cmd/gcm

.PHONY: build test run install clean

build:
	go build -o $(BINARY) $(CMD_DIR)

test:
	go test ./...

run: build
	./$(BINARY)

install:
	go install $(CMD_DIR)

clean:
	rm -f $(BINARY)
