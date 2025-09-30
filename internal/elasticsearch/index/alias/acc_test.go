package alias_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccResourceAlias(t *testing.T) {
	// generate random names
	aliasName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlpha)
	indexName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlpha)
	indexName2 := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlpha)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			// Create indices directly via curl to avoid terraform index resource conflicts
			createTestIndex(t, indexName)
			createTestIndex(t, indexName2)
		},
		CheckDestroy:             checkResourceAliasDestroy,
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceAliasCreateDirect(aliasName, indexName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_alias.test_alias", "name", aliasName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_alias.test_alias", "indices.#", "1"),
					resource.TestCheckTypeSetElemAttr("elasticstack_elasticsearch_alias.test_alias", "indices.*", indexName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_alias.test_alias", "is_hidden", "false"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_alias.test_alias", "is_write_index", "false"),
				),
			},
			{
				Config: testAccResourceAliasUpdateDirect(aliasName, indexName, indexName2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_alias.test_alias", "name", aliasName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_alias.test_alias", "indices.#", "2"),
					resource.TestCheckTypeSetElemAttr("elasticstack_elasticsearch_alias.test_alias", "indices.*", indexName),
					resource.TestCheckTypeSetElemAttr("elasticstack_elasticsearch_alias.test_alias", "indices.*", indexName2),
				),
			},
			{
				Config: testAccResourceAliasWithFilterDirect(aliasName, indexName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_alias.test_alias", "name", aliasName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_alias.test_alias", "indices.#", "1"),
					resource.TestCheckTypeSetElemAttr("elasticstack_elasticsearch_alias.test_alias", "indices.*", indexName),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_alias.test_alias", "filter"),
				),
			},
		},
	})
}

func TestAccResourceAliasDataStream(t *testing.T) {
	// generate random names
	aliasName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlpha)
	dsName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlpha)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             checkResourceAliasDestroy,
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceAliasDataStreamCreate(aliasName, dsName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_alias.test_alias", "name", aliasName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_alias.test_alias", "indices.#", "1"),
					resource.TestCheckTypeSetElemAttr("elasticstack_elasticsearch_alias.test_alias", "indices.*", dsName),
				),
			},
		},
	})
}

func createTestIndex(t *testing.T, indexName string) {
	// Create index directly via Elasticsearch API to avoid terraform resource conflicts
	client, err := clients.NewAcceptanceTestingClient()
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	esClient, err := client.GetESClient()
	if err != nil {
		t.Fatalf("Failed to get ES client: %v", err)
	}

	// Create index with basic mapping
	indexBody := `{
		"mappings": {
			"properties": {
				"title": { "type": "text" },
				"status": { "type": "keyword" }
			}
		}
	}`

	_, err = esClient.Indices.Create(indexName, esClient.Indices.Create.WithBody(strings.NewReader(indexBody)))
	if err != nil {
		t.Fatalf("Failed to create index %s: %v", indexName, err)
	}
}

func testAccResourceAliasCreateDirect(aliasName, indexName string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_alias" "test_alias" {
  name    = "%s"
  indices = ["%s"]
}
	`, aliasName, indexName)
}

func testAccResourceAliasUpdateDirect(aliasName, indexName, indexName2 string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_alias" "test_alias" {
  name    = "%s"
  indices = ["%s", "%s"]
}
	`, aliasName, indexName, indexName2)
}

func testAccResourceAliasWithFilterDirect(aliasName, indexName string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_alias" "test_alias" {
  name    = "%s"
  indices = ["%s"]
  filter  = jsonencode({
    term = {
      status = "published"
    }
  })
}
	`, aliasName, indexName)
}

func testAccResourceAliasDataStreamCreate(aliasName, dsName string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_index_template" "test_ds_template" {
  name           = "%s"
  index_patterns = ["%s"]
  data_stream {}
}

resource "elasticstack_elasticsearch_data_stream" "test_ds" {
  name = "%s"
  depends_on = [
    elasticstack_elasticsearch_index_template.test_ds_template
  ]
}

resource "elasticstack_elasticsearch_alias" "test_alias" {
  name    = "%s"
  indices = [elasticstack_elasticsearch_data_stream.test_ds.name]
}
	`, dsName, dsName, dsName, aliasName)
}

func checkResourceAliasDestroy(s *terraform.State) error {
	client, err := clients.NewAcceptanceTestingClient()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "elasticstack_elasticsearch_alias" {
			continue
		}

		// Handle the case where ID might not be in the expected format
		aliasName := rs.Primary.ID
		if compId, err := clients.CompositeIdFromStr(rs.Primary.ID); err == nil {
			aliasName = compId.ResourceId
		}

		esClient, err := client.GetESClient()
		if err != nil {
			return err
		}

		res, err := esClient.Indices.GetAlias(
			esClient.Indices.GetAlias.WithName(aliasName),
		)
		if err != nil {
			return err
		}
		defer res.Body.Close()

		if res.StatusCode != 404 {
			return fmt.Errorf("Alias (%s) still exists", aliasName)
		}
	}
	return nil
}
