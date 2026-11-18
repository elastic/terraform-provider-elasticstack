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

// mapPopupItemFromAPI extracts a popup item sub-object (message/enabled) from the API response map.
func mapPopupItemFromAPI(ctx context.Context, data map[string]any, key string) (types.Object, diag.Diagnostics) {
	itemData := getMap(data, key)
	return types.ObjectValueFrom(ctx, popupItemAttrTypes(), popupItemModel{
		Message: typeutils.StringFromMap(itemData, attrMessage),
		Enabled: typeutils.BoolFromMap(itemData, attrEnabled),
	})
}

// setPopupItem extracts a popup item from a Terraform object and adds it to the map.
func setPopupItem(ctx context.Context, m map[string]any, key string, obj types.Object, diags *diag.Diagnostics) {
	if obj.IsNull() || obj.IsUnknown() {
		return
	}
	var pm popupItemModel
	d := obj.As(ctx, &pm, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})
	diags.Append(d...)
	item := map[string]any{}
	typeutils.SetStringInMap(item, attrMessage, pm.Message)
	typeutils.SetBoolInMap(item, attrEnabled, pm.Enabled)
	m[key] = item
}

func mapWindowsPopupFromAPI(ctx context.Context, data map[string]any) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics
	if len(data) == 0 {
		return types.ObjectNull(windowsPopupAttrTypes()), diags
	}

	malwareObj, d := mapPopupItemFromAPI(ctx, data, attrMalware)
	diags.Append(d...)

	ransomwareObj, d := mapPopupItemFromAPI(ctx, data, attrRansomware)
	diags.Append(d...)

	memProtObj, d := mapPopupItemFromAPI(ctx, data, attrMemoryProtection)
	diags.Append(d...)

	behProtObj, d := mapPopupItemFromAPI(ctx, data, attrBehaviorProtection)
	diags.Append(d...)

	deviceControlObj, d := mapPopupItemFromAPI(ctx, data, attrDeviceControl)
	diags.Append(d...)

	obj, d := types.ObjectValueFrom(ctx, windowsPopupAttrTypes(), windowsPopupModel{
		Malware:            malwareObj,
		Ransomware:         ransomwareObj,
		MemoryProtection:   memProtObj,
		BehaviorProtection: behProtObj,
		DeviceControl:      deviceControlObj,
	})
	diags.Append(d...)
	return obj, diags
}

// mapMacPopupFromAPI maps the macOS popup settings, which share the Windows shape.
func mapMacPopupFromAPI(ctx context.Context, data map[string]any) (types.Object, diag.Diagnostics) {
	return mapWindowsPopupFromAPI(ctx, data)
}

func mapLinuxPopupFromAPI(ctx context.Context, data map[string]any) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics
	if len(data) == 0 {
		return types.ObjectNull(linuxPopupAttrTypes()), diags
	}

	malwareObj, d := mapPopupItemFromAPI(ctx, data, attrMalware)
	diags.Append(d...)

	memProtObj, d := mapPopupItemFromAPI(ctx, data, attrMemoryProtection)
	diags.Append(d...)

	behProtObj, d := mapPopupItemFromAPI(ctx, data, attrBehaviorProtection)
	diags.Append(d...)

	obj, d := types.ObjectValueFrom(ctx, linuxPopupAttrTypes(), linuxPopupModel{
		Malware:            malwareObj,
		MemoryProtection:   memProtObj,
		BehaviorProtection: behProtObj,
	})
	diags.Append(d...)
	return obj, diags
}

func popupItemAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		attrMessage: types.StringType,
		attrEnabled: types.BoolType,
	}
}

func windowsPopupAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		attrMalware:            types.ObjectType{AttrTypes: popupItemAttrTypes()},
		attrRansomware:         types.ObjectType{AttrTypes: popupItemAttrTypes()},
		attrMemoryProtection:   types.ObjectType{AttrTypes: popupItemAttrTypes()},
		attrBehaviorProtection: types.ObjectType{AttrTypes: popupItemAttrTypes()},
		attrDeviceControl:      types.ObjectType{AttrTypes: popupItemAttrTypes()},
	}
}

// macPopupAttrTypes returns the macOS popup attribute types, which share the
// Windows shape.
func macPopupAttrTypes() map[string]attr.Type {
	return windowsPopupAttrTypes()
}

func linuxPopupAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		attrMalware:            types.ObjectType{AttrTypes: popupItemAttrTypes()},
		attrMemoryProtection:   types.ObjectType{AttrTypes: popupItemAttrTypes()},
		attrBehaviorProtection: types.ObjectType{AttrTypes: popupItemAttrTypes()},
	}
}
