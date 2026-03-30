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

func TestAnalyzer_DefaultConfig(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, NewAnalyzer(Config{}), "github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/cases/default")
}

func TestAnalyzer_AllowedWrappers(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, NewAnalyzer(Config{
		Wrappers: []string{
			"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/cases/helpers.NewSDKClient",
			"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/cases/helpers.NewFrameworkClient",
		},
	}), "github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/cases/wrappers")
}

func TestAnalyzer_ControlFlowAndAssignment(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, NewAnalyzer(Config{}), "github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/cases/flow")
}

func TestAnalyzer_MultiParamSinks(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, NewAnalyzer(Config{}), "github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/cases/args")
}

func TestAnalyzer_ScopeBoundary(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, NewAnalyzer(Config{}), "github.com/elastic/terraform-provider-elasticstack/internal/kibana/cases/scope")
}

func TestAnalyzer_InferredWrapperFacts(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, NewAnalyzer(Config{}), "github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/cases/facts")
}
