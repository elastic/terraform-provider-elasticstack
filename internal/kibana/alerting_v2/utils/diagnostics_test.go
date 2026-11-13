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

package utils_test

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/alerting_v2/utils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/stretchr/testify/require"
)

func TestAlreadyExistsDiagnostic_recommendsImport(t *testing.T) {
	t.Parallel()

	d := utils.AlreadyExistsDiagnostic("rule", "team-a/my-rule")

	require.Equal(t, diag.SeverityError, d.Severity())
	require.Contains(t, d.Summary(), "already exists")
	require.Contains(t, d.Detail(), "team-a/my-rule")
	require.Contains(t, d.Detail(), "terraform import")
	require.NotContains(t, d.Detail(), "replaced")
}

func TestAlreadyReplacedDiagnostic_statesReplacementHappened(t *testing.T) {
	t.Parallel()

	d := utils.AlreadyReplacedDiagnostic("rule", "team-a/my-rule")

	require.Equal(t, diag.SeverityError, d.Severity())
	require.Contains(t, d.Summary(), "already exists")
	require.Contains(t, d.Detail(), "team-a/my-rule")
	require.Contains(t, d.Detail(), "terraform import")
	require.Contains(t, d.Detail(), "replaced")
}
