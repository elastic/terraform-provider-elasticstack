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

	eventsObj, d := mapOptionalObject(ctx, data, "events", windowsEventsAttrTypes(), func(m map[string]any) windowsEventsModel {
		return windowsEventsModel{
			Process:          typeutils.BoolFromMap(m, attrProcess),
			Network:          typeutils.BoolFromMap(m, "network"),
			File:             typeutils.BoolFromMap(m, "file"),
			DllAndDriverLoad: typeutils.BoolFromMap(m, "dll_and_driver_load"),
			DNS:              typeutils.BoolFromMap(m, "dns"),
			Registry:         typeutils.BoolFromMap(m, "registry"),
			Security:         typeutils.BoolFromMap(m, "security"),
			Authentication:   typeutils.BoolFromMap(m, "authentication"),
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

	ransomwareObj, d := mapOptionalObject(ctx, data, attrRansomware, protectionModeAttrTypes(), func(m map[string]any) protectionModeModel {
		return protectionModeModel{
			Mode:      typeutils.StringFromMap(m, "mode"),
			Supported: typeutils.BoolFromMap(m, attrSupported),
		}
	})
	diags.Append(d...)

	common, d := mapCommonPolicyFieldsFromAPI(ctx, data)
	diags.Append(d...)

	popupData := getMap(data, attrPopup)
	popupObj, d := mapWindowsPopupFromAPI(ctx, popupData)
	diags.Append(d...)

	avrObj, d := mapOptionalObject(ctx, data, "antivirus_registration", antivirusRegistrationAttrTypes(), func(m map[string]any) antivirusRegistrationModel {
		return antivirusRegistrationModel{
			Mode:    typeutils.StringFromMap(m, "mode"),
			Enabled: typeutils.BoolFromMap(m, "enabled"),
		}
	})
	diags.Append(d...)

	// attack_surface_reduction contains a nested credential_hardening object,
	// so it requires two levels of mapOptionalObject.
	asrData := getMap(data, "attack_surface_reduction")
	var asrObj types.Object
	if asrData != nil {
		chObj, d := mapOptionalObject(ctx, asrData, "credential_hardening", credentialHardeningAttrTypes(), func(m map[string]any) credentialHardeningModel {
			return credentialHardeningModel{
				Enabled: typeutils.BoolFromMap(m, "enabled"),
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

	if typeutils.IsKnown(wm.Events) {
		var em windowsEventsModel
		d = wm.Events.As(ctx, &em, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		events := map[string]any{}
		typeutils.SetBoolInMap(events, attrProcess, em.Process)
		typeutils.SetBoolInMap(events, "network", em.Network)
		typeutils.SetBoolInMap(events, "file", em.File)
		typeutils.SetBoolInMap(events, "dll_and_driver_load", em.DllAndDriverLoad)
		typeutils.SetBoolInMap(events, "dns", em.DNS)
		typeutils.SetBoolInMap(events, "registry", em.Registry)
		typeutils.SetBoolInMap(events, "security", em.Security)
		typeutils.SetBoolInMap(events, "authentication", em.Authentication)
		win["events"] = events
	}

	if typeutils.IsKnown(wm.Malware) {
		var mm malwareFullModel
		d = wm.Malware.As(ctx, &mm, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		malware := map[string]any{}
		typeutils.SetStringInMap(malware, "mode", mm.Mode)
		typeutils.SetBoolInMap(malware, "blocklist", mm.Blocklist)
		typeutils.SetBoolInMap(malware, attrOnWriteScan, mm.OnWriteScan)
		typeutils.SetBoolInMap(malware, attrNotifyUser, mm.NotifyUser)
		win["malware"] = malware
	}

	if typeutils.IsKnown(wm.Ransomware) {
		var rm protectionModeModel
		d = wm.Ransomware.As(ctx, &rm, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		ransomware := map[string]any{}
		typeutils.SetStringInMap(ransomware, "mode", rm.Mode)
		typeutils.SetBoolInMap(ransomware, attrSupported, rm.Supported)
		win[attrRansomware] = ransomware
	}

	if typeutils.IsKnown(wm.MemoryProtection) {
		var mm protectionModeModel
		d = wm.MemoryProtection.As(ctx, &mm, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		memProt := map[string]any{}
		typeutils.SetStringInMap(memProt, "mode", mm.Mode)
		typeutils.SetBoolInMap(memProt, attrSupported, mm.Supported)
		win["memory_protection"] = memProt
	}

	if typeutils.IsKnown(wm.BehaviorProtection) {
		var bm behaviorProtectionModel
		d = wm.BehaviorProtection.As(ctx, &bm, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		behProt := map[string]any{}
		typeutils.SetStringInMap(behProt, "mode", bm.Mode)
		typeutils.SetBoolInMap(behProt, attrSupported, bm.Supported)
		typeutils.SetBoolInMap(behProt, attrReputationService, bm.ReputationService)
		win["behavior_protection"] = behProt
	}

	if typeutils.IsKnown(wm.Popup) {
		var pm windowsPopupModel
		d = wm.Popup.As(ctx, &pm, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		popup := map[string]any{}
		setPopupItem(ctx, popup, "malware", pm.Malware, &diags)
		setPopupItem(ctx, popup, attrRansomware, pm.Ransomware, &diags)
		setPopupItem(ctx, popup, "memory_protection", pm.MemoryProtection, &diags)
		setPopupItem(ctx, popup, "behavior_protection", pm.BehaviorProtection, &diags)
		win[attrPopup] = popup
	}

	if typeutils.IsKnown(wm.Logging) {
		var lm loggingModel
		d = wm.Logging.As(ctx, &lm, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		logging := map[string]any{}
		typeutils.SetStringInMap(logging, "file", lm.File)
		win["logging"] = logging
	}

	if typeutils.IsKnown(wm.AntivirusRegistration) {
		var am antivirusRegistrationModel
		d = wm.AntivirusRegistration.As(ctx, &am, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		avr := map[string]any{}
		typeutils.SetStringInMap(avr, "mode", am.Mode)
		typeutils.SetBoolInMap(avr, "enabled", am.Enabled)
		win["antivirus_registration"] = avr
	}

	if typeutils.IsKnown(wm.AttackSurfaceReduction) {
		var am attackSurfaceReductionModel
		d = wm.AttackSurfaceReduction.As(ctx, &am, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		asr := map[string]any{}
		if typeutils.IsKnown(am.CredentialHardening) {
			var cm credentialHardeningModel
			d = am.CredentialHardening.As(ctx, &cm, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})
			diags.Append(d...)
			if diags.HasError() {
				return nil, diags
			}
			ch := map[string]any{}
			typeutils.SetBoolInMap(ch, "enabled", cm.Enabled)
			asr["credential_hardening"] = ch
		}
		win["attack_surface_reduction"] = asr
	}

	return win, diags
}

func windowsEventsAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		attrProcess:           types.BoolType,
		attrNetwork:           types.BoolType,
		attrFile:              types.BoolType,
		"dll_and_driver_load": types.BoolType,
		"dns":                 types.BoolType,
		"registry":            types.BoolType,
		"security":            types.BoolType,
		"authentication":      types.BoolType,
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
		attrMemoryProtection:       types.ObjectType{AttrTypes: protectionModeAttrTypes()},
		attrBehaviorProtection:     types.ObjectType{AttrTypes: behaviorProtectionAttrTypes()},
		attrPopup:                  types.ObjectType{AttrTypes: windowsPopupAttrTypes()},
		attrLogging:                types.ObjectType{AttrTypes: loggingAttrTypes()},
		"antivirus_registration":   types.ObjectType{AttrTypes: antivirusRegistrationAttrTypes()},
		"attack_surface_reduction": types.ObjectType{AttrTypes: attackSurfaceReductionAttrTypes()},
	}
}
