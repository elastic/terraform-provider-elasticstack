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

package streams

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (r *Resource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = getSchema()
}

func getSchema() schema.Schema {
	return schema.Schema{
		MarkdownDescription: "Manages Kibana [Streams](https://www.elastic.co/docs/api/doc/kibana/group/endpoint-streams). " +
			"Streams is an experimental feature for managing data ingestion in Kibana. " +
			"Requires Elastic Stack 9.4.0 or higher (the stream type discriminator field used by this resource was introduced in 9.4.0). " +
			"This functionality is in technical preview and may be changed or removed in a future release.\n\n" +
			"Three stream types are supported:\n" +
			"- **Wired streams** (`wired_config`): fully managed data streams with typed field mappings and routing rules.\n" +
			"- **Classic streams** (`classic_config`): adopt existing Elasticsearch data streams — they cannot be created or deleted via this resource, only imported and updated.\n" +
			"- **Query streams** (`query_config`): virtual streams defined by an ES|QL query.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Generated composite identifier for the stream (`space_id/name`).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"space_id": schema.StringAttribute{
				MarkdownDescription: "An identifier for the space. If not provided, the default space is used.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("default"),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the stream. Stream names follow Elasticsearch data stream naming conventions (e.g. `logs.nginx`).",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "A human-readable description of the stream.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
			},
			"wired_config": schema.SingleNestedAttribute{
				MarkdownDescription: "Configuration for a wired stream. Wired streams are fully managed data streams with explicit field mappings and routing rules. " +
					"Mutually exclusive with `classic_config` and `query_config`.",
				Optional:   true,
				Attributes: getWiredConfigSchema(),
				Validators: []validator.Object{
					objectvalidator.ConflictsWith(
						path.MatchRoot("classic_config"),
						path.MatchRoot("query_config"),
					),
				},
			},
			"classic_config": schema.SingleNestedAttribute{
				MarkdownDescription: "Configuration for a classic stream. Classic streams adopt pre-existing Elasticsearch data streams. " +
					"They cannot be created or deleted via this resource — use `terraform import` to manage them. " +
					"Mutually exclusive with `wired_config` and `query_config`.",
				Optional:   true,
				Attributes: getClassicConfigSchema(),
				Validators: []validator.Object{
					objectvalidator.ConflictsWith(
						path.MatchRoot("wired_config"),
						path.MatchRoot("query_config"),
					),
				},
			},
			"query_config": schema.SingleNestedAttribute{
				MarkdownDescription: "Configuration for a query stream. Query streams are virtual streams defined by an ES|QL query. " +
					"Mutually exclusive with `wired_config` and `classic_config`.",
				Optional:   true,
				Attributes: getQueryConfigSchema(),
				Validators: []validator.Object{
					objectvalidator.ConflictsWith(
						path.MatchRoot("wired_config"),
						path.MatchRoot("classic_config"),
					),
				},
			},
			"dashboards": schema.ListAttribute{
				MarkdownDescription: "List of dashboard IDs to link to this stream.",
				ElementType:         types.StringType,
				Optional:            true,
			},
			"queries": schema.ListNestedAttribute{
				MarkdownDescription: "ES|QL queries attached to this stream.",
				Optional:            true,
				NestedObject:        getStreamQuerySchema(),
			},
		},
	}
}

// getWiredConfigSchema returns the schema for wired stream configuration.
func getWiredConfigSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"processing_steps": schema.ListAttribute{
			MarkdownDescription: "Processing pipeline steps in streamlang format. Each element is a JSON-encoded " +
				"step object (e.g. `jsonencode({ action = \"grok\", from = \"message\", patterns = [\"...\"] })`). " +
				"Steps are applied in order during ingest. Storing each step as a separate list element gives " +
				"granular per-step diffs in Terraform plans. Conditions and nested steps are supported by " +
				"embedding the full streamlang object as JSON.",
			Optional:    true,
			ElementType: jsontypes.NormalizedType{},
		},
		"fields_json": schema.StringAttribute{
			MarkdownDescription: "Field type mappings as a JSON object. Maps field names to their type definitions " +
				"(e.g. `{\"host.name\": {\"type\": \"keyword\"}}`). Wired streams enforce these mappings across routed data.",
			CustomType: jsontypes.NormalizedType{},
			Optional:   true,
		},
		"routing_json": schema.StringAttribute{
			MarkdownDescription: "Routing rules as a JSON array. Each rule defines a destination child stream and a " +
				"filter condition (`where`) that determines which documents are routed there. " +
				"Example: `[{\"destination\": \"logs.nginx.errors\", \"where\": {\"field\": \"http.response.status_code\", \"gte\": 400}}]`.",
			CustomType: jsontypes.NormalizedType{},
			Optional:   true,
		},
		"lifecycle_json": schema.StringAttribute{
			MarkdownDescription: "Lifecycle configuration as a JSON object. Supports DSL (`{\"dsl\": {\"data_retention\": \"30d\"}}`), " +
				"ILM (`{\"ilm\": {\"policy\": \"my-policy\"}}`), or inherited lifecycle (`{\"inherit\": {}}`). " +
				"When not set, the previous state value is preserved on update; on first create defaults to `{\"inherit\":{}}` " +
				"and the server value is stored in state.",
			CustomType: jsontypes.NormalizedType{},
			Optional:   true,
			Computed:   true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"failure_store_json": schema.StringAttribute{
			MarkdownDescription: "Failure store configuration as a JSON object. Controls where failed ingest documents are stored. " +
				"Supports `{\"inherit\": {}}`, `{\"disabled\": {}}`, or a lifecycle-enabled configuration. " +
				"When not set, defaults to `{\"disabled\":{}}` and the server value is stored in state.",
			CustomType: jsontypes.NormalizedType{},
			Optional:   true,
			Computed:   true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"index_number_of_shards": schema.Int64Attribute{
			MarkdownDescription: "Number of primary shards for the underlying index.",
			Optional:            true,
		},
		"index_number_of_replicas": schema.Int64Attribute{
			MarkdownDescription: "Number of replica shards for the underlying index.",
			Optional:            true,
		},
		"index_refresh_interval": schema.StringAttribute{
			MarkdownDescription: "How often to refresh the index (e.g. `1s`, `5s`, `-1` to disable). " +
				"Accepts a duration string or `-1`.",
			Optional: true,
		},
	}
}

// getClassicConfigSchema returns the schema for classic stream configuration.
func getClassicConfigSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"processing_steps": schema.ListAttribute{
			MarkdownDescription: "Processing pipeline steps in streamlang format. Each element is a JSON-encoded " +
				"step object. Steps are applied in order during ingest.",
			Optional:    true,
			ElementType: jsontypes.NormalizedType{},
		},
		"field_overrides_json": schema.StringAttribute{
			MarkdownDescription: "Field override definitions as a JSON object. Maps field names to override configurations " +
				"for classic stream field handling.",
			CustomType: jsontypes.NormalizedType{},
			Optional:   true,
		},
		"lifecycle_json": schema.StringAttribute{
			MarkdownDescription: "Lifecycle configuration as a JSON object. Supports DSL, ILM, or inherited lifecycle. " +
				"When not set, the previous state value is preserved on update.",
			CustomType: jsontypes.NormalizedType{},
			Optional:   true,
			Computed:   true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"failure_store_json": schema.StringAttribute{
			MarkdownDescription: "Failure store configuration as a JSON object. " +
				"When not set, defaults to `{\"disabled\":{}}` and the server value is stored in state.",
			CustomType: jsontypes.NormalizedType{},
			Optional:   true,
			Computed:   true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"index_number_of_shards": schema.Int64Attribute{
			MarkdownDescription: "Number of primary shards for the underlying index.",
			Optional:            true,
		},
		"index_number_of_replicas": schema.Int64Attribute{
			MarkdownDescription: "Number of replica shards for the underlying index.",
			Optional:            true,
		},
		"index_refresh_interval": schema.StringAttribute{
			MarkdownDescription: "How often to refresh the index (e.g. `1s`, `5s`, `-1` to disable).",
			Optional:            true,
		},
	}
}

// getQueryConfigSchema returns the schema for query stream configuration.
func getQueryConfigSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"esql": schema.StringAttribute{
			MarkdownDescription: "The ES|QL query that defines this virtual stream (e.g. `FROM logs* | WHERE host.name == \"web-01\"`).",
			Required:            true,
		},
		"view": schema.StringAttribute{
			MarkdownDescription: "Optional view name for the query stream.",
			Optional:            true,
			Computed:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
	}
}

// getStreamQuerySchema returns the schema for ES|QL queries attached to a stream.
func getStreamQuerySchema() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "A unique identifier for the query.",
				Required:            true,
			},
			"title": schema.StringAttribute{
				MarkdownDescription: "A human-readable title for the query.",
				Required:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "A human-readable description for the query.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
			},
			"esql": schema.StringAttribute{
				MarkdownDescription: "The ES|QL query string.",
				Required:            true,
			},
			"severity_score": schema.Float64Attribute{
				MarkdownDescription: "Optional severity score for the query (0–100).",
				Optional:            true,
			},
			"evidence": schema.ListAttribute{
				MarkdownDescription: "Optional list of evidence field names for the query.",
				ElementType:         types.StringType,
				Optional:            true,
			},
		},
	}
}
