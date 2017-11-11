#!/bin/bash

set -e
set -v

# Install all dependencies using Yarn
yarn install

# Run the full build for server and client
yarn run build:mini

# Clean up all non-production dependencies
yarn install --production

# Clean up the Yarn cache
yarn cache clean
