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

package connector

import (
	"context"
	"encoding/json"

	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/writeonlyhash"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	fwtypes "github.com/hashicorp/terraform-plugin-framework/types"
)

var secretHasher = writeonlyhash.New("elasticsearch_connector")

const (
	privateStateUnavailableSummary = "Failed to persist write-only secret hash"
	privateStateUnavailableDetail  = "internal: private state writer unavailable; drift detection will not function. This is a bug."
)

func secretHashKey(mapKey string) string {
	return secretHasher.PrivateStateKey(`configuration_values["` + mapKey + `"].secret_value`)
}

// encodeSecretHashForPrivateState wraps a bcrypt hash as a JSON string. The
// Terraform Plugin Framework rejects private-state values that are not valid
// JSON, so raw bcrypt bytes (which contain '$' and other non-JSON characters)
// must be JSON-encoded before storage.
func encodeSecretHashForPrivateState(hash []byte) ([]byte, error) {
	return json.Marshal(string(hash))
}

func decodeSecretHashFromPrivateState(raw []byte) ([]byte, error) {
	if len(raw) == 0 {
		return nil, nil
	}
	var stored string
	if err := json.Unmarshal(raw, &stored); err != nil {
		return nil, err
	}
	return []byte(stored), nil
}

// secretHashKeysToClear reports the configuration keys whose private-state
// hash should be removed because the prior state held a secret_value branch
// that the new config no longer holds (either the key is gone or its branch
// is no longer secret_value).
func secretHashKeysToClear(priorMap, configMap map[string]ConfigurationValueModel) []string {
	var keys []string
	for key, priorElem := range priorMap {
		if !typeutils.IsKnown(priorElem.SecretValue) {
			continue
		}
		if configElem, inConfig := configMap[key]; inConfig && typeutils.IsKnown(configElem.SecretValue) {
			continue
		}
		keys = append(keys, key)
	}
	return keys
}

func storeSecretHashes(
	ctx context.Context,
	private entitycore.PrivateStateStorage,
	configMap map[string]ConfigurationValueModel,
	diags *diag.Diagnostics,
) {
	if private == nil {
		for _, elem := range configMap {
			if typeutils.IsKnown(elem.SecretValue) {
				diags.AddError(privateStateUnavailableSummary, privateStateUnavailableDetail)
				return
			}
		}
		return
	}
	for key, elem := range configMap {
		if !typeutils.IsKnown(elem.SecretValue) {
			continue
		}
		value := elem.SecretValue.ValueString()

		// bcrypt is intentionally slow (cost 10). Skip re-hashing when the
		// existing hash already verifies the value — common on no-op applies
		// where ModifyPlan triggered an update for non-secret reasons.
		existingRaw, getDiags := private.GetKey(ctx, secretHashKey(key))
		diags.Append(getDiags...)
		if diags.HasError() {
			return
		}
		if existing, err := decodeSecretHashFromPrivateState(existingRaw); err == nil && len(existing) > 0 && secretHasher.Matches(value, existing) {
			continue
		}

		hash, err := secretHasher.Compute(value)
		if err != nil {
			diags.AddError("Failed to hash write-only attribute", err.Error())
			return
		}
		encoded, err := encodeSecretHashForPrivateState(hash)
		if err != nil {
			diags.AddError("Failed to encode write-only secret hash", err.Error())
			return
		}
		diags.Append(private.SetKey(ctx, secretHashKey(key), encoded)...)
	}
}

func clearRemovedSecretHashes(
	ctx context.Context,
	private entitycore.PrivateStateStorage,
	priorMap, configMap map[string]ConfigurationValueModel,
	diags *diag.Diagnostics,
) {
	keys := secretHashKeysToClear(priorMap, configMap)
	if len(keys) == 0 {
		return
	}
	if private == nil {
		diags.AddError(privateStateUnavailableSummary, privateStateUnavailableDetail)
		return
	}
	for _, key := range keys {
		diags.Append(private.SetKey(ctx, secretHashKey(key), nil)...)
	}
}

func clearAllSecretHashesFromPrior(
	ctx context.Context,
	private entitycore.PrivateStateStorage,
	prior ContentConnectorData,
	diags *diag.Diagnostics,
) {
	if private == nil || prior.ConfigurationValues.IsNull() || !typeutils.IsKnown(prior.ConfigurationValues) {
		return
	}
	priorMap := typeutils.MapTypeAs[ConfigurationValueModel](ctx, prior.ConfigurationValues, configurationValuesPath, diags)
	if diags.HasError() {
		return
	}
	for key, elem := range priorMap {
		if activeConfigurationBranch(elem) != secretValueBranchAttrName {
			continue
		}
		diags.Append(private.SetKey(ctx, secretHashKey(key), nil)...)
	}
}

func configurationValuesFromModel(
	ctx context.Context,
	config fwtypes.Map,
	diags *diag.Diagnostics,
) map[string]ConfigurationValueModel {
	if config.IsNull() || !typeutils.IsKnown(config) {
		return nil
	}
	return typeutils.MapTypeAs[ConfigurationValueModel](ctx, config, configurationValuesPath, diags)
}
