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
	"fmt"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const externalIDStreamVarKey = "aws.credentials.external_id"

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

func managedIntegrationStreamVarString(prop *kbapi.KibanaHTTPAPIsManagedIntegration_Inputs_Streams_Vars_AdditionalProperties) (string, bool) {
	if prop == nil {
		return "", false
	}
	if s, err := prop.AsKibanaHTTPAPIsManagedIntegrationInputsStreamsVars0(); err == nil {
		return s, true
	}
	return managedIntegrationStreamVarStringWrapped(prop)
}

func managedIntegrationStreamVarStringWrapped(prop *kbapi.KibanaHTTPAPIsManagedIntegration_Inputs_Streams_Vars_AdditionalProperties) (string, bool) {
	raw, err := json.Marshal(prop)
	if err != nil {
		return "", false
	}
	var wrapper struct {
		Value json.RawMessage `json:"value"`
	}
	if err := json.Unmarshal(raw, &wrapper); err != nil {
		return "", false
	}
	var got string
	if err := json.Unmarshal(wrapper.Value, &got); err != nil {
		return "", false
	}
	return got, true
}

func managedIntegrationStreamVarSecretRefID(prop *kbapi.KibanaHTTPAPIsManagedIntegration_Inputs_Streams_Vars_AdditionalProperties) (string, bool) {
	if prop == nil {
		return "", false
	}
	if ref, err := prop.AsKibanaHTTPAPIsManagedIntegrationInputsStreamsVars5(); err == nil {
		if ref.IsSecretRef && ref.Id != "" {
			return ref.Id, true
		}
	}
	return managedIntegrationStreamVarSecretRefWrapped(prop)
}

func managedIntegrationStreamVarSecretRefWrapped(prop *kbapi.KibanaHTTPAPIsManagedIntegration_Inputs_Streams_Vars_AdditionalProperties) (string, bool) {
	raw, err := json.Marshal(prop)
	if err != nil {
		return "", false
	}
	var wrapper struct {
		Value json.RawMessage `json:"value"`
	}
	if err := json.Unmarshal(raw, &wrapper); err != nil {
		return "", false
	}
	var secretRef struct {
		ID          string `json:"id"`
		IsSecretRef bool   `json:"isSecretRef"`
	}
	if err := json.Unmarshal(wrapper.Value, &secretRef); err != nil {
		return "", false
	}
	if !secretRef.IsSecretRef || secretRef.ID == "" {
		return "", false
	}
	return secretRef.ID, true
}

func TestManagedIntegrationStreamVarStringDecode(t *testing.T) {
	t.Parallel()

	t.Run("bare string union arm", func(t *testing.T) {
		t.Parallel()
		item := mustManagedIntegrationProbeResponse(t, `{
			"aws.account_type": "organization-account"
		}`)
		v := streamVarFromProbe(t, item, "aws.account_type")
		got, ok := managedIntegrationStreamVarString(v)
		assert.True(t, ok)
		assert.Equal(t, "organization-account", got)
	})

	t.Run("wrapped value fallback", func(t *testing.T) {
		t.Parallel()
		item := mustManagedIntegrationProbeResponse(t, `{
			"aws.account_type": {"value": "single-account"}
		}`)
		v := streamVarFromProbe(t, item, "aws.account_type")
		got, ok := managedIntegrationStreamVarString(v)
		assert.True(t, ok)
		assert.Equal(t, "single-account", got)
	})
}

func TestManagedIntegrationStreamVarSecretRefDecode(t *testing.T) {
	t.Parallel()

	t.Run("bare secret ref union arm", func(t *testing.T) {
		t.Parallel()
		item := mustManagedIntegrationProbeResponse(t, `{
			"aws.credentials.external_id": {"id": "secret-bare", "isSecretRef": true}
		}`)
		v := streamVarFromProbe(t, item, externalIDStreamVarKey)
		id, ok := managedIntegrationStreamVarSecretRefID(v)
		assert.True(t, ok)
		assert.Equal(t, "secret-bare", id)
	})

	t.Run("wrapped secret ref fallback", func(t *testing.T) {
		t.Parallel()
		item := mustManagedIntegrationProbeResponse(t, `{
			"aws.credentials.external_id": {"value": {"id": "secret-wrapped", "isSecretRef": true}}
		}`)
		v := streamVarFromProbe(t, item, externalIDStreamVarKey)
		id, ok := managedIntegrationStreamVarSecretRefID(v)
		assert.True(t, ok)
		assert.Equal(t, "secret-wrapped", id)
	})
}

func streamVarFromProbe(t *testing.T, item *kbapi.KibanaHTTPAPIsManagedIntegration, key string) *kbapi.KibanaHTTPAPIsManagedIntegration_Inputs_Streams_Vars_AdditionalProperties {
	t.Helper()
	v, err := managedIntegrationCSPMFindingsStreamVar(item, key)
	require.NoError(t, err)
	return v
}

func managedIntegrationCSPMFindingsStreamVar(item *kbapi.KibanaHTTPAPIsManagedIntegration, key string) (*kbapi.KibanaHTTPAPIsManagedIntegration_Inputs_Streams_Vars_AdditionalProperties, error) {
	if item == nil {
		return nil, fmt.Errorf("managed integration item is nil")
	}
	in, ok := item.Inputs[cspmMappedInputKey]
	if !ok {
		return nil, fmt.Errorf("input %q missing from managed integration API response", cspmMappedInputKey)
	}
	if in.Streams == nil {
		return nil, fmt.Errorf("input %q has no streams in API response", cspmMappedInputKey)
	}
	stream, ok := (*in.Streams)[cspmFindingsStreamKey]
	if !ok {
		return nil, fmt.Errorf("stream %q missing from managed integration API response", cspmFindingsStreamKey)
	}
	if stream.Vars == nil {
		return nil, fmt.Errorf("stream %q has no vars in API response", cspmFindingsStreamKey)
	}
	v, ok := (*stream.Vars)[key]
	if !ok {
		return nil, fmt.Errorf("stream var %q missing from managed integration API response", key)
	}
	return v, nil
}
