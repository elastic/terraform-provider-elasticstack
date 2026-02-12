package index_template_ilm_attachment

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
func (m *tfModel) GetID() (*clients.CompositeId, diag.Diagnostics) {
	compId, sdkDiags := clients.CompositeIdFromStr(m.ID.ValueString())
	if sdkDiags.HasError() {
		return nil, diagutil.FrameworkDiagsFromSDK(sdkDiags)
	}
	return compId, nil
}

// mergeILMSetting adds the ILM lifecycle.name setting to existing settings.
// Uses nested form {"index": {"lifecycle": {"name": "policy"}}} to match the Get Component
// Template API; we request nested form explicitly (flat_settings=false).
// The Put API accepts both flat and nested; we use nested for consistency with the response.
func mergeILMSetting(existingSettings map[string]interface{}, lifecycleName string) map[string]interface{} {
	if existingSettings == nil {
		existingSettings = make(map[string]interface{})
	}
	indexVal, _ := existingSettings["index"].(map[string]interface{})
	if indexVal == nil {
		indexVal = make(map[string]interface{})
		existingSettings["index"] = indexVal
	}
	indexVal["lifecycle"] = map[string]interface{}{"name": lifecycleName}
	return existingSettings
}

// removeILMSetting removes the index.lifecycle.name setting from the settings map.
// Elasticsearch uses nested form: {"index": {"lifecycle": {"name": "policy"}}}.
// We remove that path and prune empty parent maps.
func removeILMSetting(settings map[string]interface{}) map[string]interface{} {
	if settings == nil {
		return nil
	}
	indexVal, ok := settings["index"].(map[string]interface{})
	if !ok {
		return pruneEmpty(settings)
	}
	lifecycleVal, ok := indexVal["lifecycle"].(map[string]interface{})
	if !ok {
		return pruneEmpty(settings)
	}
	delete(lifecycleVal, "name")
	if len(lifecycleVal) == 0 {
		delete(indexVal, "lifecycle")
	}
	if len(indexVal) == 0 {
		delete(settings, "index")
	}
	return pruneEmpty(settings)
}

func pruneEmpty(settings map[string]interface{}) map[string]interface{} {
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
// It returns (diags, true) on success, (diags, false) on SDK error, and (nil, false) when the
// template or ILM setting is missing. The caller decides how to handle "not found" (e.g. Read
// removes from state, Create/Update report an error).
func readILMAttachment(ctx context.Context, client *clients.ApiClient, model *tfModel) (diag.Diagnostics, bool) {
	var diags diag.Diagnostics

	componentTemplateName := model.getComponentTemplateName()

	tpl, sdkDiags := elasticsearch.GetComponentTemplate(ctx, client, componentTemplateName)
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
