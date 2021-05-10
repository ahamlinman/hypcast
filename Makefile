hypcast-server: go.mod go.sum $(shell find . -name '*.go') client/build.zip
	go build -v -tags embedclientzip ./cmd/hypcast-server

client/build.zip: client/build
	rm -f client/build.zip
	cd client/build; zip -r ../build.zip .

client/build: client/node_modules client/tsconfig.json $(shell find client/public client/src -type f)
	cd client; yarn build
	touch client/build

client/node_modules: client/package.json client/yarn.lock
	cd client; yarn install
	touch client/node_modules

.PHONY: install clean

install: hypcast-server
	go install -v -tags embedclientzip ./cmd/hypcast-server

clean:
	rm -f ./hypcast-server

# Configure clangd to resolve dependencies for C files in the project.
compile_flags.txt:
	pkg-config --cflags gstreamer-1.0 | tr ' ' '\n' | sed '/^$$/d' > compile_flags.txt

.PHONY: compile_flags.txt
