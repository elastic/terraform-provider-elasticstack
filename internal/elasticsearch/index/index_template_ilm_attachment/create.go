package index_template_ilm_attachment

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func (r *Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
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

	// Generate the resource ID
	componentTemplateName := plan.getComponentTemplateName()
	id, sdkDiags := r.client.ID(ctx, componentTemplateName)
	if sdkDiags.HasError() {
		resp.Diagnostics.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
		return
	}
	plan.ID = types.StringValue(id.String())

	// Read existing component template (if any) to preserve other settings
	existing, sdkDiags := elasticsearch.GetComponentTemplate(ctx, r.client, componentTemplateName)
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

		// Warn if an ILM setting already exists (potential conflict)
		existingILM := extractILMSetting(componentTemplate.Template)
		if existingILM != "" {
			tflog.Warn(ctx, "Component template already has an ILM policy configured. This resource will overwrite it. If this is unexpected, another process may be managing this setting.", map[string]interface{}{
				"component_template": componentTemplateName,
				"existing_ilm":       existingILM,
				"new_ilm":            plan.LifecycleName.ValueString(),
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
	resp.Diagnostics.Append(readILMAttachment(ctx, r.client, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}
