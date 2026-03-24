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

package kibanaoapi

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// StreamResponse is our own typed response struct for the Streams API.
//
// NOTE: The generated kbapi response types (GetStreamsNameResponse,
// PutStreamsNameResponse) only carry raw Body []byte — there are no typed JSON200
// fields. We therefore unmarshal manually into this struct. If the kbapi
// generator is ever updated to emit typed response bodies for these endpoints,
// this struct can be replaced.
type StreamResponse struct {
	// Stream is the discriminated stream definition.
	Stream StreamDefinition `json:"stream"`
	// Dashboards contains IDs of dashboards linked to this stream.
	Dashboards []string `json:"dashboards,omitempty"`
	// Queries contains ES|QL queries attached to this stream.
	Queries []StreamQuery `json:"queries,omitempty"`
}

// StreamDefinition is the top-level discriminated union for stream definitions.
// The "type" field acts as the discriminator: "wired", "classic", or "query".
type StreamDefinition struct {
	// Type discriminates the stream variant: "wired", "classic", or "query".
	Type string `json:"type"`
	// Name is the stream name (returned on read, not required on write).
	Name string `json:"name,omitempty"`
	// Description is a human-readable description of the stream.
	Description string `json:"description"`

	// Ingest is populated for wired and classic streams.
	Ingest *StreamIngest `json:"ingest,omitempty"`

	// Query is populated for query streams.
	Query *StreamQueryESQLDef `json:"query,omitempty"`
}

// StreamIngest holds ingest settings shared by wired and classic streams.
type StreamIngest struct {
	Processing   StreamProcessing     `json:"processing"`
	Settings     StreamIngestSettings `json:"settings,omitempty"` //nolint:modernize
	Lifecycle    json.RawMessage      `json:"lifecycle,omitempty"`
	FailureStore json.RawMessage      `json:"failure_store,omitempty"`
	// Wired is populated only for wired streams.
	Wired *StreamIngestWired `json:"wired,omitempty"`
	// Classic is populated only for classic streams.
	Classic *StreamIngestClassic `json:"classic,omitempty"`
}

// StreamProcessing holds the processing pipeline steps.
type StreamProcessing struct {
	Steps     json.RawMessage `json:"steps,omitempty"`
	UpdatedAt any             `json:"updated_at,omitempty"`
}

// StreamIngestSettings holds simple index settings.
type StreamIngestSettings struct {
	IndexNumberOfReplicas *StreamIngestSettingValue `json:"index.number_of_replicas,omitempty"`
	IndexNumberOfShards   *StreamIngestSettingValue `json:"index.number_of_shards,omitempty"`
	IndexRefreshInterval  *StreamIngestSettingValue `json:"index.refresh_interval,omitempty"`
}

// StreamIngestSettingValue wraps an index setting value.
type StreamIngestSettingValue struct {
	Value any `json:"value"`
}

// StreamIngestWired holds wired-stream specific ingest config.
type StreamIngestWired struct {
	// Fields is a map of field_name -> field_definition. The field definition
	// itself is a union type (string type name, or a full definition object),
	// so we keep it as raw JSON.
	Fields  json.RawMessage     `json:"fields,omitempty"`
	Routing []StreamRoutingRule `json:"routing,omitempty"`
}

// StreamRoutingRule defines a single routing rule for a wired stream.
type StreamRoutingRule struct {
	Destination string          `json:"destination"`
	Status      *string         `json:"status,omitempty"`
	Where       json.RawMessage `json:"where"`
}

// StreamIngestClassic holds classic-stream specific ingest config.
type StreamIngestClassic struct {
	// FieldOverrides is a map of field_name -> field_override_definition.
	FieldOverrides json.RawMessage `json:"field_overrides,omitempty"`
}

// StreamQueryESQLDef holds the query definition for a query stream.
type StreamQueryESQLDef struct {
	Esql string `json:"esql"`
	View string `json:"view,omitempty"`
}

// StreamQuery holds an ES|QL query attached to a stream.
type StreamQuery struct {
	ID            string          `json:"id"`
	Title         string          `json:"title"`
	Description   string          `json:"description"`
	Esql          StreamQueryEsql `json:"esql"`
	SeverityScore *float32        `json:"severity_score,omitempty"`
	Evidence      *[]string       `json:"evidence,omitempty"`
}

// StreamQueryEsql holds the ES|QL query string.
type StreamQueryEsql struct {
	Query string `json:"query"`
}

// StreamUpsertRequest is the body for PUT /api/streams/{name}.
type StreamUpsertRequest struct {
	Stream     StreamDefinition `json:"stream"`
	Dashboards []string         `json:"dashboards,omitempty"`
	Queries    []StreamQuery    `json:"queries,omitempty"`
}

// GetStream reads a specific stream from the API.
func GetStream(ctx context.Context, client *Client, spaceID string, name string) (*StreamResponse, diag.Diagnostics) {
	resp, err := client.API.GetStreamsNameWithResponse(
		ctx, name, kbapi.GetStreamsNameJSONRequestBody{},
		spaceAwarePathRequestEditor(spaceID),
	)
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		var streamResp StreamResponse
		if jsonErr := json.Unmarshal(resp.Body, &streamResp); jsonErr != nil {
			return nil, diagutil.FrameworkDiagFromError(jsonErr)
		}
		return &streamResp, nil
	case http.StatusNotFound:
		return nil, nil
	default:
		return nil, reportUnknownError(resp.StatusCode(), resp.Body)
	}
}

// UpsertStream creates or updates a stream via PUT /api/streams/{name}.
func UpsertStream(ctx context.Context, client *Client, spaceID string, name string, req StreamUpsertRequest) (*StreamResponse, diag.Diagnostics) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	// The kbapi request body is a discriminated union; UnmarshalJSON/MarshalJSON
	// are now generated, so we can use the typed path directly.
	var kbReq kbapi.PutStreamsNameJSONRequestBody
	if jsonErr := json.Unmarshal(body, &kbReq); jsonErr != nil {
		return nil, diagutil.FrameworkDiagFromError(jsonErr)
	}

	resp, err := client.API.PutStreamsNameWithResponse(
		ctx, name, kbReq,
		spaceAwarePathRequestEditor(spaceID),
	)
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		var streamResp StreamResponse
		if jsonErr := json.Unmarshal(resp.Body, &streamResp); jsonErr != nil {
			return nil, diagutil.FrameworkDiagFromError(jsonErr)
		}
		return &streamResp, nil
	default:
		return nil, reportUnknownError(resp.StatusCode(), resp.Body)
	}
}

// DeleteStream deletes a wired or query stream via DELETE /api/streams/{name}.
// For classic streams this is a no-op (classic streams cannot be deleted via the API).
func DeleteStream(ctx context.Context, client *Client, spaceID string, name string) diag.Diagnostics {
	resp, err := client.API.DeleteStreamsNameWithResponse(
		ctx, name, kbapi.DeleteStreamsNameJSONRequestBody{},
		spaceAwarePathRequestEditor(spaceID),
	)
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK, http.StatusNoContent:
		return nil
	case http.StatusNotFound:
		return nil
	default:
		return reportUnknownError(resp.StatusCode(), resp.Body)
	}
}
