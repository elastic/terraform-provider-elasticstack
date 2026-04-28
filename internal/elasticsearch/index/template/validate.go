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

package template

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ValidateConfig enforces plan-time rules that cannot be expressed as block Required flags (REQ-038 / REQ-032 scenario).
func (r *Resource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var templateObj types.Object
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("template"), &templateObj)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if templateObj.IsNull() || templateObj.IsUnknown() {
		return
	}

	dsoVal, ok := templateObj.Attributes()["data_stream_options"]
	if !ok {
		return
	}
	dso, ok := dsoVal.(types.Object)
	if !ok || dso.IsNull() || dso.IsUnknown() {
		return
	}

	fsVal, ok := dso.Attributes()["failure_store"]
	if !ok {
		return
	}
	fs, ok := fsVal.(types.Object)
	if !ok {
		return
	}
	if fs.IsUnknown() {
		return
	}
	if fs.IsNull() {
		resp.Diagnostics.AddAttributeError(
			path.Root("template").AtName("data_stream_options").AtName("failure_store"),
			errSummaryMissingFailureStore,
			errDetailMissingFailureStore,
		)
	}
}
