package kibana

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func DataSourceRole() *schema.Resource {
	roleSchema := map[string]*schema.Schema{
		"name": {
			Description: "The name for the role.",
			Type:        schema.TypeString,
			Required:    true,
		},
		"elasticsearch": {
			Description: "Elasticsearch cluster and index privileges.",
			Type:        schema.TypeSet,
			Computed:    true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"cluster": {
						Description: "List of the cluster privileges.",
						Type:        schema.TypeSet,
						Computed:    true,
						Elem: &schema.Schema{
							Type: schema.TypeString,
						},
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
								"query": {
									Description: "A search query that defines the documents the owners of the role have read access to.",
									Type:        schema.TypeString,
									Computed:    true,
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
						Computed:    true,
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
			Computed:    true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"base": {
						Description: "A base privilege. When specified, the base must be [\"all\"] or [\"read\"].",
						Type:        schema.TypeSet,
						Computed:    true,
						Elem: &schema.Schema{
							Type: schema.TypeString,
						},
					},
					"feature": {
						Description: "List of privileges for specific features. When the feature privileges are specified, you are unable to use the \"base\" section.",
						Type:        schema.TypeSet,
						Computed:    true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"name": {
									Description: "Feature name.",
									Type:        schema.TypeString,
									Computed:    true,
								},
								"privileges": {
									Description: "Feature privileges.",
									Type:        schema.TypeSet,
									Computed:    true,
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
						Computed:    true,
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
		Description: "Retrieve a specific role. See, https://www.elastic.co/guide/en/kibana/current/role-management-specific-api-get.html",
		ReadContext: dataSourceSecurityRoleRead,
		Schema:      roleSchema,
	}
}

func dataSourceSecurityRoleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	roleId := d.Get("name").(string)
	d.SetId(roleId)

	return resourceRoleRead(ctx, d, meta)
}
