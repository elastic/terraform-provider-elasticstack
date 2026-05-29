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
	"encoding/json"
	"fmt"
	"time"

	getconnector "github.com/elastic/go-elasticsearch/v8/typedapi/connector/get"
	estypes "github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types/enums/syncstatus"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	esclient "github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	fwtypes "github.com/hashicorp/terraform-plugin-framework/types"
)

// ContentConnectorDataSourceModel is the Terraform state model for the content connector data source.
type ContentConnectorDataSourceModel struct {
	entitycore.ElasticsearchConnectionField
	ID                               fwtypes.String       `tfsdk:"id"`
	ConnectorID                      fwtypes.String       `tfsdk:"connector_id"`
	ServiceType                      fwtypes.String       `tfsdk:"service_type"`
	Name                             fwtypes.String       `tfsdk:"name"`
	Description                      fwtypes.String       `tfsdk:"description"`
	IndexName                        fwtypes.String       `tfsdk:"index_name"`
	IsNative                         fwtypes.Bool         `tfsdk:"is_native"`
	Language                         fwtypes.String       `tfsdk:"language"`
	APIKeyID                         fwtypes.String       `tfsdk:"api_key_id"`
	APIKeySecretID                   fwtypes.String       `tfsdk:"api_key_secret_id"`
	Pipeline                         fwtypes.Object       `tfsdk:"pipeline"`
	Scheduling                       fwtypes.Object       `tfsdk:"scheduling"`
	Features                         fwtypes.Object       `tfsdk:"features"`
	Status                           fwtypes.String       `tfsdk:"status"`
	LastSeen                         fwtypes.String       `tfsdk:"last_seen"`
	LastSynced                       fwtypes.String       `tfsdk:"last_synced"`
	LastSyncStatus                   fwtypes.String       `tfsdk:"last_sync_status"`
	LastIndexedDocumentCount         fwtypes.Int64        `tfsdk:"last_indexed_document_count"`
	LastDeletedDocumentCount         fwtypes.Int64        `tfsdk:"last_deleted_document_count"`
	LastSyncScheduledAt              fwtypes.String       `tfsdk:"last_sync_scheduled_at"`
	LastSyncError                    fwtypes.String       `tfsdk:"last_sync_error"`
	LastAccessControlSyncStatus      fwtypes.String       `tfsdk:"last_access_control_sync_status"`
	LastAccessControlSyncError       fwtypes.String       `tfsdk:"last_access_control_sync_error"`
	LastAccessControlSyncScheduledAt fwtypes.String       `tfsdk:"last_access_control_sync_scheduled_at"`
	LastIncrementalSyncScheduledAt   fwtypes.String       `tfsdk:"last_incremental_sync_scheduled_at"`
	Error                            fwtypes.String       `tfsdk:"error"`
	Filtering                        jsontypes.Normalized `tfsdk:"filtering"`
	CustomScheduling                 jsontypes.Normalized `tfsdk:"custom_scheduling"`
	Configuration                    jsontypes.Normalized `tfsdk:"configuration"`
	SyncCursor                       jsontypes.Normalized `tfsdk:"sync_cursor"`
	SyncNow                          fwtypes.Bool         `tfsdk:"sync_now"`
}

var _ entitycore.WithVersionRequirements = ContentConnectorDataSourceModel{}

// GetVersionRequirements satisfies [entitycore.WithVersionRequirements].
func (ContentConnectorDataSourceModel) GetVersionRequirements() ([]entitycore.VersionRequirement, diag.Diagnostics) {
	return []entitycore.VersionRequirement{{
		MinVersion:   *MinSupportedVersion,
		ErrorMessage: "elasticstack_elasticsearch_connector requires Elasticsearch 8.12.0 or later (connector APIs GA).",
	}}, nil
}

func readContentConnectorDataSource(
	ctx context.Context,
	client *clients.ElasticsearchScopedClient,
	model ContentConnectorDataSourceModel,
) (ContentConnectorDataSourceModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	connectorID := model.ConnectorID.ValueString()
	resp, getDiags := esclient.GetConnector(ctx, client, connectorID)
	diags.Append(getDiags...)
	if diags.HasError() {
		return model, diags
	}

	if resp == nil {
		diags.AddError(
			"Connector not found",
			fmt.Sprintf("Connector %q was not found.", connectorID),
		)
		return model, diags
	}

	populateContentConnectorDataSourceFromAPI(ctx, client, connectorID, resp, &model, &diags)
	return model, diags
}

func populateContentConnectorDataSourceFromAPI(
	ctx context.Context,
	client *clients.ElasticsearchScopedClient,
	connectorID string,
	resp *getconnector.Response,
	model *ContentConnectorDataSourceModel,
	diags *diag.Diagnostics,
) {
	if resp.ServiceType != nil {
		model.ServiceType = fwtypes.StringValue(*resp.ServiceType)
	} else {
		model.ServiceType = fwtypes.StringNull()
	}
	if resp.Name != nil {
		model.Name = fwtypes.StringValue(*resp.Name)
	} else {
		model.Name = fwtypes.StringNull()
	}
	if resp.Description != nil {
		model.Description = fwtypes.StringValue(*resp.Description)
	} else {
		model.Description = fwtypes.StringNull()
	}
	if resp.IndexName != nil {
		model.IndexName = fwtypes.StringValue(*resp.IndexName)
	} else {
		model.IndexName = fwtypes.StringNull()
	}
	model.IsNative = fwtypes.BoolValue(resp.IsNative)
	if resp.Language != nil {
		model.Language = fwtypes.StringValue(*resp.Language)
	} else {
		model.Language = fwtypes.StringNull()
	}
	if resp.ApiKeyId != nil {
		model.APIKeyID = fwtypes.StringValue(*resp.ApiKeyId)
	} else {
		model.APIKeyID = fwtypes.StringNull()
	}
	if resp.ApiKeySecretId != nil {
		model.APIKeySecretID = fwtypes.StringValue(*resp.ApiKeySecretId)
	} else {
		model.APIKeySecretID = fwtypes.StringNull()
	}

	model.Pipeline = populatePipelineFromAPI(ctx, resp.Pipeline, diags)
	model.Scheduling = populateSchedulingFromAPI(ctx, resp.Scheduling, diags)
	model.Features = populateFeaturesFromAPI(ctx, resp.Features, diags)

	model.Status = fwtypes.StringValue(resp.Status.String())
	model.LastSeen = connectorDateTimeToString(resp.LastSeen)
	model.LastSynced = connectorDateTimeToString(resp.LastSynced)
	model.LastSyncStatus = connectorSyncStatusToString(resp.LastSyncStatus)
	model.LastIndexedDocumentCount = connectorInt64PtrToValue(resp.LastIndexedDocumentCount)
	model.LastDeletedDocumentCount = connectorInt64PtrToValue(resp.LastDeletedDocumentCount)
	model.LastSyncScheduledAt = connectorDateTimeToString(resp.LastSyncScheduledAt)
	if resp.LastSyncError != nil {
		model.LastSyncError = fwtypes.StringValue(*resp.LastSyncError)
	} else {
		model.LastSyncError = fwtypes.StringNull()
	}
	model.LastAccessControlSyncStatus = connectorSyncStatusToString(resp.LastAccessControlSyncStatus)
	if resp.LastAccessControlSyncError != nil {
		model.LastAccessControlSyncError = fwtypes.StringValue(*resp.LastAccessControlSyncError)
	} else {
		model.LastAccessControlSyncError = fwtypes.StringNull()
	}
	model.LastAccessControlSyncScheduledAt = connectorDateTimeToString(resp.LastAccessControlSyncScheduledAt)
	model.LastIncrementalSyncScheduledAt = connectorDateTimeToString(resp.LastIncrementalSyncScheduledAt)
	if resp.Error != nil {
		model.Error = fwtypes.StringValue(*resp.Error)
	} else {
		model.Error = fwtypes.StringNull()
	}

	model.Filtering = marshalConnectorJSONField("filtering", resp.Filtering, diags)
	model.CustomScheduling = marshalConnectorJSONField("custom_scheduling", resp.CustomScheduling, diags)
	model.Configuration = marshalConnectorJSONField("configuration", resp.Configuration, diags)
	model.SyncCursor = marshalConnectorRawJSONField("sync_cursor", resp.SyncCursor, diags)
	model.SyncNow = fwtypes.BoolValue(resp.SyncNow)

	id, idDiags := client.ID(ctx, connectorID)
	diags.Append(idDiags...)
	if diags.HasError() {
		return
	}

	model.ID = fwtypes.StringValue(id.String())
	model.ConnectorID = fwtypes.StringValue(connectorID)
}

func connectorSyncStatusToString(status *syncstatus.SyncStatus) fwtypes.String {
	if status == nil {
		return fwtypes.StringNull()
	}
	return fwtypes.StringValue(status.Name)
}

func connectorInt64PtrToValue(v *int64) fwtypes.Int64 {
	if v == nil {
		return fwtypes.Int64Null()
	}
	return fwtypes.Int64Value(*v)
}

func connectorDateTimeToString(dt estypes.DateTime) fwtypes.String {
	if dt == nil {
		return fwtypes.StringNull()
	}
	ms, ok := connectorDateTimeToMillis(dt)
	if !ok || ms == 0 {
		return fwtypes.StringNull()
	}
	return typeutils.TimeToStringValue(time.UnixMilli(ms).UTC())
}

func connectorDateTimeToMillis(v any) (int64, bool) {
	switch x := v.(type) {
	case float64:
		return int64(x), true
	case int64:
		return x, true
	case int:
		return int64(x), true
	case uint64:
		return int64(x), true
	case json.Number:
		i, err := x.Int64()
		if err == nil {
			return i, true
		}
		f, err := x.Float64()
		if err != nil {
			return 0, false
		}
		return int64(f), true
	case string:
		if x == "" {
			return 0, false
		}
		t, err := time.Parse(time.RFC3339, x)
		if err != nil {
			return 0, false
		}
		return t.UnixMilli(), true
	case estypes.DateTime:
		return connectorDateTimeToMillis(any(x))
	default:
		return 0, false
	}
}

func marshalConnectorJSONField(attr string, value any, diags *diag.Diagnostics) jsontypes.Normalized {
	if value == nil {
		return jsontypes.NewNormalizedNull()
	}
	bytes, err := json.Marshal(value)
	if err != nil {
		diags.AddError(
			"Failed to encode connector field",
			fmt.Sprintf("%s: %s", attr, err.Error()),
		)
		return jsontypes.NewNormalizedNull()
	}
	if len(bytes) == 0 {
		return jsontypes.NewNormalizedNull()
	}
	return jsontypes.NewNormalizedValue(string(bytes))
}

func marshalConnectorRawJSONField(attr string, raw json.RawMessage, diags *diag.Diagnostics) jsontypes.Normalized {
	if len(raw) == 0 {
		return jsontypes.NewNormalizedNull()
	}
	if !json.Valid(raw) {
		diags.AddError(
			"Failed to encode connector field",
			fmt.Sprintf("%s: invalid JSON from API", attr),
		)
		return jsontypes.NewNormalizedNull()
	}
	return jsontypes.NewNormalizedValue(string(raw))
}
