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

package filter_test

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/elastic/go-elasticsearch/v8/typedapi/ml/putjob"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types/enums/ruleaction"
	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

const mlFilterResourceAddress = "elasticstack_elasticsearch_ml_filter.test"

// mlFilterDestroyBlockedMinElasticsearch is the minimum Elasticsearch version where the acceptance
// tests that assert "delete filter is blocked while a job references it via scoped custom_rules"
// behave consistently with current server semantics in CI (older 8.x matrix images either did
// not reject deletion or diverged on job configuration).
var mlFilterDestroyBlockedMinElasticsearch = version.Must(version.NewVersion("8.5.0"))

func TestAccResourceMLFilter(t *testing.T) {
	filterID := fmt.Sprintf("test-filter-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum))

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"filter_id": config.StringVariable(filterID),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(mlFilterResourceAddress, "filter_id", filterID),
					resource.TestCheckResourceAttr(mlFilterResourceAddress, "description", "Safe domains filter"),
					resource.TestCheckResourceAttr(mlFilterResourceAddress, "items.#", "2"),
					resource.TestCheckTypeSetElemAttr(mlFilterResourceAddress, "items.*", "*.example.com"),
					resource.TestCheckTypeSetElemAttr(mlFilterResourceAddress, "items.*", "trusted.org"),
					resource.TestMatchResourceAttr(mlFilterResourceAddress, "id",
						regexp.MustCompile(`^[A-Za-z0-9_-]{22}/`+regexp.QuoteMeta(filterID)+`$`)),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"filter_id": config.StringVariable(filterID),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(mlFilterResourceAddress, "filter_id", filterID),
					resource.TestCheckResourceAttr(mlFilterResourceAddress, "description", "Updated safe domains filter"),
					resource.TestCheckResourceAttr(mlFilterResourceAddress, "items.#", "3"),
					resource.TestCheckTypeSetElemAttr(mlFilterResourceAddress, "items.*", "*.example.com"),
					resource.TestCheckTypeSetElemAttr(mlFilterResourceAddress, "items.*", "trusted.org"),
					resource.TestCheckTypeSetElemAttr(mlFilterResourceAddress, "items.*", "*.safe.net"),
					resource.TestCheckResourceAttrSet(mlFilterResourceAddress, "id"),
				),
			},
		},
	})
}

func TestAccResourceMLFilterNoItems(t *testing.T) {
	filterID := fmt.Sprintf("test-filter-noitems-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum))

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"filter_id": config.StringVariable(filterID),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(mlFilterResourceAddress, "filter_id", filterID),
					resource.TestCheckResourceAttr(mlFilterResourceAddress, "description", "Empty filter"),
					resource.TestCheckNoResourceAttr(mlFilterResourceAddress, "items.#"),
					resource.TestCheckResourceAttrSet(mlFilterResourceAddress, "id"),
				),
			},
		},
	})
}

func TestAccResourceMLFilterNoDescription(t *testing.T) {
	filterID := fmt.Sprintf("test-filter-nodesc-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum))

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"filter_id": config.StringVariable(filterID),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(mlFilterResourceAddress, "filter_id", filterID),
					resource.TestCheckNoResourceAttr(mlFilterResourceAddress, "description"),
					resource.TestCheckResourceAttrSet(mlFilterResourceAddress, "id"),
				),
			},
		},
	})
}

func TestAccResourceMLFilterImport(t *testing.T) {
	filterID := fmt.Sprintf("test-filter-import-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum))

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"filter_id": config.StringVariable(filterID),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(mlFilterResourceAddress, "filter_id", filterID),
					resource.TestCheckResourceAttr(mlFilterResourceAddress, "description", "Filter for import test"),
					resource.TestCheckTypeSetElemAttr(mlFilterResourceAddress, "items.*", "item-one"),
					resource.TestCheckTypeSetElemAttr(mlFilterResourceAddress, "items.*", "item-two"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ResourceName:             mlFilterResourceAddress,
				ImportState:              true,
				ImportStateVerify:        true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs := s.RootModule().Resources[mlFilterResourceAddress]
					return rs.Primary.ID, nil
				},
				ConfigVariables: config.Variables{
					"filter_id": config.StringVariable(filterID),
				},
			},
		},
	})
}

func TestAccResourceMLFilterFilterIDReplace(t *testing.T) {
	filterID1 := fmt.Sprintf("test-filter-repl-a-%s", sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlphaNum))
	filterID2 := fmt.Sprintf("test-filter-repl-b-%s", sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlphaNum))
	t.Cleanup(func() {
		deleteMLFilterBestEffort(t.Context(), t, filterID1)
		deleteMLFilterBestEffort(t.Context(), t, filterID2)
	})

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"filter_id": config.StringVariable(filterID1),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(mlFilterResourceAddress, "filter_id", filterID1),
					resource.TestCheckResourceAttrSet(mlFilterResourceAddress, "id"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"filter_id": config.StringVariable(filterID2),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(mlFilterResourceAddress, "filter_id", filterID2),
					resource.TestCheckResourceAttrSet(mlFilterResourceAddress, "id"),
					func(_ *terraform.State) error {
						if err := assertMLFilterAbsentES(t.Context(), t, filterID1); err != nil {
							return err
						}
						return assertMLFilterPresentES(t.Context(), t, filterID2)
					},
				),
			},
		},
	})
}

func TestAccResourceMLFilterReconcileAfterOutOfBandChange(t *testing.T) {
	filterID := fmt.Sprintf("test-filter-reconcile-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum))
	t.Cleanup(func() {
		deleteMLFilterBestEffort(t.Context(), t, filterID)
	})

	vars := config.Variables{
		"filter_id": config.StringVariable(filterID),
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          vars,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(mlFilterResourceAddress, "filter_id", filterID),
					resource.TestCheckResourceAttr(mlFilterResourceAddress, "description", "Baseline description for drift reconcile"),
					resource.TestCheckResourceAttr(mlFilterResourceAddress, "items.#", "3"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				PreConfig: func() {
					updateMLFilterDescriptionOutOfBand(t.Context(), t, filterID, "Drifted out-of-band description")
				},
				ConfigDirectory: acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: vars,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(mlFilterResourceAddress, "description", "Baseline description for drift reconcile"),
					resource.TestCheckResourceAttr(mlFilterResourceAddress, "items.#", "3"),
				),
			},
		},
	})
}

func TestAccResourceMLFilterImportFailures(t *testing.T) {
	filterID := fmt.Sprintf("test-filter-impfail-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum))
	t.Cleanup(func() {
		deleteMLFilterBestEffort(t.Context(), t, filterID)
	})
	importVars := config.Variables{
		"filter_id": config.StringVariable(filterID),
	}
	// Capture the cluster UUID segment of the composite import id while the resource is in state.
	// The final import step must run with this address absent from state; otherwise Terraform core
	// returns "Resource already managed by Terraform" before the provider can surface a read error.
	clusterUUID := make([]string, 1)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          importVars,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(mlFilterResourceAddress, "filter_id", filterID),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources[mlFilterResourceAddress]
						if !ok || rs.Primary.ID == "" {
							return fmt.Errorf("no %s in state", mlFilterResourceAddress)
						}
						parts := strings.SplitN(rs.Primary.ID, "/", 2)
						if len(parts) != 2 {
							return fmt.Errorf("unexpected composite id %q", rs.Primary.ID)
						}
						clusterUUID[0] = parts[0]
						return nil
					},
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          importVars,
				ResourceName:             mlFilterResourceAddress,
				ImportState:              true,
				// Default ImportStatePersist=false runs import in a temp working dir while the harness
				// replaces the main dir's config with provider stubs; post-test destroy then loses the
				// elasticsearch block. Persist keeps the full module config on the main working dir.
				ImportStatePersist: true,
				ImportStateVerify:  false,
				ImportStateId:      "not-a-composite-import-id",
				ExpectError:        regexp.MustCompile(`Wrong resource ID`),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          importVars,
				ResourceName:             mlFilterResourceAddress,
				ImportState:              true,
				ImportStatePersist:       true,
				ImportStateVerify:        false,
				ImportStateId:            "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee/extra/bad",
				ExpectError:              regexp.MustCompile(`Wrong resource ID`),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          importVars,
				Destroy:                  true,
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          importVars,
				ResourceName:             mlFilterResourceAddress,
				ImportState:              true,
				ImportStatePersist:       true,
				ImportStateVerify:        false,
				ImportStateId:            fmt.Sprintf("%s/nonexistent-filter-id-for-import-test", clusterUUID[0]),
				ExpectError:              regexp.MustCompile(`Failed to get ML filter|Unable to get ML filter|Cannot import non-existent`),
			},
		},
	})
}

func TestAccResourceMLFilterManyItems(t *testing.T) {
	filterID := fmt.Sprintf("test-filter-many-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum))
	t.Cleanup(func() {
		deleteMLFilterBestEffort(t.Context(), t, filterID)
	})

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"filter_id": config.StringVariable(filterID),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(mlFilterResourceAddress, "filter_id", filterID),
					resource.TestCheckResourceAttr(mlFilterResourceAddress, "items.#", "250"),
				),
			},
		},
	})
}

func TestAccResourceMLFilterDescriptionTooLong(t *testing.T) {
	filterID := fmt.Sprintf("test-filter-longdesc-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum))
	t.Cleanup(func() {
		deleteMLFilterBestEffort(t.Context(), t, filterID)
	})

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"filter_id": config.StringVariable(filterID),
				},
				ExpectError: regexp.MustCompile(`Invalid Attribute Value Length|4096|expected length`),
			},
		},
	})
}

func TestAccResourceMLFilterEmptyDescription(t *testing.T) {
	filterID := fmt.Sprintf("test-filter-emptydesc-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum))
	t.Cleanup(func() {
		deleteMLFilterBestEffort(t.Context(), t, filterID)
	})

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"filter_id": config.StringVariable(filterID),
				},
				ExpectError: regexp.MustCompile(`Invalid Attribute Value Length|between 1 and 4096|got: 0`),
			},
		},
	})
}

func TestAccResourceMLFilterInvalidFilterID(t *testing.T) {
	filterID := "INVALID_UPPERCASE_ID"

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"filter_id": config.StringVariable(filterID),
				},
				ExpectError: regexp.MustCompile(`lowercase|must contain`),
			},
		},
	})
}

func TestAccResourceMLFilterCreateWhenFilterExists(t *testing.T) {
	filterID := fmt.Sprintf("test-filter-dup-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum))
	t.Cleanup(func() {
		deleteMLFilterBestEffort(t.Context(), t, filterID)
	})

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				PreConfig: func() {
					putMLFilterOutOfBand(t.Context(), t, filterID, "Out-of-band filter before Terraform create", []string{"oob.example.com"})
				},
				ConfigDirectory: acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"filter_id": config.StringVariable(filterID),
				},
				ExpectError: regexp.MustCompile(`Failed to create ML filter`),
			},
		},
	})
}

func TestAccResourceMLFilterRecreateAfterRemoteDeleted(t *testing.T) {
	filterID := fmt.Sprintf("test-filter-recreate-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum))
	t.Cleanup(func() {
		deleteMLFilterBestEffort(t.Context(), t, filterID)
	})

	vars := config.Variables{
		"filter_id": config.StringVariable(filterID),
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          vars,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(mlFilterResourceAddress, "filter_id", filterID),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				PreConfig: func() {
					deleteMLFilterStrict(t.Context(), t, filterID)
				},
				ConfigDirectory: acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: vars,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(mlFilterResourceAddress, "filter_id", filterID),
					resource.TestCheckResourceAttr(mlFilterResourceAddress, "description", "Filter deleted out-of-band then recreated by Terraform"),
					resource.TestCheckResourceAttr(mlFilterResourceAddress, "items.#", "2"),
				),
			},
		},
	})
}

func TestAccResourceMLFilterUpdateWhenRemoteDeleted(t *testing.T) {
	filterID := fmt.Sprintf("test-filter-upd-miss-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum))
	t.Cleanup(func() {
		deleteMLFilterBestEffort(t.Context(), t, filterID)
	})

	vars := config.Variables{
		"filter_id": config.StringVariable(filterID),
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		// Without this, the pre-apply plan refreshes state, sees the remote filter is gone,
		// drops the instance from state, and the step becomes a successful create instead of
		// an update that hits the "filter missing" path in the provider.
		AdditionalCLIOptions: &resource.AdditionalCLIOptions{
			Plan: resource.PlanOptions{
				NoRefresh: true,
			},
		},
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          vars,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(mlFilterResourceAddress, "filter_id", filterID),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				PreConfig: func() {
					deleteMLFilterStrict(t.Context(), t, filterID)
				},
				ConfigDirectory: acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: vars,
				ExpectError:     regexp.MustCompile(`Filter not found`),
			},
		},
	})
}

func TestAccResourceMLFilterDestroyWhenRemoteDeleted(t *testing.T) {
	filterID := fmt.Sprintf("test-filter-del-miss-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum))
	t.Cleanup(func() {
		deleteMLFilterBestEffort(t.Context(), t, filterID)
	})

	vars := config.Variables{
		"filter_id": config.StringVariable(filterID),
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          vars,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(mlFilterResourceAddress, "filter_id", filterID),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				PreConfig: func() {
					deleteMLFilterBestEffort(t.Context(), t, filterID)
				},
				ConfigDirectory: acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: vars,
				Destroy:         true,
			},
		},
	})
}

// TestAccResourceMLFilterDestroyBlockedByReferencedJob checks that deleting an ML filter fails while an
// anomaly detection job still references it via detector custom_rules scope (Elasticsearch refuses delete).
// The job is created with the Elasticsearch API because the Terraform ML anomaly detection job resource does not
// yet expose scope on custom rules.
//
// Requires TF_ACC=1 and a cluster where ML anomaly detection jobs can be created (same expectation as other ML acceptance tests).
func TestAccResourceMLFilterDestroyBlockedByReferencedJob(t *testing.T) {
	versionutils.SkipIfUnsupported(t, mlFilterDestroyBlockedMinElasticsearch, versionutils.FlavorAny)

	filterID := fmt.Sprintf("test-filter-blockdel-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum))
	jobID := fmt.Sprintf("test-ad-blockdel-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum))

	t.Cleanup(func() {
		deleteMLJobBestEffort(t.Context(), t, jobID)
		deleteMLFilterBestEffort(t.Context(), t, filterID)
	})

	vars := config.Variables{
		"filter_id": config.StringVariable(filterID),
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          vars,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(mlFilterResourceAddress, "filter_id", filterID),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				PreConfig: func() {
					putMLJobReferencingFilter(t.Context(), t, jobID, filterID)
				},
				ConfigDirectory: acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: vars,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(mlFilterResourceAddress, "filter_id", filterID),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          vars,
				Destroy:                  true,
				ExpectError:              regexp.MustCompile(`Failed to delete ML filter`),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				PreConfig: func() {
					deleteMLJobBestEffort(t.Context(), t, jobID)
				},
				ConfigDirectory: acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: vars,
				Destroy:         true,
			},
		},
	})
}

func TestAccResourceMLFilterDestroyBlockedByTwoReferencedJobs(t *testing.T) {
	versionutils.SkipIfUnsupported(t, mlFilterDestroyBlockedMinElasticsearch, versionutils.FlavorAny)

	filterID := fmt.Sprintf("test-filter-2jobs-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum))
	jobID1 := fmt.Sprintf("test-ad-2jobs-a-%s", sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlphaNum))
	jobID2 := fmt.Sprintf("test-ad-2jobs-b-%s", sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlphaNum))

	t.Cleanup(func() {
		deleteMLJobBestEffort(t.Context(), t, jobID1)
		deleteMLJobBestEffort(t.Context(), t, jobID2)
		deleteMLFilterBestEffort(t.Context(), t, filterID)
	})

	vars := config.Variables{
		"filter_id": config.StringVariable(filterID),
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          vars,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(mlFilterResourceAddress, "filter_id", filterID),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				PreConfig: func() {
					putMLJobReferencingFilter(t.Context(), t, jobID1, filterID)
					putMLJobReferencingFilter(t.Context(), t, jobID2, filterID)
				},
				ConfigDirectory: acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: vars,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(mlFilterResourceAddress, "filter_id", filterID),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          vars,
				Destroy:                  true,
				ExpectError:              regexp.MustCompile(`Failed to delete ML filter`),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				PreConfig: func() {
					deleteMLJobBestEffort(t.Context(), t, jobID1)
					deleteMLJobBestEffort(t.Context(), t, jobID2)
				},
				ConfigDirectory: acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: vars,
				Destroy:         true,
			},
		},
	})
}

func putMLJobReferencingFilter(ctx context.Context, t *testing.T, jobID, filterID string) {
	t.Helper()

	client, err := clients.NewAcceptanceTestingElasticsearchScopedClient()
	if err != nil {
		t.Fatalf("acceptance ES client: %v", err)
	}
	typed, err := client.GetESClient()
	if err != nil {
		t.Fatalf("typed ES client: %v", err)
	}

	detectorFunction := "count"
	partitionField := "host"
	timeField := "@timestamp"
	timeFormat := "epoch_ms"

	_, err = typed.Ml.PutJob(jobID).Request(&putjob.Request{
		AnalysisConfig: types.AnalysisConfig{
			BucketSpan: types.Duration("15m"),
			Detectors: []types.Detector{
				{
					Function:           &detectorFunction,
					PartitionFieldName: &partitionField,
					CustomRules: []types.DetectionRule{
						{
							Actions: []ruleaction.RuleAction{ruleaction.Skipresult},
							Scope: map[string]types.FilterRef{
								"host": {FilterId: filterID},
							},
						},
					},
				},
			},
		},
		DataDescription: types.DataDescription{
			TimeField:  &timeField,
			TimeFormat: &timeFormat,
		},
	}).Do(ctx)
	if err != nil {
		t.Fatalf("Ml.PutJob: %v", err)
	}
}

func deleteMLJobBestEffort(ctx context.Context, t *testing.T, jobID string) {
	t.Helper()

	client, err := clients.NewAcceptanceTestingElasticsearchScopedClient()
	if err != nil {
		t.Logf("Ml.DeleteJob cleanup: no client: %v", err)
		return
	}
	typed, err := client.GetESClient()
	if err != nil {
		t.Logf("Ml.DeleteJob cleanup: %v", err)
		return
	}

	_, err = typed.Ml.DeleteJob(jobID).Force(true).Do(ctx)
	if err == nil {
		return
	}
	var esErr *types.ElasticsearchError
	if errors.As(err, &esErr) && esErr.Status == 404 {
		return
	}
	t.Logf("Ml.DeleteJob %q: %v", jobID, err)
}

func putMLFilterOutOfBand(ctx context.Context, t *testing.T, filterID, description string, items []string) {
	t.Helper()

	client, err := clients.NewAcceptanceTestingElasticsearchScopedClient()
	if err != nil {
		t.Fatalf("acceptance ES client: %v", err)
	}
	typed, err := client.GetESClient()
	if err != nil {
		t.Fatalf("typed ES client: %v", err)
	}

	put := typed.Ml.PutFilter(filterID).Description(description)
	if len(items) > 0 {
		put = put.Items(items...)
	}
	_, err = put.Do(ctx)
	if err != nil {
		t.Fatalf("Ml.PutFilter out-of-band: %v", err)
	}
}

func deleteMLFilterBestEffort(ctx context.Context, t *testing.T, filterID string) {
	t.Helper()

	client, err := clients.NewAcceptanceTestingElasticsearchScopedClient()
	if err != nil {
		t.Logf("Ml.DeleteFilter cleanup: no client: %v", err)
		return
	}
	typed, err := client.GetESClient()
	if err != nil {
		t.Logf("Ml.DeleteFilter cleanup: %v", err)
		return
	}

	_, err = typed.Ml.DeleteFilter(filterID).Do(ctx)
	if err == nil {
		return
	}
	var esErr *types.ElasticsearchError
	if errors.As(err, &esErr) && esErr.Status == 404 {
		return
	}
	t.Logf("Ml.DeleteFilter %q: %v", filterID, err)
}

func deleteMLFilterStrict(ctx context.Context, t *testing.T, filterID string) {
	t.Helper()

	client, err := clients.NewAcceptanceTestingElasticsearchScopedClient()
	if err != nil {
		t.Fatalf("acceptance ES client: %v", err)
	}
	typed, err := client.GetESClient()
	if err != nil {
		t.Fatalf("typed ES client: %v", err)
	}

	_, err = typed.Ml.DeleteFilter(filterID).Do(ctx)
	if err == nil {
		return
	}
	var esErr *types.ElasticsearchError
	if errors.As(err, &esErr) && esErr.Status == 404 {
		return
	}
	t.Fatalf("Ml.DeleteFilter %q: %v", filterID, err)
}

func updateMLFilterDescriptionOutOfBand(ctx context.Context, t *testing.T, filterID, newDescription string) {
	t.Helper()

	client, err := clients.NewAcceptanceTestingElasticsearchScopedClient()
	if err != nil {
		t.Fatalf("acceptance ES client: %v", err)
	}
	typed, err := client.GetESClient()
	if err != nil {
		t.Fatalf("typed ES client: %v", err)
	}

	_, err = typed.Ml.UpdateFilter(filterID).Description(newDescription).Do(ctx)
	if err != nil {
		t.Fatalf("Ml.UpdateFilter out-of-band: %v", err)
	}
}

func assertMLFilterAbsentES(ctx context.Context, t *testing.T, filterID string) error {
	t.Helper()

	client, err := clients.NewAcceptanceTestingElasticsearchScopedClient()
	if err != nil {
		return fmt.Errorf("acceptance ES client: %w", err)
	}
	typed, err := client.GetESClient()
	if err != nil {
		return fmt.Errorf("typed ES client: %w", err)
	}

	res, err := typed.Ml.GetFilters().FilterId(filterID).Do(ctx)
	if err != nil {
		var esErr *types.ElasticsearchError
		if errors.As(err, &esErr) && esErr.Status == 404 {
			return nil
		}
		return fmt.Errorf("get filter %q: %w", filterID, err)
	}
	if len(res.Filters) == 0 {
		return nil
	}
	return fmt.Errorf("expected ML filter %q to be absent in Elasticsearch", filterID)
}

func assertMLFilterPresentES(ctx context.Context, t *testing.T, filterID string) error {
	t.Helper()

	client, err := clients.NewAcceptanceTestingElasticsearchScopedClient()
	if err != nil {
		return fmt.Errorf("acceptance ES client: %w", err)
	}
	typed, err := client.GetESClient()
	if err != nil {
		return fmt.Errorf("typed ES client: %w", err)
	}

	res, err := typed.Ml.GetFilters().FilterId(filterID).Do(ctx)
	if err != nil {
		return fmt.Errorf("get filter %q: %w", filterID, err)
	}
	if len(res.Filters) == 0 {
		return fmt.Errorf("expected ML filter %q to exist in Elasticsearch", filterID)
	}
	return nil
}
