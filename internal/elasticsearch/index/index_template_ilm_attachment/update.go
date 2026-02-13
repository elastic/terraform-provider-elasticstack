package index_template_ilm_attachment

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func (r *Resource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan tfModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Check Elasticsearch version
	serverVersion, sdkDiags := r.client.ServerVersion(ctx)
	if sdkDiags.HasError() {
		resp.Diagnostics.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
		return
	}

	if serverVersion.LessThan(MinVersion) {
		resp.Diagnostics.AddError(
			"Unsupported Elasticsearch Version",
			fmt.Sprintf(
				"This resource requires Elasticsearch %s or later (current: %s). "+
					"This resource is not supported on this Elasticsearch version.",
				MinVersion, serverVersion,
			),
		)
		return
	}

	componentTemplateName := plan.getComponentTemplateName()

	// Read existing component template to preserve other settings
	existing, sdkDiags := elasticsearch.GetComponentTemplate(ctx, r.client, componentTemplateName, true)
	if sdkDiags.HasError() {
		resp.Diagnostics.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
		return
	}

	// Build component template, preserving existing content if any
	var componentTemplate models.ComponentTemplate
	if existing != nil {
		componentTemplate = existing.ComponentTemplate

		// Warn if the existing template has a version field
		if componentTemplate.Version != nil {
			tflog.Warn(ctx, "Existing component template has a version field. This resource does not update the version when modifying the template. If you rely on version tracking for change detection, consider using elasticstack_elasticsearch_component_template instead.", map[string]interface{}{
				"component_template": componentTemplateName,
				"existing_version":   *componentTemplate.Version,
			})
		}
	}
	componentTemplate.Name = componentTemplateName

	// Ensure template exists
	if componentTemplate.Template == nil {
		componentTemplate.Template = &models.Template{}
	}

	// Merge the ILM setting
	componentTemplate.Template.Settings = mergeILMSetting(
		componentTemplate.Template.Settings,
		plan.LifecycleName.ValueString(),
	)

	// Write the component template
	if sdkDiags := elasticsearch.PutComponentTemplate(ctx, r.client, &componentTemplate); sdkDiags.HasError() {
		resp.Diagnostics.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
		return
	}

	// Read back to ensure state consistency
	diags, found := readILMAttachment(ctx, r.client, &plan)
	resp.Diagnostics.Append(diags...)
	if !found && !resp.Diagnostics.HasError() {
		resp.Diagnostics.AddError(
			"Component template not found",
			fmt.Sprintf("Component template %s was not found after update", plan.getComponentTemplateName()),
		)
	}
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}
