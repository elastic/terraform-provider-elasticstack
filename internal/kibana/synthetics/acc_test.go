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
	tcpMonitorId  = "elasticstack_kibana_synthetics_monitor.tcp-monitor"

	providerConfig = `
provider "elasticstack" {
  	elasticsearch {}
	kibana {}
	fleet{}
}
`

	privateLocationConfig = `

resource "elasticstack_fleet_agent_policy" "test" {
	name            = "TestMonitorResource Agent Policy - test"
	namespace       = "testacc"
	description     = "TestMonitorResource Agent Policy"
	monitor_logs    = true
	monitor_metrics = true
	skip_destroy    = false
}

resource "elasticstack_kibana_synthetics_private_location" "test" {
	label = "TestMonitorResource-label"
	space_id = "testacc"
	agent_policy_id = elasticstack_fleet_agent_policy.test.policy_id
}

`

	httpMonitorConfig = `

resource "elasticstack_kibana_synthetics_monitor" "http-monitor" {
	name = "TestHttpMonitorResource"
	space_id = "testacc"
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
		max_redirects = 10
		mode = "any"
		ipv4 = true
		ipv6 = false
		proxy_url = "http://localhost:8080"
	}
}
`

	httpMonitorUpdated = `
resource "elasticstack_kibana_synthetics_monitor" "http-monitor" {
	name = "TestHttpMonitorResource Updated"
	space_id = "testacc"
	schedule = 10
	private_locations = [elasticstack_kibana_synthetics_private_location.test.label]
	enabled = false
	tags = ["c", "d", "e"]
	alert = {
		status = {
			enabled = true
		}
		tls = {
			enabled = false
		}
	}
	service_name = "test apm service"
	timeout = 30
	http = {
		url = "http://localhost:8080"
		ssl_verification_mode = "full"
		ssl_supported_protocols = ["TLSv1.2"]
		max_redirects = 10
		mode = "all"
		ipv4 = true
		ipv6 = true
		proxy_url = "http://localhost"
	}
}

`

	tcpMonitorConfig = `

resource "elasticstack_kibana_synthetics_monitor" "tcp-monitor" {
	name = "TestTcpMonitorResource"
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
	tcp = {
		host = "http://localhost:5601"
		ssl_verification_mode = "full"
		ssl_supported_protocols = ["TLSv1.0", "TLSv1.1", "TLSv1.2"]
		proxy_url = "http://localhost:8080"
		proxy_use_local_resolver = true
	}
}
`

	tcpMonitorUpdated = `
resource "elasticstack_kibana_synthetics_monitor" "tcp-monitor" {
	name = "TestTcpMonitorResource Updated"
	space_id = "default"
	schedule = 10
	private_locations = [elasticstack_kibana_synthetics_private_location.test.label]
	enabled = false
	tags = ["c", "d", "e"]
	alert = {
		status = {
			enabled = true
		}
		tls = {
			enabled = false
		}
	}
	service_name = "test apm service"
	timeout = 30
	tcp = {
		host = "http://localhost:8080"
		ssl_verification_mode = "full"
		ssl_supported_protocols = ["TLSv1.2"]
		proxy_url = "http://localhost"
		proxy_use_local_resolver = false
	}
}

`
)

func TestSyntheticMonitorResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			// Create and Read http monitor
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minKibanaVersion),
				Config:   providerConfig + privateLocationConfig + httpMonitorConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(httpMonitorId, "id"),
					resource.TestCheckResourceAttr(httpMonitorId, "name", "TestHttpMonitorResource"),
					resource.TestCheckResourceAttr(httpMonitorId, "space_id", "testacc"),
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
					resource.TestCheckResourceAttr(httpMonitorId, "http.ssl_verification_mode", "full"),
					resource.TestCheckResourceAttr(httpMonitorId, "http.ssl_supported_protocols.#", "3"),
					resource.TestCheckResourceAttr(httpMonitorId, "http.ssl_supported_protocols.0", "TLSv1.0"),
					resource.TestCheckResourceAttr(httpMonitorId, "http.ssl_supported_protocols.1", "TLSv1.1"),
					resource.TestCheckResourceAttr(httpMonitorId, "http.ssl_supported_protocols.2", "TLSv1.2"),
					resource.TestCheckResourceAttr(httpMonitorId, "http.max_redirects", "10"),
					resource.TestCheckResourceAttr(httpMonitorId, "http.mode", "any"),
					resource.TestCheckResourceAttr(httpMonitorId, "http.ipv4", "true"),
					resource.TestCheckResourceAttr(httpMonitorId, "http.ipv6", "false"),
					resource.TestCheckResourceAttr(httpMonitorId, "http.proxy_url", "http://localhost:8080"),
				),
			},
			// ImportState testing
			{
				SkipFunc:          versionutils.CheckIfVersionIsUnsupported(minKibanaVersion),
				ResourceName:      httpMonitorId,
				ImportState:       true,
				ImportStateVerify: true,
				Config:            providerConfig + privateLocationConfig + httpMonitorConfig,
			},
			// Update and Read testing http monitor
			{
				SkipFunc:     versionutils.CheckIfVersionIsUnsupported(minKibanaVersion),
				ResourceName: httpMonitorId,
				Config:       providerConfig + privateLocationConfig + httpMonitorUpdated,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(httpMonitorId, "id"),
					resource.TestCheckResourceAttr(httpMonitorId, "name", "TestHttpMonitorResource Updated"),
					resource.TestCheckResourceAttr(httpMonitorId, "space_id", "testacc"),
					resource.TestCheckResourceAttr(httpMonitorId, "schedule", "10"),
					resource.TestCheckResourceAttr(httpMonitorId, "private_locations.#", "1"),
					resource.TestCheckResourceAttrSet(httpMonitorId, "private_locations.0"),
					resource.TestCheckResourceAttr(httpMonitorId, "enabled", "false"),
					resource.TestCheckResourceAttr(httpMonitorId, "tags.#", "3"),
					resource.TestCheckResourceAttr(httpMonitorId, "tags.0", "c"),
					resource.TestCheckResourceAttr(httpMonitorId, "tags.1", "d"),
					resource.TestCheckResourceAttr(httpMonitorId, "tags.2", "e"),
					resource.TestCheckResourceAttr(httpMonitorId, "alert.status.enabled", "true"),
					resource.TestCheckResourceAttr(httpMonitorId, "alert.tls.enabled", "false"),
					resource.TestCheckResourceAttr(httpMonitorId, "service_name", "test apm service"),
					resource.TestCheckResourceAttr(httpMonitorId, "timeout", "30"),
					resource.TestCheckResourceAttr(httpMonitorId, "http.url", "http://localhost:8080"),
					resource.TestCheckResourceAttr(httpMonitorId, "http.ssl_verification_mode", "full"),
					resource.TestCheckResourceAttr(httpMonitorId, "http.ssl_supported_protocols.#", "1"),
					resource.TestCheckResourceAttr(httpMonitorId, "http.ssl_supported_protocols.0", "TLSv1.2"),
					resource.TestCheckResourceAttr(httpMonitorId, "http.max_redirects", "10"),
					resource.TestCheckResourceAttr(httpMonitorId, "http.mode", "all"),
					resource.TestCheckResourceAttr(httpMonitorId, "http.ipv4", "true"),
					resource.TestCheckResourceAttr(httpMonitorId, "http.ipv6", "true"),
					resource.TestCheckResourceAttr(httpMonitorId, "http.proxy_url", "http://localhost"),
					resource.TestCheckNoResourceAttr(httpMonitorId, "tcp"),
				),
			},
			// Create and Read tcp monitor
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minKibanaVersion),
				Config:   providerConfig + privateLocationConfig + tcpMonitorConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(tcpMonitorId, "id"),
					resource.TestCheckResourceAttr(tcpMonitorId, "name", "TestTcpMonitorResource"),
					resource.TestCheckResourceAttr(tcpMonitorId, "space_id", "default"),
					resource.TestCheckResourceAttr(tcpMonitorId, "schedule", "5"),
					resource.TestCheckResourceAttr(tcpMonitorId, "private_locations.#", "1"),
					resource.TestCheckResourceAttrSet(tcpMonitorId, "private_locations.0"),
					resource.TestCheckResourceAttr(tcpMonitorId, "enabled", "true"),
					resource.TestCheckResourceAttr(tcpMonitorId, "tags.#", "2"),
					resource.TestCheckResourceAttr(tcpMonitorId, "tags.0", "a"),
					resource.TestCheckResourceAttr(tcpMonitorId, "tags.1", "b"),
					resource.TestCheckResourceAttr(tcpMonitorId, "alert.status.enabled", "true"),
					resource.TestCheckResourceAttr(tcpMonitorId, "alert.tls.enabled", "true"),
					resource.TestCheckResourceAttr(tcpMonitorId, "service_name", "test apm service"),
					resource.TestCheckResourceAttr(tcpMonitorId, "timeout", "30"),
					resource.TestCheckResourceAttr(tcpMonitorId, "tcp.host", "http://localhost:5601"),
					resource.TestCheckResourceAttr(tcpMonitorId, "tcp.ssl_verification_mode", "full"),
					resource.TestCheckResourceAttr(tcpMonitorId, "tcp.ssl_supported_protocols.#", "3"),
					resource.TestCheckResourceAttr(tcpMonitorId, "tcp.ssl_supported_protocols.0", "TLSv1.0"),
					resource.TestCheckResourceAttr(tcpMonitorId, "tcp.ssl_supported_protocols.1", "TLSv1.1"),
					resource.TestCheckResourceAttr(tcpMonitorId, "tcp.ssl_supported_protocols.2", "TLSv1.2"),
					resource.TestCheckResourceAttr(tcpMonitorId, "tcp.proxy_url", "http://localhost:8080"),
					resource.TestCheckResourceAttr(tcpMonitorId, "tcp.proxy_use_local_resolver", "true"),
				),
			},
			// ImportState testing
			{
				SkipFunc:          versionutils.CheckIfVersionIsUnsupported(minKibanaVersion),
				ResourceName:      tcpMonitorId,
				ImportState:       true,
				ImportStateVerify: true,
				Config:            providerConfig + privateLocationConfig + tcpMonitorConfig,
			},
			// Update and Read tcp monitor
			{
				SkipFunc:     versionutils.CheckIfVersionIsUnsupported(minKibanaVersion),
				ResourceName: tcpMonitorId,
				Config:       providerConfig + privateLocationConfig + tcpMonitorUpdated,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(tcpMonitorId, "id"),
					resource.TestCheckResourceAttr(tcpMonitorId, "name", "TestTcpMonitorResource Updated"),
					resource.TestCheckResourceAttr(tcpMonitorId, "space_id", "default"),
					resource.TestCheckResourceAttr(tcpMonitorId, "schedule", "10"),
					resource.TestCheckResourceAttr(tcpMonitorId, "private_locations.#", "1"),
					resource.TestCheckResourceAttrSet(tcpMonitorId, "private_locations.0"),
					resource.TestCheckResourceAttr(tcpMonitorId, "enabled", "false"),
					resource.TestCheckResourceAttr(tcpMonitorId, "tags.#", "3"),
					resource.TestCheckResourceAttr(tcpMonitorId, "tags.0", "c"),
					resource.TestCheckResourceAttr(tcpMonitorId, "tags.1", "d"),
					resource.TestCheckResourceAttr(tcpMonitorId, "tags.2", "e"),
					resource.TestCheckResourceAttr(tcpMonitorId, "alert.status.enabled", "true"),
					resource.TestCheckResourceAttr(tcpMonitorId, "alert.tls.enabled", "false"),
					resource.TestCheckResourceAttr(tcpMonitorId, "service_name", "test apm service"),
					resource.TestCheckResourceAttr(tcpMonitorId, "timeout", "30"),
					resource.TestCheckResourceAttr(tcpMonitorId, "tcp.host", "http://localhost:8080"),
					resource.TestCheckResourceAttr(tcpMonitorId, "tcp.ssl_verification_mode", "full"),
					resource.TestCheckResourceAttr(tcpMonitorId, "tcp.ssl_supported_protocols.#", "1"),
					resource.TestCheckResourceAttr(tcpMonitorId, "tcp.ssl_supported_protocols.0", "TLSv1.2"),
					resource.TestCheckResourceAttr(tcpMonitorId, "tcp.proxy_url", "http://localhost"),
					resource.TestCheckResourceAttr(tcpMonitorId, "tcp.proxy_use_local_resolver", "false"),
					resource.TestCheckNoResourceAttr(tcpMonitorId, "http"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
