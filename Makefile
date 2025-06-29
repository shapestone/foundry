.PHONY: build run clean

build:
	go build -o bin/foundry cmd/foundry/main.go

run:
	go run cmd/foundry/main.go

clean:
	rm -rf bin/
