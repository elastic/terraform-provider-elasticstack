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

package spacesettings

import (
	"context"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func prefixSet(prefixes ...string) types.Set {
	elements := make([]attr.Value, len(prefixes))
	for i, prefix := range prefixes {
		elements[i] = types.StringValue(prefix)
	}
	return types.SetValueMust(types.StringType, elements)
}

func TestPopulateFromAPI_setsIdentityPrefixesAndManagedBy(t *testing.T) {
	t.Parallel()

	var m spaceSettingsModel
	diags := m.populateFromAPI(context.Background(), "team-a", &fleet.SpaceSettings{
		AllowedNamespacePrefixes: []string{"team_a", "shared"},
		ManagedBy:                new("kibana"),
	})

	require.False(t, diags.HasError(), "%v", diags)
	require.Equal(t, types.StringValue("team-a"), m.ID)
	require.Equal(t, types.StringValue("team-a"), m.SpaceID)
	require.Equal(t, prefixSet("team_a", "shared"), m.AllowedNamespacePrefixes)
	require.Equal(t, types.StringValue("kibana"), m.ManagedBy)
}

func TestPopulateFromAPI_missingPrefixesBecomeEmptySet(t *testing.T) {
	t.Parallel()

	var m spaceSettingsModel
	diags := m.populateFromAPI(context.Background(), "team-a", &fleet.SpaceSettings{})

	require.False(t, diags.HasError(), "%v", diags)
	require.False(t, m.AllowedNamespacePrefixes.IsNull())
	require.Equal(t, prefixSet(), m.AllowedNamespacePrefixes)
}

func TestPopulateFromAPI_absentManagedByIsNull(t *testing.T) {
	t.Parallel()

	var m spaceSettingsModel
	diags := m.populateFromAPI(context.Background(), "team-a", &fleet.SpaceSettings{AllowedNamespacePrefixes: []string{"team_a"}})

	require.False(t, diags.HasError(), "%v", diags)
	require.True(t, m.ManagedBy.IsNull())
}

func TestGetVersionRequirements_requiresStack910(t *testing.T) {
	t.Parallel()

	reqs, diags := spaceSettingsModel{}.GetVersionRequirements(context.Background())

	require.False(t, diags.HasError(), "%v", diags)
	require.Len(t, reqs, 1)
	require.Equal(t, "9.1.0", reqs[0].MinVersion.String())
	require.NotEmpty(t, reqs[0].ErrorMessage)
}

func TestPrefixesToWrite(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		prefixes types.Set
		want     []string
	}{
		{name: "configured prefixes are all written", prefixes: prefixSet("team_a", "shared"), want: []string{"team_a", "shared"}},
		{name: "empty set is written as an empty slice", prefixes: prefixSet(), want: []string{}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, diags := spaceSettingsModel{AllowedNamespacePrefixes: tc.prefixes}.prefixesToWrite(context.Background())

			require.False(t, diags.HasError(), "%v", diags)
			require.Equal(t, tc.want, got)
		})
	}
}
