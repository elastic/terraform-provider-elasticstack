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

package lenscommon

import (
	"encoding/json"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func TestPartitionEsqlGroupByCollapseByRoundTrip(t *testing.T) {
	var collapseBy kbapi.KibanaHTTPAPIsCollapseBy
	require.NoError(t, json.Unmarshal([]byte(`"avg"`), &collapseBy))

	fromAPIDiags := diag.Diagnostics{}
	fromAPI := PopulatePartitionEsqlGroupByFromAPI([]EsqlGroupByAPIFields{{
		Column:     "service.name",
		CollapseBy: &collapseBy,
	}}, &fromAPIDiags)
	require.False(t, fromAPIDiags.HasError(), "%s", fromAPIDiags)
	require.Equal(t, "avg", fromAPI[0].CollapseBy.ValueString())

	toAPIDiags := diag.Diagnostics{}
	toAPI := BuildPartitionEsqlGroupByForAPI([]models.PartitionEsqlGroupByModel{{
		Column:     types.StringValue("service.name"),
		CollapseBy: types.StringValue("sum"),
		ColorJSON:  jsontypes.NewNormalizedValue(`{}`),
	}}, &toAPIDiags)
	require.False(t, toAPIDiags.HasError(), "%s", toAPIDiags)
	require.NotNil(t, toAPI[0].CollapseBy)

	collapseByJSON, err := json.Marshal(toAPI[0].CollapseBy)
	require.NoError(t, err)
	require.JSONEq(t, `"sum"`, string(collapseByJSON))
}
