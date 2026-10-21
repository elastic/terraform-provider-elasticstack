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

package agentbuilderagent

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/agentbuilder"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

type accessControlModel struct {
	AccessMode types.String `tfsdk:"access_mode"`
}

func accessControlAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{attrAccessMode: types.StringType}
}

func (model agentModel) accessMode(ctx context.Context) (string, diag.Diagnostics) {
	var ac accessControlModel
	diags := model.AccessControl.As(ctx, &ac, basetypes.ObjectAsOptions{})
	return ac.AccessMode.ValueString(), diags
}

func (model *agentModel) populateAccessControl(ctx context.Context, data *models.AgentAccessControl) diag.Diagnostics {
	if data == nil {
		model.AccessControl = types.ObjectNull(accessControlAttrTypes())
		return nil
	}
	obj, diags := types.ObjectValueFrom(ctx, accessControlAttrTypes(), accessControlModel{
		AccessMode: types.StringValue(data.AccessMode),
	})
	model.AccessControl = obj
	return diags
}

// enforceAccessControlVersion returns a version-requirement error when a write
// that sends access_control targets a server older than 9.5.0.
func enforceAccessControlVersion(ctx context.Context, client *clients.KibanaScopedClient, sending bool) diag.Diagnostics {
	if !sending {
		return nil
	}
	supported, diags := client.EnforceMinVersion(ctx, agentbuilder.MinAccessControlAPIVersion)
	if diags.HasError() {
		return diags
	}
	if !supported {
		diags.AddAttributeError(path.Root(attrAccessControl), "Unsupported server version",
			fmt.Sprintf("The access_control attribute requires Elastic Stack v%s or later.", agentbuilder.MinAccessControlAPIVersion))
	}
	return diags
}
