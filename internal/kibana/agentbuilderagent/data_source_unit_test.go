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

package agentbuilderagent

import (
	"context"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/stretchr/testify/require"
)

// Compile-time assertion: agentDataSourceModel implements the optional
// version-requirements interface. This ensures the envelope will invoke
// GetVersionRequirements for the agent data source.
var _ entitycore.KibanaDataSourceWithVersionRequirements = (*agentDataSourceModel)(nil)

// =============================================================================
// Subtask 5.1 — NewDataSource() interface compliance
// =============================================================================

// TestNewDataSource_implementsDataSource verifies at runtime that NewDataSource()
// returns a value that satisfies datasource.DataSource and
// datasource.DataSourceWithConfigure. The compile-time assertions in
// data_source.go also cover this, but a runtime assertion provides a clear test
// failure message when interfaces drift.
func TestNewDataSource_implementsDataSource(t *testing.T) {
	t.Parallel()
	ds := NewDataSource()
	require.NotNil(t, ds)
	require.Implements(t, (*datasource.DataSource)(nil), ds)
	require.Implements(t, (*datasource.DataSourceWithConfigure)(nil), ds)
}

// =============================================================================
// Subtask 5.2 — Metadata returns the expected type name
// =============================================================================

// TestNewDataSource_metadata verifies that the envelope computes the correct
// TypeName for the agent data source: "elasticstack_kibana_agentbuilder_agent".
func TestNewDataSource_metadata(t *testing.T) {
	t.Parallel()
	ds := NewDataSource()

	var resp datasource.MetadataResponse
	ds.Metadata(context.Background(), datasource.MetadataRequest{
		ProviderTypeName: "elasticstack",
	}, &resp)

	require.Equal(t, "elasticstack_kibana_agentbuilder_agent", resp.TypeName)
}

// =============================================================================
// Subtask 5.3 — Schema includes kibana_connection and all agent attributes
// =============================================================================

// TestNewDataSource_schemaAttributes verifies that the envelope-injected schema
// includes:
//   - kibana_connection (in Blocks, injected by the envelope)
//   - all Agent Builder attributes (in Attributes, from getDataSourceSchema)
func TestNewDataSource_schemaAttributes(t *testing.T) {
	t.Parallel()
	ds := NewDataSource()

	var resp datasource.SchemaResponse
	ds.Schema(context.Background(), datasource.SchemaRequest{}, &resp)

	require.False(t, resp.Diagnostics.HasError(), "Schema must not produce diagnostics: %v", resp.Diagnostics)

	// kibana_connection is injected by the envelope into Blocks.
	require.Contains(t, resp.Schema.Blocks, "kibana_connection",
		"schema Blocks must contain kibana_connection (injected by envelope)")

	// Agent Builder attributes come from getDataSourceSchema.
	wantAttrs := []string{
		"id",
		"agent_id",
		"space_id",
		"include_dependencies",
		"name",
		"description",
		"avatar_color",
		"avatar_symbol",
		"labels",
		"tools",
	}
	for _, attr := range wantAttrs {
		require.Contains(t, resp.Schema.Attributes, attr,
			"schema Attributes must contain %q", attr)
	}
}

// =============================================================================
// Additional: GetVersionRequirements returns a non-empty requirement
// =============================================================================

// TestAgentDataSourceModel_GetVersionRequirements confirms that
// agentDataSourceModel.GetVersionRequirements returns exactly one requirement
// with MinVersion equal to minKibanaAgentBuilderAPIVersion and a non-empty
// ErrorMessage.
func TestAgentDataSourceModel_GetVersionRequirements(t *testing.T) {
	t.Parallel()
	var m agentDataSourceModel
	reqs, diags := m.GetVersionRequirements()
	require.False(t, diags.HasError(), "GetVersionRequirements must not return error diagnostics")
	require.Len(t, reqs, 1, "GetVersionRequirements must return exactly one requirement")
	require.Equal(t, *minKibanaAgentBuilderAPIVersion, reqs[0].MinVersion,
		"MinVersion must equal minKibanaAgentBuilderAPIVersion")
	require.NotEmpty(t, reqs[0].ErrorMessage,
		"ErrorMessage must not be empty")
}
