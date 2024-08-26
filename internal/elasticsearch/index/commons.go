package index

import (
	"encoding/json"

	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ExpandIndexAliases(definedAliases *schema.Set) (map[string]models.IndexAlias, diag.Diagnostics) {
	var diags diag.Diagnostics
	aliases := make(map[string]models.IndexAlias, definedAliases.Len())

	for _, a := range definedAliases.List() {
		alias := a.(map[string]interface{})
		ia, diags := ExpandIndexAlias(alias)
		if diags.HasError() {
			return nil, diags
		}
		aliases[alias["name"].(string)] = *ia
	}
	return aliases, diags
}

func ExpandIndexAlias(alias map[string]interface{}) (*models.IndexAlias, diag.Diagnostics) {
	var diags diag.Diagnostics
	ia := models.IndexAlias{}
	ia.Name = alias["name"].(string)

	if f, ok := alias["filter"]; ok {
		if f.(string) != "" {
			filterMap := make(map[string]interface{})
			if err := json.Unmarshal([]byte(f.(string)), &filterMap); err != nil {
				return nil, diag.FromErr(err)
			}
			ia.Filter = filterMap
		}
	}
	ia.IndexRouting = alias["index_routing"].(string)
	ia.IsHidden = alias["is_hidden"].(bool)
	ia.IsWriteIndex = alias["is_write_index"].(bool)
	ia.Routing = alias["routing"].(string)
	ia.SearchRouting = alias["search_routing"].(string)
	return &ia, diags
}

func FlattenIndexAliases(aliases map[string]models.IndexAlias) (interface{}, diag.Diagnostics) {
	var diags diag.Diagnostics

	als := make([]interface{}, 0)
	for k, v := range aliases {
		alias, diags := FlattenIndexAlias(k, v)
		if diags.HasError() {
			return nil, diags
		}
		als = append(als, alias)
	}
	return als, diags
}

func FlattenIndexAlias(name string, alias models.IndexAlias) (interface{}, diag.Diagnostics) {
	var diags diag.Diagnostics
	a := make(map[string]interface{})
	a["name"] = name
	if alias.Filter != nil {
		f, err := json.Marshal(alias.Filter)
		if err != nil {
			return nil, diag.FromErr(err)
		}
		a["filter"] = string(f)
	} else {
		a["filter"] = ""
	}
	a["index_routing"] = alias.IndexRouting
	a["is_hidden"] = alias.IsHidden
	a["is_write_index"] = alias.IsWriteIndex
	a["routing"] = alias.Routing
	a["search_routing"] = alias.SearchRouting

	return a, diags
}

func ExpandLifecycle(definedLifecycle *schema.Set) *models.LifecycleSettings {
	if definedLifecycle.Len() == 0 {
		return nil
	}
	lifecycleMap := definedLifecycle.List()[0].(map[string]interface{})
	if lifecycleMap != nil {
		lifecycle := &models.LifecycleSettings{}
		if s, ok := lifecycleMap["data_retention"]; ok {
			lifecycle.DataRetention = s.(string)
		}
		return lifecycle
	}
	return nil
}

func FlattenLifecycle(lifecycle *models.LifecycleSettings) interface{} {
	lf := make([]interface{}, 1)
	lfSettings := make(map[string]interface{})
	lfSettings["data_retention"] = lifecycle.DataRetention

	lf[0] = lfSettings

	return lf
}
