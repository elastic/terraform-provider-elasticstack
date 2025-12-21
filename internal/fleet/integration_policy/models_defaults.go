package integration_policy

import (
	"encoding/json"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/mitchellh/mapstructure"
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
	Name   string                   `json:"name"`
	Inputs []apiPolicyTemplateInput `json:"inputs"`
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
	Name    string      `json:"name"`
	Default interface{} `json:"default"`
	Multi   bool        `json:"multi"`
}

func (v apiVars) defaults() (jsontypes.Normalized, diag.Diagnostics) {
	varDefaults := map[string]interface{}{}
	for _, inputVar := range v {
		if inputVar.Default != nil {
			varDefaults[inputVar.Name] = inputVar.Default
			continue
		}

		if inputVar.Multi {
			varDefaults[inputVar.Name] = []interface{}{}
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

	return defaults, diags
}

func varsFromPackageInfo(pkg *kbapi.PackageInfo) (apiVars, diag.Diagnostics) {
	if pkg == nil {
		return nil, nil
	}

	var diags diag.Diagnostics
	var vars apiVars

	if pkg.Vars != nil && len(*pkg.Vars) > 0 {
		err := mapstructure.Decode(pkg.Vars, &vars)
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
		err := mapstructure.Decode(pkg.PolicyTemplates, &policyTemplates)
		if err != nil {
			diags.AddError("Failed to decode package policy templates", err.Error())
			return nil, nil, diags
		}
	}

	if pkg.DataStreams != nil {
		err := mapstructure.Decode(pkg.DataStreams, &dataStreams)
		if err != nil {
			diags.AddError("Failed to decode package data streams", err.Error())
			return nil, nil, diags
		}
	}

	return policyTemplates, dataStreams, nil
}
