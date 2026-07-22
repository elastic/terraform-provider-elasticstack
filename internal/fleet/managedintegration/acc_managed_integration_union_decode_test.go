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
	in, ok := item.Inputs[cspmMappedInputKey]
	require.True(t, ok)
	require.NotNil(t, in.Streams)
	stream, ok := (*in.Streams)[cspmFindingsStreamKey]
	require.True(t, ok)
	require.NotNil(t, stream.Vars)
	v, ok := (*stream.Vars)[key]
	require.True(t, ok)
	return v
}
