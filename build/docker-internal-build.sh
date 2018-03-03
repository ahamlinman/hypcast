#!/bin/bash

set -xeuo pipefail

# Install all build-time dependencies using Yarn. Usually I'd ignore optional
# packages, but Yarn is buggy here: https://github.com/yarnpkg/yarn/issues/4876
yarn install

# Run the full build for server and client
yarn run build:mini

# Clean up all non-production dependencies
yarn install --production

# Clean up the Yarn cache
yarn cache clean
