package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"golang.org/x/exp/slog"
	"gopkg.in/yaml.v2"
)

type Field struct {
	Name        string
	Type        string
	MachineName string
}

type StructData struct {
	StructName string
	Fields     []Field
}

func main() {
	nodeCexYaml := flag.String("node-cex-yaml", "", "Path to the node CEX YAML file")
	output := flag.String("output", "./generated_structs.go", "Output file for generated structs")
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

	fields := []Field{}

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

		fieldTypeGo := mapFieldTypeToGoType(fieldType)
		fields = append(fields, Field{
			Name:        toCamelCase(fieldName),
			Type:        fieldTypeGo,
			MachineName: fieldName,
		})
	}

	structName := toCamelCase(nodeType)
	structData := StructData{
		StructName: structName,
		Fields:     fields,
	}

	structCode, err := generateGoStruct(structData)
	if err != nil {
		slog.Error("Error generating Go struct: %s", err)
		os.Exit(1)
	}

	err = os.WriteFile(*output, []byte(structCode), 0644)
	if err != nil {
		slog.Error("Error writing output file: %s", err)
		os.Exit(1)
	}

	slog.Info("Structs generated and written", "file", *output)
}

func generateGoStruct(data StructData) (string, error) {
	tmpl, err := template.ParseFiles("node.go.tmpl")
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
