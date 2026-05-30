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

package resource

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/connector"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	fwtypes "github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func TestEvaluateSecretPlanChanges_driftDetected(t *testing.T) {
	t.Parallel()

	hash, err := secretHasher.Compute("pw1")
	require.NoError(t, err)

	configMap := map[string]connector.ConfigurationValueModel{
		"password": {SecretValue: fwtypes.StringValue("pw2")},
	}
	outcome, diags := evaluateSecretPlanChanges(configMap, nil, func(_ string) ([]byte, diag.Diagnostics) {
		return hash, nil
	})
	require.False(t, diags.HasError())
	require.True(t, outcome.NeedsUpdate)
	require.Len(t, outcome.Warnings, 1)
	require.Equal(
		t,
		`Detected a change to write-only attribute configuration_values["password"].secret_value; the resource will be updated.`,
		outcome.Warnings[0],
	)
}

func TestEvaluateSecretPlanChanges_noHashBaseline(t *testing.T) {
	t.Parallel()

	configMap := map[string]connector.ConfigurationValueModel{
		"password": {SecretValue: fwtypes.StringValue("pw1")},
	}
	outcome, diags := evaluateSecretPlanChanges(configMap, nil, func(_ string) ([]byte, diag.Diagnostics) {
		return nil, nil
	})
	require.False(t, diags.HasError())
	require.False(t, outcome.NeedsUpdate)
	require.Empty(t, outcome.Warnings)
}

func TestEvaluateSecretPlanChanges_clearsRemovedSecret(t *testing.T) {
	t.Parallel()

	configMap := map[string]connector.ConfigurationValueModel{}
	stateMap := map[string]connector.ConfigurationValueModel{
		"password": {SecretValue: fwtypes.StringValue("placeholder")},
	}
	outcome, diags := evaluateSecretPlanChanges(configMap, stateMap, func(_ string) ([]byte, diag.Diagnostics) {
		return []byte("hash"), nil
	})
	require.False(t, diags.HasError())
	require.Equal(t, []string{"password"}, outcome.KeysToClear)
}
