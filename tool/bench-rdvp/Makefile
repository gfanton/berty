go_files := $(shell find . -type f -name '*.go')

build: build/bench
build/bench: $(go_files)
	mkdir -p build
	go build -o $@ .
.PHONY: build
