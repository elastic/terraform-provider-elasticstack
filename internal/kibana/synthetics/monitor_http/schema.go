package monitor_http

import (
	"context"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/synthetics"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

//TODO:
// Ensure provider defined types fully satisfy framework interfaces

type Resource struct {
	client *clients.ApiClient
}

func HttpMonitorModeSchema() schema.Attribute {
	return schema.StringAttribute{
		Optional:    true,
		Description: "",
	}
}

func HTTPMonitorFieldsSchema() schema.Attribute {
	return schema.SingleNestedAttribute{
		Optional:    false,
		Description: "",
		Attributes: map[string]schema.Attribute{
			"url": schema.StringAttribute{
				Optional:    false,
				Description: "",
			},
			"ssl_setting": synthetics.JsonObjectSchema(),
			"max_redirects": schema.StringAttribute{
				Optional:    true,
				Description: "",
			},
			"mode": HttpMonitorModeSchema(),
			"ipv4": schema.BoolAttribute{
				Optional:    true,
				Description: "",
			},
			"ipv6": schema.BoolAttribute{
				Optional:    true,
				Description: "",
			},
			"username": schema.StringAttribute{
				Optional:    true,
				Description: "",
			},
			"password": schema.StringAttribute{
				Optional:    true,
				Description: "",
			},
			"proxy_header": synthetics.JsonObjectSchema(),
			"proxy_url": schema.StringAttribute{
				Optional:    true,
				Description: "",
			},
			"response": synthetics.JsonObjectSchema(),
			"check":    synthetics.JsonObjectSchema(),
		},
	}
}

func (r *Resource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = getSchema()
}

func getSchema() schema.Schema {
	return schema.Schema{}
}
