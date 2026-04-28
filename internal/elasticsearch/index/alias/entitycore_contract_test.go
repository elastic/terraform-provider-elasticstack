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

package alias

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

func TestResource_embedsEntityCoreResourceBase(t *testing.T) {
	t.Parallel()
	rt := reflect.TypeFor[aliasResource]()
	field, ok := rt.FieldByName("ResourceBase")
	require.True(t, ok)
	require.True(t, field.Anonymous)
	require.Equal(t, reflect.TypeFor[*entitycore.ResourceBase](), field.Type)
}

func TestAliasResource_importState_passthroughCompoundID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	r, ok := any(newAliasResource()).(resource.ResourceWithImportState)
	require.True(t, ok)
	st := providerfwtest.EmptyImportState(t, r)
	resp := &resource.ImportStateResponse{State: st}

	const importID = "cluster/uuid/alias/name"
	r.ImportState(ctx, resource.ImportStateRequest{ID: importID}, resp)
	require.False(t, resp.Diagnostics.HasError())

	var id types.String
	resp.Diagnostics.Append(resp.State.GetAttribute(ctx, path.Root("id"), &id)...)
	require.False(t, resp.Diagnostics.HasError())
	require.Equal(t, importID, id.ValueString())
}
