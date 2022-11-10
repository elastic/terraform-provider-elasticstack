package ingest_test

import (
	"fmt"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccResourceIngestPipeline(t *testing.T) {
	pipelineName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		CheckDestroy:      checkResourceIngestPipelineDestroy,
		ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceIngestPipelineCreate(pipelineName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ingest_pipeline.test_pipeline", "name", pipelineName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ingest_pipeline.test_pipeline", "description", "Test Pipeline"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ingest_pipeline.test_pipeline", "processors.#", "2"),
				),
			},
			{
				Config: testAccResourceIngestPipelineUpdate(pipelineName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ingest_pipeline.test_pipeline", "name", pipelineName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ingest_pipeline.test_pipeline", "description", "Test Pipeline"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ingest_pipeline.test_pipeline", "processors.#", "1"),
				),
			},
		},
	})
}

func testAccResourceIngestPipelineCreate(name string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_ingest_pipeline" "test_pipeline" {
  name        = "%s"
  description = "Test Pipeline"

  processors = [
    jsonencode({
      set = {
        description = "My set processor description"
        field       = "_meta"
        value       = "indexed"
      }
    }),
    <<EOF
    {"json": {
      "field": "data",
      "target_field": "parsed_data"
    }}
EOF
    ,
  ]
}
	`, name)
}

func testAccResourceIngestPipelineUpdate(name string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_ingest_pipeline" "test_pipeline" {
  name        = "%s"
  description = "Test Pipeline"

  processors = [
    jsonencode({
      set = {
        description = "My set processor description"
        field       = "_meta"
        value       = "indexed"
      }
    })
  ]
}
	`, name)
}

func checkResourceIngestPipelineDestroy(s *terraform.State) error {
	client := acctest.Provider.Meta().(*clients.ApiClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "elasticstack_elasticsearch_ingest_pipeline" {
			continue
		}
		compId, _ := clients.CompositeIdFromStr(rs.Primary.ID)

		res, err := client.GetESClient().Indices.Get([]string{compId.ResourceId})
		if err != nil {
			return err
		}

		if res.StatusCode != 404 {
			return fmt.Errorf("Ingest pipeline (%s) still exists", compId.ResourceId)
		}
	}
	return nil
}
