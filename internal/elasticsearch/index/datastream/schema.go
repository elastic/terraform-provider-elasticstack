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

package datastream

import (
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

const (
	dataStreamNameAllowedCharsError = "must contain lower case alphanumeric characters and selected punctuation, see: " +
		"https://www.elastic.co/guide/en/elasticsearch/reference/current/indices-create-data-stream.html" +
		"#indices-create-data-stream-api-path-params"
)

// GetSchema returns the Plugin Framework schema for the data stream resource.
// The elasticsearch_connection block is injected by the entitycore envelope.
func GetSchema() schema.Schema {
	return schema.Schema{
		MarkdownDescription: "Managing Elasticsearch data streams, see: https://www.elastic.co/guide/en/elasticsearch/reference/current/data-stream-apis.html",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Internal identifier of the resource",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the data stream to create.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 255),
					stringvalidator.NoneOf(".", ".."),
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^[^-_+]`),
						"cannot start with -, _, +",
					),
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^[a-z0-9!$%&'()+.;=@[\]^{}~_-]+$`),
						dataStreamNameAllowedCharsError,
					),
				},
			},
			"timestamp_field": schema.StringAttribute{
				MarkdownDescription: "Contains information about the data stream's @timestamp field.",
				Computed:            true,
			},
			"indices": schema.ListNestedAttribute{
				MarkdownDescription: "Array of objects containing information about the data stream's backing indices. The last item in this array contains information about the stream's current write index.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"index_name": schema.StringAttribute{
							MarkdownDescription: "Name of the backing index.",
							Computed:            true,
						},
						"index_uuid": schema.StringAttribute{
							MarkdownDescription: "Universally unique identifier (UUID) for the index.",
							Computed:            true,
						},
					},
				},
			},
			"generation": schema.Int64Attribute{
				MarkdownDescription: "Current generation for the data stream.",
				Computed:            true,
			},
			"metadata": schema.StringAttribute{
				MarkdownDescription: "Custom metadata for the stream, copied from the _meta object of the stream's matching index template.",
				Computed:            true,
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "Health status of the data stream.",
				Computed:            true,
			},
			"template": schema.StringAttribute{
				MarkdownDescription: "Name of the index template used to create the data stream's backing indices.",
				Computed:            true,
			},
			"ilm_policy": schema.StringAttribute{
				MarkdownDescription: "Name of the current ILM lifecycle policy in the stream's matching index template.",
				Computed:            true,
			},
			"hidden": schema.BoolAttribute{
				MarkdownDescription: "If `true`, the data stream is hidden.",
				Computed:            true,
			},
			"system": schema.BoolAttribute{
				MarkdownDescription: "If `true`, the data stream is created and managed by an Elastic stack component and cannot be modified through normal user interaction.",
				Computed:            true,
			},
			"replicated": schema.BoolAttribute{
				MarkdownDescription: "If `true`, the data stream is created and managed by cross-cluster replication and the local cluster can not write into this data stream or change its mappings.",
				Computed:            true,
			},
		},
	}
}

// indicesElementType returns the element type for the indices list attribute.
func indicesElementType() attr.Type {
	return GetSchema().Attributes["indices"].GetType().(attr.TypeWithElementType).ElementType()
}
