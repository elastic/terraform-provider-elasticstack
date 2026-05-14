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
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
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
// Call only from resource.TestCase.PreCheck after acctest.PreCheck so no Elasticsearch
// work runs when acceptance prerequisites are not satisfied.
func setupAccMLCalendar(t *testing.T, calendarID string) {
	t.Helper()
	ctx := context.Background()
	client, err := clients.NewAcceptanceTestingElasticsearchScopedClient()
	if err != nil {
		t.Fatalf("acceptance elasticsearch client: %v", err)
	}
	es, err := client.GetESClient()
	if err != nil {
		t.Fatalf("get elasticsearch client: %v", err)
	}
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
	return regexp.MustCompile(`^[^/]+/` + regexp.QuoteMeta(calendarID+"|"+jobID) + `$`)
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
		PreCheck: func() {
			acctest.PreCheck(t)
			setupAccMLCalendar(t, calendarID)
		},
		Steps: []resource.TestStep{
			{
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
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs := s.RootModule().Resources[addr]
					return rs.Primary.ID, nil
				},
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
		PreCheck: func() {
			acctest.PreCheck(t)
			setupAccMLCalendar(t, calendarID)
		},
		Steps: []resource.TestStep{
			{
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
		PreCheck: func() {
			acctest.PreCheck(t)
			setupAccMLCalendar(t, calendarID)
		},
		Steps: []resource.TestStep{
			{
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
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs := s.RootModule().Resources[addr]
					return rs.Primary.ID, nil
				},
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
		PreCheck: func() {
			acctest.PreCheck(t)
			setupAccMLCalendar(t, calendarID1)
			setupAccMLCalendar(t, calendarID2)
		},
		Steps: []resource.TestStep{
			{
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
		PreCheck: func() {
			acctest.PreCheck(t)
			setupAccMLCalendar(t, calendarID)
		},
		Steps: []resource.TestStep{
			{
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
		PreCheck: func() {
			acctest.PreCheck(t)
			setupAccMLCalendar(t, calendarID)
		},
		Steps: []resource.TestStep{
			{
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
			{
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
				ResourceName:            addr,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"elasticsearch_connection"},
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs := s.RootModule().Resources[addr]
					return rs.Primary.ID, nil
				},
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
				ExpectError: regexp.MustCompile(`calendar_id|lowercase|must contain|Invalid Attribute Value Match`),
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
				ExpectError: regexp.MustCompile(`job_id|lowercase|must contain|Invalid Attribute Value Match`),
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
				ExpectError: regexp.MustCompile(`calendar_id|Invalid Attribute Value Length|between 1 and 64|string length`),
			},
		},
	})
}

func TestAccResourceMLCalendarJob_applyCalendarNotFound(t *testing.T) {
	// Elasticsearch accepts PutCalendarJob for a job_id that does not exist yet, so
	// "missing job" does not reliably fail apply. A calendar that was never created does.
	missingCalendarID := fmt.Sprintf("test-cal-job-misscal-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum))
	jobID := fmt.Sprintf("test-cal-job-misscal-ad-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum))

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
				ExpectError: regexp.MustCompile(`(?i)failed to assign|unable to assign|not.*found|resource_not_found|unknown.*calendar|no.*calendar`),
			},
		},
	})
}
