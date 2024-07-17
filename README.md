# go-islandora

Drupal/Islandora go library

Generate an Open API spec based on your Islandora data model

```
./go-islandora generate node-structs \
  --node-cex-yaml=path/to/drupal/config/sync/node.type.islandora_object.yml \
  --output=api.yml
```


# Create Crossref XML for a journal that only has volumes

```
$ go-islandora export csv \
  --baseUrl https://your.islandora.url \
  --nid NODE
$ go-islandora generate sheets-structs
$ go-islandora transform csv crossref \
  --source merged.csv \
  --target journal.xml \
  --type issueless-journal \
  --registrant "Lehigh University Libraries" \
  --depositor-name "Lehigh University Libraries" \
  --depositor-email inpresrv@lehigh.edu
```

## Resources

### Crossref

- [Crossref XML documentation](https://data.crossref.org/reports/help/schema_doc/5.3.1/index.html)
- [Crossref XML checker](https://www.crossref.org/02publishers/parser.html)
