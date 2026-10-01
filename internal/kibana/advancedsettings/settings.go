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

package advancedsettings

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	kibanaoapi "github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// settingsFromModel converts the settings map attribute into Go values keyed
// by setting name. A null or unknown map yields a nil map.
func settingsFromModel(ctx context.Context, settings types.Map) (map[string]jsontypes.Normalized, diag.Diagnostics) {
	if settings.IsNull() || settings.IsUnknown() {
		return nil, nil
	}
	var result map[string]jsontypes.Normalized
	diags := settings.ElementsAs(ctx, &result, false)
	return result, diags
}

// settingsToModel converts settings keyed by name back into the settings map
// attribute.
func settingsToModel(ctx context.Context, settings map[string]jsontypes.Normalized) (types.Map, diag.Diagnostics) {
	return types.MapValueFrom(ctx, jsontypes.NormalizedType{}, settings)
}

// buildChanges returns the request body changes for the planned settings.
// Settings tracked in prior but absent from planned are sent as nil, which
// resets them to their Kibana defaults.
func buildChanges(planned, prior map[string]jsontypes.Normalized) (map[string]any, diag.Diagnostics) {
	var diags diag.Diagnostics
	changes := make(map[string]any, len(planned)+len(prior))
	for key := range prior {
		if _, ok := planned[key]; !ok {
			changes[key] = nil
		}
	}
	for key, value := range planned {
		var decoded any
		if err := json.Unmarshal([]byte(value.ValueString()), &decoded); err != nil {
			diags.AddError(
				"Invalid advanced setting value",
				fmt.Sprintf("The value of setting %q is not valid JSON: %s", key, err),
			)
			continue
		}
		changes[key] = decoded
	}
	return changes, diags
}

// resetChanges returns the request body changes that reset every tracked
// setting to its Kibana default.
func resetChanges(tracked map[string]jsontypes.Normalized) map[string]any {
	changes := make(map[string]any, len(tracked))
	for key := range tracked {
		changes[key] = nil
	}
	return changes
}

// trackedSettingsFromAPI returns the value Kibana reports for each tracked
// setting. Tracked settings that Kibana no longer reports (for example because
// they were reset to their default outside Terraform) are omitted so the next
// plan sets them again. When Kibana reports a value that is semantically equal
// to the tracked one, the tracked string is kept to avoid formatting-only
// differences.
func trackedSettingsFromAPI(
	ctx context.Context,
	tracked map[string]jsontypes.Normalized,
	actual map[string]kibanaoapi.AdvancedSetting,
) (map[string]jsontypes.Normalized, diag.Diagnostics) {
	var diags diag.Diagnostics
	result := make(map[string]jsontypes.Normalized, len(tracked))
	for key, trackedValue := range tracked {
		setting, ok := actual[key]
		if !ok {
			continue
		}

		encoded, err := encodeSettingValue(setting.Value)
		if err != nil {
			diags.AddError(
				"Failed to encode advanced setting value",
				fmt.Sprintf("Unable to encode the value of setting %q returned by Kibana: %s", key, err),
			)
			continue
		}
		value := jsontypes.NewNormalizedValue(encoded)

		if !trackedValue.IsNull() && !trackedValue.IsUnknown() {
			equal, eqDiags := trackedValue.StringSemanticEquals(ctx, value)
			diags.Append(eqDiags...)
			if equal {
				value = trackedValue
			}
		}
		result[key] = value
	}
	return result, diags
}

// encodeSettingValue JSON-encodes a setting value without HTML escaping, so
// values containing characters such as < or & stay readable in state.
func encodeSettingValue(value any) (string, error) {
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(value); err != nil {
		return "", err
	}
	return strings.TrimSuffix(buf.String(), "\n"), nil
}
