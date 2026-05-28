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

package output_test

// TestAccReproduceIssue1856 guards the fix for
// https://github.com/elastic/terraform-provider-elasticstack/issues/1856:
// "Provider produced inconsistent result after apply in elasticstack_fleet_output"
//
// Background: Fleet treats an omitted `config_yaml` in an update request as
// "no change" and echoes the previously stored value (or an empty string for
// outputs that never had one) back in the response. Before the fix the
// provider wrote that API-echoed value into state, while the plan held null
// for the sensitive attribute, tripping Terraform's post-apply consistency
// check with `.config_yaml: inconsistent values for sensitive attribute`.
//
// The fix has two parts:
//  1. The shared reader normalizes nil/empty config_yaml from the API to null
//     state, so outputs that never had a config_yaml don't flip null↔"".
//  2. The Update handler captures the planned config_yaml and restores a
//     plan-null over any API echo, so removing config_yaml from configuration
//     applies cleanly even when Fleet keeps the previously stored value.

import (
	"fmt"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/fleet/output"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccReproduceIssue1856(t *testing.T) {
	versionutils.SkipIfUnsupported(t, output.MinVersionOutputKafka, versionutils.FlavorAny)

	outputID := fmt.Sprintf("issue-1856-%s", sdkacctest.RandString(8))
	vars := config.Variables{
		"output_id": config.StringVariable(outputID),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceOutputDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          vars,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_output.repro_1856", "type", "kafka"),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.repro_1856", "output_id", outputID),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.repro_1856", "config_yaml", "bulk_max_size: 100\n"),
				),
			},
			{
				// With the fix in place this apply must complete cleanly and
				// drop config_yaml from state, mirroring the user's intent —
				// even though Fleet preserves the previously stored value
				// server-side and echoes it back in the update response.
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables:          vars,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_output.repro_1856", "kafka.topic", "test-topic-updated"),
					resource.TestCheckNoResourceAttr("elasticstack_fleet_output.repro_1856", "config_yaml"),
				),
			},
		},
	})
}
