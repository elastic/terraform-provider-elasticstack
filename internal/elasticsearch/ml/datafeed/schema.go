package datafeed

import (
	"context"
	"regexp"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/validators"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	providerschema "github.com/elastic/terraform-provider-elasticstack/internal/schema"
)

func (r *datafeedResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = GetSchema()
}

func GetSchema() schema.Schema {
	return schema.Schema{
		MarkdownDescription: "Creates and manages Machine Learning datafeeds. Datafeeds retrieve data from Elasticsearch for analysis by an anomaly detection job. Each anomaly detection job can have only one associated datafeed. See the [ML Datafeed API documentation](https://www.elastic.co/guide/en/elasticsearch/reference/current/ml-put-datafeed.html) for more details.",
		Blocks: map[string]schema.Block{
			"elasticsearch_connection": providerschema.GetEsFWConnectionBlock("elasticsearch_connection", false),
		},
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Internal identifier of the resource",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"datafeed_id": schema.StringAttribute{
				MarkdownDescription: "A numerical character string that uniquely identifies the datafeed. This identifier can contain lowercase alphanumeric characters (a-z and 0-9), hyphens, and underscores. It must start and end with alphanumeric characters.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 64),
					stringvalidator.RegexMatches(regexp.MustCompile(`^[a-z0-9][a-z0-9_-]*[a-z0-9]$|^[a-z0-9]$`), "must contain lowercase alphanumeric characters (a-z and 0-9), hyphens, and underscores. It must start and end with alphanumeric characters"),
				},
			},
			"job_id": schema.StringAttribute{
				MarkdownDescription: "Identifier for the anomaly detection job. The job must exist before creating the datafeed.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"indices": schema.ListAttribute{
				MarkdownDescription: "An array of index names. Wildcards are supported. If any of the indices are in remote clusters, the machine learning nodes must have the `remote_cluster_client` role.",
				Required:            true,
				ElementType:         types.StringType,
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
			},
			"query": schema.StringAttribute{
				MarkdownDescription: "The Elasticsearch query domain-specific language (DSL). This value corresponds to the query object in an Elasticsearch search POST body. All the options that are supported by Elasticsearch can be used, as this object is passed verbatim to Elasticsearch. By default uses `{\"match_all\": {\"boost\": 1}}`.",
				Optional:            true,
				Computed:            true,
				CustomType:          jsontypes.NormalizedType{},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"aggregations": schema.StringAttribute{
				MarkdownDescription: "If set, the datafeed performs aggregation searches. Support for aggregations is limited and should be used only with low cardinality data. This should be a JSON object representing the aggregations to be performed.",
				Optional:            true,
				CustomType:          jsontypes.NormalizedType{},
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.Root("script_fields").Expression()),
				},
			},
			"script_fields": schema.StringAttribute{
				MarkdownDescription: "Specifies scripts that evaluate custom expressions and returns script fields to the datafeed. The detector configuration objects in a job can contain functions that use these script fields. This should be a JSON object representing the script fields.",
				Optional:            true,
				CustomType:          customtypes.NewJSONWithDefaultsType(populateScriptFieldsDefaults),
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.Root("aggregations").Expression()),
				},
			},
			"runtime_mappings": schema.StringAttribute{
				MarkdownDescription: "Specifies runtime fields for the datafeed search. This should be a JSON object representing the runtime field mappings.",
				Optional:            true,
				CustomType:          jsontypes.NormalizedType{},
			},
			"scroll_size": schema.Int64Attribute{
				MarkdownDescription: "The size parameter that is used in Elasticsearch searches when the datafeed does not use aggregations. The maximum value is the value of `index.max_result_window`, which is 10,000 by default.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
				Validators: []validator.Int64{
					int64validator.AtLeast(1),
				},
			},
			"frequency": schema.StringAttribute{
				MarkdownDescription: "The interval at which scheduled queries are made while the datafeed runs in real time. The default value is either the bucket span for short bucket spans, or, for longer bucket spans, a sensible fraction of the bucket span. When `frequency` is shorter than the bucket span, interim results for the last (partial) bucket are written then eventually overwritten by the full bucket results. If the datafeed uses aggregations, this value must be divisible by the interval of the date histogram aggregation.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(`^\d+[nsumdh]$`), "must be a valid duration (e.g., 150s, 10m, 1h)"),
				},
			},
			"query_delay": schema.StringAttribute{
				MarkdownDescription: "The number of seconds behind real time that data is queried. For example, if data from 10:04 a.m. might not be searchable in Elasticsearch until 10:06 a.m., set this property to 120 seconds. The default value is randomly selected between `60s` and `120s`. This randomness improves the query performance when there are multiple jobs running on the same node.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(`^\d+[nsumdh]$`), "must be a valid duration (e.g., 60s, 2m)"),
				},
			},
			"max_empty_searches": schema.Int64Attribute{
				MarkdownDescription: "If a real-time datafeed has never seen any data (including during any initial training period), it automatically stops and closes the associated job after this many real-time searches return no documents. In other words, it stops after `frequency` times `max_empty_searches` of real-time operation. If not set, a datafeed with no end time that sees no data remains started until it is explicitly stopped.",
				Optional:            true,
				Validators: []validator.Int64{
					int64validator.AtLeast(1),
				},
			},
			"chunking_config": schema.SingleNestedAttribute{
				MarkdownDescription: "Datafeeds might search over long time periods, for several months or years. This search is split into time chunks in order to ensure the load on Elasticsearch is managed. Chunking configuration controls how the size of these time chunks are calculated; it is an advanced configuration option.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.UseStateForUnknown(),
				},
				Attributes: map[string]schema.Attribute{
					"mode": schema.StringAttribute{
						MarkdownDescription: "The chunking mode. Can be `auto`, `manual`, or `off`. In `auto` mode, the chunk size is dynamically calculated. In `manual` mode, chunking is applied according to the specified `time_span`. In `off` mode, no chunking is applied.",
						Required:            true,
						Validators: []validator.String{
							stringvalidator.OneOf("auto", "manual", "off"),
						},
					},
					"time_span": schema.StringAttribute{
						MarkdownDescription: "The time span for each chunk. Only applicable and required when mode is `manual`. Must be a valid duration.",
						Optional:            true,
						Computed:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
						Validators: []validator.String{
							stringvalidator.RegexMatches(regexp.MustCompile(`^\d+[nsumdh]$`), "must be a valid duration (e.g., 1h, 1d)"),
							validators.AllowedIfDependentPathEquals(path.Root("chunking_config").AtName("mode"), "manual"),
						},
					},
				},
			},
			"delayed_data_check_config": schema.SingleNestedAttribute{
				MarkdownDescription: "Specifies whether the datafeed checks for missing data and the size of the window. The datafeed can optionally search over indices that have already been read in an effort to determine whether any data has subsequently been added to the index. If missing data is found, it is a good indication that the `query_delay` is set too low and the data is being indexed after the datafeed has passed that moment in time. This check runs only on real-time datafeeds.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.UseStateForUnknown(),
				},
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{
						MarkdownDescription: "Specifies whether the datafeed periodically checks for delayed data.",
						Required:            true,
					},
					"check_window": schema.StringAttribute{
						MarkdownDescription: "The window of time that is searched for late data. This window of time ends with the latest finalized bucket. It defaults to null, which causes an appropriate `check_window` to be calculated when the real-time datafeed runs.",
						Optional:            true,
						Computed:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
						Validators: []validator.String{
							stringvalidator.RegexMatches(regexp.MustCompile(`^\d+[nsumdh]$`), "must be a valid duration (e.g., 2h, 1d)"),
						},
					},
				},
			},
			"indices_options": schema.SingleNestedAttribute{
				MarkdownDescription: "Specifies index expansion options that are used during search.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.UseStateForUnknown(),
				},
				Attributes: map[string]schema.Attribute{
					"expand_wildcards": schema.ListAttribute{
						MarkdownDescription: "Type of index that wildcard patterns can match. If the request can target data streams, this argument determines whether wildcard expressions match hidden data streams. Supports comma-separated values.",
						Optional:            true,
						Computed:            true,
						ElementType:         types.StringType,
						PlanModifiers: []planmodifier.List{
							listplanmodifier.UseStateForUnknown(),
						},
						Validators: []validator.List{
							listvalidator.ValueStringsAre(
								stringvalidator.OneOf("all", "open", "closed", "hidden", "none"),
							),
						},
					},
					"ignore_unavailable": schema.BoolAttribute{
						MarkdownDescription: "If true, unavailable indices (missing or closed) are ignored.",
						Optional:            true,
						Computed:            true,
						PlanModifiers: []planmodifier.Bool{
							boolplanmodifier.UseStateForUnknown(),
						},
					},
					"allow_no_indices": schema.BoolAttribute{
						MarkdownDescription: "If true, wildcard indices expressions that resolve into no concrete indices are ignored. This includes the `_all` string or when no indices are specified.",
						Optional:            true,
						Computed:            true,
						PlanModifiers: []planmodifier.Bool{
							boolplanmodifier.UseStateForUnknown(),
						},
					},
					"ignore_throttled": schema.BoolAttribute{
						MarkdownDescription: "If true, concrete, expanded, or aliased indices are ignored when frozen. This setting is deprecated.",
						Optional:            true,
						Computed:            true,
						PlanModifiers: []planmodifier.Bool{
							boolplanmodifier.UseStateForUnknown(),
						},
						DeprecationMessage: "indices_options.ignore_throttled is deprecated and will be removed in a future version.",
					},
				},
			},
		},
	}
}

// GetChunkingConfigAttrTypes returns the attribute types for chunking_config
func GetChunkingConfigAttrTypes() map[string]attr.Type {
	return GetSchema().Attributes["chunking_config"].GetType().(attr.TypeWithAttributeTypes).AttributeTypes()
}

// GetDelayedDataCheckConfigAttrTypes returns the attribute types for delayed_data_check_config
func GetDelayedDataCheckConfigAttrTypes() map[string]attr.Type {
	return GetSchema().Attributes["delayed_data_check_config"].GetType().(attr.TypeWithAttributeTypes).AttributeTypes()
}

// GetIndicesOptionsAttrTypes returns the attribute types for indices_options
func GetIndicesOptionsAttrTypes() map[string]attr.Type {
	return GetSchema().Attributes["indices_options"].GetType().(attr.TypeWithAttributeTypes).AttributeTypes()
}
