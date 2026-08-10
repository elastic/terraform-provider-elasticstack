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

package alertingrule

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func writeTempGuide(t *testing.T, body string) string {
	t.Helper()
	fpath := filepath.Join(t.TempDir(), "guide.md")
	require.NoError(t, os.WriteFile(fpath, []byte(body), 0o600))
	return fpath
}

// nil / inline / unknown-path guides are all no-ops for checksum drift.
func Test_contentPathChecksumChanged_noOpCases(t *testing.T) {
	cases := []struct {
		name string
		ig   *investigationGuideModel
	}{
		{name: "nil guide", ig: nil},
		{
			name: "inline content (content_path null)",
			ig: &investigationGuideModel{
				Content:     types.StringValue("inline"),
				ContentPath: types.StringNull(),
			},
		},
		{
			name: "unknown content_path is skipped (not read)",
			ig: &investigationGuideModel{
				Content:     types.StringNull(),
				ContentPath: types.StringUnknown(),
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// hasPriorState=false would otherwise force changed=true; asserting
			// false proves these branches short-circuit before that check.
			changed, diags := contentPathChecksumChanged(tc.ig, "", false)
			require.False(t, diags.HasError())
			require.False(t, changed)
		})
	}
}

func Test_contentPathChecksumChanged_createMarksChanged(t *testing.T) {
	fpath := writeTempGuide(t, "body")
	ig := &investigationGuideModel{
		Content:     types.StringNull(),
		ContentPath: types.StringValue(fpath),
	}
	// No prior state -> create -> always changed.
	changed, diags := contentPathChecksumChanged(ig, "", false)
	require.False(t, diags.HasError())
	require.True(t, changed)
}

func Test_contentPathChecksumChanged_unchangedWhenChecksumMatches(t *testing.T) {
	fpath := writeTempGuide(t, "stable body")
	prior, err := fileSHA256(fpath)
	require.NoError(t, err)

	ig := &investigationGuideModel{
		Content:     types.StringNull(),
		ContentPath: types.StringValue(fpath),
	}
	changed, diags := contentPathChecksumChanged(ig, prior, true)
	require.False(t, diags.HasError())
	require.False(t, changed)
}

func Test_contentPathChecksumChanged_changedWhenFileDiffersFromState(t *testing.T) {
	fpath := writeTempGuide(t, "new body")
	ig := &investigationGuideModel{
		Content:     types.StringNull(),
		ContentPath: types.StringValue(fpath),
	}
	changed, diags := contentPathChecksumChanged(ig, "stale-checksum", true)
	require.False(t, diags.HasError())
	require.True(t, changed)
}

func Test_contentPathChecksumChanged_missingFileErrors(t *testing.T) {
	ig := &investigationGuideModel{
		Content:     types.StringNull(),
		ContentPath: types.StringValue(filepath.Join(t.TempDir(), "missing.md")),
	}
	changed, diags := contentPathChecksumChanged(ig, "", false)
	require.True(t, diags.HasError())
	require.False(t, changed)
}
