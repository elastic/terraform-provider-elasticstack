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

package maintenancewindow_test

import (
	"fmt"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/google/uuid"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
)

var minMaintenanceWindowAPISupport = version.Must(version.NewVersion("9.1.0"))

func TestAccResourceMaintenanceWindow(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minMaintenanceWindowAPISupport, versionutils.FlavorAny)

	addr := "elasticstack_kibana_maintenance_window.test_maintenance_window"

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(addr, "id"),
					resource.TestCheckResourceAttr(addr, "space_id", "default"),
					resource.TestCheckResourceAttr(addr, "title", "Terraform Maintenance Window"),
					resource.TestCheckResourceAttr(addr, "enabled", "true"),
					resource.TestCheckResourceAttr(addr, "custom_schedule.start", "1992-01-01T05:00:00.200Z"),
					resource.TestCheckResourceAttr(addr, "custom_schedule.duration", "10d"),
					resource.TestCheckResourceAttr(addr, "custom_schedule.timezone", "UTC"),
					resource.TestCheckResourceAttr(addr, "custom_schedule.recurring.every", "20d"),
					resource.TestCheckResourceAttr(addr, "custom_schedule.recurring.end", "2029-05-17T05:05:00.000Z"),
					resource.TestCheckResourceAttr(addr, "custom_schedule.recurring.on_week_day.0", "MO"),
					resource.TestCheckResourceAttr(addr, "custom_schedule.recurring.on_week_day.1", "TU"),
					resource.TestCheckResourceAttr(addr, "custom_schedule.recurring.on_month_day.#", "0"),
					resource.TestCheckResourceAttr(addr, "custom_schedule.recurring.on_month.#", "0"),
					resource.TestCheckResourceAttr(addr, "scope.alerting.kql", "_id: '1234'"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(addr, "title", "Terraform Maintenance Window UPDATED"),
					resource.TestCheckResourceAttr(addr, "enabled", "false"),
					resource.TestCheckResourceAttr(addr, "custom_schedule.start", "1999-02-02T05:00:00.200Z"),
					resource.TestCheckResourceAttr(addr, "custom_schedule.duration", "12d"),
					resource.TestCheckResourceAttr(addr, "custom_schedule.timezone", "Asia/Taipei"),
					resource.TestCheckResourceAttr(addr, "custom_schedule.recurring.every", "21d"),
					resource.TestCheckNoResourceAttr(addr, "custom_schedule.recurring.end"),
					resource.TestCheckResourceAttr(addr, "custom_schedule.recurring.on_week_day.#", "0"),
					resource.TestCheckResourceAttr(addr, "custom_schedule.recurring.on_month_day.0", "1"),
					resource.TestCheckResourceAttr(addr, "custom_schedule.recurring.on_month_day.1", "2"),
					resource.TestCheckResourceAttr(addr, "custom_schedule.recurring.on_month_day.2", "3"),
					resource.TestCheckResourceAttr(addr, "custom_schedule.recurring.on_month.0", "4"),
					resource.TestCheckResourceAttr(addr, "custom_schedule.recurring.on_month.1", "5"),
					resource.TestCheckResourceAttr(addr, "scope.alerting.kql", "_id: 'foobar'"),
				),
			},
			// ImportState testing: confirms the full resource, including
			// nested custom_schedule and scope attributes, round-trips
			// through import.
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ResourceName:             addr,
				ImportState:              true,
				ImportStateVerify:        true,
			},
			// Removing scope cannot be applied in place (the Kibana API
			// rejects an explicit null for scope on PATCH), so the scope
			// attribute's RequiresReplaceIf forces a destroy-then-create
			// here instead of an update.
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("remove_scope"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(addr, plancheck.ResourceActionDestroyBeforeCreate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(addr, "title", "Terraform Maintenance Window UPDATED"),
					resource.TestCheckNoResourceAttr(addr, "scope.alerting.kql"),
				),
			},
		},
	})
}

// TestAccResourceMaintenanceWindowSpace verifies that the maintenance window
// can be created in a non-default Kibana space, and that changing space_id
// forces replacement as documented by its RequiresReplace plan modifier.
func TestAccResourceMaintenanceWindowSpace(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minMaintenanceWindowAPISupport, versionutils.FlavorAny)

	addr := "elasticstack_kibana_maintenance_window.test_maintenance_window"
	spaceID := fmt.Sprintf("test-space-%s", uuid.New().String()[:8])
	newSpaceID := fmt.Sprintf("test-space-2-%s", uuid.New().String()[:8])

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"space_id": config.StringVariable(spaceID),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(addr, "id"),
					resource.TestCheckResourceAttr(addr, "space_id", spaceID),
					resource.TestCheckResourceAttr(addr, "title", "Terraform Maintenance Window In Space"),
				),
			},
			{
				// Changing space_id must force a destroy-then-create,
				// matching the attribute's documented RequiresReplace plan
				// modifier.
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("change_space"),
				ConfigVariables: config.Variables{
					"space_id":     config.StringVariable(spaceID),
					"new_space_id": config.StringVariable(newSpaceID),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(addr, plancheck.ResourceActionDestroyBeforeCreate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(addr, "space_id", newSpaceID),
					resource.TestCheckResourceAttr(addr, "title", "Terraform Maintenance Window In Space"),
				),
			},
		},
	})
}

// TestAccResourceMaintenanceWindowNoScope verifies that a maintenance window
// can be created without scope, and covers occurrences, nth-day on_week_day
// values, the enabled default (false when omitted), and the computed
// "UTC" default for custom_schedule.timezone when it is omitted.
func TestAccResourceMaintenanceWindowNoScope(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minMaintenanceWindowAPISupport, versionutils.FlavorAny)

	addr := "elasticstack_kibana_maintenance_window.test_maintenance_window"

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(addr, "id"),
					resource.TestCheckResourceAttr(addr, "title", "Terraform Maintenance Window NTH DAY"),
					resource.TestCheckResourceAttr(addr, "enabled", "false"),
					resource.TestCheckResourceAttr(addr, "custom_schedule.timezone", "UTC"),
					resource.TestCheckResourceAttr(addr, "custom_schedule.recurring.every", "1w"),
					resource.TestCheckResourceAttr(addr, "custom_schedule.recurring.occurrences", "5"),
					resource.TestCheckResourceAttr(addr, "custom_schedule.recurring.on_week_day.0", "+1MO"),
					resource.TestCheckResourceAttr(addr, "custom_schedule.recurring.on_week_day.1", "-2FR"),
					resource.TestCheckNoResourceAttr(addr, "scope.alerting.kql"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(addr, "custom_schedule.timezone", "UTC"),
					resource.TestCheckResourceAttr(addr, "custom_schedule.recurring.occurrences", "10"),
					resource.TestCheckResourceAttr(addr, "custom_schedule.recurring.on_week_day.0", "+1MO"),
					resource.TestCheckResourceAttr(addr, "custom_schedule.recurring.on_week_day.1", "-2FR"),
				),
			},
		},
	})
}
