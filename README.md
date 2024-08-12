# go-islandora

Drupal/Islandora go library

Generate an Open API spec based on your Islandora data model

```
go-islandora generate node-structs \
  --node-cex-yaml=path/to/drupal/config/sync/node.type.islandora_object.yml \
  --output=api.yml

go-islandora generate sheets-structs --output=workbench.yml
```
