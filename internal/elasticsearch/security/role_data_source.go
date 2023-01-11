package security

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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
	client, diags := clients.NewApiClient(d, meta)
	if diags.HasError() {
		return diags
	}

	roleId := d.Get("name").(string)
	id, diags := client.ID(ctx, roleId)
	if diags.HasError() {
		return diags
	}
	d.SetId(id.String())

	return resourceSecurityRoleRead(ctx, d, meta)
}
