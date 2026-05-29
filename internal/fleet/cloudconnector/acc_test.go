// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package cloudconnector_test

import (
	"fmt"
	"maps"
	"regexp"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccResourceFleetCloudConnector_AWSLifecycle(t *testing.T) {
	t.Parallel()
	versionutils.SkipIfUnsupported(t, minCloudConnectorVersion, versionutils.FlavorAny)

	suffix := accRandSuffix()
	name := fmt.Sprintf("acc-aws-%s", suffix)
	roleArn := "arn:aws:iam::123456789012:role/ElasticFleetAcc"
	externalID := accExternalID()
	updatedName := fmt.Sprintf("acc-aws-upd-%s", suffix)
	updatedRoleArn := "arn:aws:iam::123456789012:role/ElasticFleetAccUpdated"

	awsVars := config.Variables{
		"role_arn":    config.StringVariable(roleArn),
		"external_id": config.StringVariable(externalID),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkCloudConnectorDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: mergeConfigVariables(awsVars, config.Variables{
					"name": config.StringVariable(name),
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceName, "cloud_provider", "aws"),
					resource.TestCheckResourceAttr(resourceName, "aws.role_arn", roleArn),
					resource.TestCheckResourceAttrSet(resourceName, "cloud_connector_id"),
					testCheckCloudConnectorHasTypedAWS(),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update_name"),
				ConfigVariables: mergeConfigVariables(awsVars, config.Variables{
					"name": config.StringVariable(updatedName),
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", updatedName),
					testCheckCloudConnectorHasTypedAWS(),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update_role_arn"),
				ConfigVariables: mergeConfigVariables(awsVars, config.Variables{
					"name":     config.StringVariable(updatedName),
					"role_arn": config.StringVariable(updatedRoleArn),
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "aws.role_arn", updatedRoleArn),
					testCheckCloudConnectorHasTypedAWS(),
				),
			},
		},
	})
}

func TestAccResourceFleetCloudConnector_AzureLifecycle(t *testing.T) {
	t.Parallel()
	versionutils.SkipIfUnsupported(t, minCloudConnectorVersion, versionutils.FlavorAny)

	suffix := accRandSuffix()
	name := fmt.Sprintf("acc-azure-%s", suffix)
	tenantID := accExternalID()
	clientID := accExternalID()
	connectorID := fmt.Sprintf("azure-conn-%s", suffix)
	updatedName := fmt.Sprintf("acc-azure-upd-%s", suffix)

	azureVars := config.Variables{
		"tenant_id":          config.StringVariable(tenantID),
		"client_id":          config.StringVariable(clientID),
		"cloud_connector_id": config.StringVariable(connectorID),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkCloudConnectorDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: mergeConfigVariables(azureVars, config.Variables{
					"name": config.StringVariable(name),
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceName, "cloud_provider", "azure"),
					resource.TestCheckResourceAttr(resourceName, "azure.cloud_connector_id", connectorID),
					testCheckCloudConnectorHasTypedAzure(),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update_name"),
				ConfigVariables: mergeConfigVariables(azureVars, config.Variables{
					"name": config.StringVariable(updatedName),
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", updatedName),
					testCheckCloudConnectorHasTypedAzure(),
				),
			},
		},
	})
}

// TestAccResourceFleetCloudConnector_VarsAllUnionArms exercises bare string, number,
// bool, and structured text union arms via a GCP connector. Password union arms are
// covered by AWSLifecycle and WriteOnlyDrift.
func TestAccResourceFleetCloudConnector_VarsAllUnionArms(t *testing.T) {
	t.Parallel()
	versionutils.SkipIfUnsupported(t, minCloudConnectorVersion, versionutils.FlavorAny)

	suffix := accRandSuffix()
	name := fmt.Sprintf("acc-vars-%s", suffix)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkCloudConnectorDestroy,
		Steps: []resource.TestStep{{
			ProtoV6ProviderFactories: acctest.Providers,
			ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
			ConfigVariables: config.Variables{
				"name":               config.StringVariable(name),
				"gcp_credentials_id": config.StringVariable(fmt.Sprintf("gcp-credentials-id-%s", suffix)),
			},
			Check: resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr(resourceName, "cloud_provider", "gcp"),
				testCheckCloudConnectorHasVarKeys(
					"service_account",
					"audience",
					"gcp_credentials_cloud_connector_id",
					"custom_struct_text",
				),
				resource.TestCheckResourceAttr(resourceName, "vars.custom_string.string", "bare-string-arm"),
				resource.TestCheckResourceAttr(resourceName, "vars.custom_number.number", "42.5"),
				resource.TestCheckResourceAttr(resourceName, "vars.custom_bool.bool", "true"),
				resource.TestCheckResourceAttr(resourceName, "vars.custom_struct_text.value", "structured-text-arm"),
			),
		}},
	})
}

func TestAccResourceFleetCloudConnector_DualStatePopulation(t *testing.T) {
	t.Parallel()
	versionutils.SkipIfUnsupported(t, minCloudConnectorVersion, versionutils.FlavorAny)

	t.Run("typed_aws_block", func(t *testing.T) {
		t.Parallel()
		suffix := accRandSuffix()
		resource.Test(t, resource.TestCase{
			PreCheck:     func() { acctest.PreCheck(t) },
			CheckDestroy: checkCloudConnectorDestroy,
			Steps: []resource.TestStep{{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"name":        config.StringVariable(fmt.Sprintf("acc-dual-aws-%s", suffix)),
					"role_arn":    config.StringVariable("arn:aws:iam::123456789012:role/ElasticFleetAcc"),
					"external_id": config.StringVariable(accExternalID()),
				},
				Check: testCheckCloudConnectorHasTypedAWS(),
			}},
		})
	})

	t.Run("vars_matching_aws_keys", func(t *testing.T) {
		// Plugin Framework limitation: vars-only create cannot plan the read-populated aws
		// sibling when aws contains write-only children (see design.md Decision 4).
		// Covered by unit tests and typed_aws_block acc scenario.
		t.Skip("vars-only create cannot plan the read-populated aws sibling under Plugin Framework rules; covered by unit tests and typed_aws_block acc scenario")
	})

	t.Run("vars_with_extra_key", func(t *testing.T) {
		t.Parallel()
		suffix := accRandSuffix()
		resource.Test(t, resource.TestCase{
			PreCheck:     func() { acctest.PreCheck(t) },
			CheckDestroy: checkCloudConnectorDestroy,
			Steps: []resource.TestStep{{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"name":               config.StringVariable(fmt.Sprintf("acc-dual-extra-%s", suffix)),
					"gcp_credentials_id": config.StringVariable(fmt.Sprintf("gcp-extra-%s", suffix)),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckNoResourceAttr(resourceName, "aws.role_arn"),
					testCheckCloudConnectorHasVarKeys("service_account", "audience", "gcp_credentials_cloud_connector_id"),
					resource.TestCheckResourceAttr(resourceName, "vars.custom_thing.string", "extra"),
				),
			}},
		})
	})
}

func TestAccResourceFleetCloudConnector_Import(t *testing.T) {
	t.Parallel()
	versionutils.SkipIfUnsupported(t, minCloudConnectorVersion, versionutils.FlavorAny)

	suffix := accRandSuffix()
	name := fmt.Sprintf("acc-import-%s", suffix)
	importVars := config.Variables{
		"name":        config.StringVariable(name),
		"role_arn":    config.StringVariable("arn:aws:iam::123456789012:role/ElasticFleetAcc"),
		"external_id": config.StringVariable(accExternalID()),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkCloudConnectorDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          importVars,
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          importVars,
				ResourceName:             resourceName,
				ImportState:              true,
				ImportStateVerify:        true,
				ImportStateVerifyIgnore: []string{
					"aws.external_id",
					"vars.external_id.secret_value",
					"force_delete",
				},
			},
		},
	})
}

// TestAccResourceFleetCloudConnector_ForceDelete validates the in-use delete
// error path when a cloud connector is referenced by a package policy.
func TestAccResourceFleetCloudConnector_ForceDelete(t *testing.T) {
	t.Parallel()
	versionutils.SkipIfUnsupported(t, minCloudConnectorVersion, versionutils.FlavorAny)

	suffix := accRandSuffix()
	name := fmt.Sprintf("acc-force-%s", suffix)
	externalID := accExternalID()
	forceVars := config.Variables{
		"name":        config.StringVariable(name),
		"role_arn":    config.StringVariable("arn:aws:iam::123456789012:role/ElasticFleetAcc"),
		"external_id": config.StringVariable(externalID),
	}

	var connectorID string

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkCloudConnectorDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          forceVars,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "cloud_connector_id"),
					testCheckCaptureCloudConnectorID(&connectorID),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("destroy_without_force"),
				ConfigVariables:          forceVars,
				PreConfig: func() {
					attachPackagePolicyToCloudConnector(t, connectorID)
				},
				Destroy:     true,
				ExpectError: regexp.MustCompile(`(?i)(force_delete|package polic)`),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("force_delete"),
				ConfigVariables:          forceVars,
				Check:                    resource.TestCheckResourceAttr(resourceName, "force_delete", "true"),
			},
		},
	})
}

func TestAccResourceFleetCloudConnector_WriteOnlyDrift(t *testing.T) {
	t.Parallel()
	versionutils.SkipIfUnsupported(t, minCloudConnectorVersion, versionutils.FlavorAny)

	suffix := accRandSuffix()
	name := fmt.Sprintf("acc-wo-drift-%s", suffix)
	roleArn := "arn:aws:iam::123456789012:role/ElasticFleetAcc"
	secretV1 := accExternalID()
	secretV2 := accExternalID()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkCloudConnectorDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"name":        config.StringVariable(name),
					"role_arn":    config.StringVariable(roleArn),
					"external_id": config.StringVariable(secretV1),
				},
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update_secret"),
				ConfigVariables: config.Variables{
					"name":        config.StringVariable(name),
					"role_arn":    config.StringVariable(roleArn),
					"external_id": config.StringVariable(secretV2),
				},
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks:   expectWriteOnlyDriftPlanChecks("aws.external_id"),
			},
		},
	})
}

func TestAccResourceFleetCloudConnector_VersionGating(t *testing.T) {
	t.Skip("requires a Kibana older than the cloud connector minimum version; gating is unit-tested in TestGetVersionRequirements")
}

func TestAccDataSourceFleetCloudConnectors(t *testing.T) {
	t.Parallel()
	versionutils.SkipIfUnsupported(t, minCloudConnectorVersion, versionutils.FlavorAny)

	suffix := accRandSuffix()
	spaceID := fmt.Sprintf("cc-ds-%s", suffix)
	emptySpaceID := fmt.Sprintf("cc-ds-empty-%s", suffix)
	awsName1 := fmt.Sprintf("acc-ds-aws1-%s", suffix)
	awsName2 := fmt.Sprintf("acc-ds-aws2-%s", suffix)
	azureName := fmt.Sprintf("acc-ds-azure-%s", suffix)
	awsExternalID1 := accExternalID()
	awsExternalID2 := accExternalID()
	azureTenantID := accExternalID()
	azureClientID := accExternalID()
	azureConnectorID := fmt.Sprintf("azure-ds-%s", suffix)

	dsVars := config.Variables{
		"space_id":                 config.StringVariable(spaceID),
		"empty_space_id":           config.StringVariable(emptySpaceID),
		"space_name":               config.StringVariable(fmt.Sprintf("Cloud Connector DS %s", suffix)),
		"empty_space_name":         config.StringVariable(fmt.Sprintf("Cloud Connector DS Empty %s", suffix)),
		"aws_name_1":               config.StringVariable(awsName1),
		"aws_name_2":               config.StringVariable(awsName2),
		"azure_name":               config.StringVariable(azureName),
		"role_arn":                 config.StringVariable("arn:aws:iam::123456789012:role/ElasticFleetAcc"),
		"aws_external_id_1":        config.StringVariable(awsExternalID1),
		"aws_external_id_2":        config.StringVariable(awsExternalID2),
		"azure_tenant_id":          config.StringVariable(azureTenantID),
		"azure_client_id":          config.StringVariable(azureClientID),
		"azure_cloud_connector_id": config.StringVariable(azureConnectorID),
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          dsVars,
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				ConfigVariables:          dsVars,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_fleet_cloud_connectors.all", "cloud_connectors.#", "3"),
					resource.TestCheckResourceAttr("data.elasticstack_fleet_cloud_connectors.aws_only", "cloud_connectors.#", "2"),
					resource.TestCheckResourceAttr("data.elasticstack_fleet_cloud_connectors.aws_only", "cloud_connectors.0.cloud_provider", "aws"),
					resource.TestCheckResourceAttr("data.elasticstack_fleet_cloud_connectors.aws_only", "cloud_connectors.1.cloud_provider", "aws"),
					resource.TestCheckResourceAttr("data.elasticstack_fleet_cloud_connectors.empty_space", "cloud_connectors.#", "0"),
				),
			},
		},
	})
}

func mergeConfigVariables(base, extra config.Variables) config.Variables {
	merged := make(config.Variables, len(base)+len(extra))
	maps.Copy(merged, base)
	maps.Copy(merged, extra)
	return merged
}
