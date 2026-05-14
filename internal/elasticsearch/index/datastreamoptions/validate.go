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
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ValidateRequiresFailureStore enforces the rule that when template.data_stream_options
// is set, the failure_store block must also be set. Shared between index templates and
// component templates: both use the same template.data_stream_options.failure_store path.
func ValidateRequiresFailureStore(ctx context.Context, config tfsdk.Config) diag.Diagnostics {
	var diags diag.Diagnostics
	var templateObj types.Object
	diags.Append(config.GetAttribute(ctx, path.Root("template"), &templateObj)...)
	if diags.HasError() {
		return diags
	}
	if templateObj.IsNull() || templateObj.IsUnknown() {
		return diags
	}

	dsoVal, ok := templateObj.Attributes()["data_stream_options"]
	if !ok {
		return diags
	}
	dso, ok := dsoVal.(types.Object)
	if !ok || dso.IsNull() || dso.IsUnknown() {
		return diags
	}

	fsVal, ok := dso.Attributes()["failure_store"]
	if !ok {
		return diags
	}
	fs, ok := fsVal.(types.Object)
	if !ok {
		return diags
	}
	if fs.IsUnknown() {
		return diags
	}
	if fs.IsNull() {
		diags.AddAttributeError(
			path.Root("template").AtName("data_stream_options").AtName("failure_store"),
			ErrSummaryMissingFailureStore,
			ErrDetailMissingFailureStore,
		)
	}
	return diags
}
