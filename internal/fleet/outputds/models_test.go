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
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

// TestOutputItemModel_populateFromAPI_CommonFields verifies that the shared
// output fields are mapped identically regardless of which concrete
// KibanaHTTPAPIsOutputResponse* type came out of the union, guarding the
// fromAPIOutputResponse generic collapse against regressions per-type.
func TestOutputItemModel_populateFromAPI_CommonFields(t *testing.T) {
	t.Parallel()

	id := "output-id"
	caSha256 := "ca-sha-256"
	caTrustedFingerprint := "ca-trusted-fingerprint"
	configYaml := "config: yaml"
	isDefault := true
	isDefaultMonitoring := false

	tests := []struct {
		name  string
		union func(t *testing.T) kbapi.OutputUnion
	}{
		{
			name: "elasticsearch",
			union: func(t *testing.T) kbapi.OutputUnion {
				var union kbapi.OutputUnion
				require.NoError(t, union.FromKibanaHTTPAPIsOutputResponseElasticsearch(kbapi.KibanaHTTPAPIsOutputResponseElasticsearch{
					Id: &id, Name: "test-output", Type: kbapi.KibanaHTTPAPIsOutputResponseElasticsearchTypeElasticsearch,
					Hosts: []string{"https://elasticsearch:9200"}, CaSha256: &caSha256,
					CaTrustedFingerprint: &caTrustedFingerprint,
					IsDefault:            &isDefault, IsDefaultMonitoring: &isDefaultMonitoring,
					ConfigYaml: &configYaml,
				}))
				return union
			},
		},
		{
			name: "kafka",
			union: func(t *testing.T) kbapi.OutputUnion {
				var union kbapi.OutputUnion
				require.NoError(t, union.FromKibanaHTTPAPIsOutputResponseKafka(kbapi.KibanaHTTPAPIsOutputResponseKafka{
					Id: &id, Name: "test-output", Type: kbapi.KibanaHTTPAPIsOutputResponseKafkaTypeKafka,
					AuthType: "user_pass",
					Hosts:    []string{"https://elasticsearch:9200"}, CaSha256: &caSha256,
					CaTrustedFingerprint: &caTrustedFingerprint,
					IsDefault:            &isDefault, IsDefaultMonitoring: &isDefaultMonitoring,
					ConfigYaml: &configYaml,
				}))
				return union
			},
		},
		{
			name: "logstash",
			union: func(t *testing.T) kbapi.OutputUnion {
				var union kbapi.OutputUnion
				require.NoError(t, union.FromKibanaHTTPAPIsOutputResponseLogstash(kbapi.KibanaHTTPAPIsOutputResponseLogstash{
					Id: &id, Name: "test-output", Type: kbapi.KibanaHTTPAPIsOutputResponseLogstashTypeLogstash,
					Hosts: []string{"https://elasticsearch:9200"}, CaSha256: &caSha256,
					CaTrustedFingerprint: &caTrustedFingerprint,
					IsDefault:            &isDefault, IsDefaultMonitoring: &isDefaultMonitoring,
					ConfigYaml: &configYaml,
				}))
				return union
			},
		},
		{
			name: "remote_elasticsearch",
			union: func(t *testing.T) kbapi.OutputUnion {
				var union kbapi.OutputUnion
				require.NoError(t, union.FromKibanaHTTPAPIsOutputResponseRemoteElasticsearch(kbapi.KibanaHTTPAPIsOutputResponseRemoteElasticsearch{
					Id: &id, Name: "test-output", Type: kbapi.KibanaHTTPAPIsOutputResponseRemoteElasticsearchTypeRemoteElasticsearch,
					Hosts: []string{"https://elasticsearch:9200"}, CaSha256: &caSha256,
					CaTrustedFingerprint: &caTrustedFingerprint,
					IsDefault:            &isDefault, IsDefaultMonitoring: &isDefaultMonitoring,
					ConfigYaml: &configYaml,
				}))
				return union
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			union := tc.union(t)
			model := outputItemModel{}
			diags := model.populateFromAPI(t.Context(), &union)

			require.False(t, diags.HasError(), diags)
			require.Equal(t, types.StringValue(id), model.ID)
			require.Equal(t, types.StringValue("test-output"), model.Name)
			require.Equal(t, types.StringValue(tc.name), model.Type)
			require.Equal(t, types.StringValue(caSha256), model.CaSha256)
			require.Equal(t, types.StringValue(caTrustedFingerprint), model.CaTrustedFingerprint)
			require.Equal(t, types.BoolValue(isDefault), model.DefaultIntegrations)
			require.Equal(t, types.BoolValue(isDefaultMonitoring), model.DefaultMonitoring)
			require.Equal(t, types.StringValue(configYaml), model.ConfigYaml)

			var hosts []string
			require.False(t, model.Hosts.ElementsAs(t.Context(), &hosts, false).HasError())
			require.Equal(t, []string{"https://elasticsearch:9200"}, hosts)
		})
	}
}
