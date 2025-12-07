package role_test

import (
	_ "embed"
	"fmt"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/security/role"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccResourceSecurityRole(t *testing.T) {
	// generate a random username
	roleName := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	roleNameRemoteIndices := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	roleNameDescription := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceSecurityRoleDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"role_name": config.StringVariable(roleName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_security_role.test", "name", roleName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_security_role.test", "indices.0.allow_restricted_indices", "true"),
					resource.TestCheckTypeSetElemAttr("elasticstack_elasticsearch_security_role.test", "indices.*.names.*", "index1"),
					resource.TestCheckTypeSetElemAttr("elasticstack_elasticsearch_security_role.test", "indices.*.names.*", "index2"),
					resource.TestCheckTypeSetElemAttr("elasticstack_elasticsearch_security_role.test", "cluster.*", "all"),
					resource.TestCheckTypeSetElemAttr("elasticstack_elasticsearch_security_role.test", "run_as.*", "other_user"),
					resource.TestCheckNoResourceAttr("elasticstack_elasticsearch_security_role.test", "global"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"role_name": config.StringVariable(roleName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_security_role.test", "name", roleName),
					resource.TestCheckTypeSetElemAttr("elasticstack_elasticsearch_security_role.test", "indices.*.names.*", "index1"),
					resource.TestCheckTypeSetElemAttr("elasticstack_elasticsearch_security_role.test", "indices.*.names.*", "index2"),
					resource.TestCheckTypeSetElemAttr("elasticstack_elasticsearch_security_role.test", "cluster.*", "all"),
					resource.TestCheckNoResourceAttr("elasticstack_elasticsearch_security_role.test", "run_as.#"),
					resource.TestCheckNoResourceAttr("elasticstack_elasticsearch_security_role.test", "global"),
					resource.TestCheckNoResourceAttr("elasticstack_elasticsearch_security_role.test", "applications.#"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(role.MinSupportedRemoteIndicesVersion),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("remote_indices_create"),
				ConfigVariables: config.Variables{
					"role_name": config.StringVariable(roleNameRemoteIndices),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_security_role.test", "name", roleNameRemoteIndices),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_security_role.test", "indices.0.allow_restricted_indices", "true"),
					resource.TestCheckTypeSetElemAttr("elasticstack_elasticsearch_security_role.test", "indices.*.names.*", "index1"),
					resource.TestCheckTypeSetElemAttr("elasticstack_elasticsearch_security_role.test", "indices.*.names.*", "index2"),
					resource.TestCheckTypeSetElemAttr("elasticstack_elasticsearch_security_role.test", "cluster.*", "all"),
					resource.TestCheckTypeSetElemAttr("elasticstack_elasticsearch_security_role.test", "run_as.*", "other_user"),
					resource.TestCheckNoResourceAttr("elasticstack_elasticsearch_security_role.test", "global"),
					resource.TestCheckTypeSetElemAttr("elasticstack_elasticsearch_security_role.test", "remote_indices.*.clusters.*", "test-cluster"),
					resource.TestCheckTypeSetElemAttr("elasticstack_elasticsearch_security_role.test", "remote_indices.*.names.*", "sample"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(role.MinSupportedRemoteIndicesVersion),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("remote_indices_update"),
				ConfigVariables: config.Variables{
					"role_name": config.StringVariable(roleNameRemoteIndices),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_security_role.test", "name", roleNameRemoteIndices),
					resource.TestCheckTypeSetElemAttr("elasticstack_elasticsearch_security_role.test", "indices.*.names.*", "index1"),
					resource.TestCheckTypeSetElemAttr("elasticstack_elasticsearch_security_role.test", "indices.*.names.*", "index2"),
					resource.TestCheckTypeSetElemAttr("elasticstack_elasticsearch_security_role.test", "cluster.*", "all"),
					resource.TestCheckNoResourceAttr("elasticstack_elasticsearch_security_role.test", "run_as.#"),
					resource.TestCheckNoResourceAttr("elasticstack_elasticsearch_security_role.test", "global"),
					resource.TestCheckNoResourceAttr("elasticstack_elasticsearch_security_role.test", "applications.#"),
					resource.TestCheckTypeSetElemAttr("elasticstack_elasticsearch_security_role.test", "remote_indices.*.clusters.*", "test-cluster2"),
					resource.TestCheckTypeSetElemAttr("elasticstack_elasticsearch_security_role.test", "remote_indices.*.names.*", "sample2"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(role.MinSupportedDescriptionVersion),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("description_create"),
				ConfigVariables: config.Variables{
					"role_name": config.StringVariable(roleNameDescription),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_security_role.test", "name", roleNameDescription),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_security_role.test", "description", "test description"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(role.MinSupportedDescriptionVersion),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("description_update"),
				ConfigVariables: config.Variables{
					"role_name": config.StringVariable(roleNameDescription),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_security_role.test", "name", roleNameDescription),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_security_role.test", "description", "updated test description"),
				),
			},
		},
	})
}

//go:embed testdata/TestAccResourceSecurityRoleFromSDK/create/main.tf
var sdkCreateTestConfig string

func TestAccResourceSecurityRoleFromSDK(t *testing.T) {
	roleName := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				// Create the role with the last provider version where the role resource was built on the SDK
				ExternalProviders: map[string]resource.ExternalProvider{
					"elasticstack": {
						Source:            "elastic/elasticstack",
						VersionConstraint: "0.12.2",
					},
				},
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(role.MinSupportedRemoteIndicesVersion),
				Config:   sdkCreateTestConfig,
				ConfigVariables: config.Variables{
					"role_name": config.StringVariable(roleName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_security_role.test", "name", roleName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_security_role.test", "indices.0.allow_restricted_indices", "true"),
					resource.TestCheckTypeSetElemAttr("elasticstack_elasticsearch_security_role.test", "indices.*.names.*", "index1"),
					resource.TestCheckTypeSetElemAttr("elasticstack_elasticsearch_security_role.test", "indices.*.names.*", "index2"),
					resource.TestCheckTypeSetElemAttr("elasticstack_elasticsearch_security_role.test", "cluster.*", "all"),
					resource.TestCheckTypeSetElemAttr("elasticstack_elasticsearch_security_role.test", "run_as.*", "other_user"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"role_name": config.StringVariable(roleName),
				},
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(role.MinSupportedRemoteIndicesVersion),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_security_role.test", "name", roleName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_security_role.test", "indices.0.allow_restricted_indices", "true"),
					resource.TestCheckTypeSetElemAttr("elasticstack_elasticsearch_security_role.test", "indices.*.names.*", "index1"),
					resource.TestCheckTypeSetElemAttr("elasticstack_elasticsearch_security_role.test", "indices.*.names.*", "index2"),
					resource.TestCheckTypeSetElemAttr("elasticstack_elasticsearch_security_role.test", "cluster.*", "all"),
					resource.TestCheckTypeSetElemAttr("elasticstack_elasticsearch_security_role.test", "run_as.*", "other_user"),
				),
			},
		},
	})
}

func checkResourceSecurityRoleDestroy(s *terraform.State) error {
	client, err := clients.NewAcceptanceTestingClient()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "elasticstack_elasticsearch_security_role" {
			continue
		}
		compId, _ := clients.CompositeIdFromStr(rs.Primary.ID)

		esClient, err := client.GetESClient()
		if err != nil {
			return err
		}
		req := esClient.Security.GetRole.WithName(compId.ResourceId)
		res, err := esClient.Security.GetRole(req)
		if err != nil {
			return err
		}

		if res.StatusCode != 404 {
			return fmt.Errorf("role (%s) still exists", compId.ResourceId)
		}
	}
	return nil
}
