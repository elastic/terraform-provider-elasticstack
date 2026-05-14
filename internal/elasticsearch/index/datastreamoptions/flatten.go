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

package datastreamoptions

import (
	estypes "github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Flatten converts a wire data_stream_options object back into a Terraform object.
// The caller is responsible for checking dso != nil and dso.FailureStore != nil
// before invoking Flatten; an absent block should map to types.ObjectNull(AttrTypes()).
func Flatten(dso *estypes.DataStreamOptions) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics
	fs := dso.FailureStore
	fsAttrs := map[string]attr.Value{
		"enabled":   types.BoolValue(typeutils.Deref(fs.Enabled)),
		"lifecycle": types.ObjectNull(FailureStoreLifecycleAttrTypes()),
	}
	if fs.Lifecycle != nil {
		dataRetention := types.StringNull()
		if fs.Lifecycle.DataRetention != nil {
			if dr, ok := fs.Lifecycle.DataRetention.(string); ok && dr != "" {
				dataRetention = types.StringValue(dr)
			}
		}
		lcAttrs := map[string]attr.Value{
			"data_retention": dataRetention,
		}
		lcObj, d := types.ObjectValue(FailureStoreLifecycleAttrTypes(), lcAttrs)
		diags.Append(d...)
		if diags.HasError() {
			return types.ObjectUnknown(AttrTypes()), diags
		}
		fsAttrs["lifecycle"] = lcObj
	}
	fsObj, d := types.ObjectValue(FailureStoreAttrTypes(), fsAttrs)
	diags.Append(d...)
	if diags.HasError() {
		return types.ObjectUnknown(AttrTypes()), diags
	}
	outer := map[string]attr.Value{
		"failure_store": fsObj,
	}
	obj, d := types.ObjectValue(AttrTypes(), outer)
	diags.Append(d...)
	return obj, diags
}