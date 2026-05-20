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

package integrationpolicy

import (
	"encoding/json"
	"fmt"
	"maps"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type inputDefaultsModel struct {
	Vars    jsontypes.Normalized                `tfsdk:"vars"`
	Streams map[string]inputDefaultsStreamModel `tfsdk:"streams"`
}

type inputDefaultsStreamModel struct {
	Enabled types.Bool           `tfsdk:"enabled"`
	Vars    jsontypes.Normalized `tfsdk:"vars"`
}

type apiPolicyTemplates []apiPolicyTemplate
type apiPolicyTemplate struct {
	Name string `json:"name"`
	// Inputs is populated for integration-type packages where each policy template
	// declares one or more nested inputs with their own vars.
	Inputs []apiPolicyTemplateInput `json:"inputs"`
	// Input and Vars are populated for input-type packages where the policy template
	// declares a single input and vars are defined at the policy-template level.
	// Kibana copies these vars onto an implicit stream named "<package>.<template>".
	Input string  `json:"input"`
	Vars  apiVars `json:"vars"`
}

type apiPolicyTemplateInput struct {
	Type string  `json:"type"`
	Vars apiVars `json:"vars"`
}

func (policyTemplates apiPolicyTemplates) defaults() (map[string]jsontypes.Normalized, diag.Diagnostics) {
	defaults := map[string]jsontypes.Normalized{}

	if len(policyTemplates) == 0 {
		return defaults, nil
	}

	for _, policyTemplate := range policyTemplates {
		for _, inputTemplate := range policyTemplate.Inputs {
			name := fmt.Sprintf("%s-%s", policyTemplate.Name, inputTemplate.Type)
			varDefaults, diags := inputTemplate.Vars.defaults()
			if diags.HasError() {
				return nil, diags
			}

			defaults[name] = varDefaults
		}
	}

	return defaults, nil
}

type apiDatastreams []apiDatastream
type apiDatastream struct {
	Type    string                `json:"type"`
	Dataset string                `json:"dataset"`
	Streams []apiDatastreamStream `json:"streams"`
}

type apiDatastreamStream struct {
	Input   string  `json:"input"`
	Vars    apiVars `json:"vars"`
	Enabled bool    `json:"enabled"`
}

func (dataStreams apiDatastreams) defaults() (map[string]map[string]inputDefaultsStreamModel, diag.Diagnostics) {
	defaults := map[string]map[string]inputDefaultsStreamModel{}

	if dataStreams == nil {
		return defaults, nil
	}

	for _, dataStream := range dataStreams {
		for _, stream := range dataStream.Streams {
			varDefaults, diags := stream.Vars.defaults()
			if diags.HasError() {
				return nil, diags
			}

			d := defaults[stream.Input]
			if d == nil {
				d = map[string]inputDefaultsStreamModel{}
			}
			d[dataStream.Dataset] = inputDefaultsStreamModel{
				Enabled: types.BoolValue(stream.Enabled),
				Vars:    varDefaults,
			}
			defaults[stream.Input] = d
		}
	}

	return defaults, nil
}

type apiVars []apiVar
type apiVar struct {
	Name    string `json:"name"`
	Default any    `json:"default"`
	Multi   bool   `json:"multi"`
}

func (v apiVars) defaults() (jsontypes.Normalized, diag.Diagnostics) {
	varDefaults := map[string]any{}
	for _, inputVar := range v {
		if inputVar.Default != nil {
			varDefaults[inputVar.Name] = inputVar.Default
			continue
		}

		if inputVar.Multi {
			varDefaults[inputVar.Name] = []any{}
			continue
		}
	}

	varDefaultsBytes, err := json.Marshal(varDefaults)
	if err != nil {
		var diags diag.Diagnostics
		diags.AddError("Failed to marshal default vars for input", err.Error())
		return jsontypes.NewNormalizedNull(), diags
	}

	return jsontypes.NewNormalizedValue(string(varDefaultsBytes)), nil
}

func packageInfoToDefaults(pkg *kbapi.PackageInfo) (map[string]inputDefaultsModel, diag.Diagnostics) {
	policyTemplates, datastreams, diags := policyTemplateAndDataStreamsFromPackageInfo(pkg)
	if diags.HasError() {
		return nil, diags
	}

	defaultVarsByInput, inputVarsDiags := policyTemplates.defaults()
	diags.Append(inputVarsDiags...)

	defaultStreamsByInput, streamsDiags := datastreams.defaults()
	diags.Append(streamsDiags...)

	if diags.HasError() {
		return nil, diags
	}

	defaults := map[string]inputDefaultsModel{}
	for inputID, vars := range defaultVarsByInput {
		defaults[inputID] = inputDefaultsModel{
			Vars: vars,
		}
	}

	for inputIDSuffix, streams := range defaultStreamsByInput {
		for _, policyTemplate := range policyTemplates {
			inputID := fmt.Sprintf("%s-%s", policyTemplate.Name, inputIDSuffix)
			inputDefaults, ok := defaults[inputID]
			if !ok {
				inputDefaults.Vars = jsontypes.NewNormalizedNull()
			}

			inputDefaults.Streams = streams
			defaults[inputID] = inputDefaults
		}
	}

	// For input-type packages, vars are declared on the policy template itself
	// and Kibana exposes them on an implicit stream named "<package>.<template>"
	// attached to an input keyed "<template>-<input>".
	inputPackageDefaults, inputPkgDiags := inputPackagePolicyTemplatesToDefaults(pkg, policyTemplates)
	diags.Append(inputPkgDiags...)
	if diags.HasError() {
		return nil, diags
	}
	for inputID, inputDefaults := range inputPackageDefaults {
		existing, ok := defaults[inputID]
		if !ok {
			defaults[inputID] = inputDefaults
			continue
		}

		if !existing.Vars.IsNull() && !existing.Vars.IsUnknown() {
			inputDefaults.Vars = existing.Vars
		}
		if existing.Streams != nil {
			merged := make(map[string]inputDefaultsStreamModel, len(existing.Streams)+len(inputDefaults.Streams))
			maps.Copy(merged, existing.Streams)
			// Merge input-type stream defaults with existing datastream defaults.
			// Existing entries take precedence, but missing vars are filled in
			// from the policy-template level defaults so that semantic equality
			// can reconcile Kibana's applied state.
			for k, v := range inputDefaults.Streams {
				if existingStream, present := merged[k]; present {
					varsWithDefaults, streamVarDiags := applyDefaultsToVars(existingStream.Vars, v.Vars)
					diags.Append(streamVarDiags...)
					if diags.HasError() {
						return nil, diags
					}
					existingStream.Vars = varsWithDefaults
					merged[k] = existingStream
				} else {
					merged[k] = v
				}
			}
			inputDefaults.Streams = merged
		}
		defaults[inputID] = inputDefaults
	}

	return defaults, diags
}

// inputPackagePolicyTemplatesToDefaults derives default input/stream entries for
// input-type Fleet packages. Such packages declare a single `input` and a flat
// `vars` list on each policy template; Kibana materialises these as a stream
// named "<package>.<template>" under an input keyed "<template>-<input>".
func inputPackagePolicyTemplatesToDefaults(pkg *kbapi.PackageInfo, policyTemplates apiPolicyTemplates) (map[string]inputDefaultsModel, diag.Diagnostics) {
	if pkg == nil || pkg.Name == "" {
		return nil, nil
	}

	defaults := map[string]inputDefaultsModel{}
	for _, policyTemplate := range policyTemplates {
		if policyTemplate.Input == "" || len(policyTemplate.Vars) == 0 {
			continue
		}

		varDefaults, diags := policyTemplate.Vars.defaults()
		if diags.HasError() {
			return nil, diags
		}

		inputID := fmt.Sprintf("%s-%s", policyTemplate.Name, policyTemplate.Input)
		streamID := fmt.Sprintf("%s.%s", pkg.Name, policyTemplate.Name)

		existing, ok := defaults[inputID]
		if !ok {
			existing.Vars = jsontypes.NewNormalizedNull()
			existing.Streams = map[string]inputDefaultsStreamModel{}
		} else if existing.Streams == nil {
			existing.Streams = map[string]inputDefaultsStreamModel{}
		}
		existing.Streams[streamID] = inputDefaultsStreamModel{
			Enabled: types.BoolValue(true),
			Vars:    varDefaults,
		}
		defaults[inputID] = existing
	}

	return defaults, nil
}

func varsFromPackageInfo(pkg *kbapi.PackageInfo) (apiVars, diag.Diagnostics) {
	if pkg == nil {
		return nil, nil
	}

	var diags diag.Diagnostics
	var vars apiVars

	if pkg.Vars != nil && len(*pkg.Vars) > 0 {
		err := decodePackageInfoValue(pkg.Vars, &vars)
		if err != nil {
			diags.AddError("Failed to decode package vars", err.Error())
			return nil, diags
		}
	}

	return vars, nil
}

func policyTemplateAndDataStreamsFromPackageInfo(pkg *kbapi.PackageInfo) (apiPolicyTemplates, apiDatastreams, diag.Diagnostics) {
	if pkg == nil {
		return nil, nil, nil
	}

	var diags diag.Diagnostics

	var policyTemplates apiPolicyTemplates
	var dataStreams apiDatastreams

	if pkg.PolicyTemplates != nil {
		err := decodePackageInfoValue(pkg.PolicyTemplates, &policyTemplates)
		if err != nil {
			diags.AddError("Failed to decode package policy templates", err.Error())
			return nil, nil, diags
		}
	}

	if pkg.DataStreams != nil {
		err := decodePackageInfoValue(pkg.DataStreams, &dataStreams)
		if err != nil {
			diags.AddError("Failed to decode package data streams", err.Error())
			return nil, nil, diags
		}
	}

	return policyTemplates, dataStreams, nil
}

func decodePackageInfoValue(input any, target any) error {
	data, err := json.Marshal(input)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, target)
}
