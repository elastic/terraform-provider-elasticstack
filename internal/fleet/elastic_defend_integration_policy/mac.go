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

	eventsObj, d := mapOptionalObject(ctx, data, attrEvents, macEventsAttrTypes(), func(m map[string]any) macEventsModel {
		return macEventsModel{
			Process:  typeutils.BoolFromMap(m, attrProcess),
			Network:  typeutils.BoolFromMap(m, attrNetwork),
			File:     typeutils.BoolFromMap(m, attrFile),
			DNS:      typeutils.BoolFromMap(m, attrDNS),
			Security: typeutils.BoolFromMap(m, attrSecurity),
		}
	})
	diags.Append(d...)

	malwareObj, d := mapOptionalObject(ctx, data, attrMalware, malwareFullAttrTypes(), func(m map[string]any) malwareFullModel {
		return malwareFullModel{
			Mode:        typeutils.StringFromMap(m, attrMode),
			Blocklist:   typeutils.BoolFromMap(m, attrBlocklist),
			OnWriteScan: typeutils.BoolFromMap(m, attrOnWriteScan),
			NotifyUser:  typeutils.BoolFromMap(m, attrNotifyUser),
		}
	})
	diags.Append(d...)

	ransomwareObj, d := mapOptionalObject(ctx, data, attrRansomware, protectionModeAttrTypes(), func(m map[string]any) protectionModeModel {
		return protectionModeModel{
			Mode:      typeutils.StringFromMap(m, attrMode),
			Supported: typeutils.BoolFromMap(m, attrSupported),
		}
	})
	diags.Append(d...)

	common, d := mapCommonPolicyFieldsFromAPI(ctx, data)
	diags.Append(d...)

	deviceControlObj, d := mapOptionalObject(ctx, data, attrDeviceControl, deviceControlAttrTypes(), func(m map[string]any) deviceControlModel {
		return deviceControlModel{
			Enabled:    typeutils.BoolFromMap(m, attrEnabled),
			UsbStorage: typeutils.StringFromMap(m, attrUsbStorage),
		}
	})
	diags.Append(d...)

	popupData := getMap(data, attrPopup)
	popupObj, d := mapMacPopupFromAPI(ctx, popupData)
	diags.Append(d...)

	macObj, d := types.ObjectValueFrom(ctx, macAttrTypes(), macPolicyModel{
		Events:             eventsObj,
		Malware:            malwareObj,
		Ransomware:         ransomwareObj,
		MemoryProtection:   common.MemoryProtection,
		BehaviorProtection: common.BehaviorProtection,
		DeviceControl:      deviceControlObj,
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

	if em, ok := decodeObjectField[macEventsModel](ctx, mm.Events, &diags); ok {
		events := map[string]any{}
		typeutils.SetBoolInMap(events, attrProcess, em.Process)
		typeutils.SetBoolInMap(events, attrNetwork, em.Network)
		typeutils.SetBoolInMap(events, attrFile, em.File)
		typeutils.SetBoolInMap(events, attrDNS, em.DNS)
		typeutils.SetBoolInMap(events, attrSecurity, em.Security)
		mac[attrEvents] = events
	}

	if malwareModel, ok := decodeObjectField[malwareFullModel](ctx, mm.Malware, &diags); ok {
		malware := map[string]any{}
		typeutils.SetStringInMap(malware, attrMode, malwareModel.Mode)
		typeutils.SetBoolInMap(malware, attrBlocklist, malwareModel.Blocklist)
		typeutils.SetBoolInMap(malware, attrOnWriteScan, malwareModel.OnWriteScan)
		typeutils.SetBoolInMap(malware, attrNotifyUser, malwareModel.NotifyUser)
		mac[attrMalware] = malware
	}

	if rm, ok := decodeObjectField[protectionModeModel](ctx, mm.Ransomware, &diags); ok {
		ransomware := map[string]any{}
		typeutils.SetStringInMap(ransomware, attrMode, rm.Mode)
		typeutils.SetBoolInMap(ransomware, attrSupported, rm.Supported)
		mac[attrRansomware] = ransomware
	}

	buildCommonPolicyPayloadFields(ctx, mac, commonPolicyPayloadFields{
		MemoryProtection:   mm.MemoryProtection,
		BehaviorProtection: mm.BehaviorProtection,
		Logging:            mm.Logging,
	}, &diags)

	buildDeviceControlPayloadField(ctx, mac, mm.DeviceControl, &diags)

	if pm, ok := decodeObjectField[macPopupModel](ctx, mm.Popup, &diags); ok {
		popup := map[string]any{}
		setPopupItem(ctx, popup, attrMalware, pm.Malware, &diags)
		setPopupItem(ctx, popup, attrRansomware, pm.Ransomware, &diags)
		setPopupItem(ctx, popup, attrMemoryProtection, pm.MemoryProtection, &diags)
		setPopupItem(ctx, popup, attrBehaviorProtection, pm.BehaviorProtection, &diags)
		setPopupItem(ctx, popup, attrDeviceControl, pm.DeviceControl, &diags)
		mac[attrPopup] = popup
	}

	if diags.HasError() {
		return nil, diags
	}
	return mac, diags
}

func macEventsAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		attrProcess:  types.BoolType,
		attrNetwork:  types.BoolType,
		attrFile:     types.BoolType,
		attrDNS:      types.BoolType,
		attrSecurity: types.BoolType,
	}
}

func macAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		attrEvents:             types.ObjectType{AttrTypes: macEventsAttrTypes()},
		attrMalware:            types.ObjectType{AttrTypes: malwareFullAttrTypes()},
		attrRansomware:         types.ObjectType{AttrTypes: protectionModeAttrTypes()},
		attrMemoryProtection:   types.ObjectType{AttrTypes: memoryProtectionAttrTypes()},
		attrBehaviorProtection: types.ObjectType{AttrTypes: behaviorProtectionAttrTypes()},
		attrDeviceControl:      types.ObjectType{AttrTypes: deviceControlAttrTypes()},
		attrPopup:              types.ObjectType{AttrTypes: macPopupAttrTypes()},
		attrLogging:            types.ObjectType{AttrTypes: loggingAttrTypes()},
	}
}
