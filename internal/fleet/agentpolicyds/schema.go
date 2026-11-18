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

package agentpolicyds

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/kbschema"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func getDataSourceSchema(_ context.Context) dsschema.Schema {
	return dsschema.Schema{
		Description: "Returns a Fleet agent policy looked up by `policy_id` or `name`. See https://www.elastic.co/guide/en/fleet/current/agent-policy.html",
		Attributes: map[string]dsschema.Attribute{
			"id": dsschema.StringAttribute{
				Description: "The ID of this agent policy.",
				Computed:    true,
			},
			"policy_id": dsschema.StringAttribute{
				Description: "The identifier of the agent policy. Set exactly one of `policy_id` or `name`; the other is populated from the API response.",
				Optional:    true,
				Computed:    true,
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(path.MatchRelative().AtParent().AtName("name")),
				},
			},
			"name": dsschema.StringAttribute{
				Description: "The name of the agent policy. Set exactly one of `policy_id` or `name`; the other is populated from the API response.",
				Optional:    true,
				Computed:    true,
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(path.MatchRelative().AtParent().AtName("policy_id")),
				},
			},
			"space_id": kbschema.DataSourceSpaceIDAttribute(),
			"namespace": dsschema.StringAttribute{
				Description: "The namespace of the agent policy.",
				Computed:    true,
			},
			"description": dsschema.StringAttribute{
				Description: "The description of the agent policy, as returned by the Fleet API.",
				Computed:    true,
			},
			"data_output_id": dsschema.StringAttribute{
				Description: "The identifier for the data output.",
				Computed:    true,
			},
			"monitoring_output_id": dsschema.StringAttribute{
				Description: "The identifier for monitoring output.",
				Computed:    true,
			},
			"fleet_server_host_id": dsschema.StringAttribute{
				Description: "The identifier for the Fleet server host.",
				Computed:    true,
			},
			"download_source_id": dsschema.StringAttribute{
				Description: "The identifier for the Elastic Agent binary download server.",
				Computed:    true,
			},
			"monitor_logs": dsschema.BoolAttribute{
				Description: "Whether collection of agent logs is enabled.",
				Computed:    true,
			},
			"monitor_metrics": dsschema.BoolAttribute{
				Description: "Whether collection of agent metrics is enabled.",
				Computed:    true,
			},
			"is_protected": dsschema.BoolAttribute{
				Description: "Whether tamper protection is enabled for this agent policy (Fleet API `is_protected`).",
				Computed:    true,
			},
			"host_name_format": dsschema.StringAttribute{
				Description: "Hostname format reported for agents in this policy. `hostname` is the short hostname; `fqdn` is the fully qualified domain name.",
				Computed:    true,
			},
			"supports_agentless": dsschema.BoolAttribute{
				Description: "Whether the agent policy supports agentless data collection.",
				Computed:    true,
			},
			"inactivity_timeout": dsschema.StringAttribute{
				Description: "Time an agent can be inactive before it is marked inactive. A duration string such as `30s`, `2m`, or `1h`.",
				Computed:    true,
				CustomType:  customtypes.DurationType{},
			},
			"unenrollment_timeout": dsschema.StringAttribute{
				Description: "Time an agent can be inactive before it is automatically unenrolled. A duration string such as `30s`, `2m`, or `1h`.",
				Computed:    true,
				CustomType:  customtypes.DurationType{},
			},
			"global_data_tags": dsschema.MapNestedAttribute{
				Description: "User-defined tags. Keyed by tag name; each entry has exactly one of `string_value` or `number_value`.",
				Computed:    true,
				NestedObject: dsschema.NestedAttributeObject{
					Attributes: map[string]dsschema.Attribute{
						"string_value": dsschema.StringAttribute{
							Description: "String value for the tag.",
							Computed:    true,
						},
						"number_value": dsschema.Float32Attribute{
							Description: "Number value for the tag.",
							Computed:    true,
						},
					},
				},
			},
			"space_ids": dsschema.SetAttribute{
				Description: "The Kibana space IDs that this agent policy is available in.",
				ElementType: types.StringType,
				Computed:    true,
			},
			"required_versions": dsschema.MapAttribute{
				Description: "Map of agent versions to target percentages for automatic upgrade. The key is the target version and the value is the percentage of agents to upgrade to that version.",
				ElementType: types.Int32Type,
				Computed:    true,
			},
			"advanced_settings":           advancedSettingsAttribute(),
			"advanced_monitoring_options": advancedMonitoringOptionsAttribute(),
		},
	}
}

func advancedSettingsAttribute() dsschema.SingleNestedAttribute {
	return dsschema.SingleNestedAttribute{
		Description: "Advanced agent settings for logging, resource limits, and downloads.",
		Computed:    true,
		Attributes: map[string]dsschema.Attribute{
			"logging_level": dsschema.StringAttribute{
				Description: "Logging level for the agent.",
				Computed:    true,
			},
			"logging_to_files": dsschema.BoolAttribute{
				Description: "Whether logging to files is enabled.",
				Computed:    true,
			},
			"logging_files_interval": dsschema.StringAttribute{
				Description: "Interval for log file rotation.",
				Computed:    true,
				CustomType:  customtypes.DurationType{},
			},
			"logging_files_keepfiles": dsschema.Int32Attribute{
				Description: "Number of rotated log files to keep.",
				Computed:    true,
			},
			"logging_files_rotateeverybytes": dsschema.Int64Attribute{
				Description: "Rotate log files when they reach this size in bytes.",
				Computed:    true,
			},
			"logging_metrics_period": dsschema.StringAttribute{
				Description: "Period for logging agent metrics.",
				Computed:    true,
				CustomType:  customtypes.DurationType{},
			},
			"go_max_procs": dsschema.Int32Attribute{
				Description: "Maximum number of CPUs that the agent can use (GOMAXPROCS).",
				Computed:    true,
			},
			"download_timeout": dsschema.StringAttribute{
				Description: "Timeout for downloading agent updates.",
				Computed:    true,
				CustomType:  customtypes.DurationType{},
			},
			"download_target_directory": dsschema.StringAttribute{
				Description: "Target directory for downloading agent updates.",
				Computed:    true,
			},
			"monitoring_runtime_experimental": dsschema.StringAttribute{
				Description: "Experimental runtime monitoring mode.",
				Computed:    true,
			},
			"features_disable_policy_change_acks_enabled": dsschema.BoolAttribute{
				Description: "Whether policy change acknowledgements from agents are disabled.",
				Computed:    true,
			},
			"features_include_tags_in_events_enabled": dsschema.BoolAttribute{
				Description: "Whether the agent's tags are included in events it sends.",
				Computed:    true,
			},
			"internal": dsschema.StringAttribute{
				Description: "Advanced internal agent settings as free-form YAML.",
				Computed:    true,
				CustomType:  customtypes.NormalizedYamlType{},
			},
		},
	}
}

func advancedMonitoringOptionsAttribute() dsschema.SingleNestedAttribute {
	return dsschema.SingleNestedAttribute{
		Description: "Advanced monitoring options for the agent policy, including the HTTP monitoring endpoint and diagnostic settings.",
		Computed:    true,
		Attributes: map[string]dsschema.Attribute{
			"http_monitoring_endpoint": dsschema.SingleNestedAttribute{
				Description: "HTTP monitoring endpoint configuration for agent health checks and liveness probes.",
				Computed:    true,
				Attributes: map[string]dsschema.Attribute{
					"enabled": dsschema.BoolAttribute{
						Description: "Whether the HTTP monitoring endpoint is enabled.",
						Computed:    true,
					},
					"host": dsschema.StringAttribute{
						Description: "Host for the HTTP monitoring endpoint.",
						Computed:    true,
					},
					"port": dsschema.Int32Attribute{
						Description: "Port for the HTTP monitoring endpoint.",
						Computed:    true,
					},
					"buffer_enabled": dsschema.BoolAttribute{
						Description: "Whether the monitoring buffer for the HTTP endpoint is enabled.",
						Computed:    true,
					},
					"pprof_enabled": dsschema.BoolAttribute{
						Description: "Whether /debug/pprof profiling endpoints are enabled.",
						Computed:    true,
					},
				},
			},
			"diagnostics": dsschema.SingleNestedAttribute{
				Description: "Diagnostic settings for rate limiting and file upload behavior.",
				Computed:    true,
				Attributes: map[string]dsschema.Attribute{
					"rate_limits": dsschema.SingleNestedAttribute{
						Description: "Rate limiting configuration for diagnostics requests from Fleet.",
						Computed:    true,
						Attributes: map[string]dsschema.Attribute{
							"interval": dsschema.StringAttribute{
								Description: "Rate limiting interval for diagnostics requests.",
								Computed:    true,
								CustomType:  customtypes.DurationType{},
							},
							"burst": dsschema.Int32Attribute{
								Description: "Rate limiting burst count for diagnostics requests.",
								Computed:    true,
							},
						},
					},
					"file_uploader": dsschema.SingleNestedAttribute{
						Description: "Diagnostic file upload retry configuration.",
						Computed:    true,
						Attributes: map[string]dsschema.Attribute{
							"init_duration": dsschema.StringAttribute{
								Description: "Initial duration before the first retry attempt.",
								Computed:    true,
								CustomType:  customtypes.DurationType{},
							},
							"backoff_duration": dsschema.StringAttribute{
								Description: "Maximum backoff duration between retry attempts.",
								Computed:    true,
								CustomType:  customtypes.DurationType{},
							},
							"max_retries": dsschema.Int32Attribute{
								Description: "Maximum number of retry attempts for file uploads.",
								Computed:    true,
							},
						},
					},
				},
			},
		},
	}
}
