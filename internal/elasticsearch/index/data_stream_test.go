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

package index_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccResourceDataStream(t *testing.T) {
	// generate random name
	dsName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlpha)
	dsNameUpdated := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlpha)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             checkResourceDataStreamDestroy,
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceDataStreamCreate(dsName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_data_stream.test_ds", "id"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_data_stream.test_ds", "name", dsName),
					resource.TestCheckNoResourceAttr("elasticstack_elasticsearch_data_stream.test_ds", "metadata"),
					// check some computed fields
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_data_stream.test_ds", "indices.#", "1"),
					resource.TestMatchResourceAttr("elasticstack_elasticsearch_data_stream.test_ds", "indices.0.index_name", dataStreamBackingIndexNameRegexp(dsName)),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_data_stream.test_ds", "indices.0.index_uuid"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_data_stream.test_ds", "template", dsName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_data_stream.test_ds", "ilm_policy", dsName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_data_stream.test_ds", "timestamp_field", "@timestamp"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_data_stream.test_ds", "generation", "1"),
					resource.TestMatchResourceAttr("elasticstack_elasticsearch_data_stream.test_ds", "status", regexp.MustCompile(`^(?i)(green|yellow|red)$`)),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_data_stream.test_ds", "hidden", "false"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_data_stream.test_ds", "system", "false"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_data_stream.test_ds", "replicated", "false"),
				),
			},
			{
				Config:            testAccResourceDataStreamCreate(dsName),
				ResourceName:      "elasticstack_elasticsearch_data_stream.test_ds",
				ImportState:       true,
				ImportStateVerify: true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_data_stream.test_ds", "name", dsName),
					resource.TestCheckNoResourceAttr("elasticstack_elasticsearch_data_stream.test_ds", "metadata"),
					resource.TestMatchResourceAttr("elasticstack_elasticsearch_data_stream.test_ds", "indices.0.index_name", dataStreamBackingIndexNameRegexp(dsName)),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_data_stream.test_ds", "indices.0.index_uuid"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_data_stream.test_ds", "template", dsName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_data_stream.test_ds", "ilm_policy", dsName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_data_stream.test_ds", "timestamp_field", "@timestamp"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_data_stream.test_ds", "generation", "1"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_data_stream.test_ds", "hidden", "false"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_data_stream.test_ds", "system", "false"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_data_stream.test_ds", "replicated", "false"),
				),
			},
			{
				Config: testAccResourceDataStreamCreate(dsNameUpdated),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_data_stream.test_ds", "name", dsNameUpdated),
					resource.TestMatchResourceAttr("elasticstack_elasticsearch_data_stream.test_ds", "id", regexp.MustCompile(fmt.Sprintf(".+/%s$", dsNameUpdated))),
					resource.TestCheckNoResourceAttr("elasticstack_elasticsearch_data_stream.test_ds", "metadata"),
					resource.TestMatchResourceAttr("elasticstack_elasticsearch_data_stream.test_ds", "indices.0.index_name", dataStreamBackingIndexNameRegexp(dsNameUpdated)),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_data_stream.test_ds", "template", dsNameUpdated),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_data_stream.test_ds", "ilm_policy", dsNameUpdated),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_data_stream.test_ds", "timestamp_field", "@timestamp"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_data_stream.test_ds", "generation", "1"),
					resource.TestMatchResourceAttr("elasticstack_elasticsearch_data_stream.test_ds", "status", regexp.MustCompile(`^(?i)(green|yellow|red)$`)),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_data_stream.test_ds", "hidden", "false"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_data_stream.test_ds", "system", "false"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_data_stream.test_ds", "replicated", "false"),
				),
			},
		},
	})
}

func TestAccResourceDataStreamWithMetadata(t *testing.T) {
	dsName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlpha)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             checkResourceDataStreamDestroy,
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceDataStreamWithMetadata(dsName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_data_stream.test_ds", "name", dsName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_data_stream.test_ds", "metadata", `{"env":"test","version":1}`),
				),
			},
			{
				Config:            testAccResourceDataStreamWithMetadata(dsName),
				ResourceName:      "elasticstack_elasticsearch_data_stream.test_ds",
				ImportState:       true,
				ImportStateVerify: true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_data_stream.test_ds", "name", dsName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_data_stream.test_ds", "metadata", `{"env":"test","version":1}`),
				),
			},
		},
	})
}

func TestAccResourceDataStreamNameValidation(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config:      testAccResourceDataStreamInvalidName("-leading-hyphen"),
				ExpectError: regexp.MustCompile(`cannot start with -, _, \+`),
			},
			{
				Config:      testAccResourceDataStreamInvalidName("_leading-underscore"),
				ExpectError: regexp.MustCompile(`cannot start with -, _, \+`),
			},
			{
				Config:      testAccResourceDataStreamInvalidName("+leading-plus"),
				ExpectError: regexp.MustCompile(`cannot start with -, _, \+`),
			},
			{
				Config:      testAccResourceDataStreamInvalidName("UpperCaseName"),
				ExpectError: regexp.MustCompile(`must contain lower case alphanumeric characters`),
			},
		},
	})
}

func testAccResourceDataStreamCreate(name string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_index_lifecycle" "test_ilm" {
  name = "%s"

  hot {
    min_age = "1h"
    set_priority {
      priority = 10
    }
    rollover {
      max_age = "1d"
    }
    readonly {}
  }

  delete {
    min_age = "2d"
    delete {}
  }
}

resource "elasticstack_elasticsearch_index_template" "test_ds_template" {
  name = "%s"

  index_patterns = ["%s*"]

  template {
    // make sure our template uses prepared ILM policy
    settings = jsonencode({
      "lifecycle.name" = elasticstack_elasticsearch_index_lifecycle.test_ilm.name
    })

  }

  data_stream {}
}

// and now we can create data stream based on the index template
resource "elasticstack_elasticsearch_data_stream" "test_ds" {
  name = "%s"

  // make sure that template is created before the data stream
  depends_on = [
    elasticstack_elasticsearch_index_template.test_ds_template
  ]
}
	`, name, name, name, name)
}

func testAccResourceDataStreamWithMetadata(name string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_index_lifecycle" "test_ilm" {
  name = "%s"

  hot {
    min_age = "1h"
    set_priority {
      priority = 10
    }
    rollover {
      max_age = "1d"
    }
    readonly {}
  }

  delete {
    min_age = "2d"
    delete {}
  }
}

resource "elasticstack_elasticsearch_index_template" "test_ds_template" {
  name = "%s"

  index_patterns = ["%s*"]

  metadata = jsonencode({
    env     = "test"
    version = 1
  })

  template {
    settings = jsonencode({
      "lifecycle.name" = elasticstack_elasticsearch_index_lifecycle.test_ilm.name
    })
  }

  data_stream {}
}

resource "elasticstack_elasticsearch_data_stream" "test_ds" {
  name = "%s"

  depends_on = [
    elasticstack_elasticsearch_index_template.test_ds_template
  ]
}
	`, name, name, name, name)
}

func testAccResourceDataStreamInvalidName(name string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_data_stream" "test_ds" {
  name = "%s"
}
	`, name)
}

func dataStreamBackingIndexNameRegexp(name string) *regexp.Regexp {
	return regexp.MustCompile(fmt.Sprintf(`^\.ds-%s-.*-000001$`, name))
}

func checkResourceDataStreamDestroy(s *terraform.State) error {
	client, err := clients.NewAcceptanceTestingClient()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "elasticstack_elasticsearch_data_stream" {
			continue
		}
		compID, _ := clients.CompositeIDFromStr(rs.Primary.ID)

		esClient, err := client.GetESClient()
		if err != nil {
			return err
		}
		req := esClient.Indices.GetDataStream.WithName(compID.ResourceID)
		res, err := esClient.Indices.GetDataStream(req)
		if err != nil {
			return err
		}

		if res.StatusCode != 404 {
			return fmt.Errorf("Data Stream (%s) still exists", compID.ResourceID)
		}
	}
	return nil
}
