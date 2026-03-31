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

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	kibanaoapi "github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const (
	streamTypeWired   = "wired"
	streamTypeClassic = "classic"
	streamTypeQuery   = "query"
)

// streamModel is the top-level Terraform model for elasticstack_kibana_stream.
type streamModel struct {
	ID            types.String        `tfsdk:"id"`
	SpaceID       types.String        `tfsdk:"space_id"`
	Name          types.String        `tfsdk:"name"`
	Description   types.String        `tfsdk:"description"`
	WiredConfig   *wiredConfigModel   `tfsdk:"wired_config"`
	ClassicConfig *classicConfigModel `tfsdk:"classic_config"`
	QueryConfig   *queryConfigModel   `tfsdk:"query_config"`
	Dashboards    types.List          `tfsdk:"dashboards"`
	Queries       []streamQueryModel  `tfsdk:"queries"`
}

// streamQueryModel is the Terraform model for an attached ES|QL query.
type streamQueryModel struct {
	ID            types.String  `tfsdk:"id"`
	Title         types.String  `tfsdk:"title"`
	Description   types.String  `tfsdk:"description"`
	Esql          types.String  `tfsdk:"esql"`
	SeverityScore types.Float64 `tfsdk:"severity_score"`
	Evidence      types.List    `tfsdk:"evidence"`
}

// streamType returns the stream type discriminator based on which config block is set.
func (m *streamModel) streamType() string {
	switch {
	case m.WiredConfig != nil:
		return streamTypeWired
	case m.ClassicConfig != nil:
		return streamTypeClassic
	case m.QueryConfig != nil:
		return streamTypeQuery
	default:
		return ""
	}
}

// populateFromAPI populates the Terraform model from the API response.
func (m *streamModel) populateFromAPI(ctx context.Context, resp *kibanaoapi.StreamResponse, name, spaceID string) diag.Diagnostics {
	var diags diag.Diagnostics

	resourceID := clients.CompositeID{ClusterID: spaceID, ResourceID: name}
	m.ID = types.StringValue(resourceID.String())
	m.Name = types.StringValue(name)
	m.SpaceID = types.StringValue(spaceID)
	m.Description = types.StringValue(resp.Stream.Description)

	// Map stream type config
	switch resp.Stream.Type {
	case streamTypeWired:
		if m.WiredConfig == nil {
			m.WiredConfig = &wiredConfigModel{}
		}
		diags.Append(m.WiredConfig.populateFromAPI(ctx, resp.Stream.Ingest)...)
		m.ClassicConfig = nil
		m.QueryConfig = nil
	case streamTypeClassic:
		if m.ClassicConfig == nil {
			m.ClassicConfig = &classicConfigModel{}
		}
		m.ClassicConfig.populateFromAPI(ctx, resp.Stream.Ingest)
		m.WiredConfig = nil
		m.QueryConfig = nil
	case streamTypeQuery:
		if m.QueryConfig == nil {
			m.QueryConfig = &queryConfigModel{}
		}
		m.QueryConfig.populateFromAPI(resp.Stream.Query)
		m.WiredConfig = nil
		m.ClassicConfig = nil
	}

	// Map dashboards
	if len(resp.Dashboards) > 0 {
		m.Dashboards = typeutils.SliceToListTypeString(ctx, resp.Dashboards, path.Root("dashboards"), &diags)
	} else {
		m.Dashboards = types.ListNull(types.StringType)
	}

	// Map queries
	if len(resp.Queries) > 0 {
		queries := make([]streamQueryModel, 0, len(resp.Queries))
		for _, q := range resp.Queries {
			qm := streamQueryModel{
				ID:          types.StringValue(q.ID),
				Title:       types.StringValue(q.Title),
				Description: types.StringValue(q.Description),
				Esql:        types.StringValue(q.Esql.Query),
			}
			if q.SeverityScore != nil {
				qm.SeverityScore = types.Float64Value(float64(*q.SeverityScore))
			} else {
				qm.SeverityScore = types.Float64Null()
			}
			if q.Evidence != nil && len(*q.Evidence) > 0 {
				qm.Evidence = typeutils.SliceToListTypeString(ctx, *q.Evidence, path.Root("queries"), &diags)
			} else {
				qm.Evidence = types.ListNull(types.StringType)
			}
			queries = append(queries, qm)
		}
		m.Queries = queries
	} else {
		m.Queries = nil
	}

	return diags
}

// toAPIUpsertRequest converts the Terraform model to an API upsert request.
func (m *streamModel) toAPIUpsertRequest(ctx context.Context, diags *diag.Diagnostics) kibanaoapi.StreamUpsertRequest {
	// Initialise all required array fields as empty slices, not nil.
	// The API rejects requests where these are absent or null.
	req := kibanaoapi.StreamUpsertRequest{
		Dashboards: []string{},
		Rules:      []string{},
		Queries:    []kibanaoapi.StreamQuery{},
	}

	// Build stream definition
	req.Stream = kibanaoapi.StreamDefinition{
		Type:        m.streamType(),
		Description: m.Description.ValueString(),
	}

	switch m.streamType() {
	case streamTypeWired:
		req.Stream.Ingest = m.WiredConfig.toAPIIngest(diags)
	case streamTypeClassic:
		req.Stream.Ingest = m.ClassicConfig.toAPIIngest(diags)
	case streamTypeQuery:
		req.Stream.Query = m.QueryConfig.toAPI(m.Name.ValueString())
	}

	// Map dashboards
	if typeutils.IsKnown(m.Dashboards) {
		if d := typeutils.ListTypeToSliceString(ctx, m.Dashboards, path.Root("dashboards"), diags); d != nil {
			req.Dashboards = d
		}
	}

	// Map queries — req.Queries is pre-initialised to []{}; append if present
	if len(m.Queries) > 0 {
		req.Queries = make([]kibanaoapi.StreamQuery, 0, len(m.Queries))
		for _, qm := range m.Queries {
			q := kibanaoapi.StreamQuery{
				ID:          qm.ID.ValueString(),
				Title:       qm.Title.ValueString(),
				Description: qm.Description.ValueString(),
				Esql:        kibanaoapi.StreamQueryEsql{Query: qm.Esql.ValueString()},
			}
			if typeutils.IsKnown(qm.SeverityScore) {
				score := float32(qm.SeverityScore.ValueFloat64())
				q.SeverityScore = &score
			}
			if typeutils.IsKnown(qm.Evidence) {
				evidence := typeutils.ListTypeToSliceString(ctx, qm.Evidence, path.Root("queries"), diags)
				if evidence != nil {
					q.Evidence = &evidence
				}
			}
			req.Queries = append(req.Queries, q)
		}
	}

	return req
}
