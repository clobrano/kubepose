.PHONY: build run test clean

BINARY_NAME=kubepose

build:
	go build -o $(BINARY_NAME) ./cmd/kubepose

run: build
	./$(BINARY_NAME)

test:
	go test ./...

clean:
	rm -f $(BINARY_NAME)
	go clean
