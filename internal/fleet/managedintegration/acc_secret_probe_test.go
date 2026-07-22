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

package managedintegration_test

import (
	"encoding/json"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExternalIDSecretRefFromManagedIntegration(t *testing.T) {
	t.Parallel()

	t.Run("extracts secret ref from stream var", func(t *testing.T) {
		t.Parallel()
		item := mustManagedIntegrationProbeResponse(t, `{
			"aws.credentials.external_id": {"value": {"id": "secret-abc", "isSecretRef": true}}
		}`)
		id, ok := externalIDSecretRefFromManagedIntegration(item)
		assert.True(t, ok)
		assert.Equal(t, "secret-abc", id)
	})

	t.Run("rejects plain string value", func(t *testing.T) {
		t.Parallel()
		item := mustManagedIntegrationProbeResponse(t, `{
			"aws.credentials.external_id": {"value": "plaintext"}
		}`)
		_, ok := externalIDSecretRefFromManagedIntegration(item)
		assert.False(t, ok)
	})

	t.Run("missing input", func(t *testing.T) {
		t.Parallel()
		_, ok := externalIDSecretRefFromManagedIntegration(&kbapi.KibanaHTTPAPIsManagedIntegration{})
		assert.False(t, ok)
	})
}

func mustManagedIntegrationProbeResponse(t *testing.T, streamVarsJSON string) *kbapi.KibanaHTTPAPIsManagedIntegration {
	t.Helper()
	payload := `{
		"id": "probe-1",
		"name": "probe",
		"created_at": "2024-01-01T00:00:00.000Z",
		"created_by": "elastic",
		"updated_at": "2024-01-02T00:00:00.000Z",
		"updated_by": "elastic",
		"package": {"name": "cloud_security_posture", "version": "3.4.0", "title": "t"},
		"inputs": {
			"cspm-cloudbeat/cis_aws": {
				"enabled": true,
				"streams": {
					"cloud_security_posture.findings": {
						"enabled": true,
						"vars": ` + streamVarsJSON + `
					}
				}
			}
		}
	}`
	var item kbapi.KibanaHTTPAPIsManagedIntegration
	require.NoError(t, json.Unmarshal([]byte(payload), &item))
	return &item
}
