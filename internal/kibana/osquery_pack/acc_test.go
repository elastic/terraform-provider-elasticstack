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

package osquerypack_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/acctest/checks"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	kibanaoapi "github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanautil"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/google/uuid"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

var (
	// Older Kibana versions can return HTTP 500 for deleted or missing Osquery packs.
	minOsqueryPackAccTestVersion = version.Must(version.NewVersion("9.4.0"))

	osqueryPackResourceAddr   = "elasticstack_kibana_osquery_pack.test"
	osqueryPackDataSourceAddr = "data.elasticstack_kibana_osquery_pack.test"

	invalidPlatformErrorPattern = `(?s)platform.*must be one of`

	// accTestKibanaSpaceIDCharset matches elasticstack_kibana_space space_id validation (^[a-z0-9_-]+$).
	accTestKibanaSpaceIDCharset = "abcdefghijklmnopqrstuvwxyz0123456789_-"
)

func skipOsqueryPackUnsupported() func() (bool, error) {
	return versionutils.CheckIfVersionIsUnsupported(minOsqueryPackAccTestVersion)
}

func TestAccResourceOsqueryPack(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minOsqueryPackAccTestVersion, versionutils.FlavorAny)

	suffix := sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlphaNum)
	vars := config.Variables{
		"suffix": config.StringVariable(suffix),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkOsqueryPackDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 skipOsqueryPackUnsupported(),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          vars,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(osqueryPackResourceAddr, "id"),
					resource.TestCheckResourceAttrSet(osqueryPackResourceAddr, "pack_id"),
					testAccCheckCompositeIDFormat(osqueryPackResourceAddr, clients.DefaultSpaceID),
					resource.TestCheckResourceAttr(osqueryPackResourceAddr, "space_id", "default"),
					resource.TestCheckResourceAttr(osqueryPackResourceAddr, "name", fmt.Sprintf("tf-acc-osquery-pack-%s", suffix)),
					resource.TestCheckResourceAttr(osqueryPackResourceAddr, "description", "Terraform acceptance test pack"),
					resource.TestCheckResourceAttr(osqueryPackResourceAddr, "enabled", "true"),
					resource.TestCheckResourceAttr(osqueryPackResourceAddr, "queries.%", "1"),
					resource.TestCheckResourceAttr(osqueryPackResourceAddr, "queries.find_procs.query", "SELECT pid, name FROM processes LIMIT 5;"),
					resource.TestCheckTypeSetElemAttr(osqueryPackResourceAddr, "queries.find_procs.platform.*", "darwin"),
					resource.TestCheckTypeSetElemAttr(osqueryPackResourceAddr, "queries.find_procs.platform.*", "linux"),
					resource.TestCheckResourceAttr(osqueryPackResourceAddr, "queries.find_procs.version", "1.0.0"),
					resource.TestCheckResourceAttr(osqueryPackResourceAddr, "queries.find_procs.snapshot", "false"),
					resource.TestCheckResourceAttr(osqueryPackResourceAddr, "queries.find_procs.removed", "false"),
					resource.TestCheckResourceAttr(osqueryPackResourceAddr, "queries.find_procs.ecs_mapping.process.name.field", "name"),
					resource.TestCheckResourceAttr(osqueryPackResourceAddr, "queries.find_procs.ecs_mapping.process.pid.value", "0"),
					resource.TestCheckTypeSetElemAttr(osqueryPackResourceAddr, "queries.find_procs.ecs_mapping.host.name.values.*", "host-a"),
					resource.TestCheckTypeSetElemAttr(osqueryPackResourceAddr, "queries.find_procs.ecs_mapping.host.name.values.*", "host-b"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 skipOsqueryPackUnsupported(),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables:          vars,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(osqueryPackResourceAddr, "name", fmt.Sprintf("tf-acc-osquery-pack-updated-%s", suffix)),
					resource.TestCheckResourceAttr(osqueryPackResourceAddr, "description", "Updated Terraform acceptance test pack"),
					resource.TestCheckResourceAttr(osqueryPackResourceAddr, "enabled", "false"),
					resource.TestCheckResourceAttr(osqueryPackResourceAddr, "queries.%", "2"),
					resource.TestCheckResourceAttr(osqueryPackResourceAddr, "queries.find_procs.query", "SELECT pid, name, path FROM processes LIMIT 10;"),
					resource.TestCheckTypeSetElemAttr(osqueryPackResourceAddr, "queries.find_procs.platform.*", "linux"),
					resource.TestCheckResourceAttr(osqueryPackResourceAddr, "queries.find_procs.version", "1.1.0"),
					resource.TestCheckResourceAttr(osqueryPackResourceAddr, "queries.find_procs.snapshot", "true"),
					resource.TestCheckResourceAttr(osqueryPackResourceAddr, "queries.list_users.query", "SELECT username FROM users LIMIT 5;"),
					resource.TestCheckTypeSetElemAttr(osqueryPackResourceAddr, "queries.list_users.platform.*", "linux"),
					resource.TestCheckTypeSetElemAttr(osqueryPackResourceAddr, "queries.list_users.platform.*", "windows"),
					resource.TestCheckResourceAttr(osqueryPackResourceAddr, "queries.list_users.version", "2.0.0"),
					resource.TestCheckResourceAttr(osqueryPackResourceAddr, "queries.list_users.removed", "false"),
					resource.TestCheckNoResourceAttr(osqueryPackResourceAddr, "queries.list_users.ecs_mapping.%"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 skipOsqueryPackUnsupported(),
				ResourceName:             osqueryPackResourceAddr,
				ImportState:              true,
				ImportStateVerify:        true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources[osqueryPackResourceAddr]
					if !ok {
						return "", fmt.Errorf("resource not found: %s", osqueryPackResourceAddr)
					}
					return rs.Primary.ID, nil
				},
				ConfigDirectory: acctest.NamedTestCaseDirectory("import"),
				ConfigVariables: vars,
			},
		},
	})
}

func TestAccResourceOsqueryPack_omittedDescriptionUpdate(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minOsqueryPackAccTestVersion, versionutils.FlavorAny)

	suffix := sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlphaNum)
	vars := config.Variables{
		"suffix": config.StringVariable(suffix),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkOsqueryPackDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 skipOsqueryPackUnsupported(),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("omitted_description_create"),
				ConfigVariables:          vars,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckNoResourceAttr(osqueryPackResourceAddr, "description"),
					resource.TestCheckResourceAttr(osqueryPackResourceAddr, "enabled", "true"),
					resource.TestCheckResourceAttr(osqueryPackResourceAddr, "queries.find_procs.platform.#", "0"),
					resource.TestCheckResourceAttrSet(osqueryPackResourceAddr, "queries.find_procs.snapshot"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 skipOsqueryPackUnsupported(),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("omitted_description_update"),
				ConfigVariables:          vars,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckNoResourceAttr(osqueryPackResourceAddr, "description"),
					resource.TestCheckResourceAttr(osqueryPackResourceAddr, "enabled", "false"),
				),
			},
		},
	})
}

func TestAccResourceOsqueryPack_descriptionRemoval(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minOsqueryPackAccTestVersion, versionutils.FlavorAny)

	suffix := sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlphaNum)
	vars := config.Variables{
		"suffix": config.StringVariable(suffix),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkOsqueryPackDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 skipOsqueryPackUnsupported(),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("description_set"),
				ConfigVariables:          vars,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(osqueryPackResourceAddr, "description", "initial description"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 skipOsqueryPackUnsupported(),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("description_unset"),
				ConfigVariables:          vars,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckNoResourceAttr(osqueryPackResourceAddr, "description"),
				),
			},
		},
	})
}

func TestAccResourceOsqueryPack_savedQueryID(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minOsqueryPackAccTestVersion, versionutils.FlavorAny)

	suffix := sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlphaNum)
	vars := config.Variables{
		"suffix": config.StringVariable(suffix),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkOsqueryPackDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 skipOsqueryPackUnsupported(),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("saved_query_create"),
				ConfigVariables:          vars,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(osqueryPackResourceAddr, "queries.find_procs.saved_query_id"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 skipOsqueryPackUnsupported(),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("saved_query_clear"),
				ConfigVariables:          vars,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckNoResourceAttr(osqueryPackResourceAddr, "queries.find_procs.saved_query_id"),
				),
			},
		},
	})
}

// TestAccDataSourceOsqueryPack_savedQueryID verifies the data source behaviour for
// saved_query_id. The Kibana API does not return saved_query_id in GET responses;
// the resource works around this via post-read state merging, but the data source
// has no prior state to merge from. The correct and expected result is therefore
// that saved_query_id is not set on the data source read.
func TestAccDataSourceOsqueryPack_savedQueryID(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minOsqueryPackAccTestVersion, versionutils.FlavorAny)

	suffix := sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlphaNum)
	vars := config.Variables{
		"suffix": config.StringVariable(suffix),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkOsqueryPackDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 skipOsqueryPackUnsupported(),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("saved_query_read"),
				ConfigVariables:          vars,
				Check: resource.ComposeTestCheckFunc(
					// The resource preserves saved_query_id via post-read state merging.
					resource.TestCheckResourceAttrSet(osqueryPackResourceAddr, "queries.find_procs.saved_query_id"),
					// The data source has no prior state; saved_query_id is not returned by the API.
					resource.TestCheckNoResourceAttr(osqueryPackDataSourceAddr, "queries.find_procs.saved_query_id"),
				),
			},
		},
	})
}

func TestAccResourceOsqueryPack_policyIDsAndShards(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minOsqueryPackAccTestVersion, versionutils.FlavorAny)

	suffix := sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlphaNum)
	vars := config.Variables{
		"suffix": config.StringVariable(suffix),
	}

	// policy_ids and shards require the osquery_manager Fleet integration to be
	// installed on the referenced agent policies. The empty-set case exercises the
	// schema path without that dependency.
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkOsqueryPackDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 skipOsqueryPackUnsupported(),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("policy_ids_empty"),
				ConfigVariables:          vars,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(osqueryPackResourceAddr, "policy_ids.#", "0"),
					resource.TestCheckNoResourceAttr(osqueryPackResourceAddr, "shards.%"),
				),
			},
		},
	})
}

func TestAccResourceOsqueryPack_ecsMappingValidator(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minOsqueryPackAccTestVersion, versionutils.FlavorAny)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 skipOsqueryPackUnsupported(),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("invalid_ecs_mapping"),
				PlanOnly:                 true,
				ExpectError:              regexp.MustCompile(`(?s)not more than one`),
			},
		},
	})
}

func TestAccResourceOsqueryPack_invalidPlatform(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minOsqueryPackAccTestVersion, versionutils.FlavorAny)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 skipOsqueryPackUnsupported(),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("invalid_platform"),
				PlanOnly:                 true,
				ExpectError:              regexp.MustCompile(invalidPlatformErrorPattern),
			},
		},
	})
}

func TestAccDataSourceOsqueryPack(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minOsqueryPackAccTestVersion, versionutils.FlavorAny)

	suffix := sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlphaNum)
	vars := config.Variables{
		"suffix": config.StringVariable(suffix),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkOsqueryPackDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 skipOsqueryPackUnsupported(),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				ConfigVariables:          vars,
				Check: func(s *terraform.State) error {
					checks := osqueryPackV1DataSourceParityChecks()
					checks = append([]resource.TestCheckFunc{
						testAccCheckCompositeIDFormat(osqueryPackDataSourceAddr, clients.DefaultSpaceID),
					}, checks...)
					return resource.ComposeTestCheckFunc(checks...)(s)
				},
			},
		},
	})
}

func TestAccPrebuiltOsqueryPack(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minOsqueryPackAccTestVersion, versionutils.FlavorAny)

	prebuiltPackID, skipReason := findPrebuiltOsqueryPack(t, clients.DefaultSpaceID)
	if skipReason != "" {
		t.Skip(skipReason)
	}

	vars := config.Variables{
		"prebuilt_pack_id": config.StringVariable(prebuiltPackID),
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 skipOsqueryPackUnsupported(),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("datasource"),
				ConfigVariables:          vars,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(osqueryPackDataSourceAddr, "pack_id", prebuiltPackID),
					resource.TestCheckResourceAttr(osqueryPackDataSourceAddr, "read_only", "true"),
					resource.TestCheckResourceAttrSet(osqueryPackDataSourceAddr, "name"),
					resource.TestCheckResourceAttrSet(osqueryPackDataSourceAddr, "queries.%"),
					resource.TestCheckResourceAttrSet(osqueryPackDataSourceAddr, "enabled"),
				),
			},
		},
	})
}

func TestAccPrebuiltOsqueryPack_importRejected(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minOsqueryPackAccTestVersion, versionutils.FlavorAny)

	prebuiltPackID, skipReason := findPrebuiltOsqueryPack(t, clients.DefaultSpaceID)
	if skipReason != "" {
		t.Skip(skipReason)
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 skipOsqueryPackUnsupported(),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("import_only"),
				ResourceName:             osqueryPackResourceAddr,
				ImportState:              true,
				ImportStateId:            fmt.Sprintf("%s/%s", clients.DefaultSpaceID, prebuiltPackID),
				ExpectError:              regexp.MustCompile(`read-only \(prebuilt\)`),
			},
		},
	})
}

func TestAccResourceOsqueryPack_nonDefaultSpace(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minOsqueryPackAccTestVersion, versionutils.FlavorAny)

	suffix := sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlphaNum)
	spaceID := sdkacctest.RandStringFromCharSet(12, accTestKibanaSpaceIDCharset)
	vars := config.Variables{
		"suffix":   config.StringVariable(suffix),
		"space_id": config.StringVariable(spaceID),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkOsqueryPackDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 skipOsqueryPackUnsupported(),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          vars,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(osqueryPackResourceAddr, "space_id", spaceID),
					resource.TestCheckResourceAttrSet(osqueryPackResourceAddr, "pack_id"),
					resource.TestCheckResourceAttr(osqueryPackResourceAddr, "name", fmt.Sprintf("tf-acc-osquery-pack-space-%s", suffix)),
					resource.TestMatchResourceAttr(osqueryPackResourceAddr, "id", regexp.MustCompile("^"+regexp.QuoteMeta(spaceID)+"/")),
				),
			},
		},
	})
}

func TestAccDataSourceOsqueryPack_nonDefaultSpace(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minOsqueryPackAccTestVersion, versionutils.FlavorAny)

	suffix := sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlphaNum)
	spaceID := sdkacctest.RandStringFromCharSet(12, accTestKibanaSpaceIDCharset)
	vars := config.Variables{
		"suffix":   config.StringVariable(suffix),
		"space_id": config.StringVariable(spaceID),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkOsqueryPackDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 skipOsqueryPackUnsupported(),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				ConfigVariables:          vars,
				Check: func(s *terraform.State) error {
					checks := osqueryPackV1DataSourceParityChecks()
					checks = append([]resource.TestCheckFunc{
						resource.TestCheckResourceAttr(osqueryPackDataSourceAddr, "space_id", spaceID),
						testAccCheckCompositeIDFormat(osqueryPackDataSourceAddr, spaceID),
					}, checks...)
					return resource.ComposeTestCheckFunc(checks...)(s)
				},
			},
		},
	})
}

func TestAccResourceOsqueryPack_externalDelete(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minOsqueryPackAccTestVersion, versionutils.FlavorAny)

	suffix := sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlphaNum)
	vars := config.Variables{
		"suffix": config.StringVariable(suffix),
	}
	var packSpaceID, packID string

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkOsqueryPackDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 skipOsqueryPackUnsupported(),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          vars,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(osqueryPackResourceAddr, "pack_id"),
					captureOsqueryPackID(osqueryPackResourceAddr, &packSpaceID, &packID),
				),
			},
			{
				PreConfig: func() {
					deleteOsqueryPackAPI(t, mustKibanaOapiClient(t), packSpaceID, packID)
				},
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 skipOsqueryPackUnsupported(),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          vars,
				PlanOnly:                 true,
				ExpectNonEmptyPlan:       true,
				Check:                    testAccCheckOsqueryPackAbsentFromState(osqueryPackResourceAddr),
			},
		},
	})
}

func TestAccResourceOsqueryPack_deleteIdempotent(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minOsqueryPackAccTestVersion, versionutils.FlavorAny)

	suffix := sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlphaNum)
	vars := config.Variables{
		"suffix": config.StringVariable(suffix),
	}
	var packSpaceID, packID string

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkOsqueryPackDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 skipOsqueryPackUnsupported(),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          vars,
				Check:                    captureOsqueryPackID(osqueryPackResourceAddr, &packSpaceID, &packID),
			},
			{
				PreConfig: func() {
					deleteOsqueryPackAPI(t, mustKibanaOapiClient(t), packSpaceID, packID)
				},
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 skipOsqueryPackUnsupported(),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          vars,
				Destroy:                  true,
			},
		},
	})
}

func TestAccDataSourceOsqueryPack_missingPack(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minOsqueryPackAccTestVersion, versionutils.FlavorAny)

	unknownPackID := uuid.NewString()
	vars := config.Variables{
		"unknown_pack_id": config.StringVariable(unknownPackID),
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 skipOsqueryPackUnsupported(),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				ConfigVariables:          vars,
				PlanOnly:                 true,
				ExpectError:              regexp.MustCompile(`Osquery pack not found`),
			},
		},
	})
}

var checkOsqueryPackDestroy = checks.KibanaResourceDestroyCheckCompositeID(
	"elasticstack_kibana_osquery_pack",
	func(ctx context.Context, client *kibanaoapi.Client, spaceID, packID string) (bool, error) {
		if spaceID != clients.DefaultSpaceID {
			space, diags := kibanaoapi.GetSpace(ctx, client, spaceID)
			if diags.HasError() {
				return false, fmt.Errorf("failed to get space %q while checking osquery pack destroy: %v", spaceID, diags)
			}
			if space == nil {
				// Terraform destroys resources in dependency order, so a deleted containing
				// space means the pack is gone. Avoid probing the Osquery route here: Kibana
				// currently returns HTTP 500 for Osquery APIs under non-existent spaces
				// (elastic/kibana#275119), which would make successful destroys look failed.
				return false, nil
			}
		}

		detail, diags := kibanaoapi.GetOsqueryPack(ctx, client, spaceID, packID)
		if diags.HasError() {
			return false, fmt.Errorf("failed to get osquery pack %q in space %q: %v", packID, spaceID, diags)
		}
		return detail != nil, nil
	},
)

func testAccCheckOsqueryPackAbsentFromState(addr string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if _, ok := s.RootModule().Resources[addr]; ok {
			return fmt.Errorf("expected %q to be absent from state after refresh (pack deleted out-of-band)", addr)
		}
		return nil
	}
}

func osqueryPackV1DataSourceParityChecks() []resource.TestCheckFunc {
	return []resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet(osqueryPackResourceAddr, "pack_id"),
		resource.TestCheckResourceAttrSet(osqueryPackDataSourceAddr, "pack_id"),
		resource.TestCheckResourceAttrPair(osqueryPackDataSourceAddr, "id", osqueryPackResourceAddr, "id"),
		resource.TestCheckResourceAttrPair(osqueryPackDataSourceAddr, "pack_id", osqueryPackResourceAddr, "pack_id"),
		resource.TestCheckResourceAttrPair(osqueryPackDataSourceAddr, "name", osqueryPackResourceAddr, "name"),
		resource.TestCheckResourceAttrPair(osqueryPackDataSourceAddr, "description", osqueryPackResourceAddr, "description"),
		resource.TestCheckResourceAttrPair(osqueryPackDataSourceAddr, "enabled", osqueryPackResourceAddr, "enabled"),
		resource.TestCheckResourceAttrPair(osqueryPackDataSourceAddr, "space_id", osqueryPackResourceAddr, "space_id"),
		resource.TestCheckResourceAttr(osqueryPackDataSourceAddr, "read_only", "false"),
		resource.TestCheckResourceAttrPair(osqueryPackDataSourceAddr, "queries.find_procs.query", osqueryPackResourceAddr, "queries.find_procs.query"),
		resource.TestCheckResourceAttrPair(osqueryPackDataSourceAddr, "queries.find_procs.platform.#", osqueryPackResourceAddr, "queries.find_procs.platform.#"),
		resource.TestCheckTypeSetElemAttr(osqueryPackDataSourceAddr, "queries.find_procs.platform.*", "darwin"),
		resource.TestCheckTypeSetElemAttr(osqueryPackDataSourceAddr, "queries.find_procs.platform.*", "linux"),
		resource.TestCheckTypeSetElemAttr(osqueryPackResourceAddr, "queries.find_procs.platform.*", "darwin"),
		resource.TestCheckTypeSetElemAttr(osqueryPackResourceAddr, "queries.find_procs.platform.*", "linux"),
		resource.TestCheckResourceAttrPair(osqueryPackDataSourceAddr, "queries.find_procs.version", osqueryPackResourceAddr, "queries.find_procs.version"),
		resource.TestCheckResourceAttrPair(osqueryPackDataSourceAddr, "queries.find_procs.snapshot", osqueryPackResourceAddr, "queries.find_procs.snapshot"),
		resource.TestCheckResourceAttrPair(osqueryPackDataSourceAddr, "queries.find_procs.removed", osqueryPackResourceAddr, "queries.find_procs.removed"),
		resource.TestCheckResourceAttrPair(osqueryPackDataSourceAddr, "queries.find_procs.ecs_mapping.process.name.field", osqueryPackResourceAddr, "queries.find_procs.ecs_mapping.process.name.field"),
		resource.TestCheckResourceAttrPair(osqueryPackDataSourceAddr, "queries.find_procs.ecs_mapping.process.pid.value", osqueryPackResourceAddr, "queries.find_procs.ecs_mapping.process.pid.value"),
		resource.TestCheckResourceAttrPair(osqueryPackDataSourceAddr, "queries.find_procs.ecs_mapping.host.name.values.#", osqueryPackResourceAddr, "queries.find_procs.ecs_mapping.host.name.values.#"),
		resource.TestCheckTypeSetElemAttr(osqueryPackDataSourceAddr, "queries.find_procs.ecs_mapping.host.name.values.*", "host-a"),
		resource.TestCheckTypeSetElemAttr(osqueryPackDataSourceAddr, "queries.find_procs.ecs_mapping.host.name.values.*", "host-b"),
		resource.TestCheckTypeSetElemAttr(osqueryPackResourceAddr, "queries.find_procs.ecs_mapping.host.name.values.*", "host-a"),
		resource.TestCheckTypeSetElemAttr(osqueryPackResourceAddr, "queries.find_procs.ecs_mapping.host.name.values.*", "host-b"),
		resource.TestCheckResourceAttrPair(osqueryPackDataSourceAddr, "queries.%", osqueryPackResourceAddr, "queries.%"),
		resource.TestCheckResourceAttr(osqueryPackDataSourceAddr, "policy_ids.#", "0"),
		resource.TestCheckResourceAttr(osqueryPackDataSourceAddr, "shards.%", "0"),
	}
}

func testAccCheckCompositeIDFormat(resourceAddr, spaceID string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceAddr]
		if !ok {
			return fmt.Errorf("resource %q not found in state", resourceAddr)
		}

		packID, ok := rs.Primary.Attributes["pack_id"]
		if !ok || packID == "" {
			return fmt.Errorf("resource %q missing pack_id in state", resourceAddr)
		}

		want := spaceID + "/" + packID
		if rs.Primary.ID != want {
			return fmt.Errorf("resource %q id %q, want composite %q", resourceAddr, rs.Primary.ID, want)
		}
		return nil
	}
}

func captureOsqueryPackID(resourceAddr string, spaceID, packID *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceAddr]
		if !ok {
			return fmt.Errorf("resource %q not found in state", resourceAddr)
		}

		compID, diags := clients.CompositeIDFromStr(rs.Primary.ID)
		if diags.HasError() {
			return fmt.Errorf("parse composite id %q: %v", rs.Primary.ID, diags)
		}

		*spaceID = compID.ClusterID
		*packID = compID.ResourceID
		return nil
	}
}

func mustKibanaOapiClient(t *testing.T) *kibanaoapi.Client {
	t.Helper()

	apiClient, err := clients.NewAcceptanceTestingKibanaScopedClient()
	if err != nil {
		t.Fatalf("create Kibana client: %v", err)
	}
	return apiClient.GetKibanaOapiClient()
}

func deleteOsqueryPackAPI(t *testing.T, client *kibanaoapi.Client, spaceID, packID string) {
	t.Helper()

	diags := kibanaoapi.DeleteOsqueryPack(context.Background(), client, spaceID, packID)
	if diags.HasError() {
		t.Fatalf("delete osquery pack %q in space %q: %s", packID, spaceID, diags[0].Summary())
	}
}

// findPrebuiltOsqueryPack returns a read-only prebuilt pack ID from the default space list
// endpoint, or a skip reason when none are available (osquery_manager integration not installed).
func findPrebuiltOsqueryPack(t *testing.T, spaceID string) (packID string, skipReason string) {
	t.Helper()

	apiClient, err := clients.NewAcceptanceTestingKibanaScopedClient()
	if err != nil {
		return "", fmt.Sprintf("skipping prebuilt pack test: could not create Kibana client: %v", err)
	}

	pageSize := kbapi.SecurityOsqueryAPIPageSizeOrUndefined(100)
	client := apiClient.GetKibanaOapiClient()

	for pageNum := 1; ; pageNum++ {
		page := pageNum
		resp, err := client.API.OsqueryFindPacksWithResponse(
			context.Background(),
			&kbapi.OsqueryFindPacksParams{Page: &page, PageSize: &pageSize},
			kibanautil.SpaceAwarePathRequestEditor(spaceID),
		)
		if err != nil {
			return "", fmt.Sprintf("skipping prebuilt pack test: list osquery packs failed: %v", err)
		}
		if resp.StatusCode() != 200 || resp.JSON200 == nil {
			return "", fmt.Sprintf(
				"skipping prebuilt pack test: list osquery packs returned HTTP %d (install osquery_manager integration to seed prebuilt packs)",
				resp.StatusCode(),
			)
		}

		for _, pack := range resp.JSON200.Data {
			if pack.ReadOnly != nil && *pack.ReadOnly && pack.SavedObjectId != "" {
				return pack.SavedObjectId, ""
			}
		}

		if len(resp.JSON200.Data) == 0 || pageNum*pageSize >= resp.JSON200.Total {
			break
		}
	}

	return "", "skipping prebuilt pack test: no read_only prebuilt osquery pack found in default space (install osquery_manager integration to seed prebuilt packs)"
}
