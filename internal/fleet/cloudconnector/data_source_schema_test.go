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

package cloudconnector

import (
	"context"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var _ entitycore.WithVersionRequirements = (*cloudConnectorsDataSourceModel)(nil)

func TestNewDataSource_implementsDataSource(t *testing.T) {
	t.Parallel()

	ds := NewDataSource()
	require.NotNil(t, ds)
	require.Implements(t, (*datasource.DataSource)(nil), ds)
	require.Implements(t, (*datasource.DataSourceWithConfigure)(nil), ds)
}

func TestNewDataSource_metadata(t *testing.T) {
	t.Parallel()

	ds := NewDataSource()

	var resp datasource.MetadataResponse
	ds.Metadata(context.Background(), datasource.MetadataRequest{
		ProviderTypeName: "elasticstack",
	}, &resp)

	require.Equal(t, "elasticstack_fleet_cloud_connectors", resp.TypeName)
}

func TestGetDataSourceSchema_noDiagnostics(t *testing.T) {
	t.Parallel()

	ds := NewDataSource()

	var resp datasource.SchemaResponse
	ds.Schema(context.Background(), datasource.SchemaRequest{}, &resp)

	require.False(t, resp.Diagnostics.HasError(), "schema must not produce diagnostics: %v", resp.Diagnostics)
	require.Empty(t, resp.Diagnostics.Warnings(), "schema must not produce warnings")
}

func TestGetDataSourceSchema_attributes(t *testing.T) {
	t.Parallel()

	s := getDataSourceSchema(context.Background())

	wantOptional := map[string]bool{
		attrSpaceID: true,
		attrKuery:   true,
		attrPage:    true,
		attrPerPage: true,
	}
	wantComputed := []string{
		attrID,
		attrSpaceID,
		attrKuery,
		attrPage,
		attrPerPage,
		attrCloudConnectors,
	}

	for name := range wantOptional {
		attr, ok := s.Attributes[name]
		require.True(t, ok, "expected attribute %q", name)
		switch a := attr.(type) {
		case dsschema.StringAttribute:
			assert.True(t, a.Optional, "%q should be optional", name)
		case dsschema.Int64Attribute:
			assert.True(t, a.Optional, "%q should be optional", name)
		default:
			t.Fatalf("unexpected attribute type for %q: %T", name, attr)
		}
	}

	for _, name := range wantComputed {
		attr, ok := s.Attributes[name]
		require.True(t, ok, "expected attribute %q", name)
		switch a := attr.(type) {
		case dsschema.StringAttribute:
			if name == attrSpaceID || name == attrKuery {
				continue
			}
			assert.True(t, a.Computed, "%q should be computed", name)
		case dsschema.Int64Attribute:
			if name == attrPage || name == attrPerPage {
				continue
			}
			assert.True(t, a.Computed, "%q should be computed", name)
		case dsschema.ListNestedAttribute:
			assert.True(t, a.Computed, "%q should be computed", name)
		default:
			t.Fatalf("unexpected attribute type for %q: %T", name, attr)
		}
	}

	connectorsAttr, ok := s.Attributes[attrCloudConnectors].(dsschema.ListNestedAttribute)
	require.True(t, ok)

	nestedWant := []string{
		attrID,
		attrCloudConnectorID,
		attrSpaceID,
		attrName,
		attrCloudProvider,
		attrAccountType,
		attrNamespace,
		attrPackagePolicyCount,
		attrVerificationStatus,
		attrVerificationStartedAt,
		attrVerificationFailedAt,
		attrCreatedAt,
		attrUpdatedAt,
	}
	for _, name := range nestedWant {
		nestedAttr, ok := connectorsAttr.NestedObject.Attributes[name]
		require.True(t, ok, "expected nested attribute %q", name)
		switch a := nestedAttr.(type) {
		case dsschema.StringAttribute:
			assert.True(t, a.Computed, "nested %q should be computed", name)
		case dsschema.Int64Attribute:
			assert.True(t, a.Computed, "nested %q should be computed", name)
		default:
			t.Fatalf("unexpected nested attribute type for %q: %T", name, nestedAttr)
		}
	}

	for _, excluded := range []string{attrAWSBlock, attrAzureBlock, attrVarsMap} {
		_, ok := connectorsAttr.NestedObject.Attributes[excluded]
		assert.False(t, ok, "nested attribute %q must not be exposed", excluded)
	}
}

func TestCloudConnectorsDataSourceModel_GetVersionRequirements(t *testing.T) {
	t.Parallel()

	var m cloudConnectorsDataSourceModel
	reqs, diags := m.GetVersionRequirements()
	require.False(t, diags.HasError())
	require.Len(t, reqs, 1)
	require.Equal(t, *cloudConnectorMinVersion, reqs[0].MinVersion)
	require.NotEmpty(t, reqs[0].ErrorMessage)
}
