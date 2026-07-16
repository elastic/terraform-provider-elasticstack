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
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLensChartBaseFieldsForAPI_AllKnown(t *testing.T) {
	t.Parallel()
	m := models.LensChartBaseTFModel{
		Title:               types.StringValue("my title"),
		Description:         types.StringValue("my description"),
		IgnoreGlobalFilters: types.BoolValue(true),
		Sampling:            types.Float64Value(0.5),
	}
	title, desc, ignoreGlobalFilters, sampling := LensChartBaseFieldsForAPI(m)
	require.NotNil(t, title)
	require.NotNil(t, desc)
	require.NotNil(t, ignoreGlobalFilters)
	require.NotNil(t, sampling)
	assert.Equal(t, "my title", *title)
	assert.Equal(t, "my description", *desc)
	assert.True(t, *ignoreGlobalFilters)
	assert.InDelta(t, float32(0.5), *sampling, 1e-6)
}

func TestLensChartBaseFieldsForAPI_AllNull(t *testing.T) {
	t.Parallel()
	m := models.LensChartBaseTFModel{
		Title:               types.StringNull(),
		Description:         types.StringNull(),
		IgnoreGlobalFilters: types.BoolNull(),
		Sampling:            types.Float64Null(),
	}
	title, desc, ignoreGlobalFilters, sampling := LensChartBaseFieldsForAPI(m)
	assert.Nil(t, title)
	assert.Nil(t, desc)
	assert.Nil(t, ignoreGlobalFilters)
	assert.Nil(t, sampling)
}

func TestLensChartBaseFieldsForAPI_AllUnknown(t *testing.T) {
	t.Parallel()
	m := models.LensChartBaseTFModel{
		Title:               types.StringUnknown(),
		Description:         types.StringUnknown(),
		IgnoreGlobalFilters: types.BoolUnknown(),
		Sampling:            types.Float64Unknown(),
	}
	title, desc, ignoreGlobalFilters, sampling := LensChartBaseFieldsForAPI(m)
	assert.Nil(t, title)
	assert.Nil(t, desc)
	assert.Nil(t, ignoreGlobalFilters)
	assert.Nil(t, sampling)
}

func TestLensChartBaseFieldsForAPI_Partial(t *testing.T) {
	t.Parallel()
	m := models.LensChartBaseTFModel{
		Title:               types.StringValue("partial"),
		Description:         types.StringNull(),
		IgnoreGlobalFilters: types.BoolNull(),
		Sampling:            types.Float64Value(1.0),
	}
	title, desc, ignoreGlobalFilters, sampling := LensChartBaseFieldsForAPI(m)
	require.NotNil(t, title)
	assert.Equal(t, "partial", *title)
	assert.Nil(t, desc)
	assert.Nil(t, ignoreGlobalFilters)
	require.NotNil(t, sampling)
	assert.InDelta(t, float32(1.0), *sampling, 1e-6)
}
