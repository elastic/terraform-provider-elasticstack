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

package dashboard_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanautil"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stretchr/testify/require"
)

// lazyStringVar is a config.Variable whose JSON value is resolved when Variables are
// marshaled (after TestCase.PreCheck), so bootstrap values like a saved-object id can
// be filled in by PreCheck before each step runs.
type lazyStringVar struct {
	p *string
}

func (v lazyStringVar) MarshalJSON() ([]byte, error) {
	if v.p == nil {
		return json.Marshal("")
	}
	return json.Marshal(*v.p)
}

func testAccCheckMarkdownByReferenceRefID(idPtr *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if idPtr == nil || *idPtr == "" {
			return fmt.Errorf("expected non-empty markdown library id")
		}
		return resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.markdown_config.by_reference.ref_id", *idPtr)(s)
	}
}

func TestAccResourceDashboardMarkdownByReference(t *testing.T) {
	var markdownLibID string
	markdownLibIDPtr := &markdownLibID

	dashboardTitle := "Acc md by-ref " + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			// Kibana only registers the `markdown` saved-object type on stacks where this
			// suite's dashboard API is exercised (see minDashboardAPISupport). Older
			// versions return 400 "Unsupported saved object type: 'markdown'".
			//
			// terraform-plugin-testing runs each step's PreConfig before SkipFunc, so
			// bootstrap in PreConfig cannot be gated by step SkipFunc alone; skip here
			// (TestCase PreCheck runs before any steps).
			if skip, err := versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport)(); err != nil {
				t.Fatalf("markdown by-reference acceptance test version check: %v", err)
			} else if skip {
				t.Skipf(
					"Skipping test: stack version is below %s (required for `markdown` saved-object type)",
					minDashboardAPISupport,
				)
			}
		},
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				PreConfig: func() {
					if markdownLibID != "" {
						return
					}
					markdownLibID = createAccMarkdownLibrarySavedObject(t)
					t.Cleanup(func() { deleteAccMarkdownLibrarySavedObject(t, markdownLibID) })
				},
				ConfigDirectory: acctest.NamedTestCaseDirectory("basic"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
					"markdown_lib_id": lazyStringVar{p: markdownLibIDPtr},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.type", "markdown"),
					testAccCheckMarkdownByReferenceRefID(markdownLibIDPtr),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.markdown_config.by_reference.title", "Overlay title for library markdown"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.markdown_config.by_reference.description", "Overlay description for by-reference panel"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.markdown_config.by_reference.hide_title", "true"),
					resource.TestCheckResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.markdown_config.by_reference.hide_border", "false"),
					resource.TestCheckNoResourceAttr("elasticstack_kibana_dashboard.test", "panels.0.markdown_config.by_value"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minDashboardAPISupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("basic"),
				ConfigVariables: config.Variables{
					"dashboard_title": config.StringVariable(dashboardTitle),
					"markdown_lib_id": lazyStringVar{p: markdownLibIDPtr},
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
				},
			},
		},
	})
}

func ptrSavedObjectAttr(s string) *any {
	v := any(s)
	return &v
}

func createAccMarkdownLibrarySavedObject(t *testing.T) string {
	t.Helper()

	suffix := sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlphaNum)
	title := "tf-acc-md-lib-" + suffix

	client, err := clients.NewAcceptanceTestingKibanaScopedClient()
	require.NoError(t, err)
	oapi, err := client.GetKibanaOapiClient()
	require.NoError(t, err)

	body := kbapi.PostSavedObjectsTypeJSONRequestBody{
		Attributes: map[string]*any{
			"title":   ptrSavedObjectAttr(title),
			"content": ptrSavedObjectAttr("# Acceptance markdown library item\n\n_" + suffix + "_\n"),
		},
	}

	resp, err := oapi.API.PostSavedObjectsTypeWithResponse(
		context.Background(),
		"markdown",
		&kbapi.PostSavedObjectsTypeParams{},
		body,
		kibanautil.SpaceAwarePathRequestEditor(""),
	)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equalf(t, http.StatusOK, resp.StatusCode(),
		"create markdown library saved object: %s", string(resp.Body))

	var parsed struct {
		ID string `json:"id"`
	}
	require.NoError(t, json.Unmarshal(resp.Body, &parsed))
	require.NotEmpty(t, parsed.ID, string(resp.Body))
	return parsed.ID
}

func deleteAccMarkdownLibrarySavedObject(t *testing.T, id string) {
	t.Helper()
	if id == "" {
		return
	}
	client, err := clients.NewAcceptanceTestingKibanaScopedClient()
	if err != nil {
		t.Logf("markdown library cleanup: no client: %v", err)
		return
	}
	oapi, err := client.GetKibanaOapiClient()
	if err != nil {
		t.Logf("markdown library cleanup: no oapi client: %v", err)
		return
	}
	resp, err := oapi.API.DeleteSavedObjectsTypeIdWithResponse(
		context.Background(),
		"markdown",
		id,
		&kbapi.DeleteSavedObjectsTypeIdParams{},
		kibanautil.SpaceAwarePathRequestEditor(""),
	)
	if err != nil {
		t.Logf("markdown library cleanup delete: %v", err)
		return
	}
	if resp.StatusCode() != http.StatusOK {
		t.Logf("markdown library cleanup: unexpected status %d: %s", resp.StatusCode(), string(resp.Body))
	}
}
