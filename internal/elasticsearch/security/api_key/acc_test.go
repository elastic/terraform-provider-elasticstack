package apikey_test

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
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccResourceSecurityAPIKey(t *testing.T) {
	// generate a random name
	apiKeyName := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             checkResourceSecurityAPIKeyDestroy,
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(apikey.MinVersion),
				Config:   testAccResourceSecurityAPIKeyCreate(apiKeyName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_security_apikey.test", "name", apiKeyName),
					resource.TestCheckResourceAttrWith("elasticstack_elasticsearch_security_apikey.test", "role_descriptors", func(testValue string) error {
						var testRoleDescriptor map[string]models.APIKeyRoleDescriptor
						if err := json.Unmarshal([]byte(testValue), &testRoleDescriptor); err != nil {
							return err
						}

						expectedRoleDescriptor := map[string]models.APIKeyRoleDescriptor{
							"role-a": {
								Cluster: []string{"all"},
								Indices: []models.IndexPerms{{
									Names:                  []string{"index-a*"},
									Privileges:             []string{"read"},
									AllowRestrictedIndices: schemautil.Pointer(false),
								}},
							},
						}

						if !reflect.DeepEqual(testRoleDescriptor, expectedRoleDescriptor) {
							return fmt.Errorf("%v doesn't match %v", testRoleDescriptor, expectedRoleDescriptor)
						}

						return nil
					}),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_security_apikey.test", "expiration"),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_security_apikey.test", "api_key"),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_security_apikey.test", "encoded"),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_security_apikey.test", "id"),
				),
			},
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(apikey.MinVersionWithUpdate),
				Config:   testAccResourceSecurityAPIKeyUpdate(apiKeyName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_security_apikey.test", "name", apiKeyName),
					resource.TestCheckResourceAttrWith("elasticstack_elasticsearch_security_apikey.test", "role_descriptors", func(testValue string) error {
						var testRoleDescriptor map[string]models.APIKeyRoleDescriptor
						if err := json.Unmarshal([]byte(testValue), &testRoleDescriptor); err != nil {
							return err
						}

						expectedRoleDescriptor := map[string]models.APIKeyRoleDescriptor{
							"role-a": {
								Cluster: []string{"manage"},
								Indices: []models.IndexPerms{{
									Names:                  []string{"index-b*"},
									Privileges:             []string{"read"},
									AllowRestrictedIndices: schemautil.Pointer(false),
								}},
							},
						}

						if !reflect.DeepEqual(testRoleDescriptor, expectedRoleDescriptor) {
							return fmt.Errorf("%v doesn't match %v", testRoleDescriptor, expectedRoleDescriptor)
						}

						return nil
					}),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_security_apikey.test", "expiration"),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_security_apikey.test", "api_key"),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_security_apikey.test", "encoded"),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_security_apikey.test", "id"),
				),
			},
		},
	})
}

func TestAccResourceSecurityAPIKeyWithRemoteIndices(t *testing.T) {
	minSupportedRemoteIndicesVersion := version.Must(version.NewSemver("8.10.0"))

	// generate a random name
	apiKeyName := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             checkResourceSecurityAPIKeyDestroy,
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minSupportedRemoteIndicesVersion),
				Config:   testAccResourceSecurityAPIKeyRemoteIndices(apiKeyName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_security_apikey.test", "name", apiKeyName),
					resource.TestCheckResourceAttrWith("elasticstack_elasticsearch_security_apikey.test", "role_descriptors", func(testValue string) error {
						var testRoleDescriptor map[string]models.APIKeyRoleDescriptor
						if err := json.Unmarshal([]byte(testValue), &testRoleDescriptor); err != nil {
							return err
						}

						expectedRoleDescriptor := map[string]models.APIKeyRoleDescriptor{
							"role-a": {
								Cluster: []string{"all"},
								Indices: []models.IndexPerms{{
									Names:                  []string{"index-a*"},
									Privileges:             []string{"read"},
									AllowRestrictedIndices: schemautil.Pointer(false),
								}},
								RemoteIndices: []models.RemoteIndexPerms{{
									Clusters: []string{"*"},
									IndexPerms: models.IndexPerms{
										Names:                  []string{"index-a*"},
										Privileges:             []string{"read"},
										AllowRestrictedIndices: schemautil.Pointer(true),
									},
								}},
							},
						}

						if !reflect.DeepEqual(testRoleDescriptor, expectedRoleDescriptor) {
							return fmt.Errorf("%v doesn't match %v", testRoleDescriptor, expectedRoleDescriptor)
						}

						return nil
					}),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_security_apikey.test", "expiration"),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_security_apikey.test", "api_key"),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_security_apikey.test", "encoded"),
				),
			},
		},
	})
}

func TestAccResourceSecurityAPIKeyWithWorkflowRestriction(t *testing.T) {
	// generate a random name
	apiKeyName := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             checkResourceSecurityAPIKeyDestroy,
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(apikey.MinVersionWithRestriction),
				Config:   testAccResourceSecurityAPIKeyCreateWithWorkflowRestriction(apiKeyName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_security_apikey.test", "name", apiKeyName),
					resource.TestCheckResourceAttrWith("elasticstack_elasticsearch_security_apikey.test", "role_descriptors", func(testValue string) error {
						var testRoleDescriptor map[string]models.APIKeyRoleDescriptor
						if err := json.Unmarshal([]byte(testValue), &testRoleDescriptor); err != nil {
							return err
						}

						allowRestrictedIndices := false
						expectedRoleDescriptor := map[string]models.APIKeyRoleDescriptor{
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
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_security_apikey.test", "expiration"),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_security_apikey.test", "api_key"),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_security_apikey.test", "encoded"),
				),
			},
		},
	})
}

func TestAccResourceSecurityAPIKeyWithWorkflowRestrictionOnElasticPre8_9_x(t *testing.T) {
	// generate a random name
	apiKeyName := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	errorPattern := fmt.Sprintf(".*Specifying `restriction` on an API key role description is not supported in this version of Elasticsearch. Role descriptor\\(s\\) %s.*", "role-a")
	errorPattern = strings.ReplaceAll(errorPattern, " ", "\\s+")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             checkResourceSecurityAPIKeyDestroy,
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				SkipFunc:    SkipWhenAPIKeysAreNotSupportedOrRestrictionsAreSupported(apikey.MinVersion, apikey.MinVersionWithRestriction),
				Config:      testAccResourceSecurityAPIKeyCreateWithWorkflowRestriction(apiKeyName),
				ExpectError: regexp.MustCompile(errorPattern),
			},
		},
	})
}

func SkipWhenAPIKeysAreNotSupportedOrRestrictionsAreSupported(minAPIKeySupportedVersion *version.Version, minRestrictionSupportedVersion *version.Version) func() (bool, error) {
	return func() (b bool, err error) {
		client, err := clients.NewAcceptanceTestingClient()
		if err != nil {
			return false, err
		}
		serverVersion, diags := client.ServerVersion(context.Background())
		if diags.HasError() {
			return false, fmt.Errorf("failed to parse the elasticsearch version %v", diags)
		}

		return serverVersion.LessThan(minAPIKeySupportedVersion) || serverVersion.GreaterThanOrEqual(minRestrictionSupportedVersion), nil
	}
}

func TestAccResourceSecurityAPIKeyFromSDK(t *testing.T) {
	// generate a random name
	apiKeyName := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	var initialAPIKey string

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceSecurityAPIKeyDestroy,
		Steps: []resource.TestStep{
			{
				// Create the api_key with the last provider version where the api_key resource was built on the SDK
				ExternalProviders: map[string]resource.ExternalProvider{
					"elasticstack": {
						Source:            "elastic/elasticstack",
						VersionConstraint: "0.11.9",
					},
				},
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(apikey.MinVersion),
				Config:   testAccResourceSecurityAPIKeyWithoutExpiration(apiKeyName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_security_apikey.test", "name", apiKeyName),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_security_apikey.test", "role_descriptors"),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_security_apikey.test", "encoded"),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_security_apikey.test", "id"),
					resource.TestCheckResourceAttrWith("elasticstack_elasticsearch_security_apikey.test", "api_key", func(value string) error {
						initialAPIKey = value

						if value == "" {
							return fmt.Errorf("expected api_key to be non-empty")
						}

						return nil
					}),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(apikey.MinVersion),
				Config:                   testAccResourceSecurityAPIKeyWithoutExpiration(apiKeyName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrWith("elasticstack_elasticsearch_security_apikey.test", "api_key", func(value string) error {
						if value != initialAPIKey {
							return fmt.Errorf("expected api_key to be unchanged")
						}

						return nil
					}),
				),
			},
		},
	})
}

func testAccResourceSecurityAPIKeyCreate(apiKeyName string) string {
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

func testAccResourceSecurityAPIKeyUpdate(apiKeyName string) string {
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

func testAccResourceSecurityAPIKeyWithoutExpiration(apiKeyName string) string {
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
}
	`, apiKeyName)
}

func testAccResourceSecurityAPIKeyRemoteIndices(apiKeyName string) string {
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

func testAccResourceSecurityAPIKeyCreateWithWorkflowRestriction(apiKeyName string) string {
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

func checkResourceSecurityAPIKeyDestroy(s *terraform.State) error {
	client, err := clients.NewAcceptanceTestingClient()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "elasticstack_elasticsearch_security_api_key" {
			continue
		}
		compID, _ := clients.CompositeIDFromStr(rs.Primary.ID)

		apiKey, diags := elasticsearch.GetAPIKey(client, compID.ResourceID)
		if diags.HasError() {
			return fmt.Errorf("Unable to get API key %v", diags)
		}

		if !apiKey.Invalidated {
			return fmt.Errorf("API key (%s) has not been invalidated", compID.ResourceID)
		}
	}
	return nil
}

func TestAccResourceSecurityAPIKeyCrossCluster(t *testing.T) {
	// generate a random name
	apiKeyName := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             checkResourceSecurityAPIKeyDestroy,
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(apikey.MinVersionWithCrossCluster),
				Config:   testAccResourceSecurityAPIKeyCrossClusterCreate(apiKeyName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_security_apikey.test", "name", apiKeyName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_security_apikey.test", "type", "cross_cluster"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_security_apikey.test", "access.search.0.names.0", "logs-*"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_security_apikey.test", "access.search.0.names.1", "metrics-*"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_security_apikey.test", "access.replication.0.names.0", "archive-*"),
				),
			},
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(apikey.MinVersionWithCrossCluster),
				Config:   testAccResourceSecurityAPIKeyCrossClusterUpdate(apiKeyName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_security_apikey.test", "name", apiKeyName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_security_apikey.test", "type", "cross_cluster"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_security_apikey.test", "access.search.0.names.0", "log-*"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_security_apikey.test", "access.search.0.names.1", "metrics-*"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_security_apikey.test", "access.replication.0.names.0", "archives-*"),
				),
			},
		},
	})
}

func testAccResourceSecurityAPIKeyCrossClusterCreate(apiKeyName string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_security_api_key" "test" {
  name = "%s"
  type = "cross_cluster"

  access = {
    search = [
      {
        names = ["logs-*", "metrics-*"]
      }
    ]
    replication = [
      {
        names = ["archive-*"]
      }
    ]
  }

  expiration = "30d"

  metadata = jsonencode({
    description = "Cross-cluster test key"
    environment = "test"
  })
}
	`, apiKeyName)
}

func testAccResourceSecurityAPIKeyCrossClusterUpdate(apiKeyName string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_security_api_key" "test" {
  name = "%s"
  type = "cross_cluster"

  access = {
    search = [
      {
        names = ["log-*", "metrics-*"]
      }
    ]
    replication = [
      {
        names = ["archives-*"]
      }
    ]
  }

  expiration = "30d"

  metadata = jsonencode({
    description = "Cross-cluster test key"
    environment = "test"
  })
}
	`, apiKeyName)
}

func TestAccResourceSecurityAPIKeyWithDefaultAllowRestrictedIndices(t *testing.T) {
	// generate a random name
	apiKeyName := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             checkResourceSecurityAPIKeyDestroy,
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(apikey.MinVersion),
				Config:   testAccResourceSecurityAPIKeyWithoutAllowRestrictedIndices(apiKeyName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_security_apikey.test", "name", apiKeyName),
					resource.TestCheckResourceAttrWith("elasticstack_elasticsearch_security_apikey.test", "role_descriptors", func(testValue string) error {
						var testRoleDescriptor map[string]models.APIKeyRoleDescriptor
						if err := json.Unmarshal([]byte(testValue), &testRoleDescriptor); err != nil {
							return err
						}

						expectedRoleDescriptor := map[string]models.APIKeyRoleDescriptor{
							"role-default": {
								Cluster: []string{"monitor"},
								Indices: []models.IndexPerms{{
									Names:      []string{"logs-*", "metrics-*"},
									Privileges: []string{"read", "view_index_metadata"},
								}},
							},
						}

						if !reflect.DeepEqual(testRoleDescriptor, expectedRoleDescriptor) {
							return fmt.Errorf("role descriptor mismatch:\nexpected: %+v\nactual: %+v", expectedRoleDescriptor, testRoleDescriptor)
						}

						return nil
					}),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_security_apikey.test", "api_key"),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_security_apikey.test", "encoded"),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_security_apikey.test", "id"),
				),
			},
		},
	})
}

func testAccResourceSecurityAPIKeyWithoutAllowRestrictedIndices(apiKeyName string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_security_api_key" "test" {
  name = "%s"

  role_descriptors = jsonencode({
    role-default = {
      cluster = ["monitor"]
      indices = [{
        names = ["logs-*", "metrics-*"]
        privileges = ["read", "view_index_metadata"]
        # Note: allow_restricted_indices is NOT specified here - should default to false
      }]
    }
  })

  expiration = "2d"
}
	`, apiKeyName)
}
