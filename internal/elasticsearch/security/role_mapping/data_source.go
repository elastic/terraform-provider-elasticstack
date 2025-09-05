package role_mapping

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	providerschema "github.com/elastic/terraform-provider-elasticstack/internal/schema"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewRoleMappingDataSource() datasource.DataSource {
	return &roleMappingDataSource{}
}

type roleMappingDataSource struct {
	client *clients.ApiClient
}

func (d *roleMappingDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_elasticsearch_security_role_mapping"
}

func (d *roleMappingDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieves role mappings. See, https://www.elastic.co/guide/en/elasticsearch/reference/current/security-api-get-role-mapping.html",
		Blocks: map[string]schema.Block{
			"elasticsearch_connection": providerschema.GetEsFWConnectionBlock("elasticsearch_connection", false),
		},
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Internal identifier of the resource",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The distinct name that identifies the role mapping, used solely as an identifier.",
				Required:            true,
			},
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "Mappings that have `enabled` set to `false` are ignored when role mapping is performed.",
				Computed:            true,
			},
			"rules": schema.StringAttribute{
				MarkdownDescription: "The rules that determine which users should be matched by the mapping. A rule is a logical condition that is expressed by using a JSON DSL.",
				Computed:            true,
			},
			"roles": schema.SetAttribute{
				MarkdownDescription: "A list of role names that are granted to the users that match the role mapping rules.",
				ElementType:         types.StringType,
				Computed:            true,
			},
			"role_templates": schema.StringAttribute{
				MarkdownDescription: "A list of mustache templates that will be evaluated to determine the roles names that should granted to the users that match the role mapping rules.",
				Computed:            true,
			},
			"metadata": schema.StringAttribute{
				MarkdownDescription: "Additional metadata that helps define which roles are assigned to each user. Keys beginning with `_` are reserved for system usage.",
				Computed:            true,
			},
		},
	}
}

func (d *roleMappingDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	client, diags := clients.ConvertProviderData(req.ProviderData)
	resp.Diagnostics.Append(diags...)
	d.client = client
}

func (d *roleMappingDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data RoleMappingData
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	roleMappingName := data.Name.ValueString()
	id, sdkDiags := d.client.ID(ctx, roleMappingName)
	resp.Diagnostics.Append(utils.FrameworkDiagsFromSDK(sdkDiags)...)
	if resp.Diagnostics.HasError() {
		return
	}

	data.Id = types.StringValue(id.String())

	// Reuse the resource read logic
	resourceRead := &roleMappingResource{client: d.client}
	
	// Create a mock state and request
	state := tfsdk.State{
		Schema: GetSchema(),
	}
	state.Set(ctx, &data)
	
	readReq := resource.ReadRequest{State: state}
	readResp := &resource.ReadResponse{
		State: state,
	}
	
	resourceRead.Read(ctx, readReq, readResp)
	resp.Diagnostics.Append(readResp.Diagnostics...)
	if resp.Diagnostics.HasError() {
		return
	}

	var resultData RoleMappingData
	resp.Diagnostics.Append(readResp.State.Get(ctx, &resultData)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &resultData)...)
}