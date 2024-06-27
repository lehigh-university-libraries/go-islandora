#!/usr/bin/env bash

set -eou pipefail

if [ ! -d ./islandora-starter-site ]; then
  git clone git@github.com:Islandora-Devops/islandora-starter-site
fi

go run main.go \
  --node-cex-yaml=./islandora-starter-site/config/sync/node.type.islandora_object.yml \
  --output=islandora_object.go

go fmt islandora_object.go > /dev/null

diff islandora_object.go fixtures/islandora_object.go || (echo "Failure Maybe starter site updated its data model?" && exit 1)

echo "Generated struct matches expected output 🚀"
