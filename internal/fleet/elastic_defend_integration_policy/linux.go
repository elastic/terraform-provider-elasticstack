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

func mapLinuxPolicyFromAPI(ctx context.Context, data map[string]any) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics
	if data == nil {
		return types.ObjectNull(linuxAttrTypes()), diags
	}

	eventsObj, d := mapOptionalObject(ctx, data, "events", linuxEventsAttrTypes(), func(m map[string]any) linuxEventsModel {
		return linuxEventsModel{
			Process:     typeutils.BoolFromMap(m, attrProcess),
			Network:     typeutils.BoolFromMap(m, "network"),
			File:        typeutils.BoolFromMap(m, "file"),
			SessionData: typeutils.BoolFromMap(m, "session_data"),
			TtyIO:       typeutils.BoolFromMap(m, "tty_io"),
		}
	})
	diags.Append(d...)

	malwareObj, d := mapOptionalObject(ctx, data, "malware", malwareLinuxAttrTypes(), func(m map[string]any) malwareLinuxModel {
		return malwareLinuxModel{
			Mode:      typeutils.StringFromMap(m, "mode"),
			Blocklist: typeutils.BoolFromMap(m, "blocklist"),
		}
	})
	diags.Append(d...)

	common, d := mapCommonPolicyFieldsFromAPI(ctx, data)
	diags.Append(d...)

	popupData := getMap(data, attrPopup)
	popupObj, d := mapMacLinuxPopupFromAPI(ctx, popupData)
	diags.Append(d...)

	linuxObj, d := types.ObjectValueFrom(ctx, linuxAttrTypes(), linuxPolicyModel{
		Events:             eventsObj,
		Malware:            malwareObj,
		MemoryProtection:   common.MemoryProtection,
		BehaviorProtection: common.BehaviorProtection,
		Popup:              popupObj,
		Logging:            common.Logging,
	})
	diags.Append(d...)
	return linuxObj, diags
}

func buildLinuxPolicyPayload(ctx context.Context, linuxObj types.Object) (map[string]any, diag.Diagnostics) {
	var diags diag.Diagnostics
	if linuxObj.IsNull() || linuxObj.IsUnknown() {
		return nil, diags
	}

	var lm linuxPolicyModel
	d := linuxObj.As(ctx, &lm, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})
	diags.Append(d...)
	if diags.HasError() {
		return nil, diags
	}

	linux := map[string]any{}

	if typeutils.IsKnown(lm.Events) {
		var em linuxEventsModel
		d = lm.Events.As(ctx, &em, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		events := map[string]any{}
		typeutils.SetBoolInMap(events, attrProcess, em.Process)
		typeutils.SetBoolInMap(events, "network", em.Network)
		typeutils.SetBoolInMap(events, "file", em.File)
		typeutils.SetBoolInMap(events, "session_data", em.SessionData)
		typeutils.SetBoolInMap(events, "tty_io", em.TtyIO)
		linux["events"] = events
	}

	if typeutils.IsKnown(lm.Malware) {
		var mm malwareLinuxModel
		d = lm.Malware.As(ctx, &mm, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		malware := map[string]any{}
		typeutils.SetStringInMap(malware, "mode", mm.Mode)
		typeutils.SetBoolInMap(malware, "blocklist", mm.Blocklist)
		linux["malware"] = malware
	}

	if typeutils.IsKnown(lm.MemoryProtection) {
		var pm protectionModeModel
		d = lm.MemoryProtection.As(ctx, &pm, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		memProt := map[string]any{}
		typeutils.SetStringInMap(memProt, "mode", pm.Mode)
		typeutils.SetBoolInMap(memProt, attrSupported, pm.Supported)
		linux["memory_protection"] = memProt
	}

	if typeutils.IsKnown(lm.BehaviorProtection) {
		var bm behaviorProtectionModel
		d = lm.BehaviorProtection.As(ctx, &bm, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		behProt := map[string]any{}
		typeutils.SetStringInMap(behProt, "mode", bm.Mode)
		typeutils.SetBoolInMap(behProt, attrSupported, bm.Supported)
		typeutils.SetBoolInMap(behProt, attrReputationService, bm.ReputationService)
		linux["behavior_protection"] = behProt
	}

	if typeutils.IsKnown(lm.Popup) {
		var pm macLinuxPopupModel
		d = lm.Popup.As(ctx, &pm, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		popup := map[string]any{}
		setPopupItem(ctx, popup, "malware", pm.Malware, &diags)
		setPopupItem(ctx, popup, "memory_protection", pm.MemoryProtection, &diags)
		setPopupItem(ctx, popup, "behavior_protection", pm.BehaviorProtection, &diags)
		linux[attrPopup] = popup
	}

	if typeutils.IsKnown(lm.Logging) {
		var logm loggingModel
		d = lm.Logging.As(ctx, &logm, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		logging := map[string]any{}
		typeutils.SetStringInMap(logging, "file", logm.File)
		linux["logging"] = logging
	}

	return linux, diags
}

func linuxEventsAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		attrProcess:    types.BoolType,
		attrNetwork:    types.BoolType,
		attrFile:       types.BoolType,
		"session_data": types.BoolType,
		"tty_io":       types.BoolType,
	}
}

func malwareLinuxAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		attrMode:      types.StringType,
		attrBlocklist: types.BoolType,
	}
}

func linuxAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"events":              types.ObjectType{AttrTypes: linuxEventsAttrTypes()},
		"malware":             types.ObjectType{AttrTypes: malwareLinuxAttrTypes()},
		"memory_protection":   types.ObjectType{AttrTypes: protectionModeAttrTypes()},
		"behavior_protection": types.ObjectType{AttrTypes: behaviorProtectionAttrTypes()},
		attrPopup:             types.ObjectType{AttrTypes: macLinuxPopupAttrTypes()},
		"logging":             types.ObjectType{AttrTypes: loggingAttrTypes()},
	}
}
