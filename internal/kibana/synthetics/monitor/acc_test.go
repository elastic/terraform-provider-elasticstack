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

package monitor_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/synthetics/monitor"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

var (
	minKibanaVersion = version.Must(version.NewVersion("8.14.0"))
	kibana816Version = version.Must(version.NewVersion("8.16.0"))
)

// accTestKibanaSpaceIDCharset matches elasticstack_kibana_space space_id validation (^[a-z0-9_-]+$).
const accTestKibanaSpaceIDCharset = "abcdefghijklmnopqrstuvwxyz0123456789_-"

const (
	httpCheckExpectedUpdated = `{"request":{"body":"name=first\u0026email=someemail@someemailprovider.com",` +
		`"headers":{"Content-Type":"application/x-www-form-urlencoded"},"method":"POST"},` +
		`"response":{"body":{"positive":["foo","bar"]},"status":[200,201,301]}}`
)

func validationTest(t *testing.T, check resource.ErrorCheckFunc) {
	name := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minKibanaVersion),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("validate"),
				ConfigVariables: config.Variables{
					"name": config.StringVariable(name),
				},
			},
		},
		ErrorCheck: check,
	})

}

func TestSyntheticMonitorSchemaValidation(t *testing.T) {

	validationTest(t, func(err error) error {
		if !strings.Contains(err.Error(), "Attribute locations[0] value must be one of") {
			return errors.New("expected to get locations validation error")
		}
		if !strings.Contains(err.Error(), "Attribute namespace namespace must not contain any of the following") {
			return errors.New("expected to get namespace validation error")
		}
		return nil
	})
}

func TestSyntheticMonitorSchemaValidationNoLocation(t *testing.T) {

	t.Setenv("TF_ELASTICSTACK_SKIP_LOCATION_VALIDATION", "true")
	validationTest(t, func(err error) error {
		if strings.Contains(err.Error(), "Attribute locations[0] value must be one of") {
			return errors.New("not expected to get locations validation error")
		}
		if !strings.Contains(err.Error(), "Attribute namespace namespace must not contain any of the following") {
			return errors.New("expected to get namespace validation error")
		}
		return nil
	})
}

func TestSyntheticMonitorHTTPResource(t *testing.T) {

	httpMonitorID := "elasticstack_kibana_synthetics_monitor.http-monitor"
	bmMonitorID := "elasticstack_kibana_synthetics_monitor.http-monitor-min"
	sslHTTPMonitorID := "elasticstack_kibana_synthetics_monitor.http-monitor-ssl"

	name := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)
	bmName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)
	sslName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			// Create and Read http monitor with minimum fields
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minKibanaVersion),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("http_min"),
				ConfigVariables: config.Variables{
					"name": config.StringVariable(bmName),
				},
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
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(kibana816Version),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("http_ssl"),
				ConfigVariables: config.Variables{
					"name": config.StringVariable(sslName),
				},
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
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(kibana816Version),
				ResourceName:             sslHTTPMonitorID,
				ImportState:              true,
				ImportStateVerify:        true,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("http_ssl"),
				ConfigVariables: config.Variables{
					"name": config.StringVariable(sslName),
				},
			},
			// Create and Read http monitor
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minKibanaVersion),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("http_create"),
				ConfigVariables: config.Variables{
					"name": config.StringVariable(name),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(httpMonitorID, "id"),
					resource.TestCheckResourceAttr(httpMonitorID, "name", "TestHttpMonitorResource - "+name),
					resource.TestCheckResourceAttr(httpMonitorID, "space_id", ""),
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
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minKibanaVersion),
				ResourceName:             httpMonitorID,
				ImportState:              true,
				ImportStateVerify:        true,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("http_create"),
				ConfigVariables: config.Variables{
					"name": config.StringVariable(name),
				},
			},
			// Update and Read testing http monitor
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minKibanaVersion),
				ResourceName:             httpMonitorID,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("http_update"),
				ConfigVariables: config.Variables{
					"name": config.StringVariable(name),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(httpMonitorID, "id"),
					resource.TestCheckResourceAttr(httpMonitorID, "name", "TestHttpMonitorResource Updated - "+name),
					resource.TestCheckResourceAttr(httpMonitorID, "space_id", ""),
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

	tcpMonitorID := "elasticstack_kibana_synthetics_monitor.tcp-monitor"
	bmMonitorID := "elasticstack_kibana_synthetics_monitor.tcp-monitor-min"
	sslTCPMonitorID := "elasticstack_kibana_synthetics_monitor.tcp-monitor-ssl"

	name := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)
	bmName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)
	sslName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			// Create and Read tcp monitor with minimum fields
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minKibanaVersion),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("tcp_min"),
				ConfigVariables: config.Variables{
					"name": config.StringVariable(bmName),
				},
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
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(kibana816Version),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("tcp_ssl"),
				ConfigVariables: config.Variables{
					"name": config.StringVariable(sslName),
				},
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
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(kibana816Version),
				ResourceName:             sslTCPMonitorID,
				ImportState:              true,
				ImportStateVerify:        true,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("tcp_ssl"),
				ConfigVariables: config.Variables{
					"name": config.StringVariable(sslName),
				},
			},
			// Create and Read tcp monitor
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minKibanaVersion),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("tcp_create"),
				ConfigVariables: config.Variables{
					"name": config.StringVariable(name),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(tcpMonitorID, "id"),
					resource.TestCheckResourceAttr(tcpMonitorID, "name", "TestTcpMonitorResource - "+name),
					resource.TestCheckResourceAttr(tcpMonitorID, "space_id", ""),
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
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minKibanaVersion),
				ResourceName:             tcpMonitorID,
				ImportState:              true,
				ImportStateVerify:        true,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("tcp_create"),
				ConfigVariables: config.Variables{
					"name": config.StringVariable(name),
				},
			},
			// Update and Read tcp monitor
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minKibanaVersion),
				ResourceName:             tcpMonitorID,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("tcp_update"),
				ConfigVariables: config.Variables{
					"name": config.StringVariable(name),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(tcpMonitorID, "id"),
					resource.TestCheckResourceAttr(tcpMonitorID, "name", "TestTcpMonitorResource Updated - "+name),
					resource.TestCheckResourceAttr(tcpMonitorID, "space_id", ""),
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

	icmpMonitorID := "elasticstack_kibana_synthetics_monitor.icmp-monitor"
	bmMonitorID := "elasticstack_kibana_synthetics_monitor.icmp-monitor-min"

	name := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)
	bmName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			// Create and Read icmp monitor with minimum fields
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minKibanaVersion),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("icmp_min"),
				ConfigVariables: config.Variables{
					"name": config.StringVariable(bmName),
				},
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
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minKibanaVersion),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("icmp_create"),
				ConfigVariables: config.Variables{
					"name": config.StringVariable(name),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(icmpMonitorID, "id"),
					resource.TestCheckResourceAttr(icmpMonitorID, "name", "TestIcmpMonitorResource - "+name),
					resource.TestCheckResourceAttr(icmpMonitorID, "space_id", ""),
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
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minKibanaVersion),
				ResourceName:             icmpMonitorID,
				ImportState:              true,
				ImportStateVerify:        true,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("icmp_create"),
				ConfigVariables: config.Variables{
					"name": config.StringVariable(name),
				},
			},
			// Update and Read icmp monitor
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minKibanaVersion),
				ResourceName:             icmpMonitorID,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("icmp_update"),
				ConfigVariables: config.Variables{
					"name": config.StringVariable(name),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(icmpMonitorID, "id"),
					resource.TestCheckResourceAttr(icmpMonitorID, "name", "TestIcmpMonitorResource Updated - "+name),
					resource.TestCheckResourceAttr(icmpMonitorID, "space_id", ""),
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

	browserMonitorID := "elasticstack_kibana_synthetics_monitor.browser-monitor"
	bmMonitorID := "elasticstack_kibana_synthetics_monitor.browser-monitor-min"

	name := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)
	bmName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			// Create and Read browser monitor with minimum fields
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minKibanaVersion),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("browser_min"),
				ConfigVariables: config.Variables{
					"name": config.StringVariable(bmName),
				},
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
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minKibanaVersion),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("browser_create"),
				ConfigVariables: config.Variables{
					"name": config.StringVariable(name),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(browserMonitorID, "id"),
					resource.TestCheckResourceAttr(browserMonitorID, "name", "TestBrowserMonitorResource - "+name),
					resource.TestCheckResourceAttr(browserMonitorID, "space_id", ""),
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
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(kibana816Version),
				ResourceName:             browserMonitorID,
				ImportState:              true,
				ImportStateVerify:        true,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("browser_create"),
				ConfigVariables: config.Variables{
					"name": config.StringVariable(name),
				},
			},
			// Update and Read browser monitor
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minKibanaVersion),
				ResourceName:             browserMonitorID,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("browser_update"),
				ConfigVariables: config.Variables{
					"name": config.StringVariable(name),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(browserMonitorID, "id"),
					resource.TestCheckResourceAttr(browserMonitorID, "name", "TestBrowserMonitorResource Updated - "+name),
					resource.TestCheckResourceAttr(browserMonitorID, "space_id", ""),
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

func TestSyntheticMonitorHTTPResource_nonDefaultSpace(t *testing.T) {
	httpMonitorID := "elasticstack_kibana_synthetics_monitor.http-monitor"
	name := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)
	spaceID := sdkacctest.RandStringFromCharSet(12, accTestKibanaSpaceIDCharset)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minKibanaVersion),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("http_non_default_space"),
				ConfigVariables: config.Variables{
					"name":     config.StringVariable(name),
					"space_id": config.StringVariable(spaceID),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(httpMonitorID, "id"),
					resource.TestCheckResourceAttr(httpMonitorID, "name", "TestHttpMonitorResource - "+name),
					resource.TestCheckResourceAttr(httpMonitorID, "space_id", spaceID),
					resource.TestCheckResourceAttr(httpMonitorID, "namespace", "test_namespace"),
					resource.TestCheckResourceAttr(httpMonitorID, "schedule", "5"),
					resource.TestCheckResourceAttr(httpMonitorID, "private_locations.#", "1"),
					resource.TestCheckResourceAttrSet(httpMonitorID, "private_locations.0"),
					resource.TestCheckResourceAttr(httpMonitorID, "http.url", "http://localhost:5601"),
					resource.TestCheckResourceAttr(httpMonitorID, "http.mode", "any"),
				),
			},
		},
	})
}

func TestSyntheticMonitorTCPResource_nonDefaultSpace(t *testing.T) {
	tcpMonitorID := "elasticstack_kibana_synthetics_monitor.tcp-monitor"
	name := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)
	spaceID := sdkacctest.RandStringFromCharSet(12, accTestKibanaSpaceIDCharset)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minKibanaVersion),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("tcp_non_default_space"),
				ConfigVariables: config.Variables{
					"name":     config.StringVariable(name),
					"space_id": config.StringVariable(spaceID),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(tcpMonitorID, "id"),
					resource.TestCheckResourceAttr(tcpMonitorID, "name", "TestTcpMonitorResource - "+name),
					resource.TestCheckResourceAttr(tcpMonitorID, "space_id", spaceID),
					resource.TestCheckResourceAttr(tcpMonitorID, "namespace", "testacc_test"),
					resource.TestCheckResourceAttr(tcpMonitorID, "schedule", "5"),
					resource.TestCheckResourceAttr(tcpMonitorID, "private_locations.#", "1"),
					resource.TestCheckResourceAttrSet(tcpMonitorID, "private_locations.0"),
					resource.TestCheckResourceAttr(tcpMonitorID, "tcp.host", "http://localhost:5601"),
				),
			},
		},
	})
}

func TestSyntheticMonitorICMPResource_nonDefaultSpace(t *testing.T) {
	icmpMonitorID := "elasticstack_kibana_synthetics_monitor.icmp-monitor"
	name := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)
	spaceID := sdkacctest.RandStringFromCharSet(12, accTestKibanaSpaceIDCharset)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minKibanaVersion),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("icmp_non_default_space"),
				ConfigVariables: config.Variables{
					"name":     config.StringVariable(name),
					"space_id": config.StringVariable(spaceID),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(icmpMonitorID, "id"),
					resource.TestCheckResourceAttr(icmpMonitorID, "name", "TestIcmpMonitorResource - "+name),
					resource.TestCheckResourceAttr(icmpMonitorID, "space_id", spaceID),
					resource.TestCheckResourceAttr(icmpMonitorID, "namespace", "testacc_namespace"),
					resource.TestCheckResourceAttr(icmpMonitorID, "schedule", "5"),
					resource.TestCheckResourceAttr(icmpMonitorID, "private_locations.#", "1"),
					resource.TestCheckResourceAttrSet(icmpMonitorID, "private_locations.0"),
					resource.TestCheckResourceAttr(icmpMonitorID, "icmp.host", "localhost"),
				),
			},
		},
	})
}

func TestSyntheticMonitorBrowserResource_nonDefaultSpace(t *testing.T) {
	browserMonitorID := "elasticstack_kibana_synthetics_monitor.browser-monitor"
	name := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)
	spaceID := sdkacctest.RandStringFromCharSet(12, accTestKibanaSpaceIDCharset)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minKibanaVersion),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("browser_non_default_space"),
				ConfigVariables: config.Variables{
					"name":     config.StringVariable(name),
					"space_id": config.StringVariable(spaceID),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(browserMonitorID, "id"),
					resource.TestCheckResourceAttr(browserMonitorID, "name", "TestBrowserMonitorResource - "+name),
					resource.TestCheckResourceAttr(browserMonitorID, "space_id", spaceID),
					resource.TestCheckResourceAttr(browserMonitorID, "namespace", "testacc_ns"),
					resource.TestCheckResourceAttr(browserMonitorID, "schedule", "5"),
					resource.TestCheckResourceAttr(browserMonitorID, "private_locations.#", "1"),
					resource.TestCheckResourceAttrSet(browserMonitorID, "private_locations.0"),
					resource.TestCheckResourceAttr(browserMonitorID, "browser.inline_script", "step('Go to https://google.com.co', () => page.goto('https://www.google.com'))"),
				),
			},
		},
	})
}

func TestSyntheticMonitorLabelsResource(t *testing.T) {
	labelsMonitorID := "elasticstack_kibana_synthetics_monitor.http-monitor-labels"
	name := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			// Create and Read monitor with labels
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(monitor.MinLabelsVersion),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("labels_create"),
				ConfigVariables: config.Variables{
					"name": config.StringVariable(name),
				},
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
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(monitor.MinLabelsVersion),
				ResourceName:             labelsMonitorID,
				ImportState:              true,
				ImportStateVerify:        true,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("labels_create"),
				ConfigVariables: config.Variables{
					"name": config.StringVariable(name),
				},
			},
			// Update labels - change values but keep same keys
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(monitor.MinLabelsVersion),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("labels_update"),
				ConfigVariables: config.Variables{
					"name": config.StringVariable(name),
				},
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
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(monitor.MinLabelsVersion),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("labels_removed"),
				ConfigVariables: config.Variables{
					"name": config.StringVariable(name),
				},
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
