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

package lensxy_test

import (
	"regexp"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/dashboardacctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccReproduceIssue3707 covers the regression where using operation="percentile"
// (with a numeric percentile value) in a bar_horizontal layer's y config_json
// caused the provider to terminate the shell (panic).
//
// Related to: https://github.com/elastic/terraform-provider-elasticstack/issues/3707
func TestAccReproduceIssue3707(t *testing.T) {
	dashboardTitle := "Repro Issue 3707 percentile bar " + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)

	versionutils.SkipIfUnsupported(t, dashboardacctest.MinDashboardAPISupport, versionutils.FlavorAny)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				// The provider injects empty_as_null=false into percentile metric configs via
				// PopulateLensMetricDefaults, but the Kibana API rejects that field for percentile
				// operations. This causes an HTTP 400 instead of a clean apply.
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("percentile_bar_horizontal"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
				},
				ExpectError: regexp.MustCompile(`(?i)(400|empty_as_null|additional properties|percentile)`),
			},
		},
	})
}
