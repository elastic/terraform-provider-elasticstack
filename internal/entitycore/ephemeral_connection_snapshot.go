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

package entitycore

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

const ephemeralConnectionKey = "entitycore.ephemeral.connection"

var ephemeralConnectionNullMarker = []byte("null")

// ephemeralESConnectionSnapshot stores elasticsearch_connection values as plain
// Go types so they round-trip through JSON private state without loss.
type ephemeralESConnectionSnapshot struct {
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
	CAFingerprint          string            `json:"ca_fingerprint,omitempty"`
	CertFile               string            `json:"cert_file,omitempty"`
	CertData               string            `json:"cert_data,omitempty"`
	KeyFile                string            `json:"key_file,omitempty"`
	KeyData                string            `json:"key_data,omitempty"`
}

type ephemeralKibanaConnectionSnapshot struct {
	Username    string   `json:"username,omitempty"`
	Password    string   `json:"password,omitempty"`
	APIKey      string   `json:"api_key,omitempty"`
	BearerToken string   `json:"bearer_token,omitempty"`
	Endpoints   []string `json:"endpoints,omitempty"`
	CACerts     []string `json:"ca_certs,omitempty"`
	Insecure    *bool    `json:"insecure,omitempty"`
}

func encodeEphemeralConnection[Snap any](
	ctx context.Context,
	connection types.List,
	snapshotFrom func(context.Context, types.List) (*Snap, diag.Diagnostics),
	errLabel string,
) ([]byte, diag.Diagnostics) {
	var diags diag.Diagnostics
	if !typeutils.IsKnown(connection) {
		return ephemeralConnectionNullMarker, diags
	}

	snapshot, snapshotDiags := snapshotFrom(ctx, connection)
	diags.Append(snapshotDiags...)
	if diags.HasError() || snapshot == nil {
		return ephemeralConnectionNullMarker, diags
	}

	data, err := json.Marshal(snapshot)
	if err != nil {
		diags.AddError("Failed to marshal "+errLabel+" for Close", err.Error())
		return nil, diags
	}
	return data, diags
}

func decodeEphemeralConnection[Snap any](
	ctx context.Context,
	data []byte,
	nullList func() types.List,
	errLabel string,
	listFrom func(context.Context, *Snap) (types.List, diag.Diagnostics),
) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	if len(data) == 0 || string(data) == "null" {
		return nullList(), diags
	}

	var snapshot Snap
	if err := json.Unmarshal(data, &snapshot); err != nil {
		diags.AddError("Failed to parse "+errLabel+" from ephemeral private data", err.Error())
		return nullList(), diags
	}

	return listFrom(ctx, &snapshot)
}

func encodeElasticsearchConnection(ctx context.Context, connection types.List) ([]byte, diag.Diagnostics) {
	return encodeEphemeralConnection(ctx, connection, esConnectionSnapshotFromList, blockElasticsearchConnection)
}

func decodeElasticsearchConnection(ctx context.Context, data []byte) (types.List, diag.Diagnostics) {
	return decodeEphemeralConnection(ctx, data, providerschema.ElasticsearchConnectionNullList, blockElasticsearchConnection, esConnectionListFromSnapshot)
}

func encodeKibanaConnection(ctx context.Context, connection types.List) ([]byte, diag.Diagnostics) {
	return encodeEphemeralConnection(ctx, connection, kibanaConnectionSnapshotFromList, blockKibanaConnection)
}

func decodeKibanaConnection(ctx context.Context, data []byte) (types.List, diag.Diagnostics) {
	return decodeEphemeralConnection(ctx, data, providerschema.KibanaConnectionNullList, blockKibanaConnection, kibanaConnectionListFromSnapshot)
}

func esConnectionSnapshotFromList(ctx context.Context, connection types.List) (*ephemeralESConnectionSnapshot, diag.Diagnostics) {
	var diags diag.Diagnostics

	var connections []clientconfig.ElasticsearchConnection
	diags.Append(connection.ElementsAs(ctx, &connections, false)...)
	if diags.HasError() || len(connections) == 0 {
		return nil, diags
	}

	return snapshotFromElasticsearchConnection(ctx, connections[0])
}

func snapshotFromElasticsearchConnection(ctx context.Context, conn clientconfig.ElasticsearchConnection) (*ephemeralESConnectionSnapshot, diag.Diagnostics) {
	var diags diag.Diagnostics
	snapshot := &ephemeralESConnectionSnapshot{
		Username:               knownStringValue(conn.Username),
		Password:               knownStringValue(conn.Password),
		APIKey:                 knownStringValue(conn.APIKey),
		BearerToken:            knownStringValue(conn.BearerToken),
		ESClientAuthentication: knownStringValue(conn.ESClientAuthentication),
		CAFile:                 knownStringValue(conn.CAFile),
		CAData:                 knownStringValue(conn.CAData),
		CAFingerprint:          knownStringValue(conn.CAFingerprint),
		CertFile:               knownStringValue(conn.CertFile),
		CertData:               knownStringValue(conn.CertData),
		KeyFile:                knownStringValue(conn.KeyFile),
		KeyData:                knownStringValue(conn.KeyData),
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

func esConnectionListFromSnapshot(ctx context.Context, snapshot *ephemeralESConnectionSnapshot) (types.List, diag.Diagnostics) {
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

func elasticsearchConnectionFromSnapshot(snapshot *ephemeralESConnectionSnapshot) (clientconfig.ElasticsearchConnection, diag.Diagnostics) {
	var diags diag.Diagnostics

	conn := clientconfig.ElasticsearchConnection{
		Username:               typeutils.NonEmptyStringishValue(snapshot.Username),
		Password:               typeutils.NonEmptyStringishValue(snapshot.Password),
		APIKey:                 typeutils.NonEmptyStringishValue(snapshot.APIKey),
		BearerToken:            typeutils.NonEmptyStringishValue(snapshot.BearerToken),
		ESClientAuthentication: typeutils.NonEmptyStringishValue(snapshot.ESClientAuthentication),
		Insecure:               types.BoolPointerValue(snapshot.Insecure),
		CAFile:                 typeutils.NonEmptyStringishValue(snapshot.CAFile),
		CAData:                 typeutils.NonEmptyStringishValue(snapshot.CAData),
		CAFingerprint:          typeutils.NonEmptyStringishValue(snapshot.CAFingerprint),
		CertFile:               typeutils.NonEmptyStringishValue(snapshot.CertFile),
		CertData:               typeutils.NonEmptyStringishValue(snapshot.CertData),
		KeyFile:                typeutils.NonEmptyStringishValue(snapshot.KeyFile),
		KeyData:                typeutils.NonEmptyStringishValue(snapshot.KeyData),
	}

	if len(snapshot.Endpoints) > 0 {
		conn.Endpoints = typeutils.StringsToListMust(snapshot.Endpoints)
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

func kibanaConnectionSnapshotFromList(ctx context.Context, connection types.List) (*ephemeralKibanaConnectionSnapshot, diag.Diagnostics) {
	var diags diag.Diagnostics

	var connections []clientconfig.KibanaConnection
	diags.Append(connection.ElementsAs(ctx, &connections, false)...)
	if diags.HasError() || len(connections) == 0 {
		return nil, diags
	}

	return snapshotFromKibanaConnection(ctx, connections[0])
}

func snapshotFromKibanaConnection(ctx context.Context, conn clientconfig.KibanaConnection) (*ephemeralKibanaConnectionSnapshot, diag.Diagnostics) {
	var diags diag.Diagnostics
	snapshot := &ephemeralKibanaConnectionSnapshot{
		Username:    knownStringValue(conn.Username),
		Password:    knownStringValue(conn.Password),
		APIKey:      knownStringValue(conn.APIKey),
		BearerToken: knownStringValue(conn.BearerToken),
	}

	if typeutils.IsKnown(conn.Endpoints) {
		diags.Append(conn.Endpoints.ElementsAs(ctx, &snapshot.Endpoints, false)...)
	}
	if typeutils.IsKnown(conn.CACerts) {
		diags.Append(conn.CACerts.ElementsAs(ctx, &snapshot.CACerts, false)...)
	}
	if typeutils.IsKnown(conn.Insecure) {
		insecure := conn.Insecure.ValueBool()
		snapshot.Insecure = &insecure
	}

	return snapshot, diags
}

func kibanaConnectionListFromSnapshot(ctx context.Context, snapshot *ephemeralKibanaConnectionSnapshot) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	if snapshot == nil {
		return providerschema.KibanaConnectionNullList(), diags
	}

	conn := kibanaConnectionFromSnapshot(snapshot)
	connection, listDiags := types.ListValueFrom(ctx, providerschema.KibanaConnectionObjectType(), []clientconfig.KibanaConnection{conn})
	diags.Append(listDiags...)
	if diags.HasError() {
		return providerschema.KibanaConnectionNullList(), diags
	}

	return connection, diags
}

func kibanaConnectionFromSnapshot(snapshot *ephemeralKibanaConnectionSnapshot) clientconfig.KibanaConnection {
	conn := clientconfig.KibanaConnection{
		Username:    typeutils.NonEmptyStringishValue(snapshot.Username),
		Password:    typeutils.NonEmptyStringishValue(snapshot.Password),
		APIKey:      typeutils.NonEmptyStringishValue(snapshot.APIKey),
		BearerToken: typeutils.NonEmptyStringishValue(snapshot.BearerToken),
		Insecure:    types.BoolPointerValue(snapshot.Insecure),
	}

	if len(snapshot.Endpoints) > 0 {
		conn.Endpoints = typeutils.StringsToListMust(snapshot.Endpoints)
	} else {
		conn.Endpoints = types.ListNull(types.StringType)
	}

	if len(snapshot.CACerts) > 0 {
		conn.CACerts = typeutils.StringsToListMust(snapshot.CACerts)
	} else {
		conn.CACerts = types.ListNull(types.StringType)
	}

	return conn
}

func knownStringValue(value types.String) string {
	if !typeutils.IsKnown(value) {
		return ""
	}
	return value.ValueString()
}
