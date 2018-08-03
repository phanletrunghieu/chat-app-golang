.PHONY: build run

build:
	go build -o server ./cmd/app/main.go
run: build
	./server