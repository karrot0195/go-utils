.PHONY: build clean

clean: 
	@rm -rf ./bin/*

build: clean
	@go build -o ./bin ./cmd/openroute