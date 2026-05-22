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

package ephemeral

import (
	"context"
	"encoding/json"

	clientconfig "github.com/elastic/terraform-provider-elasticstack/internal/clients/config"
	providerschema "github.com/elastic/terraform-provider-elasticstack/internal/schema"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ephemeralConnectionSnapshot stores elasticsearch_connection values as plain
// Go types so they round-trip through JSON private state without loss.
type ephemeralConnectionSnapshot struct {
	Username               string            `json:"username,omitempty"`
	Password               string            `json:"password,omitempty"`
	APIKey                 string            `json:"api_key,omitempty"`
	BearerToken            string            `json:"bearer_token,omitempty"`
	ESClientAuthentication string            `json:"es_client_authentication,omitempty"`
	Endpoints              []string          `json:"endpoints,omitempty"`
	Headers                map[string]string `json:"headers,omitempty"`
	Insecure               *bool             `json:"insecure,omitempty"`
	CAFile                 string            `json:"ca_file,omitempty"`
	CAData                 string            `json:"ca_data,omitempty"`
	CertFile               string            `json:"cert_file,omitempty"`
	CertData               string            `json:"cert_data,omitempty"`
	KeyFile                string            `json:"key_file,omitempty"`
	KeyData                string            `json:"key_data,omitempty"`
}

func encodeElasticsearchConnection(ctx context.Context, connection types.List) (string, diag.Diagnostics) {
	var diags diag.Diagnostics
	if !typeutils.IsKnown(connection) {
		return "", diags
	}

	snapshot, snapshotDiags := connectionSnapshotFromList(ctx, connection)
	diags.Append(snapshotDiags...)
	if diags.HasError() || snapshot == nil {
		return "", diags
	}

	bytes, err := json.Marshal(snapshot)
	if err != nil {
		diags.AddError("Failed to marshal elasticsearch_connection for Close", err.Error())
		return "", diags
	}

	return string(bytes), diags
}

func decodeElasticsearchConnection(ctx context.Context, connectionJSON string) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	if connectionJSON == "" {
		return providerschema.ElasticsearchConnectionNullList(), diags
	}

	var snapshot ephemeralConnectionSnapshot
	if err := json.Unmarshal([]byte(connectionJSON), &snapshot); err != nil {
		diags.AddError("Failed to parse elasticsearch_connection from ephemeral private data", err.Error())
		return providerschema.ElasticsearchConnectionNullList(), diags
	}

	return connectionListFromSnapshot(ctx, &snapshot)
}

func connectionSnapshotFromList(ctx context.Context, connection types.List) (*ephemeralConnectionSnapshot, diag.Diagnostics) {
	var diags diag.Diagnostics

	var connections []clientconfig.ElasticsearchConnection
	diags.Append(connection.ElementsAs(ctx, &connections, false)...)
	if diags.HasError() || len(connections) == 0 {
		return nil, diags
	}

	return snapshotFromElasticsearchConnection(ctx, connections[0])
}

func snapshotFromElasticsearchConnection(ctx context.Context, conn clientconfig.ElasticsearchConnection) (*ephemeralConnectionSnapshot, diag.Diagnostics) {
	var diags diag.Diagnostics
	snapshot := &ephemeralConnectionSnapshot{
		Username:               knownString(conn.Username),
		Password:               knownString(conn.Password),
		APIKey:                 knownString(conn.APIKey),
		BearerToken:            knownString(conn.BearerToken),
		ESClientAuthentication: knownString(conn.ESClientAuthentication),
		CAFile:                 knownString(conn.CAFile),
		CAData:                 knownString(conn.CAData),
		CertFile:               knownString(conn.CertFile),
		CertData:               knownString(conn.CertData),
		KeyFile:                knownString(conn.KeyFile),
		KeyData:                knownString(conn.KeyData),
	}

	if typeutils.IsKnown(conn.Endpoints) {
		diags.Append(conn.Endpoints.ElementsAs(ctx, &snapshot.Endpoints, false)...)
	}
	if typeutils.IsKnown(conn.Headers) {
		diags.Append(conn.Headers.ElementsAs(ctx, &snapshot.Headers, false)...)
	}
	if typeutils.IsKnown(conn.Insecure) {
		insecure := conn.Insecure.ValueBool()
		snapshot.Insecure = &insecure
	}

	return snapshot, diags
}

func knownString(value types.String) string {
	if !typeutils.IsKnown(value) {
		return ""
	}
	return value.ValueString()
}

func connectionListFromSnapshot(ctx context.Context, snapshot *ephemeralConnectionSnapshot) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	if snapshot == nil {
		return providerschema.ElasticsearchConnectionNullList(), diags
	}

	conn, connDiags := elasticsearchConnectionFromSnapshot(snapshot)
	diags.Append(connDiags...)
	if diags.HasError() {
		return providerschema.ElasticsearchConnectionNullList(), diags
	}

	connection, listDiags := types.ListValueFrom(ctx, providerschema.ElasticsearchConnectionObjectType(), []clientconfig.ElasticsearchConnection{conn})
	diags.Append(listDiags...)
	if diags.HasError() {
		return providerschema.ElasticsearchConnectionNullList(), diags
	}

	return connection, diags
}

func elasticsearchConnectionFromSnapshot(snapshot *ephemeralConnectionSnapshot) (clientconfig.ElasticsearchConnection, diag.Diagnostics) {
	var diags diag.Diagnostics

	conn := clientconfig.ElasticsearchConnection{
		Username:               optionalStringValue(snapshot.Username),
		Password:               optionalStringValue(snapshot.Password),
		APIKey:                 optionalStringValue(snapshot.APIKey),
		BearerToken:            optionalStringValue(snapshot.BearerToken),
		ESClientAuthentication: optionalStringValue(snapshot.ESClientAuthentication),
		Insecure:               optionalBoolPointerValue(snapshot.Insecure),
		CAFile:                 optionalStringValue(snapshot.CAFile),
		CAData:                 optionalStringValue(snapshot.CAData),
		CertFile:               optionalStringValue(snapshot.CertFile),
		CertData:               optionalStringValue(snapshot.CertData),
		KeyFile:                optionalStringValue(snapshot.KeyFile),
		KeyData:                optionalStringValue(snapshot.KeyData),
	}

	if len(snapshot.Endpoints) > 0 {
		endpointValues := make([]attr.Value, 0, len(snapshot.Endpoints))
		for _, endpoint := range snapshot.Endpoints {
			endpointValues = append(endpointValues, types.StringValue(endpoint))
		}
		endpoints, endpointsDiags := types.ListValue(types.StringType, endpointValues)
		diags.Append(endpointsDiags...)
		conn.Endpoints = endpoints
	} else {
		conn.Endpoints = types.ListNull(types.StringType)
	}

	if len(snapshot.Headers) > 0 {
		headerValues := make(map[string]attr.Value, len(snapshot.Headers))
		for key, value := range snapshot.Headers {
			headerValues[key] = types.StringValue(value)
		}
		headers, headersDiags := types.MapValue(types.StringType, headerValues)
		diags.Append(headersDiags...)
		conn.Headers = headers
	} else {
		conn.Headers = types.MapNull(types.StringType)
	}

	return conn, diags
}

func optionalStringValue(value string) types.String {
	if value == "" {
		return types.StringNull()
	}
	return types.StringValue(value)
}

func optionalBoolPointerValue(value *bool) types.Bool {
	if value == nil {
		return types.BoolNull()
	}
	return types.BoolValue(*value)
}
