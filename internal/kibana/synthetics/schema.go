package synthetics

import (
	"fmt"
	"github.com/disaster37/go-kibana-rest/v8/kbapi"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

//TODO: monitor support is from 8.14.0

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
	URL          types.String         `tfsdk:"url"`
	SSLSetting   jsontypes.Normalized `tfsdk:"ssl_setting"`
	MaxRedirects types.String         `tfsdk:"max_redirects"`
	Mode         types.String         `tfsdk:"mode"`
	IPv4         types.Bool           `tfsdk:"ipv4"`
	IPv6         types.Bool           `tfsdk:"ipv6"`
	Username     types.String         `tfsdk:"username"`
	Password     types.String         `tfsdk:"password"`
	ProxyHeader  jsontypes.Normalized `tfsdk:"proxy_header"`
	ProxyURL     types.String         `tfsdk:"proxy_url"`
	Response     jsontypes.Normalized `tfsdk:"response"`
	Check        jsontypes.Normalized `tfsdk:"check"`
}

type tfTCPMonitorFieldsV0 struct {
	Host                  types.String         `tfsdk:"host"`
	SSL                   jsontypes.Normalized `tfsdk:"ssl"`
	Check                 jsontypes.Normalized `tfsdk:"check"`
	ProxyURL              types.String         `tfsdk:"proxy_url"`
	ProxyUseLocalResolver types.Bool           `tfsdk:"proxy_use_local_resolver"`
}

type tfModelV0 struct {
	ID               types.String           `tfsdk:"id"`
	Name             types.String           `tfsdk:"name"`
	SpaceID          types.String           `tfsdk:"space_id"`
	Schedule         types.Int64            `tfsdk:"schedule"`
	Locations        []types.String         `tfsdk:"locations"`
	PrivateLocations []types.String         `tfsdk:"private_locations"`
	Enabled          types.Bool             `tfsdk:"enabled"`
	Tags             []types.String         `tfsdk:"tags"`
	Alert            *tfAlertConfigV0       `tfsdk:"alert"`
	ServiceName      types.String           `tfsdk:"service_name"`
	Timeout          types.Int64            `tfsdk:"timeout"`
	Params           jsontypes.Normalized   `tfsdk:"params"`
	RetestOnFailure  types.Bool             `tfsdk:"retest_on_failure"`
	HTTP             *tfHTTPMonitorFieldsV0 `tfsdk:"http"`
	TCP              *tfTCPMonitorFieldsV0  `tfsdk:"tcp"`
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
				},
			},
			"name": schema.StringAttribute{
				Optional:            false,
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
				Computed:            true,
				MarkdownDescription: "The monitor’s schedule in minutes. Supported values are 1, 3, 5, 10, 15, 30, 60, 120 and 240.",
				Validators: []validator.Int64{
					int64validator.OneOf(1, 3, 5, 10, 15, 30, 60, 120, 240),
				},
			},
			"locations": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "Where to deploy the monitor. Monitors can be deployed in multiple locations so that you can detect differences in availability and response times across those locations.",
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
			"params": jsonObjectSchema("Monitor parameters"),
			"retest_on_failure": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "Enable or disable retesting when a monitor fails. By default, monitors are automatically retested if the monitor goes from \"up\" to \"down\". If the result of the retest is also \"down\", an error will be created, and if configured, an alert sent. Then the monitor will resume running according to the defined schedule. Using retest_on_failure can reduce noise related to transient problems. Default: `true`.",
			},
			"http": httpMonitorFieldsSchema(),
			"tcp":  tcpMonitorFieldsSchema(),
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
		MarkdownDescription: "",
		Attributes: map[string]schema.Attribute{
			"url": schema.StringAttribute{
				Optional:            false,
				Required:            true,
				MarkdownDescription: "URL to monitor.",
			},
			"ssl_setting": jsonObjectSchema("The TLS/SSL connection settings for use with the HTTPS endpoint. If you don’t specify settings, the system defaults are used. See https://www.elastic.co/guide/en/beats/heartbeat/current/configuration-ssl.html for full SSL Options."),
			"max_redirects": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The maximum number of redirects to follow. Default: `0`",
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
			},
			"response": jsonObjectSchema("Controls the indexing of the HTTP response body contents to the `http.response.body.contents` field."),
			"check":    jsonObjectSchema("The check request settings."),
		},
	}
}

func tcpMonitorFieldsSchema() schema.Attribute {
	return schema.SingleNestedAttribute{
		Optional:            true,
		MarkdownDescription: "",
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				Optional:            false,
				Required:            true,
				MarkdownDescription: "The host to monitor; it can be an IP address or a hostname. The host can include the port using a colon (e.g., \"example.com:9200\").",
			},
			"ssl":   jsonObjectSchema(" The TLS/SSL connection settings for use with the HTTPS endpoint. If you don’t specify settings, the system defaults are used. See https://www.elastic.co/guide/en/beats/heartbeat/current/configuration-ssl.html for full SSL Options."),
			"check": jsonObjectSchema("An optional payload string to send to the remote host and the expected answer. If no payload is specified, the endpoint is assumed to be available if the connection attempt was successful. If send is specified without receive, any response is accepted as OK. If receive is specified without send, no payload is sent, but the client expects to receive a payload in the form of a \"hello message\" or \"banner\" on connect."),
			"proxy_url": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The URL of the SOCKS5 proxy to use when connecting to the server. The value must be a URL with a scheme of `socks5://`. If the SOCKS5 proxy server requires client authentication, then a username and password can be embedded in the URL. When using a proxy, hostnames are resolved on the proxy server instead of on the client. You can change this behavior by setting the `proxy_use_local_resolver` option.",
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

func (m *tfModelV0) toPrivateLocation() kibanaAPIRequest {
	return kibanaAPIRequest{
		fields: kbapi.HTTPMonitorFields{},
		config: kbapi.SyntheticsMonitorConfig{},
	}
}

func toModelV0(api kbapi.SyntheticsMonitor) tfModelV0 {
	return tfModelV0{}
}
