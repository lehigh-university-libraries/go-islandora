package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	"golang.org/x/exp/slog"
	"gopkg.in/yaml.v2"
)

type Field struct {
	Name           string
	Type           string
	Title          string
	Description    string
	MachineName    string
	Required       bool
	OapiProperties map[string]string

	// see https://github.com/oapi-codegen/oapi-codegen?tab=readme-ov-file#openapi-extensions
	GoType     string
	TypeImport TypeImport
}

type StructData struct {
	StructName string
	Fields     []Field
}

type TypeImport struct {
	Path string
	Name string
}

func main() {
	nodeCexYaml := flag.String("node-cex-yaml", "", "Path to the node config export YAML file")
	output := flag.String("output", "./api.yaml", "Output file for generated Open API spec")
	flag.Parse()

	if *nodeCexYaml == "" {
		slog.Error("The --node-cex-yaml flag is required")
		os.Exit(1)
	}

	dir := filepath.Dir(*nodeCexYaml)
	baseName := filepath.Base(*nodeCexYaml)
	nodeType := strings.TrimSuffix(strings.TrimPrefix(baseName, "node.type."), ".yml")

	pattern := fmt.Sprintf("field.field.node.%s.*", nodeType)

	files, err := filepath.Glob(filepath.Join(dir, pattern))
	if err != nil {
		slog.Error("Error scanning directory: %v", err)
		os.Exit(1)
	}

	if len(files) == 0 {
		slog.Error("No files found matching pattern: %s", pattern)
		os.Exit(1)
	}

	fields := nodeFields()

	for _, file := range files {
		yamlFile, err := os.ReadFile(file)
		if err != nil {
			slog.Error("Error reading YAML file: %s", err)
			os.Exit(1)
		}

		var data map[string]interface{}
		err = yaml.Unmarshal(yamlFile, &data)
		if err != nil {
			slog.Error("Error unmarshalling YAML: %s", err)
			os.Exit(1)
		}

		fieldName := data["field_name"].(string)
		fieldType := data["field_type"].(string)
		fields = append(fields, Field{
			Name:           toCamelCase(fieldName),
			OapiProperties: mapFieldTypeToOapiProperties(fieldType),
			Title:          data["label"].(string),
			Description:    strings.ReplaceAll(data["description"].(string), `"`, `\"`),
			MachineName:    fieldName,
			Required:       data["required"].(bool),
			GoType:         mapFieldTypeToGoType(fieldType),
			TypeImport: TypeImport{
				Path: "github.com/lehigh-university-libraries/go-islandora/model",
				Name: "islandoraModel",
			},
		})
	}

	structName := toCamelCase(nodeType)
	structData := StructData{
		StructName: structName,
		Fields:     fields,
	}

	structCode, err := generateOapiSpec(structData)
	if err != nil {
		slog.Error("Error generating Go struct: %s", err)
		os.Exit(1)
	}

	err = os.WriteFile(*output, []byte(structCode), 0644)
	if err != nil {
		slog.Error("Error writing output file: %s", err)
		os.Exit(1)
	}

	slog.Info("Open API Spec generated and written", "file", *output)
	cmd := exec.Command("go", "generate", "./api")
	if err := cmd.Run(); err != nil {
		slog.Error("Unable to run go generate", "err", err)
		os.Exit(1)
	}
	slog.Info("Structs generated in api/islandora.gen.go")

}

func generateOapiSpec(data StructData) (string, error) {
	tmpl, err := template.ParseFiles("api.yaml.tmpl")
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

func mapFieldTypeToOapiProperties(fieldType string) map[string]string {
	properties := map[string]string{}
	switch fieldType {
	case "boolean":
		properties["value"] = "boolean"
	case "entity_reference":
		properties["target_id"] = "integer"
	case "integer":
		properties["value"] = "integer"
	case "geolocation":
		properties["lat"] = "number"
		properties["lng"] = "number"
	case "hierarchical_geographic":
		properties["continent"] = "string"
		properties["country"] = "string"
		properties["region"] = "string"
		properties["state"] = "string"
		properties["territory"] = "string"
		properties["county"] = "string"
		properties["city"] = "string"
		properties["city_section"] = "string"
		properties["island"] = "string"
		properties["area"] = "string"
		properties["extraterrestrial_area"] = "string"
	case "typed_relation":
		properties["rel_type"] = "string"
		properties["target_id"] = "integer"
	case "related_item":
		properties["identifier"] = "string"
		properties["identifier_type"] = "string"
		properties["number"] = "string"
		properties["title"] = "string"
	case "textfield_attr", "textarea_attr":
		properties["attr0"] = "string"
		properties["attr1"] = "string"
		properties["value"] = "string"
		if fieldType == "textarea_attr" {
			properties["format"] = "string"
		}
	case "part_detail":
		properties["type"] = "string"
		properties["caption"] = "string"
		properties["number"] = "string"
		properties["title"] = "string"
	case "config_reference":
		properties["target_id"] = "string"
	default:
		properties["value"] = "string"
	}

	return properties
}

func mapFieldTypeToGoType(fieldType string) string {
	switch fieldType {
	case "boolean":
		return "islandoraModel.BoolField"
	case "entity_reference":
		return "islandoraModel.EntityReferenceField"
	case "edtf":
		return "islandoraModel.EdtfField"
	case "email":
		return "islandoraModel.EmailField"
	case "integer":
		return "islandoraModel.IntField"
	case "geolocation":
		return "islandoraModel.GeoLocationField"
	case "hierarchical_geographic":
		return "islandoraModel.HierarchicalGeographicField"
	case "typed_relation":
		return "islandoraModel.TypedRelationField"
	case "related_item":
		return "islandoraModel.RelatedItemField"
	case "textfield_attr", "textarea_attr":
		return "islandoraModel.TypedTextField"
	case "part_detail":
		return "islandoraModel.PartDetailField"
	case "config_reference":
		return "islandoraModel.ConfigReferenceField"
	default:
		return "islandoraModel.GenericField"
	}
}

func toCamelCase(input string) string {
	output := ""
	capitalizeNext := true
	for _, ch := range input {
		if ch == '_' || ch == '-' {
			capitalizeNext = true
		} else if capitalizeNext {
			output += strings.ToUpper(string(ch))
			capitalizeNext = false
		} else {
			output += string(ch)
		}
	}
	return output
}

// some base properties for the node entity
func nodeFields() []Field {
	fields := []Field{}
	f := map[string]string{
		"nid":                "integer",
		"vid":                "integer",
		"uuid":               "string",
		"language":           "string",
		"revision_timestamp": "string",
		"revision_uid":       "entity_reference",
		"revision_log":       "string",
		"uid":                "entity_reference",
		"title":              "string",
		"type":               "config_reference",
		"status":             "boolean",
		"created":            "string",
		"changed":            "string",
	}
	for fieldName, fieldType := range f {
		fields = append(fields, Field{
			Name:           toCamelCase(fieldName),
			OapiProperties: mapFieldTypeToOapiProperties(fieldType),
			Title:          toCamelCase(fieldName),
			Description:    "",
			MachineName:    fieldName,
			GoType:         mapFieldTypeToGoType(fieldType),
			TypeImport: TypeImport{
				Path: "github.com/lehigh-university-libraries/go-islandora/model",
				Name: "islandoraModel",
			},
		})
	}

	return fields
}
