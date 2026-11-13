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
	"context"
	"encoding/json"
	"errors"
	"io"
	"strings"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	stringSettingsPath = path.Root("string_settings")
	jsonSettingsPath   = path.Root("json_settings")
)

func buildChanges(plan advancedSettingsModel, prior *advancedSettingsModel) (map[string]any, diag.Diagnostics) {
	var diags diag.Diagnostics
	changes := map[string]any{}

	if prior == nil {
		prior = &advancedSettingsModel{}
	}
	priorStrings, priorJSON := prior.StringSettings.Elements(), prior.JSONSettings.Elements()
	planStrings, planJSON := plan.StringSettings.Elements(), plan.JSONSettings.Elements()

	for key, element := range planStrings {
		if element.IsNull() {
			diags.Append(nullEntryDiagnostic(stringSettingsPath, key))
			continue
		}
		value, _ := element.(types.String)
		if previous, managed := priorStrings[key]; managed && previous.Equal(value) {
			continue
		}
		changes[key] = value.ValueString()
	}

	// Converting the elements would validate the JSON first, and that diagnostic does not name the key.
	for key, element := range planJSON {
		if element.IsNull() {
			diags.Append(nullEntryDiagnostic(jsonSettingsPath, key))
			continue
		}
		value, _ := element.(jsontypes.Normalized)
		parsed, jsonDiags := parseJSONSetting(key, value.ValueString())
		diags.Append(jsonDiags...)
		if jsonDiags.HasError() {
			continue
		}
		if previous, managed := priorJSON[key]; managed && sameJSONValue(previous, parsed) {
			continue
		}
		changes[key] = parsed
	}

	for _, managedBefore := range []map[string]attr.Value{priorStrings, priorJSON} {
		for key := range managedBefore {
			_, declaredAsString := planStrings[key]
			_, declaredAsJSON := planJSON[key]
			if !declaredAsString && !declaredAsJSON {
				changes[key] = nil
			}
		}
	}

	return changes, diags
}

func sameJSONValue(recorded attr.Value, parsed any) bool {
	text, _ := recorded.(jsontypes.Normalized)
	recordedParsed, err := parseJSONKeepingNumbers(text.ValueString())
	return err == nil && jsonValuesEqual(recordedParsed, parsed)
}

func parseJSONKeepingNumbers(text string) (any, error) {
	var parsed any
	decoder := json.NewDecoder(strings.NewReader(text))
	decoder.UseNumber()
	if err := decoder.Decode(&parsed); err != nil {
		return nil, err
	}
	if _, err := decoder.Token(); !errors.Is(err, io.EOF) {
		return nil, errors.New("unexpected content after the JSON value")
	}
	return parsed, nil
}

func refreshState(ctx context.Context, model advancedSettingsModel, userProvided map[string]json.RawMessage) (advancedSettingsModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	model.StringSettings = refreshSettings(ctx, model.StringSettings, userProvided, stringSettingsPath, refreshedStringValue, &diags)
	model.JSONSettings = refreshSettings(ctx, model.JSONSettings, userProvided, jsonSettingsPath, refreshedJSONValue, &diags)

	return model, diags
}

func refreshSettings[T attr.Value](
	ctx context.Context,
	current types.Map,
	userProvided map[string]json.RawMessage,
	attributePath path.Path,
	toElement func(recorded T, raw json.RawMessage) T,
	diags *diag.Diagnostics,
) types.Map {
	if current.IsNull() {
		return current
	}

	managed := typeutils.MapTypeAs[T](ctx, current, attributePath, diags)
	refreshed := make(map[string]T, len(managed))
	for key, recorded := range managed {
		if raw, found := userProvided[key]; found {
			refreshed[key] = toElement(recorded, raw)
		}
	}
	return typeutils.MapValueFrom(ctx, refreshed, current.ElementType(ctx), attributePath, diags)
}

// Terraform requires the state after apply to match the plan, and jsontypes.Normalized
// compares number text exactly, so a value Kibana holds as 1.5 must keep the text 1.50.
func refreshedJSONValue(recorded jsontypes.Normalized, raw json.RawMessage) jsontypes.Normalized {
	if kibanaValue, err := parseJSONKeepingNumbers(string(raw)); err == nil && sameJSONValue(recorded, kibanaValue) {
		return recorded
	}
	return jsontypes.NewNormalizedValue(string(raw))
}

func refreshedStringValue(_ types.String, raw json.RawMessage) types.String {
	var text string
	if err := json.Unmarshal(raw, &text); err != nil {
		return types.StringValue(string(raw))
	}
	return types.StringValue(text)
}
