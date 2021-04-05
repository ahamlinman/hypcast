hypcast-server: go.mod go.sum $(shell find . -name '*.go')
	go build -v ./cmd/hypcast-server

# Configure clangd to resolve dependencies for C files in the project.
compile_flags.txt:
	pkg-config --cflags gstreamer-1.0 | tr ' ' '\n' > compile_flags.txt
