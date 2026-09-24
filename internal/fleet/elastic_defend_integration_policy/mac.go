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

package elasticdefendintegrationpolicy

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

func mapMacPolicyFromAPI(ctx context.Context, data map[string]any) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics
	if data == nil {
		return types.ObjectNull(macAttrTypes()), diags
	}

	eventsObj, d := mapOptionalObject(ctx, data, "events", macEventsAttrTypes(), func(m map[string]any) macEventsModel {
		return macEventsModel{
			Process: typeutils.BoolFromMap(m, attrProcess),
			Network: typeutils.BoolFromMap(m, "network"),
			File:    typeutils.BoolFromMap(m, "file"),
		}
	})
	diags.Append(d...)

	malwareObj, d := mapOptionalObject(ctx, data, "malware", malwareFullAttrTypes(), func(m map[string]any) malwareFullModel {
		return malwareFullModel{
			Mode:        typeutils.StringFromMap(m, "mode"),
			Blocklist:   typeutils.BoolFromMap(m, "blocklist"),
			OnWriteScan: typeutils.BoolFromMap(m, attrOnWriteScan),
			NotifyUser:  typeutils.BoolFromMap(m, attrNotifyUser),
		}
	})
	diags.Append(d...)

	common, d := mapCommonPolicyFieldsFromAPI(ctx, data)
	diags.Append(d...)

	popupData := getMap(data, attrPopup)
	popupObj, d := mapMacLinuxPopupFromAPI(ctx, popupData)
	diags.Append(d...)

	macObj, d := types.ObjectValueFrom(ctx, macAttrTypes(), macPolicyModel{
		Events:             eventsObj,
		Malware:            malwareObj,
		MemoryProtection:   common.MemoryProtection,
		BehaviorProtection: common.BehaviorProtection,
		Popup:              popupObj,
		Logging:            common.Logging,
	})
	diags.Append(d...)
	return macObj, diags
}

func buildMacPolicyPayload(ctx context.Context, macObj types.Object) (map[string]any, diag.Diagnostics) {
	var diags diag.Diagnostics
	if macObj.IsNull() || macObj.IsUnknown() {
		return nil, diags
	}

	var mm macPolicyModel
	d := macObj.As(ctx, &mm, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})
	diags.Append(d...)
	if diags.HasError() {
		return nil, diags
	}

	mac := map[string]any{}

	if typeutils.IsKnown(mm.Events) {
		var em macEventsModel
		d = mm.Events.As(ctx, &em, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		events := map[string]any{}
		typeutils.SetBoolInMap(events, attrProcess, em.Process)
		typeutils.SetBoolInMap(events, "network", em.Network)
		typeutils.SetBoolInMap(events, "file", em.File)
		mac["events"] = events
	}

	if typeutils.IsKnown(mm.Malware) {
		var malwareModel malwareFullModel
		d = mm.Malware.As(ctx, &malwareModel, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		malware := map[string]any{}
		typeutils.SetStringInMap(malware, "mode", malwareModel.Mode)
		typeutils.SetBoolInMap(malware, "blocklist", malwareModel.Blocklist)
		typeutils.SetBoolInMap(malware, attrOnWriteScan, malwareModel.OnWriteScan)
		typeutils.SetBoolInMap(malware, attrNotifyUser, malwareModel.NotifyUser)
		mac["malware"] = malware
	}

	if typeutils.IsKnown(mm.MemoryProtection) {
		var pm protectionModeModel
		d = mm.MemoryProtection.As(ctx, &pm, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		memProt := map[string]any{}
		typeutils.SetStringInMap(memProt, "mode", pm.Mode)
		typeutils.SetBoolInMap(memProt, attrSupported, pm.Supported)
		mac["memory_protection"] = memProt
	}

	if typeutils.IsKnown(mm.BehaviorProtection) {
		var bm behaviorProtectionModel
		d = mm.BehaviorProtection.As(ctx, &bm, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		behProt := map[string]any{}
		typeutils.SetStringInMap(behProt, "mode", bm.Mode)
		typeutils.SetBoolInMap(behProt, attrSupported, bm.Supported)
		typeutils.SetBoolInMap(behProt, attrReputationService, bm.ReputationService)
		mac["behavior_protection"] = behProt
	}

	if typeutils.IsKnown(mm.Popup) {
		var pm macLinuxPopupModel
		d = mm.Popup.As(ctx, &pm, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		popup := map[string]any{}
		setPopupItem(ctx, popup, "malware", pm.Malware, &diags)
		setPopupItem(ctx, popup, "memory_protection", pm.MemoryProtection, &diags)
		setPopupItem(ctx, popup, "behavior_protection", pm.BehaviorProtection, &diags)
		mac[attrPopup] = popup
	}

	if typeutils.IsKnown(mm.Logging) {
		var lm loggingModel
		d = mm.Logging.As(ctx, &lm, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		logging := map[string]any{}
		typeutils.SetStringInMap(logging, "file", lm.File)
		mac["logging"] = logging
	}

	return mac, diags
}

func macEventsAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		attrProcess: types.BoolType,
		attrNetwork: types.BoolType,
		attrFile:    types.BoolType,
	}
}

func macAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		attrEvents:             types.ObjectType{AttrTypes: macEventsAttrTypes()},
		attrMalware:            types.ObjectType{AttrTypes: malwareFullAttrTypes()},
		attrMemoryProtection:   types.ObjectType{AttrTypes: protectionModeAttrTypes()},
		attrBehaviorProtection: types.ObjectType{AttrTypes: behaviorProtectionAttrTypes()},
		attrPopup:              types.ObjectType{AttrTypes: macLinuxPopupAttrTypes()},
		attrLogging:            types.ObjectType{AttrTypes: loggingAttrTypes()},
	}
}
