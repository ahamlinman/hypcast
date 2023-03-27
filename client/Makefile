assets.zip: dist
	rm -f assets.zip
	cd dist; zip -r ../assets.zip .

dist: node_modules tsconfig.json tsconfig.node.json vite.config.ts $(shell find src -type f)
	rm -rf dist
	yarn build

node_modules:
	yarn install
	touch node_modules

.PHONY: clean
clean:
	rm -rf dist assets.zip
