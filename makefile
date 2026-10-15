APPS := $(notdir $(wildcard ./cmd/*))

.PHONY: build clean

clean:
	@rm -rf ./bin/*

build: clean
	@for app in $(APPS); do \
		go build -o ./bin/$$app ./cmd/$$app; \
		echo "Build $$app"; \
	done
