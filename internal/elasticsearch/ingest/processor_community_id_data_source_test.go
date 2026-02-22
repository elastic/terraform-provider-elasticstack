package ingest_test

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDataSourceIngestProcessorCommunityID(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceIngestProcessorCommunityID,
				Check: resource.ComposeTestCheckFunc(
					CheckResourceJSON("data.elasticstack_elasticsearch_ingest_processor_community_id.test", "json", expectedJSONCommunityID),
				),
			},
		},
	})
}

const expectedJSONCommunityID = `{
	"community_id": {
		"seed": 0,
		"ignore_failure": false,
		"ignore_missing": false
	}
}`

const testAccDataSourceIngestProcessorCommunityID = `
provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_community_id" "test" {}
`
