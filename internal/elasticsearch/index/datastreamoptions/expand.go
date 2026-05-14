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
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Expand converts a Terraform data_stream_options object into its wire model.
// Returns nil when the object is null, unknown, or carries no failure_store value.
func Expand(obj types.Object) (*models.DataStreamOptions, diag.Diagnostics) {
	var diags diag.Diagnostics
	if obj.IsNull() || obj.IsUnknown() {
		return nil, diags
	}
	attrs := obj.Attributes()
	fsVal, ok := attrs["failure_store"]
	if !ok || fsVal.IsNull() || fsVal.IsUnknown() {
		return nil, diags
	}
	fsObj, ok := fsVal.(types.Object)
	if !ok {
		diags.AddError("Internal error", fmt.Sprintf("expected Object for failure_store, got %T", fsVal))
		return nil, diags
	}
	fsAttrs := fsObj.Attributes()
	out := &models.DataStreamOptions{
		FailureStore: &models.FailureStoreOptions{},
	}
	if en, ok := fsAttrs["enabled"]; ok && !en.IsNull() && !en.IsUnknown() {
		if b, ok := en.(types.Bool); ok {
			out.FailureStore.Enabled = b.ValueBool()
		}
	}
	if lcVal, ok := fsAttrs["lifecycle"]; ok && !lcVal.IsNull() && !lcVal.IsUnknown() {
		lcObj, ok := lcVal.(types.Object)
		if !ok {
			diags.AddError("Internal error", fmt.Sprintf("expected Object for failure_store.lifecycle, got %T", lcVal))
			return nil, diags
		}
		lcAttrs := lcObj.Attributes()
		if drAttr, ok := lcAttrs["data_retention"]; ok && !drAttr.IsNull() && !drAttr.IsUnknown() {
			if drStr, ok := drAttr.(types.String); ok {
				out.FailureStore.Lifecycle = &models.FailureStoreLifecycle{
					DataRetention: drStr.ValueString(),
				}
			}
		}
	}
	return out, diags
}