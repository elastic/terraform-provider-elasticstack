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

package rule

import (
	"context"
	"reflect"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/elastic/terraform-provider-elasticstack/internal/providerfwtest"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func importState(t *testing.T, importID string) *resource.ImportStateResponse {
	t.Helper()
	r, ok := any(newResource()).(resource.ResourceWithImportState)
	require.True(t, ok)
	resp := &resource.ImportStateResponse{State: providerfwtest.EmptyImportState(t, r)}
	r.ImportState(context.Background(), resource.ImportStateRequest{ID: importID}, resp)
	return resp
}

func TestResource_importState_rejectsMalformedIDsBeforeWritingState(t *testing.T) {
	t.Parallel()

	for name, importID := range map[string]string{
		"no separator":    "my-rule",
		"empty space":     "/my-rule",
		"empty rule id":   "team-a/",
		"only separator":  "/",
		"nested segments": "team-a/my/rule",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			resp := importState(t, importID)

			require.True(t, resp.Diagnostics.HasError())
			require.Contains(t, resp.Diagnostics[0].Detail(), "<space_id>/<rule_id>")

			var id types.String
			require.False(t, resp.State.GetAttribute(context.Background(), path.Root("id"), &id).HasError())
			require.True(t, id.IsNull(), "id must not be written to state")
		})
	}
}

func TestResource_importState_seedsCompositeIdentity(t *testing.T) {
	t.Parallel()

	resp := importState(t, "team-a/my-rule")
	require.False(t, resp.Diagnostics.HasError(), "unexpected diagnostics: %v", resp.Diagnostics)

	ctx := context.Background()
	var id, spaceID, ruleID types.String
	require.False(t, resp.State.GetAttribute(ctx, path.Root("id"), &id).HasError())
	require.False(t, resp.State.GetAttribute(ctx, path.Root("space_id"), &spaceID).HasError())
	require.False(t, resp.State.GetAttribute(ctx, path.Root("rule_id"), &ruleID).HasError())
	require.Equal(t, "team-a/my-rule", id.ValueString())
	require.Equal(t, "team-a", spaceID.ValueString())
	require.Equal(t, "my-rule", ruleID.ValueString())
}

func TestResource_embedsEntityCoreKibanaResource(t *testing.T) {
	t.Parallel()

	rt := reflect.TypeFor[Resource]()
	field, ok := rt.FieldByName("KibanaResource")
	require.True(t, ok)
	require.True(t, field.Anonymous)
	require.Equal(t, reflect.TypeFor[*entitycore.KibanaResource[Model]](), field.Type)
}

func TestResource_metadata_typeName(t *testing.T) {
	t.Parallel()

	resp := &resource.MetadataResponse{}
	newResource().Metadata(context.Background(), resource.MetadataRequest{ProviderTypeName: "elasticstack"}, resp)

	require.Equal(t, "elasticstack_kibana_alerting_v2_rule", resp.TypeName)
}
