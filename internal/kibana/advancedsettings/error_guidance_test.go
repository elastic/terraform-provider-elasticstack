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

package advancedsettings

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWithJSONSettingsHint_advisesOnEveryErrorAndLeavesOtherDiagnosticsAlone(t *testing.T) {
	t.Parallel()
	var diags diag.Diagnostics
	diags.AddError("first failure", "Kibana rejected the first change")
	diags.AddWarning("a warning", "nothing to do with types")
	diags.AddError("second failure", "Kibana rejected the second change")

	hinted := withJSONSettingsHint(diags)

	require.Len(t, hinted, 3)
	require.Len(t, hinted.Errors(), 2)
	assert.Equal(t, "first failure", hinted.Errors()[0].Summary())
	assert.Contains(t, hinted.Errors()[0].Detail(), "Kibana rejected the first change")
	assert.Contains(t, hinted.Errors()[0].Detail(), "non-string values")
	assert.Contains(t, hinted.Errors()[0].Detail(), "json_settings")
	assert.Equal(t, "second failure", hinted.Errors()[1].Summary())
	assert.Contains(t, hinted.Errors()[1].Detail(), "Kibana rejected the second change")
	assert.Contains(t, hinted.Errors()[1].Detail(), "json_settings")
	require.Len(t, hinted.Warnings(), 1)
	assert.Equal(t, "nothing to do with types", hinted.Warnings()[0].Detail())
}
