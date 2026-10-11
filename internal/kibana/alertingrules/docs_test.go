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

package alertingrules

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGeneratedDocs_describeInputsRulesAndReplacement(t *testing.T) {
	t.Parallel()

	_, thisFile, _, ok := runtime.Caller(0)
	require.True(t, ok)
	repoRoot := filepath.Clean(filepath.Join(filepath.Dir(thisFile), "..", "..", ".."))
	page, err := os.ReadFile(filepath.Join(repoRoot, "docs", "data-sources", "kibana_alerting_rules.md"))
	require.NoError(t, err)

	body := string(page)
	for _, phrase := range []string{
		"space_id",
		"rule_id",
		"filter",
		"last_execution_status",
		"last_execution_date",
		"scheduled_task_id",
		"null",
		"many rules",
		"removed",
		`data "elasticstack_kibana_alerting_rules"`,
	} {
		require.Contains(t, strings.ToLower(body), strings.ToLower(phrase))
	}
}
