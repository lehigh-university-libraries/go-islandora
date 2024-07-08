#!/usr/bin/env bash

set -eou pipefail

if [ ! -d ./islandora-starter-site ]; then
  git clone https://github.com/Islandora-Devops/islandora-starter-site
fi

go run main.go \
  --node-cex-yaml=./islandora-starter-site/config/sync/node.type.islandora_object.yml \
  --output=islandora/islandora_object.yaml

diff islandora/islandora_object.yaml fixtures/islandora_object.yaml || (echo "Failure Maybe starter site updated its data model?" && exit 1)

echo "Generated Open API spec matches expected output 🚀"
