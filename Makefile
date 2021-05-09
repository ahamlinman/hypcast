hypcast-server: go.mod go.sum $(shell find . -name '*.go') client/build
	go build -v ./cmd/hypcast-server

client/build:
	$(MAKE) -C client build

.PHONY: clean

clean:
	rm -f ./hypcast-server

# Configure clangd to resolve dependencies for C files in the project.
compile_flags.txt:
	pkg-config --cflags gstreamer-1.0 | tr ' ' '\n' | sed '/^$$/d' > compile_flags.txt

.PHONY: compile_flags.txt
