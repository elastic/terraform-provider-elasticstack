package cluster

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
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

	utils.AddConnectionSchema(settingsSchema)

	return &schema.Resource{
		Description: "Updates cluster-wide settings. If the Elasticsearch security features are enabled, you must have the manage cluster privilege to use this API. See, https://www.elastic.co/guide/en/elasticsearch/reference/current/cluster-update-settings.html",

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

func resourceClusterSettingsPut(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, diags := clients.NewApiClient(d, meta)
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
			old, new := d.GetChange(v)
			diags = updateRemovedSettings(v, old, new, settings)
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
func updateRemovedSettings(name string, old, new interface{}, settings map[string]interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	if old != nil && new != nil {
		oldSettings := make(map[string]interface{})
		if len(old.([]interface{})) > 0 {
			oldSettings, _ = expandSettings(old)
		}
		newSettings := make(map[string]interface{})
		if len(new.([]interface{})) > 0 {
			newSettings, diags = expandSettings(new)
			if diags.HasError() {
				return diags
			}
		}

		if !utils.MapsEqual(oldSettings, newSettings) {
			for s := range oldSettings {
				if _, ok := newSettings[s]; !ok {
					if settings[name] == nil {
						settings[name] = make(map[string]interface{})
					}
					// make sure to remove the setting from the ES as well
					settings[name].(map[string]interface{})[s] = nil
				}
			}
		}
	}
	return diags
}

func getConfiguredSettings(d *schema.ResourceData) (map[string]interface{}, diag.Diagnostics) {
	var diags diag.Diagnostics
	settings := make(map[string]interface{})
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

func expandSettings(s interface{}) (map[string]interface{}, diag.Diagnostics) {
	var diags diag.Diagnostics
	settings := s.([]interface{})[0].(map[string]interface{})["setting"].(*schema.Set)
	result := make(map[string]interface{}, settings.Len())
	for _, v := range settings.List() {
		setting := v.(map[string]interface{})
		settingName := setting["name"].(string)
		if _, ok := result[settingName]; ok {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  fmt.Sprintf(`Unable to set "%s".`, settingName),
				Detail:   fmt.Sprintf(`Found setting "%s" have been already configured.`, settingName),
			})
		}

		// check if the setting has value or value_list and act accordingly
		if setting["value"].(string) != "" && len(setting["value_list"].([]interface{})) > 0 {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  `Only one of "value" or "value_list" can be set.`,
				Detail:   `Only one of "value" or "value_list" can be set.`,
			})
			return nil, diags
		} else if setting["value"].(string) == "" && len(setting["value_list"].([]interface{})) == 0 {
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

func resourceClusterSettingsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, diags := clients.NewApiClient(d, meta)
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

func flattenSettings(name string, old, new map[string]interface{}) []interface{} {
	setting := make(map[string]interface{})
	settings := make([]interface{}, 0)
	result := make([]interface{}, 0)

	if old[name] != nil {
		for k := range old[name].(map[string]interface{}) {
			if new[name] != nil {
				if v, ok := new[name].(map[string]interface{})[k]; ok {
					s := make(map[string]interface{})
					s["name"] = k

					// decide which value to set
					switch t := v.(type) {
					case string:
						s["value"] = t
					case []interface{}:
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

func resourceClusterSettingsDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, diags := clients.NewApiClient(d, meta)
	if diags.HasError() {
		return diags
	}

	configuredSettings, _ := getConfiguredSettings(d)
	pSettings := make(map[string]interface{})
	if v := configuredSettings["persistent"]; v != nil {
		for k := range v.(map[string]interface{}) {
			pSettings[k] = nil
		}
	}
	tSettings := make(map[string]interface{})
	if v := configuredSettings["transient"]; v != nil {
		for k := range v.(map[string]interface{}) {
			tSettings[k] = nil
		}
	}

	settings := map[string]interface{}{
		"persistent": pSettings,
		"transient":  tSettings,
	}
	if diags := elasticsearch.PutSettings(ctx, client, settings); diags.HasError() {
		return diags
	}

	return diags
}
