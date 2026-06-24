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

package osquerysavedquery_test

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"testing"
	"time"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/acctest/checks"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	kibanaoapi "github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/google/uuid"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

const (
	resourceName   = "elasticstack_kibana_osquery_saved_query.test"
	dataSourceName = "data.elasticstack_kibana_osquery_saved_query.test"
)

var minOsquerySavedQueryVersion = version.Must(version.NewVersion("8.5.0"))

func TestAccResourceOsquerySavedQuery(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minOsquerySavedQueryVersion, versionutils.FlavorAny)

	savedQueryID := "tf-osquery-" + uuid.New().String()
	spaceID := clients.DefaultSpaceID

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceOsquerySavedQueryDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"saved_query_id": config.StringVariable(savedQueryID),
					"space_id":       config.StringVariable(spaceID),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "saved_query_id", savedQueryID),
					resource.TestCheckResourceAttr(resourceName, "space_id", spaceID),
					resource.TestCheckResourceAttr(resourceName, "id", fmt.Sprintf("%s/%s", spaceID, savedQueryID)),
					resource.TestCheckResourceAttr(resourceName, "query", "SELECT pid, name FROM processes LIMIT 5;"),
					resource.TestCheckResourceAttr(resourceName, "description", "Terraform acceptance create"),
					resource.TestCheckResourceAttr(resourceName, "interval", "3600"),
					resource.TestCheckResourceAttr(resourceName, "version", "1.0.0"),
					resource.TestCheckResourceAttr(resourceName, "snapshot", "false"),
					resource.TestCheckResourceAttr(resourceName, "removed", "false"),
					resource.TestCheckTypeSetElemAttr(resourceName, "platform.*", "darwin"),
					resource.TestCheckTypeSetElemAttr(resourceName, "platform.*", "linux"),
					resource.TestCheckResourceAttr(resourceName, "ecs_mapping.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "ecs_mapping.process.name.field", "cmdline"),
					resource.TestCheckResourceAttr(resourceName, "ecs_mapping.event.category.value", "process"),
					resource.TestCheckResourceAttr(resourceName, "ecs_mapping.event.type.values.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "ecs_mapping.event.type.values.*", "end"),
					resource.TestCheckTypeSetElemAttr(resourceName, "ecs_mapping.event.type.values.*", "start"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ResourceName:             resourceName,
				ImportState:              true,
				ImportStateVerify:        true,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"saved_query_id": config.StringVariable(savedQueryID),
					"space_id":       config.StringVariable(spaceID),
				},
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"saved_query_id": config.StringVariable(savedQueryID),
					"space_id":       config.StringVariable(spaceID),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "query", "SELECT pid, name FROM processes LIMIT 10;"),
					resource.TestCheckResourceAttr(resourceName, "description", "Terraform acceptance update"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update_clear_optionals"),
				ConfigVariables: config.Variables{
					"saved_query_id": config.StringVariable(savedQueryID),
					"space_id":       config.StringVariable(spaceID),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "query", "SELECT pid, name FROM processes LIMIT 10;"),
					resource.TestCheckNoResourceAttr(resourceName, "description"),
					resource.TestCheckNoResourceAttr(resourceName, "platform"),
					resource.TestCheckNoResourceAttr(resourceName, "interval"),
					resource.TestCheckNoResourceAttr(resourceName, "version"),
					resource.TestCheckNoResourceAttr(resourceName, "ecs_mapping"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update_clear_optionals"),
				ConfigVariables: config.Variables{
					"saved_query_id": config.StringVariable(savedQueryID),
					"space_id":       config.StringVariable(spaceID),
				},
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func TestAccResourceOsquerySavedQuery_ExplicitSavedQueryID(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minOsquerySavedQueryVersion, versionutils.FlavorAny)

	firstID := "tf-osquery-first-" + uuid.New().String()
	secondID := "tf-osquery-second-" + uuid.New().String()
	var initialCompositeID string

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceOsquerySavedQueryDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"saved_query_id": config.StringVariable(firstID),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "saved_query_id", firstID),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources[resourceName]
						if !ok {
							return fmt.Errorf("resource not found: %s", resourceName)
						}
						initialCompositeID = rs.Primary.ID
						return nil
					},
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("replace"),
				ConfigVariables: config.Variables{
					"saved_query_id": config.StringVariable(secondID),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "saved_query_id", secondID),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources[resourceName]
						if !ok {
							return fmt.Errorf("resource not found: %s", resourceName)
						}
						if rs.Primary.ID == initialCompositeID {
							return fmt.Errorf("expected new composite id after saved_query_id change (RequiresReplace), got same id: %s", initialCompositeID)
						}
						return nil
					},
				),
			},
		},
	})
}

func TestAccResourceOsquerySavedQuery_Validation(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minOsquerySavedQueryVersion, versionutils.FlavorAny)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("missing_saved_query_id"),
				ExpectError:              regexp.MustCompile(`(?s)(saved_query_id.*required|The argument "saved_query_id" is required)`),
				PlanOnly:                 true,
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("ecs_mapping_two_fields"),
				ConfigVariables: config.Variables{
					"saved_query_id": config.StringVariable("tf-osquery-validation-" + uuid.New().String()),
				},
				ExpectError: regexp.MustCompile(`Exactly one of.*field.*value.*values`),
				PlanOnly:    true,
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("ecs_mapping_empty_element"),
				ConfigVariables: config.Variables{
					"saved_query_id": config.StringVariable("tf-osquery-validation-" + uuid.New().String()),
				},
				ExpectError: regexp.MustCompile(`Exactly one of.*field.*value.*values`),
				PlanOnly:    true,
			},
		},
	})
}

func TestAccResourceOsquerySavedQuery_Platform(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minOsquerySavedQueryVersion, versionutils.FlavorAny)

	savedQueryID := "tf-osquery-platform-" + uuid.New().String()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceOsquerySavedQueryDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"saved_query_id": config.StringVariable(savedQueryID),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "platform.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "platform.*", "darwin"),
					resource.TestCheckTypeSetElemAttr(resourceName, "platform.*", "linux"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"saved_query_id": config.StringVariable(savedQueryID),
				},
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func TestAccResourceOsquerySavedQuery_Space(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minOsquerySavedQueryVersion, versionutils.FlavorAny)

	spaceID := "tf-osquery-space-" + uuid.New().String()[:8]
	savedQueryID := "tf-osquery-space-query-" + uuid.New().String()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceOsquerySavedQueryDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("space_create"),
				ConfigVariables: config.Variables{
					"space_id":       config.StringVariable(spaceID),
					"saved_query_id": config.StringVariable(savedQueryID),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_space.test", "space_id", spaceID),
					resource.TestCheckResourceAttr(resourceName, "space_id", spaceID),
					resource.TestCheckResourceAttr(resourceName, "saved_query_id", savedQueryID),
					resource.TestCheckResourceAttr(resourceName, "id", fmt.Sprintf("%s/%s", spaceID, savedQueryID)),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ResourceName:             resourceName,
				ImportState:              true,
				ImportStateVerify:        true,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("space_create"),
				ConfigVariables: config.Variables{
					"space_id":       config.StringVariable(spaceID),
					"saved_query_id": config.StringVariable(savedQueryID),
				},
			},
		},
	})
}

func TestAccResourceOsquerySavedQuery_PrebuiltImport(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minOsquerySavedQueryVersion, versionutils.FlavorAny)

	prebuiltID, ok := discoverPrebuiltSavedQueryID(t)
	if !ok {
		t.Skip("no prebuilt osquery saved query found in test environment")
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("prebuilt_import"),
				ConfigVariables: config.Variables{
					"saved_query_id": config.StringVariable(prebuiltID),
				},
				ResourceName:  resourceName,
				ImportState:   true,
				ImportStateId: fmt.Sprintf("%s/%s", clients.DefaultSpaceID, prebuiltID),
				ExpectError:   regexp.MustCompile("Prebuilt Osquery saved query"),
			},
		},
	})
}

func TestAccDataSourceOsquerySavedQuery(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minOsquerySavedQueryVersion, versionutils.FlavorAny)

	savedQueryID := "tf-osquery-ds-" + uuid.New().String()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceOsquerySavedQueryDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				ConfigVariables: config.Variables{
					"saved_query_id": config.StringVariable(savedQueryID),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "saved_query_id", savedQueryID),
					resource.TestCheckResourceAttr(dataSourceName, "saved_query_id", savedQueryID),
					resource.TestCheckResourceAttrPair(dataSourceName, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "query", resourceName, "query"),
					resource.TestCheckResourceAttrPair(dataSourceName, "description", resourceName, "description"),
					resource.TestCheckResourceAttrPair(dataSourceName, "interval", resourceName, "interval"),
					resource.TestCheckResourceAttrPair(dataSourceName, "version", resourceName, "version"),
					resource.TestCheckResourceAttr(dataSourceName, "prebuilt", "false"),
				),
			},
		},
	})
}

func TestAccDataSourceOsquerySavedQuery_Prebuilt(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minOsquerySavedQueryVersion, versionutils.FlavorAny)

	prebuiltID, ok := discoverPrebuiltSavedQueryID(t)
	if !ok {
		t.Skip("no prebuilt osquery saved query found in test environment")
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("prebuilt_read"),
				ConfigVariables: config.Variables{
					"saved_query_id": config.StringVariable(prebuiltID),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "saved_query_id", prebuiltID),
					resource.TestCheckResourceAttr(dataSourceName, "prebuilt", "true"),
					resource.TestCheckResourceAttrSet(dataSourceName, "query"),
				),
			},
		},
	})
}

var checkResourceOsquerySavedQueryDestroy = checks.KibanaResourceDestroyCheckCompositeID(
	"elasticstack_kibana_osquery_saved_query",
	func(ctx context.Context, client *kibanaoapi.Client, spaceID, savedQueryID string) (bool, error) {
		entity, diags := kibanaoapi.GetOsquerySavedQuery(ctx, client, spaceID, savedQueryID)
		if diags.HasError() {
			return false, fmt.Errorf("failed to check osquery saved query %q in space %q: %v", savedQueryID, spaceID, diags)
		}
		return entity != nil, nil
	},
)

func discoverPrebuiltSavedQueryID(t *testing.T) (string, bool) {
	t.Helper()
	if os.Getenv("TF_ACC") == "" {
		return "", false
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client, err := clients.NewAcceptanceTestingKibanaScopedClient()
	if err != nil {
		t.Skipf("skipping prebuilt query discovery: %v", err)
	}

	pageSize := kbapi.SecurityOsqueryAPIPageSizeOrUndefined(100)
	resp, err := client.GetKibanaOapiClient().API.OsqueryFindSavedQueriesWithResponse(ctx, &kbapi.OsqueryFindSavedQueriesParams{
		PageSize: &pageSize,
	})
	if err != nil {
		t.Skipf("skipping prebuilt query discovery: %v", err)
	}
	if resp.StatusCode() != http.StatusOK || resp.JSON200 == nil {
		t.Skipf("skipping prebuilt query discovery: unexpected list response status=%d", resp.StatusCode())
	}

	for _, item := range resp.JSON200.Data {
		if item.Prebuilt != nil && *item.Prebuilt && item.Id != "" {
			return item.Id, true
		}
	}

	return "", false
}
