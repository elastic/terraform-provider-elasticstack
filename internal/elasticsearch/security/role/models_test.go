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

package role

import (
	"context"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func TestData_satisfiesElasticsearchResourceModelContract(t *testing.T) {
	t.Parallel()
	var _ entitycore.ElasticsearchResourceModel = Data{}
	var _ entitycore.WithVersionRequirements = Data{}
}

func TestModel_GetVersionRequirements(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	remoteIndicesSet := func(t *testing.T, count int) types.Set {
		t.Helper()
		attrTypes := getRemoteIndexPermsAttrTypes()
		if count == 0 {
			return types.SetValueMust(types.ObjectType{AttrTypes: attrTypes}, []attr.Value{})
		}

		clustersSet := types.SetValueMust(types.StringType, []attr.Value{types.StringValue("remote-cluster")})
		namesSet := types.SetValueMust(types.StringType, []attr.Value{types.StringValue("logs-*")})
		privilegesSet := types.SetValueMust(types.StringType, []attr.Value{types.StringValue("read")})
		remoteIndexObj := types.ObjectValueMust(attrTypes, map[string]attr.Value{
			attrAllowRestrictedIndices: types.BoolNull(),
			attrClusters:               clustersSet,
			attrFieldSecurity:          types.ObjectNull(getFieldSecurityAttrTypes()),
			attrQuery:                  jsontypes.NewNormalizedNull(),
			attrNames:                  namesSet,
			attrPrivileges:             privilegesSet,
		})
		elements := make([]attr.Value, count)
		for i := range elements {
			elements[i] = remoteIndexObj
		}
		return types.SetValueMust(types.ObjectType{AttrTypes: attrTypes}, elements)
	}

	t.Run("neither configured", func(t *testing.T) {
		t.Parallel()
		data := Data{
			Description:   types.StringNull(),
			RemoteIndices: types.SetNull(types.ObjectType{AttrTypes: getRemoteIndexPermsAttrTypes()}),
		}
		reqs, diags := data.GetVersionRequirements(ctx)
		require.False(t, diags.HasError())
		require.Empty(t, reqs)
	})

	t.Run("description empty string", func(t *testing.T) {
		t.Parallel()
		data := Data{
			Description: types.StringValue(""),
		}
		reqs, diags := data.GetVersionRequirements(ctx)
		require.False(t, diags.HasError())
		require.Len(t, reqs, 1)
		require.True(t, reqs[0].MinVersion.Equal(MinSupportedDescriptionVersion))
		require.Contains(t, reqs[0].ErrorMessage, "'description'")
		require.Contains(t, reqs[0].ErrorMessage, MinSupportedDescriptionVersion.String())
	})

	t.Run("description only", func(t *testing.T) {
		t.Parallel()
		data := Data{
			Description:   types.StringValue("role description"),
			RemoteIndices: remoteIndicesSet(t, 0),
		}
		reqs, diags := data.GetVersionRequirements(ctx)
		require.False(t, diags.HasError())
		require.Len(t, reqs, 1)
		require.True(t, reqs[0].MinVersion.Equal(MinSupportedDescriptionVersion))
		require.Contains(t, reqs[0].ErrorMessage, "'description'")
	})

	t.Run("remote_indices empty set, no description", func(t *testing.T) {
		t.Parallel()
		data := Data{
			Description:   types.StringNull(),
			RemoteIndices: remoteIndicesSet(t, 0),
		}
		reqs, diags := data.GetVersionRequirements(ctx)
		require.False(t, diags.HasError())
		require.Empty(t, reqs)
	})

	t.Run("remote_indices only", func(t *testing.T) {
		t.Parallel()
		data := Data{
			Description:   types.StringNull(),
			RemoteIndices: remoteIndicesSet(t, 1),
		}
		reqs, diags := data.GetVersionRequirements(ctx)
		require.False(t, diags.HasError())
		require.Len(t, reqs, 1)
		require.True(t, reqs[0].MinVersion.Equal(MinSupportedRemoteIndicesVersion))
		require.Contains(t, reqs[0].ErrorMessage, "'remote_indices'")
	})

	t.Run("both configured", func(t *testing.T) {
		t.Parallel()
		data := Data{
			Description:   types.StringValue("role description"),
			RemoteIndices: remoteIndicesSet(t, 1),
		}
		reqs, diags := data.GetVersionRequirements(ctx)
		require.False(t, diags.HasError())
		require.Len(t, reqs, 2)
		require.True(t, reqs[0].MinVersion.Equal(MinSupportedDescriptionVersion))
		require.Contains(t, reqs[0].ErrorMessage, "'description'")
		require.Contains(t, reqs[0].ErrorMessage, MinSupportedDescriptionVersion.String())
		require.True(t, reqs[1].MinVersion.Equal(MinSupportedRemoteIndicesVersion))
		require.Contains(t, reqs[1].ErrorMessage, "'remote_indices'")
		require.Contains(t, reqs[1].ErrorMessage, MinSupportedRemoteIndicesVersion.String())
	})
}

func TestFromAPIModel_PreservesEmptyStringDescriptionWhenAPIIsNull(t *testing.T) {
	ctx := context.Background()

	d := Data{
		Name:        types.StringValue("role-a"),
		Description: types.StringValue(""),
	}

	diags := d.fromAPIModel(ctx, &elasticsearch.Role{
		Description: nil,
	})
	require.False(t, diags.HasError(), "unexpected diags: %#v", diags)

	require.False(t, d.Description.IsNull())
	require.Empty(t, d.Description.ValueString())
}
