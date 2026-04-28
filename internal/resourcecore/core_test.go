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

package resourcecore

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/stretchr/testify/require"
)

const testProviderTypeName = "elasticstack"

func TestCore_Metadata_typeNamesPerComponent(t *testing.T) {
	cases := []struct {
		name         string
		component    Component
		resourceName string
		want         string
	}{
		{
			name:         "elasticsearch",
			component:    ComponentElasticsearch,
			resourceName: "ml_job_state",
			want:         "elasticstack_elasticsearch_ml_job_state",
		},
		{
			name:         "kibana_spec_agent_builder_tool",
			component:    ComponentKibana,
			resourceName: "agent_builder_tool",
			want:         "elasticstack_kibana_agent_builder_tool",
		},
		{
			name:         "kibana_legacy_pilot_agentbuilder_tool",
			component:    ComponentKibana,
			resourceName: "agentbuilder_tool",
			want:         "elasticstack_kibana_agentbuilder_tool",
		},
		{
			name:         "fleet",
			component:    ComponentFleet,
			resourceName: "integration",
			want:         "elasticstack_fleet_integration",
		},
		{
			name:         "apm",
			component:    ComponentAPM,
			resourceName: "agent_configuration",
			want:         "elasticstack_apm_agent_configuration",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			c := New(tc.component, tc.resourceName)
			var resp resource.MetadataResponse
			c.Metadata(context.Background(), resource.MetadataRequest{ProviderTypeName: testProviderTypeName}, &resp)
			require.Equal(t, tc.want, resp.TypeName)
		})
	}
}

func TestCore_Client_nilSafe(t *testing.T) {
	t.Run("nil_receiver", func(t *testing.T) {
		t.Parallel()
		var c *Core
		require.Nil(t, c.Client())
	})

	t.Run("non_nil_before_configure", func(t *testing.T) {
		t.Parallel()
		c := New(ComponentFleet, "integration")
		require.Nil(t, c.Client())
	})
}

// embedCoreTestResource is a minimal [resource.Resource] that embeds [Core] as
// pilot resources will. The [resource.Resource] and [resource.ResourceWithConfigure]
// assignments are compile-time interface checks. The no-import case (embedding
// [Core] does not satisfy [resource.ResourceWithImportState]) is checked at
// runtime in TestEmbedCore_importStateAndConfigure (see subtest
// "no_explicit_import" below in this file), not here.
type embedCoreTestResource struct {
	*Core
}

var (
	_ resource.Resource              = (*embedCoreTestResource)(nil)
	_ resource.ResourceWithConfigure = (*embedCoreTestResource)(nil)
	// [resource.ResourceWithImportState] is not asserted here; see TestEmbedCore_importStateAndConfigure.
)

// embedCoreWithImport is the same shape as a pilot resource that defines its own import.
type embedCoreWithImport struct {
	*Core
}

var (
	_ resource.Resource                = (*embedCoreWithImport)(nil)
	_ resource.ResourceWithConfigure   = (*embedCoreWithImport)(nil)
	_ resource.ResourceWithImportState = (*embedCoreWithImport)(nil)
)

func (r *embedCoreTestResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{Attributes: map[string]schema.Attribute{}}
}

func (r *embedCoreTestResource) Create(context.Context, resource.CreateRequest, *resource.CreateResponse) {
}
func (r *embedCoreTestResource) Read(context.Context, resource.ReadRequest, *resource.ReadResponse) {}
func (r *embedCoreTestResource) Update(context.Context, resource.UpdateRequest, *resource.UpdateResponse) {
}
func (r *embedCoreTestResource) Delete(context.Context, resource.DeleteRequest, *resource.DeleteResponse) {
}

func (r *embedCoreWithImport) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{Attributes: map[string]schema.Attribute{}}
}

func (r *embedCoreWithImport) Create(context.Context, resource.CreateRequest, *resource.CreateResponse) {
}
func (r *embedCoreWithImport) Read(context.Context, resource.ReadRequest, *resource.ReadResponse) {}
func (r *embedCoreWithImport) Update(context.Context, resource.UpdateRequest, *resource.UpdateResponse) {
}
func (r *embedCoreWithImport) Delete(context.Context, resource.DeleteRequest, *resource.DeleteResponse) {
}

func (r *embedCoreWithImport) ImportState(context.Context, resource.ImportStateRequest, *resource.ImportStateResponse) {
}

func TestEmbedCore_importStateAndConfigure(t *testing.T) {
	t.Run("no_explicit_import", func(t *testing.T) {
		t.Parallel()
		r := &embedCoreTestResource{Core: New(ComponentElasticsearch, "x")}
		anyR := any(r)

		_, okCfg := anyR.(resource.ResourceWithConfigure)
		require.True(t, okCfg, "embedded Core should allow ResourceWithConfigure via promoted Configure")

		_, okImp := anyR.(resource.ResourceWithImportState)
		require.False(t, okImp, "Core must not promote ImportState (accidental importability)")
	})

	t.Run("explicit_custom_import", func(t *testing.T) {
		t.Parallel()
		r := &embedCoreWithImport{Core: New(ComponentKibana, "agentbuilder_tool")}
		anyR := any(r)

		_, okImp := anyR.(resource.ResourceWithImportState)
		require.True(t, okImp, "ImportState on the concrete type must still satisfy ResourceWithImportState when Core has none")

		_, okCfg := anyR.(resource.ResourceWithConfigure)
		require.True(t, okCfg, "concrete type with import should still implement ResourceWithConfigure")
	})
}
