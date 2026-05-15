BINARY := gcm
CMD_DIR := ./cmd/gcm

.PHONY: build test run clean

build:
	go build -o $(BINARY) $(CMD_DIR)

test:
	go test ./...

run: build
	./$(BINARY)

clean:
	rm -f $(BINARY)
