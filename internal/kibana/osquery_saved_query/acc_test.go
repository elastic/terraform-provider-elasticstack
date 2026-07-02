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
	"strings"
	"testing"
	"time"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/acctest/checks"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	kibanaoapi "github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	osquerysavedquery "github.com/elastic/terraform-provider-elasticstack/internal/kibana/osquery_saved_query"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

const (
	resourceName   = "elasticstack_kibana_osquery_saved_query.test"
	dataSourceName = "data.elasticstack_kibana_osquery_saved_query.test"

	osqueryPrebuiltSavedQueryIDEnvVar = "TF_OSQUERY_PREBUILT_SAVED_QUERY_ID"
)

var (
	importStateVerifyIgnore = []string{"snapshot", "removed"}
)

func TestAccResourceOsquerySavedQuery(t *testing.T) {
	versionutils.SkipIfUnsupported(t, osquerysavedquery.MinSupportedVersion, versionutils.FlavorAny)
	skipIfOsquerySavedQueryAPIUnavailable(t)

	savedQueryID := "tf-osquery-" + uuid.New().String()
	spaceID := clients.DefaultSpaceID
	var initialSavedObjectID string

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
					resource.TestCheckResourceAttrSet(resourceName, "saved_object_id"),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources[resourceName]
						if !ok {
							return fmt.Errorf("resource not found: %s", resourceName)
						}
						initialSavedObjectID = rs.Primary.Attributes["saved_object_id"]
						return nil
					},
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
				ImportStateVerifyIgnore:  importStateVerifyIgnore,
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
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update_interval_version"),
				ConfigVariables: config.Variables{
					"saved_query_id": config.StringVariable(savedQueryID),
					"space_id":       config.StringVariable(spaceID),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "interval", "7200"),
					resource.TestCheckResourceAttr(resourceName, "version", "2.0.0"),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources[resourceName]
						if !ok {
							return fmt.Errorf("resource not found: %s", resourceName)
						}
						got := rs.Primary.Attributes["saved_object_id"]
						if got == "" {
							return fmt.Errorf("saved_object_id is empty after update")
						}
						if initialSavedObjectID == "" {
							return fmt.Errorf("initial saved_object_id was not captured")
						}
						if got != initialSavedObjectID {
							return fmt.Errorf("saved_object_id changed unexpectedly: was %q, now %q", initialSavedObjectID, got)
						}
						return nil
					},
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update_snapshot_removed"),
				ConfigVariables: config.Variables{
					"saved_query_id": config.StringVariable(savedQueryID),
					"space_id":       config.StringVariable(spaceID),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "snapshot", "true"),
					resource.TestCheckResourceAttr(resourceName, "removed", "true"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update_ecs_mapping"),
				ConfigVariables: config.Variables{
					"saved_query_id": config.StringVariable(savedQueryID),
					"space_id":       config.StringVariable(spaceID),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "ecs_mapping.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "ecs_mapping.host.name.field", "hostname"),
					resource.TestCheckResourceAttr(resourceName, "ecs_mapping.event.kind.value", "event"),
					resource.TestCheckResourceAttr(resourceName, "ecs_mapping.event.outcome.values.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "ecs_mapping.event.outcome.values.*", "failure"),
					resource.TestCheckTypeSetElemAttr(resourceName, "ecs_mapping.event.outcome.values.*", "success"),
					resource.TestCheckNoResourceAttr(resourceName, "ecs_mapping.process.name"),
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
	versionutils.SkipIfUnsupported(t, osquerysavedquery.MinSupportedVersion, versionutils.FlavorAny)
	skipIfOsquerySavedQueryAPIUnavailable(t)

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
				ConfigDirectory:          acctest.NamedTestCaseDirectory("missing_query"),
				ConfigVariables: config.Variables{
					"saved_query_id": config.StringVariable("tf-osquery-validation-" + uuid.New().String()),
				},
				ExpectError: regexp.MustCompile(`(?s)(query.*required|The argument "query" is required)`),
				PlanOnly:    true,
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("invalid_platform"),
				ConfigVariables: config.Variables{
					"saved_query_id": config.StringVariable("tf-osquery-validation-" + uuid.New().String()),
				},
				ExpectError: regexp.MustCompile(`(?s)(platform|ios|must be one of)`),
				PlanOnly:    true,
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
	versionutils.SkipIfUnsupported(t, osquerysavedquery.MinSupportedVersion, versionutils.FlavorAny)
	skipIfOsquerySavedQueryAPIUnavailable(t)

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
	versionutils.SkipIfUnsupported(t, osquerysavedquery.MinSupportedVersion, versionutils.FlavorAny)
	skipIfOsquerySavedQueryAPIUnavailable(t)

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
				ImportStateVerifyIgnore:  importStateVerifyIgnore,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("space_create"),
				ConfigVariables: config.Variables{
					"space_id":       config.StringVariable(spaceID),
					"saved_query_id": config.StringVariable(savedQueryID),
				},
			},
		},
	})
}

func TestAccResourceOsquerySavedQuery_SpaceRequiresReplace(t *testing.T) {
	versionutils.SkipIfUnsupported(t, osquerysavedquery.MinSupportedVersion, versionutils.FlavorAny)
	skipIfOsquerySavedQueryAPIUnavailable(t)

	spaceID1 := "tf-oq-sp1-" + uuid.New().String()[:8]
	spaceID2 := "tf-oq-sp2-" + uuid.New().String()[:8]
	savedQueryID := "tf-oq-space-replace-" + uuid.New().String()
	var firstCompositeID string

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceOsquerySavedQueryDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("space_create"),
				ConfigVariables: config.Variables{
					"space_id":       config.StringVariable(spaceID1),
					"saved_query_id": config.StringVariable(savedQueryID),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "space_id", spaceID1),
					resource.TestCheckResourceAttr(resourceName, "id", fmt.Sprintf("%s/%s", spaceID1, savedQueryID)),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources[resourceName]
						if !ok {
							return fmt.Errorf("resource not found: %s", resourceName)
						}
						firstCompositeID = rs.Primary.ID
						return nil
					},
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("space_replace"),
				ConfigVariables: config.Variables{
					"space_id":       config.StringVariable(spaceID2),
					"saved_query_id": config.StringVariable(savedQueryID),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "space_id", spaceID2),
					resource.TestCheckResourceAttr(resourceName, "id", fmt.Sprintf("%s/%s", spaceID2, savedQueryID)),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources[resourceName]
						if !ok {
							return fmt.Errorf("resource not found: %s", resourceName)
						}
						if rs.Primary.ID == firstCompositeID {
							return fmt.Errorf("expected new composite id after space_id change (RequiresReplace), got same id: %s", firstCompositeID)
						}
						return nil
					},
					checkOsquerySavedQueryNotFound(spaceID1, savedQueryID),
				),
			},
		},
	})
}

func TestAccResourceOsquerySavedQuery_PrebuiltImport(t *testing.T) {
	versionutils.SkipIfUnsupported(t, osquerysavedquery.MinSupportedVersion, versionutils.FlavorAny)

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
	versionutils.SkipIfUnsupported(t, osquerysavedquery.MinSupportedVersion, versionutils.FlavorAny)
	skipIfOsquerySavedQueryAPIUnavailable(t)

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
					resource.TestCheckResourceAttr(dataSourceName, "space_id", clients.DefaultSpaceID),
					resource.TestCheckResourceAttrPair(dataSourceName, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "saved_object_id", resourceName, "saved_object_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "query", resourceName, "query"),
					resource.TestCheckResourceAttrPair(dataSourceName, "description", resourceName, "description"),
					resource.TestCheckResourceAttrPair(dataSourceName, "interval", resourceName, "interval"),
					resource.TestCheckResourceAttrPair(dataSourceName, "version", resourceName, "version"),
					resource.TestCheckResourceAttrPair(dataSourceName, "snapshot", resourceName, "snapshot"),
					resource.TestCheckResourceAttrPair(dataSourceName, "removed", resourceName, "removed"),
					resource.TestCheckResourceAttr(dataSourceName, "prebuilt", "false"),
					resource.TestCheckTypeSetElemAttr(dataSourceName, "platform.*", "darwin"),
					resource.TestCheckTypeSetElemAttr(dataSourceName, "platform.*", "linux"),
					resource.TestCheckResourceAttr(dataSourceName, "ecs_mapping.%", "3"),
					resource.TestCheckResourceAttr(dataSourceName, "ecs_mapping.process.name.field", "cmdline"),
					resource.TestCheckResourceAttr(dataSourceName, "ecs_mapping.event.category.value", "process"),
					resource.TestCheckResourceAttr(dataSourceName, "ecs_mapping.event.type.values.#", "2"),
					resource.TestCheckTypeSetElemAttr(dataSourceName, "ecs_mapping.event.type.values.*", "end"),
					resource.TestCheckTypeSetElemAttr(dataSourceName, "ecs_mapping.event.type.values.*", "start"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("minimal_read"),
				ConfigVariables: config.Variables{
					"saved_query_id": config.StringVariable(savedQueryID),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "ecs_mapping.%", "0"),
					resource.TestCheckNoResourceAttr(dataSourceName, "platform"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("snapshot_removed_true"),
				ConfigVariables: config.Variables{
					"saved_query_id": config.StringVariable(savedQueryID),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "snapshot", "true"),
					resource.TestCheckResourceAttr(dataSourceName, "removed", "true"),
				),
			},
		},
	})
}

func TestAccDataSourceOsquerySavedQuery_Validation(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("missing_saved_query_id"),
				ExpectError:              regexp.MustCompile(`(?s)(saved_query_id.*required|The argument "saved_query_id" is required)`),
				PlanOnly:                 true,
			},
		},
	})
}

func TestAccDataSourceOsquerySavedQuery_NotFound(t *testing.T) {
	versionutils.SkipIfUnsupported(t, osquerysavedquery.MinSupportedVersion, versionutils.FlavorAny)
	skipIfOsquerySavedQueryAPIUnavailable(t)

	missingID := "tf-osquery-missing-" + uuid.New().String()

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("not_found"),
				ConfigVariables: config.Variables{
					"saved_query_id": config.StringVariable(missingID),
				},
				ExpectError: regexp.MustCompile(`Osquery saved query not found`),
			},
		},
	})
}

func TestAccDataSourceOsquerySavedQuery_Prebuilt(t *testing.T) {
	versionutils.SkipIfUnsupported(t, osquerysavedquery.MinSupportedVersion, versionutils.FlavorAny)

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
					resource.TestCheckResourceAttr(dataSourceName, "space_id", clients.DefaultSpaceID),
					resource.TestCheckResourceAttrSet(dataSourceName, "id"),
					resource.TestCheckResourceAttrSet(dataSourceName, "saved_object_id"),
					resource.TestCheckResourceAttr(dataSourceName, "prebuilt", "true"),
					resource.TestCheckResourceAttrSet(dataSourceName, "query"),
					resource.TestCheckResourceAttrSet(dataSourceName, "description"),
					resource.TestCheckResourceAttrSet(dataSourceName, "snapshot"),
					resource.TestCheckResourceAttrSet(dataSourceName, "removed"),
				),
			},
		},
	})
}

func TestAccDataSourceOsquerySavedQuery_Space(t *testing.T) {
	versionutils.SkipIfUnsupported(t, osquerysavedquery.MinSupportedVersion, versionutils.FlavorAny)
	skipIfOsquerySavedQueryAPIUnavailable(t)

	spaceID := "tf-oq-ds-space-" + uuid.New().String()[:8]
	savedQueryID := "tf-oq-ds-" + uuid.New().String()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceOsquerySavedQueryDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("space_read"),
				ConfigVariables: config.Variables{
					"space_id":       config.StringVariable(spaceID),
					"saved_query_id": config.StringVariable(savedQueryID),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "space_id", spaceID),
					resource.TestCheckResourceAttrPair(dataSourceName, "query", resourceName, "query"),
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
			if osquerySavedQueryAbsenceCheckCanTreatAsAbsent(diags) {
				return false, nil
			}
			return false, fmt.Errorf("failed to check osquery saved query %q in space %q: %v", savedQueryID, spaceID, diags)
		}
		return entity != nil, nil
	},
)

func checkOsquerySavedQueryNotFound(spaceID, savedQueryID string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		client, err := clients.NewAcceptanceTestingKibanaScopedClient()
		if err != nil {
			return err
		}

		entity, diags := kibanaoapi.GetOsquerySavedQuery(context.Background(), client.GetKibanaOapiClient(), spaceID, savedQueryID)
		if diags.HasError() {
			if osquerySavedQueryAbsenceCheckCanTreatAsAbsent(diags) {
				return nil
			}
			return fmt.Errorf("failed to verify osquery saved query %q in space %q: %v", savedQueryID, spaceID, diags)
		}
		if entity != nil {
			return fmt.Errorf("expected osquery saved query %q to be absent from space %q after RequiresReplace", savedQueryID, spaceID)
		}
		return nil
	}
}

func osquerySavedQueryAbsenceCheckCanTreatAsAbsent(diags diag.Diagnostics) bool {
	for _, d := range diags.Errors() {
		if strings.Contains(d.Detail(), `"statusCode":500`) {
			return true
		}
	}

	return false
}

func skipIfOsquerySavedQueryAPIUnavailable(t *testing.T) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client, err := clients.NewAcceptanceTestingKibanaScopedClient()
	if err != nil {
		t.Skipf("skipping Osquery saved query acceptance tests: %v", err)
	}

	oapiClient := client.GetKibanaOapiClient()
	spaceID := clients.DefaultSpaceID

	page := 1
	pageSize := kbapi.SecurityOsqueryAPIPageSizeOrUndefined(1)
	resp, err := oapiClient.API.OsqueryFindSavedQueriesWithResponse(ctx, &kbapi.OsqueryFindSavedQueriesParams{
		Page:     &page,
		PageSize: &pageSize,
	})
	if err != nil {
		t.Skipf("skipping Osquery saved query acceptance tests: %v", err)
	}
	if resp.StatusCode() == http.StatusInternalServerError {
		t.Skipf("skipping Osquery saved query acceptance tests: API unavailable in this stack: %s", string(resp.Body))
	}
	if resp.StatusCode() != http.StatusOK {
		t.Fatalf("unexpected Osquery saved query availability response status=%d body=%s", resp.StatusCode(), string(resp.Body))
	}

	// Some matrix stacks return 200 on find but cannot serve create/read round-trips.
	savedQueryID := "tf-osquery-probe-" + uuid.New().String()
	query := "SELECT 1;"
	interval := "3600"
	entity, createDiags := kibanaoapi.CreateOsquerySavedQuery(ctx, oapiClient, spaceID, kbapi.OsqueryCreateSavedQueryJSONRequestBody{
		Id:       &savedQueryID,
		Query:    &query,
		Interval: &interval,
	})
	if createDiags.HasError() {
		t.Skipf("skipping Osquery saved query acceptance tests: create probe failed: %v", createDiags)
	}
	if entity == nil || entity.SavedObjectID == "" {
		t.Skip("skipping Osquery saved query acceptance tests: create probe returned no saved_object_id")
	}
	defer func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cleanupCancel()
		_ = kibanaoapi.DeleteOsquerySavedQueryBySavedObjectID(cleanupCtx, oapiClient, spaceID, entity.SavedObjectID)
	}()

	readEntity, readDiags := kibanaoapi.GetOsquerySavedQueryBySavedObjectID(ctx, oapiClient, spaceID, entity.SavedObjectID)
	if readDiags.HasError() {
		t.Skipf("skipping Osquery saved query acceptance tests: read probe failed: %v", readDiags)
	}
	if readEntity == nil {
		t.Skip("skipping Osquery saved query acceptance tests: read probe returned not found after create")
	}
}

func discoverPrebuiltSavedQueryID(t *testing.T) (string, bool) {
	t.Helper()
	if os.Getenv("TF_ACC") == "" {
		return "", false
	}

	if override := os.Getenv(osqueryPrebuiltSavedQueryIDEnvVar); override != "" {
		return override, true
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client, err := clients.NewAcceptanceTestingKibanaScopedClient()
	if err != nil {
		t.Skipf("skipping prebuilt query discovery: %v", err)
	}

	pageSize := kbapi.SecurityOsqueryAPIPageSizeOrUndefined(100)
	page := 1

	for {
		pageParam := page
		resp, err := client.GetKibanaOapiClient().API.OsqueryFindSavedQueriesWithResponse(ctx, &kbapi.OsqueryFindSavedQueriesParams{
			Page:     &pageParam,
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

		if len(resp.JSON200.Data) == 0 || page*pageSize >= resp.JSON200.Total {
			break
		}
		page++
	}

	return "", false
}
