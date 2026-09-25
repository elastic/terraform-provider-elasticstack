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

package lensdatatable

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPopulateDatatableLensAttributes_injectsMetricVisibleAlignmentWithoutAxis(t *testing.T) {
	t.Parallel()

	attrs := populateDatatableLensAttributes(map[string]any{
		"metrics": []any{
			map[string]any{"operation": "count"},
		},
		"rows": []any{
			map[string]any{"operation": "terms", "field": "host.name"},
		},
	})

	metrics := attrs["metrics"].([]any)
	require.Len(t, metrics, 1)
	metric := metrics[0].(map[string]any)
	assert.Equal(t, true, metric["visible"])
	assert.Equal(t, "right", metric["alignment"])
	_, hasAxis := metric["axis"]
	assert.False(t, hasAxis)

	rows := attrs["rows"].([]any)
	require.Len(t, rows, 1)
	_, rowHasAxis := rows[0].(map[string]any)["axis"]
	assert.False(t, rowHasAxis)
}
