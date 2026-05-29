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

	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/writeonlyhash"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	fwtypes "github.com/hashicorp/terraform-plugin-framework/types"
)

var secretHasher = writeonlyhash.New("elasticsearch_connector")

func secretHashKey(mapKey string) string {
	return secretHasher.PrivateStateKey(`configuration_values["` + mapKey + `"].secret_value`)
}

func storeSecretHashes(
	ctx context.Context,
	private entitycore.PrivateStateStorage,
	configMap map[string]ConfigurationValueModel,
	diags *diag.Diagnostics,
) {
	if private == nil {
		return
	}
	for key, elem := range configMap {
		if !typeutils.IsKnown(elem.SecretValue) {
			continue
		}
		hash, err := secretHasher.Compute(elem.SecretValue.ValueString())
		if err != nil {
			diags.AddError("Failed to hash write-only attribute", err.Error())
			return
		}
		diags.Append(private.SetKey(ctx, secretHashKey(key), hash)...)
	}
}

func clearRemovedSecretHashes(
	ctx context.Context,
	private entitycore.PrivateStateStorage,
	priorMap, configMap map[string]ConfigurationValueModel,
	diags *diag.Diagnostics,
) {
	if private == nil || priorMap == nil {
		return
	}
	for key, priorElem := range priorMap {
		if !typeutils.IsKnown(priorElem.SecretValue) {
			continue
		}
		if _, stillPresent := configMap[key]; stillPresent {
			if typeutils.IsKnown(configMap[key].SecretValue) {
				continue
			}
		}
		diags.Append(private.SetKey(ctx, secretHashKey(key), nil)...)
	}
}

func clearAllSecretHashesFromPrior(
	ctx context.Context,
	private any,
	prior ContentConnectorData,
	diags *diag.Diagnostics,
) {
	ps, ok := private.(entitycore.PrivateStateStorage)
	if !ok || ps == nil || prior.ConfigurationValues.IsNull() || !typeutils.IsKnown(prior.ConfigurationValues) {
		return
	}
	priorMap := typeutils.MapTypeAs[ConfigurationValueModel](ctx, prior.ConfigurationValues, configurationValuesPath, diags)
	if diags.HasError() {
		return
	}
	for key := range priorMap {
		diags.Append(ps.SetKey(ctx, secretHashKey(key), nil)...)
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
