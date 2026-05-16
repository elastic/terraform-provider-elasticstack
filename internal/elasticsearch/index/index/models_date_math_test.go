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

package index

import (
	"context"
	"testing"

	estypes "github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPopulateFromAPI_PreservesConfiguredName verifies that when the state already
// holds a configured name (e.g. a date math expression), populateFromAPI does NOT
// overwrite it with the concrete index name returned by Elasticsearch.
func TestPopulateFromAPI_PreservesConfiguredName(t *testing.T) {
	ctx := context.Background()
	dateMathName := `<logs-{now/d}>`
	concreteName := "logs-2024.01.15"

	model := tfModel{
		Name:         basetypes.NewStringValue(dateMathName),
		Alias:        basetypes.NewSetNull(basetypes.ObjectType{}),
		Settings:     basetypes.NewListNull(basetypes.ObjectType{}),
		ConcreteName: basetypes.NewStringNull(),
	}

	diags := model.populateFromAPI(ctx, concreteName, estypes.IndexState{})
	require.False(t, diags.HasError(), "unexpected error: %v", diags)

	// The configured date math expression must be preserved.
	assert.Equal(t, dateMathName, model.Name.ValueString(), "Name should remain the configured date math expression")
	// The concrete name should be set to the concrete index.
	assert.Equal(t, concreteName, model.ConcreteName.ValueString(), "ConcreteName should be the concrete index from Elasticsearch")
}

// TestPopulateFromAPI_BackfillsNameWhenAbsent verifies that when no name is
// present in state (e.g. after an import), populateFromAPI backfills name from
// the concrete index name so the resource remains readable.
func TestPopulateFromAPI_BackfillsNameWhenAbsent(t *testing.T) {
	ctx := context.Background()
	concreteName := "logs-2024.01.15"

	model := tfModel{
		Name:         basetypes.NewStringNull(),
		Alias:        basetypes.NewSetNull(basetypes.ObjectType{}),
		Settings:     basetypes.NewListNull(basetypes.ObjectType{}),
		ConcreteName: basetypes.NewStringNull(),
	}

	diags := model.populateFromAPI(ctx, concreteName, estypes.IndexState{})
	require.False(t, diags.HasError(), "unexpected error: %v", diags)

	// Both name and concrete_name should be set to the concrete index.
	assert.Equal(t, concreteName, model.Name.ValueString(), "Name should be backfilled from concrete name when absent")
	assert.Equal(t, concreteName, model.ConcreteName.ValueString(), "ConcreteName should be the concrete index from Elasticsearch")
}

// TestPopulateFromAPI_StaticName verifies that static index names are handled
// identically before and after this change: both name and concrete_name reflect
// the same value.
func TestPopulateFromAPI_StaticName(t *testing.T) {
	ctx := context.Background()
	staticName := "my-static-index"

	model := tfModel{
		Name:         basetypes.NewStringValue(staticName),
		Alias:        basetypes.NewSetNull(basetypes.ObjectType{}),
		Settings:     basetypes.NewListNull(basetypes.ObjectType{}),
		ConcreteName: basetypes.NewStringNull(),
	}

	diags := model.populateFromAPI(ctx, staticName, estypes.IndexState{})
	require.False(t, diags.HasError(), "unexpected error: %v", diags)

	assert.Equal(t, staticName, model.Name.ValueString())
	assert.Equal(t, staticName, model.ConcreteName.ValueString())
}
