package role

import (
	"context"
	_ "embed"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	providerschema "github.com/elastic/terraform-provider-elasticstack/internal/schema"
)

//go:embed resource-description.md
var roleResourceDescription string

func (r *roleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = GetSchema()
}

func GetSchema() schema.Schema {
	return schema.Schema{
		MarkdownDescription: roleResourceDescription,
		Blocks: map[string]schema.Block{
			"elasticsearch_connection": providerschema.GetEsFWConnectionBlock("elasticsearch_connection", false),
		},
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Internal identifier of the resource",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the role.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "The description of the role.",
				Optional:            true,
			},
			"applications": schema.SetNestedAttribute{
				MarkdownDescription: "A list of application privilege entries.",
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"application": schema.StringAttribute{
							MarkdownDescription: "The name of the application to which this entry applies.",
							Required:            true,
						},
						"privileges": schema.SetAttribute{
							MarkdownDescription: "A list of strings, where each element is the name of an application privilege or action.",
							Required:            true,
							ElementType:         types.StringType,
						},
						"resources": schema.SetAttribute{
							MarkdownDescription: "A list resources to which the privileges are applied.",
							Required:            true,
							ElementType:         types.StringType,
						},
					},
				},
			},
			"global": schema.StringAttribute{
				MarkdownDescription: "An object defining global privileges.",
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"cluster": schema.SetAttribute{
				MarkdownDescription: "A list of cluster privileges. These privileges define the cluster level actions that users with this role are able to execute.",
				Optional:            true,
				ElementType:         types.StringType,
			},
			"indices": schema.SetNestedAttribute{
				MarkdownDescription: "A list of indices permissions entries.",
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"field_security": schema.ListNestedAttribute{
							MarkdownDescription: "The document fields that the owners of the role have read access to.",
							Optional:            true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"grant": schema.SetAttribute{
										MarkdownDescription: "List of the fields to grant the access to.",
										Optional:            true,
										ElementType:         types.StringType,
									},
									"except": schema.SetAttribute{
										MarkdownDescription: "List of the fields to which the grants will not be applied.",
										Optional:            true,
										ElementType:         types.StringType,
									},
								},
							},
						},
						"names": schema.SetAttribute{
							MarkdownDescription: "A list of indices (or index name patterns) to which the permissions in this entry apply.",
							Required:            true,
							ElementType:         types.StringType,
						},
						"privileges": schema.SetAttribute{
							MarkdownDescription: "The index level privileges that the owners of the role have on the specified indices.",
							Required:            true,
							ElementType:         types.StringType,
						},
						"query": schema.StringAttribute{
							MarkdownDescription: "A search query that defines the documents the owners of the role have read access to.",
							Optional:            true,
						},
						"allow_restricted_indices": schema.BoolAttribute{
							MarkdownDescription: "Include matching restricted indices in names parameter. Usage is strongly discouraged as it can grant unrestricted operations on critical data, make the entire system unstable or leak sensitive information.",
							Optional:            true,
							Computed:            true,
							Default:             booldefault.StaticBool(false),
						},
					},
				},
			},
			"remote_indices": schema.SetNestedAttribute{
				MarkdownDescription: "A list of remote indices permissions entries. Remote indices are effective for remote clusters configured with the API key based model. They have no effect for remote clusters configured with the certificate based model.",
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"clusters": schema.SetAttribute{
							MarkdownDescription: "A list of cluster aliases to which the permissions in this entry apply.",
							Required:            true,
							ElementType:         types.StringType,
						},
						"field_security": schema.ListNestedAttribute{
							MarkdownDescription: "The document fields that the owners of the role have read access to.",
							Optional:            true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"grant": schema.SetAttribute{
										MarkdownDescription: "List of the fields to grant the access to.",
										Optional:            true,
										ElementType:         types.StringType,
									},
									"except": schema.SetAttribute{
										MarkdownDescription: "List of the fields to which the grants will not be applied.",
										Optional:            true,
										ElementType:         types.StringType,
									},
								},
							},
						},
						"query": schema.StringAttribute{
							MarkdownDescription: "A search query that defines the documents the owners of the role have read access to.",
							Optional:            true,
						},
						"names": schema.SetAttribute{
							MarkdownDescription: "A list of indices (or index name patterns) to which the permissions in this entry apply.",
							Required:            true,
							ElementType:         types.StringType,
						},
						"privileges": schema.SetAttribute{
							MarkdownDescription: "The index level privileges that the owners of the role have on the specified indices.",
							Required:            true,
							ElementType:         types.StringType,
						},
					},
				},
			},
			"metadata": schema.StringAttribute{
				MarkdownDescription: "Optional meta-data.",
				Optional:            true,
				Computed:            true,
			},
			"run_as": schema.SetAttribute{
				MarkdownDescription: "A list of users that the owners of this role can impersonate.",
				Optional:            true,
				ElementType:         types.StringType,
			},
		},
	}
}