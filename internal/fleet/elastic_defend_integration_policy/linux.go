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

	eventsObj, d := mapOptionalObject(ctx, data, attrEvents, linuxEventsAttrTypes(), func(m map[string]any) linuxEventsModel {
		return linuxEventsModel{
			Process:     typeutils.BoolFromMap(m, attrProcess),
			Network:     typeutils.BoolFromMap(m, attrNetwork),
			File:        typeutils.BoolFromMap(m, attrFile),
			SessionData: typeutils.BoolFromMap(m, attrSessionData),
			TtyIO:       typeutils.BoolFromMap(m, attrTtyIO),
			DNS:         typeutils.BoolFromMap(m, attrDNS),
		}
	})
	diags.Append(d...)

	malwareObj, d := mapOptionalObject(ctx, data, attrMalware, malwareLinuxAttrTypes(), func(m map[string]any) malwareLinuxModel {
		return malwareLinuxModel{
			Mode:        typeutils.StringFromMap(m, attrMode),
			Blocklist:   typeutils.BoolFromMap(m, attrBlocklist),
			OnWriteScan: typeutils.BoolFromMap(m, attrOnWriteScan),
		}
	})
	diags.Append(d...)

	common, d := mapCommonPolicyFieldsFromAPI(ctx, data)
	diags.Append(d...)

	popupData := getMap(data, attrPopup)
	popupObj, d := mapLinuxPopupFromAPI(ctx, popupData)
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

	if em, ok := decodeObjectField[linuxEventsModel](ctx, lm.Events, &diags); ok {
		events := map[string]any{}
		typeutils.SetBoolInMap(events, attrProcess, em.Process)
		typeutils.SetBoolInMap(events, attrNetwork, em.Network)
		typeutils.SetBoolInMap(events, attrFile, em.File)
		typeutils.SetBoolInMap(events, attrSessionData, em.SessionData)
		typeutils.SetBoolInMap(events, attrTtyIO, em.TtyIO)
		typeutils.SetBoolInMap(events, attrDNS, em.DNS)
		linux[attrEvents] = events
	}

	if mm, ok := decodeObjectField[malwareLinuxModel](ctx, lm.Malware, &diags); ok {
		malware := map[string]any{}
		typeutils.SetStringInMap(malware, attrMode, mm.Mode)
		typeutils.SetBoolInMap(malware, attrBlocklist, mm.Blocklist)
		typeutils.SetBoolInMap(malware, attrOnWriteScan, mm.OnWriteScan)
		linux[attrMalware] = malware
	}

	buildCommonPolicyPayloadFields(ctx, linux, commonPolicyPayloadFields{
		MemoryProtection:   lm.MemoryProtection,
		BehaviorProtection: lm.BehaviorProtection,
		Logging:            lm.Logging,
	}, &diags)

	if pm, ok := decodeObjectField[linuxPopupModel](ctx, lm.Popup, &diags); ok {
		popup := map[string]any{}
		setPopupItem(ctx, popup, attrMalware, pm.Malware, &diags)
		setPopupItem(ctx, popup, attrMemoryProtection, pm.MemoryProtection, &diags)
		setPopupItem(ctx, popup, attrBehaviorProtection, pm.BehaviorProtection, &diags)
		linux[attrPopup] = popup
	}

	if diags.HasError() {
		return nil, diags
	}
	return linux, diags
}

func linuxEventsAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		attrProcess:     types.BoolType,
		attrNetwork:     types.BoolType,
		attrFile:        types.BoolType,
		attrSessionData: types.BoolType,
		attrTtyIO:       types.BoolType,
		attrDNS:         types.BoolType,
	}
}

func malwareLinuxAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		attrMode:        types.StringType,
		attrBlocklist:   types.BoolType,
		attrOnWriteScan: types.BoolType,
	}
}

func linuxAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		attrEvents:             types.ObjectType{AttrTypes: linuxEventsAttrTypes()},
		attrMalware:            types.ObjectType{AttrTypes: malwareLinuxAttrTypes()},
		attrMemoryProtection:   types.ObjectType{AttrTypes: memoryProtectionAttrTypes()},
		attrBehaviorProtection: types.ObjectType{AttrTypes: behaviorProtectionAttrTypes()},
		attrPopup:              types.ObjectType{AttrTypes: linuxPopupAttrTypes()},
		attrLogging:            types.ObjectType{AttrTypes: loggingAttrTypes()},
	}
}
