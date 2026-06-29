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

package osquerysavedquery

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/kbschema"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/osquery"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// MinSupportedVersion is the minimum Elastic Stack version supported by Osquery saved query resources and data sources.
var MinSupportedVersion = version.Must(version.NewVersion("8.5.0"))

const (
	attrID            = "id"
	attrSavedObjectID = "saved_object_id"
	attrSavedQueryID  = "saved_query_id"
	attrSpaceID       = "space_id"
	attrQuery         = "query"
	attrDescription   = "description"
	attrPlatform      = "platform"
	attrInterval      = "interval"
	attrVersion       = "version"
	attrSnapshot      = "snapshot"
	attrRemoved       = "removed"
	attrEcsMapping    = "ecs_mapping"
	attrPrebuilt      = "prebuilt"
)

var osqueryPlatformValues = osquery.PlatformValues

func getSchema(_ context.Context) schema.Schema {
	return schema.Schema{
		MarkdownDescription: "Manages a user-defined Osquery saved query in Kibana. Requires Kibana 8.5.0 or later. " +
			"Prebuilt queries shipped with the osquery_manager integration cannot be managed by this resource; " +
			"use the `elasticstack_kibana_osquery_saved_query` data source to read them instead. " +
			"Import of prebuilt queries fails; use the data source for prebuilt queries.",
		Attributes: map[string]schema.Attribute{
			attrID: schema.StringAttribute{
				MarkdownDescription: "Composite identifier in the form `<space_id>/<saved_query_id>`.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			attrSavedObjectID: schema.StringAttribute{
				MarkdownDescription: "Kibana saved object identifier used internally by Kibana's Osquery saved query detail, update, and delete APIs.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			attrSavedQueryID: schema.StringAttribute{
				MarkdownDescription: "Stable user-facing identifier for the saved query. Forces replacement when changed.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			attrSpaceID: kbschema.ResourceSpaceIDAttribute(),
			attrQuery: schema.StringAttribute{
				MarkdownDescription: "Osquery SQL query text.",
				Required:            true,
			},
			attrDescription: schema.StringAttribute{
				MarkdownDescription: "Human-readable description of the saved query.",
				Optional:            true,
			},
			attrPlatform: schema.SetAttribute{
				MarkdownDescription: "Target platforms for the query. Allowed values: `linux`, `darwin`, `windows`.",
				Optional:            true,
				ElementType:         types.StringType,
				Validators: []validator.Set{
					setvalidator.ValueStringsAre(stringvalidator.OneOf(osqueryPlatformValues...)),
				},
			},
			attrInterval: schema.Int64Attribute{
				MarkdownDescription: "Query execution interval in seconds. Required by the Kibana Osquery API on create and update.",
				Required:            true,
			},
			attrVersion: schema.StringAttribute{
				MarkdownDescription: "Saved query version string.",
				Optional:            true,
			},
			attrSnapshot: schema.BoolAttribute{
				MarkdownDescription: "Whether the saved query is a snapshot. Returned by the API and may be set explicitly in configuration. " +
					"When omitted or unknown at plan time, the prior state value is preserved (`UseStateForUnknown`).",
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			attrRemoved: schema.BoolAttribute{
				MarkdownDescription: "Whether the saved query is marked removed. Returned by the API and may be set explicitly in configuration. " +
					"When omitted or unknown at plan time, the prior state value is preserved (`UseStateForUnknown`).",
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			attrEcsMapping: osquery.ECSMappingSchema(),
		},
	}
}
