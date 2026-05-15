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
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

const testAccFieldAttrsDataViewAddress = "elasticstack_kibana_data_view.fa_dv"

// resolveFieldAttrsDataView returns the OAPI client and parsed (spaceID, viewID) for the
// managed data view, so per-test helpers don't have to repeat the same plumbing.
func resolveFieldAttrsDataView(s *terraform.State) (*kibanaoapi.Client, string, string, error) {
	rs := s.RootModule().Resources[testAccFieldAttrsDataViewAddress]
	if rs == nil {
		return nil, "", "", fmt.Errorf("%s not found in state", testAccFieldAttrsDataViewAddress)
	}
	composite, diags := clients.CompositeIDFromStr(rs.Primary.ID)
	if diags.HasError() {
		return nil, "", "", fmt.Errorf("parse data view id: %v", diags)
	}
	apiClient, err := clients.NewAcceptanceTestingKibanaScopedClient()
	if err != nil {
		return nil, "", "", fmt.Errorf("acceptance kibana client: %w", err)
	}
	kc, err := apiClient.GetKibanaOapiClient()
	if err != nil {
		return nil, "", "", fmt.Errorf("kibana openapi client: %w", err)
	}
	return kc, composite.ClusterID, composite.ResourceID, nil
}

// testAccInjectHostHostnameFieldCount writes a server-side popularity count for host.hostname
// (simulates Discover usage) by reusing the production UpdateFieldMetadata wrapper, so this
// helper cannot drift from the space-aware path / payload shape used at runtime.
func testAccInjectHostHostnameFieldCount(t *testing.T, s *terraform.State) error {
	t.Helper()
	client, spaceID, viewID, err := resolveFieldAttrsDataView(s)
	if err != nil {
		return err
	}
	diags := kibanaoapi.UpdateFieldMetadata(context.Background(), client, spaceID, viewID, map[string]any{
		"host.hostname": map[string]any{"count": 5},
	})
	if diags.HasError() {
		return fmt.Errorf("inject host.hostname count: %v", diags)
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

// testAccCheckFieldAttrsCustomLabelServerSide queries Kibana via GetDataView and verifies the
// requested field's customLabel matches `want`. Pass want="" to assert the entry is either
// absent or has a null customLabel (i.e. it was actually cleared server-side, not just dropped
// from Terraform state).
func testAccCheckFieldAttrsCustomLabelServerSide(t *testing.T, fieldKey, want string) resource.TestCheckFunc {
	t.Helper()
	return func(s *terraform.State) error {
		client, spaceID, viewID, err := resolveFieldAttrsDataView(s)
		if err != nil {
			return err
		}
		view, diags := kibanaoapi.GetDataView(context.Background(), client, spaceID, viewID)
		if diags.HasError() {
			return fmt.Errorf("get data view %s/%s: %v", spaceID, viewID, diags)
		}
		if view == nil {
			return fmt.Errorf("data view %s/%s not found", spaceID, viewID)
		}

		var entry struct {
			label   *string
			present bool
		}
		if view.DataView.FieldAttrs != nil {
			fa, ok := (*view.DataView.FieldAttrs)[fieldKey]
			entry.present = ok
			if ok {
				entry.label = fa.CustomLabel
			}
		}

		if want == "" {
			if !entry.present || entry.label == nil || *entry.label == "" {
				return nil
			}
			return fmt.Errorf("server-side field_attrs[%q].customLabel = %q, want cleared", fieldKey, *entry.label)
		}
		if !entry.present {
			return fmt.Errorf("server-side field_attrs[%q] not found (want customLabel=%q)", fieldKey, want)
		}
		if entry.label == nil {
			return fmt.Errorf("server-side field_attrs[%q].customLabel is null (want %q)", fieldKey, want)
		}
		if *entry.label != want {
			return fmt.Errorf("server-side field_attrs[%q].customLabel = %q, want %q", fieldKey, *entry.label, want)
		}
		return nil
	}
}
