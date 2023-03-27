hypcast-server: go.mod go.sum $(shell find . -name '*.go') client/assets.zip
	go build -v -tags embedclientzip ./cmd/hypcast-server

client/assets.zip:
	$(MAKE) -C client assets.zip

.PHONY: install clean

install: hypcast-server
	go install -v -tags embedclientzip ./cmd/hypcast-server

clean:
	$(MAKE) -C client clean
	rm -rf ./hypcast-server

# Configure clangd to resolve dependencies for C files in the project.
.PHONY: compile_flags.txt
compile_flags.txt:
	pkg-config --cflags gstreamer-1.0 | tr ' ' '\n' | sed '/^$$/d' > compile_flags.txt
