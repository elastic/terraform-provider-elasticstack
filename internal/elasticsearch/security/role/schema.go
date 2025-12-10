package role

import (
	"context"
	_ "embed"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	providerschema "github.com/elastic/terraform-provider-elasticstack/internal/schema"
)

const CurrentSchemaVersion = 1

//go:embed resource-description.md
var roleResourceDescription string

func (r *roleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = GetSchema(CurrentSchemaVersion)
}

func GetSchema(version int64) schema.Schema {
	return schema.Schema{
		Version:             version,
		MarkdownDescription: roleResourceDescription,
		Blocks: map[string]schema.Block{
			"elasticsearch_connection": providerschema.GetEsFWConnectionBlock("elasticsearch_connection", false),
			"applications": schema.SetNestedBlock{
				MarkdownDescription: "A list of application privilege entries.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"application": schema.StringAttribute{
							MarkdownDescription: "The name of the application to which this entry applies.",
							Required:            true,
						},
						"privileges": schema.SetAttribute{
							MarkdownDescription: "A list of strings, where each element is the name of an application privilege or action.",
							Required:            true,
							ElementType:         types.StringType,
							Validators: []validator.Set{
								setvalidator.SizeAtLeast(1),
							},
						},
						"resources": schema.SetAttribute{
							MarkdownDescription: "A list resources to which the privileges are applied.",
							Required:            true,
							ElementType:         types.StringType,
							Validators: []validator.Set{
								setvalidator.SizeAtLeast(1),
							},
						},
					},
				},
			},
			"indices": schema.SetNestedBlock{
				MarkdownDescription: "A list of indices permissions entries.",
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"field_security": schema.SingleNestedBlock{
							MarkdownDescription: "The document fields that the owners of the role have read access to.",
							Attributes: map[string]schema.Attribute{
								"grant": schema.SetAttribute{
									MarkdownDescription: "List of the fields to grant the access to.",
									Optional:            true,
									ElementType:         types.StringType,
									Validators: []validator.Set{
										setvalidator.SizeAtLeast(1),
									},
								},
								"except": schema.SetAttribute{
									MarkdownDescription: "List of the fields to which the grants will not be applied.",
									Optional:            true,
									Computed:            true,
									ElementType:         types.StringType,
									PlanModifiers: []planmodifier.Set{
										setplanmodifier.UseStateForUnknown(),
									},
								},
							},
						},
					},
					Attributes: map[string]schema.Attribute{
						"names": schema.SetAttribute{
							MarkdownDescription: "A list of indices (or index name patterns) to which the permissions in this entry apply.",
							Required:            true,
							ElementType:         types.StringType,
							Validators: []validator.Set{
								setvalidator.SizeAtLeast(1),
							},
						},
						"privileges": schema.SetAttribute{
							MarkdownDescription: "The index level privileges that the owners of the role have on the specified indices.",
							Required:            true,
							ElementType:         types.StringType,
							Validators: []validator.Set{
								setvalidator.SizeAtLeast(1),
							},
						},
						"query": schema.StringAttribute{
							MarkdownDescription: "A search query that defines the documents the owners of the role have read access to.",
							Optional:            true,
							CustomType:          jsontypes.NormalizedType{},
						},
						"allow_restricted_indices": schema.BoolAttribute{
							MarkdownDescription: "Include matching restricted indices in names parameter. Usage is strongly discouraged as it can grant unrestricted operations on critical data, make the entire system unstable or leak sensitive information.",
							Optional:            true,
							Computed:            true,
							PlanModifiers: []planmodifier.Bool{
								boolplanmodifier.UseStateForUnknown(),
							},
						},
					},
				},
			},
			"remote_indices": schema.SetNestedBlock{
				MarkdownDescription: "A list of remote indices permissions entries. Remote indices are effective for remote clusters configured with the API key based model. They have no effect for remote clusters configured with the certificate based model.",
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"field_security": schema.SingleNestedBlock{
							MarkdownDescription: "The document fields that the owners of the role have read access to.",
							Attributes: map[string]schema.Attribute{
								"grant": schema.SetAttribute{
									MarkdownDescription: "List of the fields to grant the access to.",
									Optional:            true,
									ElementType:         types.StringType,
									Validators: []validator.Set{
										setvalidator.SizeAtLeast(1),
									},
								},
								"except": schema.SetAttribute{
									MarkdownDescription: "List of the fields to which the grants will not be applied.",
									Optional:            true,
									Computed:            true,
									ElementType:         types.StringType,
									PlanModifiers: []planmodifier.Set{
										setplanmodifier.UseStateForUnknown(),
									},
								},
							},
						},
					},
					Attributes: map[string]schema.Attribute{
						"clusters": schema.SetAttribute{
							MarkdownDescription: "A list of cluster aliases to which the permissions in this entry apply.",
							Required:            true,
							ElementType:         types.StringType,
						},
						"query": schema.StringAttribute{
							MarkdownDescription: "A search query that defines the documents the owners of the role have read access to.",
							Optional:            true,
							CustomType:          jsontypes.NormalizedType{},
						},
						"names": schema.SetAttribute{
							MarkdownDescription: "A list of indices (or index name patterns) to which the permissions in this entry apply.",
							Required:            true,
							ElementType:         types.StringType,
							Validators: []validator.Set{
								setvalidator.SizeAtLeast(1),
							},
						},
						"privileges": schema.SetAttribute{
							MarkdownDescription: "The index level privileges that the owners of the role have on the specified indices.",
							Required:            true,
							ElementType:         types.StringType,
							Validators: []validator.Set{
								setvalidator.SizeAtLeast(1),
							},
						},
					},
				},
			},
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
			"global": schema.StringAttribute{
				MarkdownDescription: "An object defining global privileges.",
				Optional:            true,
				CustomType:          jsontypes.NormalizedType{},
			},
			"cluster": schema.SetAttribute{
				MarkdownDescription: "A list of cluster privileges. These privileges define the cluster level actions that users with this role are able to execute.",
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
				},
			},
			"metadata": schema.StringAttribute{
				MarkdownDescription: "Optional meta-data.",
				Optional:            true,
				Computed:            true,
				CustomType:          jsontypes.NormalizedType{},
			},
			"run_as": schema.SetAttribute{
				MarkdownDescription: "A list of users that the owners of this role can impersonate.",
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

// Helper functions to get attribute types from the schema
func getApplicationAttrTypes() map[string]attr.Type {
	attrs := GetSchema(CurrentSchemaVersion).Blocks["applications"].(schema.SetNestedBlock).NestedObject.Attributes
	result := make(map[string]attr.Type)
	for name, attr := range attrs {
		result[name] = attr.GetType()
	}
	return result
}

func getFieldSecurityAttrTypes() map[string]attr.Type {
	attrs := GetSchema(CurrentSchemaVersion).Blocks["indices"].(schema.SetNestedBlock).NestedObject.Blocks["field_security"].(schema.SingleNestedBlock).Attributes
	result := make(map[string]attr.Type)
	for name, attr := range attrs {
		result[name] = attr.GetType()
	}
	return result
}

func getIndexPermsAttrTypes() map[string]attr.Type {
	nestedObj := GetSchema(CurrentSchemaVersion).Blocks["indices"].(schema.SetNestedBlock).NestedObject
	result := make(map[string]attr.Type)
	// Add attributes
	for name, attr := range nestedObj.Attributes {
		result[name] = attr.GetType()
	}
	// Add blocks as attributes (field_security is a block in indices)
	for name, block := range nestedObj.Blocks {
		switch b := block.(type) {
		case schema.SingleNestedBlock:
			// For SingleNestedBlock, the type is ObjectType
			blockAttrs := make(map[string]attr.Type)
			for attrName, attr := range b.Attributes {
				blockAttrs[attrName] = attr.GetType()
			}
			result[name] = types.ObjectType{AttrTypes: blockAttrs}
		}
	}
	return result
}

func getRemoteIndexPermsAttrTypes() map[string]attr.Type {
	nestedObj := GetSchema(CurrentSchemaVersion).Blocks["remote_indices"].(schema.SetNestedBlock).NestedObject
	result := make(map[string]attr.Type)
	// Add attributes
	for name, attr := range nestedObj.Attributes {
		result[name] = attr.GetType()
	}
	// Add blocks as attributes (field_security is a block in remote_indices)
	for name, block := range nestedObj.Blocks {
		switch b := block.(type) {
		case schema.SingleNestedBlock:
			// For SingleNestedBlock, the type is ObjectType
			blockAttrs := make(map[string]attr.Type)
			for attrName, attr := range b.Attributes {
				blockAttrs[attrName] = attr.GetType()
			}
			result[name] = types.ObjectType{AttrTypes: blockAttrs}
		}
	}
	return result
}

func getRemoteFieldSecurityAttrTypes() map[string]attr.Type {
	attrs := GetSchema(CurrentSchemaVersion).Blocks["remote_indices"].(schema.SetNestedBlock).NestedObject.Blocks["field_security"].(schema.SingleNestedBlock).Attributes
	result := make(map[string]attr.Type)
	for name, attr := range attrs {
		result[name] = attr.GetType()
	}
	return result
}
