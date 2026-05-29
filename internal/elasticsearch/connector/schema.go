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

package connector

import (
	"context"

	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// MinSupportedVersion is the minimum Elasticsearch version supported by this resource.
var MinSupportedVersion = version.Must(version.NewVersion("8.12.0"))

// schemaFactory returns the schema for the content connector resource. The
// elasticsearch_connection block is injected automatically by the envelope.
func schemaFactory(_ context.Context) schema.Schema {
	return schema.Schema{
		MarkdownDescription: contentConnectorResourceMarkdownDescription,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Composite identifier in the form `<cluster_uuid>/<connector_id>`.",
				Computed:            true,
			},
			"connector_id": schema.StringAttribute{
				MarkdownDescription: "Unique connector identifier. When omitted, Elasticsearch auto-generates an ID on create.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"service_type": schema.StringAttribute{
				MarkdownDescription: "Connector service type (for example `postgresql`, `mysql`, `github`). New service types may be added over time; the provider does not validate against a fixed enum.",
				Required:            true,
			},
			nameAttrName: schema.StringAttribute{
				MarkdownDescription: "Human-readable connector name.",
				Optional:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Connector description.",
				Optional:            true,
			},
			"index_name": schema.StringAttribute{
				MarkdownDescription: "Destination Elasticsearch index name. When omitted, Elasticsearch may assign a default on create.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"is_native": schema.BoolAttribute{
				MarkdownDescription: "Whether this is an Elastic-managed connector (`true`) or self-managed (`false`). Defaults to `false` on the Elasticsearch side when omitted.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"language": schema.StringAttribute{
				MarkdownDescription: "Analyzer language for the connector index.",
				Optional:            true,
			},
			"api_key_id": schema.StringAttribute{
				MarkdownDescription: "ID of the API key used by the connector service for authorization.",
				Optional:            true,
			},
			"api_key_secret_id": schema.StringAttribute{
				MarkdownDescription: "ID of the connector secret holding the API key (Elastic-managed connectors only).",
				Optional:            true,
			},
			"pipeline":             pipelineSingleNestedAttribute(),
			"scheduling":           schedulingSingleNestedAttribute(),
			"features":             featuresSingleNestedAttribute(),
			"configuration_values": configurationValuesMapNestedAttribute(),
		},
	}
}

func pipelineSingleNestedAttribute() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		MarkdownDescription: "Ingest pipeline settings applied to synced documents. Changes trigger `PUT /_connector/{id}/_pipeline`.",
		Optional:            true,
		Computed:            true,
		PlanModifiers: []planmodifier.Object{
			objectplanmodifier.UseStateForUnknown(),
		},
		Attributes: map[string]schema.Attribute{
			nameAttrName: schema.StringAttribute{
				MarkdownDescription: "Ingest pipeline name.",
				Required:            true,
			},
			"extract_binary_content": schema.BoolAttribute{
				MarkdownDescription: "Whether to extract binary content during ingestion.",
				Required:            true,
			},
			"reduce_whitespace": schema.BoolAttribute{
				MarkdownDescription: "Whether to reduce whitespace in extracted text.",
				Required:            true,
			},
			"run_ml_inference": schema.BoolAttribute{
				MarkdownDescription: "Whether to run ML inference during ingestion.",
				Required:            true,
			},
		},
	}
}

func schedulingSingleNestedAttribute() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		MarkdownDescription: "Sync scheduling for full, incremental, and access-control jobs. Changes trigger `PUT /_connector/{id}/_scheduling`.",
		Optional:            true,
		Computed:            true,
		PlanModifiers: []planmodifier.Object{
			objectplanmodifier.UseStateForUnknown(),
		},
		Attributes: map[string]schema.Attribute{
			"full":           scheduleEntrySingleNestedAttribute("full"),
			"incremental":    scheduleEntrySingleNestedAttribute("incremental"),
			"access_control": scheduleEntrySingleNestedAttribute("access_control"),
		},
	}
}

func scheduleEntrySingleNestedAttribute(jobKind string) schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		MarkdownDescription: "Schedule for the `" + jobKind + "` sync job type.",
		Optional:            true,
		Attributes: map[string]schema.Attribute{
			enabledAttrName: schema.BoolAttribute{
				MarkdownDescription: "Whether this scheduled job type is enabled.",
				Required:            true,
			},
			intervalAttrName: schema.StringAttribute{
				MarkdownDescription: "Cron expression accepted by the Elasticsearch scheduler.",
				Required:            true,
			},
		},
	}
}

func featuresSingleNestedAttribute() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		MarkdownDescription: "Connector feature flags. Changes trigger `PUT /_connector/{id}/_features`.",
		Optional:            true,
		Computed:            true,
		PlanModifiers: []planmodifier.Object{
			objectplanmodifier.UseStateForUnknown(),
		},
		Attributes: map[string]schema.Attribute{
			"document_level_security":   featureFlagSingleNestedAttribute("document_level_security"),
			"incremental_sync":          featureFlagSingleNestedAttribute("incremental_sync"),
			"native_connector_api_keys": featureFlagSingleNestedAttribute("native_connector_api_keys"),
			"sync_rules":                syncRulesSingleNestedAttribute(),
		},
	}
}

func featureFlagSingleNestedAttribute(featureName string) schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		MarkdownDescription: "Feature flag for `" + featureName + "`.",
		Optional:            true,
		Attributes: map[string]schema.Attribute{
			enabledAttrName: schema.BoolAttribute{
				MarkdownDescription: "Whether the feature is enabled.",
				Required:            true,
			},
		},
	}
}

func syncRulesSingleNestedAttribute() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		MarkdownDescription: "Sync rules feature flags.",
		Optional:            true,
		Attributes: map[string]schema.Attribute{
			"basic":    featureFlagSingleNestedAttribute("basic"),
			"advanced": featureFlagSingleNestedAttribute("advanced"),
		},
	}
}

func configurationValuesMapNestedAttribute() schema.MapNestedAttribute {
	return schema.MapNestedAttribute{
		MarkdownDescription: "User-supplied connector configuration values keyed by field name. " +
			"Each element must set exactly one of `string`, `number`, `bool`, `json`, or `secret_value`. " +
			"Removing a key stops managing it but does not unset the value server-side.",
		Optional: true,
		NestedObject: schema.NestedAttributeObject{
			Validators: []validator.Object{
				configurationValueBranchValidator{},
			},
			Attributes: map[string]schema.Attribute{
				stringBranchAttrName: schema.StringAttribute{
					MarkdownDescription: "String configuration value.",
					Optional:            true,
				},
				numberBranchAttrName: schema.NumberAttribute{
					MarkdownDescription: "Numeric configuration value (integer or float).",
					Optional:            true,
				},
				boolBranchAttrName: schema.BoolAttribute{
					MarkdownDescription: "Boolean configuration value.",
					Optional:            true,
				},
				jsonBranchAttrName: schema.StringAttribute{
					// jsontypes.NormalizedType enforces syntactic JSON validity at decode time (REQ-008).
					MarkdownDescription: "JSON-encoded object or array configuration value.",
					Optional:            true,
					CustomType:          jsontypes.NormalizedType{},
				},
				secretValueBranchAttrName: schema.StringAttribute{
					MarkdownDescription: "Write-only secret configuration value. Drift is detected via private-state hashing (see resource documentation).",
					Optional:            true,
					WriteOnly:           true,
					Sensitive:           true,
				},
			},
		},
	}
}
