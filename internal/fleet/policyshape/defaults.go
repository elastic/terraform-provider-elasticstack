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

package policyshape

import (
	"encoding/json"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

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

// inputID builds the identifier Kibana uses for an input:
// "<policyTemplate.Name>-<inputType>".
func inputID(templateName, inputType string) string {
	return fmt.Sprintf("%s-%s", templateName, inputType)
}

// streamID builds the identifier Kibana uses for an input-type stream:
// "<pkgName>.<templateName>".
func streamID(pkgName, templateName string) string {
	return fmt.Sprintf("%s.%s", pkgName, templateName)
}

func (policyTemplates apiPolicyTemplates) defaults(pkgName string) (map[string]InputDefaultsModel, diag.Diagnostics) {
	defaults := map[string]InputDefaultsModel{}

	if len(policyTemplates) == 0 {
		return defaults, nil
	}

	for _, policyTemplate := range policyTemplates {
		// Integration-type packages: nested inputs with their own vars.
		for _, inputTemplate := range policyTemplate.Inputs {
			id := inputID(policyTemplate.Name, inputTemplate.Type)
			varDefaults, diags := inputTemplate.Vars.defaults()
			if diags.HasError() {
				return nil, diags
			}

			defaults[id] = InputDefaultsModel{
				Vars: varDefaults,
			}
		}

		// Input-type packages: single input and vars at the policy-template level.
		if pkgName != "" && policyTemplate.Input != "" && len(policyTemplate.Vars) > 0 {
			varDefaults, diags := policyTemplate.Vars.defaults()
			if diags.HasError() {
				return nil, diags
			}

			id := inputID(policyTemplate.Name, policyTemplate.Input)
			sid := streamID(pkgName, policyTemplate.Name)

			existing, ok := defaults[id]
			if !ok {
				existing.Vars = jsontypes.NewNormalizedNull()
				existing.Streams = map[string]InputDefaultsStreamModel{}
			} else if existing.Streams == nil {
				existing.Streams = map[string]InputDefaultsStreamModel{}
			}

			existing.Streams[sid] = InputDefaultsStreamModel{
				Enabled: types.BoolValue(true),
				Vars:    varDefaults,
			}
			defaults[id] = existing
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

func (dataStreams apiDatastreams) defaults() (map[string]map[string]InputDefaultsStreamModel, diag.Diagnostics) {
	defaults := map[string]map[string]InputDefaultsStreamModel{}

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
				d = map[string]InputDefaultsStreamModel{}
			}
			d[dataStream.Dataset] = InputDefaultsStreamModel{
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

// PackageInfoToDefaults derives the per-input default values (vars and
// per-stream vars/enabled state) declared by a Fleet package's metadata
// (policy templates and data streams), keyed by input ID. Callers merge the
// result into user-supplied input/stream config via applyDefaultsToInput (or
// equivalent) so that an unset var falls back to the package's declared
// default without ever writing that default into the request body.
func PackageInfoToDefaults(pkg *kbapi.KibanaHTTPAPIsGetPackageInfo) (map[string]InputDefaultsModel, diag.Diagnostics) {
	policyTemplates, datastreams, diags := policyTemplateAndDataStreamsFromPackageInfo(pkg)
	if diags.HasError() {
		return nil, diags
	}

	var pkgName string
	if pkg != nil {
		pkgName = pkg.Name
	}

	defaultInputsByID, inputVarsDiags := policyTemplates.defaults(pkgName)
	diags.Append(inputVarsDiags...)

	defaultStreamsByInput, streamsDiags := datastreams.defaults()
	diags.Append(streamsDiags...)

	if diags.HasError() {
		return nil, diags
	}

	// Merge datastream defaults into policy template defaults.
	// Datastream input suffixes (e.g. "kafka/metrics") are mapped to full
	// input IDs using each policy template's name.
	for inputIDSuffix, streams := range defaultStreamsByInput {
		for _, policyTemplate := range policyTemplates {
			id := inputID(policyTemplate.Name, inputIDSuffix)
			inputDefaults, ok := defaultInputsByID[id]
			if !ok {
				inputDefaults.Vars = jsontypes.NewNormalizedNull()
				inputDefaults.Streams = streams
				defaultInputsByID[id] = inputDefaults
				continue
			}

			if inputDefaults.Streams == nil {
				// Integration-type packages: datastreams define all streams.
				inputDefaults.Streams = streams
			} else {
				// Input-type packages: merge datastream streams with policy-template
				// stream defaults. Datastream vars take precedence, but missing keys
				// are filled from policy-template defaults.
				for sID, stream := range streams {
					if existingStream, present := inputDefaults.Streams[sID]; present {
						varsWithDefaults, streamVarDiags := applyDefaultsToVars(stream.Vars, existingStream.Vars)
						diags.Append(streamVarDiags...)
						if diags.HasError() {
							return nil, diags
						}
						existingStream.Vars = varsWithDefaults
						existingStream.Enabled = stream.Enabled
						inputDefaults.Streams[sID] = existingStream
					} else {
						inputDefaults.Streams[sID] = stream
					}
				}
			}

			defaultInputsByID[id] = inputDefaults
		}
	}

	return defaultInputsByID, diags
}

func varsFromPackageInfo(pkg *kbapi.KibanaHTTPAPIsGetPackageInfo) (apiVars, diag.Diagnostics) {
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

func policyTemplateAndDataStreamsFromPackageInfo(pkg *kbapi.KibanaHTTPAPIsGetPackageInfo) (apiPolicyTemplates, apiDatastreams, diag.Diagnostics) {
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
