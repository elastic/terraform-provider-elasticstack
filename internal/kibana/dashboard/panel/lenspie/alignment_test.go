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

package lenspie

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_alignPieConfigStateFromPlan_preservesNullLegendTruncateAndNestedWhenKibanaInjectsDefaults(t *testing.T) {
	t.Parallel()

	plan := &models.PieChartConfigModel{
		Legend: &models.PartitionLegendModel{
			Size:              types.StringValue("auto"),
			Visible:           types.StringValue("auto"),
			TruncateAfterLine: types.Int64Null(),
			Nested:            types.BoolNull(),
		},
	}
	state := &models.PieChartConfigModel{
		Legend: &models.PartitionLegendModel{
			Size:              types.StringValue("auto"),
			Visible:           types.StringValue("auto"),
			TruncateAfterLine: types.Int64Value(1),
			Nested:            types.BoolValue(false),
		},
	}

	alignPieConfigStateFromPlan(t.Context(), plan, state)

	require.NotNil(t, state.Legend)
	assert.True(t, state.Legend.TruncateAfterLine.IsNull())
	assert.True(t, state.Legend.Nested.IsNull())
}

func Test_alignPieConfigStateFromPlan_doesNotOverwriteExplicitLegendTruncateAndNested(t *testing.T) {
	t.Parallel()

	plan := &models.PieChartConfigModel{
		Legend: &models.PartitionLegendModel{
			Size:              types.StringValue("auto"),
			Visible:           types.StringValue("visible"),
			TruncateAfterLine: types.Int64Value(5),
			Nested:            types.BoolValue(true),
		},
	}
	state := &models.PieChartConfigModel{
		Legend: &models.PartitionLegendModel{
			Size:              types.StringValue("auto"),
			Visible:           types.StringValue("visible"),
			TruncateAfterLine: types.Int64Value(5),
			Nested:            types.BoolValue(true),
		},
	}

	alignPieConfigStateFromPlan(t.Context(), plan, state)

	require.NotNil(t, state.Legend)
	assert.Equal(t, int64(5), state.Legend.TruncateAfterLine.ValueInt64())
	assert.True(t, state.Legend.Nested.ValueBool())
}
