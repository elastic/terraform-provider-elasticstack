package synthetics

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/disaster37/go-kibana-rest/v8/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"strconv"
)

const (
	MetadataPrefix = "_kibana_synthetics_"
)

type kibanaAPIRequest struct {
	fields kbapi.MonitorFields
	config kbapi.SyntheticsMonitorConfig
}

type tfStatusConfigV0 struct {
	Enabled types.Bool `tfsdk:"enabled"`
}

type tfAlertConfigV0 struct {
	Status *tfStatusConfigV0 `tfsdk:"status"`
	TLS    *tfStatusConfigV0 `tfsdk:"tls"`
}

type tfHTTPMonitorFieldsV0 struct {
	URL                   types.String         `tfsdk:"url"`
	SslVerificationMode   types.String         `tfsdk:"ssl_verification_mode"`
	SslSupportedProtocols types.List           `tfsdk:"ssl_supported_protocols"`
	MaxRedirects          types.Int64          `tfsdk:"max_redirects"`
	Mode                  types.String         `tfsdk:"mode"`
	IPv4                  types.Bool           `tfsdk:"ipv4"`
	IPv6                  types.Bool           `tfsdk:"ipv6"`
	ProxyURL              types.String         `tfsdk:"proxy_url"`
	ProxyHeader           jsontypes.Normalized `tfsdk:"proxy_header"`
	Username              types.String         `tfsdk:"username"`
	Password              types.String         `tfsdk:"password"`
	Response              jsontypes.Normalized `tfsdk:"response"`
	Check                 jsontypes.Normalized `tfsdk:"check"`
}

type tfTCPMonitorFieldsV0 struct {
	Host                  types.String `tfsdk:"host"`
	SslVerificationMode   types.String `tfsdk:"ssl_verification_mode"`
	SslSupportedProtocols types.List   `tfsdk:"ssl_supported_protocols"`
	CheckSend             types.String `tfsdk:"check_send"`
	CheckReceive          types.String `tfsdk:"check_receive"`
	ProxyURL              types.String `tfsdk:"proxy_url"`
	ProxyUseLocalResolver types.Bool   `tfsdk:"proxy_use_local_resolver"`
}

type tfICMPMonitorFieldsV0 struct {
	Host types.String `tfsdk:"host"`
	Wait types.Int64  `tfsdk:"wait"`
}

type tfBrowserMonitorFieldsV0 struct {
	InlineScript      types.String         `tfsdk:"inline_script"`
	Screenshots       types.String         `tfsdk:"screenshots"`
	SyntheticsArgs    []types.String       `tfsdk:"synthetics_args"`
	IgnoreHttpsErrors types.Bool           `tfsdk:"ignore_https_errors"`
	PlaywrightOptions jsontypes.Normalized `tfsdk:"playwright_options"`
}

type tfModelV0 struct {
	ID               types.String              `tfsdk:"id"`
	Name             types.String              `tfsdk:"name"`
	SpaceID          types.String              `tfsdk:"space_id"`
	Schedule         types.Int64               `tfsdk:"schedule"`
	Locations        []types.String            `tfsdk:"locations"`
	PrivateLocations []types.String            `tfsdk:"private_locations"`
	Enabled          types.Bool                `tfsdk:"enabled"`
	Tags             []types.String            `tfsdk:"tags"`
	Alert            *tfAlertConfigV0          `tfsdk:"alert"`
	APMServiceName   types.String              `tfsdk:"service_name"`
	TimeoutSeconds   types.Int64               `tfsdk:"timeout"`
	HTTP             *tfHTTPMonitorFieldsV0    `tfsdk:"http"`
	TCP              *tfTCPMonitorFieldsV0     `tfsdk:"tcp"`
	ICMP             *tfICMPMonitorFieldsV0    `tfsdk:"icmp"`
	Browser          *tfBrowserMonitorFieldsV0 `tfsdk:"browser"`
	Params           jsontypes.Normalized      `tfsdk:"params"`
	RetestOnFailure  types.Bool                `tfsdk:"retest_on_failure"`
}

func GetCompositeId(id string) (*clients.CompositeId, diag.Diagnostics) {
	compositeID, sdkDiag := clients.CompositeIdFromStr(id)
	dg := diag.Diagnostics{}
	if sdkDiag.HasError() {
		dg.AddError(fmt.Sprintf("Failed to parse monitor ID %s", id), fmt.Sprintf("Resource ID must have following format: <cluster_uuid>/<resource identifier>. Current value: %s", id))
		return nil, dg
	}
	return compositeID, dg
}

func monitorConfigSchema() schema.Schema {
	return schema.Schema{
		MarkdownDescription: "Synthetics monitor config, see https://www.elastic.co/guide/en/kibana/current/add-monitor-api.html for more details. The monitor must have one of the following: http, tcp, icmp or browser.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Generated identifier for the monitor",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Optional:            false,
				Required:            true,
				MarkdownDescription: "The monitor’s name.",
			},
			"space_id": schema.StringAttribute{
				MarkdownDescription: "The namespace field should be lowercase and not contain spaces. The namespace must not include any of the following characters: *, \\, /, ?, \", <, >, |, whitespace, ,, #, :, or -. Default: `default`",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"schedule": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "The monitor’s schedule in minutes. Supported values are 1, 3, 5, 10, 15, 30, 60, 120 and 240.",
				Validators: []validator.Int64{
					int64validator.OneOf(1, 3, 5, 10, 15, 30, 60, 120, 240),
				},
			},
			"locations": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "Where to deploy the monitor. Monitors can be deployed in multiple locations so that you can detect differences in availability and response times across those locations.",
				Validators: []validator.List{
					listvalidator.ValueStringsAre(
						stringvalidator.OneOf(
							"japan",
							"india",
							"singapore",
							"australia_east",
							"united_kingdom",
							"germany",
							"canada_east",
							"brazil",
							"us_east",
							"us_west",
						),
					),
				},
			},
			"private_locations": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "These Private Locations refer to locations hosted and managed by you, whereas locations are hosted by Elastic. You can specify a Private Location using the location’s name.",
			},
			"enabled": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "Whether the monitor is enabled. Default: `true`",
			},
			"tags": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "An array of tags.",
			},
			"alert": monitorAlertConfigSchema(),
			"service_name": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The APM service name.",
			},
			"timeout": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "The monitor timeout in seconds, monitor will fail if it doesn’t complete within this time. Default: `16`",
			},
			"params":  jsonObjectSchema("Monitor parameters"),
			"http":    httpMonitorFieldsSchema(),
			"tcp":     tcpMonitorFieldsSchema(),
			"icmp":    icmpMonitorFieldsSchema(),
			"browser": browserMonitorFieldsSchema(),
			"retest_on_failure": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "Enable or disable retesting when a monitor fails. By default, monitors are automatically retested if the monitor goes from \"up\" to \"down\". If the result of the retest is also \"down\", an error will be created, and if configured, an alert sent. Then the monitor will resume running according to the defined schedule. Using retest_on_failure can reduce noise related to transient problems. Default: `true`.",
			},
		},
	}
}

func browserMonitorFieldsSchema() schema.Attribute {
	return schema.SingleNestedAttribute{
		Optional:            true,
		MarkdownDescription: "Browser Monitor specific fields",
		Attributes: map[string]schema.Attribute{
			"inline_script": schema.StringAttribute{
				Optional:            false,
				Required:            true,
				MarkdownDescription: "The inline script.",
			},
			"screenshots": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Controls the behavior of the screenshots feature.",
				Validators: []validator.String{
					stringvalidator.OneOf("on", "off", "only-on-failure"),
				},
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"synthetics_args": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "Synthetics agent CLI arguments.",
			},
			"ignore_https_errors": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "Whether to ignore HTTPS errors.",
				Computed:            true,
				PlanModifiers:       []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
			},
			"playwright_options": jsonObjectSchema("Playwright options."),
		},
	}
}

func icmpMonitorFieldsSchema() schema.Attribute {
	return schema.SingleNestedAttribute{
		Optional:            true,
		MarkdownDescription: "ICMP Monitor specific fields",
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				Optional:            false,
				Required:            true,
				MarkdownDescription: "Host to ping; it can be an IP address or a hostname.",
			},
			"wait": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: " Wait time in seconds. Default: `1`",
				PlanModifiers:       []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
				Computed:            true,
			},
		},
	}
}

func jsonObjectSchema(doc string) schema.Attribute {
	return schema.StringAttribute{
		Optional:            true,
		MarkdownDescription: fmt.Sprintf("%s. Raw JSON object, use `jsonencode` function to represent JSON", doc),
		CustomType:          jsontypes.NormalizedType{},
	}
}

func statusConfigSchema() schema.Attribute {
	return schema.SingleNestedAttribute{
		Optional: true,
		Attributes: map[string]schema.Attribute{
			"enabled": schema.BoolAttribute{
				Optional: true,
			},
		},
	}
}

func monitorAlertConfigSchema() schema.Attribute {
	return schema.SingleNestedAttribute{
		Optional:            true,
		MarkdownDescription: "Alert configuration. Default: `{ status: { enabled: true }, tls: { enabled: true } }`.",
		Attributes: map[string]schema.Attribute{
			"status": statusConfigSchema(),
			"tls":    statusConfigSchema(),
		},
	}
}

func httpMonitorFieldsSchema() schema.Attribute {
	return schema.SingleNestedAttribute{
		Optional:            true,
		MarkdownDescription: "HTTP Monitor specific fields",
		Attributes: map[string]schema.Attribute{
			"url": schema.StringAttribute{
				Optional:            false,
				Required:            true,
				MarkdownDescription: "URL to monitor.",
			},
			"ssl_verification_mode": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Controls the verification of server certificates. ",
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"ssl_supported_protocols": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "List of allowed SSL/TLS versions.",
				Computed:            true,
				PlanModifiers:       []planmodifier.List{listplanmodifier.UseStateForUnknown()},
			},
			"max_redirects": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "The maximum number of redirects to follow. Default: `0`",
				PlanModifiers:       []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
				Computed:            true,
			},
			"mode": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The mode of the monitor. Can be \"all\" or \"any\". If you’re using a DNS-load balancer and want to ping every IP address for the specified hostname, you should use all.",
				Validators: []validator.String{
					stringvalidator.OneOf("any", "all"),
				},
			},
			"ipv4": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "Whether to ping using the ipv4 protocol.",
			},
			"ipv6": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "Whether to ping using the ipv6 protocol.",
			},
			"username": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The username for authenticating with the server. The credentials are passed with the request.",
			},
			"password": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The password for authenticating with the server. The credentials are passed with the request.",
			},
			"proxy_header": jsonObjectSchema("Additional headers to send to proxies during CONNECT requests."),
			"proxy_url": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The URL of the proxy to use for this monitor.",
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"response": jsonObjectSchema("Controls the indexing of the HTTP response body contents to the `http.response.body.contents` field."),
			"check":    jsonObjectSchema("The check request settings."),
		},
	}
}

func tcpMonitorFieldsSchema() schema.Attribute {
	return schema.SingleNestedAttribute{
		Optional:            true,
		MarkdownDescription: "TCP Monitor specific fields",
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				Optional:            false,
				Required:            true,
				MarkdownDescription: "The host to monitor; it can be an IP address or a hostname. The host can include the port using a colon (e.g., \"example.com:9200\").",
			},
			"ssl_verification_mode": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Controls the verification of server certificates. ",
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"ssl_supported_protocols": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "List of allowed SSL/TLS versions.",
				Computed:            true,
				PlanModifiers:       []planmodifier.List{listplanmodifier.UseStateForUnknown()},
			},
			"check_send": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "An optional payload string to send to the remote host.",
			},
			"check_receive": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The expected answer. ",
			},
			"proxy_url": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The URL of the SOCKS5 proxy to use when connecting to the server. The value must be a URL with a scheme of `socks5://`. If the SOCKS5 proxy server requires client authentication, then a username and password can be embedded in the URL. When using a proxy, hostnames are resolved on the proxy server instead of on the client. You can change this behavior by setting the `proxy_use_local_resolver` option.",
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"proxy_use_local_resolver": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: " A Boolean value that determines whether hostnames are resolved locally instead of being resolved on the proxy server. The default value is false, which means that name resolution occurs on the proxy server.",
			},
		},
	}
}

func GeoConfigSchema() schema.Attribute {
	return schema.SingleNestedAttribute{
		Optional:    true,
		Description: "Geographic coordinates (WGS84) for the location",
		Attributes: map[string]schema.Attribute{
			"lat": schema.Float64Attribute{
				Optional:            false,
				Required:            true,
				MarkdownDescription: "The latitude of the location.",
			},
			"lon": schema.Float64Attribute{
				Optional:            false,
				Required:            true,
				MarkdownDescription: "The longitude of the location.",
			},
		},
	}
}

type TFGeoConfigV0 struct {
	Lat types.Float64 `tfsdk:"lat"`
	Lon types.Float64 `tfsdk:"lon"`
}

func (m *TFGeoConfigV0) ToSyntheticGeoConfig() *kbapi.SyntheticGeoConfig {
	return &kbapi.SyntheticGeoConfig{
		Lat: m.Lat.ValueFloat64(),
		Lon: m.Lon.ValueFloat64(),
	}
}

func FromSyntheticGeoConfig(v *kbapi.SyntheticGeoConfig) *TFGeoConfigV0 {
	if v == nil {
		return nil
	}
	return &TFGeoConfigV0{
		Lat: types.Float64Value(v.Lat),
		Lon: types.Float64Value(v.Lon),
	}
}

func ValueStringSlice(v []types.String) []string {
	var res []string
	for _, s := range v {
		res = append(res, s.ValueString())
	}
	return res
}

func StringSliceValue(v []string) []types.String {
	var res []types.String
	for _, s := range v {
		res = append(res, types.StringValue(s))
	}
	return res
}

func toNormalizedValue(jsObj kbapi.JsonObject) (jsontypes.Normalized, error) {
	res, err := json.Marshal(jsObj)
	if err != nil {
		return jsontypes.NewNormalizedUnknown(), err
	}
	return jsontypes.NewNormalizedValue(string(res)), nil
}

func toJsonObject(v jsontypes.Normalized) (kbapi.JsonObject, diag.Diagnostics) {
	if v.IsNull() {
		return nil, diag.Diagnostics{}
	}
	var res kbapi.JsonObject
	dg := v.Unmarshal(&res)
	if dg.HasError() {
		return nil, dg
	}
	return res, diag.Diagnostics{}
}

func stringToInt64(v string) (int64, error) {
	var res int64
	var err error
	if v != "" {
		res, err = strconv.ParseInt(v, 10, 64)
	}
	return res, err
}

func (v *tfModelV0) toModelV0(ctx context.Context, api *kbapi.SyntheticsMonitor) (*tfModelV0, diag.Diagnostics) {
	var schedule int64
	var err error
	dg := diag.Diagnostics{}
	if api.Schedule != nil {
		schedule, err = stringToInt64(api.Schedule.Number)
		if err != nil {
			dg.AddError("Failed to convert schedule to int64", err.Error())
			return nil, dg
		}
	}

	var privateLocLabels []string
	for _, l := range api.Locations {
		if !l.IsServiceManaged {
			privateLocLabels = append(privateLocLabels, l.Label)
		}
	}

	timeout, err := stringToInt64(string(api.Timeout))
	if err != nil {
		dg.AddError("Failed to convert timeout to int64", err.Error())
		return nil, dg
	}

	var http *tfHTTPMonitorFieldsV0
	var tcp *tfTCPMonitorFieldsV0
	var icmp *tfICMPMonitorFieldsV0
	var browser *tfBrowserMonitorFieldsV0

	switch mType := api.Type; mType {
	case kbapi.Http:
		http = &tfHTTPMonitorFieldsV0{}
		if v.HTTP != nil {
			http = v.HTTP
		}
		http = http.toTfHTTPMonitorFieldsV0(ctx, dg, api)
	case kbapi.Tcp:
		tcp = &tfTCPMonitorFieldsV0{}
		if v.TCP != nil {
			tcp = v.TCP
		}
		tcp = tcp.toTfTCPMonitorFieldsV0(ctx, dg, api)
	case kbapi.Icmp:
		icmp = &tfICMPMonitorFieldsV0{}
		if v.ICMP != nil {
			icmp = v.ICMP
		}
		icmp, err = icmp.toTfICMPMonitorFieldsV0(api)
	case kbapi.Browser:
		browser = &tfBrowserMonitorFieldsV0{}
		if v.Browser != nil {
			browser = v.Browser
		}
		browser, err = browser.toTfBrowserMonitorFieldsV0(api)
	default:
		err = fmt.Errorf("unsupported monitor type: %s", mType)
	}

	if err != nil {
		dg.AddError("Failed to convert monitor fields", err.Error())
		return nil, dg
	}

	params := v.Params
	if api.Params != nil {
		params, err = toNormalizedValue(api.Params)
		if err != nil {
			dg.AddError("Failed to parse params", err.Error())
			return nil, dg
		}
	}

	resourceID := clients.CompositeId{
		ClusterId:  api.Namespace,
		ResourceId: string(api.Id),
	}

	return &tfModelV0{
		ID:               types.StringValue(resourceID.String()),
		Name:             types.StringValue(api.Name),
		SpaceID:          types.StringValue(api.Namespace),
		Schedule:         types.Int64Value(schedule),
		Locations:        v.Locations,
		PrivateLocations: StringSliceValue(privateLocLabels),
		Enabled:          types.BoolPointerValue(api.Enabled),
		Tags:             StringSliceValue(api.Tags),
		Alert:            toTfAlertConfigV0(api.Alert),
		APMServiceName:   types.StringValue(api.APMServiceName),
		TimeoutSeconds:   types.Int64Value(timeout),
		Params:           params,
		HTTP:             http,
		TCP:              tcp,
		ICMP:             icmp,
		Browser:          browser,
		RetestOnFailure:  v.RetestOnFailure,
	}, dg
}

func (v *tfTCPMonitorFieldsV0) toTfTCPMonitorFieldsV0(ctx context.Context, dg diag.Diagnostics, api *kbapi.SyntheticsMonitor) *tfTCPMonitorFieldsV0 {
	checkSend := v.CheckSend
	if api.CheckSend != "" {
		checkSend = types.StringValue(api.CheckSend)
	}
	checkReceive := v.CheckReceive
	if api.CheckReceive != "" {
		checkReceive = types.StringValue(api.CheckReceive)
	}
	sslSupportedProtocols := utils.SliceToListType_String(ctx, api.SslSupportedProtocols, path.Root("tcp").AtName("ssl_supported_protocols"), dg)
	if dg.HasError() {
		return nil
	}
	return &tfTCPMonitorFieldsV0{
		Host:                  types.StringValue(api.Host),
		SslVerificationMode:   types.StringValue(api.SslVerificationMode),
		SslSupportedProtocols: sslSupportedProtocols,
		CheckSend:             checkSend,
		CheckReceive:          checkReceive,
		ProxyURL:              types.StringValue(api.ProxyUrl),
		ProxyUseLocalResolver: types.BoolPointerValue(api.ProxyUseLocalResolver),
	}
}

func (v *tfICMPMonitorFieldsV0) toTfICMPMonitorFieldsV0(api *kbapi.SyntheticsMonitor) (*tfICMPMonitorFieldsV0, error) {
	wait, err := stringToInt64(string(api.Wait))
	if err != nil {
		return nil, err
	}
	return &tfICMPMonitorFieldsV0{
		Host: types.StringValue(api.Host),
		Wait: types.Int64Value(wait),
	}, nil
}

func (v *tfBrowserMonitorFieldsV0) toTfBrowserMonitorFieldsV0(api *kbapi.SyntheticsMonitor) (*tfBrowserMonitorFieldsV0, error) {

	var err error
	playwrightOptions := v.PlaywrightOptions
	if api.PlaywrightOptions != nil {
		playwrightOptions, err = toNormalizedValue(api.PlaywrightOptions)
		if err != nil {
			return nil, err
		}
	}

	syntheticsArgs := v.SyntheticsArgs
	if api.SyntheticsArgs != nil {
		syntheticsArgs = StringSliceValue(api.SyntheticsArgs)
	}

	inlineScript := v.InlineScript
	if api.InlineScript != "" {
		inlineScript = types.StringValue(api.InlineScript)
	}

	return &tfBrowserMonitorFieldsV0{
		InlineScript:      inlineScript,
		Screenshots:       types.StringValue(api.Screenshots),
		SyntheticsArgs:    syntheticsArgs,
		IgnoreHttpsErrors: types.BoolPointerValue(api.IgnoreHttpsErrors),
		PlaywrightOptions: playwrightOptions,
	}, nil
}

func (v *tfHTTPMonitorFieldsV0) toTfHTTPMonitorFieldsV0(ctx context.Context, dg diag.Diagnostics, api *kbapi.SyntheticsMonitor) *tfHTTPMonitorFieldsV0 {

	var err error
	proxyHeaders := v.ProxyHeader
	if api.ProxyHeaders != nil {
		proxyHeaders, err = toNormalizedValue(api.ProxyHeaders)
		if err != nil {
			dg.AddError("Failed to parse proxy_headers", err.Error())
			return nil
		}
	}

	username := v.Username
	if api.Username != "" {
		username = types.StringValue(api.Username)
	}
	password := v.Password
	if api.Password != "" {
		password = types.StringValue(api.Password)
	}

	maxRedirects, err := stringToInt64(api.MaxRedirects)
	if err != nil {
		dg.AddError("Failed to parse max_redirects", err.Error())
		return nil
	}

	sslSupportedProtocols := utils.SliceToListType_String(ctx, api.SslSupportedProtocols, path.Root("http").AtName("ssl_supported_protocols"), dg)

	if dg.HasError() {
		return nil
	}
	return &tfHTTPMonitorFieldsV0{
		URL:                   types.StringValue(api.Url),
		SslVerificationMode:   types.StringValue(api.SslVerificationMode),
		SslSupportedProtocols: sslSupportedProtocols,
		MaxRedirects:          types.Int64Value(maxRedirects),
		Mode:                  types.StringValue(string(api.Mode)),
		IPv4:                  types.BoolPointerValue(api.Ipv4),
		IPv6:                  types.BoolPointerValue(api.Ipv6),
		Username:              username,
		Password:              password,
		ProxyHeader:           proxyHeaders,
		ProxyURL:              types.StringValue(api.ProxyUrl),
		Check:                 v.Check,
		Response:              v.Response,
	}
}

func toTfAlertConfigV0(alert *kbapi.MonitorAlertConfig) *tfAlertConfigV0 {
	if alert == nil {
		return nil
	}
	return &tfAlertConfigV0{
		Status: toTfStatusConfigV0(alert.Status),
		TLS:    toTfStatusConfigV0(alert.Tls),
	}
}

func toTfStatusConfigV0(status *kbapi.SyntheticsStatusConfig) *tfStatusConfigV0 {
	if status == nil {
		return nil
	}
	return &tfStatusConfigV0{
		Enabled: types.BoolPointerValue(status.Enabled),
	}
}

func (v *tfModelV0) toKibanaAPIRequest(ctx context.Context) (*kibanaAPIRequest, diag.Diagnostics) {

	fields, dg := v.toMonitorFields(ctx)
	if dg.HasError() {
		return nil, dg
	}
	config, dg := v.toSyntheticsMonitorConfig()
	if dg.HasError() {
		return nil, dg
	}
	return &kibanaAPIRequest{
		fields: fields,
		config: *config,
	}, dg
}

func (v *tfModelV0) toMonitorFields(ctx context.Context) (kbapi.MonitorFields, diag.Diagnostics) {
	dg := diag.Diagnostics{}

	if v.HTTP != nil {
		return v.toHttpMonitorFields(ctx)
	} else if v.TCP != nil {
		return v.toTCPMonitorFields(ctx)
	} else if v.ICMP != nil {
		return v.toICMPMonitorFields(), dg
	} else if v.Browser != nil {
		return v.toBrowserMonitorFields()
	}

	dg.AddError("Unsupported monitor type config", "one of http,tcp monitor fields is required")
	return nil, dg
}

func (v *tfModelV0) toSyntheticsMonitorConfig() (*kbapi.SyntheticsMonitorConfig, diag.Diagnostics) {
	locations := Map[types.String, kbapi.MonitorLocation](v.Locations, func(s types.String) kbapi.MonitorLocation { return kbapi.MonitorLocation(s.ValueString()) })
	params, dg := toJsonObject(v.Params)
	if dg.HasError() {
		return nil, dg
	}

	var alert *kbapi.MonitorAlertConfig
	if v.Alert != nil {
		alert = v.Alert.toTfAlertConfigV0()
	}

	return &kbapi.SyntheticsMonitorConfig{
		Name:             v.Name.ValueString(),
		Schedule:         kbapi.MonitorSchedule(v.Schedule.ValueInt64()),
		Locations:        locations,
		PrivateLocations: ValueStringSlice(v.PrivateLocations),
		Enabled:          v.Enabled.ValueBoolPointer(),
		Tags:             ValueStringSlice(v.Tags),
		Alert:            alert,
		APMServiceName:   v.APMServiceName.ValueString(),
		TimeoutSeconds:   int(v.TimeoutSeconds.ValueInt64()),
		Namespace:        v.SpaceID.ValueString(),
		Params:           params,
		RetestOnFailure:  v.RetestOnFailure.ValueBoolPointer(),
	}, diag.Diagnostics{} //dg
}

func tfInt64ToString(v types.Int64) string {
	res := ""
	if !v.IsUnknown() && !v.IsNull() { // handle omitempty case
		return strconv.FormatInt(v.ValueInt64(), 10)
	}
	return res
}

func (v *tfModelV0) toHttpMonitorFields(ctx context.Context) (kbapi.MonitorFields, diag.Diagnostics) {
	proxyHeaders, dg := toJsonObject(v.HTTP.ProxyHeader)
	if dg.HasError() {
		return nil, dg
	}
	response, dg := toJsonObject(v.HTTP.Response)
	if dg.HasError() {
		return nil, dg
	}
	check, dg := toJsonObject(v.HTTP.Check)
	if dg.HasError() {
		return nil, dg
	}

	sslSupportedProtocols := utils.ListTypeToSlice_String(ctx, v.HTTP.SslSupportedProtocols, path.Root("http").AtName("ssl_supported_protocols"), dg)
	if dg.HasError() {
		return nil, dg
	}

	maxRedirects := tfInt64ToString(v.HTTP.MaxRedirects)
	return kbapi.HTTPMonitorFields{
		Url:                   v.HTTP.URL.ValueString(),
		SslVerificationMode:   v.HTTP.SslVerificationMode.ValueString(),
		SslSupportedProtocols: sslSupportedProtocols,
		MaxRedirects:          maxRedirects,
		Mode:                  kbapi.HttpMonitorMode(v.HTTP.Mode.ValueString()),
		Ipv4:                  v.HTTP.IPv4.ValueBoolPointer(),
		Ipv6:                  v.HTTP.IPv6.ValueBoolPointer(),
		Username:              v.HTTP.Username.ValueString(),
		Password:              v.HTTP.Password.ValueString(),
		ProxyHeader:           proxyHeaders,
		ProxyUrl:              v.HTTP.ProxyURL.ValueString(),
		Response:              response,
		Check:                 check,
	}, dg
}

func (v *tfModelV0) toTCPMonitorFields(ctx context.Context) (kbapi.MonitorFields, diag.Diagnostics) {
	dg := diag.Diagnostics{}
	sslSupportedProtocols := utils.ListTypeToSlice_String(ctx, v.TCP.SslSupportedProtocols, path.Root("tcp").AtName("ssl_supported_protocols"), dg)
	if dg.HasError() {
		return nil, dg
	}

	return kbapi.TCPMonitorFields{
		Host:                  v.TCP.Host.ValueString(),
		SslVerificationMode:   v.TCP.SslVerificationMode.ValueString(),
		SslSupportedProtocols: sslSupportedProtocols,
		CheckSend:             v.TCP.CheckSend.ValueString(),
		CheckReceive:          v.TCP.CheckReceive.ValueString(),
		ProxyUrl:              v.TCP.ProxyURL.ValueString(),
		ProxyUseLocalResolver: v.TCP.ProxyUseLocalResolver.ValueBoolPointer(),
	}, dg
}

func (v *tfModelV0) toICMPMonitorFields() kbapi.MonitorFields {
	return kbapi.ICMPMonitorFields{
		Host: v.ICMP.Host.ValueString(),
		Wait: tfInt64ToString(v.ICMP.Wait),
	}
}

func (v *tfModelV0) toBrowserMonitorFields() (kbapi.MonitorFields, diag.Diagnostics) {
	playwrightOptions, dg := toJsonObject(v.Browser.PlaywrightOptions)
	if dg.HasError() {
		return nil, dg
	}
	return kbapi.BrowserMonitorFields{
		InlineScript:      v.Browser.InlineScript.ValueString(),
		Screenshots:       kbapi.ScreenshotOption(v.Browser.Screenshots.ValueString()),
		SyntheticsArgs:    ValueStringSlice(v.Browser.SyntheticsArgs),
		IgnoreHttpsErrors: v.Browser.IgnoreHttpsErrors.ValueBoolPointer(),
		PlaywrightOptions: playwrightOptions,
	}, diag.Diagnostics{} //dg
}

func Map[T, U any](ts []T, f func(T) U) []U {
	var us []U
	for _, v := range ts {
		us = append(us, f(v))
	}
	return us
}

func (v tfAlertConfigV0) toTfAlertConfigV0() *kbapi.MonitorAlertConfig {
	var status *kbapi.SyntheticsStatusConfig
	if v.Status != nil {
		status = v.Status.toTfStatusConfigV0()
	}
	var tls *kbapi.SyntheticsStatusConfig
	if v.TLS != nil {
		tls = v.TLS.toTfStatusConfigV0()
	}
	return &kbapi.MonitorAlertConfig{
		Status: status,
		Tls:    tls,
	}
}

func (v tfStatusConfigV0) toTfStatusConfigV0() *kbapi.SyntheticsStatusConfig {
	return &kbapi.SyntheticsStatusConfig{
		Enabled: v.Enabled.ValueBoolPointer(),
	}
}
