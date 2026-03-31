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

package datastreamlifecycle_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/index/datastreamlifecycle"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccResourceDataStreamLifecycle(t *testing.T) {
	dsName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlpha)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             checkResourceDataStreamLifecycleDestroy,
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				SkipFunc:        versionutils.CheckIfVersionIsUnsupported(datastreamlifecycle.MinVersion),
				ConfigDirectory: acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{"name": config.StringVariable(dsName)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr("elasticstack_elasticsearch_data_stream_lifecycle.test_ds_lifecycle", "id", dataStreamLifecycleIDRegexp(dsName+"-one")),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_data_stream_lifecycle.test_ds_lifecycle", "name", dsName+"-one"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_data_stream_lifecycle.test_ds_lifecycle", "data_retention", "3d"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_data_stream_lifecycle.test_ds_lifecycle", "enabled", "true"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_data_stream_lifecycle.test_ds_lifecycle", "expand_wildcards", "open"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_data_stream_lifecycle.test_ds_lifecycle", "downsampling.#", "2"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_data_stream_lifecycle.test_ds_lifecycle", "downsampling.0.after", "1d"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_data_stream_lifecycle.test_ds_lifecycle", "downsampling.0.fixed_interval", "10m"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_data_stream_lifecycle.test_ds_lifecycle", "downsampling.1.after", "7d"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_data_stream_lifecycle.test_ds_lifecycle", "downsampling.1.fixed_interval", "1d"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_data_stream_lifecycle.test_ds_lifecycle_multiple", "name", dsName+"-multiple-*"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_data_stream_lifecycle.test_ds_lifecycle_multiple", "data_retention", "3d"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_data_stream_lifecycle.test_ds_lifecycle_multiple", "enabled", "true"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_data_stream_lifecycle.test_ds_lifecycle_multiple", "expand_wildcards", "open"),
				),
			},
			{
				SkipFunc:                versionutils.CheckIfVersionIsUnsupported(datastreamlifecycle.MinVersion),
				ResourceName:            "elasticstack_elasticsearch_data_stream_lifecycle.test_ds_lifecycle",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"name", "enabled", "expand_wildcards"},
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"elasticstack_elasticsearch_data_stream_lifecycle.test_ds_lifecycle",
						"id",
						dataStreamLifecycleIDRegexp(dsName+"-one"),
					),
				),
			},
			{
				SkipFunc:        versionutils.CheckIfVersionIsUnsupported(datastreamlifecycle.MinVersion),
				ConfigDirectory: acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{"name": config.StringVariable(dsName)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr("elasticstack_elasticsearch_data_stream_lifecycle.test_ds_lifecycle", "id", dataStreamLifecycleIDRegexp(dsName+"-one")),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_data_stream_lifecycle.test_ds_lifecycle", "name", dsName+"-one"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_data_stream_lifecycle.test_ds_lifecycle", "data_retention", "2d"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_data_stream_lifecycle.test_ds_lifecycle", "enabled", "false"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_data_stream_lifecycle.test_ds_lifecycle", "expand_wildcards", "all"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_data_stream_lifecycle.test_ds_lifecycle", "downsampling.#", "2"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_data_stream_lifecycle.test_ds_lifecycle", "downsampling.0.after", "2d"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_data_stream_lifecycle.test_ds_lifecycle", "downsampling.0.fixed_interval", "30m"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_data_stream_lifecycle.test_ds_lifecycle", "downsampling.1.after", "9d"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_data_stream_lifecycle.test_ds_lifecycle", "downsampling.1.fixed_interval", "2d"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_data_stream_lifecycle.test_ds_lifecycle_multiple", "name", dsName+"-multiple-*"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_data_stream_lifecycle.test_ds_lifecycle_multiple", "data_retention", "2d"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_data_stream_lifecycle.test_ds_lifecycle_multiple", "enabled", "true"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_data_stream_lifecycle.test_ds_lifecycle_multiple", "expand_wildcards", "open"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_data_stream_lifecycle.test_ds_lifecycle_multiple", "downsampling.0.after", "1d"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_data_stream_lifecycle.test_ds_lifecycle_multiple", "downsampling.0.fixed_interval", "10m"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_data_stream_lifecycle.test_ds_lifecycle_multiple", "downsampling.1.after", "7d"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_data_stream_lifecycle.test_ds_lifecycle_multiple", "downsampling.1.fixed_interval", "1d"),
				),
			},
			{
				SkipFunc:        versionutils.CheckIfVersionIsUnsupported(datastreamlifecycle.MinVersion),
				ConfigDirectory: acctest.NamedTestCaseDirectory("reenable"),
				ConfigVariables: config.Variables{"name": config.StringVariable(dsName)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr("elasticstack_elasticsearch_data_stream_lifecycle.test_ds_lifecycle", "id", dataStreamLifecycleIDRegexp(dsName+"-one")),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_data_stream_lifecycle.test_ds_lifecycle", "name", dsName+"-one"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_data_stream_lifecycle.test_ds_lifecycle", "data_retention", "4d"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_data_stream_lifecycle.test_ds_lifecycle", "enabled", "true"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_data_stream_lifecycle.test_ds_lifecycle", "expand_wildcards", "hidden"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_data_stream_lifecycle.test_ds_lifecycle", "downsampling.#", "2"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_data_stream_lifecycle.test_ds_lifecycle", "downsampling.0.after", "4d"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_data_stream_lifecycle.test_ds_lifecycle", "downsampling.0.fixed_interval", "45m"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_data_stream_lifecycle.test_ds_lifecycle", "downsampling.1.after", "11d"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_data_stream_lifecycle.test_ds_lifecycle", "downsampling.1.fixed_interval", "3d"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_data_stream_lifecycle.test_ds_lifecycle_multiple", "name", dsName+"-multiple-*"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_data_stream_lifecycle.test_ds_lifecycle_multiple", "data_retention", "2d"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_data_stream_lifecycle.test_ds_lifecycle_multiple", "enabled", "true"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_data_stream_lifecycle.test_ds_lifecycle_multiple", "expand_wildcards", "open"),
				),
			},
			{
				SkipFunc:        versionutils.CheckIfVersionIsUnsupported(datastreamlifecycle.MinVersion),
				ConfigDirectory: acctest.NamedTestCaseDirectory("single_downsampling"),
				ConfigVariables: config.Variables{"name": config.StringVariable(dsName)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr("elasticstack_elasticsearch_data_stream_lifecycle.test_ds_lifecycle", "id", dataStreamLifecycleIDRegexp(dsName+"-one")),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_data_stream_lifecycle.test_ds_lifecycle", "name", dsName+"-one"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_data_stream_lifecycle.test_ds_lifecycle", "data_retention", "5d"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_data_stream_lifecycle.test_ds_lifecycle", "enabled", "true"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_data_stream_lifecycle.test_ds_lifecycle", "expand_wildcards", "closed"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_data_stream_lifecycle.test_ds_lifecycle", "downsampling.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_data_stream_lifecycle.test_ds_lifecycle", "downsampling.0.after", "5d"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_data_stream_lifecycle.test_ds_lifecycle", "downsampling.0.fixed_interval", "1h"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_data_stream_lifecycle.test_ds_lifecycle_multiple", "name", dsName+"-multiple-*"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_data_stream_lifecycle.test_ds_lifecycle_multiple", "data_retention", "2d"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_data_stream_lifecycle.test_ds_lifecycle_multiple", "enabled", "true"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_data_stream_lifecycle.test_ds_lifecycle_multiple", "expand_wildcards", "open"),
				),
			},
			{
				SkipFunc:        versionutils.CheckIfVersionIsUnsupported(datastreamlifecycle.MinVersion),
				ConfigDirectory: acctest.NamedTestCaseDirectory("remove_retention"),
				ConfigVariables: config.Variables{"name": config.StringVariable(dsName)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr("elasticstack_elasticsearch_data_stream_lifecycle.test_ds_lifecycle", "id", dataStreamLifecycleIDRegexp(dsName+"-one")),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_data_stream_lifecycle.test_ds_lifecycle", "name", dsName+"-one"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_data_stream_lifecycle.test_ds_lifecycle", "enabled", "false"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_data_stream_lifecycle.test_ds_lifecycle", "expand_wildcards", "all"),
					resource.TestCheckNoResourceAttr("elasticstack_elasticsearch_data_stream_lifecycle.test_ds_lifecycle", "data_retention"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_data_stream_lifecycle.test_ds_lifecycle", "downsampling.#", "0"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_data_stream_lifecycle.test_ds_lifecycle_multiple", "name", dsName+"-multiple-*"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_data_stream_lifecycle.test_ds_lifecycle_multiple", "data_retention", "2d"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_data_stream_lifecycle.test_ds_lifecycle_multiple", "enabled", "true"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_data_stream_lifecycle.test_ds_lifecycle_multiple", "expand_wildcards", "open"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_data_stream_lifecycle.test_ds_lifecycle_multiple", "downsampling.0.after", "1d"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_data_stream_lifecycle.test_ds_lifecycle_multiple", "downsampling.0.fixed_interval", "10m"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_data_stream_lifecycle.test_ds_lifecycle_multiple", "downsampling.1.after", "7d"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_data_stream_lifecycle.test_ds_lifecycle_multiple", "downsampling.1.fixed_interval", "1d"),
				),
			},
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(datastreamlifecycle.MinVersion),
				PreConfig: func() {
					client, err := clients.NewAcceptanceTestingClient()
					if err != nil {
						t.Fatalf("Failed to create testing client: %s", err)
					}

					esClient, err := client.GetESClient()
					if err != nil {
						t.Fatalf("Failed to get es client: %s", err)
					}

					lifecycle := models.LifecycleSettings{
						DataRetention: "10d",
						Downsampling: []models.Downsampling{
							{After: "10d", FixedInterval: "5d"},
							{After: "20d", FixedInterval: "10d"},
						},
					}
					lifecycleBytes, err := json.Marshal(lifecycle)
					if err != nil {
						t.Fatalf("Cannot marshal lifecycle: %s", err)
					}
					_, err = esClient.Indices.PutDataLifecycle(
						[]string{dsName + "-multiple-two"},
						esClient.Indices.PutDataLifecycle.WithBody(bytes.NewReader(lifecycleBytes)),
					)
					if err != nil {
						t.Fatalf("Cannot update lifecycle: %s", err)
					}
				},
				ConfigDirectory: acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{"name": config.StringVariable(dsName)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr("elasticstack_elasticsearch_data_stream_lifecycle.test_ds_lifecycle", "id", dataStreamLifecycleIDRegexp(dsName+"-one")),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_data_stream_lifecycle.test_ds_lifecycle", "name", dsName+"-one"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_data_stream_lifecycle.test_ds_lifecycle", "data_retention", "2d"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_data_stream_lifecycle.test_ds_lifecycle", "enabled", "false"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_data_stream_lifecycle.test_ds_lifecycle", "expand_wildcards", "all"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_data_stream_lifecycle.test_ds_lifecycle", "downsampling.0.after", "2d"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_data_stream_lifecycle.test_ds_lifecycle", "downsampling.0.fixed_interval", "30m"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_data_stream_lifecycle.test_ds_lifecycle", "downsampling.1.after", "9d"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_data_stream_lifecycle.test_ds_lifecycle", "downsampling.1.fixed_interval", "2d"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_data_stream_lifecycle.test_ds_lifecycle_multiple", "name", dsName+"-multiple-*"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_data_stream_lifecycle.test_ds_lifecycle_multiple", "data_retention", "2d"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_data_stream_lifecycle.test_ds_lifecycle_multiple", "enabled", "true"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_data_stream_lifecycle.test_ds_lifecycle_multiple", "expand_wildcards", "open"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_data_stream_lifecycle.test_ds_lifecycle_multiple", "downsampling.0.after", "1d"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_data_stream_lifecycle.test_ds_lifecycle_multiple", "downsampling.0.fixed_interval", "10m"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_data_stream_lifecycle.test_ds_lifecycle_multiple", "downsampling.1.after", "7d"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_data_stream_lifecycle.test_ds_lifecycle_multiple", "downsampling.1.fixed_interval", "1d"),
				),
			},
		},
	})
}

func TestAccResourceDataStreamLifecycleInvalidExpandWildcards(t *testing.T) {
	dsName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlpha)

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config:      testAccResourceDataStreamLifecycleInvalidExpandWildcards(dsName, "invalid"),
				ExpectError: regexp.MustCompile(`value must be one of`),
			},
		},
	})
}

func dataStreamLifecycleIDRegexp(name string) *regexp.Regexp {
	return regexp.MustCompile(fmt.Sprintf(`.+/%s$`, regexp.QuoteMeta(name)))
}

func testAccResourceDataStreamLifecycleInvalidExpandWildcards(name, expandWildcards string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_index_template" "test_ds_template" {
  name = "%[1]s"

  index_patterns = ["%[1]s*"]

  data_stream {}
}

resource "elasticstack_elasticsearch_data_stream" "test_ds_one" {
  name = "%[1]s-one"

  depends_on = [
    elasticstack_elasticsearch_index_template.test_ds_template
  ]
}

resource "elasticstack_elasticsearch_data_stream_lifecycle" "test_ds_lifecycle" {
	name = "%[1]s-one"
	expand_wildcards = "%[2]s"

	depends_on = [
		elasticstack_elasticsearch_data_stream.test_ds_one
	]
}
`, name, expandWildcards)
}

func checkResourceDataStreamLifecycleDestroy(s *terraform.State) error {
	client, err := clients.NewAcceptanceTestingClient()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "elasticstack_elasticsearch_data_stream_lifecycle" {
			continue
		}
		compID, _ := clients.CompositeIDFromStr(rs.Primary.ID)

		esClient, err := client.GetESClient()
		if err != nil {
			return err
		}

		res, err := esClient.Indices.GetDataLifecycle([]string{compID.ResourceID})
		if err != nil {
			return err
		}

		// for lifecycle without wildcard 404 is returned when no ds matches
		if res.StatusCode == 404 {
			return nil
		}

		defer res.Body.Close()

		dStreams := struct {
			DataStreams []models.DataStreamLifecycle `json:"data_streams,omitempty"`
		}{}

		if err := json.NewDecoder(res.Body).Decode(&dStreams); err != nil {
			return err
		}

		// for lifecycle with wildcard empty array is returned
		if len(dStreams.DataStreams) > 0 {
			return fmt.Errorf("Data Stream Lifecycle (%s) still exists", compID.ResourceID)
		}
	}
	return nil
}
