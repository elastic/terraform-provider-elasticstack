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

package anomalydetectionjob_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccReproduceIssue3403 reproduces the "Value Conversion Error" that occurs
// when analysis_config is assigned from an unknown value, as happens with
// for_each (e.g. analysis_config = each.value.job.analysis_config).
//
// The Terraform testing framework does not support for_each, so this test uses
// terraform_data.source.output which is unknown during the first plan (before
// terraform_data is applied) to simulate the same code path.
//
// Root cause: TFModel.AnalysisConfig is typed as *AnalysisConfigTFModel, which
// cannot hold unknown values. The fix requires changing it to basetypes.ObjectValue
// (as suggested in the error: "Suggested Type: basetypes.ObjectValue").
//
// Related to: https://github.com/elastic/terraform-provider-elasticstack/issues/3403
func TestAccReproduceIssue3403(t *testing.T) {
	jobID := fmt.Sprintf("test-ad-repro-3403-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum))

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"job_id": config.StringVariable(jobID),
				},
				// The bug manifests as a plan-time "Value Conversion Error" because
				// *AnalysisConfigTFModel cannot hold the unknown value that
				// terraform_data.source.output has before first apply.
				ExpectError: regexp.MustCompile(`(?i)Value Conversion Error`),
			},
		},
	})
}
