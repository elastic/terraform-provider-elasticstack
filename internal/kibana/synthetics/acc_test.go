package synthetics_test

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

var (
	minKibanaVersion = version.Must(version.NewVersion("8.14.0"))
)

const (
	httpMonitorId = "elasticstack_kibana_synthetics_monitor.http-monitor"

	providerConfig = `
provider "elasticstack" {
  	elasticsearch {}
	kibana {}
	fleet{}
}

resource "elasticstack_fleet_agent_policy" "test" {
	name            = "TestMonitorResource Agent Policy - test"
	namespace       = "default"
	description     = "TestMonitorResource Agent Policy"
	monitor_logs    = true
	monitor_metrics = true
	skip_destroy    = false
}

resource "elasticstack_kibana_synthetics_private_location" "test" {
	label = "TestMonitorResource-label"
	space_id = "default"
	agent_policy_id = elasticstack_fleet_agent_policy.test.policy_id
}

resource "elasticstack_kibana_synthetics_monitor" "http-monitor" {
	name = "TestHttpMonitorResource"
	space_id = "default"
	schedule = 5
	private_locations = [elasticstack_kibana_synthetics_private_location.test.label]
	enabled = true
	tags = ["a", "b"]
	alert = {
		status = {
			enabled = true
		}
		tls = {
			enabled = true
		}
	}
	service_name = "test apm service"
	timeout = 30
	http = {
		url = "http://localhost:5601"
		ssl_verification_mode = "full"
		ssl_supported_protocols = ["TLSv1.0", "TLSv1.1", "TLSv1.2"]
		max_redirects = "10"
		mode = "any"
		ipv4 = true
		ipv6 = false
		proxy_url = "http://localhost:8080"
	}
}
`

	/*
		check.send = "Hello"
		check.receive = "World"
		proxy_use_local_resolver = true

	*/
)

func TestSyntheticMonitorResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minKibanaVersion),
				Config:   providerConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(httpMonitorId, "id"),
					resource.TestCheckResourceAttr(httpMonitorId, "name", "TestHttpMonitorResource"),
					resource.TestCheckResourceAttr(httpMonitorId, "space_id", "default"),
					resource.TestCheckResourceAttr(httpMonitorId, "schedule", "5"),
					resource.TestCheckResourceAttr(httpMonitorId, "private_locations.#", "1"),
					resource.TestCheckResourceAttrSet(httpMonitorId, "private_locations.0"),
					resource.TestCheckResourceAttr(httpMonitorId, "enabled", "true"),
					resource.TestCheckResourceAttr(httpMonitorId, "tags.#", "2"),
					resource.TestCheckResourceAttr(httpMonitorId, "tags.0", "a"),
					resource.TestCheckResourceAttr(httpMonitorId, "tags.1", "b"),
					resource.TestCheckResourceAttr(httpMonitorId, "alert.status.enabled", "true"),
					resource.TestCheckResourceAttr(httpMonitorId, "alert.tls.enabled", "true"),
					resource.TestCheckResourceAttr(httpMonitorId, "service_name", "test apm service"),
					resource.TestCheckResourceAttr(httpMonitorId, "timeout", "30"),
					resource.TestCheckResourceAttr(httpMonitorId, "http.url", "http://localhost:5601"),
				),
			},
			// ImportState testing
			{
				SkipFunc:          versionutils.CheckIfVersionIsUnsupported(minKibanaVersion),
				ResourceName:      httpMonitorId,
				ImportState:       true,
				ImportStateVerify: true,
				Config:            providerConfig,
			},
			// Update and Read testing
			/*
				{
					SkipFunc: versionutils.CheckIfVersionIsUnsupported(minKibanaVersion),
					Config:   "", // TODO
					Check:    resource.ComposeAggregateTestCheckFunc(
					// TODO
					),
				},
			*/
			// Delete testing automatically occurs in TestCase
		},
	})
}
