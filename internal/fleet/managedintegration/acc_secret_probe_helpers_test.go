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

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
)

const externalIDStreamVarKey = "aws.credentials.external_id"

// externalIDSecretRefFromManagedIntegration extracts a Fleet secret reference ID
// from a managed integration GET response for the CSPM external_id stream var.
func externalIDSecretRefFromManagedIntegration(item *kbapi.KibanaHTTPAPIsManagedIntegration) (string, bool) {
	if item == nil {
		return "", false
	}
	in, ok := item.Inputs[cspmMappedInputKey]
	if !ok || in.Streams == nil {
		return "", false
	}
	stream, ok := (*in.Streams)[cspmFindingsStreamKey]
	if !ok || stream.Vars == nil {
		return "", false
	}
	v, ok := (*stream.Vars)[externalIDStreamVarKey]
	if !ok || v == nil {
		return "", false
	}
	raw, err := json.Marshal(v)
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
