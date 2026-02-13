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

	// Check if the component template is now empty
	if isComponentTemplateEmpty(existing.ComponentTemplate.Template) {
		// Delete the entire component template
		tflog.Debug(ctx, "Component template is empty after removing ILM setting, deleting", map[string]interface{}{
			"name": componentTemplateName,
		})
		if sdkDiags := elasticsearch.DeleteComponentTemplate(ctx, r.client, componentTemplateName); sdkDiags.HasError() {
			resp.Diagnostics.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
			return
		}
	} else {
		// Update the component template with the ILM setting removed
		tflog.Debug(ctx, "Component template has other settings, preserving", map[string]interface{}{
			"name": componentTemplateName,
		})
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
}
