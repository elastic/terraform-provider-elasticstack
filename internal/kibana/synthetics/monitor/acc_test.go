package monitor_test

import (
	"fmt"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/synthetics/monitor"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/go-version"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

var (
	minKibanaVersion = version.Must(version.NewVersion("8.14.0"))
	kibana816Version = version.Must(version.NewVersion("8.16.0"))
)

const (
	httpCheckExpectedUpdated = `{"request":{"body":"name=first\u0026email=someemail@someemailprovider.com",` +
		`"headers":{"Content-Type":"application/x-www-form-urlencoded"},"method":"POST"},` +
		`"response":{"body":{"positive":["foo","bar"]},"status":[200,201,301]}}`

	httpMonitorMinConfig = `

resource "elasticstack_kibana_synthetics_monitor" "%s" {
	name = "TestHttpMonitorResource - %s"
	private_locations = [elasticstack_kibana_synthetics_private_location.%s.label]
	http = {
		url = "http://localhost:5601"
	}
}
`
	httpMonitorConfig = `

resource "elasticstack_kibana_synthetics_monitor" "%s" {
	name = "TestHttpMonitorResource - %s"
	space_id = "testacc"
	namespace = "test_namespace"
	schedule = 5
	private_locations = [elasticstack_kibana_synthetics_private_location.%s.label]
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
		mode = "any"
		ipv4 = true
		ipv6 = false
	}
}
`
	httpMonitorSslConfig = `

resource "elasticstack_kibana_synthetics_monitor" "%s" {
	name = "TestHttpMonitorResource - %s"
	private_locations = [elasticstack_kibana_synthetics_private_location.%s.label]
	http = {
		url = "http://localhost:5601"
		ssl_verification_mode = "full"
		ssl_supported_protocols = ["TLSv1.2"]
		ssl_certificate_authorities = ["ca1", "ca2"]
		ssl_certificate = "cert"
		ssl_key = "key"
		ssl_key_passphrase = "pass"
	}
}
`

	httpMonitorUpdated = `
resource "elasticstack_kibana_synthetics_monitor" "%s" {
	name = "TestHttpMonitorResource Updated - %s"
	space_id = "testacc"
	schedule = 10
	private_locations = [elasticstack_kibana_synthetics_private_location.%s.label]
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
		proxy_header = jsonencode({
			"header-name" = "header-value-updated"
		})
		username = "testupdated"
		password = "testpassword-updated"
		check = jsonencode({
			"request": {
				"method": "POST",
				"headers": {
					"Content-Type": "application/x-www-form-urlencoded",
				},
				"body": "name=first&email=someemail@someemailprovider.com",
			},
			"response": {
				"status": [200, 201, 301],
				"body": {
					"positive": ["foo", "bar"]
				}
			}
		})
		response = jsonencode({
			"include_body":           "never",
			"include_body_max_bytes": "1024",
		})
	}
	params = jsonencode({
		"param-name" = "param-value-updated"
  	})
	retest_on_failure = false
}

`

	tcpMonitorMinConfig = `

resource "elasticstack_kibana_synthetics_monitor" "%s" {
	name = "TestTcpMonitorResource - %s"
	private_locations = [elasticstack_kibana_synthetics_private_location.%s.label]
	tcp = {
		host = "http://localhost:5601"
	}
}
`

	tcpMonitorSslConfig = `

resource "elasticstack_kibana_synthetics_monitor" "%s" {
	name = "TestHttpMonitorResource - %s"
	private_locations = [elasticstack_kibana_synthetics_private_location.%s.label]
	tcp = {
		host = "http://localhost:5601"
		ssl_verification_mode = "full"
		ssl_supported_protocols = ["TLSv1.2"]
		ssl_certificate_authorities = ["ca1", "ca2"]
		ssl_certificate = "cert"
		ssl_key = "key"
		ssl_key_passphrase = "pass"
	}
}
`

	tcpMonitorConfig = `

resource "elasticstack_kibana_synthetics_monitor" "%s" {
	name = "TestTcpMonitorResource - %s"
	space_id = "testacc"
	namespace = "testacc_test"
	schedule = 5
	private_locations = [elasticstack_kibana_synthetics_private_location.%s.label]
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
		proxy_use_local_resolver = true
	}
}
`

	tcpMonitorUpdated = `
resource "elasticstack_kibana_synthetics_monitor" "%s" {
	name = "TestTcpMonitorResource Updated - %s"
	space_id = "testacc"
	schedule = 10
	private_locations = [elasticstack_kibana_synthetics_private_location.%s.label]
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
		check_send = "Hello Updated"
		check_receive = "World Updated"
	}
}
`

	icmpMonitorMinConfig = `

resource "elasticstack_kibana_synthetics_monitor" "%s" {
	name = "TestIcmpMonitorResource - %s"
	private_locations = [elasticstack_kibana_synthetics_private_location.%s.label]
	icmp = {
		host = "localhost"
	}
}
`
	icmpMonitorConfig = `

resource "elasticstack_kibana_synthetics_monitor" "%s" {
	name = "TestIcmpMonitorResource - %s"
	space_id = "testacc"
	namespace = "testacc_namespace"
	schedule = 5
	private_locations = [elasticstack_kibana_synthetics_private_location.%s.label]
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
	icmp = {
		host = "localhost"
	}
}
`

	icmpMonitorUpdated = `
resource "elasticstack_kibana_synthetics_monitor" "%s" {
	name = "TestIcmpMonitorResource Updated - %s"
	space_id = "testacc"
	schedule = 10
	private_locations = [elasticstack_kibana_synthetics_private_location.%s.label]
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
	icmp = {
		host = "google.com"
		wait = 10
	}
}
`
	browserMonitorConfig = `

resource "elasticstack_kibana_synthetics_monitor" "%s" {
	name = "TestBrowserMonitorResource - %s"
	space_id = "testacc"
	namespace = "testacc_ns"
	schedule = 5
	private_locations = [elasticstack_kibana_synthetics_private_location.%s.label]
	enabled = true
	tags = ["a", "b"]
	service_name = "test apm service"
	timeout = 30
	browser = {
		inline_script = "step('Go to https://google.com.co', () => page.goto('https://www.google.com'))"
	}
}
`
	browserMonitorMinConfig = `

resource "elasticstack_kibana_synthetics_monitor" "%s" {
	name = "TestBrowserMonitorResource - %s"
	private_locations = [elasticstack_kibana_synthetics_private_location.%s.label]
	alert = {
		status = {
			enabled = true
		}
		tls = {
			enabled = true
		}
	}
	browser = {
		inline_script = "step('Go to https://google.com.co', () => page.goto('https://www.google.com'))"
	}
}
`

	browserMonitorUpdated = `
resource "elasticstack_kibana_synthetics_monitor" "%s" {
	name = "TestBrowserMonitorResource Updated - %s"
	space_id = "testacc"
	schedule = 10
	private_locations = [elasticstack_kibana_synthetics_private_location.%s.label]
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
	browser = {
		inline_script = "step('Go to https://google.de', () => page.goto('https://www.google.de'))"
		synthetics_args = ["--no-sandbox", "--disable-setuid-sandbox"]
		screenshots = "off"
		ignore_https_errors = true
		playwright_options = jsonencode({"httpCredentials":{"password":"test","username":"test"},"ignoreHTTPSErrors":false})
	}
}
`

	httpMonitorLabelsConfig = `
resource "elasticstack_kibana_synthetics_monitor" "%s" {
	name = "TestHttpMonitorLabels - %s"
	private_locations = [elasticstack_kibana_synthetics_private_location.%s.label]
	labels = {
		environment = "production"
		team = "platform"
		service = "web-app"
	}
	http = {
		url = "http://localhost:5601"
	}
}
`

	httpMonitorLabelsUpdated = `
resource "elasticstack_kibana_synthetics_monitor" "%s" {
	name = "TestHttpMonitorLabels Updated - %s"
	private_locations = [elasticstack_kibana_synthetics_private_location.%s.label]
	labels = {
		environment = "staging"
		team = "platform-updated"
		service = "web-app-v2"
	}
	http = {
		url = "http://localhost:5601"
	}
}
`

	httpMonitorLabelsRemoved = `
resource "elasticstack_kibana_synthetics_monitor" "%s" {
	name = "TestHttpMonitorLabels Removed - %s"
	private_locations = [elasticstack_kibana_synthetics_private_location.%s.label]
	http = {
		url = "http://localhost:5601"
	}
}
`
)

func TestSyntheticMonitorHTTPResource(t *testing.T) {

	name := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)
	id := "http-monitor"
	httpMonitorID, config := testMonitorConfig(id, httpMonitorConfig, name)

	bmName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)
	bmMonitorID, bmConfig := testMonitorConfig("http-monitor-min", httpMonitorMinConfig, bmName)

	sslName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)
	sslHTTPMonitorID, sslConfig := testMonitorConfig("http-monitor-ssl", httpMonitorSslConfig, sslName)

	_, configUpdated := testMonitorConfig(id, httpMonitorUpdated, name)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			// Create and Read http monitor with minimum fields
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minKibanaVersion),
				Config:   bmConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(bmMonitorID, "id"),
					resource.TestCheckResourceAttr(bmMonitorID, "name", "TestHttpMonitorResource - "+bmName),
					resource.TestCheckResourceAttr(bmMonitorID, "space_id", ""),
					resource.TestCheckResourceAttr(bmMonitorID, "namespace", "default"),
					resource.TestCheckResourceAttr(bmMonitorID, "alert.status.enabled", "true"),
					resource.TestCheckResourceAttr(bmMonitorID, "alert.tls.enabled", "true"),
					resource.TestCheckResourceAttr(bmMonitorID, "http.url", "http://localhost:5601"),
				),
			},
			// Create and Read http monitor with ssl fields, starting from ES 8.16.0
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(kibana816Version),
				Config:   sslConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(sslHTTPMonitorID, "id"),
					resource.TestCheckResourceAttr(sslHTTPMonitorID, "name", "TestHttpMonitorResource - "+sslName),
					resource.TestCheckResourceAttr(sslHTTPMonitorID, "space_id", ""),
					resource.TestCheckResourceAttr(sslHTTPMonitorID, "namespace", "default"),
					resource.TestCheckResourceAttr(sslHTTPMonitorID, "http.url", "http://localhost:5601"),
					resource.TestCheckResourceAttr(sslHTTPMonitorID, "http.ssl_verification_mode", "full"),
					resource.TestCheckResourceAttr(sslHTTPMonitorID, "http.ssl_supported_protocols.#", "1"),
					resource.TestCheckResourceAttr(sslHTTPMonitorID, "http.ssl_supported_protocols.0", "TLSv1.2"),
					resource.TestCheckResourceAttr(sslHTTPMonitorID, "http.ssl_certificate_authorities.#", "2"),
					resource.TestCheckResourceAttr(sslHTTPMonitorID, "http.ssl_certificate_authorities.0", "ca1"),
					resource.TestCheckResourceAttr(sslHTTPMonitorID, "http.ssl_certificate_authorities.1", "ca2"),
					resource.TestCheckResourceAttr(sslHTTPMonitorID, "http.ssl_certificate", "cert"),
					resource.TestCheckResourceAttr(sslHTTPMonitorID, "http.ssl_key", "key"),
					resource.TestCheckResourceAttr(sslHTTPMonitorID, "http.ssl_key_passphrase", "pass"),
				),
			},
			// ImportState testing ssl fields
			{
				SkipFunc:          versionutils.CheckIfVersionIsUnsupported(kibana816Version),
				ResourceName:      sslHTTPMonitorID,
				ImportState:       true,
				ImportStateVerify: true,
				Config:            sslConfig,
			},
			// Create and Read http monitor
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minKibanaVersion),
				Config:   config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(httpMonitorID, "id"),
					resource.TestCheckResourceAttr(httpMonitorID, "name", "TestHttpMonitorResource - "+name),
					resource.TestCheckResourceAttr(httpMonitorID, "space_id", "testacc"),
					resource.TestCheckResourceAttr(httpMonitorID, "namespace", "test_namespace"),
					resource.TestCheckResourceAttr(httpMonitorID, "schedule", "5"),
					resource.TestCheckResourceAttr(httpMonitorID, "private_locations.#", "1"),
					resource.TestCheckResourceAttrSet(httpMonitorID, "private_locations.0"),
					resource.TestCheckResourceAttr(httpMonitorID, "enabled", "true"),
					resource.TestCheckResourceAttr(httpMonitorID, "tags.#", "2"),
					resource.TestCheckResourceAttr(httpMonitorID, "tags.0", "a"),
					resource.TestCheckResourceAttr(httpMonitorID, "tags.1", "b"),
					resource.TestCheckResourceAttr(httpMonitorID, "alert.status.enabled", "true"),
					resource.TestCheckResourceAttr(httpMonitorID, "alert.tls.enabled", "true"),
					resource.TestCheckResourceAttr(httpMonitorID, "service_name", "test apm service"),
					resource.TestCheckResourceAttr(httpMonitorID, "timeout", "30"),
					resource.TestCheckResourceAttr(httpMonitorID, "http.url", "http://localhost:5601"),
					resource.TestCheckResourceAttr(httpMonitorID, "http.ssl_verification_mode", "full"),
					resource.TestCheckResourceAttr(httpMonitorID, "http.ssl_supported_protocols.#", "3"),
					resource.TestCheckResourceAttr(httpMonitorID, "http.ssl_supported_protocols.0", "TLSv1.1"),
					resource.TestCheckResourceAttr(httpMonitorID, "http.ssl_supported_protocols.1", "TLSv1.2"),
					resource.TestCheckResourceAttr(httpMonitorID, "http.ssl_supported_protocols.2", "TLSv1.3"),
					resource.TestCheckResourceAttr(httpMonitorID, "http.max_redirects", "0"),
					resource.TestCheckResourceAttr(httpMonitorID, "http.mode", "any"),
					resource.TestCheckResourceAttr(httpMonitorID, "http.ipv4", "true"),
					resource.TestCheckResourceAttr(httpMonitorID, "http.ipv6", "false"),
					resource.TestCheckResourceAttr(httpMonitorID, "http.proxy_url", ""),
				),
			},
			// ImportState testing
			{
				SkipFunc:          versionutils.CheckIfVersionIsUnsupported(minKibanaVersion),
				ResourceName:      httpMonitorID,
				ImportState:       true,
				ImportStateVerify: true,
				Config:            config,
			},
			// Update and Read testing http monitor
			{
				SkipFunc:     versionutils.CheckIfVersionIsUnsupported(minKibanaVersion),
				ResourceName: httpMonitorID,
				Config:       configUpdated,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(httpMonitorID, "id"),
					resource.TestCheckResourceAttr(httpMonitorID, "name", "TestHttpMonitorResource Updated - "+name),
					resource.TestCheckResourceAttr(httpMonitorID, "space_id", "testacc"),
					resource.TestCheckResourceAttr(httpMonitorID, "namespace", "test_namespace"),
					resource.TestCheckResourceAttr(httpMonitorID, "schedule", "10"),
					resource.TestCheckResourceAttr(httpMonitorID, "private_locations.#", "1"),
					resource.TestCheckResourceAttrSet(httpMonitorID, "private_locations.0"),
					resource.TestCheckResourceAttr(httpMonitorID, "enabled", "false"),
					resource.TestCheckResourceAttr(httpMonitorID, "tags.#", "3"),
					resource.TestCheckResourceAttr(httpMonitorID, "tags.0", "c"),
					resource.TestCheckResourceAttr(httpMonitorID, "tags.1", "d"),
					resource.TestCheckResourceAttr(httpMonitorID, "tags.2", "e"),
					resource.TestCheckResourceAttr(httpMonitorID, "alert.status.enabled", "true"),
					resource.TestCheckResourceAttr(httpMonitorID, "alert.tls.enabled", "false"),
					resource.TestCheckResourceAttr(httpMonitorID, "service_name", "test apm service"),
					resource.TestCheckResourceAttr(httpMonitorID, "timeout", "30"),
					resource.TestCheckResourceAttr(httpMonitorID, "http.url", "http://localhost:8080"),
					resource.TestCheckResourceAttr(httpMonitorID, "http.ssl_verification_mode", "full"),
					resource.TestCheckResourceAttr(httpMonitorID, "http.ssl_supported_protocols.#", "1"),
					resource.TestCheckResourceAttr(httpMonitorID, "http.ssl_supported_protocols.0", "TLSv1.2"),
					resource.TestCheckResourceAttr(httpMonitorID, "http.max_redirects", "10"),
					resource.TestCheckResourceAttr(httpMonitorID, "http.mode", "all"),
					resource.TestCheckResourceAttr(httpMonitorID, "http.ipv4", "true"),
					resource.TestCheckResourceAttr(httpMonitorID, "http.ipv6", "true"),
					resource.TestCheckResourceAttr(httpMonitorID, "http.proxy_url", "http://localhost"),
					resource.TestCheckNoResourceAttr(httpMonitorID, "tcp"),
					resource.TestCheckNoResourceAttr(httpMonitorID, "browser"),
					resource.TestCheckNoResourceAttr(httpMonitorID, "icmp"),
					// check for merge attributes
					resource.TestCheckResourceAttr(httpMonitorID, "http.proxy_header", `{"header-name":"header-value-updated"}`),
					resource.TestCheckResourceAttr(httpMonitorID, "http.username", "testupdated"),
					resource.TestCheckResourceAttr(httpMonitorID, "http.password", "testpassword-updated"),
					resource.TestCheckResourceAttr(httpMonitorID, "http.check", httpCheckExpectedUpdated),
					resource.TestCheckResourceAttr(httpMonitorID, "http.response", `{"include_body":"never","include_body_max_bytes":"1024"}`),
					resource.TestCheckResourceAttr(httpMonitorID, "params", `{"param-name":"param-value-updated"}`),
					resource.TestCheckResourceAttr(httpMonitorID, "retest_on_failure", "false"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestSyntheticMonitorTCPResource(t *testing.T) {

	name := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)
	id := "tcp-monitor"
	tcpMonitorID, config := testMonitorConfig(id, tcpMonitorConfig, name)
	_, configUpdated := testMonitorConfig(id, tcpMonitorUpdated, name)

	bmName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)
	bmMonitorID, bmConfig := testMonitorConfig("tcp-monitor-min", tcpMonitorMinConfig, bmName)

	sslName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)
	sslTCPMonitorID, sslConfig := testMonitorConfig("tcp-monitor-ssl", tcpMonitorSslConfig, sslName)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			// Create and Read tcp monitor with minimum fields
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minKibanaVersion),
				Config:   bmConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(bmMonitorID, "id"),
					resource.TestCheckResourceAttr(bmMonitorID, "name", "TestTcpMonitorResource - "+bmName),
					resource.TestCheckResourceAttr(bmMonitorID, "space_id", ""),
					resource.TestCheckResourceAttr(bmMonitorID, "namespace", "default"),
					resource.TestCheckResourceAttr(bmMonitorID, "tcp.host", "http://localhost:5601"),
					resource.TestCheckResourceAttr(bmMonitorID, "alert.status.enabled", "true"),
					resource.TestCheckResourceAttr(bmMonitorID, "alert.tls.enabled", "true"),
				),
			},
			// Create and Read tcp monitor with ssl fields, starting from ES 8.16.0
			// Create and Read tcp monitor with ssl fields, starting from ES 8.16.0
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(kibana816Version),
				Config:   sslConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(sslTCPMonitorID, "id"),
					resource.TestCheckResourceAttr(sslTCPMonitorID, "name", "TestHttpMonitorResource - "+sslName),
					resource.TestCheckResourceAttr(sslTCPMonitorID, "space_id", ""),
					resource.TestCheckResourceAttr(sslTCPMonitorID, "namespace", "default"),
					resource.TestCheckResourceAttr(sslTCPMonitorID, "tcp.host", "http://localhost:5601"),
					resource.TestCheckResourceAttr(sslTCPMonitorID, "tcp.ssl_verification_mode", "full"),
					resource.TestCheckResourceAttr(sslTCPMonitorID, "tcp.ssl_supported_protocols.#", "1"),
					resource.TestCheckResourceAttr(sslTCPMonitorID, "tcp.ssl_supported_protocols.0", "TLSv1.2"),
					resource.TestCheckResourceAttr(sslTCPMonitorID, "tcp.ssl_certificate_authorities.#", "2"),
					resource.TestCheckResourceAttr(sslTCPMonitorID, "tcp.ssl_certificate_authorities.0", "ca1"),
					resource.TestCheckResourceAttr(sslTCPMonitorID, "tcp.ssl_certificate_authorities.1", "ca2"),
					resource.TestCheckResourceAttr(sslTCPMonitorID, "tcp.ssl_certificate", "cert"),
					resource.TestCheckResourceAttr(sslTCPMonitorID, "tcp.ssl_key", "key"),
					resource.TestCheckResourceAttr(sslTCPMonitorID, "tcp.ssl_key_passphrase", "pass"),
				),
			},
			// ImportState testing ssl fields
			{
				SkipFunc:          versionutils.CheckIfVersionIsUnsupported(kibana816Version),
				ResourceName:      sslTCPMonitorID,
				ImportState:       true,
				ImportStateVerify: true,
				Config:            sslConfig,
			},
			// Create and Read tcp monitor
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minKibanaVersion),
				Config:   config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(tcpMonitorID, "id"),
					resource.TestCheckResourceAttr(tcpMonitorID, "name", "TestTcpMonitorResource - "+name),
					resource.TestCheckResourceAttr(tcpMonitorID, "space_id", "testacc"),
					resource.TestCheckResourceAttr(tcpMonitorID, "namespace", "testacc_test"),
					resource.TestCheckResourceAttr(tcpMonitorID, "schedule", "5"),
					resource.TestCheckResourceAttr(tcpMonitorID, "private_locations.#", "1"),
					resource.TestCheckResourceAttrSet(tcpMonitorID, "private_locations.0"),
					resource.TestCheckResourceAttr(tcpMonitorID, "enabled", "true"),
					resource.TestCheckResourceAttr(tcpMonitorID, "tags.#", "2"),
					resource.TestCheckResourceAttr(tcpMonitorID, "tags.0", "a"),
					resource.TestCheckResourceAttr(tcpMonitorID, "tags.1", "b"),
					resource.TestCheckResourceAttr(tcpMonitorID, "alert.status.enabled", "true"),
					resource.TestCheckResourceAttr(tcpMonitorID, "alert.tls.enabled", "true"),
					resource.TestCheckResourceAttr(tcpMonitorID, "service_name", "test apm service"),
					resource.TestCheckResourceAttr(tcpMonitorID, "timeout", "30"),
					resource.TestCheckResourceAttr(tcpMonitorID, "tcp.host", "http://localhost:5601"),
					resource.TestCheckResourceAttr(tcpMonitorID, "tcp.ssl_verification_mode", "full"),
					resource.TestCheckResourceAttr(tcpMonitorID, "tcp.ssl_supported_protocols.#", "3"),
					resource.TestCheckResourceAttr(tcpMonitorID, "tcp.ssl_supported_protocols.0", "TLSv1.1"),
					resource.TestCheckResourceAttr(tcpMonitorID, "tcp.ssl_supported_protocols.1", "TLSv1.2"),
					resource.TestCheckResourceAttr(tcpMonitorID, "tcp.ssl_supported_protocols.2", "TLSv1.3"),
					resource.TestCheckResourceAttr(tcpMonitorID, "tcp.proxy_url", ""),
					resource.TestCheckResourceAttr(tcpMonitorID, "tcp.proxy_use_local_resolver", "true"),
				),
			},
			// ImportState testing
			{
				SkipFunc:          versionutils.CheckIfVersionIsUnsupported(minKibanaVersion),
				ResourceName:      tcpMonitorID,
				ImportState:       true,
				ImportStateVerify: true,
				Config:            config,
			},
			// Update and Read tcp monitor
			{
				SkipFunc:     versionutils.CheckIfVersionIsUnsupported(minKibanaVersion),
				ResourceName: tcpMonitorID,
				Config:       configUpdated,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(tcpMonitorID, "id"),
					resource.TestCheckResourceAttr(tcpMonitorID, "name", "TestTcpMonitorResource Updated - "+name),
					resource.TestCheckResourceAttr(tcpMonitorID, "space_id", "testacc"),
					resource.TestCheckResourceAttr(tcpMonitorID, "namespace", "testacc_test"),
					resource.TestCheckResourceAttr(tcpMonitorID, "schedule", "10"),
					resource.TestCheckResourceAttr(tcpMonitorID, "private_locations.#", "1"),
					resource.TestCheckResourceAttrSet(tcpMonitorID, "private_locations.0"),
					resource.TestCheckResourceAttr(tcpMonitorID, "enabled", "false"),
					resource.TestCheckResourceAttr(tcpMonitorID, "tags.#", "3"),
					resource.TestCheckResourceAttr(tcpMonitorID, "tags.0", "c"),
					resource.TestCheckResourceAttr(tcpMonitorID, "tags.1", "d"),
					resource.TestCheckResourceAttr(tcpMonitorID, "tags.2", "e"),
					resource.TestCheckResourceAttr(tcpMonitorID, "alert.status.enabled", "true"),
					resource.TestCheckResourceAttr(tcpMonitorID, "alert.tls.enabled", "false"),
					resource.TestCheckResourceAttr(tcpMonitorID, "service_name", "test apm service"),
					resource.TestCheckResourceAttr(tcpMonitorID, "timeout", "30"),
					resource.TestCheckResourceAttr(tcpMonitorID, "tcp.host", "http://localhost:8080"),
					resource.TestCheckResourceAttr(tcpMonitorID, "tcp.ssl_verification_mode", "full"),
					resource.TestCheckResourceAttr(tcpMonitorID, "tcp.ssl_supported_protocols.#", "1"),
					resource.TestCheckResourceAttr(tcpMonitorID, "tcp.ssl_supported_protocols.0", "TLSv1.2"),
					resource.TestCheckResourceAttr(tcpMonitorID, "tcp.proxy_url", "http://localhost"),
					resource.TestCheckResourceAttr(tcpMonitorID, "tcp.proxy_use_local_resolver", "false"),
					resource.TestCheckNoResourceAttr(tcpMonitorID, "http"),
					resource.TestCheckNoResourceAttr(tcpMonitorID, "browser"),
					resource.TestCheckNoResourceAttr(tcpMonitorID, "icmp"),
					// check for merge attributes
					resource.TestCheckResourceAttr(tcpMonitorID, "tcp.check_send", "Hello Updated"),
					resource.TestCheckResourceAttr(tcpMonitorID, "tcp.check_receive", "World Updated"),
				),
			},
			// Delete testing automatically occurs in TestCase

		},
	})
}

func TestSyntheticMonitorICMPResource(t *testing.T) {

	name := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)
	id := "icmp-monitor"
	icmpMonitorID, config := testMonitorConfig(id, icmpMonitorConfig, name)
	_, configUpdated := testMonitorConfig(id, icmpMonitorUpdated, name)

	bmName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)
	bmMonitorID, bmConfig := testMonitorConfig("icmp-monitor-min", icmpMonitorMinConfig, bmName)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			// Create and Read icmp monitor with minimum fields
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minKibanaVersion),
				Config:   bmConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(bmMonitorID, "id"),
					resource.TestCheckResourceAttr(bmMonitorID, "name", "TestIcmpMonitorResource - "+bmName),
					resource.TestCheckResourceAttr(bmMonitorID, "space_id", ""),
					resource.TestCheckResourceAttr(bmMonitorID, "namespace", "default"),
					resource.TestCheckResourceAttr(bmMonitorID, "icmp.host", "localhost"),
					resource.TestCheckResourceAttr(bmMonitorID, "alert.status.enabled", "true"),
					resource.TestCheckResourceAttr(bmMonitorID, "alert.tls.enabled", "true"),
				),
			},

			// Create and Read icmp monitor
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minKibanaVersion),
				Config:   config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(icmpMonitorID, "id"),
					resource.TestCheckResourceAttr(icmpMonitorID, "name", "TestIcmpMonitorResource - "+name),
					resource.TestCheckResourceAttr(icmpMonitorID, "space_id", "testacc"),
					resource.TestCheckResourceAttr(icmpMonitorID, "namespace", "testacc_namespace"),
					resource.TestCheckResourceAttr(icmpMonitorID, "schedule", "5"),
					resource.TestCheckResourceAttr(icmpMonitorID, "private_locations.#", "1"),
					resource.TestCheckResourceAttrSet(icmpMonitorID, "private_locations.0"),
					resource.TestCheckResourceAttr(icmpMonitorID, "enabled", "true"),
					resource.TestCheckResourceAttr(icmpMonitorID, "tags.#", "2"),
					resource.TestCheckResourceAttr(icmpMonitorID, "tags.0", "a"),
					resource.TestCheckResourceAttr(icmpMonitorID, "tags.1", "b"),
					resource.TestCheckResourceAttr(icmpMonitorID, "alert.status.enabled", "true"),
					resource.TestCheckResourceAttr(icmpMonitorID, "alert.tls.enabled", "true"),
					resource.TestCheckResourceAttr(icmpMonitorID, "service_name", "test apm service"),
					resource.TestCheckResourceAttr(icmpMonitorID, "timeout", "30"),
					resource.TestCheckResourceAttr(icmpMonitorID, "icmp.host", "localhost"),
				),
			},
			// ImportState testing
			{
				SkipFunc:          versionutils.CheckIfVersionIsUnsupported(minKibanaVersion),
				ResourceName:      icmpMonitorID,
				ImportState:       true,
				ImportStateVerify: true,
				Config:            config,
			},
			// Update and Read icmp monitor
			{
				SkipFunc:     versionutils.CheckIfVersionIsUnsupported(minKibanaVersion),
				ResourceName: icmpMonitorID,
				Config:       configUpdated,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(icmpMonitorID, "id"),
					resource.TestCheckResourceAttr(icmpMonitorID, "name", "TestIcmpMonitorResource Updated - "+name),
					resource.TestCheckResourceAttr(icmpMonitorID, "space_id", "testacc"),
					resource.TestCheckResourceAttr(icmpMonitorID, "namespace", "testacc_namespace"),
					resource.TestCheckResourceAttr(icmpMonitorID, "schedule", "10"),
					resource.TestCheckResourceAttr(icmpMonitorID, "private_locations.#", "1"),
					resource.TestCheckResourceAttrSet(icmpMonitorID, "private_locations.0"),
					resource.TestCheckResourceAttr(icmpMonitorID, "enabled", "false"),
					resource.TestCheckResourceAttr(icmpMonitorID, "tags.#", "3"),
					resource.TestCheckResourceAttr(icmpMonitorID, "tags.0", "c"),
					resource.TestCheckResourceAttr(icmpMonitorID, "tags.1", "d"),
					resource.TestCheckResourceAttr(icmpMonitorID, "tags.2", "e"),
					resource.TestCheckResourceAttr(icmpMonitorID, "alert.status.enabled", "true"),
					resource.TestCheckResourceAttr(icmpMonitorID, "alert.tls.enabled", "false"),
					resource.TestCheckResourceAttr(icmpMonitorID, "service_name", "test apm service"),
					resource.TestCheckResourceAttr(icmpMonitorID, "timeout", "30"),
					resource.TestCheckResourceAttr(icmpMonitorID, "icmp.host", "google.com"),
					resource.TestCheckResourceAttr(icmpMonitorID, "icmp.wait", "10"),
					resource.TestCheckNoResourceAttr(icmpMonitorID, "http"),
					resource.TestCheckNoResourceAttr(icmpMonitorID, "browser"),
					resource.TestCheckNoResourceAttr(icmpMonitorID, "tcp"),
				),
			},
			// Delete testing automatically occurs in TestCase

		},
	})
}

func TestSyntheticMonitorBrowserResource(t *testing.T) {

	name := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)
	id := "browser-monitor"
	browserMonitorID, config := testMonitorConfig(id, browserMonitorConfig, name)
	_, configUpdated := testMonitorConfig(id, browserMonitorUpdated, name)

	bmName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)
	bmMonitorID, bmConfig := testMonitorConfig("browser-monitor-min", browserMonitorMinConfig, bmName)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			// Create and Read browser monitor with minimum fields
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minKibanaVersion),
				Config:   bmConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(bmMonitorID, "id"),
					resource.TestCheckResourceAttr(bmMonitorID, "name", "TestBrowserMonitorResource - "+bmName),
					resource.TestCheckResourceAttr(bmMonitorID, "space_id", ""),
					resource.TestCheckResourceAttr(bmMonitorID, "namespace", "default"),
					resource.TestCheckResourceAttr(bmMonitorID, "browser.inline_script", "step('Go to https://google.com.co', () => page.goto('https://www.google.com'))"),
					resource.TestCheckResourceAttr(bmMonitorID, "alert.status.enabled", "true"),
					resource.TestCheckResourceAttr(bmMonitorID, "alert.tls.enabled", "true"),
				),
			},
			// Create and Read browser monitor
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minKibanaVersion),
				Config:   config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(browserMonitorID, "id"),
					resource.TestCheckResourceAttr(browserMonitorID, "name", "TestBrowserMonitorResource - "+name),
					resource.TestCheckResourceAttr(browserMonitorID, "space_id", "testacc"),
					resource.TestCheckResourceAttr(browserMonitorID, "namespace", "testacc_ns"),
					resource.TestCheckResourceAttr(browserMonitorID, "schedule", "5"),
					resource.TestCheckResourceAttr(browserMonitorID, "private_locations.#", "1"),
					resource.TestCheckResourceAttrSet(browserMonitorID, "private_locations.0"),
					resource.TestCheckResourceAttr(browserMonitorID, "enabled", "true"),
					resource.TestCheckResourceAttr(browserMonitorID, "tags.#", "2"),
					resource.TestCheckResourceAttr(browserMonitorID, "tags.0", "a"),
					resource.TestCheckResourceAttr(browserMonitorID, "tags.1", "b"),
					resource.TestCheckResourceAttr(browserMonitorID, "alert.status.enabled", "true"),
					resource.TestCheckResourceAttr(browserMonitorID, "alert.tls.enabled", "true"),
					resource.TestCheckResourceAttr(browserMonitorID, "service_name", "test apm service"),
					resource.TestCheckResourceAttr(browserMonitorID, "timeout", "30"),
					resource.TestCheckResourceAttr(browserMonitorID, "browser.inline_script", "step('Go to https://google.com.co', () => page.goto('https://www.google.com'))"),
				),
			},
			// ImportState testing
			{
				SkipFunc:          versionutils.CheckIfVersionIsUnsupported(kibana816Version),
				ResourceName:      browserMonitorID,
				ImportState:       true,
				ImportStateVerify: true,
				Config:            config,
			},
			// Update and Read browser monitor
			{
				SkipFunc:     versionutils.CheckIfVersionIsUnsupported(minKibanaVersion),
				ResourceName: browserMonitorID,
				Config:       configUpdated,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(browserMonitorID, "id"),
					resource.TestCheckResourceAttr(browserMonitorID, "name", "TestBrowserMonitorResource Updated - "+name),
					resource.TestCheckResourceAttr(browserMonitorID, "space_id", "testacc"),
					resource.TestCheckResourceAttr(browserMonitorID, "namespace", "testacc_ns"),
					resource.TestCheckResourceAttr(browserMonitorID, "schedule", "10"),
					resource.TestCheckResourceAttr(browserMonitorID, "private_locations.#", "1"),
					resource.TestCheckResourceAttrSet(browserMonitorID, "private_locations.0"),
					resource.TestCheckResourceAttr(browserMonitorID, "enabled", "false"),
					resource.TestCheckResourceAttr(browserMonitorID, "tags.#", "3"),
					resource.TestCheckResourceAttr(browserMonitorID, "tags.0", "c"),
					resource.TestCheckResourceAttr(browserMonitorID, "tags.1", "d"),
					resource.TestCheckResourceAttr(browserMonitorID, "tags.2", "e"),
					resource.TestCheckResourceAttr(browserMonitorID, "alert.status.enabled", "true"),
					resource.TestCheckResourceAttr(browserMonitorID, "alert.tls.enabled", "false"),
					resource.TestCheckResourceAttr(browserMonitorID, "service_name", "test apm service"),
					resource.TestCheckResourceAttr(browserMonitorID, "timeout", "30"),
					resource.TestCheckResourceAttr(browserMonitorID, "browser.inline_script", "step('Go to https://google.de', () => page.goto('https://www.google.de'))"),
					resource.TestCheckResourceAttr(browserMonitorID, "browser.synthetics_args.#", "2"),
					resource.TestCheckResourceAttr(browserMonitorID, "browser.synthetics_args.0", "--no-sandbox"),
					resource.TestCheckResourceAttr(browserMonitorID, "browser.synthetics_args.1", "--disable-setuid-sandbox"),
					resource.TestCheckResourceAttr(browserMonitorID, "browser.screenshots", "off"),
					resource.TestCheckResourceAttr(browserMonitorID, "browser.ignore_https_errors", "true"),
					resource.TestCheckResourceAttr(browserMonitorID, "browser.playwright_options", `{"httpCredentials":{"password":"test","username":"test"},"ignoreHTTPSErrors":false}`),
					resource.TestCheckNoResourceAttr(browserMonitorID, "http"),
					resource.TestCheckNoResourceAttr(browserMonitorID, "icmp"),
					resource.TestCheckNoResourceAttr(browserMonitorID, "tcp"),
				),
			},
			// Delete testing automatically occurs in TestCase

		},
	})
}

func TestSyntheticMonitorLabelsResource(t *testing.T) {
	name := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)
	id := "http-monitor-labels"
	labelsMonitorID, labelsConfig := testMonitorConfig(id, httpMonitorLabelsConfig, name)
	_, labelsConfigUpdated := testMonitorConfig(id, httpMonitorLabelsUpdated, name)
	_, labelsConfigRemoved := testMonitorConfig(id, httpMonitorLabelsRemoved, name)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			// Create and Read monitor with labels
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(monitor.MinLabelsVersion),
				Config:   labelsConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(labelsMonitorID, "id"),
					resource.TestCheckResourceAttr(labelsMonitorID, "name", "TestHttpMonitorLabels - "+name),
					resource.TestCheckResourceAttr(labelsMonitorID, "labels.%", "3"),
					resource.TestCheckResourceAttr(labelsMonitorID, "labels.environment", "production"),
					resource.TestCheckResourceAttr(labelsMonitorID, "labels.team", "platform"),
					resource.TestCheckResourceAttr(labelsMonitorID, "labels.service", "web-app"),
					resource.TestCheckResourceAttr(labelsMonitorID, "http.url", "http://localhost:5601"),
				),
			},
			// ImportState testing
			{
				SkipFunc:          versionutils.CheckIfVersionIsUnsupported(monitor.MinLabelsVersion),
				ResourceName:      labelsMonitorID,
				ImportState:       true,
				ImportStateVerify: true,
				Config:            labelsConfig,
			},
			// Update labels - change values but keep same keys
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(monitor.MinLabelsVersion),
				Config:   labelsConfigUpdated,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(labelsMonitorID, "id"),
					resource.TestCheckResourceAttr(labelsMonitorID, "name", "TestHttpMonitorLabels Updated - "+name),
					resource.TestCheckResourceAttr(labelsMonitorID, "labels.%", "3"),
					resource.TestCheckResourceAttr(labelsMonitorID, "labels.environment", "staging"),
					resource.TestCheckResourceAttr(labelsMonitorID, "labels.team", "platform-updated"),
					resource.TestCheckResourceAttr(labelsMonitorID, "labels.service", "web-app-v2"),
				),
			},
			// Remove all labels - this tests the round-trip consistency fix
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(monitor.MinLabelsVersion),
				Config:   labelsConfigRemoved,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(labelsMonitorID, "id"),
					resource.TestCheckResourceAttr(labelsMonitorID, "name", "TestHttpMonitorLabels Removed - "+name),
					resource.TestCheckNoResourceAttr(labelsMonitorID, "labels.%"),
					resource.TestCheckNoResourceAttr(labelsMonitorID, "labels.environment"),
					resource.TestCheckNoResourceAttr(labelsMonitorID, "labels.team"),
					resource.TestCheckNoResourceAttr(labelsMonitorID, "labels.service"),
					resource.TestCheckNoResourceAttr(labelsMonitorID, "labels.version"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testMonitorConfig(id, cfg, name string) (string, string) {

	resourceID := "elasticstack_kibana_synthetics_monitor." + id
	privateLocationID := "pl-" + id
	agentPolicyID := "apl-" + id

	provider := fmt.Sprintf(`
provider "elasticstack" {
  	elasticsearch {}
	kibana {}
	fleet{}
}

resource "elasticstack_fleet_agent_policy" "%s" {
	name            = "TestMonitorResource Agent Policy - %s"
	namespace       = "testacc"
	description     = "TestMonitorResource Agent Policy"
	monitor_logs    = true
	monitor_metrics = true
	skip_destroy    = false
}

resource "elasticstack_kibana_synthetics_private_location" "%s" {
	label = "monitor-pll-%s"
	agent_policy_id = elasticstack_fleet_agent_policy.%s.policy_id
}
`, agentPolicyID, name, privateLocationID, name, agentPolicyID)

	config := fmt.Sprintf(cfg, id, name, privateLocationID)

	return resourceID, provider + config
}
