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

package managedintegration

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const mappedPackagePolicyForBridgeJSON = `{
	"id": "policy-1",
	"name": "test-policy",
	"namespace": "default",
	"enabled": true,
	"created_at": "2024-01-01T00:00:00.000Z",
	"created_by": "elastic",
	"updated_at": "2024-01-02T00:00:00.000Z",
	"updated_by": "elastic",
	"revision": 1,
	"package": {"name": "cloud_security_posture", "version": "3.4.0", "title": "Security Posture Management"},
	"inputs": {
		"cspm-cloudbeat/cis_aws": {
			"enabled": true,
			"vars": {"cloud_formation_template": "x"}
		}
	}
}`

const typedArrayPackagePolicyForBridgeJSON = `{
	"id": "policy-1",
	"name": "test-policy",
	"namespace": "default",
	"enabled": true,
	"created_at": "2024-01-01T00:00:00.000Z",
	"created_by": "elastic",
	"updated_at": "2024-01-02T00:00:00.000Z",
	"updated_by": "elastic",
	"revision": 1,
	"package": {"name": "cloud_security_posture", "version": "3.4.0"},
	"inputs": [
		{
			"type": "cloudbeat/cis_aws",
			"policy_template": "cspm",
			"enabled": true
		}
	]
}`

func TestManagedIntegrationFromPackagePolicyReadResponse_mappedInputs(t *testing.T) {
	t.Parallel()

	data := mustPackagePolicyFromJSON(t, mappedPackagePolicyForBridgeJSON)
	item, diags := managedIntegrationFromPackagePolicyReadResponse(data)
	require.False(t, diags.HasError(), "%v", diags)
	assert.Equal(t, "policy-1", item.Id)
	require.Contains(t, item.Inputs, "cspm-cloudbeat/cis_aws")
	assert.NotNil(t, item.Inputs["cspm-cloudbeat/cis_aws"].Enabled)
	assert.True(t, *item.Inputs["cspm-cloudbeat/cis_aws"].Enabled)
}

func TestManagedIntegrationFromPackagePolicyReadResponse_typedInputsFail(t *testing.T) {
	t.Parallel()

	data := mustPackagePolicyFromJSON(t, typedArrayPackagePolicyForBridgeJSON)
	_, diags := managedIntegrationFromPackagePolicyReadResponse(data)
	require.True(t, diags.HasError(), "typed-array inputs must not be silently dropped")
	requireDiagnosticAtPath(t, diags, path.Root("inputs"), "Unexpected package policy inputs format")
}
