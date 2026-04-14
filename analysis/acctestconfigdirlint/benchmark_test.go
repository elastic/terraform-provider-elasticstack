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

// BenchmarkAnalyzer_Compliant benchmarks the analyzer on a package of compliant acceptance tests.
// This exercises the narrowed traversal path (only _test.go files, only candidate calls).
func BenchmarkAnalyzer_Compliant(b *testing.B) {
	testdata := analysistest.TestData()
	for b.Loop() {
		analysistest.Run(b, testdata, NewAnalyzer(), "github.com/elastic/terraform-provider-elasticstack/internal/acctestcases/compliant")
	}
}

// BenchmarkAnalyzer_Violations benchmarks the analyzer on a package of violating acceptance tests.
// This exercises the diagnostic-reporting paths in addition to narrowed traversal.
func BenchmarkAnalyzer_Violations(b *testing.B) {
	testdata := analysistest.TestData()
	for b.Loop() {
		analysistest.Run(b, testdata, NewAnalyzer(), "github.com/elastic/terraform-provider-elasticstack/internal/acctestcases/violations")
	}
}
