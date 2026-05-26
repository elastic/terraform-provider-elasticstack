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

// TestAccReproduceIssue1856 reproduces the bug reported in
// https://github.com/elastic/terraform-provider-elasticstack/issues/1856:
// "Provider produced inconsistent result after apply in elasticstack_fleet_output"
//
// Root cause: when the Fleet API returns a previously stored config_yaml value
// in the update response (even though the update request omitted config_yaml),
// the provider's Update handler overwrites the plan's null ConfigYaml with the
// API-returned non-null value before writing state.  Because config_yaml is
// marked Sensitive, Terraform detects the null-vs-value discrepancy and emits:
//   ".config_yaml: inconsistent values for sensitive attribute"
//
// Reproduction steps:
//  1. Create a Kafka output with config_yaml set.
//  2. Remove config_yaml from the config and apply (in-place update).
//     The API preserves the stored value and echoes it back; the provider
//     writes it into state while the plan expected null → inconsistency error.

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/fleet/output"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccReproduceIssue1856(t *testing.T) {
	versionutils.SkipIfUnsupported(t, output.MinVersionOutputKafka, versionutils.FlavorAny)

	outputID := fmt.Sprintf("issue-1856-%s", sdkacctest.RandString(8))

	// Step 1: create a Kafka output with config_yaml set.
	configWithYaml := fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_fleet_output" "repro_1856" {
  output_id            = %q
  name                 = "Issue 1856 Repro"
  type                 = "kafka"
  default_integrations = false
  default_monitoring   = false
  hosts                = ["kafka:9092"]
  config_yaml          = "bulk_max_size: 100\n"

  kafka = {
    auth_type       = "none"
    connection_type = "plaintext"
    topic           = "test-topic"
    partition       = "round_robin"
  }
}
`, outputID)

	// Step 2: update the Kafka output without config_yaml.
	// The Fleet API will echo back the previously stored config_yaml value
	// in its response, causing the provider to write a non-null value into
	// state while the plan expected null.  This triggers:
	//   "Provider produced inconsistent result after apply:
	//    .config_yaml: inconsistent values for sensitive attribute"
	configWithoutYaml := fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_fleet_output" "repro_1856" {
  output_id            = %q
  name                 = "Issue 1856 Repro (updated)"
  type                 = "kafka"
  default_integrations = false
  default_monitoring   = false
  hosts                = ["kafka:9092"]

  kafka = {
    auth_type       = "none"
    connection_type = "plaintext"
    topic           = "test-topic-updated"
    partition       = "round_robin"
  }
}
`, outputID)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceOutputDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				Config:                   configWithYaml,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_output.repro_1856", "type", "kafka"),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.repro_1856", "output_id", outputID),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				Config:                   configWithoutYaml,
				// Bug: the provider writes the API-echoed config_yaml back into
				// state after the update, producing a mismatch with the plan's
				// null value for the sensitive config_yaml attribute.
				// The Terraform CLI word-wraps at ~80 chars so we match a shorter
				// substring that is unambiguous and fits on a single line.
				ExpectError: regexp.MustCompile(`inconsistent values for sensitive`),
			},
		},
	})
}
