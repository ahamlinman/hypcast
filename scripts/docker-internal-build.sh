#!/bin/bash

set -e
set -v

# Install Yarn
npm install --global yarn

# Install all dependencies using Yarn
yarn install

# Run the full build for server and client
npm run build:mini

# Clean up all non-production dependencies
yarn install --production

# Clean up the Yarn cache
yarn cache clean

# Uninstall Yarn
npm uninstall --global yarn

# Clean up the NPM cache
npm cache clean
