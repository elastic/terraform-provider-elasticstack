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
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccResourceMLCalendarEvent(t *testing.T) {
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
				),
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
