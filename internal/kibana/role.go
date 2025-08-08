package kibana

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/disaster37/go-kibana-rest/v8/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var (
	minSupportedRemoteIndicesVersion = version.Must(version.NewVersion("8.10.0"))
	minSupportedDescriptionVersion   = version.Must(version.NewVersion("8.15.0"))
)

func ResourceRole() *schema.Resource {
	roleSchema := map[string]*schema.Schema{
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
					"remote_indices": {
						Description: "A list of remote indices permissions entries. Remote indices are effective for remote clusters configured with the API key based model. They have no effect for remote clusters configured with the certificate based model.",
						Type:        schema.TypeSet,
						Optional:    true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"clusters": {
									Description: "A list of cluster aliases to which the permissions in this entry apply.",
									Type:        schema.TypeSet,
									Required:    true,
									Elem: &schema.Schema{
										Type: schema.TypeString,
									},
								},
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
						Description: "A base privilege. When specified, the base must be [\"all\"] or [\"read\"]. When the base privileges are specified, you are unable to use the \"feature\" section.",
						Type:        schema.TypeSet,
						Optional:    true,
						MaxItems:    1,
						Elem: &schema.Schema{
							Type:         schema.TypeString,
							ValidateFunc: validation.StringInSlice([]string{"all", "read"}, true),
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
		"description": {
			Description: "Optional description for the role",
			Type:        schema.TypeString,
			Optional:    true,
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

		Schema: roleSchema,
	}
}

func resourceRoleUpsert(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, diags := clients.NewApiClientFromSDKResource(d, meta)
	if diags.HasError() {
		return diags
	}

	serverVersion, diags := client.ServerVersion(ctx)
	if diags.HasError() {
		return diags
	}

	kibana, err := client.GetKibanaClient()
	if err != nil {
		return diag.FromErr(err)
	}
	kibanaRole := kbapi.KibanaRole{
		Name:          d.Get("name").(string),
		Kibana:        []kbapi.KibanaRoleKibana{},
		Elasticsearch: &kbapi.KibanaRoleElasticsearch{},
		CreateOnly:    d.IsNewResource(),
	}

	if v, ok := d.GetOk("kibana"); ok {
		kibanaRole.Kibana, diags = expandKibanaRoleKibana(v)
		if diags != nil {
			return diags
		}
	}

	if v, ok := d.GetOk("elasticsearch"); ok {
		kibanaRole.Elasticsearch, diags = expandKibanaRoleElasticsearch(v, serverVersion)
		if diags != nil {
			return diags
		}
	}

	if v, ok := d.GetOk("metadata"); ok {
		kibanaRole.Metadata, diags = expandKibanaRoleMetadata(v)
		if diags != nil {
			return diags
		}
	}

	if v, ok := d.GetOk("description"); ok {
		if serverVersion.LessThan(minSupportedDescriptionVersion) {
			return diag.FromErr(fmt.Errorf("'description' is supported only for Kibana v%s and above", minSupportedDescriptionVersion.String()))
		}

		kibanaRole.Description = v.(string)
	}

	roleManageResponse, err := kibana.KibanaRoleManagement.CreateOrUpdate(&kibanaRole)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(roleManageResponse.Name)
	return resourceRoleRead(ctx, d, meta)
}

func resourceRoleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, diags := clients.NewApiClientFromSDKResource(d, meta)
	if diags.HasError() {
		return diags
	}

	name := d.Id()

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
	if err := d.Set("elasticsearch", flattenKibanaRoleElasticsearchData(role.Elasticsearch)); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("kibana", flattenKibanaRoleKibanaData(&role.Kibana)); err != nil {
		return diag.FromErr(err)
	}
	// Only set description if it's not empty to avoid Terraform validation errors
	if strings.TrimSpace(role.Description) != "" {
		if err := d.Set("description", role.Description); err != nil {
			return diag.FromErr(err)
		}
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
	client, diags := clients.NewApiClientFromSDKResource(d, meta)
	if diags.HasError() {
		return diags
	}
	resourceId := d.Id()

	kibana, err := client.GetKibanaClient()
	if err != nil {
		return diag.FromErr(err)
	}

	err = kibana.KibanaRoleManagement.Delete(resourceId)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")
	return diags
}

// Helper functions

func expandKibanaRoleMetadata(v interface{}) (map[string]interface{}, diag.Diagnostics) {
	metadata := make(map[string]interface{})
	if err := json.NewDecoder(strings.NewReader(v.(string))).Decode(&metadata); err != nil {
		return nil, diag.FromErr(err)
	}
	return metadata, nil
}

func expandKibanaRoleElasticsearch(v interface{}, serverVersion *version.Version) (*kbapi.KibanaRoleElasticsearch, diag.Diagnostics) {
	elasticConfig := &kbapi.KibanaRoleElasticsearch{}
	var diags diag.Diagnostics

	if definedElasticConfigs := v.(*schema.Set); definedElasticConfigs.Len() > 0 {
		userElasticConfig := definedElasticConfigs.List()[0].(map[string]interface{})
		if v, ok := userElasticConfig["cluster"]; ok {
			definedCluster := v.(*schema.Set)
			cls := make([]string, definedCluster.Len())
			for i, cl := range definedCluster.List() {
				cls[i] = cl.(string)
			}
			elasticConfig.Cluster = cls

			if v, ok := userElasticConfig["indices"]; ok {
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
				elasticConfig.Indices = indices
			}

			if v, ok := userElasticConfig["remote_indices"]; ok {
				definedRemoteIndices := v.(*schema.Set)
				if definedRemoteIndices.Len() > 0 {
					if serverVersion.LessThan(minSupportedRemoteIndicesVersion) {
						return nil, diag.FromErr(fmt.Errorf("'remote_indices' is supported only for Kibana v%s and above", minSupportedRemoteIndicesVersion.String()))
					}
				}
				remote_indices := make([]kbapi.KibanaRoleElasticsearchRemoteIndice, definedRemoteIndices.Len())
				for i, idx := range definedRemoteIndices.List() {
					index := idx.(map[string]interface{})

					definedNames := index["names"].(*schema.Set)
					names := make([]string, definedNames.Len())
					for i, name := range definedNames.List() {
						names[i] = name.(string)
					}
					definedClusters := index["clusters"].(*schema.Set)
					clusters := make([]string, definedClusters.Len())
					for i, cluster := range definedClusters.List() {
						clusters[i] = cluster.(string)
					}
					definedPrivileges := index["privileges"].(*schema.Set)
					privileges := make([]string, definedPrivileges.Len())
					for i, pr := range definedPrivileges.List() {
						privileges[i] = pr.(string)
					}

					newRemoteIndex := kbapi.KibanaRoleElasticsearchRemoteIndice{
						Names:      names,
						Clusters:   clusters,
						Privileges: privileges,
					}

					if query := index["query"].(string); query != "" {
						newRemoteIndex.Query = &query
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
						newRemoteIndex.FieldSecurity = fieldSecurity
					}

					remote_indices[i] = newRemoteIndex
				}
				elasticConfig.RemoteIndices = remote_indices
			}

			if v, ok := userElasticConfig["run_as"]; ok {
				definedRuns := v.(*schema.Set)
				runs := make([]string, definedRuns.Len())
				for i, run := range definedRuns.List() {
					runs[i] = run.(string)
				}
				elasticConfig.RunAs = runs
			}
		}
	}
	return elasticConfig, diags
}

func expandKibanaRoleKibana(v interface{}) ([]kbapi.KibanaRoleKibana, diag.Diagnostics) {
	kibanaConfigs := []kbapi.KibanaRoleKibana{}
	definedKibanaConfigs := v.(*schema.Set)

	for _, item := range definedKibanaConfigs.List() {
		each := item.(map[string]interface{})
		config := kbapi.KibanaRoleKibana{
			Base:    []string{},
			Feature: map[string][]string{},
		}

		if basePrivileges, ok := each["base"].(*schema.Set); ok && basePrivileges.Len() > 0 {
			if _features, ok := each["feature"].(*schema.Set); ok && _features.Len() > 0 {
				return nil, diag.Errorf("Only one of the `feature` or `base` privileges allowed!")
			}
			config.Base = make([]string, basePrivileges.Len())
			for i, name := range basePrivileges.List() {
				config.Base[i] = name.(string)
			}
		} else if kibanaFeatures, ok := each["feature"].(*schema.Set); ok && kibanaFeatures.Len() > 0 {
			for _, item := range kibanaFeatures.List() {
				featureData := item.(map[string]interface{})
				featurePrivileges := featureData["privileges"].(*schema.Set)
				_features := make([]string, featurePrivileges.Len())
				for i, f := range featurePrivileges.List() {
					_features[i] = f.(string)
				}
				config.Feature[featureData["name"].(string)] = _features
			}
		} else {
			return nil, diag.Errorf("Either on of the `feature` or `base` privileges must be set for kibana role!")
		}

		if roleSpaces, ok := each["spaces"].(*schema.Set); ok && roleSpaces.Len() > 0 {
			config.Spaces = make([]string, roleSpaces.Len())
			for i, name := range roleSpaces.List() {
				config.Spaces[i] = name.(string)
			}
		}
		kibanaConfigs = append(kibanaConfigs, config)
	}
	return kibanaConfigs, nil
}

func flattenKibanaRoleIndicesData(indices []kbapi.KibanaRoleElasticsearchIndice) []interface{} {
	oindx := make([]interface{}, len(indices))

	for i, index := range indices {
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

func flattenKibanaRoleRemoteIndicesData(indices []kbapi.KibanaRoleElasticsearchRemoteIndice) []interface{} {
	oindx := make([]interface{}, len(indices))

	for i, index := range indices {
		oi := make(map[string]interface{})
		oi["clusters"] = index.Clusters
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

func flattenKibanaRoleElasticsearchData(elastic *kbapi.KibanaRoleElasticsearch) []interface{} {
	if elastic != nil {
		result := make(map[string]interface{})
		if len(elastic.Cluster) > 0 {
			result["cluster"] = elastic.Cluster
		}
		result["indices"] = flattenKibanaRoleIndicesData(elastic.Indices)
		result["remote_indices"] = flattenKibanaRoleRemoteIndicesData(elastic.RemoteIndices)
		if len(elastic.RunAs) > 0 {
			result["run_as"] = elastic.RunAs
		}
		return []interface{}{result}
	}
	return make([]interface{}, 0)
}

func flattenKibanaRoleKibanaFeatureData(features map[string][]string) []interface{} {
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

func flattenKibanaRoleKibanaData(kibana_configs *[]kbapi.KibanaRoleKibana) []interface{} {
	if kibana_configs != nil {
		result := make([]interface{}, len(*kibana_configs))
		for i, index := range *kibana_configs {
			nk := make(map[string]interface{})
			nk["base"] = index.Base
			nk["feature"] = flattenKibanaRoleKibanaFeatureData(index.Feature)
			nk["spaces"] = index.Spaces
			result[i] = nk
		}
		return result
	}
	return make([]interface{}, 0)
}
