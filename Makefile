hypcast-server: go.mod go.sum $(shell find . -name '*.go') client/assets.zip
	go build -v -tags embedclientzip ./cmd/hypcast-server

client/assets.zip: client/dist
	rm -f client/assets.zip
	cd client/dist && zip -r ../assets.zip .

client/dist: client/node_modules client/tsconfig.json client/tsconfig.node.json client/vite.config.ts $(shell find client/src -type f)
	rm -rf client/dist
	cd client && yarn build

client/node_modules: client/package.json client/yarn.lock
	cd client && yarn install
	touch client/node_modules

install: hypcast-server
	go install -v -tags embedclientzip ./cmd/hypcast-server

clean:
	$(MAKE) -C client clean
	rm -rf ./hypcast-server

# Configure clangd to resolve dependencies for C files in the project.
.PHONY: compile_flags.txt
compile_flags.txt:
	pkg-config --cflags gstreamer-1.0 | tr ' ' '\n' | sed '/^$$/d' > compile_flags.txt
