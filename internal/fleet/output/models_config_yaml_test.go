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

package output

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestToAPICreate_OmitConfigYamlWhenNull asserts that when the user does not
// configure the optional `config_yaml` attribute, the serialized request body
// for every Fleet output type omits the field entirely rather than sending
// `"config_yaml": null`.
//
// Regression test for https://github.com/elastic/terraform-provider-elasticstack/issues/1067:
// Older Kibana spec generations emitted `ConfigYaml *string` without a
// `,omitempty` JSON tag (or used a nullable wrapper), so a nil pointer
// marshaled to explicit JSON null. The Fleet API rejects explicit null with
// HTTP 400 `expected value of type [string] but got [null]`. The current
// generated types carry `omitempty` for every output flavour; this test
// locks that in so a future `make generate` cannot silently regress it.
func TestToAPICreate_OmitConfigYamlWhenNull(t *testing.T) {
	t.Parallel()

	baseHosts := types.ListValueMust(types.StringType, []attr.Value{
		types.StringValue("https://127.0.0.1:443"),
	})

	mkModel := func(outputType string) outputModel {
		return outputModel{
			Name:       types.StringValue("example-" + outputType),
			Type:       types.StringValue(outputType),
			Hosts:      baseHosts,
			ConfigYaml: types.StringNull(), // the user-facing scenario from #1067
		}
	}

	tests := []struct {
		name string
		call func(t *testing.T) []byte
	}{
		{
			name: "elasticsearch create omits config_yaml",
			call: func(t *testing.T) []byte {
				m := mkModel("elasticsearch")
				union, diags := m.toAPICreateElasticsearchModel(context.Background())
				require.False(t, diags.HasError(), "diags: %v", diags)
				b, err := union.MarshalJSON()
				require.NoError(t, err)
				return b
			},
		},
		{
			name: "elasticsearch update omits config_yaml",
			call: func(t *testing.T) []byte {
				m := mkModel("elasticsearch")
				union, diags := m.toAPIUpdateElasticsearchModel(context.Background())
				require.False(t, diags.HasError(), "diags: %v", diags)
				b, err := union.MarshalJSON()
				require.NoError(t, err)
				return b
			},
		},
		{
			name: "logstash create omits config_yaml",
			call: func(t *testing.T) []byte {
				m := mkModel("logstash")
				union, diags := m.toAPICreateLogstashModel(context.Background())
				require.False(t, diags.HasError(), "diags: %v", diags)
				b, err := union.MarshalJSON()
				require.NoError(t, err)
				return b
			},
		},
		{
			name: "logstash update omits config_yaml",
			call: func(t *testing.T) []byte {
				m := mkModel("logstash")
				union, diags := m.toAPIUpdateLogstashModel(context.Background())
				require.False(t, diags.HasError(), "diags: %v", diags)
				b, err := union.MarshalJSON()
				require.NoError(t, err)
				return b
			},
		},
		{
			name: "remote_elasticsearch create omits config_yaml",
			call: func(t *testing.T) []byte {
				m := mkModel("remote_elasticsearch")
				m.ServiceToken = types.StringValue("token")
				union, diags := m.toAPICreateRemoteElasticsearchModel(context.Background())
				require.False(t, diags.HasError(), "diags: %v", diags)
				b, err := union.MarshalJSON()
				require.NoError(t, err)
				return b
			},
		},
		{
			name: "remote_elasticsearch update omits config_yaml",
			call: func(t *testing.T) []byte {
				m := mkModel("remote_elasticsearch")
				m.ServiceToken = types.StringValue("token")
				union, diags := m.toAPIUpdateRemoteElasticsearchModel(context.Background())
				require.False(t, diags.HasError(), "diags: %v", diags)
				b, err := union.MarshalJSON()
				require.NoError(t, err)
				return b
			},
		},
		{
			name: "kafka create omits config_yaml",
			call: func(t *testing.T) []byte {
				m := mkModel("kafka")
				union, diags := m.toAPICreateKafkaModel(context.Background())
				require.False(t, diags.HasError(), "diags: %v", diags)
				b, err := union.MarshalJSON()
				require.NoError(t, err)
				return b
			},
		},
		{
			name: "kafka update omits config_yaml",
			call: func(t *testing.T) []byte {
				m := mkModel("kafka")
				union, diags := m.toAPIUpdateKafkaModel(context.Background())
				require.False(t, diags.HasError(), "diags: %v", diags)
				b, err := union.MarshalJSON()
				require.NoError(t, err)
				return b
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			body := tc.call(t)
			// Parse the body so the assertion is keyed on the literal
			// "config_yaml" top-level attribute, not any substring (e.g.
			// `otel_exporter_config_yaml` would false-trigger a raw-string
			// Contains check).
			var parsed map[string]any
			require.NoError(t, json.Unmarshal(body, &parsed),
				"body should be valid JSON: %s", string(body))
			_, present := parsed["config_yaml"]
			assert.False(t, present,
				"config_yaml key should be absent when ConfigYaml is null, got body: %s", string(body))
		})
	}

	// Positive control: when ConfigYaml is configured, it MUST be present so
	// this test can't accidentally pass by masking a different bug.
	t.Run("elasticsearch create includes config_yaml when set", func(t *testing.T) {
		m := outputModel{
			Name:       types.StringValue("example"),
			Type:       types.StringValue("elasticsearch"),
			Hosts:      baseHosts,
			ConfigYaml: types.StringValue("bulk_max_size: 100\n"),
		}
		union, diags := m.toAPICreateElasticsearchModel(context.Background())
		require.False(t, diags.HasError())
		b, err := union.MarshalJSON()
		require.NoError(t, err)

		var parsed map[string]any
		require.NoError(t, json.Unmarshal(b, &parsed),
			"body should be valid JSON: %s", string(b))
		got, present := parsed["config_yaml"]
		require.True(t, present,
			"config_yaml key should be present when set, got: %s", string(b))
		assert.Equal(t, "bulk_max_size: 100\n", got)
	})
}
