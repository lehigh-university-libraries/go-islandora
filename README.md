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

The Preserve embargo date comes only from `DISS_repository/DISS_delayed_release`.
Dates and timestamps are accepted; `never deliver` uses the existing indefinite
embargo date `2999-12-31`. An absent or empty value means no IR embargo date.
`DISS_sales_restriction/@remove` and durations implied by ProQuest embargo codes
do not set or extend the Preserve embargo, even when the ProQuest date is later.
Codes outside 0–4 still cause an error.

Open-access and campus-only options do not clear an explicit IR embargo.
Campus-only access additionally sets `field_local_restriction=true`. Unknown
access options or unresolvable repository restrictions (including periods without
an explicit end date) cause an error.

This separation follows [ProQuest's embargo guidance](https://pq-static-content.proquest.com/collateral/media2/documents/umi_embargoesrestrictions_guide.pdf):
IR dissemination policies are managed independently by the university.
[Ex Libris's ETD Administrator mapping](https://knowledge.exlibrisgroup.com/Esploro/Product_Documentation/Esploro_Online_Help_%28English%29/Esploro_Integration/ETD_Administrator_Mapping_to_Esploro_Asset/ETD_Administrator_Mapping_to_Esploro_Assets)
also identifies `DISS_delayed_release` as the IR embargo value. Its Esploro-specific
mapping ignores that value when access is explicitly open or campus-only; Preserve
instead retains an explicit IR embargo as a local precaution. The `2999-12-31`
sentinel and rejection of periods without a date are also local implementation
choices, not ProQuest requirements.

The separate IR PDF embargo form is not represented in the delivered XML. College
coordinators must reconcile that form with the student's IR selection in the ETD
system. An XML-only audit cannot detect a discrepancy between the form and the
system; audits must also compare the approved IR forms with those settings.

The ETD ZIP ingest also includes `Local Restriction` (mapped to
`field_local_restriction`). The date-backfill command compares and emits
`field_local_restriction` alongside dates; an omitted restriction column in the
input CSV is treated as false.
Backfill matches records by full title, falling back to the first 255 characters.

## Resources

### Crossref

- [Crossref XML documentation](https://data.crossref.org/reports/help/schema_doc/5.3.1/index.html)
- [Crossref XML checker](https://www.crossref.org/02publishers/parser.html)
