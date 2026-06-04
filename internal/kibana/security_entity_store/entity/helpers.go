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

package entity

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	jsontypes "github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// Common attribute keys used throughout schema and helpers
const (
	attrName                = "name"
	attrType                = "type"
	attrRisk                = "risk"
	attrAsset               = "asset"
	attrCalculatedLevel     = "calculated_level"
	attrCalculatedScore     = "calculated_score"
	attrCalculatedScoreNorm = "calculated_score_norm"
	attrDomain              = "domain"
	attrEmail               = "email"
	attrProvider            = "provider"
	attrReason              = "reason"
	attrValue               = "value"

	descCalculatedLevel     = "The calculated risk level."
	descCalculatedScore     = "The raw numeric value of the given entity's risk score."
	descCalculatedScoreNorm = "The normalized numeric value of the given entity's risk score."

	// Attribute keys used in maps (to satisfy goconst)
	attrTimestamp           = "@timestamp"
	attrEntity              = "entity"
	attrHost                = "host"
	attrUser                = "user"
	attrService             = "service"
	attrCloud               = "cloud"
	attrOrchestrator        = "orchestrator"
	attrEvent               = "event"
	attrLabels              = "labels"
	attrTags                = "tags"
	attrDocumentJSON        = "document_json"
	attrAttributes          = "attributes"
	attrBehaviors           = "behaviors"
	attrLifecycle           = "lifecycle"
	attrRelationships       = "relationships"
	attrSubType             = "sub_type"
	attrID                  = "id"
	attrSource              = "source"
	attrCriticalityFeedback = "criticality_feedback"
	attrOwner               = "owner"
	attrOs                  = "os"
)

// canonicalJSON normalizes a Go value to canonical JSON (sorted keys).
func canonicalJSON(v any) (string, diag.Diagnostics) {
	var diags diag.Diagnostics
	b, err := json.Marshal(v)
	if err != nil {
		diags.AddError("JSON marshal error", err.Error())
		return "", diags
	}
	var tmp any
	if err := json.Unmarshal(b, &tmp); err != nil {
		diags.AddError("JSON unmarshal error", err.Error())
		return "", diags
	}
	b, err = json.Marshal(tmp)
	if err != nil {
		diags.AddError("JSON marshal error", err.Error())
		return "", diags
	}
	return string(b), diags
}

// canonicalMapJSON returns canonical JSON for a map[string]any, or empty string for nil.
func canonicalMapJSON(m map[string]any) string {
	if m == nil {
		return ""
	}
	s, diags := canonicalJSON(m)
	if diags.HasError() {
		return ""
	}
	return s
}

// getStringValue returns types.StringValue if the key exists and is a string, else Null.
func getStringValue(m map[string]any, key string) types.String {
	if m == nil {
		return types.StringNull()
	}
	if v, ok := m[key].(string); ok {
		return types.StringValue(v)
	}
	return types.StringNull()
}

// getBoolValue returns types.BoolValue if the key exists and is a bool, else Null.
func getBoolValue(m map[string]any, key string) types.Bool {
	if m == nil {
		return types.BoolNull()
	}
	if v, ok := m[key].(bool); ok {
		return types.BoolValue(v)
	}
	return types.BoolNull()
}

// getFloat64Value returns types.Float64Value if the key exists and is numeric, else Null.
func getFloat64Value(m map[string]any, key string) types.Float64 {
	if m == nil {
		return types.Float64Null()
	}
	if v, ok := m[key].(float64); ok {
		return types.Float64Value(v)
	}
	if v, ok := m[key].(int); ok {
		return types.Float64Value(float64(v))
	}
	if v, ok := m[key].(int64); ok {
		return types.Float64Value(float64(v))
	}
	return types.Float64Null()
}

// getStringSetValue converts a []any of strings to a types.Set of strings.
func getStringSetValue(m map[string]any, key string) types.Set {
	if m == nil {
		return types.SetNull(types.StringType)
	}
	raw, ok := m[key]
	if !ok {
		return types.SetNull(types.StringType)
	}
	arr, ok := raw.([]any)
	if !ok {
		return types.SetNull(types.StringType)
	}
	vals := make([]attr.Value, 0, len(arr))
	for _, v := range arr {
		if s, ok := v.(string); ok {
			vals = append(vals, types.StringValue(s))
		}
	}
	set, _ := types.SetValue(types.StringType, vals)
	return set
}

// appendStringSetToMap appends a types.Set of strings to a map as []string if non-empty.
func appendStringSetToMap(m map[string]any, key string, set types.Set) {
	if set.IsNull() || set.IsUnknown() || len(set.Elements()) == 0 {
		return
	}
	vals := make([]string, 0, len(set.Elements()))
	for _, v := range set.Elements() {
		if s, ok := v.(types.String); ok {
			vals = append(vals, s.ValueString())
		}
	}
	if len(vals) > 0 {
		m[key] = vals
	}
}

// ---------------------------------------------------------------------------
// Attribute type helpers
// ---------------------------------------------------------------------------

func BlockAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":            types.StringType,
		attrName:        types.StringType,
		attrType:        types.StringType,
		"sub_type":      types.StringType,
		"source":        types.SetType{ElemType: types.StringType},
		"attributes":    types.ObjectType{AttrTypes: AttributesBlockAttrTypes()},
		"behaviors":     types.ObjectType{AttrTypes: BehaviorsBlockAttrTypes()},
		"lifecycle":     types.ObjectType{AttrTypes: LifecycleBlockAttrTypes()},
		attrRisk:        types.ObjectType{AttrTypes: RiskBlockAttrTypes()},
		"relationships": types.ObjectType{AttrTypes: RelationshipsBlockAttrTypes()},
	}
}

func AttributesBlockAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		attrAsset:     types.BoolType,
		"managed":     types.BoolType,
		"privileged":  types.BoolType,
		"mfa_enabled": types.BoolType,
	}
}

func BehaviorsBlockAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"brute_force_victim": types.BoolType,
		"new_country_login":  types.BoolType,
		"used_usb_device":    types.BoolType,
	}
}

func LifecycleBlockAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"first_seen":    types.StringType,
		"last_seen":     types.StringType,
		"last_activity": types.StringType,
	}
}

func RiskBlockAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		attrCalculatedLevel:     types.StringType,
		attrCalculatedScore:     types.Float64Type,
		attrCalculatedScoreNorm: types.Float64Type,
	}
}

func RelationshipsBlockAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"owned_by":               types.SetType{ElemType: types.StringType},
		"owns":                   types.SetType{ElemType: types.StringType},
		"supervised_by":          types.SetType{ElemType: types.StringType},
		"supervises":             types.SetType{ElemType: types.StringType},
		"depends_on":             types.SetType{ElemType: types.StringType},
		"dependent_of":           types.SetType{ElemType: types.StringType},
		"communicates_with":      types.SetType{ElemType: types.StringType},
		"accesses_frequently":    types.SetType{ElemType: types.StringType},
		"accessed_frequently_by": types.SetType{ElemType: types.StringType},
		"accesses_infrequently":  types.SetType{ElemType: types.StringType},
	}
}

func HostBlockAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		attrName:       types.StringType,
		attrDomain:     types.SetType{ElemType: types.StringType},
		"hostname":     types.SetType{ElemType: types.StringType},
		"id":           types.SetType{ElemType: types.StringType},
		"ip":           types.SetType{ElemType: types.StringType},
		"mac":          types.SetType{ElemType: types.StringType},
		attrType:       types.SetType{ElemType: types.StringType},
		"architecture": types.SetType{ElemType: types.StringType},
		"os":           types.ObjectType{AttrTypes: HostOsBlockAttrTypes()},
		attrRisk:       types.ObjectType{AttrTypes: RiskBlockAttrTypes()},
	}
}

func HostOsBlockAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"family":   types.StringType,
		"full":     types.StringType,
		"kernel":   types.StringType,
		attrName:   types.StringType,
		"platform": types.StringType,
		attrType:   types.StringType,
		"version":  types.StringType,
	}
}

func UserBlockAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		attrName:    types.StringType,
		attrDomain:  types.SetType{ElemType: types.StringType},
		attrEmail:   types.SetType{ElemType: types.StringType},
		"full_name": types.SetType{ElemType: types.StringType},
		"hash":      types.SetType{ElemType: types.StringType},
		"id":        types.SetType{ElemType: types.StringType},
		"roles":     types.SetType{ElemType: types.StringType},
		attrRisk:    types.ObjectType{AttrTypes: RiskBlockAttrTypes()},
	}
}

func ServiceBlockAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		attrName: types.StringType,
		attrRisk: types.ObjectType{AttrTypes: RiskBlockAttrTypes()},
	}
}

func OrchestratorBlockAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		attrName:          types.StringType,
		attrType:          types.StringType,
		"namespace":       types.StringType,
		"cluster_id":      types.StringType,
		"cluster_name":    types.StringType,
		"cluster_version": types.StringType,
		"resource_id":     types.StringType,
		"resource_name":   types.StringType,
		"resource_type":   types.StringType,
	}
}

func CloudBlockAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		attrProvider:   types.StringType,
		"region":       types.StringType,
		"account_id":   types.StringType,
		"account_name": types.StringType,
		"project_id":   types.StringType,
		"project_name": types.StringType,
		"service_name": types.StringType,
	}
}

func EventBlockAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"category":   types.StringType,
		attrType:     types.StringType,
		"dataset":    types.StringType,
		"kind":       types.StringType,
		"outcome":    types.StringType,
		attrProvider: types.StringType,
		"action":     types.StringType,
		"code":       types.StringType,
		"reference":  types.StringType,
		attrReason:   types.StringType,
		"severity":   types.StringType,
		"timezone":   types.StringType,
		"url":        types.StringType,
		"ingested":   types.StringType,
	}
}

func AssetBlockAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"criticality":          types.StringType,
		"criticality_feedback": types.ObjectType{AttrTypes: AssetCriticalityFeedbackBlockAttrTypes()},
		"owner":                types.ObjectType{AttrTypes: AssetOwnerBlockAttrTypes()},
		attrValue:              types.Float64Type,
	}
}

func AssetCriticalityFeedbackBlockAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"notes":    types.StringType,
		attrReason: types.StringType,
	}
}

func AssetOwnerBlockAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		attrName:     types.StringType,
		"department": types.StringType,
		attrEmail:    types.StringType,
		"ext":        types.StringType,
	}
}

// ---------------------------------------------------------------------------
// Model -> API body conversion
// ---------------------------------------------------------------------------

func modelToAPIBody(ctx context.Context, model tfModel) (map[string]any, diag.Diagnostics) {
	body := make(map[string]any)
	var diags diag.Diagnostics

	if !model.Timestamp.IsNull() && !model.Timestamp.IsUnknown() {
		body["@timestamp"] = model.Timestamp.ValueString()
	}

	if !model.Tags.IsNull() && !model.Tags.IsUnknown() {
		tags := make([]string, 0)
		for _, v := range model.Tags.Elements() {
			if s, ok := v.(types.String); ok {
				tags = append(tags, s.ValueString())
			}
		}
		body["tags"] = tags
	}

	if !model.Labels.IsNull() && !model.Labels.IsUnknown() {
		labels := make(map[string]any)
		for k, v := range model.Labels.Elements() {
			if s, ok := v.(types.String); ok {
				labels[k] = s.ValueString()
			}
		}
		body["labels"] = labels
	} else if !model.LabelsJSON.IsNull() && !model.LabelsJSON.IsUnknown() {
		labels := typeutils.NormalizedTypeToMap[any](model.LabelsJSON, path.Root("labels_json"), &diags)
		if labels != nil {
			body["labels"] = labels
		}
	}

	if !model.Entity.IsNull() && !model.Entity.IsUnknown() {
		m := entityBlockToMap(ctx, model.Entity, &diags)
		if m != nil {
			body["entity"] = m
		}
	} else if !model.EntityJSON.IsNull() && !model.EntityJSON.IsUnknown() {
		m := typeutils.NormalizedTypeToMap[any](model.EntityJSON, path.Root("entity_json"), &diags)
		if m != nil {
			body["entity"] = m
		}
	}

	if !model.Host.IsNull() && !model.Host.IsUnknown() {
		m := hostBlockToMap(ctx, model.Host, &diags)
		if m != nil {
			body["host"] = m
		}
	} else if !model.HostJSON.IsNull() && !model.HostJSON.IsUnknown() {
		m := typeutils.NormalizedTypeToMap[any](model.HostJSON, path.Root("host_json"), &diags)
		if m != nil {
			body["host"] = m
		}
	}

	if !model.User.IsNull() && !model.User.IsUnknown() {
		m := userBlockToMap(ctx, model.User, &diags)
		if m != nil {
			body["user"] = m
		}
	} else if !model.UserJSON.IsNull() && !model.UserJSON.IsUnknown() {
		m := typeutils.NormalizedTypeToMap[any](model.UserJSON, path.Root("user_json"), &diags)
		if m != nil {
			body["user"] = m
		}
	}

	if !model.Service.IsNull() && !model.Service.IsUnknown() {
		m := serviceBlockToMap(ctx, model.Service, &diags)
		if m != nil {
			body["service"] = m
		}
	} else if !model.ServiceJSON.IsNull() && !model.ServiceJSON.IsUnknown() {
		m := typeutils.NormalizedTypeToMap[any](model.ServiceJSON, path.Root("service_json"), &diags)
		if m != nil {
			body["service"] = m
		}
	}

	if !model.Orchestrator.IsNull() && !model.Orchestrator.IsUnknown() {
		m := orchestratorBlockToMap(ctx, model.Orchestrator, &diags)
		if m != nil {
			body["orchestrator"] = m
		}
	} else if !model.OrchestratorJSON.IsNull() && !model.OrchestratorJSON.IsUnknown() {
		m := typeutils.NormalizedTypeToMap[any](model.OrchestratorJSON, path.Root("orchestrator_json"), &diags)
		if m != nil {
			body["orchestrator"] = m
		}
	}

	if !model.Cloud.IsNull() && !model.Cloud.IsUnknown() {
		m := cloudBlockToMap(ctx, model.Cloud, &diags)
		if m != nil {
			body["cloud"] = m
		}
	} else if !model.CloudJSON.IsNull() && !model.CloudJSON.IsUnknown() {
		m := typeutils.NormalizedTypeToMap[any](model.CloudJSON, path.Root("cloud_json"), &diags)
		if m != nil {
			body["cloud"] = m
		}
	}

	if !model.Event.IsNull() && !model.Event.IsUnknown() {
		m := eventBlockToMap(ctx, model.Event, &diags)
		if m != nil {
			body["event"] = m
		}
	} else if !model.EventJSON.IsNull() && !model.EventJSON.IsUnknown() {
		m := typeutils.NormalizedTypeToMap[any](model.EventJSON, path.Root("event_json"), &diags)
		if m != nil {
			body["event"] = m
		}
	}

	if !model.Asset.IsNull() && !model.Asset.IsUnknown() {
		m := assetBlockToMap(ctx, model.Asset, &diags)
		if m != nil {
			body[attrAsset] = m
		}
	} else if !model.AssetJSON.IsNull() && !model.AssetJSON.IsUnknown() {
		m := typeutils.NormalizedTypeToMap[any](model.AssetJSON, path.Root("asset_json"), &diags)
		if m != nil {
			body[attrAsset] = m
		}
	}

	return body, diags
}

func entityBlockToMap(ctx context.Context, obj types.Object, diags *diag.Diagnostics) map[string]any {
	if obj.IsNull() || obj.IsUnknown() {
		return nil
	}
	var model entityBlockModel
	d := obj.As(ctx, &model, basetypes.ObjectAsOptions{})
	diags.Append(d...)
	if diags.HasError() {
		return nil
	}

	m := map[string]any{"id": model.ID.ValueString()}
	if !model.Name.IsNull() && !model.Name.IsUnknown() {
		m[attrName] = model.Name.ValueString()
	}
	if !model.Type.IsNull() && !model.Type.IsUnknown() {
		m[attrType] = model.Type.ValueString()
	}
	if !model.SubType.IsNull() && !model.SubType.IsUnknown() {
		m["sub_type"] = model.SubType.ValueString()
	}
	if !model.Source.IsNull() && !model.Source.IsUnknown() {
		appendStringSetToMap(m, "source", model.Source)
	}
	if !model.Attributes.IsNull() && !model.Attributes.IsUnknown() {
		var attr entityAttributesBlockModel
		d := model.Attributes.As(ctx, &attr, basetypes.ObjectAsOptions{})
		diags.Append(d...)
		if !diags.HasError() {
			am := make(map[string]any)
			if !attr.Asset.IsNull() {
				am[attrAsset] = attr.Asset.ValueBool()
			}
			if !attr.Managed.IsNull() {
				am["managed"] = attr.Managed.ValueBool()
			}
			if !attr.Privileged.IsNull() {
				am["privileged"] = attr.Privileged.ValueBool()
			}
			if !attr.MfaEnabled.IsNull() {
				am["mfa_enabled"] = attr.MfaEnabled.ValueBool()
			}
			if len(am) > 0 {
				m["attributes"] = am
			}
		}
	}
	if !model.Behaviors.IsNull() && !model.Behaviors.IsUnknown() {
		var beh entityBehaviorsBlockModel
		d := model.Behaviors.As(ctx, &beh, basetypes.ObjectAsOptions{})
		diags.Append(d...)
		if !diags.HasError() {
			bm := make(map[string]any)
			if !beh.BruteForceVictim.IsNull() {
				bm["brute_force_victim"] = beh.BruteForceVictim.ValueBool()
			}
			if !beh.NewCountryLogin.IsNull() {
				bm["new_country_login"] = beh.NewCountryLogin.ValueBool()
			}
			if !beh.UsedUsbDevice.IsNull() {
				bm["used_usb_device"] = beh.UsedUsbDevice.ValueBool()
			}
			if len(bm) > 0 {
				m["behaviors"] = bm
			}
		}
	}
	if !model.Lifecycle.IsNull() && !model.Lifecycle.IsUnknown() {
		var lc entityLifecycleBlockModel
		d := model.Lifecycle.As(ctx, &lc, basetypes.ObjectAsOptions{})
		diags.Append(d...)
		if !diags.HasError() {
			lm := make(map[string]any)
			if !lc.FirstSeen.IsNull() {
				lm["first_seen"] = lc.FirstSeen.ValueString()
			}
			if !lc.LastSeen.IsNull() {
				lm["last_seen"] = lc.LastSeen.ValueString()
			}
			if !lc.LastActivity.IsNull() {
				lm["last_activity"] = lc.LastActivity.ValueString()
			}
			if len(lm) > 0 {
				m["lifecycle"] = lm
			}
		}
	}
	if !model.Risk.IsNull() && !model.Risk.IsUnknown() {
		if rm := riskBlockToMap(ctx, model.Risk, diags); rm != nil {
			m[attrRisk] = rm
		}
	}
	if !model.Relationships.IsNull() && !model.Relationships.IsUnknown() {
		var rel entityRelationshipsBlockModel
		d := model.Relationships.As(ctx, &rel, basetypes.ObjectAsOptions{})
		diags.Append(d...)
		if !diags.HasError() {
			rm := make(map[string]any)
			appendStringSetToMap(rm, "owned_by", rel.OwnedBy)
			appendStringSetToMap(rm, "owns", rel.Owns)
			appendStringSetToMap(rm, "supervised_by", rel.SupervisedBy)
			appendStringSetToMap(rm, "supervises", rel.Supervises)
			appendStringSetToMap(rm, "depends_on", rel.DependsOn)
			appendStringSetToMap(rm, "dependent_of", rel.DependentOf)
			appendStringSetToMap(rm, "communicates_with", rel.CommunicatesWith)
			appendStringSetToMap(rm, "accesses_frequently", rel.AccessesFrequently)
			appendStringSetToMap(rm, "accessed_frequently_by", rel.AccessedFrequentlyBy)
			appendStringSetToMap(rm, "accesses_infrequently", rel.AccessesInfrequently)
			if len(rm) > 0 {
				m["relationships"] = rm
			}
		}
	}
	return m
}

func riskBlockToMap(ctx context.Context, obj types.Object, diags *diag.Diagnostics) map[string]any {
	if obj.IsNull() || obj.IsUnknown() {
		return nil
	}
	var model entityRiskBlockModel
	d := obj.As(ctx, &model, basetypes.ObjectAsOptions{})
	diags.Append(d...)
	if diags.HasError() {
		return nil
	}
	m := make(map[string]any)
	if !model.CalculatedLevel.IsNull() {
		m[attrCalculatedLevel] = model.CalculatedLevel.ValueString()
	}
	if !model.CalculatedScore.IsNull() {
		m[attrCalculatedScore] = model.CalculatedScore.ValueFloat64()
	}
	if !model.CalculatedScoreNorm.IsNull() {
		m[attrCalculatedScoreNorm] = model.CalculatedScoreNorm.ValueFloat64()
	}
	return m
}

func hostBlockToMap(ctx context.Context, obj types.Object, diags *diag.Diagnostics) map[string]any {
	if obj.IsNull() || obj.IsUnknown() {
		return nil
	}
	var model hostBlockModel
	d := obj.As(ctx, &model, basetypes.ObjectAsOptions{})
	diags.Append(d...)
	if diags.HasError() {
		return nil
	}
	m := map[string]any{attrName: model.Name.ValueString()}
	appendStringSetToMap(m, attrDomain, model.Domain)
	appendStringSetToMap(m, "hostname", model.Hostname)
	appendStringSetToMap(m, "id", model.ID)
	appendStringSetToMap(m, "ip", model.IP)
	appendStringSetToMap(m, "mac", model.Mac)
	appendStringSetToMap(m, attrType, model.Type)
	appendStringSetToMap(m, "architecture", model.Architecture)
	if !model.Os.IsNull() && !model.Os.IsUnknown() {
		var osModel hostOsBlockModel
		d := model.Os.As(ctx, &osModel, basetypes.ObjectAsOptions{})
		diags.Append(d...)
		if !diags.HasError() {
			om := make(map[string]any)
			if !osModel.Family.IsNull() {
				om["family"] = osModel.Family.ValueString()
			}
			if !osModel.Full.IsNull() {
				om["full"] = osModel.Full.ValueString()
			}
			if !osModel.Kernel.IsNull() {
				om["kernel"] = osModel.Kernel.ValueString()
			}
			if !osModel.Name.IsNull() {
				om[attrName] = osModel.Name.ValueString()
			}
			if !osModel.Platform.IsNull() {
				om["platform"] = osModel.Platform.ValueString()
			}
			if !osModel.Type.IsNull() {
				om[attrType] = osModel.Type.ValueString()
			}
			if !osModel.Version.IsNull() {
				om["version"] = osModel.Version.ValueString()
			}
			if len(om) > 0 {
				m["os"] = om
			}
		}
	}
	if !model.Risk.IsNull() && !model.Risk.IsUnknown() {
		if rm := riskBlockToMap(ctx, model.Risk, diags); rm != nil {
			m[attrRisk] = rm
		}
	}
	return m
}

func userBlockToMap(ctx context.Context, obj types.Object, diags *diag.Diagnostics) map[string]any {
	if obj.IsNull() || obj.IsUnknown() {
		return nil
	}
	var model userBlockModel
	d := obj.As(ctx, &model, basetypes.ObjectAsOptions{})
	diags.Append(d...)
	if diags.HasError() {
		return nil
	}
	m := map[string]any{attrName: model.Name.ValueString()}
	appendStringSetToMap(m, attrDomain, model.Domain)
	appendStringSetToMap(m, attrEmail, model.Email)
	appendStringSetToMap(m, "full_name", model.FullName)
	appendStringSetToMap(m, "hash", model.Hash)
	appendStringSetToMap(m, "id", model.ID)
	appendStringSetToMap(m, "roles", model.Roles)
	if !model.Risk.IsNull() && !model.Risk.IsUnknown() {
		if rm := riskBlockToMap(ctx, model.Risk, diags); rm != nil {
			m[attrRisk] = rm
		}
	}
	return m
}

func serviceBlockToMap(ctx context.Context, obj types.Object, diags *diag.Diagnostics) map[string]any {
	if obj.IsNull() || obj.IsUnknown() {
		return nil
	}
	var model serviceBlockModel
	d := obj.As(ctx, &model, basetypes.ObjectAsOptions{})
	diags.Append(d...)
	if diags.HasError() {
		return nil
	}
	m := map[string]any{attrName: model.Name.ValueString()}
	if !model.Risk.IsNull() && !model.Risk.IsUnknown() {
		if rm := riskBlockToMap(ctx, model.Risk, diags); rm != nil {
			m[attrRisk] = rm
		}
	}
	return m
}

func orchestratorBlockToMap(ctx context.Context, obj types.Object, diags *diag.Diagnostics) map[string]any {
	if obj.IsNull() || obj.IsUnknown() {
		return nil
	}
	var model orchestratorBlockModel
	d := obj.As(ctx, &model, basetypes.ObjectAsOptions{})
	diags.Append(d...)
	if diags.HasError() {
		return nil
	}
	m := make(map[string]any)
	if !model.Name.IsNull() {
		m[attrName] = model.Name.ValueString()
	}
	if !model.Type.IsNull() {
		m[attrType] = model.Type.ValueString()
	}
	if !model.Namespace.IsNull() {
		m["namespace"] = model.Namespace.ValueString()
	}
	if !model.ClusterID.IsNull() {
		m["cluster_id"] = model.ClusterID.ValueString()
	}
	if !model.ClusterName.IsNull() {
		m["cluster_name"] = model.ClusterName.ValueString()
	}
	if !model.ClusterVersion.IsNull() {
		m["cluster_version"] = model.ClusterVersion.ValueString()
	}
	if !model.ResourceID.IsNull() {
		m["resource_id"] = model.ResourceID.ValueString()
	}
	if !model.ResourceName.IsNull() {
		m["resource_name"] = model.ResourceName.ValueString()
	}
	if !model.ResourceType.IsNull() {
		m["resource_type"] = model.ResourceType.ValueString()
	}
	return m
}

func cloudBlockToMap(ctx context.Context, obj types.Object, diags *diag.Diagnostics) map[string]any {
	if obj.IsNull() || obj.IsUnknown() {
		return nil
	}
	var model cloudBlockModel
	d := obj.As(ctx, &model, basetypes.ObjectAsOptions{})
	diags.Append(d...)
	if diags.HasError() {
		return nil
	}
	m := make(map[string]any)
	if !model.Provider.IsNull() {
		m[attrProvider] = model.Provider.ValueString()
	}
	if !model.Region.IsNull() {
		m["region"] = model.Region.ValueString()
	}
	if !model.AccountID.IsNull() {
		m["account_id"] = model.AccountID.ValueString()
	}
	if !model.AccountName.IsNull() {
		m["account_name"] = model.AccountName.ValueString()
	}
	if !model.ProjectID.IsNull() {
		m["project_id"] = model.ProjectID.ValueString()
	}
	if !model.ProjectName.IsNull() {
		m["project_name"] = model.ProjectName.ValueString()
	}
	if !model.ServiceName.IsNull() {
		m["service_name"] = model.ServiceName.ValueString()
	}
	return m
}

func eventBlockToMap(ctx context.Context, obj types.Object, diags *diag.Diagnostics) map[string]any {
	if obj.IsNull() || obj.IsUnknown() {
		return nil
	}
	var model eventBlockModel
	d := obj.As(ctx, &model, basetypes.ObjectAsOptions{})
	diags.Append(d...)
	if diags.HasError() {
		return nil
	}
	m := make(map[string]any)
	if !model.Category.IsNull() {
		m["category"] = model.Category.ValueString()
	}
	if !model.Type.IsNull() {
		m[attrType] = model.Type.ValueString()
	}
	if !model.Dataset.IsNull() {
		m["dataset"] = model.Dataset.ValueString()
	}
	if !model.Kind.IsNull() {
		m["kind"] = model.Kind.ValueString()
	}
	if !model.Outcome.IsNull() {
		m["outcome"] = model.Outcome.ValueString()
	}
	if !model.Provider.IsNull() {
		m[attrProvider] = model.Provider.ValueString()
	}
	if !model.Action.IsNull() {
		m["action"] = model.Action.ValueString()
	}
	if !model.Code.IsNull() {
		m["code"] = model.Code.ValueString()
	}
	if !model.Reference.IsNull() {
		m["reference"] = model.Reference.ValueString()
	}
	if !model.Reason.IsNull() {
		m[attrReason] = model.Reason.ValueString()
	}
	if !model.Severity.IsNull() {
		m["severity"] = model.Severity.ValueString()
	}
	if !model.Timezone.IsNull() {
		m["timezone"] = model.Timezone.ValueString()
	}
	if !model.URL.IsNull() {
		m["url"] = model.URL.ValueString()
	}
	if !model.Ingested.IsNull() {
		m["ingested"] = model.Ingested.ValueString()
	}
	return m
}

func assetBlockToMap(ctx context.Context, obj types.Object, diags *diag.Diagnostics) map[string]any {
	if obj.IsNull() || obj.IsUnknown() {
		return nil
	}
	var model assetBlockModel
	d := obj.As(ctx, &model, basetypes.ObjectAsOptions{})
	diags.Append(d...)
	if diags.HasError() {
		return nil
	}
	m := make(map[string]any)
	if !model.Criticality.IsNull() {
		m["criticality"] = model.Criticality.ValueString()
	}
	if !model.Value.IsNull() {
		m[attrValue] = model.Value.ValueFloat64()
	}
	if !model.CriticalityFeedback.IsNull() && !model.CriticalityFeedback.IsUnknown() {
		var fb assetCriticalityFeedbackBlockModel
		d := model.CriticalityFeedback.As(ctx, &fb, basetypes.ObjectAsOptions{})
		diags.Append(d...)
		if !diags.HasError() {
			fbm := make(map[string]any)
			if !fb.Notes.IsNull() {
				fbm["notes"] = fb.Notes.ValueString()
			}
			if !fb.Reason.IsNull() {
				fbm[attrReason] = fb.Reason.ValueString()
			}
			if len(fbm) > 0 {
				m["criticality_feedback"] = fbm
			}
		}
	}
	if !model.Owner.IsNull() && !model.Owner.IsUnknown() {
		var owner assetOwnerBlockModel
		d := model.Owner.As(ctx, &owner, basetypes.ObjectAsOptions{})
		diags.Append(d...)
		if !diags.HasError() {
			om := make(map[string]any)
			if !owner.Name.IsNull() {
				om[attrName] = owner.Name.ValueString()
			}
			if !owner.Department.IsNull() {
				om["department"] = owner.Department.ValueString()
			}
			if !owner.Email.IsNull() {
				om[attrEmail] = owner.Email.ValueString()
			}
			if !owner.Ext.IsNull() {
				om["ext"] = owner.Ext.ValueString()
			}
			if len(om) > 0 {
				m["owner"] = om
			}
		}
	}
	return m
}

// ---------------------------------------------------------------------------
// API body -> Model conversion
// ---------------------------------------------------------------------------

func apiBodyToModel(ctx context.Context, body map[string]any, model *tfModel, diags *diag.Diagnostics) {
	if ts, ok := body["@timestamp"].(string); ok {
		model.Timestamp = types.StringValue(ts)
	} else {
		model.Timestamp = types.StringNull()
	}

	model.DocumentJSON = jsontypes.NewNormalizedValue(canonicalMapJSON(body))

	if labelsRaw, ok := body["labels"].(map[string]any); ok {
		labelsTyped := make(map[string]attr.Value, len(labelsRaw))
		allStrings := true
		for k, v := range labelsRaw {
			if s, ok := v.(string); ok {
				labelsTyped[k] = types.StringValue(s)
			} else {
				allStrings = false
				break
			}
		}
		if allStrings {
			lv, d := types.MapValue(types.StringType, labelsTyped)
			diags.Append(d...)
			model.Labels = lv
		} else {
			model.Labels = types.MapNull(types.StringType)
		}
		model.LabelsJSON = jsontypes.NewNormalizedValue(canonicalMapJSON(labelsRaw))
	} else {
		model.Labels = types.MapNull(types.StringType)
		model.LabelsJSON = jsontypes.NewNormalizedNull()
	}

	if tagsRaw, ok := body["tags"].([]any); ok {
		tagsVals := make([]attr.Value, 0, len(tagsRaw))
		for _, t := range tagsRaw {
			if s, ok := t.(string); ok {
				tagsVals = append(tagsVals, types.StringValue(s))
			}
		}
		tv, d := types.SetValue(types.StringType, tagsVals)
		diags.Append(d...)
		model.Tags = tv
	} else {
		model.Tags = types.SetNull(types.StringType)
	}

	if entityRaw, ok := body["entity"].(map[string]any); ok {
		em := mapToEntityBlockModel(ctx, entityRaw, diags)
		ev, d := types.ObjectValueFrom(ctx, BlockAttrTypes(), em)
		diags.Append(d...)
		model.Entity = ev
	} else {
		model.Entity = types.ObjectNull(BlockAttrTypes())
	}

	if hostRaw, ok := body["host"].(map[string]any); ok {
		hm := mapToHostBlockModel(ctx, hostRaw, diags)
		hv, d := types.ObjectValueFrom(ctx, HostBlockAttrTypes(), hm)
		diags.Append(d...)
		model.Host = hv
	} else {
		model.Host = types.ObjectNull(HostBlockAttrTypes())
	}

	if userRaw, ok := body["user"].(map[string]any); ok {
		um := mapToUserBlockModel(ctx, userRaw, diags)
		uv, d := types.ObjectValueFrom(ctx, UserBlockAttrTypes(), um)
		diags.Append(d...)
		model.User = uv
	} else {
		model.User = types.ObjectNull(UserBlockAttrTypes())
	}

	if serviceRaw, ok := body["service"].(map[string]any); ok {
		sm := mapToServiceBlockModel(ctx, serviceRaw, diags)
		sv, d := types.ObjectValueFrom(ctx, ServiceBlockAttrTypes(), sm)
		diags.Append(d...)
		model.Service = sv
	} else {
		model.Service = types.ObjectNull(ServiceBlockAttrTypes())
	}

	if orchRaw, ok := body["orchestrator"].(map[string]any); ok {
		om := mapToOrchestratorBlockModel(ctx, orchRaw)
		ov, d := types.ObjectValueFrom(ctx, OrchestratorBlockAttrTypes(), om)
		diags.Append(d...)
		model.Orchestrator = ov
	} else {
		model.Orchestrator = types.ObjectNull(OrchestratorBlockAttrTypes())
	}

	if cloudRaw, ok := body["cloud"].(map[string]any); ok {
		cm := mapToCloudBlockModel(ctx, cloudRaw)
		cv, d := types.ObjectValueFrom(ctx, CloudBlockAttrTypes(), cm)
		diags.Append(d...)
		model.Cloud = cv
	} else {
		model.Cloud = types.ObjectNull(CloudBlockAttrTypes())
	}

	if eventRaw, ok := body["event"].(map[string]any); ok {
		em := mapToEventBlockModel(ctx, eventRaw)
		ev, d := types.ObjectValueFrom(ctx, EventBlockAttrTypes(), em)
		diags.Append(d...)
		model.Event = ev
	} else {
		model.Event = types.ObjectNull(EventBlockAttrTypes())
	}

	if assetRaw, ok := body[attrAsset].(map[string]any); ok {
		am := mapToAssetBlockModel(ctx, assetRaw, diags)
		av, d := types.ObjectValueFrom(ctx, AssetBlockAttrTypes(), am)
		diags.Append(d...)
		model.Asset = av
	} else {
		model.Asset = types.ObjectNull(AssetBlockAttrTypes())
	}
}

func mapToEntityBlockModel(ctx context.Context, m map[string]any, _ *diag.Diagnostics) entityBlockModel {
	model := entityBlockModel{
		ID:      getStringValue(m, "id"),
		Name:    getStringValue(m, attrName),
		Type:    getStringValue(m, attrType),
		SubType: getStringValue(m, "sub_type"),
		Source:  getStringSetValue(m, "source"),
	}
	if attrsRaw, ok := m["attributes"].(map[string]any); ok {
		attr := entityAttributesBlockModel{
			Asset:      getBoolValue(attrsRaw, attrAsset),
			Managed:    getBoolValue(attrsRaw, "managed"),
			Privileged: getBoolValue(attrsRaw, "privileged"),
			MfaEnabled: getBoolValue(attrsRaw, "mfa_enabled"),
		}
		model.Attributes, _ = types.ObjectValueFrom(ctx, AttributesBlockAttrTypes(), attr)
	} else {
		model.Attributes = types.ObjectNull(AttributesBlockAttrTypes())
	}
	if behRaw, ok := m["behaviors"].(map[string]any); ok {
		beh := entityBehaviorsBlockModel{
			BruteForceVictim: getBoolValue(behRaw, "brute_force_victim"),
			NewCountryLogin:  getBoolValue(behRaw, "new_country_login"),
			UsedUsbDevice:    getBoolValue(behRaw, "used_usb_device"),
		}
		model.Behaviors, _ = types.ObjectValueFrom(ctx, BehaviorsBlockAttrTypes(), beh)
	} else {
		model.Behaviors = types.ObjectNull(BehaviorsBlockAttrTypes())
	}
	if lcRaw, ok := m["lifecycle"].(map[string]any); ok {
		lc := entityLifecycleBlockModel{
			FirstSeen:    getStringValue(lcRaw, "first_seen"),
			LastSeen:     getStringValue(lcRaw, "last_seen"),
			LastActivity: getStringValue(lcRaw, "last_activity"),
		}
		model.Lifecycle, _ = types.ObjectValueFrom(ctx, LifecycleBlockAttrTypes(), lc)
	} else {
		model.Lifecycle = types.ObjectNull(LifecycleBlockAttrTypes())
	}
	if riskRaw, ok := m[attrRisk].(map[string]any); ok {
		model.Risk = mapToRiskBlockModel(ctx, riskRaw)
	} else {
		model.Risk = types.ObjectNull(RiskBlockAttrTypes())
	}
	if relRaw, ok := m["relationships"].(map[string]any); ok {
		rel := entityRelationshipsBlockModel{
			OwnedBy:              getStringSetValue(relRaw, "owned_by"),
			Owns:                 getStringSetValue(relRaw, "owns"),
			SupervisedBy:         getStringSetValue(relRaw, "supervised_by"),
			Supervises:           getStringSetValue(relRaw, "supervises"),
			DependsOn:            getStringSetValue(relRaw, "depends_on"),
			DependentOf:          getStringSetValue(relRaw, "dependent_of"),
			CommunicatesWith:     getStringSetValue(relRaw, "communicates_with"),
			AccessesFrequently:   getStringSetValue(relRaw, "accesses_frequently"),
			AccessedFrequentlyBy: getStringSetValue(relRaw, "accessed_frequently_by"),
			AccessesInfrequently: getStringSetValue(relRaw, "accesses_infrequently"),
		}
		model.Relationships, _ = types.ObjectValueFrom(ctx, RelationshipsBlockAttrTypes(), rel)
	} else {
		model.Relationships = types.ObjectNull(RelationshipsBlockAttrTypes())
	}
	return model
}

func mapToRiskBlockModel(ctx context.Context, m map[string]any) types.Object {
	model := entityRiskBlockModel{
		CalculatedLevel:     getStringValue(m, attrCalculatedLevel),
		CalculatedScore:     getFloat64Value(m, attrCalculatedScore),
		CalculatedScoreNorm: getFloat64Value(m, attrCalculatedScoreNorm),
	}
	obj, _ := types.ObjectValueFrom(ctx, RiskBlockAttrTypes(), model)
	return obj
}

func mapToHostBlockModel(ctx context.Context, m map[string]any, _ *diag.Diagnostics) hostBlockModel {
	model := hostBlockModel{
		Name:         getStringValue(m, attrName),
		Domain:       getStringSetValue(m, attrDomain),
		Hostname:     getStringSetValue(m, "hostname"),
		ID:           getStringSetValue(m, "id"),
		IP:           getStringSetValue(m, "ip"),
		Mac:          getStringSetValue(m, "mac"),
		Type:         getStringSetValue(m, attrType),
		Architecture: getStringSetValue(m, "architecture"),
	}
	if osRaw, ok := m["os"].(map[string]any); ok {
		osModel := hostOsBlockModel{
			Family:   getStringValue(osRaw, "family"),
			Full:     getStringValue(osRaw, "full"),
			Kernel:   getStringValue(osRaw, "kernel"),
			Name:     getStringValue(osRaw, attrName),
			Platform: getStringValue(osRaw, "platform"),
			Type:     getStringValue(osRaw, attrType),
			Version:  getStringValue(osRaw, "version"),
		}
		model.Os, _ = types.ObjectValueFrom(ctx, HostOsBlockAttrTypes(), osModel)
	} else {
		model.Os = types.ObjectNull(HostOsBlockAttrTypes())
	}
	if riskRaw, ok := m[attrRisk].(map[string]any); ok {
		model.Risk = mapToRiskBlockModel(ctx, riskRaw)
	} else {
		model.Risk = types.ObjectNull(RiskBlockAttrTypes())
	}
	return model
}

func mapToUserBlockModel(ctx context.Context, m map[string]any, _ *diag.Diagnostics) userBlockModel {
	model := userBlockModel{
		Name:     getStringValue(m, attrName),
		Domain:   getStringSetValue(m, attrDomain),
		Email:    getStringSetValue(m, attrEmail),
		FullName: getStringSetValue(m, "full_name"),
		Hash:     getStringSetValue(m, "hash"),
		ID:       getStringSetValue(m, "id"),
		Roles:    getStringSetValue(m, "roles"),
	}
	if riskRaw, ok := m[attrRisk].(map[string]any); ok {
		model.Risk = mapToRiskBlockModel(ctx, riskRaw)
	} else {
		model.Risk = types.ObjectNull(RiskBlockAttrTypes())
	}
	return model
}

func mapToServiceBlockModel(ctx context.Context, m map[string]any, _ *diag.Diagnostics) serviceBlockModel {
	model := serviceBlockModel{
		Name: getStringValue(m, attrName),
	}
	if riskRaw, ok := m[attrRisk].(map[string]any); ok {
		model.Risk = mapToRiskBlockModel(ctx, riskRaw)
	} else {
		model.Risk = types.ObjectNull(RiskBlockAttrTypes())
	}
	return model
}

func mapToOrchestratorBlockModel(_ context.Context, m map[string]any) orchestratorBlockModel {
	model := orchestratorBlockModel{
		Name:           getStringValue(m, attrName),
		Type:           getStringValue(m, attrType),
		Namespace:      getStringValue(m, "namespace"),
		ClusterID:      getStringValue(m, "cluster_id"),
		ClusterName:    getStringValue(m, "cluster_name"),
		ClusterVersion: getStringValue(m, "cluster_version"),
		ResourceID:     getStringValue(m, "resource_id"),
		ResourceName:   getStringValue(m, "resource_name"),
		ResourceType:   getStringValue(m, "resource_type"),
	}
	return model
}

func mapToCloudBlockModel(_ context.Context, m map[string]any) cloudBlockModel {
	return cloudBlockModel{
		Provider:    getStringValue(m, attrProvider),
		Region:      getStringValue(m, "region"),
		AccountID:   getStringValue(m, "account_id"),
		AccountName: getStringValue(m, "account_name"),
		ProjectID:   getStringValue(m, "project_id"),
		ProjectName: getStringValue(m, "project_name"),
		ServiceName: getStringValue(m, "service_name"),
	}
}

func mapToEventBlockModel(_ context.Context, m map[string]any) eventBlockModel {
	return eventBlockModel{
		Category:  getStringValue(m, "category"),
		Type:      getStringValue(m, attrType),
		Dataset:   getStringValue(m, "dataset"),
		Kind:      getStringValue(m, "kind"),
		Outcome:   getStringValue(m, "outcome"),
		Provider:  getStringValue(m, attrProvider),
		Action:    getStringValue(m, "action"),
		Code:      getStringValue(m, "code"),
		Reference: getStringValue(m, "reference"),
		Reason:    getStringValue(m, attrReason),
		Severity:  getStringValue(m, "severity"),
		Timezone:  getStringValue(m, "timezone"),
		URL:       getStringValue(m, "url"),
		Ingested:  getStringValue(m, "ingested"),
	}
}

func mapToAssetBlockModel(ctx context.Context, m map[string]any, _ *diag.Diagnostics) assetBlockModel {
	model := assetBlockModel{
		Criticality: getStringValue(m, "criticality"),
		Value:       getFloat64Value(m, attrValue),
	}
	if fbRaw, ok := m["criticality_feedback"].(map[string]any); ok {
		fb := assetCriticalityFeedbackBlockModel{
			Notes:  getStringValue(fbRaw, "notes"),
			Reason: getStringValue(fbRaw, attrReason),
		}
		model.CriticalityFeedback, _ = types.ObjectValueFrom(ctx, AssetCriticalityFeedbackBlockAttrTypes(), fb)
	} else {
		model.CriticalityFeedback = types.ObjectNull(AssetCriticalityFeedbackBlockAttrTypes())
	}
	if ownerRaw, ok := m["owner"].(map[string]any); ok {
		owner := assetOwnerBlockModel{
			Name:       getStringValue(ownerRaw, attrName),
			Department: getStringValue(ownerRaw, "department"),
			Email:      getStringValue(ownerRaw, attrEmail),
			Ext:        getStringValue(ownerRaw, "ext"),
		}
		model.Owner, _ = types.ObjectValueFrom(ctx, AssetOwnerBlockAttrTypes(), owner)
	} else {
		model.Owner = types.ObjectNull(AssetOwnerBlockAttrTypes())
	}
	return model
}

// ItemObjectType returns the object type for items in the list data source.
// It covers the fields that apiBodyToModel populates from an API response.
func ItemObjectType() attr.Type {
	return types.ObjectType{AttrTypes: map[string]attr.Type{
		attrTimestamp:    types.StringType,
		attrEntity:       types.ObjectType{AttrTypes: BlockAttrTypes()},
		attrHost:         types.ObjectType{AttrTypes: HostBlockAttrTypes()},
		attrUser:         types.ObjectType{AttrTypes: UserBlockAttrTypes()},
		attrService:      types.ObjectType{AttrTypes: ServiceBlockAttrTypes()},
		attrCloud:        types.ObjectType{AttrTypes: CloudBlockAttrTypes()},
		attrAsset:        types.ObjectType{AttrTypes: AssetBlockAttrTypes()},
		attrOrchestrator: types.ObjectType{AttrTypes: OrchestratorBlockAttrTypes()},
		attrEvent:        types.ObjectType{AttrTypes: EventBlockAttrTypes()},
		attrLabels:       types.MapType{ElemType: types.StringType},
		attrTags:         types.SetType{ElemType: types.StringType},
		attrDocumentJSON: jsontypes.NormalizedType{},
	}}
}

// APIBodyToItem converts a raw entity document from the API list response
// into a types.Object suitable for the items list. It fills the same typed
// attributes that the resource model would have after a read.
func APIBodyToItem(ctx context.Context, body map[string]any, diags *diag.Diagnostics) types.Object {
	var item tfModel
	apiBodyToModel(ctx, body, &item, diags)
	if diags.HasError() {
		return types.ObjectNull(ItemObjectType().(types.ObjectType).AttrTypes)
	}
	obj, d := types.ObjectValue(ItemObjectType().(types.ObjectType).AttrTypes, map[string]attr.Value{
		"@timestamp":     item.Timestamp,
		attrEntity:       item.Entity,
		"host":           item.Host,
		"user":           item.User,
		"service":        item.Service,
		"cloud":          item.Cloud,
		"asset":          item.Asset,
		"orchestrator":   item.Orchestrator,
		"event":          item.Event,
		attrLabels:       item.Labels,
		attrTags:         item.Tags,
		attrDocumentJSON: item.DocumentJSON,
	})
	diags.Append(d...)
	return obj
}

// ItemModel is the struct used by the data source items list.
type ItemModel = map[string]attr.Value

// ItemAttrTypes returns the attribute types for items in the list data source.
func ItemAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		attrTimestamp:    types.StringType,
		attrEntity:       types.ObjectType{AttrTypes: BlockAttrTypes()},
		attrHost:         types.ObjectType{AttrTypes: HostBlockAttrTypes()},
		attrUser:         types.ObjectType{AttrTypes: UserBlockAttrTypes()},
		attrService:      types.ObjectType{AttrTypes: ServiceBlockAttrTypes()},
		attrCloud:        types.ObjectType{AttrTypes: CloudBlockAttrTypes()},
		attrAsset:        types.ObjectType{AttrTypes: AssetBlockAttrTypes()},
		attrOrchestrator: types.ObjectType{AttrTypes: OrchestratorBlockAttrTypes()},
		attrEvent:        types.ObjectType{AttrTypes: EventBlockAttrTypes()},
		attrLabels:       types.MapType{ElemType: types.StringType},
		attrTags:         types.SetType{ElemType: types.StringType},
		attrDocumentJSON: jsontypes.NormalizedType{},
	}
}

// QuoteKQLString escapes and quotes a value for use as a KQL string literal.
// This safely handles entity IDs that may contain quotes or backslashes.
func QuoteKQLString(v string) string {
	return strconv.Quote(v)
}

// injectEntityIDAndMarshal sets entity.id in bodyMap and marshals it to JSON.
func injectEntityIDAndMarshal(bodyMap map[string]any, entityID string) ([]byte, diag.Diagnostics) {
	if entityMap, ok := bodyMap["entity"].(map[string]any); ok {
		entityMap["id"] = entityID
		bodyMap["entity"] = entityMap
	} else {
		bodyMap["entity"] = map[string]any{"id": entityID}
	}

	bodyBytes, err := json.Marshal(bodyMap)
	if err != nil {
		return nil, diag.Diagnostics{
			diag.NewErrorDiagnostic("JSON marshal error", err.Error()),
		}
	}
	return bodyBytes, nil
}

// ExtractEntitiesFromResponse extracts the entity list from an API response map,
// trying "entities" first and falling back to "records" for older API versions.
func ExtractEntitiesFromResponse(result map[string]any) []any {
	if rawEntities, ok := result["entities"].([]any); ok {
		return rawEntities
	}
	if rawRecords, ok := result["records"].([]any); ok {
		return rawRecords
	}
	return nil
}
