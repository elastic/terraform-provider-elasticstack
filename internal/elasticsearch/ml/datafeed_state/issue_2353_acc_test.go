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

package datafeedstate_test

// TestAccReproduceIssue2353 reproduces the "Provider produced inconsistent
// result after apply" error when an explicit start timestamp is provided to
// elasticstack_elasticsearch_ml_datafeed_state.
//
// When the datafeed runs it sets SearchInterval.StartMs to the timestamp of
// the first data record it finds (which is later than the requested start).
// The provider reads this back and overwrites d.Start with the new value,
// producing a mismatch against the planned start value.
//
// Related: https://github.com/elastic/terraform-provider-elasticstack/issues/2353

import (
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// minVersionIssue2353 gates this reproducer to ES >= 8.1.0. On 8.0.x the
// datafeed running_state / search_interval shape used by the reproducer is
// not reliably available and the test fails for unrelated reasons.
var minVersionIssue2353 = version.Must(version.NewVersion("8.1.0"))

func TestAccReproduceIssue2353(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minVersionIssue2353, versionutils.FlavorAny)

	jobID := fmt.Sprintf("test-job-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum))
	datafeedID := fmt.Sprintf("datafeed-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum))
	indexName := fmt.Sprintf("test-index-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum))

	// docTimestamp is AFTER start but not on any 15m bucket boundary.
	// The datafeed will find this record and set SearchInterval.StartMs to this
	// time, which differs from the planned start ("2022-01-01T00:07:30Z"),
	// triggering the "Provider produced inconsistent result after apply" error.
	const docTimestamp = "2022-01-01T00:10:00Z"
	const plannedStart = "2022-01-01T00:07:30Z"

	configVars := config.Variables{
		"job_id":      config.StringVariable(jobID),
		"datafeed_id": config.StringVariable(datafeedID),
		"index_name":  config.StringVariable(indexName),
	}

	fullConfigVars := config.Variables{
		"job_id":        config.StringVariable(jobID),
		"datafeed_id":   config.StringVariable(datafeedID),
		"index_name":    config.StringVariable(indexName),
		"planned_start": config.StringVariable(plannedStart),
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				// Step 1: create prerequisite resources (index, job, job state,
				// datafeed). After apply, index a document so the datafeed has
				// data to consume in step 2.
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("setup"),
				ConfigVariables:          configVars,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_index.test", "id"),
					testAccIssue2353IndexDocument(indexName, docTimestamp),
				),
			},
			{
				// Step 2: add the datafeed_state resource with an explicit
				// start that precedes the indexed document. The datafeed finds
				// the document and reports SearchInterval.StartMs = docTimestamp,
				// which differs from plannedStart. The framework detects the
				// plan/state mismatch and produces the inconsistency error.
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("full"),
				ConfigVariables:          fullConfigVars,
				ExpectError:              regexp.MustCompile(`Provider produced inconsistent result after apply`),
			},
		},
	})
}

// testAccIssue2353IndexDocument returns a TestCheckFunc that indexes a
// single document into the named index via the Elasticsearch HTTP API.
// This must run after the index is created (step 1 apply) and before the
// datafeed is started (step 2 apply).
func testAccIssue2353IndexDocument(indexName, docTimestamp string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		rawEndpoints := os.Getenv("ELASTICSEARCH_ENDPOINTS")
		if rawEndpoints == "" {
			return fmt.Errorf("ELASTICSEARCH_ENDPOINTS not set")
		}
		endpoint := strings.TrimRight(strings.Split(rawEndpoints, ",")[0], "/")
		url := fmt.Sprintf("%s/%s/_doc", endpoint, indexName)

		body := fmt.Sprintf(`{"@timestamp":%q,"value":42}`, docTimestamp)
		req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(body))
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", "application/json")
		req.SetBasicAuth(
			os.Getenv("ELASTICSEARCH_USERNAME"),
			os.Getenv("ELASTICSEARCH_PASSWORD"),
		)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return fmt.Errorf("index document failed with HTTP %d", resp.StatusCode)
		}
		return nil
	}
}
