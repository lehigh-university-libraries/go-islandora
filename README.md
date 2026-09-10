# go-islandora

Drupal/Islandora go library

Generate an Open API spec based on your Islandora data model

```
go-islandora generate node-structs \
  --node-cex-yaml=path/to/drupal/config/sync/node.type.islandora_object.yml \
  --output=api.yaml

go-islandora generate sheets-structs --output=workbench.yaml
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

## Extract ETD embargo dates

Read a directory of ProQuest XML files, including subdirectories:

```sh
go-islandora transform etd-embargo --source /path/to/xml --target embargoes.csv
```

Omit `--target` to write CSV to stdout. Output includes the XML file path,
`field_edtf_date_embargo`, `field_local_restriction`, and `pdf_file`.
The `pdf_file` column contains the filename from `DISS_content/DISS_binary`
(for example, `Amin_lehigh_0105A_13188.pdf`), or is blank if it is absent.

Add `--future-only` to include only embargo dates strictly after today (UTC).
This includes indefinite embargoes (`2999-12-31`) and excludes dates ending
today, past dates, and empty dates, including campus-only records without a date.

```sh
go-islandora transform etd-embargo --source /path/to/xml --future-only --target embargoes.csv
```

The embargo date is the later of the repository delayed-release date and the
ProQuest sales embargo date. This is our local rule for combining the two
independent restrictions. Open-access and campus-only options do not clear either
embargo. Campus-only access additionally sets `field_local_restriction=true`.
`never deliver` takes precedence over dated embargoes and uses the existing indefinite
embargo date `2999-12-31`. Unknown access options or unresolvable repository
restrictions (including periods without an explicit end date) cause an error.

The ProQuest date comes from a valid sales restriction `remove` date, or otherwise
from adding six calendar months, one year, or two years to the acceptance date for
codes 1–3. Codes outside 0–4 cause an error. Codes 0 and 4 without a valid sales
release date contribute no date; a valid repository date is still retained.

The ETD ZIP ingest also includes `Local Restriction` (mapped to
`field_local_restriction`). The date-backfill command compares and emits
`field_local_restriction` alongside dates; an omitted restriction column in the
input CSV is treated as false.
Backfill matches records by full title, falling back to the first 255 characters.

## Resources

### Crossref

- [Crossref XML documentation](https://data.crossref.org/reports/help/schema_doc/5.3.1/index.html)
- [Crossref XML checker](https://www.crossref.org/02publishers/parser.html)
