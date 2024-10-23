package data_stream_lifecycle_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/index/data_stream_lifecycle"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccResourceDataStreamLifecycle(t *testing.T) {
	dsName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlpha)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             checkResourceDataStreamLifecycleDestroy,
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(data_stream_lifecycle.MinVersion),
				Config:   testAccResourceDataStreamLifecycleCreate(dsName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_data_stream_lifecycle.test_ds_lifecycle", "name", dsName+"-one"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_data_stream_lifecycle.test_ds_lifecycle", "data_retention", "3d"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_data_stream_lifecycle.test_ds_lifecycle", "downsampling.#", "2"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_data_stream_lifecycle.test_ds_lifecycle", "downsampling.0.after", "1d"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_data_stream_lifecycle.test_ds_lifecycle", "downsampling.0.fixed_interval", "10m"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_data_stream_lifecycle.test_ds_lifecycle", "downsampling.1.after", "7d"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_data_stream_lifecycle.test_ds_lifecycle", "downsampling.1.fixed_interval", "1d"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_data_stream_lifecycle.test_ds_lifecycle_multiple", "name", dsName+"-multiple-*"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_data_stream_lifecycle.test_ds_lifecycle_multiple", "data_retention", "3d"),
				),
			},
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(data_stream_lifecycle.MinVersion),
				Config:   testAccResourceDataStreamLifecycleUpdate(dsName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_data_stream_lifecycle.test_ds_lifecycle", "name", dsName+"-one"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_data_stream_lifecycle.test_ds_lifecycle", "data_retention", "2d"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_data_stream_lifecycle.test_ds_lifecycle", "downsampling.#", "0"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_data_stream_lifecycle.test_ds_lifecycle_multiple", "name", dsName+"-multiple-*"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_data_stream_lifecycle.test_ds_lifecycle_multiple", "data_retention", "2d"),
				),
			},
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(data_stream_lifecycle.MinVersion),
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
					_, err = esClient.Indices.PutDataLifecycle([]string{dsName + "-multiple-two"}, esClient.Indices.PutDataLifecycle.WithBody(bytes.NewReader(lifecycleBytes)))
					if err != nil {
						t.Fatalf("Cannot update lifecycle: %s", err)
					}
				},
				Config: testAccResourceDataStreamLifecycleUpdate(dsName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_data_stream_lifecycle.test_ds_lifecycle", "name", dsName+"-one"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_data_stream_lifecycle.test_ds_lifecycle", "data_retention", "2d"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_data_stream_lifecycle.test_ds_lifecycle_multiple", "name", dsName+"-multiple-*"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_data_stream_lifecycle.test_ds_lifecycle_multiple", "data_retention", "2d"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_data_stream_lifecycle.test_ds_lifecycle_multiple", "downsampling.0.after", "1d"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_data_stream_lifecycle.test_ds_lifecycle_multiple", "downsampling.0.fixed_interval", "10m"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_data_stream_lifecycle.test_ds_lifecycle_multiple", "downsampling.1.after", "7d"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_data_stream_lifecycle.test_ds_lifecycle_multiple", "downsampling.1.fixed_interval", "1d"),
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
  name = "%[1]s-multiple-one"

  depends_on = [
    elasticstack_elasticsearch_index_template.test_ds_template
  ]
}

resource "elasticstack_elasticsearch_data_stream" "test_ds_three" {
  name = "%[1]s-multiple-two"

  depends_on = [
    elasticstack_elasticsearch_index_template.test_ds_template
  ]
}
	
resource "elasticstack_elasticsearch_data_stream_lifecycle" "test_ds_lifecycle" {
	name = "%[1]s-one"
	data_retention = "3d"
	downsampling = [
	{
		after = "1d"
		fixed_interval = "10m"
	},
	{
		after = "7d"
		fixed_interval = "1d"
	}
	]

	depends_on = [
		elasticstack_elasticsearch_data_stream.test_ds_one
	]
}

resource "elasticstack_elasticsearch_data_stream_lifecycle" "test_ds_lifecycle_multiple" {
	name = "%[1]s-multiple-*"
	data_retention = "3d"
	downsampling = [
	{
		after = "1d"
		fixed_interval = "10m"
	},
	{
		after = "7d"
		fixed_interval = "1d"
	}
	]

	depends_on = [
		elasticstack_elasticsearch_data_stream.test_ds_two,
		elasticstack_elasticsearch_data_stream.test_ds_three
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
  name = "%[1]s-multiple-one"

  depends_on = [
    elasticstack_elasticsearch_index_template.test_ds_template
  ]
}

resource "elasticstack_elasticsearch_data_stream" "test_ds_three" {
  name = "%[1]s-multiple-two"

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
	name = "%[1]s-multiple-*"
	data_retention = "2d"
	downsampling = [
	{
		after = "1d"
		fixed_interval = "10m"
	},
	{
		after = "7d"
		fixed_interval = "1d"
	}
	]

	depends_on = [
		elasticstack_elasticsearch_data_stream.test_ds_two,
		elasticstack_elasticsearch_data_stream.test_ds_three
	]
}

`, name)

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
