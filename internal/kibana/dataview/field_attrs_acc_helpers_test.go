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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

const testAccFieldAttrsDataViewAddress = "elasticstack_kibana_data_view.fa_dv"

// testAccInjectHostHostnameFieldCount POSTs field metadata so Kibana records a server-side
// popularity count for host.hostname (simulates Discover usage). REQ-015 scenario 1.
func testAccInjectHostHostnameFieldCount(t *testing.T, s *terraform.State) error {
	t.Helper()

	rs := s.RootModule().Resources[testAccFieldAttrsDataViewAddress]
	if rs == nil {
		return fmt.Errorf("%s not found in state", testAccFieldAttrsDataViewAddress)
	}

	composite, diags := clients.CompositeIDFromStr(rs.Primary.ID)
	if diags.HasError() {
		return fmt.Errorf("parse data view id: %v", diags)
	}

	apiClient, err := clients.NewAcceptanceTestingKibanaScopedClient()
	if err != nil {
		return fmt.Errorf("acceptance kibana client: %w", err)
	}
	kc, err := apiClient.GetKibanaOapiClient()
	if err != nil {
		return fmt.Errorf("kibana openapi client: %w", err)
	}

	spaceID := composite.ClusterID
	viewID := composite.ResourceID

	path := fmt.Sprintf("/api/data_views/data_view/%s/fields", url.PathEscape(viewID))
	if spaceID != "" && spaceID != "default" {
		path = fmt.Sprintf("/s/%s/api/data_views/data_view/%s/fields", url.PathEscape(spaceID), url.PathEscape(viewID))
	}

	payload := map[string]any{
		"fields": map[string]any{
			"host.hostname": map[string]any{
				"count": 5,
			},
		},
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal field metadata body: %w", err)
	}

	endpoint := strings.TrimRight(kc.URL, "/") + path
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, endpoint, bytes.NewReader(raw))
	if err != nil {
		return fmt.Errorf("build POST %s: %w", endpoint, err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("kbn-xsrf", "true")

	resp, err := kc.HTTP.Do(req)
	if err != nil {
		return fmt.Errorf("POST %s: %w", endpoint, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("POST %s: status %d: %s", endpoint, resp.StatusCode, string(body))
	}

	return nil
}

// testAccCheckFieldAttrsCustomLabel matches the flattened state key for a dynamic map entry whose
// key may contain dots (e.g. "host.hostname").
func testAccCheckFieldAttrsCustomLabel(fieldKey, want string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs := s.RootModule().Resources[testAccFieldAttrsDataViewAddress]
		if rs == nil {
			return fmt.Errorf("%s not found in state", testAccFieldAttrsDataViewAddress)
		}
		const prefix = "data_view.field_attrs."
		const suffix = ".custom_label"
		for k, v := range rs.Primary.Attributes {
			if !strings.HasPrefix(k, prefix) || !strings.HasSuffix(k, suffix) {
				continue
			}
			mid := strings.TrimSuffix(strings.TrimPrefix(k, prefix), suffix)
			if mid == fieldKey {
				if v == want {
					return nil
				}
				return fmt.Errorf("%s: custom_label is %q, want %q", k, v, want)
			}
		}
		return fmt.Errorf("no field_attrs[%q].custom_label in state (want %q)", fieldKey, want)
	}
}
