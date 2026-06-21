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
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

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
		ID:      typeutils.StringFromMap(m, "id"),
		Name:    typeutils.StringFromMap(m, attrName),
		Type:    typeutils.StringFromMap(m, attrType),
		SubType: typeutils.StringFromMap(m, "sub_type"),
		Source:  getStringSetValue(m, "source"),
	}
	if attrsRaw, ok := m["attributes"].(map[string]any); ok {
		attr := entityAttributesBlockModel{
			Asset:      typeutils.BoolFromMap(attrsRaw, attrAsset),
			Managed:    typeutils.BoolFromMap(attrsRaw, "managed"),
			Privileged: typeutils.BoolFromMap(attrsRaw, "privileged"),
			MfaEnabled: typeutils.BoolFromMap(attrsRaw, "mfa_enabled"),
		}
		model.Attributes, _ = types.ObjectValueFrom(ctx, AttributesBlockAttrTypes(), attr)
	} else {
		model.Attributes = types.ObjectNull(AttributesBlockAttrTypes())
	}
	if behRaw, ok := m["behaviors"].(map[string]any); ok {
		beh := entityBehaviorsBlockModel{
			BruteForceVictim: typeutils.BoolFromMap(behRaw, "brute_force_victim"),
			NewCountryLogin:  typeutils.BoolFromMap(behRaw, "new_country_login"),
			UsedUsbDevice:    typeutils.BoolFromMap(behRaw, "used_usb_device"),
		}
		model.Behaviors, _ = types.ObjectValueFrom(ctx, BehaviorsBlockAttrTypes(), beh)
	} else {
		model.Behaviors = types.ObjectNull(BehaviorsBlockAttrTypes())
	}
	if lcRaw, ok := m["lifecycle"].(map[string]any); ok {
		lc := entityLifecycleBlockModel{
			FirstSeen:    typeutils.StringFromMap(lcRaw, "first_seen"),
			LastSeen:     typeutils.StringFromMap(lcRaw, "last_seen"),
			LastActivity: typeutils.StringFromMap(lcRaw, "last_activity"),
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
		CalculatedLevel:     typeutils.StringFromMap(m, attrCalculatedLevel),
		CalculatedScore:     typeutils.Float64FromMap(m, attrCalculatedScore),
		CalculatedScoreNorm: typeutils.Float64FromMap(m, attrCalculatedScoreNorm),
	}
	obj, _ := types.ObjectValueFrom(ctx, RiskBlockAttrTypes(), model)
	return obj
}

func mapToHostBlockModel(ctx context.Context, m map[string]any, _ *diag.Diagnostics) hostBlockModel {
	model := hostBlockModel{
		Name:         typeutils.StringFromMap(m, attrName),
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
			Family:   typeutils.StringFromMap(osRaw, "family"),
			Full:     typeutils.StringFromMap(osRaw, "full"),
			Kernel:   typeutils.StringFromMap(osRaw, "kernel"),
			Name:     typeutils.StringFromMap(osRaw, attrName),
			Platform: typeutils.StringFromMap(osRaw, "platform"),
			Type:     typeutils.StringFromMap(osRaw, attrType),
			Version:  typeutils.StringFromMap(osRaw, "version"),
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
		Name:     typeutils.StringFromMap(m, attrName),
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
		Name: typeutils.StringFromMap(m, attrName),
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
		Name:           typeutils.StringFromMap(m, attrName),
		Type:           typeutils.StringFromMap(m, attrType),
		Namespace:      typeutils.StringFromMap(m, "namespace"),
		ClusterID:      typeutils.StringFromMap(m, "cluster_id"),
		ClusterName:    typeutils.StringFromMap(m, "cluster_name"),
		ClusterVersion: typeutils.StringFromMap(m, "cluster_version"),
		ResourceID:     typeutils.StringFromMap(m, "resource_id"),
		ResourceName:   typeutils.StringFromMap(m, "resource_name"),
		ResourceType:   typeutils.StringFromMap(m, "resource_type"),
	}
	return model
}

func mapToCloudBlockModel(_ context.Context, m map[string]any) cloudBlockModel {
	return cloudBlockModel{
		Provider:    typeutils.StringFromMap(m, attrProvider),
		Region:      typeutils.StringFromMap(m, "region"),
		AccountID:   typeutils.StringFromMap(m, "account_id"),
		AccountName: typeutils.StringFromMap(m, "account_name"),
		ProjectID:   typeutils.StringFromMap(m, "project_id"),
		ProjectName: typeutils.StringFromMap(m, "project_name"),
		ServiceName: typeutils.StringFromMap(m, "service_name"),
	}
}

func mapToEventBlockModel(_ context.Context, m map[string]any) eventBlockModel {
	return eventBlockModel{
		Category:  typeutils.StringFromMap(m, "category"),
		Type:      typeutils.StringFromMap(m, attrType),
		Dataset:   typeutils.StringFromMap(m, "dataset"),
		Kind:      typeutils.StringFromMap(m, "kind"),
		Outcome:   typeutils.StringFromMap(m, "outcome"),
		Provider:  typeutils.StringFromMap(m, attrProvider),
		Action:    typeutils.StringFromMap(m, "action"),
		Code:      typeutils.StringFromMap(m, "code"),
		Reference: typeutils.StringFromMap(m, "reference"),
		Reason:    typeutils.StringFromMap(m, attrReason),
		Severity:  typeutils.StringFromMap(m, "severity"),
		Timezone:  typeutils.StringFromMap(m, "timezone"),
		URL:       typeutils.StringFromMap(m, "url"),
		Ingested:  typeutils.StringFromMap(m, "ingested"),
	}
}

func mapToAssetBlockModel(ctx context.Context, m map[string]any, _ *diag.Diagnostics) assetBlockModel {
	model := assetBlockModel{
		Criticality: typeutils.StringFromMap(m, "criticality"),
		Value:       typeutils.Float64FromMap(m, attrValue),
	}
	if fbRaw, ok := m["criticality_feedback"].(map[string]any); ok {
		fb := assetCriticalityFeedbackBlockModel{
			Notes:  typeutils.StringFromMap(fbRaw, "notes"),
			Reason: typeutils.StringFromMap(fbRaw, attrReason),
		}
		model.CriticalityFeedback, _ = types.ObjectValueFrom(ctx, AssetCriticalityFeedbackBlockAttrTypes(), fb)
	} else {
		model.CriticalityFeedback = types.ObjectNull(AssetCriticalityFeedbackBlockAttrTypes())
	}
	if ownerRaw, ok := m["owner"].(map[string]any); ok {
		owner := assetOwnerBlockModel{
			Name:       typeutils.StringFromMap(ownerRaw, attrName),
			Department: typeutils.StringFromMap(ownerRaw, "department"),
			Email:      typeutils.StringFromMap(ownerRaw, attrEmail),
			Ext:        typeutils.StringFromMap(ownerRaw, "ext"),
		}
		model.Owner, _ = types.ObjectValueFrom(ctx, AssetOwnerBlockAttrTypes(), owner)
	} else {
		model.Owner = types.ObjectNull(AssetOwnerBlockAttrTypes())
	}
	return model
}
