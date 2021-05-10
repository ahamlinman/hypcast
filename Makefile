hypcast-server: go.mod go.sum $(shell find . -name '*.go') client/build.zip
	go build -v -tags embedclientzip ./cmd/hypcast-server

.PHONY: client/build.zip

client/build.zip:
	$(MAKE) -C client build.zip

.PHONY: install clean

install: hypcast-server
	go install -v -tags embedclientzip ./cmd/hypcast-server

clean:
	rm -f ./hypcast-server

# Configure clangd to resolve dependencies for C files in the project.
compile_flags.txt:
	pkg-config --cflags gstreamer-1.0 | tr ' ' '\n' | sed '/^$$/d' > compile_flags.txt

.PHONY: compile_flags.txt
