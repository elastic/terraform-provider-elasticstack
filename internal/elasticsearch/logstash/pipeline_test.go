package logstash_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestResourceLogstashPipeline(t *testing.T) {
	pipelineID := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		CheckDestroy:      checkResourceLogstashPipelineDestroy,
		ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceLogstashPipelineCreate(pipelineID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_logstash_pipeline.test", "pipeline_id", pipelineID),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_logstash_pipeline.test", "description", "Description of Logstash Pipeline"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_logstash_pipeline.test", "pipeline", "input{} filter{} output{}"),
				),
			},
			{
				Config: testAccResourceLogstashPipelineUpdate(pipelineID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_logstash_pipeline.test", "pipeline_id", pipelineID),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_logstash_pipeline.test", "description", "Updated description of Logstash Pipeline"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_logstash_pipeline.test", "pipeline", "input{} \nfilter{} \noutput{}"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_logstash_pipeline.test", "pipeline_metadata", `{"type":"logstash_pipeline","version":"1"}`),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_logstash_pipeline.test", "pipeline_settings", `{"pipeline.workers":2,"pipeline.batch.size":150,"pipeline.batch.delay":60,"queue.type":"persisted","queue.max_bytes.number":2,"queue.max_bytes.units":"mb","queue.checkpoint.writes":2048}`),
				),
			},
		},
	})
}

func testAccResourceLogstashPipelineCreate(pipelineID string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_logstash_pipeline" "test" {
  pipeline_id = "%s"
	description = "Description of Logstash Pipeline"
	pipeline = "input{} filter{} output{}"
}
	`, pipelineID)
}

func testAccResourceLogstashPipelineUpdate(pipelineID string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_logstash_pipeline" "test" {
  pipeline_id = "%s"
	description = "Updated description of Logstash Pipeline"
	pipeline = "input{} \nfilter{} \noutput{}"
	pipeline_metadata = jsonencode({
		"type" = "logstash_pipeline",
    "version" = "1"
	})
	pipeline_settings = jsonencode({
		"pipeline.workers": 2,
    "pipeline.batch.size": 150,
    "pipeline.batch.delay": 60,
    "queue.type": "persisted",
    "queue.max_bytes.number": 2,
    "queue.max_bytes.units": "mb",
    "queue.checkpoint.writes": 2048
	})
}
	`, pipelineID)
}

func checkResourceLogstashPipelineDestroy(s *terraform.State) error {

	client := acctest.Provider.Meta().(*clients.ApiClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "elasticstack_elasticsearch_logstash_pipeline" {
			continue
		}
		compId, _ := clients.CompositeIdFromStr(rs.Primary.ID)

		res, err := client.GetESClient().LogstashGetPipeline(compId.ResourceId)
		if err != nil {
			return err
		}

		if res.StatusCode != http.StatusNotFound {
			return fmt.Errorf("logstash pipeline (%s) still exists", compId.ResourceId)
		}
	}
	return nil
}
