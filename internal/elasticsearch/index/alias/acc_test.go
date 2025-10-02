package alias_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/hashicorp/terraform-plugin-testing/config"
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
			// Create indices directly via API to avoid terraform index resource conflicts
			createTestIndex(t, indexName)
			createTestIndex(t, indexName2)
		},
		CheckDestroy:             checkResourceAliasDestroy,
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceAliasCreateDirect,
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
				Config: testAccResourceAliasUpdateDirect,
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
				Config: testAccResourceAliasWithFilterDirect,
				ConfigVariables: map[string]config.Variable{
					"alias_name":  config.StringVariable(aliasName),
					"index_name":  config.StringVariable(indexName),
					"index_name2": config.StringVariable(indexName2),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_alias.test_alias", "name", aliasName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_alias.test_alias", "write_index.name", indexName),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_alias.test_alias", "write_index.filter"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_alias.test_alias", "write_index.index_routing", "write-routing"),
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
		PreCheck: func() {
			acctest.PreCheck(t)
			// Create indices directly via API to avoid terraform index resource conflicts
			createTestIndex(t, indexName1)
			createTestIndex(t, indexName2)
			createTestIndex(t, indexName3)
		},
		CheckDestroy:             checkResourceAliasDestroy,
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			// Case 1: Single index with is_write_index=true
			{
				Config: testAccResourceAliasWriteIndexSingleDirect,
				ConfigVariables: map[string]config.Variable{
					"alias_name":  config.StringVariable(aliasName),
					"index_name1": config.StringVariable(indexName1),
					"index_name2": config.StringVariable(indexName2),
					"index_name3": config.StringVariable(indexName3),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_alias.test_alias", "name", aliasName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_alias.test_alias", "write_index.name", indexName1),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_alias.test_alias", "read_indices.#", "0"),
				),
			},
			// Case 2: Add new index with is_write_index=true, existing becomes read index
			{
				Config: testAccResourceAliasWriteIndexSwitchDirect,
				ConfigVariables: map[string]config.Variable{
					"alias_name":  config.StringVariable(aliasName),
					"index_name1": config.StringVariable(indexName1),
					"index_name2": config.StringVariable(indexName2),
					"index_name3": config.StringVariable(indexName3),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_alias.test_alias", "name", aliasName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_alias.test_alias", "write_index.name", indexName2),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_alias.test_alias", "read_indices.#", "1"),
				),
			},
			// Case 3: Add third index as write index
			{
				Config: testAccResourceAliasWriteIndexTripleDirect,
				ConfigVariables: map[string]config.Variable{
					"alias_name":  config.StringVariable(aliasName),
					"index_name1": config.StringVariable(indexName1),
					"index_name2": config.StringVariable(indexName2),
					"index_name3": config.StringVariable(indexName3),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_alias.test_alias", "name", aliasName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_alias.test_alias", "write_index.name", indexName3),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_alias.test_alias", "read_indices.#", "2"),
				),
			},
			// Case 4: Remove initial index, keep two indices with one as write index
			{
				Config: testAccResourceAliasWriteIndexRemoveFirstDirect,
				ConfigVariables: map[string]config.Variable{
					"alias_name":  config.StringVariable(aliasName),
					"index_name1": config.StringVariable(indexName1),
					"index_name2": config.StringVariable(indexName2),
					"index_name3": config.StringVariable(indexName3),
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
				Config: testAccResourceAliasDataStreamCreate,
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

const testAccResourceAliasDataStreamCreate = `
variable "alias_name" {
  description = "The alias name"
  type        = string
}

variable "ds_name" {
  description = "The data stream name"
  type        = string
}

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

// createTestIndex creates an index directly via API for testing
func createTestIndex(t *testing.T, indexName string) {
	client, err := clients.NewAcceptanceTestingClient()
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	esClient, err := client.GetESClient()
	if err != nil {
		t.Fatalf("Failed to get ES client: %v", err)
	}

	// Create index with mappings
	body := `{
		"mappings": {
			"properties": {
				"title": { "type": "text" },
				"status": { "type": "keyword" }
			}
		}
	}`

	res, err := esClient.Indices.Create(indexName, esClient.Indices.Create.WithBody(strings.NewReader(body)))
	if err != nil {
		t.Fatalf("Failed to create index %s: %v", indexName, err)
	}
	defer res.Body.Close()

	if res.IsError() {
		t.Fatalf("Failed to create index %s: %s", indexName, res.String())
	}
}

const testAccResourceAliasCreateDirect = `
variable "alias_name" {
  description = "The alias name"
  type        = string
}

variable "index_name" {
  description = "The index name"
  type        = string
}

variable "index_name2" {
  description = "The second index name"
  type        = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_alias" "test_alias" {
  name = var.alias_name

  write_index {
    name = var.index_name
  }
}
`

const testAccResourceAliasUpdateDirect = `
variable "alias_name" {
  description = "The alias name"
  type        = string
}

variable "index_name" {
  description = "The index name"
  type        = string
}

variable "index_name2" {
  description = "The second index name"
  type        = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_alias" "test_alias" {
  name = var.alias_name

  write_index {
    name = var.index_name2
  }

  read_indices {
    name = var.index_name
  }
}
`

const testAccResourceAliasWithFilterDirect = `
variable "alias_name" {
  description = "The alias name"
  type        = string
}

variable "index_name" {
  description = "The index name"
  type        = string
}

variable "index_name2" {
  description = "The second index name"
  type        = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_alias" "test_alias" {
  name = var.alias_name

  write_index {
    name    = var.index_name
    index_routing = "write-routing"
    filter = jsonencode({
      term = {
        status = "published"
      }
    })
  }

  read_indices {
    name = var.index_name2
    filter = jsonencode({
      term = {
        status = "draft"
      }
    })
  }
}
`

const testAccResourceAliasWriteIndexSingleDirect = `
variable "alias_name" {
  description = "The alias name"
  type        = string
}

variable "index_name1" {
  description = "The first index name"
  type        = string
}

variable "index_name2" {
  description = "The second index name"
  type        = string
}

variable "index_name3" {
  description = "The third index name"
  type        = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_alias" "test_alias" {
  name = var.alias_name

  write_index {
    name = var.index_name1
  }
}
`

const testAccResourceAliasWriteIndexSwitchDirect = `
variable "alias_name" {
  description = "The alias name"
  type        = string
}

variable "index_name1" {
  description = "The first index name"
  type        = string
}

variable "index_name2" {
  description = "The second index name"
  type        = string
}

variable "index_name3" {
  description = "The third index name"
  type        = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_alias" "test_alias" {
  name = var.alias_name

  write_index {
    name = var.index_name2
  }

  read_indices {
    name = var.index_name1
  }
}
`

const testAccResourceAliasWriteIndexTripleDirect = `
variable "alias_name" {
  description = "The alias name"
  type        = string
}

variable "index_name1" {
  description = "The first index name"
  type        = string
}

variable "index_name2" {
  description = "The second index name"
  type        = string
}

variable "index_name3" {
  description = "The third index name"
  type        = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_alias" "test_alias" {
  name = var.alias_name

  write_index {
    name = var.index_name3
  }

  read_indices {
    name = var.index_name1
  }

  read_indices {
    name = var.index_name2
  }
}
`

const testAccResourceAliasWriteIndexRemoveFirstDirect = `
variable "alias_name" {
  description = "The alias name"
  type        = string
}

variable "index_name1" {
  description = "The first index name"
  type        = string
}

variable "index_name2" {
  description = "The second index name"
  type        = string
}

variable "index_name3" {
  description = "The third index name"
  type        = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_alias" "test_alias" {
  name = var.alias_name

  write_index {
    name = var.index_name3
  }

  read_indices {
    name = var.index_name2
  }
}
`
