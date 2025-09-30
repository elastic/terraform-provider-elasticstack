package alias_test

import (
	"fmt"
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
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             checkResourceAliasDestroy,
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceAliasCreate(aliasName, indexName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_alias.test_alias", "name", aliasName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_alias.test_alias", "indices.#", "1"),
					resource.TestCheckTypeSetElemAttr("elasticstack_elasticsearch_alias.test_alias", "indices.*", indexName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_alias.test_alias", "is_hidden", "false"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_alias.test_alias", "is_write_index", "false"),
				),
			},
			{
				Config: testAccResourceAliasUpdate(aliasName, indexName, indexName2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_alias.test_alias", "name", aliasName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_alias.test_alias", "indices.#", "2"),
					resource.TestCheckTypeSetElemAttr("elasticstack_elasticsearch_alias.test_alias", "indices.*", indexName),
					resource.TestCheckTypeSetElemAttr("elasticstack_elasticsearch_alias.test_alias", "indices.*", indexName2),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_alias.test_alias", "is_write_index", "true"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_alias.test_alias", "routing", "test-routing"),
				),
			},
			{
				Config: testAccResourceAliasWithFilter(aliasName, indexName),
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

func testAccResourceAliasCreate(aliasName, indexName string) string {
	return fmt.Sprintf(`
resource "elasticstack_elasticsearch_index" "test_index" {
  name = "%s"
}

resource "elasticstack_elasticsearch_alias" "test_alias" {
  name    = "%s"
  indices = [elasticstack_elasticsearch_index.test_index.name]
}
	`, indexName, aliasName)
}

func testAccResourceAliasUpdate(aliasName, indexName, indexName2 string) string {
	return fmt.Sprintf(`
resource "elasticstack_elasticsearch_index" "test_index" {
  name = "%s"
}

resource "elasticstack_elasticsearch_index" "test_index2" {
  name = "%s"
}

resource "elasticstack_elasticsearch_alias" "test_alias" {
  name          = "%s"
  indices       = [elasticstack_elasticsearch_index.test_index.name, elasticstack_elasticsearch_index.test_index2.name]
  is_write_index = true
  routing        = "test-routing"
}
	`, indexName, indexName2, aliasName)
}

func testAccResourceAliasWithFilter(aliasName, indexName string) string {
	return fmt.Sprintf(`
resource "elasticstack_elasticsearch_index" "test_index" {
  name = "%s"
}

resource "elasticstack_elasticsearch_alias" "test_alias" {
  name    = "%s"
  indices = [elasticstack_elasticsearch_index.test_index.name]
  filter  = jsonencode({
    term = {
      status = "published"
    }
  })
}
	`, indexName, aliasName)
}

func testAccResourceAliasDataStreamCreate(aliasName, dsName string) string {
	return fmt.Sprintf(`
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
		compId, _ := clients.CompositeIdFromStr(rs.Primary.ID)

		esClient, err := client.GetESClient()
		if err != nil {
			return err
		}
		
		res, err := esClient.Indices.GetAlias(
			esClient.Indices.GetAlias.WithName(compId.ResourceId),
		)
		if err != nil {
			return err
		}

		if res.StatusCode != 404 {
			return fmt.Errorf("Alias (%s) still exists", compId.ResourceId)
		}
	}
	return nil
}