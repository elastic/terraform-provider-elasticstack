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

package dashboard

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func Test_waffleModeListStateFromTF_unknownSkipsCountRules(t *testing.T) {
	unknown := types.ListUnknown(types.StringType)
	st := waffleModeListStateFromTF(unknown)
	require.True(t, st.Unknown)

	diags := waffleConfigModeValidateDiags(false,
		st,
		waffleModeListStateFromSlice(0),
		waffleModeListStateFromSlice(0),
		waffleModeListStateFromSlice(0),
		nil,
	)
	require.False(t, diags.HasError(), "unknown metrics must not trigger Missing metrics")

	diags2 := waffleConfigModeValidateDiags(true,
		waffleModeListStateFromSlice(0),
		waffleModeListStateFromSlice(0),
		st,
		waffleModeListStateFromSlice(0),
		nil,
	)
	require.False(t, diags2.HasError(), "unknown esql_metrics must not trigger Missing esql_metrics")
}

func Test_waffleConfigModeValidateDiags_crossMode(t *testing.T) {
	diags := waffleConfigModeValidateDiags(true,
		waffleModeListStateFromSlice(1),
		waffleModeListStateFromSlice(0),
		waffleModeListStateFromSlice(1),
		waffleModeListStateFromSlice(0),
		nil,
	)
	require.True(t, diags.HasError())
	require.Len(t, diags, 1)
	require.Contains(t, diags[0].Detail(), "metrics")
}
