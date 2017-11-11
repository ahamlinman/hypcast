#!/bin/bash

set -ev

# Install all build-time dependencies using Yarn
yarn install --ignore-optional

# Run the full build for server and client
yarn run build:mini

# Clean up all non-production dependencies
yarn install --production

# Clean up the Yarn cache
yarn cache clean
