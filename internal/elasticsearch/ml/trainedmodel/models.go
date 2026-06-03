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
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type trainedModelData struct {
	entitycore.ElasticsearchConnectionField
	ID                   types.String         `tfsdk:"id"`
	ModelID              types.String         `tfsdk:"model_id"`
	Description          types.String         `tfsdk:"description"`
	ModelType            types.String         `tfsdk:"model_type"`
	ModelSizeBytes       types.Int64          `tfsdk:"model_size_bytes"`
	FullyDefined         types.Bool           `tfsdk:"fully_defined"`
	Tags                 types.Set            `tfsdk:"tags"`
	CreateTime           types.String         `tfsdk:"create_time"`
	CreatedBy            types.String         `tfsdk:"created_by"`
	Version              types.String         `tfsdk:"version"`
	PlatformArchitecture types.String         `tfsdk:"platform_architecture"`
	LicenseLevel         types.String         `tfsdk:"license_level"`
	InputJSON            jsontypes.Normalized `tfsdk:"input_json"`
	InferenceConfigJSON  jsontypes.Normalized `tfsdk:"inference_config_json"`
	MetadataJSON         jsontypes.Normalized `tfsdk:"metadata_json"`
	DefaultFieldMap      types.Map            `tfsdk:"default_field_map"`
}
