// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package trainedmodel

import (
	"context"
	"testing"

	estypes "github.com/elastic/go-elasticsearch/v9/typedapi/types"
	"github.com/elastic/go-elasticsearch/v9/typedapi/types/enums/trainedmodeltype"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMapTrainedModelConfig_fullResponse(t *testing.T) {
	ctx := context.Background()

	model := &estypes.TrainedModelConfig{
		ModelId:              "test-model",
		Description:          new("A test model"),
		ModelType:            &trainedmodeltype.Langident,
		ModelSizeBytes:       int64(1024),
		FullyDefined:         new(true),
		Tags:                 []string{"nlp", "test"},
		CreateTime:           estypes.DateTime("2024-06-01T12:00:00.000Z"),
		CreatedBy:            new("elastic"),
		Version:              new("8.19.0"),
		PlatformArchitecture: new("linux-x86_64"),
		LicenseLevel:         new("platinum"),
		Input: estypes.TrainedModelConfigInput{
			FieldNames: []string{"text_field"},
		},
		InferenceConfig: &estypes.InferenceConfigCreateContainer{
			TextClassification: &estypes.TextClassificationInferenceOptions{
				NumTopClasses: new(5),
			},
		},
		Metadata: &estypes.TrainedModelConfigMetadata{
			ModelAliases: []string{"alias1"},
		},
		DefaultFieldMap: map[string]string{"src": "dest"},
	}

	var data trainedModelData
	diags := mapTrainedModelConfig(ctx, model, &data)
	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)

	assert.Equal(t, types.StringValue("A test model"), data.Description)
	assert.Equal(t, types.StringValue("lang_ident"), data.ModelType)
	assert.Equal(t, types.Int64Value(1024), data.ModelSizeBytes)
	assert.Equal(t, types.BoolValue(true), data.FullyDefined)
	assert.Equal(t, types.StringValue("2024-06-01T12:00:00.000Z"), data.CreateTime)
	assert.Equal(t, types.StringValue("elastic"), data.CreatedBy)
	assert.Equal(t, types.StringValue("8.19.0"), data.Version)
	assert.Equal(t, types.StringValue("linux-x86_64"), data.PlatformArchitecture)
	assert.Equal(t, types.StringValue("platinum"), data.LicenseLevel)

	assert.False(t, data.InputJSON.IsNull())
	assert.JSONEq(t, `{"field_names":["text_field"]}`, data.InputJSON.ValueString())

	assert.False(t, data.InferenceConfigJSON.IsNull())
	assert.Contains(t, data.InferenceConfigJSON.ValueString(), "text_classification")

	assert.False(t, data.MetadataJSON.IsNull())
	assert.Contains(t, data.MetadataJSON.ValueString(), "alias1")

	var tags []string
	diags = data.Tags.ElementsAs(ctx, &tags, false)
	require.False(t, diags.HasError())
	assert.ElementsMatch(t, []string{"nlp", "test"}, tags)

	var dfm map[string]string
	diags = data.DefaultFieldMap.ElementsAs(ctx, &dfm, false)
	require.False(t, diags.HasError())
	assert.Equal(t, map[string]string{"src": "dest"}, dfm)
}

func TestMapTrainedModelConfig_nilFields(t *testing.T) {
	ctx := context.Background()

	model := &estypes.TrainedModelConfig{
		ModelId: "minimal-model",
		Tags:    nil,
		Input: estypes.TrainedModelConfigInput{
			FieldNames: nil,
		},
	}

	var data trainedModelData
	diags := mapTrainedModelConfig(ctx, model, &data)
	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)

	assert.True(t, data.Description.IsNull())
	assert.True(t, data.ModelType.IsNull())
	assert.True(t, data.ModelSizeBytes.IsNull())
	assert.True(t, data.FullyDefined.IsNull())
	assert.True(t, data.CreateTime.IsNull())
	assert.True(t, data.CreatedBy.IsNull())
	assert.True(t, data.Version.IsNull())
	assert.True(t, data.PlatformArchitecture.IsNull())
	assert.True(t, data.LicenseLevel.IsNull())
	assert.True(t, data.Tags.IsNull())
	assert.True(t, data.InputJSON.IsNull())
	assert.True(t, data.InferenceConfigJSON.IsNull())
	assert.True(t, data.MetadataJSON.IsNull())
	assert.True(t, data.DefaultFieldMap.IsNull())
}

func TestMapTrainedModelConfig_emptyInputIsNull(t *testing.T) {
	ctx := context.Background()

	model := &estypes.TrainedModelConfig{
		ModelId: "empty-input-model",
		Input: estypes.TrainedModelConfigInput{
			FieldNames: []string{},
		},
	}

	var data trainedModelData
	diags := mapTrainedModelConfig(ctx, model, &data)
	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)

	assert.True(t, data.InputJSON.IsNull(), "expected input_json to be null for empty input")
}

func TestMapTrainedModelConfig_emptyTagsSet(t *testing.T) {
	ctx := context.Background()

	model := &estypes.TrainedModelConfig{
		ModelId: "empty-tags-model",
		Tags:    []string{},
	}

	var data trainedModelData
	diags := mapTrainedModelConfig(ctx, model, &data)
	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)

	assert.False(t, data.Tags.IsNull(), "expected tags to be non-null empty set")
	assert.Empty(t, data.Tags.Elements())
}

func TestMapTrainedModelConfig_emptyDefaultFieldMap(t *testing.T) {
	ctx := context.Background()

	model := &estypes.TrainedModelConfig{
		ModelId:         "empty-dfm-model",
		DefaultFieldMap: map[string]string{},
	}

	var data trainedModelData
	diags := mapTrainedModelConfig(ctx, model, &data)
	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)

	assert.True(t, data.DefaultFieldMap.IsNull(), "expected default_field_map to be null for empty map")
}

func TestByteSizeToInt64Value(t *testing.T) {
	tests := []struct {
		name     string
		input    estypes.ByteSize
		expected types.Int64
	}{
		{"nil", nil, types.Int64Null()},
		{"int64 zero", int64(0), types.Int64Null()},
		{"int64 non-zero", int64(1024), types.Int64Value(1024)},
		{"int zero", int(0), types.Int64Null()},
		{"int non-zero", int(2048), types.Int64Value(2048)},
		{"float64 zero", float64(0), types.Int64Null()},
		{"float64 non-zero", float64(4096), types.Int64Value(4096)},
		{"string zero", "0", types.Int64Null()},
		{"string empty", "", types.Int64Null()},
		{"string non-zero", "8192", types.Int64Value(8192)},
		{"string invalid", "not-a-number", types.Int64Null()},
		{"unknown type", true, types.Int64Null()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := byteSizeToInt64Value(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDateTimeToStringValue(t *testing.T) {
	tests := []struct {
		name     string
		input    estypes.DateTime
		expected types.String
	}{
		{"nil", nil, types.StringNull()},
		{"zero float64", estypes.DateTime(float64(0)), types.StringNull()},
		{"non-zero float64 epoch", estypes.DateTime(float64(1717243200000)), types.StringValue("2024-06-01T12:00:00.000Z")},
		{"RFC3339 string", estypes.DateTime("2024-06-01T12:00:00Z"), types.StringValue("2024-06-01T12:00:00.000Z")},
		{"empty string", estypes.DateTime(""), types.StringNull()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := dateTimeToStringValue(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
