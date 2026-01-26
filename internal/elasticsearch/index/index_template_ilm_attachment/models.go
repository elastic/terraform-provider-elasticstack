package index_template_ilm_attachment

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// tfModel represents the Terraform state model for this resource.
type tfModel struct {
	ID            types.String `tfsdk:"id"`
	IndexTemplate types.String `tfsdk:"index_template"`
	LifecycleName types.String `tfsdk:"lifecycle_name"`
}

// getComponentTemplateName returns the name of the @custom component template.
func (m *tfModel) getComponentTemplateName() string {
	return m.IndexTemplate.ValueString() + "@custom"
}

// GetID parses and returns the composite ID from the model.
func (m *tfModel) GetID() (*clients.CompositeId, diag.Diagnostics) {
	compId, sdkDiags := clients.CompositeIdFromStr(m.ID.ValueString())
	if sdkDiags.HasError() {
		return nil, diagutil.FrameworkDiagsFromSDK(sdkDiags)
	}
	return compId, nil
}

// mergeILMSetting adds the ILM lifecycle.name setting to existing settings.
func mergeILMSetting(existingSettings map[string]interface{}, lifecycleName string) map[string]interface{} {
	if existingSettings == nil {
		existingSettings = make(map[string]interface{})
	}
	existingSettings["index.lifecycle.name"] = lifecycleName
	return existingSettings
}

// removeILMSetting removes the index.lifecycle.name setting from the settings map.
func removeILMSetting(settings map[string]interface{}) map[string]interface{} {
	if settings == nil {
		return nil
	}
	delete(settings, "index.lifecycle.name")
	if len(settings) == 0 {
		return nil
	}
	return settings
}

// isComponentTemplateEmpty checks if a component template has no meaningful content.
func isComponentTemplateEmpty(template *models.Template) bool {
	if template == nil {
		return true
	}
	return len(template.Settings) == 0 &&
		len(template.Mappings) == 0 &&
		len(template.Aliases) == 0
}

// extractILMSetting extracts the index.lifecycle.name setting from component template settings.
// Elasticsearch returns settings in a nested structure: {"index": {"lifecycle": {"name": "policy"}}}
func extractILMSetting(template *models.Template) string {
	if template == nil || template.Settings == nil {
		return ""
	}

	// Elasticsearch returns settings in nested structure
	if indexSettings, ok := template.Settings["index"].(map[string]interface{}); ok {
		if lifecycleSettings, ok := indexSettings["lifecycle"].(map[string]interface{}); ok {
			if lifecycleName, ok := lifecycleSettings["name"].(string); ok {
				return lifecycleName
			}
		}
	}

	return ""
}

// readILMAttachment reads the component template and updates the model with the actual ILM setting.
// This ensures state consistency after create/update operations.
func readILMAttachment(ctx context.Context, client *clients.ApiClient, model *tfModel) diag.Diagnostics {
	var diags diag.Diagnostics

	componentTemplateName := model.getComponentTemplateName()

	tpl, sdkDiags := elasticsearch.GetComponentTemplate(ctx, client, componentTemplateName)
	if sdkDiags.HasError() {
		diags.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
		return diags
	}

	if tpl == nil {
		diags.AddError(
			"Component template not found",
			fmt.Sprintf("Component template %s was not found after create/update", componentTemplateName),
		)
		return diags
	}

	lifecycleName := extractILMSetting(tpl.ComponentTemplate.Template)
	if lifecycleName == "" {
		diags.AddError(
			"ILM setting not found",
			fmt.Sprintf("ILM setting was not found in component template %s after create/update", componentTemplateName),
		)
		return diags
	}

	model.LifecycleName = types.StringValue(lifecycleName)
	return diags
}
