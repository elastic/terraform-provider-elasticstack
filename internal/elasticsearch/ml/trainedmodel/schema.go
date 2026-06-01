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

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func getDataSourceSchema(_ context.Context) dsschema.Schema {
	return dsschema.Schema{
		MarkdownDescription: "Retrieves an Elasticsearch ML trained model. See, https://www.elastic.co/guide/en/elasticsearch/reference/current/get-trained-models.html",
		Attributes: map[string]dsschema.Attribute{
			"id": dsschema.StringAttribute{
				MarkdownDescription: "Internal identifier of the resource",
				Computed:            true,
			},
			"model_id": dsschema.StringAttribute{
				MarkdownDescription: "The identifier for the trained model.",
				Required:            true,
			},
			"description": dsschema.StringAttribute{
				MarkdownDescription: "The free-text description of the trained model.",
				Computed:            true,
			},
			"model_type": dsschema.StringAttribute{
				MarkdownDescription: "The model type.",
				Computed:            true,
			},
			"model_size_bytes": dsschema.Int64Attribute{
				MarkdownDescription: "The estimated memory usage in bytes to keep the trained model in memory.",
				Computed:            true,
			},
			"fully_defined": dsschema.BoolAttribute{
				MarkdownDescription: "True if the full model definition is present.",
				Computed:            true,
			},
			"tags": dsschema.SetAttribute{
				MarkdownDescription: "A comma delimited string of tags. A trained model can have many tags, or none.",
				ElementType:         types.StringType,
				Computed:            true,
			},
			"create_time": dsschema.StringAttribute{
				MarkdownDescription: "The time when the trained model was created.",
				Computed:            true,
			},
			"created_by": dsschema.StringAttribute{
				MarkdownDescription: "Information on the creator of the trained model.",
				Computed:            true,
			},
			"version": dsschema.StringAttribute{
				MarkdownDescription: "The Elasticsearch version number in which the trained model was created.",
				Computed:            true,
			},
			"platform_architecture": dsschema.StringAttribute{
				MarkdownDescription: "The platform identifier (e.g. linux-x86_64).",
				Computed:            true,
			},
			"license_level": dsschema.StringAttribute{
				MarkdownDescription: "The license level of the trained model.",
				Computed:            true,
			},
			"input_json": dsschema.StringAttribute{
				MarkdownDescription: "JSON string of the model input field names.",
				Computed:            true,
				CustomType:          jsontypes.NormalizedType{},
			},
			"inference_config_json": dsschema.StringAttribute{
				MarkdownDescription: "JSON string of the default inference configuration.",
				Computed:            true,
				CustomType:          jsontypes.NormalizedType{},
			},
			"metadata_json": dsschema.StringAttribute{
				MarkdownDescription: "JSON string of the model metadata.",
				Computed:            true,
				CustomType:          jsontypes.NormalizedType{},
			},
			"default_field_map": dsschema.MapAttribute{
				MarkdownDescription: "Any field map described in the inference configuration takes precedence.",
				ElementType:         types.StringType,
				Computed:            true,
			},
		},
	}
}
