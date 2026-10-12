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

package outputds

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestFromAPIModel_CommonFields verifies that the per-type fromAPI*Model
// adapters all map their shared fields through fromAPICommonFields the same
// way, regardless of which generated kbapi response type they receive.
func TestFromAPIModel_CommonFields(t *testing.T) {
	t.Parallel()

	id := "output-id"
	caSha256 := "ca-sha-256"
	caTrustedFingerprint := "ca-trusted-fingerprint"
	isDefault := true
	isDefaultMonitoring := true
	configYaml := "config: yaml"
	hosts := []string{"https://example.com:9200"}

	want := outputItemModel{
		ID:                   types.StringValue(id),
		Name:                 types.StringValue("test-output"),
		Type:                 types.StringValue("elasticsearch"),
		Hosts:                mustListValue(t, hosts),
		CaSha256:             types.StringValue(caSha256),
		CaTrustedFingerprint: types.StringValue(caTrustedFingerprint),
		DefaultIntegrations:  types.BoolValue(isDefault),
		DefaultMonitoring:    types.BoolValue(isDefaultMonitoring),
		ConfigYaml:           types.StringValue(configYaml),
	}

	tests := []struct {
		name string
		run  func(t *testing.T) (outputItemModel, diag.Diagnostics)
	}{
		{
			name: "elasticsearch",
			run: func(t *testing.T) (outputItemModel, diag.Diagnostics) {
				model := outputItemModel{}
				diags := model.fromAPIElasticsearchModel(t.Context(), &kbapi.KibanaHTTPAPIsOutputResponseElasticsearch{
					Id: &id, Name: "test-output", Type: "elasticsearch",
					Hosts: hosts, CaSha256: &caSha256, CaTrustedFingerprint: &caTrustedFingerprint,
					IsDefault: &isDefault, IsDefaultMonitoring: &isDefaultMonitoring, ConfigYaml: &configYaml,
				})
				return model, diags
			},
		},
		{
			name: "kafka",
			run: func(t *testing.T) (outputItemModel, diag.Diagnostics) {
				model := outputItemModel{}
				diags := model.fromAPIKafkaModel(t.Context(), &kbapi.KibanaHTTPAPIsOutputResponseKafka{
					Id: &id, Name: "test-output", Type: "elasticsearch",
					Hosts: hosts, CaSha256: &caSha256, CaTrustedFingerprint: &caTrustedFingerprint,
					IsDefault: &isDefault, IsDefaultMonitoring: &isDefaultMonitoring, ConfigYaml: &configYaml,
				})
				return model, diags
			},
		},
		{
			name: "logstash",
			run: func(t *testing.T) (outputItemModel, diag.Diagnostics) {
				model := outputItemModel{}
				diags := model.fromAPILogstashModel(t.Context(), &kbapi.KibanaHTTPAPIsOutputResponseLogstash{
					Id: &id, Name: "test-output", Type: "elasticsearch",
					Hosts: hosts, CaSha256: &caSha256, CaTrustedFingerprint: &caTrustedFingerprint,
					IsDefault: &isDefault, IsDefaultMonitoring: &isDefaultMonitoring, ConfigYaml: &configYaml,
				})
				return model, diags
			},
		},
		{
			name: "remote_elasticsearch",
			run: func(t *testing.T) (outputItemModel, diag.Diagnostics) {
				model := outputItemModel{}
				diags := model.fromAPIRemoteElasticsearchModel(t.Context(), &kbapi.KibanaHTTPAPIsOutputResponseRemoteElasticsearch{
					Id: &id, Name: "test-output", Type: "elasticsearch",
					Hosts: hosts, CaSha256: &caSha256, CaTrustedFingerprint: &caTrustedFingerprint,
					IsDefault: &isDefault, IsDefaultMonitoring: &isDefaultMonitoring, ConfigYaml: &configYaml,
				})
				return model, diags
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, diags := tc.run(t)
			require.False(t, diags.HasError(), diags)
			assert.Equal(t, want, got)
		})
	}
}

func mustListValue(t *testing.T, values []string) types.List {
	t.Helper()
	elems := make([]types.String, len(values))
	for i, v := range values {
		elems[i] = types.StringValue(v)
	}
	list, diags := types.ListValueFrom(t.Context(), types.StringType, elems)
	require.False(t, diags.HasError(), diags)
	return list
}
