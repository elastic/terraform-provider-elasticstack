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
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_logstash_pipeline.test", "pipeline_metadata.type", "logstash_pipeline"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_logstash_pipeline.test", "pipeline_metadata.version", "1"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_logstash_pipeline.test", "pipeline_batch_delay", "100"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_logstash_pipeline.test", "pipeline_batch_size", "250"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_logstash_pipeline.test", "pipeline_ecs_compatibility", "disabled"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_logstash_pipeline.test", "pipeline_ordered", "auto"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_logstash_pipeline.test", "pipeline_plugin_classloaders", "false"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_logstash_pipeline.test", "pipeline_unsafe_shutdown", "false"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_logstash_pipeline.test", "pipeline_workers", "2"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_logstash_pipeline.test", "queue_checkpoint_acks", "1024"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_logstash_pipeline.test", "queue_checkpoint_retry", "true"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_logstash_pipeline.test", "queue_checkpoint_writes", "2048"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_logstash_pipeline.test", "queue_drain", "false"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_logstash_pipeline.test", "queue_max_bytes_number", "2"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_logstash_pipeline.test", "queue_max_bytes_units", "mb"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_logstash_pipeline.test", "queue_max_events", "0"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_logstash_pipeline.test", "queue_page_capacity", "64mb"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_logstash_pipeline.test", "queue_type", "memory"),
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

  pipeline_metadata = {
    "type" = "logstash_pipeline"
    "version" = 1
  }

  pipeline_batch_delay = 100
  pipeline_batch_size = 250
  pipeline_ecs_compatibility = "disabled"
  pipeline_ordered = "auto"
  pipeline_plugin_classloaders = false
  pipeline_unsafe_shutdown = false
  pipeline_workers = 2
  queue_checkpoint_acks = 1024
  queue_checkpoint_retry = true
  queue_checkpoint_writes = 2048
  queue_drain = false
  queue_max_bytes_number = 2
  queue_max_bytes_units = "mb"
  queue_max_events = 0
  queue_page_capacity = "64mb"
  queue_type = "memory"
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
