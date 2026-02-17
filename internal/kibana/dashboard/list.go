package dashboard

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibana_oapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/list"
	listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ list.ListResource              = &dashboardsListResource{}
	_ list.ListResourceWithConfigure = &dashboardsListResource{}
)

func NewListResource() list.ListResource {
	return &dashboardsListResource{}
}

type dashboardsListResource struct {
	client *clients.ApiClient
}

type dashboardsListConfigModelV0 struct {
	SpaceID      types.String `tfsdk:"space_id"`
	Search       types.String `tfsdk:"search"`
	TagsIncluded types.List   `tfsdk:"tags_included"`
	TagsExcluded types.List   `tfsdk:"tags_excluded"`
	PerPage      types.Int64  `tfsdk:"per_page"`
}

func (r *dashboardsListResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_%s", req.ProviderTypeName, "kibana_dashboard")
}

func (r *dashboardsListResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, diags := clients.ConvertProviderData(req.ProviderData)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	r.client = client
}

func (r *dashboardsListResource) ListResourceConfigSchema(ctx context.Context, req list.ListResourceSchemaRequest, resp *list.ListResourceSchemaResponse) {
	resp.Schema = listschema.Schema{
		Attributes: map[string]listschema.Attribute{
			"space_id": listschema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Kibana space to search. Defaults to `default`.",
			},
			"search": listschema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Elasticsearch simple_query_string query that filters dashboards by title and description.",
			},
			"tags_included": listschema.ListAttribute{
				Optional:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Only include dashboards with these tag IDs.",
			},
			"tags_excluded": listschema.ListAttribute{
				Optional:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Exclude dashboards with these tag IDs.",
			},
			"per_page": listschema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "Number of dashboards per page to request from Kibana (server-side paging).",
			},
		},
	}
}

func (r *dashboardsListResource) List(ctx context.Context, req list.ListRequest, stream *list.ListResultsStream) {
	var cfg dashboardsListConfigModelV0
	diags := req.Config.Get(ctx, &cfg)
	if diags.HasError() {
		stream.Results = list.ListResultsStreamDiagnostics(diags)
		return
	}

	if r.client == nil {
		var errDiags diag.Diagnostics
		errDiags.AddError("Provider not configured", "Missing provider configuration while listing dashboards.")
		stream.Results = list.ListResultsStreamDiagnostics(errDiags)
		return
	}

	spaceID := "default"
	if !cfg.SpaceID.IsNull() && !cfg.SpaceID.IsUnknown() && cfg.SpaceID.ValueString() != "" {
		spaceID = cfg.SpaceID.ValueString()
	}

	perPage := int64(100)
	if !cfg.PerPage.IsNull() && !cfg.PerPage.IsUnknown() && cfg.PerPage.ValueInt64() > 0 {
		perPage = cfg.PerPage.ValueInt64()
	}

	kibanaClient, err := r.client.GetKibanaOapiClient()
	if err != nil {
		var errDiags diag.Diagnostics
		errDiags.AddError("Unable to get Kibana client", err.Error())
		stream.Results = list.ListResultsStreamDiagnostics(errDiags)
		return
	}

	stream.Results = func(push func(list.ListResult) bool) {
		var pushed int64
		var page float32 = 1

		for {
			// Respect Terraform's requested limit (also enforced by core via stream termination).
			if req.Limit > 0 && pushed >= req.Limit {
				return
			}

			var bodyDiags diag.Diagnostics
			body := newDashboardsSearchBody(ctx, cfg, page, perPage, &bodyDiags)
			if bodyDiags.HasError() {
				result := req.NewListResult(ctx)
				result.Diagnostics.AddWarning("Failed to build dashboards search request", "One or more list config values could not be converted.")
				result.Diagnostics.Append(bodyDiags...)
				_ = push(result)
				return
			}

			searchResp, searchDiags := kibana_oapi.SearchDashboards(ctx, kibanaClient, spaceID, body)
			if searchDiags.HasError() {
				result := req.NewListResult(ctx)
				result.Diagnostics.AddWarning("Dashboard search failed", "The dashboards search endpoint returned an error. Partial results may have been returned.")
				result.Diagnostics.Append(searchDiags...)
				_ = push(result)
				return
			}

			if searchResp == nil || searchResp.JSON200 == nil {
				// No results / unexpected empty response: stop paging.
				return
			}

			dashboards := searchResp.JSON200.Dashboards
			if len(dashboards) == 0 {
				return
			}

			for _, d := range dashboards {
				if req.Limit > 0 && pushed >= req.Limit {
					return
				}

				result := req.NewListResult(ctx)

				compID := clients.CompositeId{ClusterId: spaceID, ResourceId: d.Id}
				result.DisplayName = d.Data.Title
				result.Diagnostics.Append(result.Identity.Set(ctx, identityModelV0{
					ID: types.StringValue(compID.String()),
				})...)

				if req.IncludeResource {
					fullModel, ok := r.fetchFullDashboard(ctx, kibanaClient, spaceID, d.Id)
					if !ok {
						// If we can't fetch full resource details, skip this result to avoid
						// returning a null resource when include_resource=true.
						continue
					}

					result.Diagnostics.Append(result.Resource.Set(ctx, fullModel)...)
				}

				if !push(result) {
					return
				}
				pushed++
			}

			// If the returned page is smaller than per_page, we're done.
			if int64(len(dashboards)) < perPage {
				return
			}

			page++
		}
	}
}

func (r *dashboardsListResource) fetchFullDashboard(ctx context.Context, kibanaClient *kibana_oapi.Client, spaceID string, dashboardID string) (dashboardModel, bool) {
	var model dashboardModel

	getResp, diags := kibana_oapi.GetDashboard(ctx, kibanaClient, spaceID, dashboardID)
	if diags.HasError() || getResp == nil || getResp.JSON200 == nil {
		return model, false
	}

	diags = model.populateFromAPI(ctx, getResp, dashboardID, spaceID)
	if diags.HasError() {
		return model, false
	}

	return model, true
}

func newDashboardsSearchBody(ctx context.Context, cfg dashboardsListConfigModelV0, page float32, perPage int64, diags *diag.Diagnostics) kbapi.PostDashboardsSearchJSONRequestBody {
	p := page
	pp := float32(perPage)

	body := kbapi.PostDashboardsSearchJSONRequestBody{
		Page:    &p,
		PerPage: &pp,
	}

	if !cfg.Search.IsNull() && !cfg.Search.IsUnknown() && cfg.Search.ValueString() != "" {
		s := cfg.Search.ValueString()
		body.Search = &s
	}

	includedVals := utils.ListTypeToSlice_String(ctx, cfg.TagsIncluded, path.Root("tags_included"), diags)
	excludedVals := utils.ListTypeToSlice_String(ctx, cfg.TagsExcluded, path.Root("tags_excluded"), diags)
	if diags.HasError() {
		return body
	}

	var included *[]string
	if includedVals != nil {
		included = &includedVals
	}
	var excluded *[]string
	if excludedVals != nil {
		excluded = &excludedVals
	}

	if included != nil || excluded != nil {
		body.Tags = &struct {
			Excluded *[]string `json:"excluded,omitempty"`
			Included *[]string `json:"included,omitempty"`
		}{
			Excluded: excluded,
			Included: included,
		}
	}

	return body
}
