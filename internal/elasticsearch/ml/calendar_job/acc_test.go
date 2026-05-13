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
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// setupAccMLCalendar creates an ML calendar via the Elasticsearch API so acceptance tests
// do not depend on the elasticstack_elasticsearch_ml_calendar resource (which may not
// be present on all branches). The calendar is deleted in t.Cleanup after the test.
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

func TestAccResourceMLCalendarJob_basic(t *testing.T) {
	acctest.PreCheck(t)

	calendarID := fmt.Sprintf("test-cal-job-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum))
	jobID := fmt.Sprintf("test-cal-job-ad-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum))
	setupAccMLCalendar(t, calendarID)

	vars := config.Variables{
		"calendar_id": config.StringVariable(calendarID),
		"job_id":      config.StringVariable(jobID),
	}
	const addr = "elasticstack_elasticsearch_ml_calendar_job.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
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
				),
			},
		},
	})
}

func TestAccResourceMLCalendarJob_import(t *testing.T) {
	acctest.PreCheck(t)

	calendarID := fmt.Sprintf("test-cal-job-imp-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum))
	jobID := fmt.Sprintf("test-cal-job-imp-ad-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum))
	setupAccMLCalendar(t, calendarID)

	vars := config.Variables{
		"calendar_id": config.StringVariable(calendarID),
		"job_id":      config.StringVariable(jobID),
	}
	const addr = "elasticstack_elasticsearch_ml_calendar_job.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
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
