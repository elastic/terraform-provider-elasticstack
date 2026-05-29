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

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

// dataSourceSchemaFactory returns the schema for the content connector data source.
// The elasticsearch_connection block is injected automatically by the envelope.
func dataSourceSchemaFactory(_ context.Context) dschema.Schema {
	return dschema.Schema{
		MarkdownDescription: contentConnectorDataSourceMarkdownDescription,
		Attributes: map[string]dschema.Attribute{
			"id": dschema.StringAttribute{
				MarkdownDescription: "Composite identifier in the form `<cluster_uuid>/<connector_id>`.",
				Computed:            true,
			},
			"connector_id": dschema.StringAttribute{
				MarkdownDescription: "Unique connector identifier to look up.",
				Required:            true,
			},
			"service_type": dschema.StringAttribute{
				MarkdownDescription: "Connector service type (for example `postgresql`, `mysql`, `github`).",
				Computed:            true,
			},
			nameAttrName: dschema.StringAttribute{
				MarkdownDescription: "Human-readable connector name.",
				Computed:            true,
			},
			"description": dschema.StringAttribute{
				MarkdownDescription: "Connector description.",
				Computed:            true,
			},
			"index_name": dschema.StringAttribute{
				MarkdownDescription: "Destination Elasticsearch index name.",
				Computed:            true,
			},
			"is_native": dschema.BoolAttribute{
				MarkdownDescription: "Whether this is an Elastic-managed connector (`true`) or self-managed (`false`).",
				Computed:            true,
			},
			"language": dschema.StringAttribute{
				MarkdownDescription: "Analyzer language for the connector index.",
				Computed:            true,
			},
			"api_key_id": dschema.StringAttribute{
				MarkdownDescription: "ID of the API key used by the connector service for authorization.",
				Computed:            true,
			},
			"api_key_secret_id": dschema.StringAttribute{
				MarkdownDescription: "ID of the connector secret holding the API key (Elastic-managed connectors only).",
				Computed:            true,
			},
			"pipeline":         dataSourcePipelineSingleNestedAttribute(),
			"scheduling":       dataSourceSchedulingSingleNestedAttribute(),
			"features":         dataSourceFeaturesSingleNestedAttribute(),
			"status":           dataSourceStatusAttribute(),
			"last_seen":        dataSourceLastSeenAttribute(),
			"last_synced":      dataSourceLastSyncedAttribute(),
			"last_sync_status": dataSourceLastSyncStatusAttribute(),
			"last_indexed_document_count": dschema.Int64Attribute{
				MarkdownDescription: "Number of documents indexed during the last sync job.",
				Computed:            true,
			},
			"last_deleted_document_count": dschema.Int64Attribute{
				MarkdownDescription: "Number of documents deleted during the last sync job.",
				Computed:            true,
			},
			"last_sync_scheduled_at": dataSourceLastSyncScheduledAtAttribute(),
			"last_sync_error": dschema.StringAttribute{
				MarkdownDescription: "Error message from the last sync job, if any.",
				Computed:            true,
			},
			"last_access_control_sync_status": dataSourceLastAccessControlSyncStatusAttribute(),
			"last_access_control_sync_error": dschema.StringAttribute{
				MarkdownDescription: "Error message from the last access-control sync job, if any.",
				Computed:            true,
			},
			"last_access_control_sync_scheduled_at": dataSourceLastAccessControlSyncScheduledAtAttribute(),
			"last_incremental_sync_scheduled_at":    dataSourceLastIncrementalSyncScheduledAtAttribute(),
			"error": dschema.StringAttribute{
				MarkdownDescription: "Connector-level error message, if any.",
				Computed:            true,
			},
			"filtering": dschema.StringAttribute{
				MarkdownDescription: "Connector filtering rules. JSON-encoded array; use `jsondecode()` to inspect.",
				Computed:            true,
				CustomType:          jsontypes.NormalizedType{},
			},
			"custom_scheduling": dschema.StringAttribute{
				MarkdownDescription: "Custom per-job-type scheduling overrides. JSON-encoded object; use `jsondecode()` to inspect.",
				Computed:            true,
				CustomType:          jsontypes.NormalizedType{},
			},
			"configuration": dschema.StringAttribute{
				MarkdownDescription: "Full registered configuration schema document from the connector service. " +
					"JSON-encoded; use `jsondecode()` to inspect.",
				Computed:   true,
				CustomType: jsontypes.NormalizedType{},
			},
			"sync_cursor": dschema.StringAttribute{
				MarkdownDescription: "Opaque connector sync cursor state. JSON-encoded; use `jsondecode()` to inspect.",
				Computed:            true,
				CustomType:          jsontypes.NormalizedType{},
			},
			"sync_now": dschema.BoolAttribute{
				MarkdownDescription: "Whether a sync job is queued to run immediately.",
				Computed:            true,
			},
		},
	}
}

func dataSourcePipelineSingleNestedAttribute() dschema.SingleNestedAttribute {
	return dschema.SingleNestedAttribute{
		MarkdownDescription: "Ingest pipeline settings applied to synced documents.",
		Computed:            true,
		Attributes: map[string]dschema.Attribute{
			nameAttrName: dschema.StringAttribute{
				MarkdownDescription: "Ingest pipeline name.",
				Computed:            true,
			},
			extractBinaryContentAttrName: dschema.BoolAttribute{
				MarkdownDescription: "Whether to extract binary content during ingestion.",
				Computed:            true,
			},
			reduceWhitespaceAttrName: dschema.BoolAttribute{
				MarkdownDescription: "Whether to reduce whitespace in extracted text.",
				Computed:            true,
			},
			runMlInferenceAttrName: dschema.BoolAttribute{
				MarkdownDescription: "Whether to run ML inference during ingestion.",
				Computed:            true,
			},
		},
	}
}

func dataSourceSchedulingSingleNestedAttribute() dschema.SingleNestedAttribute {
	return dschema.SingleNestedAttribute{
		MarkdownDescription: "Sync scheduling for full, incremental, and access-control jobs.",
		Computed:            true,
		Attributes: map[string]dschema.Attribute{
			fullScheduleAttrName:          dataSourceScheduleEntrySingleNestedAttribute(fullScheduleAttrName),
			incrementalScheduleAttrName:   dataSourceScheduleEntrySingleNestedAttribute(incrementalScheduleAttrName),
			accessControlScheduleAttrName: dataSourceScheduleEntrySingleNestedAttribute(accessControlScheduleAttrName),
		},
	}
}

func dataSourceScheduleEntrySingleNestedAttribute(jobKind string) dschema.SingleNestedAttribute {
	return dschema.SingleNestedAttribute{
		MarkdownDescription: "Schedule for the `" + jobKind + "` sync job type.",
		Computed:            true,
		Attributes: map[string]dschema.Attribute{
			enabledAttrName: dschema.BoolAttribute{
				MarkdownDescription: "Whether this scheduled job type is enabled.",
				Computed:            true,
			},
			intervalAttrName: dschema.StringAttribute{
				MarkdownDescription: "Cron expression accepted by the Elasticsearch scheduler.",
				Computed:            true,
			},
		},
	}
}

func dataSourceFeaturesSingleNestedAttribute() dschema.SingleNestedAttribute {
	return dschema.SingleNestedAttribute{
		MarkdownDescription: "Connector feature flags.",
		Computed:            true,
		Attributes: map[string]dschema.Attribute{
			documentLevelSecurityAttrName:  dataSourceFeatureFlagSingleNestedAttribute(documentLevelSecurityAttrName),
			incrementalSyncAttrName:        dataSourceFeatureFlagSingleNestedAttribute(incrementalSyncAttrName),
			nativeConnectorAPIKeysAttrName: dataSourceFeatureFlagSingleNestedAttribute(nativeConnectorAPIKeysAttrName),
			syncRulesAttrName:              dataSourceSyncRulesSingleNestedAttribute(),
		},
	}
}

func dataSourceFeatureFlagSingleNestedAttribute(featureName string) dschema.SingleNestedAttribute {
	return dschema.SingleNestedAttribute{
		MarkdownDescription: "Feature flag for `" + featureName + "`.",
		Computed:            true,
		Attributes: map[string]dschema.Attribute{
			enabledAttrName: dschema.BoolAttribute{
				MarkdownDescription: "Whether the feature is enabled.",
				Computed:            true,
			},
		},
	}
}

func dataSourceSyncRulesSingleNestedAttribute() dschema.SingleNestedAttribute {
	return dschema.SingleNestedAttribute{
		MarkdownDescription: "Sync rules feature flags.",
		Computed:            true,
		Attributes: map[string]dschema.Attribute{
			basicSyncRulesAttrName:    dataSourceFeatureFlagSingleNestedAttribute(basicSyncRulesAttrName),
			advancedSyncRulesAttrName: dataSourceFeatureFlagSingleNestedAttribute(advancedSyncRulesAttrName),
		},
	}
}

func dataSourceStatusAttribute() dschema.StringAttribute {
	return dschema.StringAttribute{
		MarkdownDescription: "Connector lifecycle status (for example `created`, `connected`, `error`).",
		Computed:            true,
	}
}

func dataSourceLastSeenAttribute() dschema.StringAttribute {
	return dschema.StringAttribute{
		MarkdownDescription: "ISO 8601 timestamp when the connector service last reported in.",
		Computed:            true,
	}
}

func dataSourceLastSyncedAttribute() dschema.StringAttribute {
	return dschema.StringAttribute{
		MarkdownDescription: "ISO 8601 timestamp of the last completed sync.",
		Computed:            true,
	}
}

func dataSourceLastSyncStatusAttribute() dschema.StringAttribute {
	return dschema.StringAttribute{
		MarkdownDescription: "Status of the last sync job.",
		Computed:            true,
	}
}

func dataSourceLastSyncScheduledAtAttribute() dschema.StringAttribute {
	return dschema.StringAttribute{
		MarkdownDescription: "ISO 8601 timestamp when the last sync job was scheduled.",
		Computed:            true,
	}
}

func dataSourceLastAccessControlSyncStatusAttribute() dschema.StringAttribute {
	return dschema.StringAttribute{
		MarkdownDescription: "Status of the last access-control sync job.",
		Computed:            true,
	}
}

func dataSourceLastAccessControlSyncScheduledAtAttribute() dschema.StringAttribute {
	return dschema.StringAttribute{
		MarkdownDescription: "ISO 8601 timestamp when the last access-control sync was scheduled.",
		Computed:            true,
	}
}

func dataSourceLastIncrementalSyncScheduledAtAttribute() dschema.StringAttribute {
	return dschema.StringAttribute{
		MarkdownDescription: "ISO 8601 timestamp when the last incremental sync was scheduled.",
		Computed:            true,
	}
}
