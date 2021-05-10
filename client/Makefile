build: node_modules tsconfig.json $(shell find public src -type f)
	yarn build
	touch build

build.zip: build
	rm -f build.zip
	cd build; zip -r ../build.zip .

node_modules: package.json yarn.lock
	yarn install
	touch node_modules
