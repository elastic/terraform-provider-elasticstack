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

package panelkit

import (
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// RejectConfigJSON returns an error diagnostic when pm.ConfigJSON is set (known and non-null),
// which is unsupported for panel types that require a typed config block instead.
// panelType is used solely for the human-readable error message (e.g. "discover_session").
func RejectConfigJSON(pm models.PanelModel, panelType string) diag.Diagnostics {
	if !typeutils.IsKnown(pm.ConfigJSON) || pm.ConfigJSON.IsNull() {
		return nil
	}
	var diags diag.Diagnostics
	diags.AddError(
		"Unsupported panel type for config_json",
		fmt.Sprintf(
			"Panel-level `config_json` is not supported for `%s` panels. Use `%s_config` instead.",
			panelType, panelType,
		),
	)
	return diags
}

// ValidateDataViewFieldName validates that data_view_id and field_name are present in attrs,
// using the flat or nested shape detected by ResolvePanelAttrsShape. cfgLabel is used as the
// error summary (e.g. "Invalid options list control configuration").
func ValidateDataViewFieldName(attrs map[string]attr.Value, configKey, cfgLabel string, attrPath path.Path) diag.Diagnostics {
	var out diag.Diagnostics
	flat, obj, shaped := ResolvePanelAttrsShape(attrs, configKey, "data_view_id", "field_name")
	if !shaped {
		return out
	}

	cfgPath := attrPath
	var dataViewAttr, fieldNameAttr attr.Value
	switch {
	case flat:
		dataViewAttr, fieldNameAttr = attrs["data_view_id"], attrs["field_name"]
	default:
		at := obj.Attributes()
		cfgPath = attrPath.AtName(configKey)
		dataViewAttr, fieldNameAttr = at["data_view_id"], at["field_name"]
	}

	writeErr := func(field, msg string) {
		out.AddAttributeError(cfgPath.AtName(field), cfgLabel, msg)
	}
	if deferDV, missDV := StringAttrDeferOrMissing(dataViewAttr); !deferDV && missDV {
		writeErr("data_view_id", "`data_view_id` is required.")
	}
	if deferFN, missFN := StringAttrDeferOrMissing(fieldNameAttr); !deferFN && missFN {
		writeErr("field_name", "`field_name` is required.")
	}
	return out
}

// ResolveConfigBlock resolves a config block's path and shaped state, handling null/unknown guards.
// It encodes the shared "shape resolution → missing-config error → null/unknown guard" pattern used
// across SLO panel ValidatePanelConfig implementations.
// Returns flat, obj, cfgPath, skip (true means caller should return immediately), and any diagnostics.
func ResolveConfigBlock(
	attrs map[string]attr.Value,
	attrPath path.Path,
	cfgKey, missingErrSummary, missingErrDetail string,
	flatKeys ...string,
) (flat bool, obj types.Object, cfgPath path.Path, skip bool, diags diag.Diagnostics) {
	cfgPath = attrPath
	flat, obj, shaped := ResolvePanelAttrsShape(attrs, cfgKey, flatKeys...)
	if !shaped {
		diags.AddAttributeError(attrPath.AtName(cfgKey), missingErrSummary, missingErrDetail)
		return false, types.Object{}, cfgPath, true, diags
	}
	if !flat {
		cfgPath = attrPath.AtName(cfgKey)
		nestedRaw := attrs[cfgKey]
		if nestedRaw != nil {
			switch {
			case nestedRaw.IsUnknown():
				return false, obj, cfgPath, true, diags
			case nestedRaw.IsNull():
				diags.AddAttributeError(cfgPath, missingErrSummary, missingErrDetail)
				return false, obj, cfgPath, true, diags
			}
		}
	}
	return flat, obj, cfgPath, false, diags
}

// ValidateRequiredStringField validates that a required string attribute is present and known.
// It looks up key from flat attrs or from obj.Attributes() based on flat, then applies
// StringAttrDeferOrMissing. Returns deferred=true when validation should be skipped (value unknown),
// and any error diagnostics when the value is missing.
func ValidateRequiredStringField(attrs map[string]attr.Value, obj types.Object, flat bool, cfgPath path.Path, key, errSummary, errDetail string) (deferred bool, diags diag.Diagnostics) {
	var v attr.Value
	if flat {
		v = attrs[key]
	} else {
		v = obj.Attributes()[key]
	}
	deferred, missing := StringAttrDeferOrMissing(v)
	if deferred {
		return true, nil
	}
	if missing {
		diags.AddAttributeError(cfgPath.AtName(key), errSummary, errDetail)
	}
	return false, diags
}

// ValidateRequiredListField validates that a required list attribute (looked up from flat attrs or
// obj.Attributes() based on flat) is present and has an element count within [minSize, maxSize]. A
// maxSize of 0 means no upper bound. requiredMsg is used when the list is null; sizeMsg is used when
// the element count is out of range. Unknown values defer validation to a later plan.
func ValidateRequiredListField(
	attrs map[string]attr.Value, obj types.Object, flat bool, cfgPath path.Path, key string,
	minSize, maxSize int, errSummary, requiredMsg, sizeMsg string,
) diag.Diagnostics {
	var out diag.Diagnostics
	var v attr.Value
	if flat {
		v = attrs[key]
	} else {
		v = obj.Attributes()[key]
	}
	switch {
	case v == nil || v.IsUnknown():
	case v.IsNull():
		out.AddAttributeError(cfgPath.AtName(key), errSummary, requiredMsg)
	default:
		list, ok := v.(types.List)
		if !ok || list.IsNull() || list.IsUnknown() {
			return out
		}
		if n := len(list.Elements()); n < minSize || (maxSize > 0 && n > maxSize) {
			out.AddAttributeError(cfgPath.AtName(key), errSummary, sizeMsg)
		}
	}
	return out
}
