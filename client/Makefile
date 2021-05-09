build: node_modules tsconfig.json $(shell find public src -type f)
	yarn build
	touch build

node_modules: package.json yarn.lock
	yarn install
	touch node_modules
