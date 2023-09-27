package kibana

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/disaster37/go-kibana-rest/v8/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func ResourceRole() *schema.Resource {
	apikeySchema := map[string]*schema.Schema{
		"name": {
			Description: "The name for the role.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"elasticsearch": {
			Description: "Elasticsearch cluster and index privileges.",
			Type:        schema.TypeSet,
			Required:    true,
			MaxItems:    1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"cluster": {
						Description: "List of the cluster privileges.",
						Type:        schema.TypeSet,
						Optional:    true,
						Elem: &schema.Schema{
							Type: schema.TypeString,
						},
					},
					"indices": {
						Description: "A list of indices permissions entries.",
						Type:        schema.TypeSet,
						Optional:    true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"field_security": {
									Description: "The document fields that the owners of the role have read access to.",
									Type:        schema.TypeList,
									Optional:    true,
									MaxItems:    1,
									Elem: &schema.Resource{
										Schema: map[string]*schema.Schema{
											"grant": {
												Description: "List of the fields to grant the access to.",
												Type:        schema.TypeSet,
												Optional:    true,
												Elem: &schema.Schema{
													Type: schema.TypeString,
												},
											},
											"except": {
												Description: "List of the fields to which the grants will not be applied.",
												Type:        schema.TypeSet,
												Optional:    true,
												Elem: &schema.Schema{
													Type: schema.TypeString,
												},
											},
										},
									},
								},
								"query": {
									Description:      "A search query that defines the documents the owners of the role have read access to.",
									Type:             schema.TypeString,
									ValidateFunc:     validation.StringIsJSON,
									DiffSuppressFunc: utils.DiffJsonSuppress,
									Optional:         true,
								},
								"names": {
									Description: "A list of indices (or index name patterns) to which the permissions in this entry apply.",
									Type:        schema.TypeSet,
									Required:    true,
									Elem: &schema.Schema{
										Type: schema.TypeString,
									},
								},
								"privileges": {
									Description: "The index level privileges that the owners of the role have on the specified indices.",
									Type:        schema.TypeSet,
									Required:    true,
									Elem: &schema.Schema{
										Type: schema.TypeString,
									},
								},
							},
						},
					},
					"run_as": {
						Description: "A list of usernames the owners of this role can impersonate.",
						Type:        schema.TypeSet,
						Optional:    true,
						Elem: &schema.Schema{
							Type: schema.TypeString,
						},
					},
				},
			},
		},
		"kibana": {
			Description: "The list of objects that specify the Kibana privileges for the role.",
			Type:        schema.TypeSet,
			Optional:    true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"base": {
						Description: "A base privilege. When specified, the base must be [\"all\"] or [\"read\"].",
						Type:        schema.TypeSet,
						Optional:    true,
						MaxItems:    1,
						Elem: &schema.Schema{
							Type: schema.TypeString,
							ValidateDiagFunc: func(v any, p cty.Path) diag.Diagnostics {
								value := v.(string)
								expected := []string{"all", "read"}
								var diags diag.Diagnostics
								for _, e := range expected {
									if e == value {
										return diags
									}
								}
								diag := diag.Diagnostic{
									Severity: diag.Error,
									Summary:  "Wrong value for base attribute",
									Detail:   fmt.Sprintf("Expected %s , got %s", strings.Join(expected, " | "), value),
								}
								diags = append(diags, diag)
								return diags
							},
						},
					},
					"feature": {
						Description: "List of privileges for specific features. When the feature privileges are specified, you are unable to use the \"base\" section.",
						Type:        schema.TypeSet,
						Optional:    true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"name": {
									Description: "Feature name.",
									Type:        schema.TypeString,
									Required:    true,
								},
								"privileges": {
									Description: "Feature privileges.",
									Type:        schema.TypeSet,
									Required:    true,
									Elem: &schema.Schema{
										Type: schema.TypeString,
									},
								},
							},
						},
					},
					"spaces": {
						Description: "The spaces to apply the privileges to. To grant access to all spaces, set to [\"*\"], or omit the value.",
						Type:        schema.TypeSet,
						Required:    true,
						Elem: &schema.Schema{
							Type: schema.TypeString,
						},
					},
				},
			},
		},
		"metadata": {
			Description:      "Optional meta-data.",
			Type:             schema.TypeString,
			Optional:         true,
			Computed:         true,
			ValidateFunc:     validation.StringIsJSON,
			DiffSuppressFunc: utils.DiffJsonSuppress,
		},
	}

	return &schema.Resource{
		Description: "Creates a Kibana role. See, https://www.elastic.co/guide/en/kibana/master/role-management-api-put.html",

		CreateContext: resourceRoleUpsert,
		UpdateContext: resourceRoleUpsert,
		ReadContext:   resourceRoleRead,
		DeleteContext: resourceRoleDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: apikeySchema,
	}
}

func resourceRoleUpsert(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, diags := clients.NewApiClient(d, meta)
	if diags.HasError() {
		return diags
	}

	kibana, err := client.GetKibanaClient()
	if err != nil {
		return diag.FromErr(err)
	}
	queryParams := ""
	if d.IsNewResource() {
		queryParams = "?createOnly=true"
	}
	kibanaRole := kbapi.KibanaRole{
		Name:          fmt.Sprintf("%s%s", d.Get("name").(string), queryParams),
		Kibana:        []kbapi.KibanaRoleKibana{},
		Elasticsearch: &kbapi.KibanaRoleElasticsearch{},
	}

	if v, ok := d.GetOk("kibana"); ok {
		definedKibanaConfigs := v.(*schema.Set)
		kibanaRole.Kibana = make([]kbapi.KibanaRoleKibana, definedKibanaConfigs.Len())
		for i, item := range definedKibanaConfigs.List() {
			each := item.(map[string]interface{})
			_config := kbapi.KibanaRoleKibana{
				Base:    []string{},
				Feature: map[string][]string{},
			}

			if basePrivileges, ok := each["base"].(*schema.Set); ok && basePrivileges.Len() > 0 {
				if _features, ok := each["feature"].(*schema.Set); ok && _features.Len() > 0 {
					return diag.Errorf("Only one of the `feature` or `base` privileges allowed!")
				}
				_config.Base = make([]string, basePrivileges.Len())
				for i, name := range basePrivileges.List() {
					_config.Base[i] = name.(string)
				}
			} else if kibanaFeatures, ok := each["feature"].(*schema.Set); ok && kibanaFeatures.Len() > 0 {
				for _, item := range kibanaFeatures.List() {
					featureData := item.(map[string]interface{})
					featurePrivileges := featureData["privileges"].(*schema.Set)
					_features := make([]string, featurePrivileges.Len())
					for i, f := range featurePrivileges.List() {
						_features[i] = f.(string)
					}
					_config.Feature[featureData["name"].(string)] = _features
				}
			} else {
				return diag.Errorf("Either on of the `feature` or `base` privileges must be set for kibana role!")
			}

			if roleSpaces, ok := each["spaces"].(*schema.Set); ok && roleSpaces.Len() > 0 {
				_config.Spaces = make([]string, roleSpaces.Len())
				for i, name := range roleSpaces.List() {
					_config.Spaces[i] = name.(string)
				}
			}
			kibanaRole.Kibana[i] = _config
		}
	}

	if v, ok := d.GetOk("elasticsearch"); ok {
		if definedElasicConfigs := v.(*schema.Set); definedElasicConfigs.Len() > 0 {
			userElasitcConfig := definedElasicConfigs.List()[0].(map[string]interface{})
			if v, ok := userElasitcConfig["cluster"]; ok {
				definedCluster := v.(*schema.Set)
				cls := make([]string, definedCluster.Len())
				for i, cl := range definedCluster.List() {
					cls[i] = cl.(string)
				}
				kibanaRole.Elasticsearch.Cluster = cls

				if v, ok := userElasitcConfig["indices"]; ok {
					definedIndices := v.(*schema.Set)
					indices := make([]kbapi.KibanaRoleElasticsearchIndice, definedIndices.Len())
					for i, idx := range definedIndices.List() {
						index := idx.(map[string]interface{})

						definedNames := index["names"].(*schema.Set)
						names := make([]string, definedNames.Len())
						for i, name := range definedNames.List() {
							names[i] = name.(string)
						}
						definedPrivileges := index["privileges"].(*schema.Set)
						privileges := make([]string, definedPrivileges.Len())
						for i, pr := range definedPrivileges.List() {
							privileges[i] = pr.(string)
						}

						newIndex := kbapi.KibanaRoleElasticsearchIndice{
							Names:      names,
							Privileges: privileges,
						}

						if query := index["query"].(string); query != "" {
							newIndex.Query = &query
						}
						if fieldSec := index["field_security"].([]interface{}); len(fieldSec) > 0 {
							fieldSecurity := map[string]interface{}{}
							// there must be only 1 entry
							definedFieldSec := fieldSec[0].(map[string]interface{})

							// grants
							if gr := definedFieldSec["grant"].(*schema.Set); gr != nil {
								grants := make([]string, gr.Len())
								for i, grant := range gr.List() {
									grants[i] = grant.(string)
								}
								fieldSecurity["grant"] = grants
							}
							// except
							if exp := definedFieldSec["except"].(*schema.Set); exp != nil {
								excepts := make([]string, exp.Len())
								for i, except := range exp.List() {
									excepts[i] = except.(string)
								}
								fieldSecurity["except"] = excepts
							}
							newIndex.FieldSecurity = fieldSecurity
						}

						indices[i] = newIndex
					}
					kibanaRole.Elasticsearch.Indices = indices
				}

				if v, ok := userElasitcConfig["run_as"]; ok {
					definedRuns := v.(*schema.Set)
					runs := make([]string, definedRuns.Len())
					for i, run := range definedRuns.List() {
						runs[i] = run.(string)
					}
					kibanaRole.Elasticsearch.RunAs = runs
				}
			}
		}
	}

	if v, ok := d.GetOk("metadata"); ok {
		metadata := make(map[string]interface{})
		if err := json.NewDecoder(strings.NewReader(v.(string))).Decode(&metadata); err != nil {
			return diag.FromErr(err)
		}
		kibanaRole.Metadata = metadata
	}

	roleManageResponse, err := kibana.KibanaRoleManagement.CreateOrUpdate(&kibanaRole)
	if err != nil {
		return diag.FromErr(err)
	}

	id, diags := client.ID(ctx, roleManageResponse.Name)
	if diags.HasError() {
		return diags
	}

	d.SetId(id.String())
	return resourceRoleRead(ctx, d, meta)
}

func resourceRoleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, diags := clients.NewApiClient(d, meta)
	if diags.HasError() {
		return diags
	}
	compId, diags := clients.CompositeIdFromStr(d.Id())
	if diags.HasError() {
		return diags
	}
	name := compId.ResourceId

	kibana, err := client.GetKibanaClient()
	if err != nil {
		return diag.FromErr(err)
	}

	role, err := kibana.KibanaRoleManagement.Get(name)
	if role == nil && err == nil {
		d.SetId("")
		return diags
	}
	if err != nil {
		return diag.FromErr(err)
	}

	// set the fields
	if err := d.Set("name", role.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("elasticsearch", flattenElasticsearchData(role.Elasticsearch)); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("kibana", flattenKibanaData(&role.Kibana)); err != nil {
		return diag.FromErr(err)
	}
	if role.Metadata != nil {
		metadata, err := json.Marshal(role.Metadata)
		if err != nil {
			return diag.FromErr(err)
		}
		if err := d.Set("metadata", string(metadata)); err != nil {
			return diag.FromErr(err)
		}
	}
	return diags
}

func resourceRoleDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, diags := clients.NewApiClient(d, meta)
	if diags.HasError() {
		return diags
	}
	compId, diags := clients.CompositeIdFromStr(d.Id())
	if diags.HasError() {
		return diags
	}

	kibana, err := client.GetKibanaClient()
	if err != nil {
		return diag.FromErr(err)
	}

	err = kibana.KibanaRoleManagement.Delete(compId.ResourceId)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")
	return diags
}

// Helper functions

func flattenIndicesData(indices *[]kbapi.KibanaRoleElasticsearchIndice) []interface{} {
	if indices != nil {
		oindx := make([]interface{}, len(*indices))

		for i, index := range *indices {
			oi := make(map[string]interface{})
			oi["names"] = index.Names
			oi["privileges"] = index.Privileges
			oi["query"] = index.Query

			if index.FieldSecurity != nil {
				fsec := make(map[string]interface{})
				if grant_v, ok := index.FieldSecurity["grant"]; ok {
					fsec["grant"] = grant_v
				}
				if except_v, ok := index.FieldSecurity["except"]; ok {
					fsec["except"] = except_v
				}
				oi["field_security"] = []interface{}{fsec}
			}
			oindx[i] = oi
		}
		return oindx
	}
	return make([]interface{}, 0)
}

func flattenElasticsearchData(elastic *kbapi.KibanaRoleElasticsearch) []interface{} {
	if elastic != nil {
		result := make(map[string]interface{})
		if len(elastic.Cluster) > 0 {
			result["cluster"] = elastic.Cluster
		}
		result["indices"] = flattenIndicesData(&elastic.Indices)
		if len(elastic.RunAs) > 0 {
			result["run_as"] = elastic.RunAs
		}
		return []interface{}{result}
	}
	return make([]interface{}, 0)
}

func flattenKibanaFeatureData(features map[string][]string) []interface{} {
	if features != nil {
		result := make([]interface{}, len(features))
		i := 0
		for k, v := range features {
			m := make(map[string]interface{})
			m["name"] = k
			m["privileges"] = v
			result[i] = m
			i += 1
		}
		return result
	}
	return make([]interface{}, 0)
}

func flattenKibanaData(kibana_configs *[]kbapi.KibanaRoleKibana) []interface{} {
	if kibana_configs != nil {
		result := make([]interface{}, len(*kibana_configs))
		for i, index := range *kibana_configs {
			nk := make(map[string]interface{})
			nk["base"] = index.Base
			nk["feature"] = flattenKibanaFeatureData(index.Feature)
			nk["spaces"] = index.Spaces
			result[i] = nk
		}
		return result
	}
	return make([]interface{}, 0)
}
