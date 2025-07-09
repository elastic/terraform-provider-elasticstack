package geminikbapi

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os"
	"reflect"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/yaml"
	yml3 "github.com/oasdiff/yaml3"
	"github.com/stretchr/testify/require"
)

func Test_YAML(t *testing.T) {
	yamlData, err := os.ReadFile("./api-spec.yaml")
	require.NoError(t, err)

	dec := yml3.NewDecoder(bytes.NewReader(yamlData))
	dec.Origin(false)

	// Convert the YAML to an object.
	var yamlObj interface{}
	if err := dec.Decode(&yamlObj); err != nil {
		// Functionality changed in v3 which means we need to ignore EOF error.
		// See https://github.com/go-yaml/yaml/issues/639
		if !errors.Is(err, io.EOF) {
			require.NoError(t, err)
		}
	}

	// YAML objects are not completely compatible with JSON objects (e.g. you
	// can have non-string keys in YAML). So, convert the YAML-compatible object
	// to a JSON-compatible object, failing with an error if irrecoverable
	// incompatibilities happen along the way.
	jsonTarget := reflect.ValueOf(yamlObj)
	jsonObj, err := convertToJSONableObject(yamlObj, &jsonTarget)
	require.NoError(t, err)

	// Convert this object to JSON and return the data.
	j, err := json.Marshal(jsonObj)
	_ = os.WriteFile("./api-spec.json", j, os.ModeAppend)

	doc := &openapi3.T{}
	err = yaml.UnmarshalWithOrigin(yamlData, doc, false)
	require.NoError(t, err)

	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true
	spec, err := loader.LoadFromData(yamlData)

	require.NoError(t, err)
	require.NotEmpty(t, spec)
}
