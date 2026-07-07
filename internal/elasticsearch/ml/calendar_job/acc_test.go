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

package calendar_job_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"

	estypes "github.com/elastic/go-elasticsearch/v9/typedapi/types"
	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	esclient "github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

const mlCalendarJobResourceAddr = "elasticstack_elasticsearch_ml_calendar_job.test"

// setupAccMLCalendar creates an ML calendar via the Elasticsearch API so acceptance tests
// do not depend on the elasticstack_elasticsearch_ml_calendar resource (which may not
// be present on all branches). The calendar is deleted in t.Cleanup after the test.
//
// Call from the first test step's PreConfig (after acctest.PreCheck in the test case's
// PreCheck) so no Elasticsearch work runs when acceptance prerequisites are not satisfied.
func setupAccMLCalendar(t *testing.T, calendarID string) {
	t.Helper()
	ctx := context.Background()
	client, err := clients.NewAcceptanceTestingElasticsearchScopedClient()
	if err != nil {
		t.Fatalf("acceptance elasticsearch client: %v", err)
	}
	es := client.GetESClient()
	desc := fmt.Sprintf("terraform acc calendar %s", calendarID)
	if _, err := es.Ml.PutCalendar(calendarID).Description(desc).Do(ctx); err != nil {
		t.Fatalf("put ML calendar %q: %v", calendarID, err)
	}
	t.Cleanup(func() {
		if _, delErr := es.Ml.DeleteCalendar(calendarID).Do(context.Background()); delErr != nil {
			t.Logf("cleanup delete ML calendar %q: %v", calendarID, delErr)
		}
	})
}

func testAccMLCalendarJobCompositeIDRegexp(calendarID, jobID string) *regexp.Regexp {
	return regexp.MustCompile(`^[^/]+/` + regexp.QuoteMeta(calendarID+"/"+jobID) + `$`)
}

func testAccMLCalendarJobESEndpoints() []string {
	raw := os.Getenv("ELASTICSEARCH_ENDPOINTS")
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func testAccCheckMLCalendarJobAbsentFromState(addr string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if _, ok := s.RootModule().Resources[addr]; ok {
			return fmt.Errorf("expected %q to be absent from state after refresh (calendar deleted out-of-band); "+
				"read must treat the missing calendar API error as not found so the resource can be dropped from state", addr)
		}
		return nil
	}
}

func importMLCalendarJobStateID(addr string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[addr]
		if !ok || rs == nil {
			return "", fmt.Errorf("resource %q not found in state", addr)
		}
		return rs.Primary.ID, nil
	}
}

// accCleanupMLAnomalyJobAfterTest registers a best-effort CloseJob + DeleteJob for jobID so
// acceptance runs that expect an apply failure still tear down an ML job if Terraform cannot
// complete destroy (mirrors patterns in other ML acc tests).
func accCleanupMLAnomalyJobAfterTest(t *testing.T, jobID string) {
	t.Helper()
	t.Cleanup(func() {
		ctx := context.Background()
		client, err := clients.NewAcceptanceTestingElasticsearchScopedClient()
		if err != nil {
			t.Logf("cleanup ML job %q: acceptance client: %v", jobID, err)
			return
		}
		es := client.GetESClient()
		if _, err := es.Ml.CloseJob(jobID).Force(true).AllowNoMatch(true).Do(ctx); err != nil {
			t.Logf("cleanup ML job %q: CloseJob: %v", jobID, err)
		}
		_, err = es.Ml.DeleteJob(jobID).Force(true).Do(ctx)
		if err == nil {
			return
		}
		var esErr *estypes.ElasticsearchError
		if errors.As(err, &esErr) && esErr.Status == 404 {
			return
		}
		t.Logf("cleanup ML job %q: DeleteJob: %v", jobID, err)
	})
}

func TestAccResourceMLCalendarJob_withJobGroup(t *testing.T) {
	calendarID := fmt.Sprintf("test-cal-job-grp-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum))
	groupName := fmt.Sprintf("test-acc-ml-grp-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum))
	jobID := fmt.Sprintf("test-acc-ml-grpjob-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum))

	const addr = mlCalendarJobResourceAddr
	const addrJob = "elasticstack_elasticsearch_ml_anomaly_detection_job.job"

	vars := config.Variables{
		"calendar_id": config.StringVariable(calendarID),
		"job_id":      config.StringVariable(jobID),
		"group_name":  config.StringVariable(groupName),
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				PreConfig:                func() { setupAccMLCalendar(t, calendarID) },
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          vars,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(addrJob, "job_id", jobID),
					resource.TestCheckResourceAttr(addrJob, "groups.#", "1"),
					resource.TestCheckResourceAttr(addrJob, "groups.0", groupName),
					resource.TestCheckResourceAttr(addr, "calendar_id", calendarID),
					resource.TestCheckResourceAttr(addr, "job_id", groupName),
					resource.TestCheckResourceAttrSet(addr, "id"),
					resource.TestMatchResourceAttr(addr, "id", testAccMLCalendarJobCompositeIDRegexp(calendarID, groupName)),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          vars,
				ResourceName:             addr,
				ImportState:              true,
				ImportStateVerify:        true,
				ImportStateVerifyIgnore:  []string{"elasticsearch_connection"},
				ImportStateIdFunc:        importMLCalendarJobStateID(addr),
			},
		},
	})
}

func TestAccResourceMLCalendarJob_basic(t *testing.T) {
	calendarID := fmt.Sprintf("test-cal-job-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum))
	jobID := fmt.Sprintf("test-cal-job-ad-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum))

	vars := config.Variables{
		"calendar_id": config.StringVariable(calendarID),
		"job_id":      config.StringVariable(jobID),
	}
	const addr = mlCalendarJobResourceAddr

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				PreConfig:                func() { setupAccMLCalendar(t, calendarID) },
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          vars,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(addr, "calendar_id", calendarID),
					resource.TestCheckResourceAttr(addr, "job_id", jobID),
					resource.TestCheckResourceAttrSet(addr, "id"),
					resource.TestMatchResourceAttr(addr, "id", testAccMLCalendarJobCompositeIDRegexp(calendarID, jobID)),
				),
			},
		},
	})
}

func TestAccResourceMLCalendarJob_import(t *testing.T) {
	calendarID := fmt.Sprintf("test-cal-job-imp-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum))
	jobID := fmt.Sprintf("test-cal-job-imp-ad-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum))

	vars := config.Variables{
		"calendar_id": config.StringVariable(calendarID),
		"job_id":      config.StringVariable(jobID),
	}
	const addr = mlCalendarJobResourceAddr

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				PreConfig:                func() { setupAccMLCalendar(t, calendarID) },
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          vars,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(addr, "calendar_id", calendarID),
					resource.TestCheckResourceAttr(addr, "job_id", jobID),
					resource.TestCheckResourceAttrSet(addr, "id"),
					resource.TestMatchResourceAttr(addr, "id", testAccMLCalendarJobCompositeIDRegexp(calendarID, jobID)),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          vars,
				ResourceName:             addr,
				ImportState:              true,
				ImportStateVerify:        true,
				ImportStateVerifyIgnore:  []string{"elasticsearch_connection"},
				ImportStateIdFunc:        importMLCalendarJobStateID(addr),
			},
		},
	})
}

func TestAccResourceMLCalendarJob_replaceCalendar(t *testing.T) {
	calendarID1 := fmt.Sprintf("test-cal-job-rc1-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum))
	calendarID2 := fmt.Sprintf("test-cal-job-rc2-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum))
	jobID := fmt.Sprintf("test-cal-job-rc-ad-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum))

	const addr = mlCalendarJobResourceAddr

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					setupAccMLCalendar(t, calendarID1)
					setupAccMLCalendar(t, calendarID2)
				},
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"calendar_id": config.StringVariable(calendarID1),
					"job_id":      config.StringVariable(jobID),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(addr, "calendar_id", calendarID1),
					resource.TestCheckResourceAttr(addr, "job_id", jobID),
					resource.TestMatchResourceAttr(addr, "id", testAccMLCalendarJobCompositeIDRegexp(calendarID1, jobID)),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("after_replace"),
				ConfigVariables: config.Variables{
					"calendar_id": config.StringVariable(calendarID2),
					"job_id":      config.StringVariable(jobID),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(addr, "calendar_id", calendarID2),
					resource.TestCheckResourceAttr(addr, "job_id", jobID),
					resource.TestMatchResourceAttr(addr, "id", testAccMLCalendarJobCompositeIDRegexp(calendarID2, jobID)),
				),
			},
		},
	})
}

func TestAccResourceMLCalendarJob_replaceJob(t *testing.T) {
	calendarID := fmt.Sprintf("test-cal-job-rj-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum))
	jobIDA := fmt.Sprintf("test-cal-job-rj-a-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum))
	jobIDB := fmt.Sprintf("test-cal-job-rj-b-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum))

	const addr = mlCalendarJobResourceAddr

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				PreConfig:                func() { setupAccMLCalendar(t, calendarID) },
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"calendar_id": config.StringVariable(calendarID),
					"job_id_a":    config.StringVariable(jobIDA),
					"job_id_b":    config.StringVariable(jobIDB),
					"attach_job":  config.StringVariable("a"),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(addr, "calendar_id", calendarID),
					resource.TestCheckResourceAttr(addr, "job_id", jobIDA),
					resource.TestMatchResourceAttr(addr, "id", testAccMLCalendarJobCompositeIDRegexp(calendarID, jobIDA)),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"calendar_id": config.StringVariable(calendarID),
					"job_id_a":    config.StringVariable(jobIDA),
					"job_id_b":    config.StringVariable(jobIDB),
					"attach_job":  config.StringVariable("b"),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(addr, "calendar_id", calendarID),
					resource.TestCheckResourceAttr(addr, "job_id", jobIDB),
					resource.TestMatchResourceAttr(addr, "id", testAccMLCalendarJobCompositeIDRegexp(calendarID, jobIDB)),
				),
			},
		},
	})
}

func TestAccResourceMLCalendarJob_explicitConnection(t *testing.T) {
	// The default provider is configured with invalid Elasticsearch endpoints and
	// credentials (see testdata/.../calendar_job.tf). The ML job uses aliased
	// provider.elasticstack.setup with elasticsearch {} so it still reaches the
	// cluster. The calendar_job resource must use its elasticsearch_connection
	// block; if it incorrectly used the default provider client, apply would fail.
	//
	// ImportState uses an empty connection list and the provider default client,
	// so import is not part of this test (see TestAccResourceMLCalendarJob_import).
	//
	// ELASTICSEARCH_* env vars may still override some client fields during
	// acceptance runs (internal/clients/config); the invalid default provider still
	// documents intent and catches mis-wiring when env does not override.
	endpoints := testAccMLCalendarJobESEndpoints()
	if len(endpoints) == 0 {
		t.Skip("ELASTICSEARCH_ENDPOINTS must be set to run this test")
	}
	endpointVars := make([]config.Variable, 0, len(endpoints))
	for _, ep := range endpoints {
		endpointVars = append(endpointVars, config.StringVariable(ep))
	}

	calendarID := fmt.Sprintf("test-cal-job-ec-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum))
	jobID := fmt.Sprintf("test-cal-job-ec-ad-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum))

	const addr = mlCalendarJobResourceAddr

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				PreConfig:                func() { setupAccMLCalendar(t, calendarID) },
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"calendar_id": config.StringVariable(calendarID),
					"job_id":      config.StringVariable(jobID),
					"endpoints":   config.ListVariable(endpointVars...),
					"api_key":     config.StringVariable(os.Getenv("ELASTICSEARCH_API_KEY")),
					"username":    config.StringVariable(os.Getenv("ELASTICSEARCH_USERNAME")),
					"password":    config.StringVariable(os.Getenv("ELASTICSEARCH_PASSWORD")),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(addr, "calendar_id", calendarID),
					resource.TestCheckResourceAttr(addr, "job_id", jobID),
					resource.TestCheckResourceAttr(addr, "elasticsearch_connection.#", "1"),
					resource.TestCheckResourceAttr(addr, "elasticsearch_connection.0.endpoints.#", fmt.Sprintf("%d", len(endpoints))),
					resource.TestCheckResourceAttr(addr, "elasticsearch_connection.0.endpoints.0", endpoints[0]),
					resource.TestCheckResourceAttr(addr, "elasticsearch_connection.0.insecure", "true"),
					resource.TestMatchResourceAttr(addr, "id", testAccMLCalendarJobCompositeIDRegexp(calendarID, jobID)),
				),
			},
		},
	})
}

func TestAccResourceMLCalendarJob_planInvalidCalendarID(t *testing.T) {
	jobID := fmt.Sprintf("test-cal-job-badcal-ad-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum))

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"calendar_id": config.StringVariable("INVALID_UPPERCASE_CAL"),
					"job_id":      config.StringVariable(jobID),
				},
				ExpectError: regexp.MustCompile(`(?s)(?:Invalid Attribute Value Match.*calendar_id|calendar_id.*Invalid Attribute Value Match).*must contain lowercase alphanumeric`),
			},
		},
	})
}

func TestAccResourceMLCalendarJob_planInvalidJobID(t *testing.T) {
	calendarID := fmt.Sprintf("test-cal-job-badj-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum))
	jobID := fmt.Sprintf("test-cal-job-badj-ad-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum))

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"calendar_id":    config.StringVariable(calendarID),
					"job_id":         config.StringVariable(jobID),
					"invalid_job_id": config.StringVariable("INVALID_UPPERCASE_JOB"),
				},
				ExpectError: regexp.MustCompile(`(?s)(?:Invalid Attribute Value Match.*job_id|job_id.*Invalid Attribute Value Match).*must contain lowercase alphanumeric`),
			},
		},
	})
}

func TestAccResourceMLCalendarJob_planCalendarIDTooLong(t *testing.T) {
	jobID := fmt.Sprintf("test-cal-job-longcal-ad-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum))
	longCalendarID := strings.Repeat("a", 65)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"calendar_id": config.StringVariable(longCalendarID),
					"job_id":      config.StringVariable(jobID),
				},
				ExpectError: regexp.MustCompile(`(?s)Invalid Attribute Value Length.*calendar_id|calendar_id.*between 1 and 64`),
			},
		},
	})
}

func TestAccResourceMLCalendarJob_applyCalendarNotFound(t *testing.T) {
	// Elasticsearch accepts PutCalendarJob for a job_id that does not exist yet, so
	// "missing job" does not reliably fail apply. A calendar that was never created does.
	missingCalendarID := fmt.Sprintf("test-cal-job-misscal-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum))
	jobID := fmt.Sprintf("test-cal-job-misscal-ad-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum))
	accCleanupMLAnomalyJobAfterTest(t, jobID)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"missing_calendar_id": config.StringVariable(missingCalendarID),
					"job_id":              config.StringVariable(jobID),
				},
				ExpectError: regexp.MustCompile(`(?i)failed to assign ml job|unable to assign job .* to calendar|resource_not_found|no such calendar|404`),
			},
		},
	})
}

// TestAccMLCalendarJob_getCalendarsMissingRepresentedAsNotFound documents how
// Elasticsearch represents a missing calendar for ml.get_calendars: some versions
// return *types.ElasticsearchError with HTTP 404; others return 200 with an empty
// calendars list. readCalendarJob treats both as absent (404 via
// IsNotFoundElasticsearchError, or len(calendars)==0).
func TestAccMLCalendarJob_getCalendarsMissingRepresentedAsNotFound(t *testing.T) {
	t.Parallel()
	acctest.PreCheck(t)

	ctx := context.Background()
	client, err := clients.NewAcceptanceTestingElasticsearchScopedClient()
	if err != nil {
		t.Fatalf("acceptance elasticsearch client: %v", err)
	}
	es := client.GetESClient()
	calendarID := fmt.Sprintf("test-cal-job-nocal-%s", sdkacctest.RandStringFromCharSet(24, sdkacctest.CharSetAlphaNum))

	res, err := es.Ml.GetCalendars().CalendarId(calendarID).Do(ctx)
	if err != nil {
		if !esclient.IsNotFoundElasticsearchError(err) {
			t.Fatalf("missing calendar error must be a 404 elasticsearch error when the API returns one; got [%T] %v", err, err)
		}
		return
	}
	if len(res.Calendars) != 0 {
		t.Fatalf("expected no calendars for missing id %q when the API returns 200, got %d", calendarID, len(res.Calendars))
	}
}

func TestAccResourceMLCalendarJob_refreshRemovesAssignmentWhenCalendarDeleted(t *testing.T) {
	calendarID := fmt.Sprintf("test-cal-job-refdel-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum))
	jobID := fmt.Sprintf("test-cal-job-refdel-ad-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum))
	accCleanupMLAnomalyJobAfterTest(t, jobID)

	vars := config.Variables{
		"calendar_id": config.StringVariable(calendarID),
		"job_id":      config.StringVariable(jobID),
	}
	const addr = mlCalendarJobResourceAddr

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				PreConfig:                func() { setupAccMLCalendar(t, calendarID) },
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          vars,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(addr, "calendar_id", calendarID),
					resource.TestCheckResourceAttr(addr, "job_id", jobID),
					resource.TestCheckResourceAttrSet(addr, "id"),
				),
			},
			{
				PreConfig: func() {
					ctx := context.Background()
					c, err := clients.NewAcceptanceTestingElasticsearchScopedClient()
					if err != nil {
						t.Fatalf("acceptance elasticsearch client: %v", err)
					}
					es := c.GetESClient()
					if _, err := es.Ml.DeleteCalendar(calendarID).Do(ctx); err != nil && !esclient.IsNotFoundElasticsearchError(err) {
						t.Fatalf("delete ML calendar %q before refresh: %v", calendarID, err)
					}
				},
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          vars,
				PlanOnly:                 true,
				ExpectNonEmptyPlan:       true,
				Check:                    testAccCheckMLCalendarJobAbsentFromState(addr),
			},
		},
	})
}
