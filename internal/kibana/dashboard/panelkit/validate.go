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
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

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
