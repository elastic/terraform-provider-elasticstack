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
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

type secretChangeOutcome struct {
	NeedsUpdate bool
	Warnings    []string
	KeysToClear []string
}

func writeOnlySecretDriftWarning(key string) string {
	return fmt.Sprintf(
		`Detected a change to write-only attribute configuration_values["%s"].secret_value; the resource will be updated.`,
		key,
	)
}

func evaluateSecretPlanChanges(
	configMap, stateMap map[string]ConfigurationValueModel,
	getHash func(key string) ([]byte, diag.Diagnostics),
) (secretChangeOutcome, diag.Diagnostics) {
	var diags diag.Diagnostics
	var out secretChangeOutcome

	for key, elem := range configMap {
		if !typeutils.IsKnown(elem.SecretValue) {
			continue
		}
		value := elem.SecretValue.ValueString()
		storedHash, hashDiags := getHash(key)
		diags.Append(hashDiags...)
		if diags.HasError() {
			return out, diags
		}
		if len(storedHash) == 0 {
			continue
		}
		if !secretHasher.Matches(value, storedHash) {
			out.NeedsUpdate = true
			out.Warnings = append(out.Warnings, writeOnlySecretDriftWarning(key))
		}
	}

	for key, priorElem := range stateMap {
		if !typeutils.IsKnown(priorElem.SecretValue) {
			continue
		}
		if _, inConfig := configMap[key]; inConfig && typeutils.IsKnown(configMap[key].SecretValue) {
			continue
		}
		out.KeysToClear = append(out.KeysToClear, key)
	}

	return out, diags
}
