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

	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	jsontypes "github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// setBlockOrJSON writes key into body from a typed block, falling back to a
// raw JSON value when the block is null or unknown.
func setBlockOrJSON(
	ctx context.Context,
	body map[string]any,
	key string,
	block types.Object,
	blockToMap func(context.Context, types.Object, *diag.Diagnostics) map[string]any,
	jsonVal jsontypes.Normalized,
	jsonPath path.Path,
	diags *diag.Diagnostics,
) {
	if !block.IsNull() && !block.IsUnknown() {
		if m := blockToMap(ctx, block, diags); m != nil {
			body[key] = m
		}
		return
	}
	if !jsonVal.IsNull() && !jsonVal.IsUnknown() {
		if m := typeutils.NormalizedTypeToMap[any](jsonVal, jsonPath, diags); m != nil {
			body[key] = m
		}
	}
}

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

	setBlockOrJSON(ctx, body, attrEntity, model.Entity, entityBlockToMap, model.EntityJSON, path.Root("entity_json"), &diags)
	setBlockOrJSON(ctx, body, attrHost, model.Host, hostBlockToMap, model.HostJSON, path.Root("host_json"), &diags)
	setBlockOrJSON(ctx, body, attrUser, model.User, userBlockToMap, model.UserJSON, path.Root("user_json"), &diags)
	setBlockOrJSON(ctx, body, attrService, model.Service, serviceBlockToMap, model.ServiceJSON, path.Root("service_json"), &diags)
	setBlockOrJSON(ctx, body, attrOrchestrator, model.Orchestrator, orchestratorBlockToMap, model.OrchestratorJSON, path.Root("orchestrator_json"), &diags)
	setBlockOrJSON(ctx, body, attrCloud, model.Cloud, cloudBlockToMap, model.CloudJSON, path.Root("cloud_json"), &diags)
	setBlockOrJSON(ctx, body, attrEvent, model.Event, eventBlockToMap, model.EventJSON, path.Root("event_json"), &diags)
	setBlockOrJSON(ctx, body, attrAsset, model.Asset, assetBlockToMap, model.AssetJSON, path.Root("asset_json"), &diags)

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
