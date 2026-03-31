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

package acctestconfigdirlint

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

// TestAnalyzer_CompliantCases verifies that fully compliant test steps produce no diagnostics.
// Covered patterns:
//   - ordinary step: ConfigDirectory: acctest.NamedTestCaseDirectory(...) inside resource.Test
//   - ordinary step: ConfigDirectory: acctest.NamedTestCaseDirectory(...) inside resource.ParallelTest
//   - compatibility step: ExternalProviders + Config: "..." inside resource.Test
//   - import-only step: neither Config nor ConfigDirectory
func TestAnalyzer_CompliantCases(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, NewAnalyzer(), "github.com/elastic/terraform-provider-elasticstack/internal/acctestcases/compliant")
}

// TestAnalyzer_Violations verifies that non-compliant test steps each produce the expected diagnostic.
// Covered violations:
//   - inline Config without ExternalProviders (violation 1)
//   - ConfigDirectory set to config.TestNameDirectory() instead of acctest.NamedTestCaseDirectory (violation 2)
//   - ExternalProviders + ConfigDirectory together (violation 4)
//   - inline Config without ExternalProviders inside resource.ParallelTest (violation 1, parallel variant)
func TestAnalyzer_Violations(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, NewAnalyzer(), "github.com/elastic/terraform-provider-elasticstack/internal/acctestcases/violations")
}
