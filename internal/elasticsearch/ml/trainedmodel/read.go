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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	estypes "github.com/elastic/go-elasticsearch/v9/typedapi/types"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func readDataSource(ctx context.Context, esClient *clients.ElasticsearchScopedClient, config trainedModelData) (trainedModelData, diag.Diagnostics) {
	var diags diag.Diagnostics

	modelID := config.ModelID.ValueString()

	// Resolve the composite ID
	id, idDiags := esClient.ID(ctx, modelID)
	diags.Append(idDiags...)
	if diags.HasError() {
		return config, diags
	}
	config.ID = types.StringValue(id.String())

	// Call GetTrainedModel
	model, found, modelDiags := elasticsearch.GetTrainedModel(ctx, esClient, modelID)
	diags.Append(modelDiags...)
	if diags.HasError() {
		return config, diags
	}

	// Not-found: return empty ID with computed attributes null
	if !found || model == nil {
		config.ID = types.StringValue("")
		config.Description = types.StringNull()
		config.ModelType = types.StringNull()
		config.ModelSizeBytes = types.Int64Null()
		config.FullyDefined = types.BoolNull()
		config.Tags = types.SetNull(types.StringType)
		config.CreateTime = types.StringNull()
		config.CreatedBy = types.StringNull()
		config.Version = types.StringNull()
		config.PlatformArchitecture = types.StringNull()
		config.LicenseLevel = types.StringNull()
		config.InputJSON = jsontypes.NewNormalizedNull()
		config.InferenceConfigJSON = jsontypes.NewNormalizedNull()
		config.MetadataJSON = jsontypes.NewNormalizedNull()
		config.DefaultFieldMap = types.MapNull(types.StringType)
		return config, diags
	}

	// Map API response to model
	diags.Append(mapTrainedModelConfig(ctx, model, &config)...)
	if diags.HasError() {
		return config, diags
	}

	return config, diags
}

func mapTrainedModelConfig(ctx context.Context, model *estypes.TrainedModelConfig, data *trainedModelData) diag.Diagnostics {
	var diags diag.Diagnostics

	data.Description = typeutils.NonEmptyStringOrNull(model.Description)
	if model.ModelType != nil {
		data.ModelType = types.StringValue(model.ModelType.String())
	} else {
		data.ModelType = types.StringNull()
	}
	data.ModelSizeBytes = byteSizeToInt64Value(model.ModelSizeBytes)
	data.FullyDefined = types.BoolPointerValue(model.FullyDefined)

	// Tags
	if model.Tags == nil {
		data.Tags = types.SetNull(types.StringType)
	} else {
		tagsSet, d := types.SetValueFrom(ctx, types.StringType, model.Tags)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}
		data.Tags = tagsSet
	}

	data.CreateTime = dateTimeToStringValue(model.CreateTime)
	data.CreatedBy = typeutils.NonEmptyStringOrNull(model.CreatedBy)
	data.Version = typeutils.NonEmptyStringOrNull(model.Version)
	data.PlatformArchitecture = typeutils.NonEmptyStringOrNull(model.PlatformArchitecture)
	data.LicenseLevel = typeutils.NonEmptyStringOrNull(model.LicenseLevel)

	// InputJSON
	inputJSON, d := marshalInputToJSON(model.Input)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}
	if inputJSON != "" {
		data.InputJSON = jsontypes.NewNormalizedValue(inputJSON)
	} else {
		data.InputJSON = jsontypes.NewNormalizedNull()
	}

	// InferenceConfigJSON
	if model.InferenceConfig != nil {
		infBytes, err := json.Marshal(model.InferenceConfig)
		if err != nil {
			diags.AddError("JSON Marshal Error", fmt.Sprintf("Error marshaling inference_config JSON: %s", err))
			return diags
		}
		if !bytes.Equal(infBytes, []byte("null")) {
			data.InferenceConfigJSON = jsontypes.NewNormalizedValue(string(infBytes))
		} else {
			data.InferenceConfigJSON = jsontypes.NewNormalizedNull()
		}
	} else {
		data.InferenceConfigJSON = jsontypes.NewNormalizedNull()
	}

	// MetadataJSON
	if model.Metadata != nil {
		metaBytes, err := json.Marshal(model.Metadata)
		if err != nil {
			diags.AddError("JSON Marshal Error", fmt.Sprintf("Error marshaling metadata JSON: %s", err))
			return diags
		}
		if !bytes.Equal(metaBytes, []byte("null")) {
			data.MetadataJSON = jsontypes.NewNormalizedValue(string(metaBytes))
		} else {
			data.MetadataJSON = jsontypes.NewNormalizedNull()
		}
	} else {
		data.MetadataJSON = jsontypes.NewNormalizedNull()
	}

	// DefaultFieldMap
	if len(model.DefaultFieldMap) == 0 {
		data.DefaultFieldMap = types.MapNull(types.StringType)
	} else {
		fmMap, d := types.MapValueFrom(ctx, types.StringType, model.DefaultFieldMap)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}
		data.DefaultFieldMap = fmMap
	}

	return diags
}

func byteSizeToInt64Value(b estypes.ByteSize) types.Int64 {
	if b == nil {
		return types.Int64Null()
	}
	switch v := b.(type) {
	case int64:
		if v == 0 {
			return types.Int64Null()
		}
		return types.Int64Value(v)
	case int:
		if v == 0 {
			return types.Int64Null()
		}
		return types.Int64Value(int64(v))
	case float64:
		if v == 0 {
			return types.Int64Null()
		}
		return types.Int64Value(int64(v))
	case string:
		if v == "" || v == "0" {
			return types.Int64Null()
		}
		// Try to parse as int64
		i, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return types.Int64Null()
		}
		return types.Int64Value(i)
	default:
		return types.Int64Null()
	}
}

func dateTimeToStringValue(dt estypes.DateTime) types.String {
	return typeutils.ElasticDateTimeToStringValue(dt)
}

func marshalInputToJSON(input estypes.TrainedModelConfigInput) (string, diag.Diagnostics) {
	var diags diag.Diagnostics

	// Treat effectively empty input as null
	if len(input.FieldNames) == 0 {
		return "", diags
	}

	bytes, err := json.Marshal(input)
	if err != nil {
		diags.AddError("JSON Marshal Error", fmt.Sprintf("Error marshaling input JSON: %s", err))
		return "", diags
	}

	return string(bytes), diags
}
