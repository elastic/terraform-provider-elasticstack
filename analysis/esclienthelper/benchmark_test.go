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

package esclienthelper

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

// BenchmarkAnalyzer_DefaultConfig benchmarks the analyzer on the default sink detection cases.
// This exercises the in-scope file precomputation and fact-export phases.
func BenchmarkAnalyzer_DefaultConfig(b *testing.B) {
	testdata := analysistest.TestData()
	for b.Loop() {
		analysistest.Run(b, testdata, NewAnalyzer(Config{}), "github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/cases/default")
	}
}

// BenchmarkAnalyzer_InferredWrapperFacts benchmarks the analyzer on packages that exercise
// the fact-import path for wrapper functions. This is the hot path for the cached metadata work.
func BenchmarkAnalyzer_InferredWrapperFacts(b *testing.B) {
	testdata := analysistest.TestData()
	for b.Loop() {
		analysistest.Run(b, testdata, NewAnalyzer(Config{}), "github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/cases/facts")
	}
}

// BenchmarkAnalyzer_ControlFlow benchmarks the analyzer on packages with branching control flow
// and multiple assignment/reassignment patterns.
func BenchmarkAnalyzer_ControlFlow(b *testing.B) {
	testdata := analysistest.TestData()
	for b.Loop() {
		analysistest.Run(b, testdata, NewAnalyzer(Config{}), "github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/cases/flow")
	}
}
