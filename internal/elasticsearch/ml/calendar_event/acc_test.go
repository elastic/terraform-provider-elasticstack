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

package calendar_event_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// mlCalendarEventOptionalSchedulingMinElasticsearch is the minimum Elasticsearch version for the
// post calendar events API fields skip_result, skip_model_update, and force_time_shift (see ES #112837).
// TestAccResourceMLCalendarEvent_optionalSchedulingFields calls SkipIfUnsupported against this
// version so acceptance runs on stacks older than 8.16 skip that test; other calendar event acc
// tests use only fields supported on the provider's minimum supported Elasticsearch version.
var mlCalendarEventOptionalSchedulingMinElasticsearch = version.Must(version.NewVersion("8.16.0"))

func TestAccResourceMLCalendarEvent(t *testing.T) {
	// The Check block asserts skip_result and skip_model_update are populated by the
	// server, which only happens on Elasticsearch 8.16+ (see ES #112837 / backport #113209).
	versionutils.SkipIfUnsupported(t, mlCalendarEventOptionalSchedulingMinElasticsearch, versionutils.FlavorAny)
	calendarID := fmt.Sprintf("test-cal-evt-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum))

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
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_calendar_event.test", "calendar_id", calendarID),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_calendar_event.test", "description", "Planned maintenance"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_calendar_event.test", "start_time", "2026-06-01T00:00:00Z"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_calendar_event.test", "end_time", "2026-06-01T06:00:00Z"),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_ml_calendar_event.test", "event_id"),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_ml_calendar_event.test", "id"),
					// Gap 1: confirm force_time_shift is absent when not configured.
					resource.TestCheckNoResourceAttr("elasticstack_elasticsearch_ml_calendar_event.test", "force_time_shift"),
					// Gaps 2 & 3: skip_result and skip_model_update are Optional+Computed; verify
					// the server populates them even when omitted from the configuration.
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_ml_calendar_event.test", "skip_result"),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_ml_calendar_event.test", "skip_model_update"),
				),
			},
		},
	})
}

func TestAccResourceMLCalendarEvent_optionalSchedulingFields(t *testing.T) {
	versionutils.SkipIfUnsupported(t, mlCalendarEventOptionalSchedulingMinElasticsearch, versionutils.FlavorAny)
	calendarID := fmt.Sprintf("test-cal-evt-opt-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum))
	vars := config.Variables{
		"calendar_id": config.StringVariable(calendarID),
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          vars,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_calendar_event.test", "calendar_id", calendarID),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_calendar_event.test", "description", "ACC outage with optional scheduling fields"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_calendar_event.test", "start_time", "2026-09-01T00:00:00Z"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_calendar_event.test", "end_time", "2026-09-01T02:00:00Z"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_calendar_event.test", "skip_result", "true"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_calendar_event.test", "skip_model_update", "true"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_calendar_event.test", "force_time_shift", "3600"),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_ml_calendar_event.test", "event_id"),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_ml_calendar_event.test", "id"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          vars,
				ResourceName:             "elasticstack_elasticsearch_ml_calendar_event.test",
				ImportState:              true,
				ImportStateVerify:        true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs := s.RootModule().Resources["elasticstack_elasticsearch_ml_calendar_event.test"]
					return rs.Primary.ID, nil
				},
			},
		},
	})
}

func TestAccResourceMLCalendarEventImport(t *testing.T) {
	calendarID := fmt.Sprintf("test-cal-evt-imp-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum))

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
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_calendar_event.test", "calendar_id", calendarID),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_calendar_event.test", "description", "Import test event"),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_ml_calendar_event.test", "event_id"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ResourceName:             "elasticstack_elasticsearch_ml_calendar_event.test",
				ImportState:              true,
				ImportStateVerify:        true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs := s.RootModule().Resources["elasticstack_elasticsearch_ml_calendar_event.test"]
					return rs.Primary.ID, nil
				},
				ConfigVariables: config.Variables{
					"calendar_id": config.StringVariable(calendarID),
				},
			},
		},
	})
}

func TestAccResourceMLCalendarEvent_validation_endBeforeStart(t *testing.T) {
	calendarID := fmt.Sprintf("test-cal-evt-time-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum))

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("plan"),
				ConfigVariables: config.Variables{
					"holder_calendar_id": config.StringVariable(calendarID),
				},
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`(?i)(Invalid event time range|end_time must be after)`),
			},
		},
	})
}

func TestAccResourceMLCalendarEvent_validation_invalidCalendarIDRegex(t *testing.T) {
	calendarID := fmt.Sprintf("test-cal-evt-hold-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum))

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("plan"),
				ConfigVariables: config.Variables{
					"holder_calendar_id": config.StringVariable(calendarID),
				},
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`(?i)(calendar_id|invalid|match|lowercase|alphanumeric)`),
			},
		},
	})
}

func TestAccResourceMLCalendarEvent_optionalSchedulingFieldsFalse(t *testing.T) {
	versionutils.SkipIfUnsupported(t, mlCalendarEventOptionalSchedulingMinElasticsearch, versionutils.FlavorAny)
	calendarID := fmt.Sprintf("test-cal-evt-fls-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum))
	vars := config.Variables{
		"calendar_id": config.StringVariable(calendarID),
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          vars,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_calendar_event.test", "calendar_id", calendarID),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_calendar_event.test", "description", "False scheduling flags test"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_calendar_event.test", "start_time", "2026-11-01T00:00:00Z"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_calendar_event.test", "end_time", "2026-11-01T02:00:00Z"),
					// Gap 4: verify skip_result=false and skip_model_update=false round-trip.
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_calendar_event.test", "skip_result", "false"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_calendar_event.test", "skip_model_update", "false"),
					// Gap 5: second force_time_shift value (7200) confirms different durations are accepted.
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_calendar_event.test", "force_time_shift", "7200"),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_ml_calendar_event.test", "event_id"),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_ml_calendar_event.test", "id"),
				),
			},
		},
	})
}

func TestAccResourceMLCalendarEvent_importWrongIDFormat(t *testing.T) {
	calendarID := fmt.Sprintf("test-cal-evt-badimp-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum))
	importVars := config.Variables{
		"calendar_id": config.StringVariable(calendarID),
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          importVars,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_ml_calendar_event.test", "id"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          importVars,
				ResourceName:             "elasticstack_elasticsearch_ml_calendar_event.test",
				ImportState:              true,
				ImportStateKind:          resource.ImportBlockWithID,
				ImportStateVerify:        false,
				ImportStateId:            "missing-slash-segment",
				ExpectError:              regexp.MustCompile(`Wrong resource ID`),
			},
		},
	})
}
