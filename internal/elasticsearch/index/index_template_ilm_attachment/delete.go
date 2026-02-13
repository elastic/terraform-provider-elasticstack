package index_template_ilm_attachment

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func (r *Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
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

	// Read existing component template
	existing, sdkDiags := elasticsearch.GetComponentTemplate(ctx, r.client, componentTemplateName, true)
	if sdkDiags.HasError() {
		resp.Diagnostics.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
		return
	}

	if existing == nil {
		// Already gone
		tflog.Debug(ctx, "Component template already deleted", map[string]interface{}{
			"name": componentTemplateName,
		})
		return
	}

	// Remove the ILM setting from the template
	if existing.ComponentTemplate.Template != nil {
		existing.ComponentTemplate.Template.Settings = removeILMSetting(existing.ComponentTemplate.Template.Settings)
	}

	// Always update the component template with the ILM setting removed; never delete it.
	// The template (e.g. logs-system.syslog@custom) is typically used by an index template
	// (e.g. from Fleet); deleting it would fail with "cannot be removed as they are still in use".
	componentTemplate := models.ComponentTemplate{
		Name:     componentTemplateName,
		Template: existing.ComponentTemplate.Template,
		Meta:     existing.ComponentTemplate.Meta,
		Version:  existing.ComponentTemplate.Version,
	}
	if sdkDiags := elasticsearch.PutComponentTemplate(ctx, r.client, &componentTemplate); sdkDiags.HasError() {
		resp.Diagnostics.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
		return
	}
}
