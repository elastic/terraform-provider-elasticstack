package ingest_test

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceIngestProcessorPipeline(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceIngestProcessorPipeline,
				Check: resource.ComposeTestCheckFunc(
					CheckResourceJson("data.elasticstack_elasticsearch_ingest_processor_pipeline.test", "json", expectedJsonPipeline),
				),
			},
		},
	})
}

const expectedJsonPipeline = `{
	"pipeline": {
		"name": "pipeline_a",
		"ignore_failure": false
	}
}`

const testAccDataSourceIngestProcessorPipeline = `
provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_append" "tags" {
  field = "tags"
  value = ["production", "{{{app}}}", "{{{owner}}}"]
}

resource "elasticstack_elasticsearch_ingest_pipeline" "pipeline_a" {
  name = "pipeline_a"

  processors = [
    data.elasticstack_elasticsearch_ingest_processor_append.tags.json
  ]
}

data "elasticstack_elasticsearch_ingest_processor_pipeline" "test" {
  name = elasticstack_elasticsearch_ingest_pipeline.pipeline_a.name
}
`
