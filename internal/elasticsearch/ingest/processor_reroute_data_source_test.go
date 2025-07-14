package ingest_test

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceIngestProcessorReroute(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceIngestProcessorRerouteDestination,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_reroute.test_destination", "destination", "my-dest-index"),
					CheckResourceJson("data.elasticstack_elasticsearch_ingest_processor_reroute.test_destination", "json", expectedJsonRerouteDestination),
				),
			},
			{
				Config: testAccDataSourceIngestProcessorRerouteDataset,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_reroute.test_dataset", "dataset.0", "my-dataset"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_reroute.test_dataset", "namespace.0", "my-namespace"),
					CheckResourceJson("data.elasticstack_elasticsearch_ingest_processor_reroute.test_dataset", "json", expectedJsonRerouteDataset),
				),
			},
			{
				Config: testAccDataSourceIngestProcessorRerouteDatasetFallback,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_reroute.test_dataset_fallback", "dataset.0", "{{field1}}"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_ingest_processor_reroute.test_dataset_fallback", "dataset.1", "my-fallback-dataset"),
					CheckResourceJson("data.elasticstack_elasticsearch_ingest_processor_reroute.test_dataset_fallback", "json", expectedJsonRerouteDatasetFallback),
				),
			},
		},
	})
}

const expectedJsonRerouteDestination = `{
	"reroute": {
		"destination": "my-dest-index",
		"ignore_failure": false
	}
}`

const testAccDataSourceIngestProcessorRerouteDestination = `
provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_reroute" "test_destination" {
  destination = "my-dest-index"
}
`

const expectedJsonRerouteDataset = `{
	"reroute": {
		"dataset": ["my-dataset"],
		"namespace": ["my-namespace"],
		"ignore_failure": false
	}
}`

const testAccDataSourceIngestProcessorRerouteDataset = `
provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_reroute" "test_dataset" {
  dataset = ["my-dataset"]
  namespace = ["my-namespace"]
}
`

const expectedJsonRerouteDatasetFallback = `{
	"reroute": {
		"dataset": ["{{field1}}", "my-fallback-dataset"],
		"ignore_failure": false
	}
}`

const testAccDataSourceIngestProcessorRerouteDatasetFallback = `
provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_reroute" "test_dataset_fallback" {
  dataset = ["{{field1}}", "my-fallback-dataset"]
}
`
