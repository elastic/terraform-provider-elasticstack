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

package datafeedstate

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// SetUnknownIfStateHasChanges returns a plan modifier that sets the current attribute to unknown
// if the state attribute has changed between state and config. During creation (no prior state),
// it sets the attribute to null when the desired state is "stopped" since a stopped datafeed has
// no start/end times.
func SetUnknownIfStateHasChanges() planmodifier.String {
	return setUnknownIfStateHasChanges{}
}

type setUnknownIfStateHasChanges struct{}

func (s setUnknownIfStateHasChanges) Description(_ context.Context) string {
	return "Sets the attribute value to unknown if the state attribute has changed, or null during creation with stopped state"
}

func (s setUnknownIfStateHasChanges) MarkdownDescription(ctx context.Context) string {
	return s.Description(ctx)
}

func (s setUnknownIfStateHasChanges) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	// Continue using the config value if it's explicitly set
	if typeutils.IsKnown(req.ConfigValue) {
		return
	}

	// During Create (no prior state), if the desired state is "stopped", the
	// attribute should be null rather than unknown — a stopped datafeed has no
	// start or end time. This prevents Terraform from expecting a computed
	// value that the provider cannot produce.
	if req.State.Raw.IsNull() {
		var configState types.String
		resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("state"), &configState)...)
		if resp.Diagnostics.HasError() {
			return
		}

		if configState.ValueString() == "stopped" {
			tflog.Debug(ctx, fmt.Sprintf("Plan modifier: setting %s to null during create with stopped state", req.Path))
			resp.PlanValue = types.StringNull()
		}
		return
	}

	if req.Config.Raw.IsNull() {
		return
	}

	// Get the state attribute from state and config to check if it has changed
	var stateValue, configValue types.String
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("state"), &stateValue)...)
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("state"), &configValue)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If the state attribute has changed between state and config, set the current attribute to Unknown
	if !stateValue.Equal(configValue) {
		tflog.Debug(ctx, fmt.Sprintf("Plan modifier: setting %s to unknown because state changed from %s to %s", req.Path, stateValue.ValueString(), configValue.ValueString()))
		resp.PlanValue = types.StringUnknown()
	}
}
