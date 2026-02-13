package index_template_ilm_attachment

import (
	"context"
	"strings"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

const customSuffix = "@custom"

func (r *Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state tfModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	compId, diags := state.GetID()
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	componentTemplateName := compId.ResourceId
	// Derive index_template from component template name for import (component name is <index_template>@custom)
	if !utils.IsKnown(state.IndexTemplate) {
		state.IndexTemplate = types.StringValue(strings.TrimSuffix(componentTemplateName, customSuffix))
	}

	diags, found := readILMAttachment(ctx, r.client, &state)
	if !found {
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		tflog.Warn(ctx, "Component template or ILM setting not found, removing from state", map[string]interface{}{
			"name": componentTemplateName,
		})
		resp.State.RemoveResource(ctx)
		return
	}
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
