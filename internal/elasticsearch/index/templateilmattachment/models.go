package templateilmattachment

import (
	"context"

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
func (m *tfModel) GetID() (*clients.CompositeID, diag.Diagnostics) {
	compID, sdkDiags := clients.CompositeIDFromStr(m.ID.ValueString())
	if sdkDiags.HasError() {
		return nil, diagutil.FrameworkDiagsFromSDK(sdkDiags)
	}
	return compID, nil
}

// mergeILMSetting adds the ILM lifecycle.name setting to existing settings.
// We use flat form (index.lifecycle.name) for simpler processing; Get is called with flat_settings=true.
func mergeILMSetting(existingSettings map[string]any, lifecycleName string) map[string]any {
	if existingSettings == nil {
		existingSettings = make(map[string]any)
	}
	existingSettings["index.lifecycle.name"] = lifecycleName
	return existingSettings
}

// removeILMSetting removes the index.lifecycle.name setting from the settings map (flat form).
func removeILMSetting(settings map[string]any) map[string]any {
	if settings == nil {
		return nil
	}
	delete(settings, "index.lifecycle.name")
	return pruneEmpty(settings)
}

func pruneEmpty(settings map[string]any) map[string]any {
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

// extractILMSetting extracts the index.lifecycle.name setting from component template settings (flat form).
func extractILMSetting(template *models.Template) string {
	if template == nil || template.Settings == nil {
		return ""
	}
	if v, ok := template.Settings["index.lifecycle.name"].(string); ok {
		return v
	}
	return ""
}

// readILMAttachment reads the component template and updates the model with the actual ILM setting.
// It returns (diags, true) on success, (diags, false) on SDK error, and (nil, false) when the
// template or ILM setting is missing. The caller decides how to handle "not found" (e.g. Read
// removes from state, Create/Update report an error).
func readILMAttachment(ctx context.Context, client *clients.APIClient, model *tfModel) (diag.Diagnostics, bool) {
	var diags diag.Diagnostics

	componentTemplateName := model.getComponentTemplateName()

	tpl, sdkDiags := elasticsearch.GetComponentTemplate(ctx, client, componentTemplateName, true)
	if sdkDiags.HasError() {
		diags.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
		return diags, false
	}

	if tpl == nil {
		return nil, false
	}

	lifecycleName := extractILMSetting(tpl.ComponentTemplate.Template)
	if lifecycleName == "" {
		return nil, false
	}

	model.LifecycleName = types.StringValue(lifecycleName)
	return diags, true
}
