package proxy

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (r *proxyResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = getSchema()
}

func getSchema() schema.Schema {
	return schema.Schema{
		Description: "Creates and manages a Fleet Proxy. Proxies can be used by Fleet Server Hosts and Elasticsearch Outputs to route traffic through a proxy server.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of this resource.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"proxy_id": schema.StringAttribute{
				Description: "Unique identifier of the proxy. If not specified, one will be generated.",
				Computed:    true,
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the proxy.",
				Required:    true,
			},
			"url": schema.StringAttribute{
				Description: "The proxy URL (e.g., https://proxy.example.com:8080).",
				Required:    true,
			},
			"certificate": schema.StringAttribute{
				Description: "PEM-encoded client certificate for TLS authentication with the proxy.",
				Optional:    true,
			},
			"certificate_authorities": schema.StringAttribute{
				Description: "PEM-encoded certificate authorities for verifying the proxy server certificate.",
				Optional:    true,
			},
			"certificate_key": schema.StringAttribute{
				Description: "PEM-encoded private key for the client certificate.",
				Optional:    true,
				Sensitive:   true,
			},
			"is_preconfigured": schema.BoolAttribute{
				Description: "Indicates if the proxy is preconfigured (managed outside Terraform). Note: This field cannot be updated after creation; changes will force resource recreation.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"space_ids": schema.SetAttribute{
				Description: "The Kibana space IDs where this proxy is available. When multiple spaces are specified, the first space ID is used for API operations. Note: The order of space IDs does not matter as this is a set.",
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}
