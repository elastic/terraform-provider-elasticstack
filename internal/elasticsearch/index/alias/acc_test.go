package alias_test

import (
	"fmt"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/config"
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
				Config: testAccResourceAliasCreate(),
				ConfigVariables: map[string]config.Variable{
					"alias_name":  config.StringVariable(aliasName),
					"index_name":  config.StringVariable(indexName),
					"index_name2": config.StringVariable(indexName2),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_alias.test_alias", "name", aliasName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_alias.test_alias", "write_index.name", indexName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_alias.test_alias", "write_index.is_hidden", "false"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_alias.test_alias", "read_indices.#", "0"),
				),
			},
			{
				Config: testAccResourceAliasUpdate(),
				ConfigVariables: map[string]config.Variable{
					"alias_name":  config.StringVariable(aliasName),
					"index_name":  config.StringVariable(indexName),
					"index_name2": config.StringVariable(indexName2),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_alias.test_alias", "name", aliasName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_alias.test_alias", "write_index.name", indexName2),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_alias.test_alias", "read_indices.#", "1"),
				),
			},
			{
				Config: testAccResourceAliasWithFilter(),
				ConfigVariables: map[string]config.Variable{
					"alias_name":  config.StringVariable(aliasName),
					"index_name":  config.StringVariable(indexName),
					"index_name2": config.StringVariable(indexName2),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_alias.test_alias", "name", aliasName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_alias.test_alias", "write_index.name", indexName),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_alias.test_alias", "write_index.filter"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_alias.test_alias", "write_index.routing", "test-routing"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_alias.test_alias", "read_indices.#", "1"),
				),
			},
		},
	})
}

func TestAccResourceAliasWriteIndex(t *testing.T) {
	// generate random names  
	aliasName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlpha)
	indexName1 := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlpha)
	indexName2 := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlpha)
	indexName3 := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlpha)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             checkResourceAliasDestroy,
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			// Case 1: Single index with is_write_index=true
			{
				Config: testAccResourceAliasWriteIndexSingle(),
				ConfigVariables: map[string]config.Variable{
					"alias_name":   config.StringVariable(aliasName),
					"index_name1":  config.StringVariable(indexName1),
					"index_name2":  config.StringVariable(indexName2),
					"index_name3":  config.StringVariable(indexName3),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_alias.test_alias", "name", aliasName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_alias.test_alias", "write_index.name", indexName1),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_alias.test_alias", "read_indices.#", "0"),
				),
			},
			// Case 2: Add new index with is_write_index=true, existing becomes read index
			{
				Config: testAccResourceAliasWriteIndexSwitch(),
				ConfigVariables: map[string]config.Variable{
					"alias_name":   config.StringVariable(aliasName),
					"index_name1":  config.StringVariable(indexName1),
					"index_name2":  config.StringVariable(indexName2),
					"index_name3":  config.StringVariable(indexName3),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_alias.test_alias", "name", aliasName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_alias.test_alias", "write_index.name", indexName2),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_alias.test_alias", "read_indices.#", "1"),
				),
			},
			// Case 3: Add third index as write index
			{
				Config: testAccResourceAliasWriteIndexTriple(),
				ConfigVariables: map[string]config.Variable{
					"alias_name":   config.StringVariable(aliasName),
					"index_name1":  config.StringVariable(indexName1),
					"index_name2":  config.StringVariable(indexName2),
					"index_name3":  config.StringVariable(indexName3),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_alias.test_alias", "name", aliasName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_alias.test_alias", "write_index.name", indexName3),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_alias.test_alias", "read_indices.#", "2"),
				),
			},
			// Case 4: Remove initial index, keep two indices with one as write index
			{
				Config: testAccResourceAliasWriteIndexRemoveFirst(),
				ConfigVariables: map[string]config.Variable{
					"alias_name":   config.StringVariable(aliasName),
					"index_name1":  config.StringVariable(indexName1),
					"index_name2":  config.StringVariable(indexName2),
					"index_name3":  config.StringVariable(indexName3),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_alias.test_alias", "name", aliasName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_alias.test_alias", "write_index.name", indexName3),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_alias.test_alias", "read_indices.#", "1"),
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
				Config: testAccResourceAliasDataStreamCreate(),
				ConfigVariables: map[string]config.Variable{
					"alias_name": config.StringVariable(aliasName),
					"ds_name":    config.StringVariable(dsName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_alias.test_alias", "name", aliasName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_alias.test_alias", "write_index.name", dsName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_alias.test_alias", "read_indices.#", "0"),
				),
			},
		},
	})
}

const testAccResourceAliasCreate = `
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_index" "test_index" {
  name                = var.index_name
  deletion_protection = false
  
  mappings = jsonencode({
    properties = {
      title = { type = "text" }
    }
  })
}

resource "elasticstack_elasticsearch_index" "test_index2" {
  name                = var.index_name2
  deletion_protection = false
  
  mappings = jsonencode({
    properties = {
      title = { type = "text" }
    }
  })
}

resource "elasticstack_elasticsearch_alias" "test_alias" {
  name = var.alias_name

  write_index {
    name = elasticstack_elasticsearch_index.test_index.name
  }
}
`

const testAccResourceAliasUpdate = `
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_index" "test_index" {
  name                = var.index_name
  deletion_protection = false
  
  mappings = jsonencode({
    properties = {
      title = { type = "text" }
    }
  })
}

resource "elasticstack_elasticsearch_index" "test_index2" {
  name                = var.index_name2
  deletion_protection = false
  
  mappings = jsonencode({
    properties = {
      title = { type = "text" }
    }
  })
}

resource "elasticstack_elasticsearch_alias" "test_alias" {
  name = var.alias_name

  write_index {
    name = elasticstack_elasticsearch_index.test_index2.name
  }

  read_indices {
    name = elasticstack_elasticsearch_index.test_index.name
  }
}
`

const testAccResourceAliasWithFilter = `
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_index" "test_index" {
  name                = var.index_name
  deletion_protection = false
  
  mappings = jsonencode({
    properties = {
      title = { type = "text" }
      status = { type = "keyword" }
    }
  })
}

resource "elasticstack_elasticsearch_index" "test_index2" {
  name                = var.index_name2
  deletion_protection = false
  
  mappings = jsonencode({
    properties = {
      title = { type = "text" }
      status = { type = "keyword" }
    }
  })
}

resource "elasticstack_elasticsearch_alias" "test_alias" {
  name = var.alias_name

  write_index {
    name    = elasticstack_elasticsearch_index.test_index.name
    routing = "test-routing"
    filter = jsonencode({
      term = {
        status = "published"
      }
    })
  }

  read_indices {
    name = elasticstack_elasticsearch_index.test_index2.name
    filter = jsonencode({
      term = {
        status = "draft"
      }
    })
  }
}
`

const testAccResourceAliasWriteIndexSingle = `
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_index" "test_index1" {
  name                = var.index_name1
  deletion_protection = false
}

resource "elasticstack_elasticsearch_index" "test_index2" {
  name                = var.index_name2
  deletion_protection = false
}

resource "elasticstack_elasticsearch_index" "test_index3" {
  name                = var.index_name3
  deletion_protection = false
}

resource "elasticstack_elasticsearch_alias" "test_alias" {
  name = var.alias_name

  write_index {
    name = elasticstack_elasticsearch_index.test_index1.name
  }
}
`

const testAccResourceAliasWriteIndexSwitch = `
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_index" "test_index1" {
  name                = var.index_name1
  deletion_protection = false
}

resource "elasticstack_elasticsearch_index" "test_index2" {
  name                = var.index_name2
  deletion_protection = false
}

resource "elasticstack_elasticsearch_index" "test_index3" {
  name                = var.index_name3
  deletion_protection = false
}

resource "elasticstack_elasticsearch_alias" "test_alias" {
  name = var.alias_name

  write_index {
    name = elasticstack_elasticsearch_index.test_index2.name
  }

  read_indices {
    name = elasticstack_elasticsearch_index.test_index1.name
  }
}
`

const testAccResourceAliasWriteIndexTriple = `
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_index" "test_index1" {
  name                = var.index_name1
  deletion_protection = false
}

resource "elasticstack_elasticsearch_index" "test_index2" {
  name                = var.index_name2
  deletion_protection = false
}

resource "elasticstack_elasticsearch_index" "test_index3" {
  name                = var.index_name3
  deletion_protection = false
}

resource "elasticstack_elasticsearch_alias" "test_alias" {
  name = var.alias_name

  write_index {
    name = elasticstack_elasticsearch_index.test_index3.name
  }

  read_indices {
    name = elasticstack_elasticsearch_index.test_index1.name
  }

  read_indices {
    name = elasticstack_elasticsearch_index.test_index2.name
  }
}
`

const testAccResourceAliasWriteIndexRemoveFirst = `
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_index" "test_index1" {
  name                = var.index_name1
  deletion_protection = false
}

resource "elasticstack_elasticsearch_index" "test_index2" {
  name                = var.index_name2
  deletion_protection = false
}

resource "elasticstack_elasticsearch_index" "test_index3" {
  name                = var.index_name3
  deletion_protection = false
}

resource "elasticstack_elasticsearch_alias" "test_alias" {
  name = var.alias_name

  write_index {
    name = elasticstack_elasticsearch_index.test_index3.name
  }

  read_indices {
    name = elasticstack_elasticsearch_index.test_index2.name
  }
}
`

const testAccResourceAliasDataStreamCreate = `
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_index_template" "test_ds_template" {
  name           = var.ds_name
  index_patterns = [var.ds_name]
  data_stream {}
}

resource "elasticstack_elasticsearch_data_stream" "test_ds" {
  name = var.ds_name
  depends_on = [
    elasticstack_elasticsearch_index_template.test_ds_template
  ]
}

resource "elasticstack_elasticsearch_alias" "test_alias" {
  name = var.alias_name

  write_index {
    name = elasticstack_elasticsearch_data_stream.test_ds.name
  }
}
`

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