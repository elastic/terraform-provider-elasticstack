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

func mapWindowsPolicyFromAPI(ctx context.Context, data map[string]any) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics
	if data == nil {
		return types.ObjectNull(windowsAttrTypes()), diags
	}

	eventsObj, d := mapOptionalObject(ctx, data, attrEvents, windowsEventsAttrTypes(), func(m map[string]any) windowsEventsModel {
		return windowsEventsModel{
			Process:          typeutils.BoolFromMap(m, attrProcess),
			Network:          typeutils.BoolFromMap(m, attrNetwork),
			File:             typeutils.BoolFromMap(m, attrFile),
			DllAndDriverLoad: typeutils.BoolFromMap(m, attrDllAndDriverLoad),
			DNS:              typeutils.BoolFromMap(m, attrDNS),
			Registry:         typeutils.BoolFromMap(m, attrRegistry),
			Security:         typeutils.BoolFromMap(m, attrSecurity),
			Authentication:   typeutils.BoolFromMap(m, attrAuthentication),
			CredentialAccess: typeutils.BoolFromMap(m, attrCredentialAccess),
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
	popupObj, d := mapWindowsPopupFromAPI(ctx, popupData)
	diags.Append(d...)

	avrObj, d := mapOptionalObject(ctx, data, attrAntivirusRegistration, antivirusRegistrationAttrTypes(), func(m map[string]any) antivirusRegistrationModel {
		return antivirusRegistrationModel{
			Mode:    typeutils.StringFromMap(m, attrMode),
			Enabled: typeutils.BoolFromMap(m, attrEnabled),
		}
	})
	diags.Append(d...)

	// attack_surface_reduction contains a nested credential_hardening object,
	// so it requires two levels of mapOptionalObject.
	asrData := getMap(data, attrAttackSurfaceReduction)
	var asrObj types.Object
	if asrData != nil {
		chObj, d := mapOptionalObject(ctx, asrData, attrCredentialHardening, credentialHardeningAttrTypes(), func(m map[string]any) credentialHardeningModel {
			return credentialHardeningModel{
				Enabled: typeutils.BoolFromMap(m, attrEnabled),
			}
		})
		diags.Append(d...)
		asrObj, d = types.ObjectValueFrom(ctx, attackSurfaceReductionAttrTypes(), attackSurfaceReductionModel{
			CredentialHardening: chObj,
		})
		diags.Append(d...)
	} else {
		asrObj = types.ObjectNull(attackSurfaceReductionAttrTypes())
	}

	winObj, d := types.ObjectValueFrom(ctx, windowsAttrTypes(), windowsPolicyModel{
		Events:                 eventsObj,
		Malware:                malwareObj,
		Ransomware:             ransomwareObj,
		MemoryProtection:       common.MemoryProtection,
		BehaviorProtection:     common.BehaviorProtection,
		DeviceControl:          deviceControlObj,
		Popup:                  popupObj,
		Logging:                common.Logging,
		AntivirusRegistration:  avrObj,
		AttackSurfaceReduction: asrObj,
	})
	diags.Append(d...)
	return winObj, diags
}

func buildWindowsPolicyPayload(ctx context.Context, winObj types.Object) (map[string]any, diag.Diagnostics) {
	var diags diag.Diagnostics
	if winObj.IsNull() || winObj.IsUnknown() {
		return nil, diags
	}

	var wm windowsPolicyModel
	d := winObj.As(ctx, &wm, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})
	diags.Append(d...)
	if diags.HasError() {
		return nil, diags
	}

	win := map[string]any{}

	if em, ok := decodeObjectField[windowsEventsModel](ctx, wm.Events, &diags); ok {
		events := map[string]any{}
		typeutils.SetBoolInMap(events, attrProcess, em.Process)
		typeutils.SetBoolInMap(events, attrNetwork, em.Network)
		typeutils.SetBoolInMap(events, attrFile, em.File)
		typeutils.SetBoolInMap(events, attrDllAndDriverLoad, em.DllAndDriverLoad)
		typeutils.SetBoolInMap(events, attrDNS, em.DNS)
		typeutils.SetBoolInMap(events, attrRegistry, em.Registry)
		typeutils.SetBoolInMap(events, attrSecurity, em.Security)
		typeutils.SetBoolInMap(events, attrAuthentication, em.Authentication)
		typeutils.SetBoolInMap(events, attrCredentialAccess, em.CredentialAccess)
		win[attrEvents] = events
	}

	if mm, ok := decodeObjectField[malwareFullModel](ctx, wm.Malware, &diags); ok {
		malware := map[string]any{}
		typeutils.SetStringInMap(malware, attrMode, mm.Mode)
		typeutils.SetBoolInMap(malware, attrBlocklist, mm.Blocklist)
		typeutils.SetBoolInMap(malware, attrOnWriteScan, mm.OnWriteScan)
		typeutils.SetBoolInMap(malware, attrNotifyUser, mm.NotifyUser)
		win[attrMalware] = malware
	}

	if rm, ok := decodeObjectField[protectionModeModel](ctx, wm.Ransomware, &diags); ok {
		ransomware := map[string]any{}
		typeutils.SetStringInMap(ransomware, attrMode, rm.Mode)
		typeutils.SetBoolInMap(ransomware, attrSupported, rm.Supported)
		win[attrRansomware] = ransomware
	}

	buildCommonPolicyPayloadFields(ctx, win, commonPolicyPayloadFields{
		MemoryProtection:   wm.MemoryProtection,
		BehaviorProtection: wm.BehaviorProtection,
		Logging:            wm.Logging,
	}, &diags)

	buildDeviceControlPayloadField(ctx, win, wm.DeviceControl, &diags)

	if pm, ok := decodeObjectField[windowsPopupModel](ctx, wm.Popup, &diags); ok {
		popup := map[string]any{}
		setPopupItem(ctx, popup, attrMalware, pm.Malware, &diags)
		setPopupItem(ctx, popup, attrRansomware, pm.Ransomware, &diags)
		setPopupItem(ctx, popup, attrMemoryProtection, pm.MemoryProtection, &diags)
		setPopupItem(ctx, popup, attrBehaviorProtection, pm.BehaviorProtection, &diags)
		setPopupItem(ctx, popup, attrDeviceControl, pm.DeviceControl, &diags)
		win[attrPopup] = popup
	}

	if am, ok := decodeObjectField[antivirusRegistrationModel](ctx, wm.AntivirusRegistration, &diags); ok {
		avr := map[string]any{}
		typeutils.SetStringInMap(avr, attrMode, am.Mode)
		typeutils.SetBoolInMap(avr, attrEnabled, am.Enabled)
		win[attrAntivirusRegistration] = avr
	}

	if am, ok := decodeObjectField[attackSurfaceReductionModel](ctx, wm.AttackSurfaceReduction, &diags); ok {
		asr := map[string]any{}
		if cm, ok := decodeObjectField[credentialHardeningModel](ctx, am.CredentialHardening, &diags); ok {
			ch := map[string]any{}
			typeutils.SetBoolInMap(ch, attrEnabled, cm.Enabled)
			asr[attrCredentialHardening] = ch
		}
		win[attrAttackSurfaceReduction] = asr
	}

	if diags.HasError() {
		return nil, diags
	}
	return win, diags
}

func windowsEventsAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		attrProcess:          types.BoolType,
		attrNetwork:          types.BoolType,
		attrFile:             types.BoolType,
		attrDllAndDriverLoad: types.BoolType,
		attrDNS:              types.BoolType,
		attrRegistry:         types.BoolType,
		attrSecurity:         types.BoolType,
		attrAuthentication:   types.BoolType,
		attrCredentialAccess: types.BoolType,
	}
}

func antivirusRegistrationAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		attrMode:    types.StringType,
		attrEnabled: types.BoolType,
	}
}

func credentialHardeningAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		attrEnabled: types.BoolType,
	}
}

func attackSurfaceReductionAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		attrCredentialHardening: types.ObjectType{AttrTypes: credentialHardeningAttrTypes()},
	}
}

func windowsAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		attrEvents:                 types.ObjectType{AttrTypes: windowsEventsAttrTypes()},
		attrMalware:                types.ObjectType{AttrTypes: malwareFullAttrTypes()},
		attrRansomware:             types.ObjectType{AttrTypes: protectionModeAttrTypes()},
		attrMemoryProtection:       types.ObjectType{AttrTypes: memoryProtectionAttrTypes()},
		attrDeviceControl:          types.ObjectType{AttrTypes: deviceControlAttrTypes()},
		attrBehaviorProtection:     types.ObjectType{AttrTypes: behaviorProtectionAttrTypes()},
		attrPopup:                  types.ObjectType{AttrTypes: windowsPopupAttrTypes()},
		attrLogging:                types.ObjectType{AttrTypes: loggingAttrTypes()},
		attrAntivirusRegistration:  types.ObjectType{AttrTypes: antivirusRegistrationAttrTypes()},
		attrAttackSurfaceReduction: types.ObjectType{AttrTypes: attackSurfaceReductionAttrTypes()},
	}
}
