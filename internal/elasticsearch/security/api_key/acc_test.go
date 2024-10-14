package api_key_test

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/go-version"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/security/api_key"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccResourceSecurityApiKey(t *testing.T) {
	// generate a random name
	apiKeyName := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             checkResourceSecurityApiKeyDestroy,
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(api_key.MinVersion),
				Config:   testAccResourceSecurityApiKeyCreate(apiKeyName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_security_api_key.test", "name", apiKeyName),
					resource.TestCheckResourceAttrWith("elasticstack_elasticsearch_security_api_key.test", "role_descriptors", func(testValue string) error {
						var testRoleDescriptor map[string]models.ApiKeyRoleDescriptor
						if err := json.Unmarshal([]byte(testValue), &testRoleDescriptor); err != nil {
							return err
						}

						expectedRoleDescriptor := map[string]models.ApiKeyRoleDescriptor{
							"role-a": {
								Cluster: []string{"all"},
								Indices: []models.IndexPerms{{
									Names:                  []string{"index-a*"},
									Privileges:             []string{"read"},
									AllowRestrictedIndices: utils.Pointer(false),
								}},
							},
						}

						if !reflect.DeepEqual(testRoleDescriptor, expectedRoleDescriptor) {
							return fmt.Errorf("%v doesn't match %v", testRoleDescriptor, expectedRoleDescriptor)
						}

						return nil
					}),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_security_api_key.test", "expiration"),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_security_api_key.test", "api_key"),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_security_api_key.test", "encoded"),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_security_api_key.test", "id"),
				),
			},
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(api_key.MinVersionWithUpdate),
				Config:   testAccResourceSecurityApiKeyUpdate(apiKeyName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_security_api_key.test", "name", apiKeyName),
					resource.TestCheckResourceAttrWith("elasticstack_elasticsearch_security_api_key.test", "role_descriptors", func(testValue string) error {
						var testRoleDescriptor map[string]models.ApiKeyRoleDescriptor
						if err := json.Unmarshal([]byte(testValue), &testRoleDescriptor); err != nil {
							return err
						}

						expectedRoleDescriptor := map[string]models.ApiKeyRoleDescriptor{
							"role-a": {
								Cluster: []string{"manage"},
								Indices: []models.IndexPerms{{
									Names:                  []string{"index-b*"},
									Privileges:             []string{"read"},
									AllowRestrictedIndices: utils.Pointer(false),
								}},
							},
						}

						if !reflect.DeepEqual(testRoleDescriptor, expectedRoleDescriptor) {
							return fmt.Errorf("%v doesn't match %v", testRoleDescriptor, expectedRoleDescriptor)
						}

						return nil
					}),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_security_api_key.test", "expiration"),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_security_api_key.test", "api_key"),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_security_api_key.test", "encoded"),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_security_api_key.test", "id"),
				),
			},
		},
	})
}

func TestAccResourceSecurityApiKeyWithRemoteIndices(t *testing.T) {
	minSupportedRemoteIndicesVersion := version.Must(version.NewSemver("8.10.0"))

	// generate a random name
	apiKeyName := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             checkResourceSecurityApiKeyDestroy,
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minSupportedRemoteIndicesVersion),
				Config:   testAccResourceSecurityApiKeyRemoteIndices(apiKeyName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_security_api_key.test", "name", apiKeyName),
					resource.TestCheckResourceAttrWith("elasticstack_elasticsearch_security_api_key.test", "role_descriptors", func(testValue string) error {
						var testRoleDescriptor map[string]models.ApiKeyRoleDescriptor
						if err := json.Unmarshal([]byte(testValue), &testRoleDescriptor); err != nil {
							return err
						}

						expectedRoleDescriptor := map[string]models.ApiKeyRoleDescriptor{
							"role-a": {
								Cluster: []string{"all"},
								Indices: []models.IndexPerms{{
									Names:                  []string{"index-a*"},
									Privileges:             []string{"read"},
									AllowRestrictedIndices: utils.Pointer(false),
								}},
								RemoteIndices: []models.RemoteIndexPerms{{
									Clusters:               []string{"*"},
									Names:                  []string{"index-a*"},
									Privileges:             []string{"read"},
									AllowRestrictedIndices: utils.Pointer(true),
								}},
							},
						}

						if !reflect.DeepEqual(testRoleDescriptor, expectedRoleDescriptor) {
							return fmt.Errorf("%v doesn't match %v", testRoleDescriptor, expectedRoleDescriptor)
						}

						return nil
					}),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_security_api_key.test", "expiration"),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_security_api_key.test", "api_key"),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_security_api_key.test", "encoded"),
				),
			},
		},
	})
}

func TestAccResourceSecurityApiKeyWithWorkflowRestriction(t *testing.T) {
	// generate a random name
	apiKeyName := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             checkResourceSecurityApiKeyDestroy,
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(api_key.MinVersionWithRestriction),
				Config:   testAccResourceSecurityApiKeyCreateWithWorkflowRestriction(apiKeyName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_security_api_key.test", "name", apiKeyName),
					resource.TestCheckResourceAttrWith("elasticstack_elasticsearch_security_api_key.test", "role_descriptors", func(testValue string) error {
						var testRoleDescriptor map[string]models.ApiKeyRoleDescriptor
						if err := json.Unmarshal([]byte(testValue), &testRoleDescriptor); err != nil {
							return err
						}

						allowRestrictedIndices := false
						expectedRoleDescriptor := map[string]models.ApiKeyRoleDescriptor{
							"role-a": {
								Cluster: []string{"all"},
								Indices: []models.IndexPerms{{
									Names:                  []string{"index-a*"},
									Privileges:             []string{"read"},
									AllowRestrictedIndices: &allowRestrictedIndices,
								}},
								Restriction: &models.Restriction{Workflows: []string{"search_application_query"}},
							},
						}

						if !reflect.DeepEqual(testRoleDescriptor, expectedRoleDescriptor) {
							return fmt.Errorf("%v doesn't match %v", testRoleDescriptor, expectedRoleDescriptor)
						}

						return nil
					}),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_security_api_key.test", "expiration"),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_security_api_key.test", "api_key"),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_security_api_key.test", "encoded"),
				),
			},
		},
	})
}

func TestAccResourceSecurityApiKeyWithWorkflowRestrictionOnElasticPre8_9_x(t *testing.T) {
	// generate a random name
	apiKeyName := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	errorPattern := fmt.Sprintf(".*Specifying `restriction` on an API key role description is not supported in this version of Elasticsearch. Role descriptor\\(s\\) %s.*", "role-a")
	errorPattern = strings.ReplaceAll(errorPattern, " ", "\\s+")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             checkResourceSecurityApiKeyDestroy,
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				SkipFunc:    SkipWhenApiKeysAreNotSupportedOrRestrictionsAreSupported(api_key.MinVersion, api_key.MinVersionWithRestriction),
				Config:      testAccResourceSecurityApiKeyCreateWithWorkflowRestriction(apiKeyName),
				ExpectError: regexp.MustCompile(errorPattern),
			},
		},
	})
}

func SkipWhenApiKeysAreNotSupportedOrRestrictionsAreSupported(minApiKeySupportedVersion *version.Version, minRestrictionSupportedVersion *version.Version) func() (bool, error) {
	return func() (b bool, err error) {
		client, err := clients.NewAcceptanceTestingClient()
		if err != nil {
			return false, err
		}
		serverVersion, diags := client.ServerVersion(context.Background())
		if diags.HasError() {
			return false, fmt.Errorf("failed to parse the elasticsearch version %v", diags)
		}

		return serverVersion.LessThan(minApiKeySupportedVersion) || serverVersion.GreaterThanOrEqual(minRestrictionSupportedVersion), nil
	}
}

func testAccResourceSecurityApiKeyCreate(apiKeyName string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_security_api_key" "test" {
  name = "%s"

  role_descriptors = jsonencode({
    role-a = {
      cluster = ["all"]
      indices = [{
        names = ["index-a*"]
        privileges = ["read"]
        allow_restricted_indices = false
      }]
	}
  })

	expiration = "1d"
}
	`, apiKeyName)
}

func testAccResourceSecurityApiKeyUpdate(apiKeyName string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_security_api_key" "test" {
  name = "%s"

  role_descriptors = jsonencode({
    role-a = {
      cluster = ["manage"]
      indices = [{
        names = ["index-b*"]
        privileges = ["read"]
        allow_restricted_indices = false
      }]
	}
  })

	expiration = "1d"
}
	`, apiKeyName)
}

func testAccResourceSecurityApiKeyRemoteIndices(apiKeyName string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_security_api_key" "test" {
  name = "%s"

  role_descriptors = jsonencode({
    role-a = {
      cluster = ["all"]
      indices = [{
        names = ["index-a*"]
        privileges = ["read"]
        allow_restricted_indices = false
      }]
      remote_indices = [{
	    clusters = ["*"]
		names = ["index-a*"]
		privileges = ["read"]
		allow_restricted_indices = true
	  }]
	}
  })

	expiration = "1d"
}
	`, apiKeyName)
}

func testAccResourceSecurityApiKeyCreateWithWorkflowRestriction(apiKeyName string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_security_api_key" "test" {
  name = "%s"

  role_descriptors = jsonencode({
    role-a = {
      cluster = ["all"]
      indices = [{
        names = ["index-a*"]
        privileges = ["read"]
        allow_restricted_indices = false
      }],
      restriction = {
		workflows = [ "search_application_query"]
      }
    }
  })

	expiration = "1d"
}
	`, apiKeyName)
}

func checkResourceSecurityApiKeyDestroy(s *terraform.State) error {
	client, err := clients.NewAcceptanceTestingClient()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "elasticstack_elasticsearch_security_api_key" {
			continue
		}
		compId, _ := clients.CompositeIdFromStr(rs.Primary.ID)

		apiKey, diags := elasticsearch.GetApiKey(client, compId.ResourceId)
		if diags.HasError() {
			return fmt.Errorf("Unabled to get API key %v", diags)
		}

		if !apiKey.Invalidated {
			return fmt.Errorf("ApiKey (%s) has not been invalidated", compId.ResourceId)
		}
	}
	return nil
}
