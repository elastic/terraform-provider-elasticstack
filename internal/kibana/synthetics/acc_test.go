package synthetics_test

import (
	"fmt"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/go-version"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

var (
	minKibanaVersion = version.Must(version.NewVersion("8.14.0"))
	kibana816Version = version.Must(version.NewVersion("8.16.0"))
)

const (
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
	httpMonitorConfigWithNamespace = `

resource "elasticstack_kibana_synthetics_monitor" "%s" {
	name = "TestHttpMonitorResource - %s"
	space_id = "testacc"
	namespace = "testnamespace"
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
)

func TestSyntheticMonitorHTTPResource(t *testing.T) {

	name := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)
	id := "http-monitor"
	httpMonitorId, config := testMonitorConfig(id, httpMonitorConfig, name)

	bmName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)
	bmMonitorId, bmConfig := testMonitorConfig("http-monitor-min", httpMonitorMinConfig, bmName)

	sslName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)
	sslHttpMonitorId, sslConfig := testMonitorConfig("http-monitor-ssl", httpMonitorSslConfig, sslName)

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
					resource.TestCheckResourceAttrSet(bmMonitorId, "id"),
					resource.TestCheckResourceAttr(bmMonitorId, "name", "TestHttpMonitorResource - "+bmName),
					resource.TestCheckResourceAttr(bmMonitorId, "space_id", "default"),
					resource.TestCheckResourceAttr(bmMonitorId, "alert.status.enabled", "true"),
					resource.TestCheckResourceAttr(bmMonitorId, "alert.tls.enabled", "true"),
					resource.TestCheckResourceAttr(bmMonitorId, "http.url", "http://localhost:5601"),
				),
			},
			// Create and Read http monitor with ssl fields, starting from ES 8.16.0
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(kibana816Version),
				Config:   sslConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(sslHttpMonitorId, "id"),
					resource.TestCheckResourceAttr(sslHttpMonitorId, "name", "TestHttpMonitorResource - "+sslName),
					resource.TestCheckResourceAttr(sslHttpMonitorId, "space_id", "default"),
					resource.TestCheckResourceAttr(sslHttpMonitorId, "http.url", "http://localhost:5601"),
					resource.TestCheckResourceAttr(sslHttpMonitorId, "http.ssl_verification_mode", "full"),
					resource.TestCheckResourceAttr(sslHttpMonitorId, "http.ssl_supported_protocols.#", "1"),
					resource.TestCheckResourceAttr(sslHttpMonitorId, "http.ssl_supported_protocols.0", "TLSv1.2"),
					resource.TestCheckResourceAttr(sslHttpMonitorId, "http.ssl_certificate_authorities.#", "2"),
					resource.TestCheckResourceAttr(sslHttpMonitorId, "http.ssl_certificate_authorities.0", "ca1"),
					resource.TestCheckResourceAttr(sslHttpMonitorId, "http.ssl_certificate_authorities.1", "ca2"),
					resource.TestCheckResourceAttr(sslHttpMonitorId, "http.ssl_certificate", "cert"),
					resource.TestCheckResourceAttr(sslHttpMonitorId, "http.ssl_key", "key"),
					resource.TestCheckResourceAttr(sslHttpMonitorId, "http.ssl_key_passphrase", "pass"),
				),
			},
			// ImportState testing ssl fields
			{
				SkipFunc:          versionutils.CheckIfVersionIsUnsupported(kibana816Version),
				ResourceName:      sslHttpMonitorId,
				ImportState:       true,
				ImportStateVerify: true,
				Config:            sslConfig,
			},
			// Create and Read http monitor
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minKibanaVersion),
				Config:   config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(httpMonitorId, "id"),
					resource.TestCheckResourceAttr(httpMonitorId, "name", "TestHttpMonitorResource - "+name),
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
					resource.TestCheckResourceAttr(httpMonitorId, "http.ssl_supported_protocols.0", "TLSv1.1"),
					resource.TestCheckResourceAttr(httpMonitorId, "http.ssl_supported_protocols.1", "TLSv1.2"),
					resource.TestCheckResourceAttr(httpMonitorId, "http.ssl_supported_protocols.2", "TLSv1.3"),
					resource.TestCheckResourceAttr(httpMonitorId, "http.max_redirects", "0"),
					resource.TestCheckResourceAttr(httpMonitorId, "http.mode", "any"),
					resource.TestCheckResourceAttr(httpMonitorId, "http.ipv4", "true"),
					resource.TestCheckResourceAttr(httpMonitorId, "http.ipv6", "false"),
					resource.TestCheckResourceAttr(httpMonitorId, "http.proxy_url", ""),
				),
			},
			// ImportState testing
			{
				SkipFunc:          versionutils.CheckIfVersionIsUnsupported(minKibanaVersion),
				ResourceName:      httpMonitorId,
				ImportState:       true,
				ImportStateVerify: true,
				Config:            config,
			},
			// Update and Read testing http monitor
			{
				SkipFunc:     versionutils.CheckIfVersionIsUnsupported(minKibanaVersion),
				ResourceName: httpMonitorId,
				Config:       configUpdated,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(httpMonitorId, "id"),
					resource.TestCheckResourceAttr(httpMonitorId, "name", "TestHttpMonitorResource Updated - "+name),
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
					resource.TestCheckNoResourceAttr(httpMonitorId, "browser"),
					resource.TestCheckNoResourceAttr(httpMonitorId, "icmp"),
					//check for merge attributes
					resource.TestCheckResourceAttr(httpMonitorId, "http.proxy_header", `{"header-name":"header-value-updated"}`),
					resource.TestCheckResourceAttr(httpMonitorId, "http.username", "testupdated"),
					resource.TestCheckResourceAttr(httpMonitorId, "http.password", "testpassword-updated"),
					resource.TestCheckResourceAttr(httpMonitorId, "http.check", `{"request":{"body":"name=first\u0026email=someemail@someemailprovider.com","headers":{"Content-Type":"application/x-www-form-urlencoded"},"method":"POST"},"response":{"body":{"positive":["foo","bar"]},"status":[200,201,301]}}`),
					resource.TestCheckResourceAttr(httpMonitorId, "http.response", `{"include_body":"never","include_body_max_bytes":"1024"}`),
					resource.TestCheckResourceAttr(httpMonitorId, "params", `{"param-name":"param-value-updated"}`),
					resource.TestCheckResourceAttr(httpMonitorId, "retest_on_failure", "false"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestSyntheticMonitorTCPResource(t *testing.T) {

	name := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)
	id := "tcp-monitor"
	tcpMonitorId, config := testMonitorConfig(id, tcpMonitorConfig, name)
	_, configUpdated := testMonitorConfig(id, tcpMonitorUpdated, name)

	bmName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)
	bmMonitorId, bmConfig := testMonitorConfig("tcp-monitor-min", tcpMonitorMinConfig, bmName)

	sslName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)
	sslTcpMonitorId, sslConfig := testMonitorConfig("tcp-monitor-ssl", tcpMonitorSslConfig, sslName)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			// Create and Read tcp monitor with minimum fields
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minKibanaVersion),
				Config:   bmConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(bmMonitorId, "id"),
					resource.TestCheckResourceAttr(bmMonitorId, "name", "TestTcpMonitorResource - "+bmName),
					resource.TestCheckResourceAttr(bmMonitorId, "space_id", "default"),
					resource.TestCheckResourceAttr(bmMonitorId, "tcp.host", "http://localhost:5601"),
					resource.TestCheckResourceAttr(bmMonitorId, "alert.status.enabled", "true"),
					resource.TestCheckResourceAttr(bmMonitorId, "alert.tls.enabled", "true"),
				),
			},
			// Create and Read tcp monitor with ssl fields, starting from ES 8.16.0
			// Create and Read tcp monitor with ssl fields, starting from ES 8.16.0
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(kibana816Version),
				Config:   sslConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(sslTcpMonitorId, "id"),
					resource.TestCheckResourceAttr(sslTcpMonitorId, "name", "TestHttpMonitorResource - "+sslName),
					resource.TestCheckResourceAttr(sslTcpMonitorId, "space_id", "default"),
					resource.TestCheckResourceAttr(sslTcpMonitorId, "tcp.host", "http://localhost:5601"),
					resource.TestCheckResourceAttr(sslTcpMonitorId, "tcp.ssl_verification_mode", "full"),
					resource.TestCheckResourceAttr(sslTcpMonitorId, "tcp.ssl_supported_protocols.#", "1"),
					resource.TestCheckResourceAttr(sslTcpMonitorId, "tcp.ssl_supported_protocols.0", "TLSv1.2"),
					resource.TestCheckResourceAttr(sslTcpMonitorId, "tcp.ssl_certificate_authorities.#", "2"),
					resource.TestCheckResourceAttr(sslTcpMonitorId, "tcp.ssl_certificate_authorities.0", "ca1"),
					resource.TestCheckResourceAttr(sslTcpMonitorId, "tcp.ssl_certificate_authorities.1", "ca2"),
					resource.TestCheckResourceAttr(sslTcpMonitorId, "tcp.ssl_certificate", "cert"),
					resource.TestCheckResourceAttr(sslTcpMonitorId, "tcp.ssl_key", "key"),
					resource.TestCheckResourceAttr(sslTcpMonitorId, "tcp.ssl_key_passphrase", "pass"),
				),
			},
			// ImportState testing ssl fields
			{
				SkipFunc:          versionutils.CheckIfVersionIsUnsupported(kibana816Version),
				ResourceName:      sslTcpMonitorId,
				ImportState:       true,
				ImportStateVerify: true,
				Config:            sslConfig,
			},
			// Create and Read tcp monitor
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minKibanaVersion),
				Config:   config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(tcpMonitorId, "id"),
					resource.TestCheckResourceAttr(tcpMonitorId, "name", "TestTcpMonitorResource - "+name),
					resource.TestCheckResourceAttr(tcpMonitorId, "space_id", "testacc"),
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
					resource.TestCheckResourceAttr(tcpMonitorId, "tcp.ssl_supported_protocols.0", "TLSv1.1"),
					resource.TestCheckResourceAttr(tcpMonitorId, "tcp.ssl_supported_protocols.1", "TLSv1.2"),
					resource.TestCheckResourceAttr(tcpMonitorId, "tcp.ssl_supported_protocols.2", "TLSv1.3"),
					resource.TestCheckResourceAttr(tcpMonitorId, "tcp.proxy_url", ""),
					resource.TestCheckResourceAttr(tcpMonitorId, "tcp.proxy_use_local_resolver", "true"),
				),
			},
			// ImportState testing
			{
				SkipFunc:          versionutils.CheckIfVersionIsUnsupported(minKibanaVersion),
				ResourceName:      tcpMonitorId,
				ImportState:       true,
				ImportStateVerify: true,
				Config:            config,
			},
			// Update and Read tcp monitor
			{
				SkipFunc:     versionutils.CheckIfVersionIsUnsupported(minKibanaVersion),
				ResourceName: tcpMonitorId,
				Config:       configUpdated,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(tcpMonitorId, "id"),
					resource.TestCheckResourceAttr(tcpMonitorId, "name", "TestTcpMonitorResource Updated - "+name),
					resource.TestCheckResourceAttr(tcpMonitorId, "space_id", "testacc"),
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
					resource.TestCheckNoResourceAttr(tcpMonitorId, "browser"),
					resource.TestCheckNoResourceAttr(tcpMonitorId, "icmp"),
					//check for merge attributes
					resource.TestCheckResourceAttr(tcpMonitorId, "tcp.check_send", "Hello Updated"),
					resource.TestCheckResourceAttr(tcpMonitorId, "tcp.check_receive", "World Updated"),
				),
			},
			// Delete testing automatically occurs in TestCase

		},
	})
}

func TestSyntheticMonitorICMPResource(t *testing.T) {

	name := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)
	id := "icmp-monitor"
	icmpMonitorId, config := testMonitorConfig(id, icmpMonitorConfig, name)
	_, configUpdated := testMonitorConfig(id, icmpMonitorUpdated, name)

	bmName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)
	bmMonitorId, bmConfig := testMonitorConfig("icmp-monitor-min", icmpMonitorMinConfig, bmName)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			// Create and Read icmp monitor with minimum fields
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minKibanaVersion),
				Config:   bmConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(bmMonitorId, "id"),
					resource.TestCheckResourceAttr(bmMonitorId, "name", "TestIcmpMonitorResource - "+bmName),
					resource.TestCheckResourceAttr(bmMonitorId, "space_id", "default"),
					resource.TestCheckResourceAttr(bmMonitorId, "icmp.host", "localhost"),
					resource.TestCheckResourceAttr(bmMonitorId, "alert.status.enabled", "true"),
					resource.TestCheckResourceAttr(bmMonitorId, "alert.tls.enabled", "true"),
				),
			},

			// Create and Read icmp monitor
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minKibanaVersion),
				Config:   config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(icmpMonitorId, "id"),
					resource.TestCheckResourceAttr(icmpMonitorId, "name", "TestIcmpMonitorResource - "+name),
					resource.TestCheckResourceAttr(icmpMonitorId, "space_id", "testacc"),
					resource.TestCheckResourceAttr(icmpMonitorId, "schedule", "5"),
					resource.TestCheckResourceAttr(icmpMonitorId, "private_locations.#", "1"),
					resource.TestCheckResourceAttrSet(icmpMonitorId, "private_locations.0"),
					resource.TestCheckResourceAttr(icmpMonitorId, "enabled", "true"),
					resource.TestCheckResourceAttr(icmpMonitorId, "tags.#", "2"),
					resource.TestCheckResourceAttr(icmpMonitorId, "tags.0", "a"),
					resource.TestCheckResourceAttr(icmpMonitorId, "tags.1", "b"),
					resource.TestCheckResourceAttr(icmpMonitorId, "alert.status.enabled", "true"),
					resource.TestCheckResourceAttr(icmpMonitorId, "alert.tls.enabled", "true"),
					resource.TestCheckResourceAttr(icmpMonitorId, "service_name", "test apm service"),
					resource.TestCheckResourceAttr(icmpMonitorId, "timeout", "30"),
					resource.TestCheckResourceAttr(icmpMonitorId, "icmp.host", "localhost"),
				),
			},
			// ImportState testing
			{
				SkipFunc:          versionutils.CheckIfVersionIsUnsupported(minKibanaVersion),
				ResourceName:      icmpMonitorId,
				ImportState:       true,
				ImportStateVerify: true,
				Config:            config,
			},
			// Update and Read icmp monitor
			{
				SkipFunc:     versionutils.CheckIfVersionIsUnsupported(minKibanaVersion),
				ResourceName: icmpMonitorId,
				Config:       configUpdated,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(icmpMonitorId, "id"),
					resource.TestCheckResourceAttr(icmpMonitorId, "name", "TestIcmpMonitorResource Updated - "+name),
					resource.TestCheckResourceAttr(icmpMonitorId, "space_id", "testacc"),
					resource.TestCheckResourceAttr(icmpMonitorId, "schedule", "10"),
					resource.TestCheckResourceAttr(icmpMonitorId, "private_locations.#", "1"),
					resource.TestCheckResourceAttrSet(icmpMonitorId, "private_locations.0"),
					resource.TestCheckResourceAttr(icmpMonitorId, "enabled", "false"),
					resource.TestCheckResourceAttr(icmpMonitorId, "tags.#", "3"),
					resource.TestCheckResourceAttr(icmpMonitorId, "tags.0", "c"),
					resource.TestCheckResourceAttr(icmpMonitorId, "tags.1", "d"),
					resource.TestCheckResourceAttr(icmpMonitorId, "tags.2", "e"),
					resource.TestCheckResourceAttr(icmpMonitorId, "alert.status.enabled", "true"),
					resource.TestCheckResourceAttr(icmpMonitorId, "alert.tls.enabled", "false"),
					resource.TestCheckResourceAttr(icmpMonitorId, "service_name", "test apm service"),
					resource.TestCheckResourceAttr(icmpMonitorId, "timeout", "30"),
					resource.TestCheckResourceAttr(icmpMonitorId, "icmp.host", "google.com"),
					resource.TestCheckResourceAttr(icmpMonitorId, "icmp.wait", "10"),
					resource.TestCheckNoResourceAttr(icmpMonitorId, "http"),
					resource.TestCheckNoResourceAttr(icmpMonitorId, "browser"),
					resource.TestCheckNoResourceAttr(icmpMonitorId, "tcp"),
				),
			},
			// Delete testing automatically occurs in TestCase

		},
	})
}

func TestSyntheticMonitorBrowserResource(t *testing.T) {

	name := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)
	id := "browser-monitor"
	browserMonitorId, config := testMonitorConfig(id, browserMonitorConfig, name)
	_, configUpdated := testMonitorConfig(id, browserMonitorUpdated, name)

	bmName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)
	bmMonitorId, bmConfig := testMonitorConfig("browser-monitor-min", browserMonitorMinConfig, bmName)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			// Create and Read browser monitor with minimum fields
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minKibanaVersion),
				Config:   bmConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(bmMonitorId, "id"),
					resource.TestCheckResourceAttr(bmMonitorId, "name", "TestBrowserMonitorResource - "+bmName),
					resource.TestCheckResourceAttr(bmMonitorId, "space_id", "default"),
					resource.TestCheckResourceAttr(bmMonitorId, "browser.inline_script", "step('Go to https://google.com.co', () => page.goto('https://www.google.com'))"),
					resource.TestCheckResourceAttr(bmMonitorId, "alert.status.enabled", "true"),
					resource.TestCheckResourceAttr(bmMonitorId, "alert.tls.enabled", "true"),
				),
			},
			// Create and Read browser monitor
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minKibanaVersion),
				Config:   config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(browserMonitorId, "id"),
					resource.TestCheckResourceAttr(browserMonitorId, "name", "TestBrowserMonitorResource - "+name),
					resource.TestCheckResourceAttr(browserMonitorId, "space_id", "testacc"),
					resource.TestCheckResourceAttr(browserMonitorId, "schedule", "5"),
					resource.TestCheckResourceAttr(browserMonitorId, "private_locations.#", "1"),
					resource.TestCheckResourceAttrSet(browserMonitorId, "private_locations.0"),
					resource.TestCheckResourceAttr(browserMonitorId, "enabled", "true"),
					resource.TestCheckResourceAttr(browserMonitorId, "tags.#", "2"),
					resource.TestCheckResourceAttr(browserMonitorId, "tags.0", "a"),
					resource.TestCheckResourceAttr(browserMonitorId, "tags.1", "b"),
					resource.TestCheckResourceAttr(browserMonitorId, "alert.status.enabled", "true"),
					resource.TestCheckResourceAttr(browserMonitorId, "alert.tls.enabled", "true"),
					resource.TestCheckResourceAttr(browserMonitorId, "service_name", "test apm service"),
					resource.TestCheckResourceAttr(browserMonitorId, "timeout", "30"),
					resource.TestCheckResourceAttr(browserMonitorId, "browser.inline_script", "step('Go to https://google.com.co', () => page.goto('https://www.google.com'))"),
				),
			},
			// ImportState testing
			{
				SkipFunc:          versionutils.CheckIfVersionIsUnsupported(kibana816Version),
				ResourceName:      browserMonitorId,
				ImportState:       true,
				ImportStateVerify: true,
				Config:            config,
			},
			// Update and Read browser monitor
			{
				SkipFunc:     versionutils.CheckIfVersionIsUnsupported(minKibanaVersion),
				ResourceName: browserMonitorId,
				Config:       configUpdated,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(browserMonitorId, "id"),
					resource.TestCheckResourceAttr(browserMonitorId, "name", "TestBrowserMonitorResource Updated - "+name),
					resource.TestCheckResourceAttr(browserMonitorId, "space_id", "testacc"),
					resource.TestCheckResourceAttr(browserMonitorId, "schedule", "10"),
					resource.TestCheckResourceAttr(browserMonitorId, "private_locations.#", "1"),
					resource.TestCheckResourceAttrSet(browserMonitorId, "private_locations.0"),
					resource.TestCheckResourceAttr(browserMonitorId, "enabled", "false"),
					resource.TestCheckResourceAttr(browserMonitorId, "tags.#", "3"),
					resource.TestCheckResourceAttr(browserMonitorId, "tags.0", "c"),
					resource.TestCheckResourceAttr(browserMonitorId, "tags.1", "d"),
					resource.TestCheckResourceAttr(browserMonitorId, "tags.2", "e"),
					resource.TestCheckResourceAttr(browserMonitorId, "alert.status.enabled", "true"),
					resource.TestCheckResourceAttr(browserMonitorId, "alert.tls.enabled", "false"),
					resource.TestCheckResourceAttr(browserMonitorId, "service_name", "test apm service"),
					resource.TestCheckResourceAttr(browserMonitorId, "timeout", "30"),
					resource.TestCheckResourceAttr(browserMonitorId, "browser.inline_script", "step('Go to https://google.de', () => page.goto('https://www.google.de'))"),
					resource.TestCheckResourceAttr(browserMonitorId, "browser.synthetics_args.#", "2"),
					resource.TestCheckResourceAttr(browserMonitorId, "browser.synthetics_args.0", "--no-sandbox"),
					resource.TestCheckResourceAttr(browserMonitorId, "browser.synthetics_args.1", "--disable-setuid-sandbox"),
					resource.TestCheckResourceAttr(browserMonitorId, "browser.screenshots", "off"),
					resource.TestCheckResourceAttr(browserMonitorId, "browser.ignore_https_errors", "true"),
					resource.TestCheckResourceAttr(browserMonitorId, "browser.playwright_options", `{"httpCredentials":{"password":"test","username":"test"},"ignoreHTTPSErrors":false}`),
					resource.TestCheckNoResourceAttr(browserMonitorId, "http"),
					resource.TestCheckNoResourceAttr(browserMonitorId, "icmp"),
					resource.TestCheckNoResourceAttr(browserMonitorId, "tcp"),
				),
			},
			// Delete testing automatically occurs in TestCase

		},
	})
}

func testMonitorConfig(id, cfg, name string) (string, string) {

	resourceId := "elasticstack_kibana_synthetics_monitor." + id
	privateLocationId := "pl-" + id
	agentPolicyId := "apl-" + id

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
`, agentPolicyId, name, privateLocationId, name, agentPolicyId)

	config := fmt.Sprintf(cfg, id, name, privateLocationId)

	return resourceId, provider + config
}

func TestSyntheticMonitorHTTPResourceWithNamespace(t *testing.T) {

	name := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)
	id := "http-monitor-namespace"
	httpMonitorId, config := testMonitorConfig(id, httpMonitorConfigWithNamespace, name)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			// Create and Read http monitor with explicit namespace
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minKibanaVersion),
				Config:   config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(httpMonitorId, "id"),
					resource.TestCheckResourceAttr(httpMonitorId, "name", "TestHttpMonitorResource - "+name),
					resource.TestCheckResourceAttr(httpMonitorId, "space_id", "testacc"),
					resource.TestCheckResourceAttr(httpMonitorId, "namespace", "testnamespace"),
					resource.TestCheckResourceAttr(httpMonitorId, "schedule", "5"),
					resource.TestCheckResourceAttr(httpMonitorId, "enabled", "true"),
					resource.TestCheckResourceAttr(httpMonitorId, "http.url", "http://localhost:5601"),
				),
			},
			// Import
			{
				SkipFunc:     versionutils.CheckIfVersionIsUnsupported(minKibanaVersion),
				ResourceName: httpMonitorId,
				ImportState:  true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					return s.RootModule().Resources[httpMonitorId].Primary.Attributes["id"], nil
				},
				ImportStateVerify: true,
			},
		},
	})
}
