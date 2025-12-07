package alias_test

import (
	"fmt"
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
		},
		CheckDestroy: checkResourceAliasDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: map[string]config.Variable{
					"alias_name":  config.StringVariable(aliasName),
					"index_name":  config.StringVariable(indexName),
					"index_name2": config.StringVariable(indexName2),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_alias.test_alias", "name", aliasName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_alias.test_alias", "write_index.name", indexName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_alias.test_alias", "write_index.is_hidden", "false"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_alias.test_alias", "read_indices.#", "0"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: map[string]config.Variable{
					"alias_name":  config.StringVariable(aliasName),
					"index_name":  config.StringVariable(indexName),
					"index_name2": config.StringVariable(indexName2),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_alias.test_alias", "name", aliasName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_alias.test_alias", "write_index.name", indexName2),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_alias.test_alias", "read_indices.#", "1"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("with_filter"),
				ConfigVariables: map[string]config.Variable{
					"alias_name":  config.StringVariable(aliasName),
					"index_name":  config.StringVariable(indexName),
					"index_name2": config.StringVariable(indexName2),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_alias.test_alias", "name", aliasName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_alias.test_alias", "write_index.name", indexName),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_index_alias.test_alias", "write_index.filter"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_alias.test_alias", "write_index.index_routing", "write-routing"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_alias.test_alias", "read_indices.#", "1"),
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
		},
		CheckDestroy: checkResourceAliasDestroy,
		Steps: []resource.TestStep{
			// Case 1: Single index with is_write_index=true
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("single"),
				ConfigVariables: map[string]config.Variable{
					"alias_name":  config.StringVariable(aliasName),
					"index_name1": config.StringVariable(indexName1),
					"index_name2": config.StringVariable(indexName2),
					"index_name3": config.StringVariable(indexName3),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_alias.test_alias", "name", aliasName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_alias.test_alias", "write_index.name", indexName1),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_alias.test_alias", "read_indices.#", "0"),
				),
			},
			// Case 2: Add new index with is_write_index=true, existing becomes read index
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("switch"),
				ConfigVariables: map[string]config.Variable{
					"alias_name":  config.StringVariable(aliasName),
					"index_name1": config.StringVariable(indexName1),
					"index_name2": config.StringVariable(indexName2),
					"index_name3": config.StringVariable(indexName3),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_alias.test_alias", "name", aliasName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_alias.test_alias", "write_index.name", indexName2),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_alias.test_alias", "read_indices.#", "1"),
				),
			},
			// Case 3: Add third index as write index
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("triple"),
				ConfigVariables: map[string]config.Variable{
					"alias_name":  config.StringVariable(aliasName),
					"index_name1": config.StringVariable(indexName1),
					"index_name2": config.StringVariable(indexName2),
					"index_name3": config.StringVariable(indexName3),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_alias.test_alias", "name", aliasName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_alias.test_alias", "write_index.name", indexName3),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_alias.test_alias", "read_indices.#", "2"),
				),
			},
			// Case 4: Remove initial index, keep two indices with one as write index
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("remove_first"),
				ConfigVariables: map[string]config.Variable{
					"alias_name":  config.StringVariable(aliasName),
					"index_name1": config.StringVariable(indexName1),
					"index_name2": config.StringVariable(indexName2),
					"index_name3": config.StringVariable(indexName3),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_alias.test_alias", "name", aliasName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_alias.test_alias", "write_index.name", indexName3),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_alias.test_alias", "read_indices.#", "1"),
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
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceAliasDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: map[string]config.Variable{
					"alias_name": config.StringVariable(aliasName),
					"ds_name":    config.StringVariable(dsName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_alias.test_alias", "name", aliasName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_alias.test_alias", "write_index.name", dsName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_alias.test_alias", "read_indices.#", "0"),
				),
			},
		},
	})
}

func checkResourceAliasDestroy(s *terraform.State) error {
	client, err := clients.NewAcceptanceTestingClient()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "elasticstack_elasticsearch_index_alias" {
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
