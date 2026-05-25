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

package dataview_test

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stretchr/testify/require"
)

var minDataViewAPISupport = version.Must(version.NewVersion("8.1.0"))
var minFullDataviewSupport = version.Must(version.NewVersion("8.8.0"))

func TestAccResourceDataView(t *testing.T) {
	indexName := "my-index-" + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)

	var dataViewID string
	captureID := func(s *terraform.State) error {
		rs := s.RootModule().Resources["elasticstack_kibana_data_view.dv"]
		if rs == nil {
			return fmt.Errorf("elasticstack_kibana_data_view.dv not found in state")
		}
		dataViewID = rs.Primary.ID
		return nil
	}
	checkIDUnchanged := func(s *terraform.State) error {
		rs := s.RootModule().Resources["elasticstack_kibana_data_view.dv"]
		if rs == nil {
			return fmt.Errorf("elasticstack_kibana_data_view.dv not found in state")
		}
		if rs.Primary.ID != dataViewID {
			return fmt.Errorf("data view was recreated: id changed from %s to %s", dataViewID, rs.Primary.ID)
		}
		return nil
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDataViewAPISupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("pre_8_8"),
				ConfigVariables: config.Variables{
					"index_name": config.StringVariable(indexName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("elasticstack_kibana_data_view.dv", "id"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minFullDataviewSupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("basic"),
				ConfigVariables: config.Variables{
					"index_name": config.StringVariable(indexName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("elasticstack_kibana_data_view.dv", "id"),
					resource.TestCheckResourceAttr("elasticstack_kibana_data_view.dv", "override", "true"),
					resource.TestCheckResourceAttr("elasticstack_kibana_data_view.dv", "data_view.title", indexName+"*"),
					resource.TestCheckResourceAttr("elasticstack_kibana_data_view.dv", "data_view.name", indexName),
					resource.TestCheckResourceAttr("elasticstack_kibana_data_view.dv", "data_view.time_field_name", "@timestamp"),
					resource.TestCheckResourceAttr("elasticstack_kibana_data_view.dv", "data_view.allow_no_index", "true"),
					resource.TestCheckResourceAttr("elasticstack_kibana_data_view.dv", "data_view.source_filters.#", "2"),
					resource.TestCheckResourceAttr("elasticstack_kibana_data_view.dv", "data_view.source_filters.0", "event_time"),
					resource.TestCheckResourceAttr("elasticstack_kibana_data_view.dv", "data_view.source_filters.1", "machine.ram"),
					resource.TestCheckResourceAttr("elasticstack_kibana_data_view.dv", "data_view.field_formats.event_time.id", "date_nanos"),
					resource.TestCheckResourceAttr("elasticstack_kibana_data_view.dv", "data_view.field_formats.machine.ram.params.pattern", "0,0.[000] b"),
					resource.TestCheckResourceAttr("elasticstack_kibana_data_view.dv", "data_view.runtime_field_map.runtime_shape_name.type", "keyword"),
					resource.TestCheckResourceAttr("elasticstack_kibana_data_view.dv", "data_view.runtime_field_map.runtime_shape_name.script_source", "emit(doc['shape_name'].value)"),
					resource.TestCheckResourceAttr("elasticstack_kibana_data_view.dv", "data_view.field_attrs.ingest_failure.custom_label", "error.ingest_failure"),
					resource.TestCheckResourceAttr("elasticstack_kibana_data_view.dv", "data_view.field_attrs.ingest_failure.count", "6"),
					captureID,
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minFullDataviewSupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("basic_updated"),
				ConfigVariables: config.Variables{
					"index_name": config.StringVariable(indexName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("elasticstack_kibana_data_view.dv", "id"),
					resource.TestCheckResourceAttr("elasticstack_kibana_data_view.dv", "override", "false"),
					resource.TestCheckResourceAttr("elasticstack_kibana_data_view.dv", "data_view.name", indexName),
					resource.TestCheckResourceAttr("elasticstack_kibana_data_view.dv", "data_view.source_filters.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_kibana_data_view.dv", "data_view.field_formats.event_time.id", "date_nanos"),
					resource.TestCheckNoResourceAttr("elasticstack_kibana_data_view.dv", "data_view.field_formats.machine.ram.%"),
					resource.TestCheckResourceAttr("elasticstack_kibana_data_view.dv", "data_view.runtime_field_map.runtime_shape_name.script_source", "emit(doc['shape_name'].value)"),
					checkIDUnchanged,
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minFullDataviewSupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("basic_omitted"),
				ConfigVariables: config.Variables{
					"index_name": config.StringVariable(indexName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("elasticstack_kibana_data_view.dv", "id"),
					resource.TestCheckResourceAttr("elasticstack_kibana_data_view.dv", "override", "false"),
					resource.TestCheckResourceAttr("elasticstack_kibana_data_view.dv", "data_view.name", indexName),
					resource.TestCheckResourceAttr("elasticstack_kibana_data_view.dv", "data_view.source_filters.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_kibana_data_view.dv", "data_view.field_formats.event_time.id", "date_nanos"),
					resource.TestCheckResourceAttr("elasticstack_kibana_data_view.dv", "data_view.runtime_field_map.runtime_shape_name.script_source", "emit(doc['shape_name'].value)"),
					checkIDUnchanged,
				),
			},
			// Re-apply the same omitted config for import-state verification.
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minFullDataviewSupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("basic_omitted"),
				ConfigVariables: config.Variables{
					"index_name": config.StringVariable(indexName),
				},
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"data_view.runtime_field_map", "data_view.field_formats", "data_view.source_filters"},
				ResourceName:            "elasticstack_kibana_data_view.dv",
			},
		},
	})
}

func TestAccResourceDataViewColorFieldFormat(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minFullDataviewSupport, versionutils.FlavorAny)

	indexName := "my-color-index-" + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"index_name": config.StringVariable(indexName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("elasticstack_kibana_data_view.color_dv", "id"),
					resource.TestCheckResourceAttr("elasticstack_kibana_data_view.color_dv", "data_view.field_formats.status.id", "color"),
					resource.TestCheckResourceAttr("elasticstack_kibana_data_view.color_dv", "data_view.field_formats.status.params.field_type", "string"),
					resource.TestCheckResourceAttr("elasticstack_kibana_data_view.color_dv", "data_view.field_formats.status.params.colors.#", "2"),
					resource.TestCheckResourceAttr("elasticstack_kibana_data_view.color_dv", "data_view.field_formats.status.params.colors.0.range", "-Infinity:Infinity"),
					resource.TestCheckResourceAttr("elasticstack_kibana_data_view.color_dv", "data_view.field_formats.status.params.colors.0.regex", "Completed"),
					resource.TestCheckResourceAttr("elasticstack_kibana_data_view.color_dv", "data_view.field_formats.status.params.colors.0.text", "#000000"),
					resource.TestCheckResourceAttr("elasticstack_kibana_data_view.color_dv", "data_view.field_formats.status.params.colors.0.background", "#54B399"),
					resource.TestCheckResourceAttr("elasticstack_kibana_data_view.color_dv", "data_view.field_formats.status.params.colors.1.regex", "Error"),
					resource.TestCheckResourceAttr("elasticstack_kibana_data_view.color_dv", "data_view.field_formats.status.params.colors.1.text", "#FFFFFF"),
					resource.TestCheckResourceAttr("elasticstack_kibana_data_view.color_dv", "data_view.field_formats.status.params.colors.1.background", "#BD271E"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("import"),
				ConfigVariables: config.Variables{
					"index_name": config.StringVariable(indexName),
				},
				ImportState: true,
				ImportStateVerifyIgnore: []string{
					"override",
				},
				ImportStateVerify: true,
				ResourceName:      "elasticstack_kibana_data_view.color_dv",
			},
		},
	})
}

func TestAccResourceDataViewCreateErrorRecovery(t *testing.T) {
	if os.Getenv("TF_ACC") != "1" {
		t.Skip("acceptance tests skipped unless TF_ACC=1")
	}

	acctest.PreCheck(t)
	unsupported, err := versionutils.CheckIfVersionIsUnsupported(minFullDataviewSupport)()
	require.NoError(t, err)
	if unsupported {
		t.Skipf("data view create recovery requires stack version %s or later", minFullDataviewSupport)
	}

	acctest.PreCheckWithExplicitKibanaEndpoint(t)

	indexName := "my-index-" + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)
	spaceID := "test-space-" + sdkacctest.RandStringFromCharSet(6, sdkacctest.CharSetAlphaNum)
	dataViewID := "test-data-view-" + sdkacctest.RandStringFromCharSet(6, sdkacctest.CharSetAlphaNum)

	proxyServer, createFailures := testAccDataViewCreateErrorProxy(t, os.Getenv("KIBANA_ENDPOINT"), spaceID)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"index_name":      config.StringVariable(indexName),
					"space_id":        config.StringVariable(spaceID),
					"data_view_id":    config.StringVariable(dataViewID),
					"kibana_endpoint": config.StringVariable(proxyServer.URL),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_data_view.dv", "id", fmt.Sprintf("%s/%s", spaceID, dataViewID)),
					resource.TestCheckResourceAttr("elasticstack_kibana_data_view.dv", "space_id", spaceID),
					resource.TestCheckResourceAttr("elasticstack_kibana_data_view.dv", "data_view.id", dataViewID),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"index_name":      config.StringVariable(indexName),
					"space_id":        config.StringVariable(spaceID),
					"data_view_id":    config.StringVariable(dataViewID),
					"kibana_endpoint": config.StringVariable(proxyServer.URL),
				},
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})

	require.Equal(t, int32(1), createFailures.Load())
}

func testAccDataViewCreateErrorProxy(t *testing.T, upstreamEndpoint, spaceID string) (*httptest.Server, *atomic.Int32) {
	targetURL, err := url.Parse(upstreamEndpoint)
	require.NoError(t, err)

	createFailures := &atomic.Int32{}
	proxy := httputil.NewSingleHostReverseProxy(targetURL)
	proxy.ModifyResponse = func(resp *http.Response) error {
		if resp.Request.Method != http.MethodPost || !testAccIsDataViewCreatePath(resp.Request.URL.Path, spaceID) {
			return nil
		}
		if !createFailures.CompareAndSwap(0, 1) {
			return nil
		}

		if resp.Body != nil {
			_ = resp.Body.Close()
		}

		body := `{"statusCode":400,"error":"Bad Request","message":"synthetic create failure after upstream persistence"}`
		resp.StatusCode = http.StatusBadRequest
		resp.Status = http.StatusText(http.StatusBadRequest)
		resp.Body = io.NopCloser(strings.NewReader(body))
		resp.ContentLength = int64(len(body))
		resp.Header.Del("Content-Encoding")
		resp.Header.Set("Content-Length", strconv.Itoa(len(body)))
		resp.Header.Set("Content-Type", "application/json")
		return nil
	}

	server := httptest.NewServer(proxy)
	t.Cleanup(server.Close)
	return server, createFailures
}

func testAccIsDataViewCreatePath(path, spaceID string) bool {
	return path == "/api/data_views/data_view" || path == fmt.Sprintf("/s/%s/api/data_views/data_view", spaceID)
}

func TestAccResourceDataViewNamespaces(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minFullDataviewSupport, versionutils.FlavorAny)

	indexName := "ns-test-" + sdkacctest.RandStringFromCharSet(6, sdkacctest.CharSetAlphaNum)
	space1 := "space-a-" + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)
	space2 := "space-b-" + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)
	space3 := "space-c-" + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)

	vars := config.Variables{
		"index_name": config.StringVariable(indexName),
		"space1":     config.StringVariable(space1),
		"space2":     config.StringVariable(space2),
		"space3":     config.StringVariable(space3),
	}
	var dataViewID string
	captureID := func(s *terraform.State) error {
		rs := s.RootModule().Resources["elasticstack_kibana_data_view.ns_dv"]
		if rs == nil {
			return fmt.Errorf("elasticstack_kibana_data_view.ns_dv not found in state")
		}
		dataViewID = rs.Primary.ID
		return nil
	}
	checkIDUnchanged := func(s *terraform.State) error {
		rs := s.RootModule().Resources["elasticstack_kibana_data_view.ns_dv"]
		if rs == nil {
			return fmt.Errorf("elasticstack_kibana_data_view.ns_dv not found in state")
		}
		if rs.Primary.ID != dataViewID {
			return fmt.Errorf("data view was recreated: id changed from %s to %s", dataViewID, rs.Primary.ID)
		}
		return nil
	}
	checkNamespacesMembership := func(expected ...string) resource.TestCheckFunc {
		checks := []resource.TestCheckFunc{
			resource.TestCheckResourceAttr("elasticstack_kibana_data_view.ns_dv", "space_id", space1),
			resource.TestCheckResourceAttr("elasticstack_kibana_data_view.ns_dv", "data_view.namespaces.#", fmt.Sprintf("%d", len(expected))),
		}
		for _, ns := range expected {
			checks = append(checks, resource.TestCheckTypeSetElemAttr(
				"elasticstack_kibana_data_view.ns_dv", "data_view.namespaces.*", ns))
		}
		return resource.ComposeTestCheckFunc(checks...)
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("initial"),
				ConfigVariables:          vars,
				Check: resource.ComposeTestCheckFunc(
					checkNamespacesMembership(space1, space2, "default"),
					captureID,
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("add_space"),
				ConfigVariables:          vars,
				Check: resource.ComposeTestCheckFunc(
					checkNamespacesMembership(space1, space2, "default", space3),
					checkIDUnchanged,
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("remove_space"),
				ConfigVariables:          vars,
				Check: resource.ComposeTestCheckFunc(
					checkNamespacesMembership(space1, space2, space3),
					checkIDUnchanged,
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("add_remove_space"),
				ConfigVariables:          vars,
				Check: resource.ComposeTestCheckFunc(
					checkNamespacesMembership(space1, "default", space3),
					checkIDUnchanged,
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("add_remove_space"),
				ConfigVariables:          vars,
				PlanOnly:                 true,
				ExpectNonEmptyPlan:       false,
			},
		},
	})
}

// TestAccResourceDataViewFieldAttrs covers fix-dataview-field-attrs-drift task 6 (REQ-006 "No
// replacement on field_attrs change", REQ-015 stability, REQ-016 in-place updates).
//
// REQ-015 scenario 1 (server-side count does not drift plan): after the first apply we POST
// field popularity for host.hostname via the Kibana HTTP API (same endpoint family as
// UpdateFieldMetadata) and then assert a PlanOnly step is empty. Failing to inject the count
// fails the test, since suppressing that drift is exactly the behaviour we are exercising.
func TestAccResourceDataViewFieldAttrs(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minFullDataviewSupport, versionutils.FlavorAny)

	indexName := "fa-test-" + sdkacctest.RandStringFromCharSet(6, sdkacctest.CharSetAlphaNum)
	vars := config.Variables{
		"index_name": config.StringVariable(indexName),
	}

	var dataViewID string
	captureID := func(s *terraform.State) error {
		rs := s.RootModule().Resources[testAccFieldAttrsDataViewAddress]
		if rs == nil {
			return fmt.Errorf("%s not found in state", testAccFieldAttrsDataViewAddress)
		}
		dataViewID = rs.Primary.ID
		return nil
	}
	checkIDUnchanged := func(s *terraform.State) error {
		rs := s.RootModule().Resources[testAccFieldAttrsDataViewAddress]
		if rs == nil {
			return fmt.Errorf("%s not found in state", testAccFieldAttrsDataViewAddress)
		}
		if rs.Primary.ID != dataViewID {
			return fmt.Errorf("data view was recreated: id changed from %s to %s", dataViewID, rs.Primary.ID)
		}
		return nil
	}

	injectHostHostnameCount := func(s *terraform.State) error {
		return testAccInjectHostHostnameFieldCount(t, s)
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("no_field_attrs"),
				ConfigVariables:          vars,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(testAccFieldAttrsDataViewAddress, "id"),
					resource.TestCheckNoResourceAttr(testAccFieldAttrsDataViewAddress, "data_view.field_attrs.%"),
					captureID,
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("no_field_attrs"),
				ConfigVariables:          vars,
				Check:                    injectHostHostnameCount,
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("no_field_attrs"),
				ConfigVariables:          vars,
				PlanOnly:                 true,
				ExpectNonEmptyPlan:       false,
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("add_label"),
				ConfigVariables:          vars,
				Check: resource.ComposeTestCheckFunc(
					checkIDUnchanged,
					testAccCheckFieldAttrsCustomLabel("host.hostname", "Host"),
					testAccCheckFieldAttrsCustomLabelServerSide(t, "host.hostname", "Host"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("with_count"),
				ConfigVariables:          vars,
				Check: resource.ComposeTestCheckFunc(
					checkIDUnchanged,
					testAccCheckFieldAttrsCustomLabel("host.hostname", "Host"),
					testAccCheckFieldAttrsCount("host.hostname", 5),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("change_label"),
				ConfigVariables:          vars,
				Check: resource.ComposeTestCheckFunc(
					checkIDUnchanged,
					testAccCheckFieldAttrsCustomLabel("host.hostname", "Hostname"),
					testAccCheckFieldAttrsCustomLabelServerSide(t, "host.hostname", "Hostname"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("remove_label"),
				ConfigVariables:          vars,
				Check: resource.ComposeTestCheckFunc(
					checkIDUnchanged,
					resource.TestCheckNoResourceAttr(testAccFieldAttrsDataViewAddress, "data_view.field_attrs.%"),
					testAccCheckFieldAttrsCustomLabelServerSide(t, "host.hostname", ""),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("remove_label"),
				ConfigVariables:          vars,
				PlanOnly:                 true,
				ExpectNonEmptyPlan:       false,
			},
		},
	})
}

func TestAccResourceDataViewDurationFieldFormat(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minFullDataviewSupport, versionutils.FlavorAny)

	indexName := "my-duration-index-" + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"index_name": config.StringVariable(indexName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("elasticstack_kibana_data_view.duration_dv", "id"),
					resource.TestCheckResourceAttr("elasticstack_kibana_data_view.duration_dv", "data_view.field_formats.response_time.id", "duration"),
					resource.TestCheckResourceAttr("elasticstack_kibana_data_view.duration_dv", "data_view.field_formats.response_time.params.input_format", "milliseconds"),
					resource.TestCheckResourceAttr("elasticstack_kibana_data_view.duration_dv", "data_view.field_formats.response_time.params.output_format", "humanizePrecise"),
					resource.TestCheckResourceAttr("elasticstack_kibana_data_view.duration_dv", "data_view.field_formats.response_time.params.output_precision", "2"),
					resource.TestCheckResourceAttr("elasticstack_kibana_data_view.duration_dv", "data_view.field_formats.response_time.params.include_space_with_suffix", "true"),
					resource.TestCheckResourceAttr("elasticstack_kibana_data_view.duration_dv", "data_view.field_formats.response_time.params.use_short_suffix", "false"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"index_name": config.StringVariable(indexName),
				},
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"override"},
				ResourceName:            "elasticstack_kibana_data_view.duration_dv",
			},
		},
	})
}

func TestAccResourceDataViewURLFieldFormat(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minFullDataviewSupport, versionutils.FlavorAny)

	indexName := "my-url-index-" + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"index_name": config.StringVariable(indexName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("elasticstack_kibana_data_view.url_dv", "id"),
					resource.TestCheckResourceAttr("elasticstack_kibana_data_view.url_dv", "data_view.field_formats.thumbnail.id", "url"),
					resource.TestCheckResourceAttr("elasticstack_kibana_data_view.url_dv", "data_view.field_formats.thumbnail.params.type", "img"),
					resource.TestCheckResourceAttr("elasticstack_kibana_data_view.url_dv", "data_view.field_formats.thumbnail.params.urltemplate", "https://example.com/images/{{value}}"),
					resource.TestCheckResourceAttr("elasticstack_kibana_data_view.url_dv", "data_view.field_formats.thumbnail.params.labeltemplate", "Image: {{value}}"),
					resource.TestCheckResourceAttr("elasticstack_kibana_data_view.url_dv", "data_view.field_formats.thumbnail.params.width", "200"),
					resource.TestCheckResourceAttr("elasticstack_kibana_data_view.url_dv", "data_view.field_formats.thumbnail.params.height", "150"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"index_name": config.StringVariable(indexName),
				},
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"override"},
				ResourceName:            "elasticstack_kibana_data_view.url_dv",
			},
		},
	})
}

func TestAccResourceDataViewStaticLookupFieldFormat(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minFullDataviewSupport, versionutils.FlavorAny)

	indexName := "my-lookup-index-" + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"index_name": config.StringVariable(indexName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("elasticstack_kibana_data_view.lookup_dv", "id"),
					resource.TestCheckResourceAttr("elasticstack_kibana_data_view.lookup_dv", "data_view.field_formats.status_code.id", "static_lookup"),
					resource.TestCheckResourceAttr("elasticstack_kibana_data_view.lookup_dv", "data_view.field_formats.status_code.params.lookup_entries.#", "2"),
					resource.TestCheckResourceAttr("elasticstack_kibana_data_view.lookup_dv", "data_view.field_formats.status_code.params.lookup_entries.0.key", "200"),
					resource.TestCheckResourceAttr("elasticstack_kibana_data_view.lookup_dv", "data_view.field_formats.status_code.params.lookup_entries.0.value", "OK"),
					resource.TestCheckResourceAttr("elasticstack_kibana_data_view.lookup_dv", "data_view.field_formats.status_code.params.lookup_entries.1.key", "404"),
					resource.TestCheckResourceAttr("elasticstack_kibana_data_view.lookup_dv", "data_view.field_formats.status_code.params.lookup_entries.1.value", "Not Found"),
					resource.TestCheckResourceAttr("elasticstack_kibana_data_view.lookup_dv", "data_view.field_formats.status_code.params.unknown_key_value", "Unknown"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"index_name": config.StringVariable(indexName),
				},
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"override"},
				ResourceName:            "elasticstack_kibana_data_view.lookup_dv",
			},
		},
	})
}
