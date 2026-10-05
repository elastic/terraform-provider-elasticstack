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
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
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

func TestAccResourceMLCalendarEvent_forceReplaceOnChange(t *testing.T) {
	calendarID := fmt.Sprintf("test-cal-evt-rep-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum))
	resourceName := "elasticstack_elasticsearch_ml_calendar_event.test"
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
					resource.TestCheckResourceAttr(resourceName, "description", "Replace test initial description"),
					resource.TestCheckResourceAttrSet(resourceName, "event_id"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables:          vars,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionReplace),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "description", "Replace test changed description"),
					resource.TestCheckResourceAttrSet(resourceName, "event_id"),
				),
			},
		},
	})
}

// TestAccResourceMLCalendarEvent_optionalSchedulingFieldsUnsupportedVersion asserts that
// setting force_time_shift against an Elasticsearch cluster older than
// mlCalendarEventOptionalSchedulingMinElasticsearch fails with the version-gate error raised
// by createCalendarEvent. It runs only on older stateful clusters.
func TestAccResourceMLCalendarEvent_optionalSchedulingFieldsUnsupportedVersion(t *testing.T) {
	isStateful, err := versionutils.CheckIfNotServerless()()
	if err != nil {
		t.Fatalf("failed to check whether stack is serverless: %v", err)
	}
	if !isStateful {
		t.Skip("serverless supports optional scheduling fields; skipping version gate test")
	}

	unsupported, err := versionutils.CheckIfVersionIsUnsupported(mlCalendarEventOptionalSchedulingMinElasticsearch)()
	if err != nil {
		t.Fatalf("failed to check stack version: %v", err)
	}
	if !unsupported {
		t.Skip("stack supports optional scheduling fields; skipping version gate test")
	}

	calendarID := fmt.Sprintf("test-cal-evt-ver-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum))

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"calendar_id": config.StringVariable(calendarID),
				},
				ExpectError: regexp.MustCompile(`optional scheduling fields not supported`),
			},
		},
	})
}

// TestAccResourceMLCalendarEvent_validation_emptyForceTimeShift asserts that force_time_shift
// rejects an empty string via its stringvalidator.LengthAtLeast(1) validator.
func TestAccResourceMLCalendarEvent_validation_emptyForceTimeShift(t *testing.T) {
	calendarID := fmt.Sprintf("test-cal-evt-fts-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum))

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
				ExpectError: regexp.MustCompile(`(?i)(string length must be at least|force_time_shift)`),
			},
		},
	})
}

// TestAccResourceMLCalendarEvent_nonZOffsetTimes confirms RFC3339 start_time/end_time values
// using an explicit non-Z UTC offset round-trip through create and refresh.
func TestAccResourceMLCalendarEvent_nonZOffsetTimes(t *testing.T) {
	calendarID := fmt.Sprintf("test-cal-evt-off-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum))
	resourceName := "elasticstack_elasticsearch_ml_calendar_event.test"

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
					resource.TestCheckResourceAttr(resourceName, "start_time", "2027-03-01T02:00:00+02:00"),
					resource.TestCheckResourceAttr(resourceName, "end_time", "2027-03-01T04:30:00+02:00"),
					resource.TestCheckResourceAttrSet(resourceName, "event_id"),
				),
			},
		},
	})
}

// TestAccResourceMLCalendarEvent_validation_malformedStartTime asserts that a malformed
// (non-RFC3339) start_time string produces a clear validation error at plan time.
func TestAccResourceMLCalendarEvent_validation_malformedStartTime(t *testing.T) {
	calendarID := fmt.Sprintf("test-cal-evt-mal-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum))

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
				ExpectError: regexp.MustCompile(`(?i)(RFC.?3339|invalid|not a valid)`),
			},
		},
	})
}

// TestAccResourceMLCalendarEvent_optionalSchedulingFieldsPartial sets only force_time_shift
// while leaving skip_result and skip_model_update omitted, confirming the raw-POST-body path
// and version gate trigger correctly for partial combinations and that the omitted fields
// remain server-populated (computed).
func TestAccResourceMLCalendarEvent_optionalSchedulingFieldsPartial(t *testing.T) {
	versionutils.SkipIfUnsupported(t, mlCalendarEventOptionalSchedulingMinElasticsearch, versionutils.FlavorAny)
	calendarID := fmt.Sprintf("test-cal-evt-par-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum))
	resourceName := "elasticstack_elasticsearch_ml_calendar_event.test"

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
					resource.TestCheckResourceAttr(resourceName, "force_time_shift", "1800"),
					resource.TestCheckResourceAttrSet(resourceName, "skip_result"),
					resource.TestCheckResourceAttrSet(resourceName, "skip_model_update"),
					resource.TestCheckResourceAttrSet(resourceName, "event_id"),
				),
			},
		},
	})
}
