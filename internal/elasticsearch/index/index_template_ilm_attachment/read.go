package index_template_ilm_attachment

import (
	"context"
	"strings"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
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

	// Read the component template
	tpl, sdkDiags := elasticsearch.GetComponentTemplate(ctx, r.client, componentTemplateName)
	if sdkDiags.HasError() {
		resp.Diagnostics.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
		return
	}

	if tpl == nil {
		// Resource was deleted outside Terraform
		tflog.Warn(ctx, "Component template not found, removing from state", map[string]interface{}{
			"name": componentTemplateName,
		})
		resp.State.RemoveResource(ctx)
		return
	}

	// Extract the lifecycle name from settings
	lifecycleName := extractILMSetting(tpl.ComponentTemplate.Template)
	if lifecycleName == "" {
		// The ILM setting was removed outside Terraform
		tflog.Warn(ctx, "ILM setting not found in component template, removing from state", map[string]interface{}{
			"name": componentTemplateName,
		})
		resp.State.RemoveResource(ctx)
		return
	}

	state.LifecycleName = types.StringValue(lifecycleName)

	// Derive the index_template from the component template name (for import support)
	// Component template name is always <index_template>@custom
	if state.IndexTemplate.IsNull() || state.IndexTemplate.IsUnknown() {
		indexTemplateName := strings.TrimSuffix(componentTemplateName, customSuffix)
		state.IndexTemplate = types.StringValue(indexTemplateName)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
