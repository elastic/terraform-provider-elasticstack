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

// Package managedintegration: model ↔ Kibana managed_integrations API
// conversion — POST/PUT request bodies (toCreateBody) and state population
// from KibanaHTTPAPIsManagedIntegration (populateFromManagedIntegration).
//
// Several kbapi request/response fields are anonymous Go structs (oapi-codegen
// emits an unnamed struct type per inline schema property, so e.g.
// KibanaHTTPAPIsCreateManagedIntegrationRequest.Inputs have no nameable
// Go type). Rather than hand-spelling those anonymous types at every call
// site (fragile, and liable to drift silently out of sync on the next kbapi
// regeneration), this file builds plain map[string]any/[]any trees matching
// the wire shape and converts via a JSON marshal/unmarshal round trip into
// the destination field (e.g. `json.Unmarshal(b, &body.Inputs)`) -- the same
// pattern already used elsewhere in this repo for anonymous API fields (see
// internal/kibana/dashboard/panel/*/api_conv.go) and by
// policyshape.VarsMapToTypedMap.
package managedintegration

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/fleet/policyshape"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// packageModel and cloudConnectorModel are the Go representations of the
// `package` and `cloud_connector` nested attributes (see models.go's field-level
// doc comment). global_data_tags uses globalDataTagsItemModel in models.go.
type packageModel struct {
	Name    types.String `tfsdk:"name"`
	Version types.String `tfsdk:"version"`
	Title   types.String `tfsdk:"title"`
}

type cloudConnectorModel struct {
	Enabled          types.Bool   `tfsdk:"enabled"`
	CloudConnectorID types.String `tfsdk:"cloud_connector_id"`
	Name             types.String `tfsdk:"name"`
	TargetCSP        types.String `tfsdk:"target_csp"`
}

// managedIntegrationInputModel is the Go representation of a single `inputs` map
// element. It deliberately mirrors managedIntegrationInputAttributeTypes() in
// schema.go (enabled/condition/vars/streams) rather than reusing
// policyshape.InputModel directly: policyshape.InputModel also declares a
// `defaults` field that managedIntegrationInputAttributeTypes() does not surface (see
// schema.go's managedIntegrationInputType doc comment), and the Plugin Framework's
// object-to-struct decoding requires the struct's tfsdk fields to match the
// object's attribute set.
type managedIntegrationInputModel struct {
	Enabled   types.Bool           `tfsdk:"enabled"`
	Condition types.String         `tfsdk:"condition"`
	Vars      jsontypes.Normalized `tfsdk:"vars"`
	Streams   types.Map            `tfsdk:"streams"` // > policyshape.InputStreamModel
}

func (m managedIntegrationInputModel) GetStreams() types.Map { return m.Streams }

// Map/attribute key constants reused across this file's raw wire-format
// builders and attr-type helpers. Consolidated into named constants (rather
// than repeating the string literals) both for consistency and because
// schema.go already uses each literal at least once, so repeating them again
// here would trip golangci-lint's goconst check.
const (
	keyEnabled          = "enabled"
	keyCloudConnectorID = "cloud_connector_id"
)

// globalDataTagRaw is the wire shape of one global_data_tags entry in the
// create/update request body. It is a type alias (not a defined type) so it
// stays assignment-compatible with the anonymous struct kbapi emits for
// KibanaHTTPAPIsCreateManagedIntegrationRequest.GlobalDataTags's element type.
type globalDataTagRaw = struct {
	Name  string                                                                   `json:"name"`
	Value kbapi.KibanaHTTPAPIsCreateManagedIntegrationRequest_GlobalDataTags_Value `json:"value"`
}

func packageAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		attrName:  types.StringType,
		"version": types.StringType,
		"title":   types.StringType,
	}
}

func cloudConnectorAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		keyEnabled:          types.BoolType,
		keyCloudConnectorID: types.StringType,
		attrName:            types.StringType,
		"target_csp":        types.StringType,
	}
}

// globalDataTagsToModel converts managed_integrations global_data_tags into
// the Terraform map attribute, or a null map when there are none.
func globalDataTagsToModel(ctx context.Context, item *kbapi.KibanaHTTPAPIsManagedIntegration, diags *diag.Diagnostics) types.Map {
	elemType := globalDataTagsElementType()
	if item == nil || item.GlobalDataTags == nil || len(*item.GlobalDataTags) == 0 {
		return types.MapNull(elemType)
	}

	map0 := make(map[string]globalDataTagsItemModel, len(*item.GlobalDataTags))
	seenNames := make(map[string]struct{}, len(*item.GlobalDataTags))
	for _, tag := range *item.GlobalDataTags {
		tagPath := path.Root(attrGlobalDataTags).AtMapKey(tag.Name)
		if _, dup := seenNames[tag.Name]; dup {
			diags.AddAttributeError(
				tagPath,
				"Duplicate global_data_tags name",
				fmt.Sprintf("API returned global_data_tags name %q more than once.", tag.Name),
			)
			continue
		}
		seenNames[tag.Name] = struct{}{}
		tagItem := globalDataTagsItemModel{}
		if num, err := tag.Value.AsKibanaHTTPAPIsManagedIntegrationGlobalDataTagsValue1(); err == nil {
			tagItem.NumberValue = types.Float32Value(num)
		} else if str, err := tag.Value.AsKibanaHTTPAPIsManagedIntegrationGlobalDataTagsValue0(); err == nil {
			tagItem.StringValue = types.StringValue(str)
		} else {
			diags.AddAttributeError(
				tagPath,
				"Unsupported global_data_tags value type",
				fmt.Sprintf("API returned an unsupported value for tag %q; expected string or number.", tag.Name),
			)
			continue
		}
		map0[tag.Name] = tagItem
	}

	if diags.HasError() {
		return types.MapNull(elemType)
	}

	return typeutils.MapValueFrom(ctx, map0, elemType, path.Root(attrGlobalDataTags), diags)
}

// globalDataTagsRawFromModel converts the `global_data_tags` map attribute
// into request-body global_data_tags using typed union values.
func globalDataTagsRawFromModel(ctx context.Context, tags types.Map, diags *diag.Diagnostics) *[]globalDataTagRaw {
	if !typeutils.IsKnown(tags) {
		return nil
	}
	items := typeutils.MapTypeAs[globalDataTagsItemModel](ctx, tags, path.Root(attrGlobalDataTags), diags)
	if diags.HasError() {
		return nil
	}

	raw := make([]globalDataTagRaw, 0, len(items))
	for key, item := range items {
		tagPath := path.Root(attrGlobalDataTags).AtMapKey(key)
		var value kbapi.KibanaHTTPAPIsCreateManagedIntegrationRequest_GlobalDataTags_Value
		var err error
		switch {
		case typeutils.IsKnown(item.StringValue):
			err = value.FromKibanaHTTPAPIsCreateManagedIntegrationRequestGlobalDataTagsValue0(item.StringValue.ValueString())
		case typeutils.IsKnown(item.NumberValue):
			err = value.FromKibanaHTTPAPIsCreateManagedIntegrationRequestGlobalDataTagsValue1(item.NumberValue.ValueFloat32())
		default:
			diags.AddAttributeError(
				tagPath,
				"Invalid global_data_tags entry",
				"Each entry in global_data_tags must have exactly one of string_value or number_value set.",
			)
			continue
		}
		if err != nil {
			diags.AddAttributeError(tagPath, "Failed to encode global_data_tags", err.Error())
			continue
		}
		raw = append(raw, globalDataTagRaw{Name: key, Value: value})
	}
	if diags.HasError() {
		return nil
	}
	return &raw
}

// varsJSONFromAny builds a policyshape.VarsJSONValue from any mapped-format
// vars value (a bare-value union map from either the create response or the
// Format=Simplified package-policy response), integration-scoped via
// lookupCachedPackageInfo so unset package-declared defaults are filled in on
// read, matching internal/fleet/integration_policy/models.go's populateFromAPI.
func varsJSONFromAny(raw any, packageName, packageVersion string, diags *diag.Diagnostics) policyshape.VarsJSONValue {
	varsMap := policyshape.VarsAnyToMap(raw)
	if len(varsMap) == 0 {
		return policyshape.NewVarsJSONNull()
	}

	b, err := json.Marshal(varsMap)
	if err != nil {
		diags.AddAttributeError(path.Root(attrVarsJSON), "Failed to marshal vars_json from API response", err.Error())
		return policyshape.NewVarsJSONNull()
	}

	v, d := policyshape.NewVarsJSONWithIntegration(string(b), packageName, packageVersion, lookupCachedPackageInfo)
	diags.Append(d...)
	return v
}

// inputsKnownKeySet captures the set of keys of inputs's map value, or nil if
// inputs is not Known (null or unknown) -- see populateInputsFromManagedIntegration.
func inputsKnownKeySet(inputs policyshape.InputsValue) map[string]struct{} {
	if !typeutils.IsKnown(inputs.MapValue) {
		return nil
	}
	keys := make(map[string]struct{}, len(inputs.Elements()))
	for k := range inputs.Elements() {
		keys[k] = struct{}{}
	}
	return keys
}

// varsUnionToMap decodes a managed-integration vars map (typed union values,
// keyed by var name) into a plain map for Normalized JSON encoding. It is
// generic over the anonymous per-property union type kbapi emits, so the same
// logic serves both input-level and stream-level vars.
//
// Malformed union payloads are rejected when kbapi unmarshals the HTTP
// response; json.Marshal here only fails on unsupported Go types, which the
// generated client does not surface — so failure paths are covered by
// attribute-path wiring tests, not by constructing invalid unions in unit tests.
func varsUnionToMap[T any](vars *map[string]*T, attrPath path.Path, diags *diag.Diagnostics) map[string]any {
	if vars == nil || len(*vars) == 0 {
		return nil
	}
	b, err := json.Marshal(vars)
	if err != nil {
		diags.AddAttributeError(attrPath, "Failed to decode vars from API response", err.Error())
		return nil
	}
	if len(b) == 0 || string(b) == "null" {
		return nil
	}
	var out map[string]any
	if err := json.Unmarshal(b, &out); err != nil {
		diags.AddAttributeError(attrPath, "Failed to decode vars from API response", err.Error())
		return nil
	}
	return out
}

// populateInputsFromManagedIntegration builds the `inputs` map attribute from
// KibanaHTTPAPIsManagedIntegration.Inputs.
func populateInputsFromManagedIntegration(ctx context.Context, item *kbapi.KibanaHTTPAPIsManagedIntegration, knownKeys map[string]struct{}, diags *diag.Diagnostics) policyshape.InputsValue {
	if item == nil {
		return policyshape.NewInputsNull(managedIntegrationInputType())
	}
	inputs := maps.Clone(item.Inputs)
	if knownKeys != nil {
		for key := range inputs {
			if _, ok := knownKeys[key]; !ok {
				delete(inputs, key)
			}
		}
	}

	if len(inputs) == 0 {
		return policyshape.NewInputsNull(managedIntegrationInputType())
	}

	models := make(map[string]managedIntegrationInputModel, len(inputs))
	for inputID, wire := range inputs {
		inputPath := path.Root(attrInputs).AtMapKey(inputID)

		m := managedIntegrationInputModel{
			Enabled:   types.BoolPointerValue(wire.Enabled),
			Condition: types.StringPointerValue(wire.Condition),
			Vars:      typeutils.MarshalToNormalized(varsUnionToMap(wire.Vars, inputPath.AtName("vars"), diags), inputPath.AtName("vars"), diags),
		}

		if wire.Streams != nil && len(*wire.Streams) > 0 {
			streams := make(map[string]policyshape.InputStreamModel, len(*wire.Streams))
			for streamID, sw := range *wire.Streams {
				streamPath := inputPath.AtName("streams").AtMapKey(streamID)
				streams[streamID] = policyshape.InputStreamModel{
					Enabled:   types.BoolPointerValue(sw.Enabled),
					Condition: types.StringPointerValue(sw.Condition),
					Vars:      typeutils.MarshalToNormalized(varsUnionToMap(sw.Vars, streamPath.AtName("vars"), diags), streamPath.AtName("vars"), diags),
				}
			}
			streamsMap, d := types.MapValueFrom(ctx, policyshape.StreamType(), streams)
			diags.Append(d...)
			m.Streams = streamsMap
		} else {
			m.Streams = types.MapNull(policyshape.StreamType())
		}

		models[inputID] = m
	}

	inputsValue, d := policyshape.NewInputsValueFrom(ctx, managedIntegrationInputType(), models)
	diags.Append(d...)
	return inputsValue
}

// decodedManagedIntegrationInput is the once-decoded form of a single `inputs` map
// element, with its nested `streams` map (if any) already decoded too.
// decodeInputs produces this so that the request-body builders
// (applyCreateInputs, buildUpdateBody) don't each independently re-run the
// reflection-based typeutils.MapTypeAs decode over the same inputs+streams
// structure.
type decodedManagedIntegrationInput = policyshape.DecodedInput[managedIntegrationInputModel]

// decodeInputs decodes the `inputs` attribute, and each input's nested
// `streams` map, exactly once. Returns nil if `inputs` itself is null/unknown
// or fails to decode.
func (m managedIntegrationModel) decodeInputs(ctx context.Context, diags *diag.Diagnostics) map[string]decodedManagedIntegrationInput {
	return policyshape.DecodeInputs[managedIntegrationInputModel](ctx, m.Inputs, path.Root(attrInputs), diags)
}

// managedIntegrationRequestOptions controls how a plan model is compiled into
// KibanaHTTPAPIsCreateManagedIntegrationRequest (shared by Create and Update).
//
// Full-replace Update semantics (sendExplicitEmptyScalars): optional API fields
// use generated `omitempty` JSON tags — omitting a key clears that field on the
// server. Known-null optional attributes (description, namespace, inputs, …)
// are therefore omitted rather than sent as empty strings. Known-empty
// collections (for example vars_json = jsonencode({})) are marshaled
// explicitly to clear. Unknown inputs on Update produce an attribute error
// instead of omitting inputs and risking destructive API behavior.
type managedIntegrationRequestOptions struct {
	// omitCreateOnlyFields drops force, create_dataset_templates, and id from
	// the body (Update PUT must not re-send create-only knobs).
	omitCreateOnlyFields bool
	// priorForCloudConnector, when non-nil, supplies cloud_connector
	// {enabled, cloud_connector_id} for Update (never name/target_csp).
	// When nil, cloud_connector is taken from the plan model (Create).
	priorForCloudConnector *managedIntegrationModel
	// sendExplicitEmptyScalars encodes known-but-empty optional fields so a
	// full-replace PUT actively clears values removed from config.
	sendExplicitEmptyScalars bool
}

// toCreateBody compiles the config/plan model into
// PostFleetManagedIntegrationsJSONRequestBody.
func (m managedIntegrationModel) toCreateBody(ctx context.Context) (kbapi.PostFleetManagedIntegrationsJSONRequestBody, diag.Diagnostics) {
	return m.toManagedIntegrationRequestBody(ctx, managedIntegrationRequestOptions{})
}

func (m managedIntegrationModel) toManagedIntegrationRequestBody(ctx context.Context, opts managedIntegrationRequestOptions) (kbapi.KibanaHTTPAPIsCreateManagedIntegrationRequest, diag.Diagnostics) {
	var diags diag.Diagnostics

	m.diagnoseUnknownUpdatePlanFields(&diags, opts)
	if diags.HasError() {
		return kbapi.KibanaHTTPAPIsCreateManagedIntegrationRequest{}, diags
	}

	decodedInputs := m.decodeInputs(ctx, &diags)

	var pkg packageModel
	diags.Append(m.Package.As(ctx, &pkg, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return kbapi.KibanaHTTPAPIsCreateManagedIntegrationRequest{}, diags
	}

	body := kbapi.KibanaHTTPAPIsCreateManagedIntegrationRequest{
		Name: m.Name.ValueString(),
		Package: kbapi.KibanaHTTPAPIsPackagePolicyPackage{
			Name:    pkg.Name.ValueString(),
			Version: pkg.Version.ValueString(),
		},
	}

	if typeutils.IsKnown(pkg.Title) {
		body.Package.Title = pkg.Title.ValueStringPointer()
	}
	if !opts.omitCreateOnlyFields && typeutils.IsKnown(m.PolicyID) {
		body.Id = m.PolicyID.ValueStringPointer()
	}
	if typeutils.IsKnown(m.Description) {
		body.Description = m.Description.ValueStringPointer()
	}
	if typeutils.IsKnown(m.Namespace) {
		body.Namespace = m.Namespace.ValueStringPointer()
	}
	if typeutils.IsKnown(m.PolicyTemplate) {
		body.PolicyTemplate = m.PolicyTemplate.ValueStringPointer()
	}
	if !opts.omitCreateOnlyFields {
		if typeutils.IsKnown(m.Force) {
			body.Force = m.Force.ValueBoolPointer()
		}
		if typeutils.IsKnown(m.CreateDatasetTemplates) {
			body.CreateDatasetTemplates = m.CreateDatasetTemplates.ValueBoolPointer()
		}
	}

	if !m.VarsJSON.IsUnknown() {
		if m.VarsJSON.IsNull() {
			if opts.sendExplicitEmptyScalars {
				empty := map[string]any{}
				if b, err := json.Marshal(empty); err != nil {
					diags.AddAttributeError(path.Root(attrVarsJSON), "Failed to encode vars_json", err.Error())
				} else if err := json.Unmarshal(b, &body.Vars); err != nil {
					diags.AddAttributeError(path.Root(attrVarsJSON), "Failed to encode vars_json for the request", err.Error())
				}
			}
		} else {
			sanitized, sd := m.VarsJSON.SanitizedValue()
			diags.Append(sd...)
			if !sd.HasError() {
				varsMap := typeutils.NormalizedTypeToMap[any](jsontypes.NewNormalizedValue(sanitized), path.Root(attrVarsJSON), &diags)
				if len(varsMap) > 0 || opts.sendExplicitEmptyScalars {
					if b, err := json.Marshal(varsMap); err != nil {
						diags.AddAttributeError(path.Root(attrVarsJSON), "Failed to encode vars_json", err.Error())
					} else if err := json.Unmarshal(b, &body.Vars); err != nil {
						diags.AddAttributeError(path.Root(attrVarsJSON), "Failed to encode vars_json for the request", err.Error())
					}
				}
			}
		}
	}

	if typeutils.IsKnown(m.VarGroupSelections) {
		vgs := map[string]string{}
		diags.Append(m.VarGroupSelections.ElementsAs(ctx, &vgs, false)...)
		if len(vgs) > 0 || opts.sendExplicitEmptyScalars {
			body.VarGroupSelections = &vgs
		}
	}

	if !m.AdditionalDatastreamsPermissions.IsUnknown() {
		if !m.AdditionalDatastreamsPermissions.IsNull() {
			var perms []string
			diags.Append(m.AdditionalDatastreamsPermissions.ElementsAs(ctx, &perms, false)...)
			if perms == nil && opts.sendExplicitEmptyScalars {
				perms = []string{}
			}
			body.AdditionalDatastreamsPermissions = &perms
		}
	}

	if !m.GlobalDataTags.IsUnknown() {
		if !m.GlobalDataTags.IsNull() {
			tagsRaw := globalDataTagsRawFromModel(ctx, m.GlobalDataTags, &diags)
			if tagsRaw == nil && opts.sendExplicitEmptyScalars {
				empty := make([]globalDataTagRaw, 0)
				tagsRaw = &empty
			}
			body.GlobalDataTags = tagsRaw
		}
	}

	ccSource := m
	if opts.priorForCloudConnector != nil {
		ccSource = *opts.priorForCloudConnector
	}
	if typeutils.IsKnown(ccSource.CloudConnector) {
		var cc cloudConnectorModel
		diags.Append(ccSource.CloudConnector.As(ctx, &cc, basetypes.ObjectAsOptions{})...)

		raw := map[string]any{}
		if typeutils.IsKnown(cc.Enabled) {
			raw[keyEnabled] = cc.Enabled.ValueBool()
		}
		if typeutils.IsKnown(cc.CloudConnectorID) {
			raw[keyCloudConnectorID] = cc.CloudConnectorID.ValueString()
		}
		// Update must never send write-only name/target_csp; Create may.
		if opts.priorForCloudConnector == nil {
			if typeutils.IsKnown(cc.Name) {
				raw[attrName] = cc.Name.ValueString()
			}
			if typeutils.IsKnown(cc.TargetCSP) {
				raw["target_csp"] = cc.TargetCSP.ValueString()
			}
		}

		if len(raw) > 0 {
			if b, err := json.Marshal(raw); err != nil {
				diags.AddAttributeError(path.Root("cloud_connector"), "Failed to encode cloud_connector", err.Error())
			} else if err := json.Unmarshal(b, &body.CloudConnector); err != nil {
				diags.AddAttributeError(path.Root("cloud_connector"), "Failed to encode cloud_connector for the request", err.Error())
			}
		}
	}

	inputOpts := applyCreateInputsOptions{
		explicitEmptyInputs: opts.sendExplicitEmptyScalars &&
			typeutils.IsKnown(m.Inputs) &&
			typeutils.IsKnown(m.Inputs.MapValue),
		includeEmptyVars: opts.sendExplicitEmptyScalars,
	}
	applyCreateInputs(&body, decodedInputs, &diags, inputOpts)

	return body, diags
}

func (m managedIntegrationModel) diagnoseUnknownUpdatePlanFields(diags *diag.Diagnostics, opts managedIntegrationRequestOptions) {
	if !opts.sendExplicitEmptyScalars {
		return
	}
	const summary = "Cannot build managed integration update request"
	for _, attr := range []struct {
		key     string
		unknown bool
	}{
		{attrDescription, m.Description.IsUnknown()},
		{attrNamespace, m.Namespace.IsUnknown()},
		{attrPolicyTemplate, m.PolicyTemplate.IsUnknown()},
		{attrVarsJSON, m.VarsJSON.IsUnknown()},
		{attrVarGroupSelections, m.VarGroupSelections.IsUnknown()},
		{attrAdditionalDatastreamsPermissions, m.AdditionalDatastreamsPermissions.IsUnknown()},
		{attrGlobalDataTags, m.GlobalDataTags.IsUnknown()},
		{attrPackage, m.Package.IsUnknown()},
		{attrInputs, m.Inputs.IsUnknown()},
	} {
		if attr.unknown {
			diags.AddAttributeError(
				path.Root(attr.key),
				summary,
				fmt.Sprintf("The %s attribute is unknown; cannot build a full-replace update body without risking unintended clears on the API.", attr.key),
			)
		}
	}
}

type applyCreateInputsOptions struct {
	explicitEmptyInputs bool
	includeEmptyVars    bool
}

// applyCreateInputs converts the already-decoded `inputs` map (see
// decodeInputs) into the create body's Inputs field (a
// map[string]struct{...} of anonymous Go type -- see this file's header
// comment) via a JSON marshal/unmarshal round trip.
func applyCreateInputs(body *kbapi.PostFleetManagedIntegrationsJSONRequestBody, decoded map[string]decodedManagedIntegrationInput, diags *diag.Diagnostics, opts applyCreateInputsOptions) {
	if len(decoded) == 0 {
		if !opts.explicitEmptyInputs {
			return
		}
		raw := map[string]any{}
		b, err := json.Marshal(raw)
		if err != nil {
			diags.AddAttributeError(path.Root(attrInputs), "Failed to encode inputs", err.Error())
			return
		}
		if err := json.Unmarshal(b, &body.Inputs); err != nil {
			diags.AddAttributeError(path.Root(attrInputs), "Failed to encode inputs for the request", err.Error())
		}
		return
	}

	raw := map[string]any{}
	for inputID, di := range decoded {
		in := di.Model
		inputPath := path.Root(attrInputs).AtMapKey(inputID)
		entry := map[string]any{}

		if typeutils.IsKnown(in.Enabled) {
			entry[keyEnabled] = in.Enabled.ValueBool()
		}
		if typeutils.IsKnown(in.Condition) {
			entry["condition"] = in.Condition.ValueString()
		}
		varsMap := typeutils.NormalizedTypeToMap[any](in.Vars, inputPath.AtName("vars"), diags)
		if len(varsMap) > 0 || opts.includeEmptyVars {
			if varsMap == nil {
				varsMap = map[string]any{}
			}
			entry["vars"] = varsMap
		}

		if len(di.Streams) > 0 || opts.includeEmptyVars {
			streamsRaw := map[string]any{}
			for streamID, s := range di.Streams {
				streamPath := inputPath.AtName("streams").AtMapKey(streamID)
				streamEntry := map[string]any{}
				if typeutils.IsKnown(s.Enabled) {
					streamEntry[keyEnabled] = s.Enabled.ValueBool()
				}
				if typeutils.IsKnown(s.Condition) {
					streamEntry["condition"] = s.Condition.ValueString()
				}
				sv := typeutils.NormalizedTypeToMap[any](s.Vars, streamPath.AtName("vars"), diags)
				if len(sv) > 0 || opts.includeEmptyVars {
					if sv == nil {
						sv = map[string]any{}
					}
					streamEntry["vars"] = sv
				}
				streamsRaw[streamID] = streamEntry
			}
			entry["streams"] = streamsRaw
		}

		raw[inputID] = entry
	}

	b, err := json.Marshal(raw)
	if err != nil {
		diags.AddAttributeError(path.Root(attrInputs), "Failed to encode inputs", err.Error())
		return
	}
	if err := json.Unmarshal(b, &body.Inputs); err != nil {
		diags.AddAttributeError(path.Root(attrInputs), "Failed to encode inputs for the request", err.Error())
	}
}

// populateFromManagedIntegration updates Terraform state from a
// KibanaHTTPAPIsManagedIntegration response. Create-only attributes
// (force, force_delete, create_dataset_templates, skip_topology_check,
// policy_template) and cloud_connector write-only fields are preserved from
// the incoming model when set; cloud_connector.enabled and
// cloud_connector_id are merged from the API response (see
// applyCloudConnectorFromAPI). Call reconcileManagedIntegrationSecretsFromPrior
// after populate on read paths when prior state/config is available.
//
// spaceIDs is optional metadata when the caller has space membership outside
// the read response; when nil, space_ids defaults from spaceID when unset on
// the model.
func (m *managedIntegrationModel) populateFromManagedIntegration(ctx context.Context, spaceID string, item *kbapi.KibanaHTTPAPIsManagedIntegration, spaceIDs *[]string) diag.Diagnostics {
	var diags diag.Diagnostics
	if item == nil {
		return diags
	}

	m.ID = types.StringValue((&clients.CompositeID{ClusterID: spaceID, ResourceID: item.Id}).String())
	m.PolicyID = types.StringValue(item.Id)

	if spaceIDs != nil && len(*spaceIDs) > 0 {
		ids, d := types.SetValueFrom(ctx, types.StringType, *spaceIDs)
		diags.Append(d...)
		m.SpaceIDs = ids
	} else if !typeutils.IsKnown(m.SpaceIDs) {
		ids, d := types.SetValueFrom(ctx, types.StringType, []string{spaceID})
		diags.Append(d...)
		m.SpaceIDs = ids
	}

	m.Name = types.StringValue(item.Name)
	m.Description = typeutils.NonEmptyStringOrNull(item.Description)
	m.Namespace = typeutils.NonEmptyStringOrNull(item.Namespace)
	m.CreatedAt = types.StringValue(item.CreatedAt)
	m.UpdatedAt = types.StringValue(item.UpdatedAt)

	if item.Package.Name != "" || item.Package.Version != "" {
		pkgObj, d := types.ObjectValueFrom(ctx, packageAttrTypes(), packageModel{
			Name:    types.StringValue(item.Package.Name),
			Version: types.StringValue(item.Package.Version),
			Title:   types.StringValue(item.Package.Title),
		})
		diags.Append(d...)
		m.Package = pkgObj
		m.VarsJSON = varsJSONFromAny(item.Vars, item.Package.Name, item.Package.Version, &diags)
	}

	if item.VarGroupSelections != nil && len(*item.VarGroupSelections) > 0 {
		vgs, d := types.MapValueFrom(ctx, types.StringType, *item.VarGroupSelections)
		diags.Append(d...)
		m.VarGroupSelections = vgs
	} else {
		m.VarGroupSelections = types.MapNull(types.StringType)
	}

	if item.AdditionalDatastreamsPermissions != nil && len(*item.AdditionalDatastreamsPermissions) > 0 {
		perms, d := types.ListValueFrom(ctx, types.StringType, *item.AdditionalDatastreamsPermissions)
		diags.Append(d...)
		m.AdditionalDatastreamsPermissions = perms
	} else {
		m.AdditionalDatastreamsPermissions = types.ListNull(types.StringType)
	}

	m.GlobalDataTags = globalDataTagsToModel(ctx, item, &diags)

	inputsKnownKeys := inputsKnownKeySet(m.Inputs)
	m.Inputs = populateInputsFromManagedIntegration(ctx, item, inputsKnownKeys, &diags)

	m.applyCloudConnectorFromAPI(ctx, item, &diags)

	return diags
}
