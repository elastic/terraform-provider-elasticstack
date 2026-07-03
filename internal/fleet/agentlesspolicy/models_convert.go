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

// Package agentlesspolicy: this file implements Task 5's conversion layer
// (openspec/changes/fleet-agentless-policy, "5. Resource: CRUD + import"):
// building the POST /api/fleet/agentless_policies request body from the
// Plugin Framework model (toCreateBody), and populating the model from the
// two distinct API response shapes this resource consumes -- the bundled
// create response (KibanaHTTPAPIsAgentlessPolicy, populateFromCreateResponse)
// and the package_policies read/update response (kbapi.PackagePolicy,
// populateFromPackagePolicy). See design.md Decision 4 for why there are two
// response shapes at all (no dedicated agentless GET/PUT endpoint exists).
//
// Several kbapi request/response fields are anonymous Go structs (oapi-codegen
// emits an unnamed struct type per inline schema property, so e.g.
// KibanaHTTPAPIsAgentlessPolicy.Inputs and
// KibanaHTTPAPIsCreateAgentlessPolicyRequest.GlobalDataTags have no nameable
// Go type). Rather than hand-spelling those anonymous types at every call
// site (fragile, and liable to drift silently out of sync on the next kbapi
// regeneration), this file builds plain map[string]any/[]any trees matching
// the wire shape and converts via a JSON marshal/unmarshal round trip into
// the destination field (e.g. `json.Unmarshal(b, &body.Inputs)`) -- the same
// pattern already used elsewhere in this repo for anonymous API fields (see
// internal/kibana/dashboard/panel/*/api_conv.go) and by
// policyshape.VarsMapToTypedMap.
package agentlesspolicy

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

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

// packageModel, cloudConnectorModel, and globalDataTagModel are the Go
// representations of the `package`, `cloud_connector`, and `global_data_tags`
// nested attributes (see models.go's field-level doc comment). They are
// decoded/encoded via types.Object(List).As/ObjectValueFrom, matching the
// convention used by internal/fleet/agentpolicy's advanced_settings and
// global_data_tags fields.
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

type globalDataTagModel struct {
	Name  types.String `tfsdk:"name"`
	Value types.String `tfsdk:"value"`
}

// agentlessInputModel is the Go representation of a single `inputs` map
// element. It deliberately mirrors agentlessInputAttributeTypes() in
// schema.go (enabled/condition/vars/streams) rather than reusing
// policyshape.InputModel directly: policyshape.InputModel also declares a
// `defaults` field that agentlessInputAttributeTypes() does not surface (see
// schema.go's agentlessInputType doc comment), and the Plugin Framework's
// object-to-struct decoding requires the struct's tfsdk fields to match the
// object's attribute set.
type agentlessInputModel struct {
	Enabled   types.Bool           `tfsdk:"enabled"`
	Condition types.String         `tfsdk:"condition"`
	Vars      jsontypes.Normalized `tfsdk:"vars"`
	Streams   types.Map            `tfsdk:"streams"` // > policyshape.InputStreamModel
}

// Map/attribute key constants reused across this file's raw wire-format
// builders and attr-type helpers. Consolidated into named constants (rather
// than repeating the string literals) both for consistency and because
// schema.go already uses each literal at least once, so repeating them again
// here would trip golangci-lint's goconst check.
const (
	keyEnabled          = "enabled"
	keyCloudConnectorID = "cloud_connector_id"
	keyValue            = "value"
)

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

func globalDataTagAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		attrName: types.StringType,
		keyValue: types.StringType,
	}
}

// mappedInputWire and mappedStreamWire mirror the on-wire shape of a single
// input/stream in the Fleet "mapped"/simplified package-policy format,
// empirically confirmed against a live Kibana 9.4.3 deployment (input keys
// are "<policy_template>-<input_type>"; stream keys are the dataset name --
// see mappedInputKey and update.go's header comment). Both the agentless
// create response (KibanaHTTPAPIsAgentlessPolicy.Inputs) and the
// package-policies read/update response (kbapi.PackagePolicyMappedInputs,
// obtained via GetPackagePolicy/UpdatePackagePolicy's Format=Simplified) use
// this wire shape, but oapi-codegen gives each a distinct (in the create
// case, anonymous) Go type. decodeMappedInputs normalizes either into this
// shared struct via a JSON round trip so populateInputsModel only needs to be
// written once.
type mappedInputWire struct {
	Enabled   *bool                       `json:"enabled,omitempty"`
	Condition *string                     `json:"condition,omitempty"`
	Vars      map[string]any              `json:"vars,omitempty"`
	Streams   map[string]mappedStreamWire `json:"streams,omitempty"`
}

type mappedStreamWire struct {
	Enabled   *bool          `json:"enabled,omitempty"`
	Condition *string        `json:"condition,omitempty"`
	Vars      map[string]any `json:"vars,omitempty"`
}

// decodeMappedInputs normalizes any mapped-format inputs value (from either
// response shape described above) into mappedInputWire via a JSON round trip.
func decodeMappedInputs(raw any) (map[string]mappedInputWire, error) {
	b, err := json.Marshal(raw)
	if err != nil {
		return nil, err
	}
	if len(b) == 0 || string(b) == "null" {
		return nil, nil
	}
	var out map[string]mappedInputWire
	if err := json.Unmarshal(b, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// globalDataTagWire mirrors the on-wire shape of a single global_data_tags
// entry (`{name, value}`, value a string|number union), shared by the
// agentless create response and the package-policies response.
type globalDataTagWire struct {
	Name  string `json:"name"`
	Value any    `json:"value"`
}

func decodeGlobalDataTags(raw any) ([]globalDataTagWire, error) {
	b, err := json.Marshal(raw)
	if err != nil {
		return nil, err
	}
	if len(b) == 0 || string(b) == "null" {
		return nil, nil
	}
	var out []globalDataTagWire
	if err := json.Unmarshal(b, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// mappedInputKey builds the "<policy_template>-<input_type>" key the Fleet
// mapped/simplified package-policy format uses to identify an input,
// confirmed empirically against a live Kibana 9.4.3 deployment (see
// update.go's header comment and design.md Decision 3). When policyTemplate
// is nil or empty, the key is just the input type.
func mappedInputKey(policyTemplate *string, inputType string) string {
	if policyTemplate == nil || *policyTemplate == "" {
		return inputType
	}
	return *policyTemplate + "-" + inputType
}

// globalDataTagValueToString renders a decoded global_data_tags value
// (string|number per the API union) as a string; this resource's schema only
// models string tag values (schema.go's global_data_tags nested object).
func globalDataTagValueToString(v any) string {
	switch val := v.(type) {
	case string:
		return val
	case float64:
		return strconv.FormatFloat(val, 'f', -1, 64)
	case nil:
		return ""
	default:
		return fmt.Sprint(val)
	}
}

// globalDataTagsToModel converts decoded wire tags into the `global_data_tags`
// list attribute, or a null list when there are none.
func globalDataTagsToModel(ctx context.Context, wire []globalDataTagWire, diags *diag.Diagnostics) types.List {
	attrTypes := globalDataTagAttrTypes()
	if len(wire) == 0 {
		return types.ListNull(types.ObjectType{AttrTypes: attrTypes})
	}

	tags := make([]globalDataTagModel, 0, len(wire))
	for _, w := range wire {
		tags = append(tags, globalDataTagModel{
			Name:  types.StringValue(w.Name),
			Value: types.StringValue(globalDataTagValueToString(w.Value)),
		})
	}

	list, d := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: attrTypes}, tags)
	diags.Append(d...)
	return list
}

// globalDataTagsRawFromModel converts the `global_data_tags` list attribute
// into a slice of plain {"name":..., "value":...} maps, suitable for a JSON
// marshal/unmarshal round trip into a request body's (structurally anonymous)
// GlobalDataTags field. Returns nil when list is null or unknown.
func globalDataTagsRawFromModel(ctx context.Context, list types.List, diags *diag.Diagnostics) []map[string]any {
	if !typeutils.IsKnown(list) {
		return nil
	}
	tags := typeutils.ListTypeAs[globalDataTagModel](ctx, list, path.Root("global_data_tags"), diags)
	raw := make([]map[string]any, 0, len(tags))
	for _, t := range tags {
		raw = append(raw, map[string]any{attrName: t.Name.ValueString(), keyValue: t.Value.ValueString()})
	}
	return raw
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
		diags.AddError("Failed to marshal vars", err.Error())
		return policyshape.NewVarsJSONNull()
	}

	v, d := policyshape.NewVarsJSONWithIntegration(string(b), packageName, packageVersion, lookupCachedPackageInfo)
	diags.Append(d...)
	return v
}

// inputsKnownKeySet captures the set of keys of inputs's map value, or nil if
// inputs is not Known (null or unknown) -- see populateInputsModel's knownKeys
// parameter for how this is used to filter an API response before it
// overwrites the very model this was read from.
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

// populateInputsModel builds the `inputs` map attribute from decoded wire
// inputs (see decodeMappedInputs), shared by populateFromCreateResponse and
// populateFromPackagePolicy so the conversion is only written once despite
// the two API response shapes.
//
// knownKeys filters wireInputs down to the given key set before decoding, or
// performs no filtering at all when nil. This works around a real behavior
// of both the POST /api/fleet/agentless_policies response and the
// package_policies GET/PUT (?format=simplified) response for
// multi-policy-template packages such as cloud_security_posture (CSPM):
// empirically (see the fleet-agentless-policy OpenSpec change's Task 8
// acceptance-test work), both responses include an entry for *every* input
// declared by the package across *all* of its policy templates (e.g. CSPM's
// kspm-cloudbeat/cis_k8s, kspm-cloudbeat/cis_eks, cspm-cloudbeat/cis_gcp,
// cspm-cloudbeat/cis_azure, and vuln_mgmt-cloudbeat/vuln_mgmt_aws all appear
// disabled, in addition to whichever single input the config actually
// enables), not just the input(s) the policy_template/config selected. The
// `inputs` schema attribute is Optional+Computed (schema.go), so when config
// sets a known, explicit inputs map (e.g. one entry), Terraform requires the
// applied state to contain *exactly* those keys -- the framework hard-fails
// with "Provider produced inconsistent result after apply: .inputs: new
// element ... has appeared" otherwise. Filtering the response down to the
// caller-supplied knownKeys (the model's own Inputs map value captured
// *before* this function overwrites it -- the plan's value on Create, or the
// prior model's value on Read/Update) makes state mirror what was actually
// planned. When inputs is unset entirely (Unknown on Create, or Known from
// prior state on Read/Update), knownKeys is nil and no filtering happens,
// preserving the widest-possible-response behavior for that case.
func populateInputsModel(ctx context.Context, wireInputs map[string]mappedInputWire, knownKeys map[string]struct{}, diags *diag.Diagnostics) policyshape.InputsValue {
	if knownKeys != nil {
		filtered := make(map[string]mappedInputWire, len(knownKeys))
		for k := range knownKeys {
			if wire, ok := wireInputs[k]; ok {
				filtered[k] = wire
			}
		}
		wireInputs = filtered
	}

	if len(wireInputs) == 0 {
		return policyshape.NewInputsNull(agentlessInputType())
	}

	models := make(map[string]agentlessInputModel, len(wireInputs))
	for inputID, wire := range wireInputs {
		inputPath := path.Root("inputs").AtMapKey(inputID)

		m := agentlessInputModel{
			Enabled:   types.BoolPointerValue(wire.Enabled),
			Condition: types.StringPointerValue(wire.Condition),
			Vars:      typeutils.MarshalToNormalized(wire.Vars, inputPath.AtName("vars"), diags),
		}

		if len(wire.Streams) > 0 {
			streams := make(map[string]policyshape.InputStreamModel, len(wire.Streams))
			for streamID, sw := range wire.Streams {
				streamPath := inputPath.AtName("streams").AtMapKey(streamID)
				streams[streamID] = policyshape.InputStreamModel{
					Enabled:   types.BoolPointerValue(sw.Enabled),
					Condition: types.StringPointerValue(sw.Condition),
					Vars:      typeutils.MarshalToNormalized(sw.Vars, streamPath.AtName("vars"), diags),
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

	inputsValue, d := policyshape.NewInputsValueFrom(ctx, agentlessInputType(), models)
	diags.Append(d...)
	return inputsValue
}

// decodedAgentlessInput is the once-decoded form of a single `inputs` map
// element, with its nested `streams` map (if any) already decoded too.
// decodeInputs produces this so that validateInputConditionSupport and the
// request-body builders (applyCreateInputs, buildUpdateInputs) don't each
// independently re-run the reflection-based typeutils.MapTypeAs decode over
// the same inputs+streams structure -- mirroring
// internal/fleet/integration_policy/models.go's decodedInput/decodeInputs.
type decodedAgentlessInput struct {
	model   agentlessInputModel
	streams map[string]policyshape.InputStreamModel // nil if streams is null/unknown
}

// decodeInputs decodes the `inputs` attribute, and each input's nested
// `streams` map, exactly once. Returns nil if `inputs` itself is null/unknown
// or fails to decode.
func (m agentlessPolicyModel) decodeInputs(ctx context.Context, diags *diag.Diagnostics) map[string]decodedAgentlessInput {
	if !typeutils.IsKnown(m.Inputs.MapValue) {
		return nil
	}

	inputsMap := typeutils.MapTypeAs[agentlessInputModel](ctx, m.Inputs.MapValue, path.Root("inputs"), diags)
	if inputsMap == nil {
		return nil
	}

	decoded := make(map[string]decodedAgentlessInput, len(inputsMap))
	for inputID, inputModel := range inputsMap {
		d := decodedAgentlessInput{model: inputModel}
		if typeutils.IsKnown(inputModel.Streams) {
			inputPath := path.Root("inputs").AtMapKey(inputID)
			d.streams = typeutils.MapTypeAs[policyshape.InputStreamModel](ctx, inputModel.Streams, inputPath.AtName("streams"), diags)
		}
		decoded[inputID] = d
	}
	return decoded
}

// validateInputConditionSupport returns an attribute-scoped error diagnostic
// for every input/stream `condition` value that is set when the connected
// Kibana version does not support the `condition` field on package-policy
// inputs/streams (added in Kibana 9.5.0; see policyshape.MinVersionCondition).
// It is a no-op when the version supports condition or when no inputs are
// configured. Mirrors internal/fleet/integration_policy/models.go's
// validateConditionSupport -- the diagnostic wording is kept identical on
// purpose, for consistency across both resources that surface `condition` on
// package-policy inputs/streams.
func validateInputConditionSupport(decoded map[string]decodedAgentlessInput, supportsCondition bool) diag.Diagnostics {
	var diags diag.Diagnostics

	if supportsCondition {
		return diags
	}

	for inputID, di := range decoded {
		inputPath := path.Root("inputs").AtMapKey(inputID)

		if typeutils.IsKnown(di.model.Condition) {
			diags.AddAttributeError(
				inputPath.AtName("condition"),
				"Unsupported Elasticsearch version",
				fmt.Sprintf("Input condition is only supported in Elastic Stack %s and above", policyshape.MinVersionCondition),
			)
		}

		for streamID, streamModel := range di.streams {
			if typeutils.IsKnown(streamModel.Condition) {
				diags.AddAttributeError(
					inputPath.AtName("streams").AtMapKey(streamID).AtName("condition"),
					"Unsupported Elasticsearch version",
					fmt.Sprintf("Stream condition is only supported in Elastic Stack %s and above", policyshape.MinVersionCondition),
				)
			}
		}
	}

	return diags
}

// toCreateBody implements Task 5.1: compiles the config/plan model into
// PostFleetAgentlessPoliciesJSONRequestBody. Per spec: cloud_connector is
// omitted entirely when the block is not present in config (typeutils.IsKnown
// returns false for a null, non-Computed attribute), but sent -- even with
// only `enabled` set -- when the block is present; force and
// create_dataset_templates are sent on create only.
func (m agentlessPolicyModel) toCreateBody(ctx context.Context, feat agentlessPolicyFeatures) (kbapi.PostFleetAgentlessPoliciesJSONRequestBody, diag.Diagnostics) {
	var diags diag.Diagnostics

	// Decode `inputs` (and each input's nested `streams`) exactly once, and
	// validate `condition` support against the connected Kibana's version
	// before doing any further work: see validateInputConditionSupport's doc
	// comment and capabilities.go's resolveAgentlessPolicyFeatures. A rejected
	// condition here surfaces as a clean attribute-scoped Terraform
	// diagnostic instead of a raw Kibana 400 from the eventual POST.
	decodedInputs := m.decodeInputs(ctx, &diags)
	if condDiags := validateInputConditionSupport(decodedInputs, feat.SupportsCondition); condDiags.HasError() {
		diags.Append(condDiags...)
		return kbapi.PostFleetAgentlessPoliciesJSONRequestBody{}, diags
	}

	var pkg packageModel
	diags.Append(m.Package.As(ctx, &pkg, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return kbapi.PostFleetAgentlessPoliciesJSONRequestBody{}, diags
	}

	body := kbapi.PostFleetAgentlessPoliciesJSONRequestBody{
		Name: m.Name.ValueString(),
		Package: kbapi.KibanaHTTPAPIsPackagePolicyPackage{
			Name:    pkg.Name.ValueString(),
			Version: pkg.Version.ValueString(),
		},
	}

	if typeutils.IsKnown(pkg.Title) {
		body.Package.Title = pkg.Title.ValueStringPointer()
	}
	if typeutils.IsKnown(m.PolicyID) {
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
	if typeutils.IsKnown(m.Force) {
		body.Force = m.Force.ValueBoolPointer()
	}
	if typeutils.IsKnown(m.CreateDatasetTemplates) {
		body.CreateDatasetTemplates = m.CreateDatasetTemplates.ValueBoolPointer()
	}

	if typeutils.IsKnown(m.VarsJSON) {
		sanitized, sd := m.VarsJSON.SanitizedValue()
		diags.Append(sd...)
		if !diags.HasError() {
			varsMap := typeutils.NormalizedTypeToMap[any](jsontypes.NewNormalizedValue(sanitized), path.Root("vars_json"), &diags)
			if len(varsMap) > 0 {
				if b, err := json.Marshal(varsMap); err != nil {
					diags.AddAttributeError(path.Root("vars_json"), "Failed to encode vars_json", err.Error())
				} else if err := json.Unmarshal(b, &body.Vars); err != nil {
					diags.AddAttributeError(path.Root("vars_json"), "Failed to encode vars_json for the create request", err.Error())
				}
			}
		}
	}

	if typeutils.IsKnown(m.VarGroupSelections) {
		vgs := map[string]string{}
		diags.Append(m.VarGroupSelections.ElementsAs(ctx, &vgs, false)...)
		if len(vgs) > 0 {
			body.VarGroupSelections = &vgs
		}
	}

	if typeutils.IsKnown(m.AdditionalDatastreamsPermissions) {
		var perms []string
		diags.Append(m.AdditionalDatastreamsPermissions.ElementsAs(ctx, &perms, false)...)
		body.AdditionalDatastreamsPermissions = &perms
	}

	if typeutils.IsKnown(m.GlobalDataTags) {
		raw := globalDataTagsRawFromModel(ctx, m.GlobalDataTags, &diags)
		if b, err := json.Marshal(raw); err != nil {
			diags.AddAttributeError(path.Root("global_data_tags"), "Failed to encode global_data_tags", err.Error())
		} else if err := json.Unmarshal(b, &body.GlobalDataTags); err != nil {
			diags.AddAttributeError(path.Root("global_data_tags"), "Failed to encode global_data_tags for the create request", err.Error())
		}
	}

	if typeutils.IsKnown(m.CloudConnector) {
		var cc cloudConnectorModel
		diags.Append(m.CloudConnector.As(ctx, &cc, basetypes.ObjectAsOptions{})...)

		raw := map[string]any{}
		if typeutils.IsKnown(cc.Enabled) {
			raw[keyEnabled] = cc.Enabled.ValueBool()
		}
		if typeutils.IsKnown(cc.CloudConnectorID) {
			raw[keyCloudConnectorID] = cc.CloudConnectorID.ValueString()
		}
		if typeutils.IsKnown(cc.Name) {
			raw[attrName] = cc.Name.ValueString()
		}
		if typeutils.IsKnown(cc.TargetCSP) {
			raw["target_csp"] = cc.TargetCSP.ValueString()
		}

		if b, err := json.Marshal(raw); err != nil {
			diags.AddAttributeError(path.Root("cloud_connector"), "Failed to encode cloud_connector", err.Error())
		} else if err := json.Unmarshal(b, &body.CloudConnector); err != nil {
			diags.AddAttributeError(path.Root("cloud_connector"), "Failed to encode cloud_connector for the create request", err.Error())
		}
	}

	applyCreateInputs(&body, decodedInputs, &diags)

	return body, diags
}

// applyCreateInputs converts the already-decoded `inputs` map (see
// decodeInputs) into the create body's Inputs field (a
// map[string]struct{...} of anonymous Go type -- see this file's header
// comment) via a JSON marshal/unmarshal round trip.
func applyCreateInputs(body *kbapi.PostFleetAgentlessPoliciesJSONRequestBody, decoded map[string]decodedAgentlessInput, diags *diag.Diagnostics) {
	if len(decoded) == 0 {
		return
	}

	raw := map[string]any{}
	for inputID, di := range decoded {
		in := di.model
		inputPath := path.Root("inputs").AtMapKey(inputID)
		entry := map[string]any{}

		if typeutils.IsKnown(in.Enabled) {
			entry[keyEnabled] = in.Enabled.ValueBool()
		}
		if typeutils.IsKnown(in.Condition) {
			entry["condition"] = in.Condition.ValueString()
		}
		if varsMap := typeutils.NormalizedTypeToMap[any](in.Vars, inputPath.AtName("vars"), diags); len(varsMap) > 0 {
			entry["vars"] = varsMap
		}

		if len(di.streams) > 0 {
			streamsRaw := map[string]any{}
			for streamID, s := range di.streams {
				streamPath := inputPath.AtName("streams").AtMapKey(streamID)
				streamEntry := map[string]any{}
				if typeutils.IsKnown(s.Enabled) {
					streamEntry[keyEnabled] = s.Enabled.ValueBool()
				}
				if typeutils.IsKnown(s.Condition) {
					streamEntry["condition"] = s.Condition.ValueString()
				}
				if sv := typeutils.NormalizedTypeToMap[any](s.Vars, streamPath.AtName("vars"), diags); len(sv) > 0 {
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
		diags.AddAttributeError(path.Root("inputs"), "Failed to encode inputs", err.Error())
		return
	}
	if err := json.Unmarshal(b, &body.Inputs); err != nil {
		diags.AddAttributeError(path.Root("inputs"), "Failed to encode inputs for the create request", err.Error())
	}
}

// responseFields captures the response fields that populateFromCreateResponse
// and populateFromPackagePolicy both apply onto the model -- the two response
// shapes they read from (KibanaHTTPAPIsAgentlessPolicy and
// kbapi.PackagePolicy respectively) diverge in pointer-vs-value fields and in
// how `inputs` is boxed (a plain raw value vs. a union type requiring
// AsPackagePolicyMappedInputs() first), so each caller extracts its own
// fields into this common shape and applyResponseFields does the actual
// state-population work once.
type responseFields struct {
	name        string
	description *string
	namespace   *string
	createdAt   string
	updatedAt   string

	// spaceIDs is the response's own space_ids, if the response shape
	// carries one (only kbapi.PackagePolicy does). nil (not just empty) means
	// "this response shape has no such field" -- applyResponseFields falls
	// through to defaulting m.SpaceIDs from spaceID when unset, exactly as
	// when the slice is present but empty.
	spaceIDs *[]string

	// hasPackage mirrors kbapi.PackagePolicy's Package pointer being nil:
	// KibanaHTTPAPIsAgentlessPolicy's Package is never nil, so
	// populateFromCreateResponse always passes true. When false, m.Package
	// and m.VarsJSON are left untouched, matching populateFromPackagePolicy's
	// original `if data.Package != nil` guard.
	hasPackage     bool
	packageName    string
	packageVersion string
	packageTitle   types.String
	varsRaw        any

	varGroupSelections               *map[string]string
	additionalDatastreamsPermissions *[]string
	globalDataTagsRaw                any
	inputsRaw                        any

	// decodeErrContext is appended to the two decode-failure diagnostic
	// messages below, preserving each caller's original wording ("... from
	// the create response" vs. no suffix at all).
	decodeErrContext string
}

// applyResponseFields is the shared body of populateFromCreateResponse and
// populateFromPackagePolicy: given the id already extracted by the caller
// (the two response shapes use different id field types) and a responseFields
// populated from the caller's own response, it sets every model field the two
// functions have in common. See responseFields' doc comment for how the two
// callers' response-shape differences are normalized before calling this.
func (m *agentlessPolicyModel) applyResponseFields(ctx context.Context, spaceID, resourceID string, f responseFields, diags *diag.Diagnostics) {
	m.ID = types.StringValue((&clients.CompositeID{ClusterID: spaceID, ResourceID: resourceID}).String())
	m.PolicyID = types.StringValue(resourceID)

	if f.spaceIDs != nil && len(*f.spaceIDs) > 0 {
		spaceIDs, d := types.SetValueFrom(ctx, types.StringType, *f.spaceIDs)
		diags.Append(d...)
		m.SpaceIDs = spaceIDs
	} else if !typeutils.IsKnown(m.SpaceIDs) {
		spaceIDs, d := types.SetValueFrom(ctx, types.StringType, []string{spaceID})
		diags.Append(d...)
		m.SpaceIDs = spaceIDs
	}

	m.Name = types.StringValue(f.name)
	// description is Optional but NOT Computed in schema.go, so state must
	// exactly mirror an unset (null) config value. Empirically, Kibana's
	// package-policy endpoints return an explicit "" (not an omitted/nil
	// field) once a description has been cleared via an update -- see
	// update.go's buildUpdateBody, which sends "" rather than omitting the
	// field to actively clear it. NonEmptyStringOrNull folds that "" back to
	// null so a cleared description doesn't show up as a permanent diff
	// against a config that never set it. namespace gets the same treatment
	// for consistency, even though it is Computed (so a "" vs null mismatch
	// there is a milder no-op-diff annoyance, not a hard error).
	m.Description = typeutils.NonEmptyStringOrNull(f.description)
	m.Namespace = typeutils.NonEmptyStringOrNull(f.namespace)
	m.CreatedAt = types.StringValue(f.createdAt)
	m.UpdatedAt = types.StringValue(f.updatedAt)

	if f.hasPackage {
		pkgObj, d := types.ObjectValueFrom(ctx, packageAttrTypes(), packageModel{
			Name:    types.StringValue(f.packageName),
			Version: types.StringValue(f.packageVersion),
			Title:   f.packageTitle,
		})
		diags.Append(d...)
		m.Package = pkgObj

		m.VarsJSON = varsJSONFromAny(f.varsRaw, f.packageName, f.packageVersion, diags)
	}

	if f.varGroupSelections != nil && len(*f.varGroupSelections) > 0 {
		vgs, d := types.MapValueFrom(ctx, types.StringType, *f.varGroupSelections)
		diags.Append(d...)
		m.VarGroupSelections = vgs
	} else {
		m.VarGroupSelections = types.MapNull(types.StringType)
	}

	if f.additionalDatastreamsPermissions != nil && len(*f.additionalDatastreamsPermissions) > 0 {
		perms, d := types.ListValueFrom(ctx, types.StringType, *f.additionalDatastreamsPermissions)
		diags.Append(d...)
		m.AdditionalDatastreamsPermissions = perms
	} else {
		m.AdditionalDatastreamsPermissions = types.ListNull(types.StringType)
	}

	tagsWire, err := decodeGlobalDataTags(f.globalDataTagsRaw)
	if err != nil {
		diags.AddError("Failed to decode global_data_tags"+f.decodeErrContext, err.Error())
	} else {
		m.GlobalDataTags = globalDataTagsToModel(ctx, tagsWire, diags)
	}

	inputsKnownKeys := inputsKnownKeySet(m.Inputs)
	inputsWire, err := decodeMappedInputs(f.inputsRaw)
	if err != nil {
		diags.AddError("Failed to decode inputs"+f.decodeErrContext, err.Error())
	} else {
		m.Inputs = populateInputsModel(ctx, inputsWire, inputsKnownKeys, diags)
	}
}

// populateFromCreateResponse implements the "Create" requirement's response
// decoding (specs/fleet-agentless-policy/spec.md): policy_id and id are set
// from the response's id field; created_at/updated_at come from the response.
//
// space_ids, cloud_connector, policy_template, force, force_delete, and
// create_dataset_templates are intentionally left untouched beyond what the
// plan already set: none of them round-trip through
// KibanaHTTPAPIsAgentlessPolicy (Decision 4), none are Computed in schema.go
// except space_ids (defaulted below when unset), and touching the others
// here risks a "Provider produced inconsistent result" error against the
// plan value the framework already validated.
func (m *agentlessPolicyModel) populateFromCreateResponse(ctx context.Context, spaceID string, item kbapi.KibanaHTTPAPIsAgentlessPolicy) diag.Diagnostics {
	var diags diag.Diagnostics

	m.applyResponseFields(ctx, spaceID, item.Id, responseFields{
		name:                             item.Name,
		description:                      item.Description,
		namespace:                        item.Namespace,
		createdAt:                        item.CreatedAt,
		updatedAt:                        item.UpdatedAt,
		hasPackage:                       true,
		packageName:                      item.Package.Name,
		packageVersion:                   item.Package.Version,
		packageTitle:                     types.StringValue(item.Package.Title),
		varsRaw:                          item.Vars,
		varGroupSelections:               item.VarGroupSelections,
		additionalDatastreamsPermissions: item.AdditionalDatastreamsPermissions,
		globalDataTagsRaw:                item.GlobalDataTags,
		inputsRaw:                        item.Inputs,
		decodeErrContext:                 " from the create response",
	}, &diags)

	return diags
}

// populateFromPackagePolicy implements the "Read" requirement (specs/
// fleet-agentless-policy/spec.md): state is updated from every API-populated
// field. force, force_delete, and create_dataset_templates are preserved
// (left untouched on m) since none are returned by GET
// /api/fleet/package_policies/{id} -- see the "Read preserves force_delete"
// and "Create-only flags are not round-tripped from the API" scenarios.
// cloud_connector and policy_template are likewise left untouched: neither
// is Computed in schema.go, so overwriting them from partial API data (the
// response includes cloud_connector_id/enabled but never name/target_csp)
// risks a "Provider produced inconsistent result" error.
func (m *agentlessPolicyModel) populateFromPackagePolicy(ctx context.Context, spaceID string, data *kbapi.PackagePolicy) diag.Diagnostics {
	var diags diag.Diagnostics
	if data == nil {
		return diags
	}

	// Extract mapped inputs from the union Inputs field (Format=Simplified,
	// set by fleet.GetPackagePolicy/fleet.UpdatePackagePolicy, returns mapped
	// inputs). An empty/nil union is treated as no inputs rather than an
	// error, matching internal/fleet/integration_policy/models.go.
	mappedInputs, err := data.Inputs.AsPackagePolicyMappedInputs()
	if err != nil {
		mappedInputs = kbapi.PackagePolicyMappedInputs{}
	}

	fields := responseFields{
		name:                             data.Name,
		description:                      data.Description,
		namespace:                        data.Namespace,
		createdAt:                        data.CreatedAt,
		updatedAt:                        data.UpdatedAt,
		spaceIDs:                         data.SpaceIds,
		hasPackage:                       data.Package != nil,
		varsRaw:                          data.Vars,
		varGroupSelections:               data.VarGroupSelections,
		additionalDatastreamsPermissions: data.AdditionalDatastreamsPermissions,
		globalDataTagsRaw:                data.GlobalDataTags,
		inputsRaw:                        mappedInputs,
	}
	if data.Package != nil {
		fields.packageName = data.Package.Name
		fields.packageVersion = data.Package.Version
		fields.packageTitle = types.StringPointerValue(data.Package.Title)
	}

	m.applyResponseFields(ctx, spaceID, typeutils.Deref(data.Id), fields, &diags)

	return diags
}
