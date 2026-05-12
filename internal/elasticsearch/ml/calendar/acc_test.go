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

package calendar_test

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccResourceMLCalendar(t *testing.T) {
	calendarID := fmt.Sprintf("test-calendar-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum))
	jobID := fmt.Sprintf("test-cal-job-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum))
	jobID2 := fmt.Sprintf("test-cal-job2-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum))

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"calendar_id": config.StringVariable(calendarID),
					"job_id":      config.StringVariable(jobID),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_calendar.test", "calendar_id", calendarID),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_calendar.test", "description", "Test calendar"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_calendar.test", "job_ids.#", "1"),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_ml_calendar.test", "id"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"calendar_id": config.StringVariable(calendarID),
					"job_id":      config.StringVariable(jobID),
					"job_id_2":    config.StringVariable(jobID2),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_calendar.test", "calendar_id", calendarID),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_calendar.test", "description", "Test calendar"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_calendar.test", "job_ids.#", "2"),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_ml_calendar.test", "id"),
				),
			},
		},
	})
}

func TestAccResourceMLCalendarNoJobs(t *testing.T) {
	calendarID := fmt.Sprintf("test-cal-nojobs-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum))

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"calendar_id": config.StringVariable(calendarID),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_calendar.test", "calendar_id", calendarID),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_calendar.test", "description", "Calendar with no jobs"),
					resource.TestCheckNoResourceAttr("elasticstack_elasticsearch_ml_calendar.test", "job_ids.#"),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_ml_calendar.test", "id"),
				),
			},
		},
	})
}

func TestAccResourceMLCalendarImport(t *testing.T) {
	calendarID := fmt.Sprintf("test-cal-import-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum))

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"calendar_id": config.StringVariable(calendarID),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_calendar.test", "calendar_id", calendarID),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_calendar.test", "description", "Calendar for import test"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ResourceName:             "elasticstack_elasticsearch_ml_calendar.test",
				ImportState:              true,
				ImportStateVerify:        true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs := s.RootModule().Resources["elasticstack_elasticsearch_ml_calendar.test"]
					return rs.Primary.ID, nil
				},
				ConfigVariables: config.Variables{
					"calendar_id": config.StringVariable(calendarID),
				},
			},
		},
	})
}

func TestAccResourceMLCalendar_validation_invalidCalendarIDRegex(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				Config: `
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_ml_calendar" "bad" {
  calendar_id = "INVALID_UPPER"
  description = "x"
}
`,
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`(?i)(calendar_id|invalid|match|lowercase|alphanumeric)`),
			},
		},
	})
}

func TestAccResourceMLCalendar_validation_calendarIDTooLong(t *testing.T) {
	longID := strings.Repeat("a", 65)
	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				Config: fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_ml_calendar" "bad" {
  calendar_id = %q
  description = "x"
}
`, longID),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`(?i)(calendar_id|64|length)`),
			},
		},
	})
}

func TestAccResourceMLCalendar_importWrongIDFormat(t *testing.T) {
	calendarID := fmt.Sprintf("test-cal-badimp-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum))
	cfg := fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_ml_calendar" "test" {
  calendar_id = %q
  description   = "Calendar for import test"
}
`, calendarID)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				Config:                   cfg,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_ml_calendar.test", "id"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				Config:                   cfg,
				ResourceName:             "elasticstack_elasticsearch_ml_calendar.test",
				ImportState:              true,
				// Default ImportStatePersist=false runs import in a temp working dir; the harness
				// still replaces the main dir's config with provider stubs first, so post-test
				// destroy would lose elasticsearch configuration. Persist keeps import on the
				// main working dir so the full config (and env-based endpoints) remain for destroy.
				ImportStatePersist: true,
				ImportStateId:      "not-a-valid-composite-id",
				ExpectError:        regexp.MustCompile(`Wrong resource ID`),
			},
		},
	})
}
