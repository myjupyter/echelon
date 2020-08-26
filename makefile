TARGET=application

.PHONY: build

build:
	go build -v -o $(TARGET) cmd/$(TARGET)/main.go

