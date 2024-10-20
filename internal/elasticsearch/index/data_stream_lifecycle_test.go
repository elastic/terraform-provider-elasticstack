package index_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccResourceDataStreamLifecycle(t *testing.T) {
	// generate renadom name
	dsName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlpha)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             checkResourceDataStreamLifecycleDestroy,
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceDataStreamLifecycleCreate(dsName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_data_stream_lifecycle.test_ds_lifecycle", "name", dsName+"-one"),

					resource.TestCheckTypeSetElemNestedAttrs("elasticstack_elasticsearch_data_stream_lifecycle.test_ds_lifecycle", "lifecycles.*", map[string]string{
						"name":           dsName + "-one",
						"data_retention": "3d",
						"enabled":        "true",
					}),

					resource.TestCheckTypeSetElemNestedAttrs("elasticstack_elasticsearch_data_stream_lifecycle.test_ds_lifecycle", "lifecycles.0.downsampling.*", map[string]string{
						"after":          "1d",
						"fixed_interval": "10m",
					}),

					resource.TestCheckTypeSetElemNestedAttrs("elasticstack_elasticsearch_data_stream_lifecycle.test_ds_lifecycle", "lifecycles.0.downsampling.*", map[string]string{
						"after":          "7d",
						"fixed_interval": "1d",
					}),

					// multiple match
					resource.TestCheckTypeSetElemNestedAttrs("elasticstack_elasticsearch_data_stream_lifecycle.test_ds_lifecycle_multiple", "lifecycles.*", map[string]string{
						"name":           dsName + "-one",
						"data_retention": "3d",
						"enabled":        "true",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("elasticstack_elasticsearch_data_stream_lifecycle.test_ds_lifecycle_multiple", "lifecycles.*", map[string]string{
						"name":           dsName + "-two",
						"data_retention": "3d",
						"enabled":        "true",
					}),
				),
			},
			{
				Config: testAccResourceDataStreamLifecycleUpdate(dsName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_data_stream_lifecycle.test_ds_lifecycle", "name", dsName+"-one"),

					resource.TestCheckTypeSetElemNestedAttrs("elasticstack_elasticsearch_data_stream_lifecycle.test_ds_lifecycle", "lifecycles.*", map[string]string{
						"name":           dsName + "-one",
						"data_retention": "2d",
						"enabled":        "true",
					}),

					resource.TestCheckResourceAttr("elasticstack_elasticsearch_data_stream_lifecycle.test_ds_lifecycle", "lifecycles.0.downsampling.#", "0"),

					// multiple match
					resource.TestCheckTypeSetElemNestedAttrs("elasticstack_elasticsearch_data_stream_lifecycle.test_ds_lifecycle_multiple", "lifecycles.*", map[string]string{
						"name":           dsName + "-one",
						"data_retention": "2d",
						"enabled":        "true",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("elasticstack_elasticsearch_data_stream_lifecycle.test_ds_lifecycle_multiple", "lifecycles.*", map[string]string{
						"name":           dsName + "-two",
						"data_retention": "2d",
						"enabled":        "true",
					}),
				),
			},
		},
	})
}

func testAccResourceDataStreamLifecycleCreate(name string) string {
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

resource "elasticstack_elasticsearch_data_stream" "test_ds_two" {
  name = "%[1]s-two"

  depends_on = [
    elasticstack_elasticsearch_index_template.test_ds_template
  ]
}
	
resource "elasticstack_elasticsearch_data_stream_lifecycle" "test_ds_lifecycle" {
	name = "%[1]s-one"
	data_retention = "3d"
	downsampling {
		after = "1d"
		fixed_interval = "10m"
	}
	downsampling {
		after = "7d"
		fixed_interval = "1d"
	}

	depends_on = [
		elasticstack_elasticsearch_data_stream.test_ds_one
	]
}

resource "elasticstack_elasticsearch_data_stream_lifecycle" "test_ds_lifecycle_multiple" {
	name = "%[1]s-*"
	data_retention = "3d"

	depends_on = [
		elasticstack_elasticsearch_data_stream.test_ds_one,
		elasticstack_elasticsearch_data_stream.test_ds_two
	]
}
`, name)

}

func testAccResourceDataStreamLifecycleUpdate(name string) string {
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

resource "elasticstack_elasticsearch_data_stream" "test_ds_two" {
  name = "%[1]s-two"

  depends_on = [
    elasticstack_elasticsearch_index_template.test_ds_template
  ]
}
	
resource "elasticstack_elasticsearch_data_stream_lifecycle" "test_ds_lifecycle" {
	name = "%[1]s-one"
	data_retention = "2d"

	depends_on = [
		elasticstack_elasticsearch_data_stream.test_ds_one
	]
}

resource "elasticstack_elasticsearch_data_stream_lifecycle" "test_ds_lifecycle_multiple" {
	name = "%[1]s-*"
	data_retention = "2d"

	depends_on = [
		elasticstack_elasticsearch_data_stream.test_ds_one,
		elasticstack_elasticsearch_data_stream.test_ds_two
	]
}	`, name)

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
		compId, _ := clients.CompositeIdFromStr(rs.Primary.ID)

		esClient, err := client.GetESClient()
		if err != nil {
			return err
		}

		res, err := esClient.Indices.GetDataLifecycle([]string{compId.ResourceId})
		if err != nil {
			return err
		}

		// for lifecycle without wildcard 404 is returned when no ds matches
		if res.StatusCode == 404 {
			return nil
		}

		defer res.Body.Close()

		dStreams := make(map[string][]models.DataStreamLifecycle)
		if err := json.NewDecoder(res.Body).Decode(&dStreams); err != nil {
			return err
		}
		// for lifecycle with wildcard empty array is returned
		if len(dStreams["data_streams"]) > 0 {
			return fmt.Errorf("Data Stream Lifecycle (%s) still exists", compId.ResourceId)
		}
	}
	return nil
}
