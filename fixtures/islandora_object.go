package islandora

import (
	islandoraModel "github.com/lehigh-university-libraries/go-islandora/model"
)

type IslandoraObject struct {
	Nid               []islandoraModel.IntField             `json:"nid" csv:"node_id"`
	Vid               []islandoraModel.IntField             `json:"vid" csv:"-"`
	Uuid              []islandoraModel.GenericField         `json:"uuid" csv:"-"`
	Language          []islandoraModel.GenericField         `json:"lang" csv:"langcode"`
	RevisionTimestamp []islandoraModel.GenericField         `json:"revision_timestamp" csv:"-"`
	RevisionUid       []islandoraModel.EntityReferenceField `json:"revision_uid" csv:"-"`
	RevisionLog       []islandoraModel.GenericField         `json:"revision_log" csv:"-"`
	Uid               []islandoraModel.EntityReferenceField `json:"uid" csv:"-"`
	Title             []islandoraModel.GenericField         `json:"title" csv:"title"`
	Type              []islandoraModel.ConfigReferenceField `json:"type" csv:"-"`
	Status            []islandoraModel.BoolField            `json:"status" csv:"published"`
	Created           []islandoraModel.GenericField         `json:"created" csv:"-"`
	Changed           []islandoraModel.GenericField         `json:"changed" csv:"-"`

	FieldAbstract              []islandoraModel.GenericField         `json:"field_abstract,omitempty" csv:"field_abstract"`
	FieldAltTitle              []islandoraModel.GenericField         `json:"field_alt_title,omitempty" csv:"field_alt_title"`
	FieldClassification        []islandoraModel.GenericField         `json:"field_classification,omitempty" csv:"field_classification"`
	FieldCoordinates           []islandoraModel.GeoLocationField     `json:"field_coordinates,omitempty" csv:"field_coordinates"`
	FieldCoordinatesText       []islandoraModel.GenericField         `json:"field_coordinates_text,omitempty" csv:"field_coordinates_text"`
	FieldCopyrightDate         []islandoraModel.EdtfField            `json:"field_copyright_date,omitempty" csv:"field_copyright_date"`
	FieldDateCaptured          []islandoraModel.EdtfField            `json:"field_date_captured,omitempty" csv:"field_date_captured"`
	FieldDateModified          []islandoraModel.EdtfField            `json:"field_date_modified,omitempty" csv:"field_date_modified"`
	FieldDateValid             []islandoraModel.EdtfField            `json:"field_date_valid,omitempty" csv:"field_date_valid"`
	FieldDescription           []islandoraModel.GenericField         `json:"field_description,omitempty" csv:"field_description"`
	FieldDeweyClassification   []islandoraModel.GenericField         `json:"field_dewey_classification,omitempty" csv:"field_dewey_classification"`
	FieldEdition               []islandoraModel.GenericField         `json:"field_edition,omitempty" csv:"field_edition"`
	FieldEdtfDate              []islandoraModel.EdtfField            `json:"field_edtf_date,omitempty" csv:"field_edtf_date"`
	FieldEdtfDateCreated       []islandoraModel.EdtfField            `json:"field_edtf_date_created,omitempty" csv:"field_edtf_date_created"`
	FieldEdtfDateIssued        []islandoraModel.EdtfField            `json:"field_edtf_date_issued,omitempty" csv:"field_edtf_date_issued"`
	FieldExtent                []islandoraModel.GenericField         `json:"field_extent,omitempty" csv:"field_extent"`
	FieldFrequency             []islandoraModel.EntityReferenceField `json:"field_frequency,omitempty" csv:"field_frequency"`
	FieldFullTitle             []islandoraModel.GenericField         `json:"field_full_title,omitempty" csv:"field_full_title"`
	FieldGenre                 []islandoraModel.EntityReferenceField `json:"field_genre,omitempty" csv:"field_genre"`
	FieldGeographicSubject     []islandoraModel.EntityReferenceField `json:"field_geographic_subject,omitempty" csv:"field_geographic_subject"`
	FieldIdentifier            []islandoraModel.GenericField         `json:"field_identifier,omitempty" csv:"field_identifier"`
	FieldIsbn                  []islandoraModel.GenericField         `json:"field_isbn,omitempty" csv:"field_isbn"`
	FieldLanguage              []islandoraModel.EntityReferenceField `json:"field_language,omitempty" csv:"field_language"`
	FieldLccClassification     []islandoraModel.GenericField         `json:"field_lcc_classification,omitempty" csv:"field_lcc_classification"`
	FieldLinkedAgent           []islandoraModel.TypedRelationField   `json:"field_linked_agent,omitempty" csv:"field_linked_agent"`
	FieldLocalIdentifier       []islandoraModel.GenericField         `json:"field_local_identifier,omitempty" csv:"field_local_identifier"`
	FieldMemberOf              []islandoraModel.EntityReferenceField `json:"field_member_of,omitempty" csv:"field_member_of"`
	FieldModeOfIssuance        []islandoraModel.EntityReferenceField `json:"field_mode_of_issuance,omitempty" csv:"field_mode_of_issuance"`
	FieldModel                 []islandoraModel.EntityReferenceField `json:"field_model,omitempty" csv:"field_model"`
	FieldNote                  []islandoraModel.GenericField         `json:"field_note,omitempty" csv:"field_note"`
	FieldOclcNumber            []islandoraModel.GenericField         `json:"field_oclc_number,omitempty" csv:"field_oclc_number"`
	FieldPhysicalForm          []islandoraModel.EntityReferenceField `json:"field_physical_form,omitempty" csv:"field_physical_form"`
	FieldPid                   []islandoraModel.GenericField         `json:"field_pid,omitempty" csv:"field_pid"`
	FieldPlacePublished        []islandoraModel.GenericField         `json:"field_place_published,omitempty" csv:"field_place_published"`
	FieldPlacePublishedCountry []islandoraModel.EntityReferenceField `json:"field_place_published_country,omitempty" csv:"field_place_published_country"`
	FieldPublisher             []islandoraModel.GenericField         `json:"field_publisher,omitempty" csv:"field_publisher"`
	FieldRepresentativeImage   []islandoraModel.EntityReferenceField `json:"field_representative_image,omitempty" csv:"field_representative_image"`
	FieldResourceType          []islandoraModel.EntityReferenceField `json:"field_resource_type,omitempty" csv:"field_resource_type"`
	FieldRights                []islandoraModel.GenericField         `json:"field_rights,omitempty" csv:"field_rights"`
	FieldSubject               []islandoraModel.EntityReferenceField `json:"field_subject,omitempty" csv:"field_subject"`
	FieldSubjectGeneral        []islandoraModel.EntityReferenceField `json:"field_subject_general,omitempty" csv:"field_subject_general"`
	FieldSubjectsName          []islandoraModel.EntityReferenceField `json:"field_subjects_name,omitempty" csv:"field_subjects_name"`
	FieldTableOfContents       []islandoraModel.GenericField         `json:"field_table_of_contents,omitempty" csv:"field_table_of_contents"`
	FieldTemporalSubject       []islandoraModel.EntityReferenceField `json:"field_temporal_subject,omitempty" csv:"field_temporal_subject"`
	FieldViewerOverride        []islandoraModel.EntityReferenceField `json:"field_viewer_override,omitempty" csv:"field_viewer_override"`
	FieldWeight                []islandoraModel.IntField             `json:"field_weight,omitempty" csv:"field_weight"`
}
