TARGET=application
DATA=data

.PHONY: build

build:
	go build -v -o $(TARGET) cmd/$(TARGET)/main.go

run:
	mkdir -p $(DATA) 
	docker-compose up -d

clean:
	rm $(TARGET)
