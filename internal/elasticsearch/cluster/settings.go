package cluster

import (
	"context"
	"fmt"
	"reflect"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	schemautil "github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ResourceSettings() *schema.Resource {
	settingSchema := &schema.Resource{
		Schema: map[string]*schema.Schema{
			"setting": {
				Description: "Defines the setting in the cluster.",
				Type:        schema.TypeSet,
				Required:    true,
				MinItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Description: "The name of the setting to set and track.",
							Type:        schema.TypeString,
							Required:    true,
						},
						"value": {
							Description: "The value of the setting to set and track.",
							Type:        schema.TypeString,
							Optional:    true,
						},
						"value_list": {
							Description: "The list of values to be set for the key, where the list is required.",
							Type:        schema.TypeList,
							Optional:    true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
			},
		},
	}

	settingsSchema := map[string]*schema.Schema{
		"id": {
			Description: "Internal identifier of the resource",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"persistent": {
			Description: "Settings will apply across restarts.",
			Type:        schema.TypeList,
			MaxItems:    1,
			Optional:    true,
			Elem:        settingSchema,
		},
		"transient": {
			Description: "Settings do not survive a full cluster restart.",
			Type:        schema.TypeList,
			MaxItems:    1,
			Optional:    true,
			Elem:        settingSchema,
		},
	}

	schemautil.AddConnectionSchema(settingsSchema)

	return &schema.Resource{
		Description: settingsResourceDescription,

		CreateContext: resourceClusterSettingsPut,
		UpdateContext: resourceClusterSettingsPut,
		ReadContext:   resourceClusterSettingsRead,
		DeleteContext: resourceClusterSettingsDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: settingsSchema,
	}
}

func resourceClusterSettingsPut(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client, diags := clients.NewAPIClientFromSDKResource(d, meta)
	if diags.HasError() {
		return diags
	}
	id, diags := client.ID(ctx, "cluster-settings")
	if diags.HasError() {
		return diags
	}

	settings, diags := getConfiguredSettings(d)
	if diags.HasError() {
		return diags
	}
	for _, v := range []string{"persistent", "transient"} {
		if d.HasChange(v) {
			oldValue, newValue := d.GetChange(v)
			diags = updateRemovedSettings(v, oldValue, newValue, settings)
			if diags.HasError() {
				return diags
			}
		}
	}
	if diags := elasticsearch.PutSettings(ctx, client, settings); diags.HasError() {
		return diags
	}
	d.SetId(id.String())
	return resourceClusterSettingsRead(ctx, d, meta)
}

// Updates the map of settings in place if there is a difference between old and new list of settings
func updateRemovedSettings(name string, oldValue, newValue any, settings map[string]any) diag.Diagnostics {
	var diags diag.Diagnostics
	if oldValue != nil && newValue != nil {
		oldSettings := make(map[string]any)
		if len(oldValue.([]any)) > 0 {
			oldSettings, _ = expandSettings(oldValue)
		}
		newSettings := make(map[string]any)
		if len(newValue.([]any)) > 0 {
			newSettings, diags = expandSettings(newValue)
			if diags.HasError() {
				return diags
			}
		}

		if !reflect.DeepEqual(oldSettings, newSettings) {
			for s := range oldSettings {
				if _, ok := newSettings[s]; !ok {
					if settings[name] == nil {
						settings[name] = make(map[string]any)
					}
					// make sure to remove the setting from the ES as well
					settings[name].(map[string]any)[s] = nil
				}
			}
		}
	}
	return diags
}

func getConfiguredSettings(d *schema.ResourceData) (map[string]any, diag.Diagnostics) {
	var diags diag.Diagnostics
	settings := make(map[string]any)
	if v, ok := d.GetOk("persistent"); ok {
		var ds diag.Diagnostics
		settings["persistent"], ds = expandSettings(v)
		diags = append(diags, ds...)
	}
	if v, ok := d.GetOk("transient"); ok {
		var ds diag.Diagnostics
		settings["transient"], ds = expandSettings(v)
		diags = append(diags, ds...)
	}
	return settings, diags
}

func expandSettings(s any) (map[string]any, diag.Diagnostics) {
	var diags diag.Diagnostics
	settings := s.([]any)[0].(map[string]any)["setting"].(*schema.Set)
	result := make(map[string]any, settings.Len())
	for _, v := range settings.List() {
		setting := v.(map[string]any)
		settingName := setting["name"].(string)
		if _, ok := result[settingName]; ok {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  fmt.Sprintf(`Unable to set "%s".`, settingName),
				Detail:   fmt.Sprintf(`Found setting "%s" have been already configured.`, settingName),
			})
		}

		// check if the setting has value or value_list and act accordingly
		if setting["value"].(string) != "" && len(setting["value_list"].([]any)) > 0 {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  `Only one of "value" or "value_list" can be set.`,
				Detail:   `Only one of "value" or "value_list" can be set.`,
			})
			return nil, diags
		} else if setting["value"].(string) == "" && len(setting["value_list"].([]any)) == 0 {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  `At least one of "value" or "value_list" must be set to not empty value.`,
				Detail:   `At least one of "value" or "value_list" must be set to not empty value.`,
			})
			return nil, diags
		}

		if vv := setting["value"].(string); vv != "" {
			result[settingName] = vv
		} else {
			result[settingName] = setting["value_list"]
		}
	}
	return result, diags
}

func resourceClusterSettingsRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client, diags := clients.NewAPIClientFromSDKResource(d, meta)
	if diags.HasError() {
		return diags
	}
	clusterSettings, diags := elasticsearch.GetSettings(ctx, client)
	if diags.HasError() {
		return diags
	}
	configuredSettings, _ := getConfiguredSettings(d)
	persistent := flattenSettings("persistent", configuredSettings, clusterSettings)
	transient := flattenSettings("transient", configuredSettings, clusterSettings)

	if err := d.Set("persistent", persistent); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("transient", transient); err != nil {
		return diag.FromErr(err)
	}
	return diags
}

func flattenSettings(name string, oldSettings, newSettings map[string]any) []any {
	setting := make(map[string]any)
	settings := make([]any, 0)
	result := make([]any, 0)

	if oldSettings[name] != nil {
		for k := range oldSettings[name].(map[string]any) {
			if newSettings[name] != nil {
				if v, ok := newSettings[name].(map[string]any)[k]; ok {
					s := make(map[string]any)
					s["name"] = k

					// decide which value to set
					switch t := v.(type) {
					case string:
						s["value"] = t
					case []any:
						s["value_list"] = t
					}
					settings = append(settings, s)
				}
			}
		}
	}

	if len(settings) > 0 {
		setting["setting"] = settings
		result = append(result, setting)
	}
	return result
}

func resourceClusterSettingsDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client, diags := clients.NewAPIClientFromSDKResource(d, meta)
	if diags.HasError() {
		return diags
	}

	configuredSettings, _ := getConfiguredSettings(d)
	pSettings := make(map[string]any)
	if v := configuredSettings["persistent"]; v != nil {
		for k := range v.(map[string]any) {
			pSettings[k] = nil
		}
	}
	tSettings := make(map[string]any)
	if v := configuredSettings["transient"]; v != nil {
		for k := range v.(map[string]any) {
			tSettings[k] = nil
		}
	}

	settings := map[string]any{
		"persistent": pSettings,
		"transient":  tSettings,
	}
	if diags := elasticsearch.PutSettings(ctx, client, settings); diags.HasError() {
		return diags
	}

	return diags
}
