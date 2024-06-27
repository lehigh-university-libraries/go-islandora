package main

import (
	"encoding/json"
	"io"
	"os"
	"testing"

	"github.com/lehigh-university-libraries/go-islandora/islandora"

	"github.com/gocarina/gocsv"
	"github.com/stretchr/testify/assert"
)

func SaveJSON(filename string, v interface{}) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	byteValue, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}

	_, err = file.Write(byteValue)
	return err
}

func SaveCsv(filename string, o islandora.IslandoraObject) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	row := []islandora.IslandoraObject{
		o,
	}
	err = gocsv.MarshalFile(row, f)

	return err
}

// LoadJSON loads a JSON file and unmarshals it into the given interface.
func LoadJSON(filename string, v interface{}) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	byteValue, err := io.ReadAll(file)
	if err != nil {
		return err
	}

	return json.Unmarshal(byteValue, v)
}

// TestLoadJSON tests loading JSON into the generated struct.
func TestLoadJSON(t *testing.T) {
	var node islandora.IslandoraObject

	err := LoadJSON("fixtures/node.json", &node)
	if err != nil {
		t.Fatalf("Failed to load JSON: %v", err)
	}

	assert.Equal(t, "Digital files found in the UTSC Library's Digital Collections are meant for research and private study used in compliance with copyright legislation. Access to digital images and text found on this website and the technical capacity to download or copy it does not imply permission to re-use. Prior written permission to publish, or otherwise use images and text found on the website must be obtained from the copyright holder. Please contact UTSC Library for further information.", node.FieldRights[0].Value)
	assert.Equal(t, "University of Toronto Scarborough", node.FieldPublisher[0].Value)
	assert.Equal(t, 4, node.FieldMemberOf[0].TargetId)
	assert.Equal(t, "1999", node.FieldEdtfDateIssued[0].Value)
	assert.Equal(t, "islandora_object", node.Type[0].TargetId)
}

func TestSaveJson(t *testing.T) {
	var node islandora.IslandoraObject

	err := LoadJSON("fixtures/node.json", &node)
	if err != nil {
		t.Fatalf("Failed to load JSON: %v", err)
	}

	err = SaveJSON("output.json", node)
	if err != nil {
		t.Fatalf("Failed to save JSON: %v", err)
	}
}

func TestSaveCsv(t *testing.T) {
	var node islandora.IslandoraObject

	err := LoadJSON("fixtures/node.json", &node)
	if err != nil {
		t.Fatalf("Failed to load JSON: %v", err)
	}

	err = SaveCsv("fixtures/node.csv", node)
	if err != nil {
		t.Fatalf("Failed to save CSV: %v", err)
	}

}
