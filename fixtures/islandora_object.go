package islandora

import (
	islandoraModel "github.com/lehigh-university-libraries/go-islandora/model"
)

type IslandoraObject struct {
	Nid               []islandoraModel.IntField             `json:"nid"`
	Vid               []islandoraModel.IntField             `json:"vid"`
	Uuid              []islandoraModel.GenericField         `json:"uuid"`
	Language          []islandoraModel.GenericField         `json:"lang"`
	RevisionTimestamp []islandoraModel.GenericField         `json:"revision_timestamp"`
	RevisionUid       []islandoraModel.EntityReferenceField `json:"revision_uid"`
	RevisionLog       []islandoraModel.GenericField         `json:"revision_log"`
	Uid               []islandoraModel.EntityReferenceField `json:"uid"`
	Title             []islandoraModel.GenericField         `json:"title"`
	Type              []islandoraModel.ConfigReferenceField `json:"type"`
	Status            []islandoraModel.BoolField            `json:"status"`
	Created           []islandoraModel.GenericField         `json:"created"`
	Changed           []islandoraModel.GenericField         `json:"changed"`

	FieldAbstract              []islandoraModel.GenericField         `json:"field_abstract,omitempty"`
	FieldAltTitle              []islandoraModel.GenericField         `json:"field_alt_title,omitempty"`
	FieldClassification        []islandoraModel.GenericField         `json:"field_classification,omitempty"`
	FieldCoordinates           []islandoraModel.GeoLocationField     `json:"field_coordinates,omitempty"`
	FieldCoordinatesText       []islandoraModel.GenericField         `json:"field_coordinates_text,omitempty"`
	FieldCopyrightDate         []islandoraModel.EdtfField            `json:"field_copyright_date,omitempty"`
	FieldDateCaptured          []islandoraModel.EdtfField            `json:"field_date_captured,omitempty"`
	FieldDateModified          []islandoraModel.EdtfField            `json:"field_date_modified,omitempty"`
	FieldDateValid             []islandoraModel.EdtfField            `json:"field_date_valid,omitempty"`
	FieldDescription           []islandoraModel.GenericField         `json:"field_description,omitempty"`
	FieldDeweyClassification   []islandoraModel.GenericField         `json:"field_dewey_classification,omitempty"`
	FieldEdition               []islandoraModel.GenericField         `json:"field_edition,omitempty"`
	FieldEdtfDate              []islandoraModel.EdtfField            `json:"field_edtf_date,omitempty"`
	FieldEdtfDateCreated       []islandoraModel.EdtfField            `json:"field_edtf_date_created,omitempty"`
	FieldEdtfDateIssued        []islandoraModel.EdtfField            `json:"field_edtf_date_issued,omitempty"`
	FieldExtent                []islandoraModel.GenericField         `json:"field_extent,omitempty"`
	FieldFrequency             []islandoraModel.EntityReferenceField `json:"field_frequency,omitempty"`
	FieldFullTitle             []islandoraModel.GenericField         `json:"field_full_title,omitempty"`
	FieldGenre                 []islandoraModel.EntityReferenceField `json:"field_genre,omitempty"`
	FieldGeographicSubject     []islandoraModel.EntityReferenceField `json:"field_geographic_subject,omitempty"`
	FieldIdentifier            []islandoraModel.GenericField         `json:"field_identifier,omitempty"`
	FieldIsbn                  []islandoraModel.GenericField         `json:"field_isbn,omitempty"`
	FieldLanguage              []islandoraModel.EntityReferenceField `json:"field_language,omitempty"`
	FieldLccClassification     []islandoraModel.GenericField         `json:"field_lcc_classification,omitempty"`
	FieldLinkedAgent           []islandoraModel.TypedRelationField   `json:"field_linked_agent,omitempty"`
	FieldLocalIdentifier       []islandoraModel.GenericField         `json:"field_local_identifier,omitempty"`
	FieldMemberOf              []islandoraModel.EntityReferenceField `json:"field_member_of,omitempty"`
	FieldModeOfIssuance        []islandoraModel.EntityReferenceField `json:"field_mode_of_issuance,omitempty"`
	FieldModel                 []islandoraModel.EntityReferenceField `json:"field_model,omitempty"`
	FieldNote                  []islandoraModel.GenericField         `json:"field_note,omitempty"`
	FieldOclcNumber            []islandoraModel.GenericField         `json:"field_oclc_number,omitempty"`
	FieldPhysicalForm          []islandoraModel.EntityReferenceField `json:"field_physical_form,omitempty"`
	FieldPid                   []islandoraModel.GenericField         `json:"field_pid,omitempty"`
	FieldPlacePublished        []islandoraModel.GenericField         `json:"field_place_published,omitempty"`
	FieldPlacePublishedCountry []islandoraModel.EntityReferenceField `json:"field_place_published_country,omitempty"`
	FieldPublisher             []islandoraModel.GenericField         `json:"field_publisher,omitempty"`
	FieldRepresentativeImage   []islandoraModel.EntityReferenceField `json:"field_representative_image,omitempty"`
	FieldResourceType          []islandoraModel.EntityReferenceField `json:"field_resource_type,omitempty"`
	FieldRights                []islandoraModel.GenericField         `json:"field_rights,omitempty"`
	FieldSubject               []islandoraModel.EntityReferenceField `json:"field_subject,omitempty"`
	FieldSubjectGeneral        []islandoraModel.EntityReferenceField `json:"field_subject_general,omitempty"`
	FieldSubjectsName          []islandoraModel.EntityReferenceField `json:"field_subjects_name,omitempty"`
	FieldTableOfContents       []islandoraModel.GenericField         `json:"field_table_of_contents,omitempty"`
	FieldTemporalSubject       []islandoraModel.EntityReferenceField `json:"field_temporal_subject,omitempty"`
	FieldViewerOverride        []islandoraModel.EntityReferenceField `json:"field_viewer_override,omitempty"`
	FieldWeight                []islandoraModel.IntField             `json:"field_weight,omitempty"`
}
