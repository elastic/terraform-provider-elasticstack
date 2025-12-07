package security

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var (
	MinSupportedDescriptionVersion = version.Must(version.NewVersion("8.15.0"))
)

func DataSourceRole() *schema.Resource {
	roleSchema := map[string]*schema.Schema{
		"id": {
			Description: "Internal identifier of the resource",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"name": {
			Description: "The name of the role.",
			Type:        schema.TypeString,
			Required:    true,
		},
		"description": {
			Description: "The description of the role.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"applications": {
			Description: "A list of application privilege entries.",
			Type:        schema.TypeSet,
			Computed:    true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"application": {
						Description: "The name of the application to which this entry applies.",
						Type:        schema.TypeString,
						Computed:    true,
					},
					"privileges": {
						Description: "A list of strings, where each element is the name of an application privilege or action.",
						Type:        schema.TypeSet,
						Elem: &schema.Schema{
							Type: schema.TypeString,
						},
						Computed: true,
					},
					"resources": {
						Description: "A list resources to which the privileges are applied.",
						Type:        schema.TypeSet,
						Elem: &schema.Schema{
							Type: schema.TypeString,
						},
						Computed: true,
					},
				},
			},
		},
		"global": {
			Description: "An object defining global privileges.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"cluster": {
			Description: "A list of cluster privileges. These privileges define the cluster level actions that users with this role are able to execute.",
			Type:        schema.TypeSet,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
			Computed: true,
		},
		"indices": {
			Description: "A list of indices permissions entries.",
			Type:        schema.TypeSet,
			Computed:    true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"field_security": {
						Description: "The document fields that the owners of the role have read access to.",
						Type:        schema.TypeList,
						Computed:    true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"grant": {
									Description: "List of the fields to grant the access to.",
									Type:        schema.TypeSet,
									Computed:    true,
									Elem: &schema.Schema{
										Type: schema.TypeString,
									},
								},
								"except": {
									Description: "List of the fields to which the grants will not be applied.",
									Type:        schema.TypeSet,
									Computed:    true,
									Elem: &schema.Schema{
										Type: schema.TypeString,
									},
								},
							},
						},
					},
					"names": {
						Description: "A list of indices (or index name patterns) to which the permissions in this entry apply.",
						Type:        schema.TypeSet,
						Computed:    true,
						Elem: &schema.Schema{
							Type: schema.TypeString,
						},
					},
					"privileges": {
						Description: "The index level privileges that the owners of the role have on the specified indices.",
						Type:        schema.TypeSet,
						Computed:    true,
						Elem: &schema.Schema{
							Type: schema.TypeString,
						},
					},
					"query": {
						Description: "A search query that defines the documents the owners of the role have read access to.",
						Type:        schema.TypeString,
						Computed:    true,
					},
					"allow_restricted_indices": {
						Description: "Include matching restricted indices in names parameter. Usage is strongly discouraged as it can grant unrestricted operations on critical data, make the entire system unstable or leak sensitive information.",
						Type:        schema.TypeBool,
						Computed:    true,
					},
				},
			},
		},
		"remote_indices": {
			Description: "A list of remote indices permissions entries. Remote indices are effective for remote clusters configured with the API key based model. They have no effect for remote clusters configured with the certificate based model.",
			Type:        schema.TypeSet,
			Computed:    true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"clusters": {
						Description: "A list of cluster aliases to which the permissions in this entry apply.",
						Type:        schema.TypeSet,
						Computed:    true,
						Elem: &schema.Schema{
							Type: schema.TypeString,
						},
					},
					"field_security": {
						Description: "The document fields that the owners of the role have read access to.",
						Type:        schema.TypeList,
						Computed:    true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"grant": {
									Description: "List of the fields to grant the access to.",
									Type:        schema.TypeSet,
									Computed:    true,
									Elem: &schema.Schema{
										Type: schema.TypeString,
									},
								},
								"except": {
									Description: "List of the fields to which the grants will not be applied.",
									Type:        schema.TypeSet,
									Computed:    true,
									Elem: &schema.Schema{
										Type: schema.TypeString,
									},
								},
							},
						},
					},
					"names": {
						Description: "A list of indices (or index name patterns) to which the permissions in this entry apply.",
						Type:        schema.TypeSet,
						Computed:    true,
						Elem: &schema.Schema{
							Type: schema.TypeString,
						},
					},
					"privileges": {
						Description: "The index level privileges that the owners of the role have on the specified indices.",
						Type:        schema.TypeSet,
						Computed:    true,
						Elem: &schema.Schema{
							Type: schema.TypeString,
						},
					},
					"query": {
						Description: "A search query that defines the documents the owners of the role have read access to.",
						Type:        schema.TypeString,
						Computed:    true,
					},
				},
			},
		},
		"metadata": {
			Description: "Optional meta-data.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"run_as": {
			Description: "A list of users that the owners of this role can impersonate.",
			Type:        schema.TypeSet,
			Optional:    true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
	}

	utils.AddConnectionSchema(roleSchema)

	return &schema.Resource{
		Description: "Retrieves roles in the native realm. See, https://www.elastic.co/guide/en/elasticsearch/reference/current/security-api-get-role.html",
		ReadContext: dataSourceSecurityRoleRead,
		Schema:      roleSchema,
	}
}

func dataSourceSecurityRoleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, diags := clients.NewApiClientFromSDKResource(d, meta)
	if diags.HasError() {
		return diags
	}

	roleId := d.Get("name").(string)
	id, diags := client.ID(ctx, roleId)
	if diags.HasError() {
		return diags
	}
	d.SetId(id.String())

	role, diags := elasticsearch.GetRole(ctx, client, roleId)
	if role == nil && diags == nil {
		tflog.Warn(ctx, fmt.Sprintf(`Role "%s" not found, removing from state`, roleId))
		d.SetId("")
		return diags
	}
	if diags.HasError() {
		return diags
	}

	// set the fields
	if err := d.Set("name", roleId); err != nil {
		return diag.FromErr(err)
	}

	// Set the description if it exists
	if role.Description != nil {
		if err := d.Set("description", *role.Description); err != nil {
			return diag.FromErr(err)
		}
	}

	apps := role.Applications
	applications := flattenApplicationsData(&apps)
	if err := d.Set("applications", applications); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("cluster", role.Cluster); err != nil {
		return diag.FromErr(err)
	}

	if role.Global != nil {
		global, err := json.Marshal(role.Global)
		if err != nil {
			return diag.FromErr(err)
		}
		if err := d.Set("global", string(global)); err != nil {
			return diag.FromErr(err)
		}
	}

	indices := flattenIndicesData(role.Indices)
	if err := d.Set("indices", indices); err != nil {
		return diag.FromErr(err)
	}
	remoteIndices := flattenRemoteIndicesData(role.RemoteIndices)
	if err := d.Set("remote_indices", remoteIndices); err != nil {
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

	if err := d.Set("run_as", role.RunAs); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func flattenApplicationsData(apps *[]models.Application) []interface{} {
	if apps != nil {
		oapps := make([]interface{}, len(*apps))
		for i, app := range *apps {
			oa := make(map[string]interface{})
			oa["application"] = app.Name
			oa["privileges"] = app.Privileges
			oa["resources"] = app.Resources
			oapps[i] = oa
		}
		return oapps
	}
	return make([]interface{}, 0)
}

func flattenIndicesData(indices []models.IndexPerms) []interface{} {
	oindx := make([]interface{}, len(indices))

	for i, index := range indices {
		oi := make(map[string]interface{})
		oi["names"] = index.Names
		oi["privileges"] = index.Privileges
		oi["query"] = index.Query
		oi["allow_restricted_indices"] = index.AllowRestrictedIndices

		if index.FieldSecurity != nil {
			fsec := make(map[string]interface{})
			fsec["grant"] = index.FieldSecurity.Grant
			fsec["except"] = index.FieldSecurity.Except
			oi["field_security"] = []interface{}{fsec}
		}
		oindx[i] = oi
	}
	return oindx
}

func flattenRemoteIndicesData(remoteIndices []models.RemoteIndexPerms) []interface{} {
	oRemoteIndx := make([]interface{}, len(remoteIndices))

	for i, remoteIndex := range remoteIndices {
		oi := make(map[string]interface{})
		oi["names"] = remoteIndex.Names
		oi["clusters"] = remoteIndex.Clusters
		oi["privileges"] = remoteIndex.Privileges
		oi["query"] = remoteIndex.Query

		if remoteIndex.FieldSecurity != nil {
			fsec := make(map[string]interface{})
			fsec["grant"] = remoteIndex.FieldSecurity.Grant
			fsec["except"] = remoteIndex.FieldSecurity.Except
			oi["field_security"] = []interface{}{fsec}
		}
		oRemoteIndx[i] = oi
	}
	return oRemoteIndx
}
