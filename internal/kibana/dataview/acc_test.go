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
	"github.com/stretchr/testify/require"
)

var minDataViewAPISupport = version.Must(version.NewVersion("8.1.0"))
var minFullDataviewSupport = version.Must(version.NewVersion("8.8.0"))

func TestAccResourceDataView(t *testing.T) {
	indexName := "my-index-" + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)

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
					resource.TestCheckResourceAttr("elasticstack_kibana_data_view.dv", "data_view.name", indexName),
					resource.TestCheckResourceAttr("elasticstack_kibana_data_view.dv", "data_view.source_filters.#", "2"),
					resource.TestCheckResourceAttr("elasticstack_kibana_data_view.dv", "data_view.field_formats.event_time.id", "date_nanos"),
					resource.TestCheckResourceAttr("elasticstack_kibana_data_view.dv", "data_view.field_formats.machine.ram.params.pattern", "0,0.[000] b"),
					resource.TestCheckResourceAttr("elasticstack_kibana_data_view.dv", "data_view.runtime_field_map.runtime_shape_name.script_source", "emit(doc['shape_name'].value)"),
					resource.TestCheckResourceAttr("elasticstack_kibana_data_view.dv", "data_view.field_attrs.ingest_failure.custom_label", "error.ingest_failure"),
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
					resource.TestCheckNoResourceAttr("elasticstack_kibana_data_view.dv", "data_view.source_filters"),
					resource.TestCheckNoResourceAttr("elasticstack_kibana_data_view.dv", "data_view.field_formats"),
					resource.TestCheckNoResourceAttr("elasticstack_kibana_data_view.dv", "data_view.runtime_field_map"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minFullDataviewSupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("basic_updated"),
				ConfigVariables: config.Variables{
					"index_name": config.StringVariable(indexName),
				},
				ImportState:       true,
				ImportStateVerify: true,
				ResourceName:      "elasticstack_kibana_data_view.dv",
			},
		},
	})
}

func TestAccResourceDataViewColorFieldFormat(t *testing.T) {
	indexName := "my-color-index-" + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minFullDataviewSupport),
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
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minFullDataviewSupport),
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
