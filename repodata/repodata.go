// Parse Conda metadata JSON files
// https://github.com/conda/schemas
package repodata

import (
	"bytes"
	"encoding/json"
	"reflect"
	"strings"
)

// https://github.com/conda/schemas/blob/bd2b05d6a6314b39d9a8c9c9802280c3eb78e788/common-1.schema.json

// https://github.com/conda/schemas/blob/bd2b05d6a6314b39d9a8c9c9802280c3eb78e788/repodata-record-1.schema.json
type RepodataRecord struct {
	// Required fields
	Subdir      string `json:"subdir"`
	Name        string `json:"name"`
	Version     string `json:"version"`
	BuildNumber int    `json:"build_number"`
	Build       string `json:"build"`
	// Schema says fn is required, but it isn't included in repodata.json
	// Fn          string `json:"fn"`
	Sha256 string `json:"sha256"`
	Size   int    `json:"size"`
	Md5    string `json:"md5"`

	// Optional fields
	Depends []string `json:"depends,omitempty"`

	// Extra fields, keep so we can serialise back to JSON
	Extra map[string]interface{} `json:"-"`
}

type RepodataInfo struct {
	Subdir string `json:"subdir"`
}

// https://github.com/conda/schemas/blob/bd2b05d6a6314b39d9a8c9c9802280c3eb78e788/repodata-1.schema.json
type Repodata struct {
	RepodataVersion int          `json:"repodata_version"`
	Info            RepodataInfo `json:"info"`
	// Package filename are the keys
	// ^.+\.tar\.bz2$
	Packages map[string]RepodataRecord `json:"packages"`
	// ^.+\.conda$
	PackagesConda map[string]RepodataRecord `json:"packages.conda"`
}

// EncodeJSON converts a value to a JSON byte array, without escaping HTML
func EncodeJSON(v any, indent string) ([]byte, error) {
	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", indent)
	if err := encoder.Encode(v); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

type _RepodataRecord RepodataRecord

// MarshalJSON marshals a RepodataRecord to JSON, including Extra fields
func (t RepodataRecord) MarshalJSON() ([]byte, error) {
	data := make(map[string]interface{})

	// Take everything in Extra
	for k, v := range t.Extra {
		data[k] = v
	}

	// Take all the struct values with a json tag
	val := reflect.ValueOf(t)
	typ := reflect.TypeOf(t)
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		fieldv := val.Field(i)
		jsonTag := strings.Split(field.Tag.Get("json"), ",")[0]
		if jsonTag != "" && jsonTag != "-" {
			data[jsonTag] = fieldv.Interface()
		}
	}

	return EncodeJSON(data, " ")
}

// UnmarshalJSON unmarshals a RepodataRecord from JSON, including Extra fields
func (t *RepodataRecord) UnmarshalJSON(b []byte) error {
	t2 := _RepodataRecord{}
	err := json.Unmarshal(b, &t2)
	if err != nil {
		return err
	}

	err = json.Unmarshal(b, &(t2.Extra))
	if err != nil {
		return err
	}

	typ := reflect.TypeOf(t2)
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		jsonTag := strings.Split(field.Tag.Get("json"), ",")[0]
		if jsonTag != "" && jsonTag != "-" {
			delete(t2.Extra, jsonTag)
		}
	}

	*t = RepodataRecord(t2)

	return nil
}
