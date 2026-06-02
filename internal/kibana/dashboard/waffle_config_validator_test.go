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

	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panel/lenswaffle"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func Test_waffleModeListStateFromTF_unknownSkipsCountRules(t *testing.T) {
	unknown := types.ListUnknown(types.StringType)
	st := waffleModeListStateFromTF(unknown)
	require.True(t, st.Unknown)

	diags := lenswaffle.WaffleConfigModeValidateDiags(false,
		st,
		lenswaffle.WaffleModeListStateFromSlice(0),
		lenswaffle.WaffleModeListStateFromSlice(0),
		lenswaffle.WaffleModeListStateFromSlice(0),
	)
	require.False(t, diags.HasError(), "unknown metrics must not trigger Missing metrics")

	diags2 := lenswaffle.WaffleConfigModeValidateDiags(true,
		lenswaffle.WaffleModeListStateFromSlice(0),
		lenswaffle.WaffleModeListStateFromSlice(0),
		st,
		lenswaffle.WaffleModeListStateFromSlice(0),
	)
	require.False(t, diags2.HasError(), "unknown esql_metrics must not trigger Missing esql_metrics")
}

func Test_waffleConfigModeValidateDiags_crossMode(t *testing.T) {
	diags := lenswaffle.WaffleConfigModeValidateDiags(true,
		lenswaffle.WaffleModeListStateFromSlice(1),
		lenswaffle.WaffleModeListStateFromSlice(0),
		lenswaffle.WaffleModeListStateFromSlice(1),
		lenswaffle.WaffleModeListStateFromSlice(0),
	)
	require.True(t, diags.HasError())
	require.Len(t, diags, 1)
	require.Contains(t, diags[0].Detail(), "metrics")
}
