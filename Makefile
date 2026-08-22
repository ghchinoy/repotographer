BINARY  := repotographer
BIN_DIR := bin
PKG     := .

.PHONY: build clean test vet fmt install run all

all: build

build:
	@mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/$(BINARY) $(PKG)

test:
	go test ./...

vet:
	go vet ./...

fmt:
	gofmt -w .

install:
	go install .

clean:
	rm -rf $(BIN_DIR)
	rm -f $(BINARY)

run: build
	./$(BIN_DIR)/$(BINARY)
