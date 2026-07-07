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

package data_source

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	getconnector "github.com/elastic/go-elasticsearch/v9/typedapi/connector/get"
	estypes "github.com/elastic/go-elasticsearch/v9/typedapi/types"
	"github.com/elastic/go-elasticsearch/v9/typedapi/types/enums/syncstatus"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	esclient "github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/connector"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	fwtypes "github.com/hashicorp/terraform-plugin-framework/types"
)

// ContentConnectorDataSourceModel is the Terraform state model for the content connector data source.
type ContentConnectorDataSourceModel struct {
	entitycore.ElasticsearchConnectionField
	connector.CoreConnectorFields
	connector.VersionGate
	ID                               fwtypes.String       `tfsdk:"id"`
	ConnectorID                      fwtypes.String       `tfsdk:"connector_id"`
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
	model.CoreConnectorFields = connector.PopulateCoreConnectorFieldsFromAPI(ctx, resp, diags)

	model.Status = fwtypes.StringValue(resp.Status.String())
	model.LastSeen = connectorDateTimeToString(resp.LastSeen)
	model.LastSynced = connectorDateTimeToString(resp.LastSynced)
	model.LastSyncStatus = connectorSyncStatusToString(resp.LastSyncStatus)
	model.LastIndexedDocumentCount = fwtypes.Int64PointerValue(resp.LastIndexedDocumentCount)
	model.LastDeletedDocumentCount = fwtypes.Int64PointerValue(resp.LastDeletedDocumentCount)
	model.LastSyncScheduledAt = connectorDateTimeToString(resp.LastSyncScheduledAt)
	model.LastSyncError = typeutils.StringishPointerValue(resp.LastSyncError)
	model.LastAccessControlSyncStatus = connectorSyncStatusToString(resp.LastAccessControlSyncStatus)
	model.LastAccessControlSyncError = typeutils.StringishPointerValue(resp.LastAccessControlSyncError)
	model.LastAccessControlSyncScheduledAt = connectorDateTimeToString(resp.LastAccessControlSyncScheduledAt)
	model.LastIncrementalSyncScheduledAt = connectorDateTimeToString(resp.LastIncrementalSyncScheduledAt)
	model.Error = typeutils.StringishPointerValue(resp.Error)

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

func connectorDateTimeToString(dt estypes.DateTime) fwtypes.String {
	return typeutils.ElasticDateTimeToStringValue(dt)
}

func isNilValue(v any) bool {
	if v == nil {
		return true
	}
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Map, reflect.Slice, reflect.Pointer, reflect.Interface:
		return rv.IsNil()
	}
	return false
}

func marshalConnectorJSONField(attr string, value any, diags *diag.Diagnostics) jsontypes.Normalized {
	if isNilValue(value) {
		return jsontypes.NewNormalizedNull()
	}
	encoded, err := json.Marshal(value)
	if err != nil {
		diags.AddError(
			"Failed to encode connector field",
			fmt.Sprintf("%s: %s", attr, err.Error()),
		)
		return jsontypes.NewNormalizedNull()
	}
	if string(encoded) == connector.JSONNullLiteral {
		return jsontypes.NewNormalizedNull()
	}
	return jsontypes.NewNormalizedValue(string(encoded))
}

func marshalConnectorRawJSONField(attr string, raw json.RawMessage, diags *diag.Diagnostics) jsontypes.Normalized {
	if len(raw) == 0 || bytes.Equal(bytes.TrimSpace(raw), []byte(connector.JSONNullLiteral)) {
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
