// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package monitor

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strconv"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	kibanaoapi "github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/synthetics"
	providerschema "github.com/elastic/terraform-provider-elasticstack/internal/schema"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

type tfStatusConfigV0 struct {
	Enabled types.Bool `tfsdk:"enabled"`
}

type tfAlertConfigV0 struct {
	Status *tfStatusConfigV0 `tfsdk:"status"`
	TLS    *tfStatusConfigV0 `tfsdk:"tls"`
}

type tfSSLConfig struct {
	SslVerificationMode       types.String   `tfsdk:"ssl_verification_mode"`
	SslSupportedProtocols     types.List     `tfsdk:"ssl_supported_protocols"`
	SslCertificateAuthorities []types.String `tfsdk:"ssl_certificate_authorities"`
	SslCertificate            types.String   `tfsdk:"ssl_certificate"`
	SslKey                    types.String   `tfsdk:"ssl_key"`
	SslKeyPassphrase          types.String   `tfsdk:"ssl_key_passphrase"`
}

type tfHTTPMonitorFieldsV0 struct {
	URL          types.String         `tfsdk:"url"`
	MaxRedirects types.Int64          `tfsdk:"max_redirects"`
	Mode         types.String         `tfsdk:"mode"`
	IPv4         types.Bool           `tfsdk:"ipv4"`
	IPv6         types.Bool           `tfsdk:"ipv6"`
	ProxyURL     types.String         `tfsdk:"proxy_url"`
	ProxyHeader  jsontypes.Normalized `tfsdk:"proxy_header"`
	Username     types.String         `tfsdk:"username"`
	Password     types.String         `tfsdk:"password"`
	Response     jsontypes.Normalized `tfsdk:"response"`
	Check        jsontypes.Normalized `tfsdk:"check"`

	tfSSLConfig
}

type tfTCPMonitorFieldsV0 struct {
	Host                  types.String `tfsdk:"host"`
	CheckSend             types.String `tfsdk:"check_send"`
	CheckReceive          types.String `tfsdk:"check_receive"`
	ProxyURL              types.String `tfsdk:"proxy_url"`
	ProxyUseLocalResolver types.Bool   `tfsdk:"proxy_use_local_resolver"`

	tfSSLConfig
}

type tfICMPMonitorFieldsV0 struct {
	Host types.String `tfsdk:"host"`
	Wait types.Int64  `tfsdk:"wait"`
}

type tfBrowserMonitorFieldsV0 struct {
	InlineScript      types.String         `tfsdk:"inline_script"`
	Screenshots       types.String         `tfsdk:"screenshots"`
	SyntheticsArgs    []types.String       `tfsdk:"synthetics_args"`
	IgnoreHTTPSErrors types.Bool           `tfsdk:"ignore_https_errors"`
	PlaywrightOptions jsontypes.Normalized `tfsdk:"playwright_options"`
}

type tfModelV0 struct {
	ID               types.String              `tfsdk:"id"`
	KibanaConnection types.List                `tfsdk:"kibana_connection"`
	Name             types.String              `tfsdk:"name"`
	SpaceID          types.String              `tfsdk:"space_id"`
	Namespace        types.String              `tfsdk:"namespace"`
	Schedule         types.Int64               `tfsdk:"schedule"`
	Locations        []types.String            `tfsdk:"locations"`
	PrivateLocations []types.String            `tfsdk:"private_locations"`
	Enabled          types.Bool                `tfsdk:"enabled"`
	Tags             []types.String            `tfsdk:"tags"`
	Labels           types.Map                 `tfsdk:"labels"`
	Alert            types.Object              `tfsdk:"alert"` // tfAlertConfigV0
	APMServiceName   types.String              `tfsdk:"service_name"`
	TimeoutSeconds   types.Int64               `tfsdk:"timeout"`
	HTTP             *tfHTTPMonitorFieldsV0    `tfsdk:"http"`
	TCP              *tfTCPMonitorFieldsV0     `tfsdk:"tcp"`
	ICMP             *tfICMPMonitorFieldsV0    `tfsdk:"icmp"`
	Browser          *tfBrowserMonitorFieldsV0 `tfsdk:"browser"`
	Params           jsontypes.Normalized      `tfsdk:"params"`
	RetestOnFailure  types.Bool                `tfsdk:"retest_on_failure"`
}

//go:embed resource-description.md
var monitorDescription string

// skipLocationValidationEnvVar: when set to "true" at validate time, managed location enum
// validation is skipped (e.g. acceptance tests using t.Setenv after provider init).
const skipLocationValidationEnvVar = "TF_ELASTICSTACK_SKIP_LOCATION_VALIDATION"

var managedElasticLocationOneOf = stringvalidator.OneOf(
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
)

type managedLocationStringValidator struct{}

func (v managedLocationStringValidator) Description(ctx context.Context) string {
	return managedElasticLocationOneOf.Description(ctx)
}

func (v managedLocationStringValidator) MarkdownDescription(ctx context.Context) string {
	return managedElasticLocationOneOf.MarkdownDescription(ctx)
}

func (v managedLocationStringValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if os.Getenv(skipLocationValidationEnvVar) == "true" {
		return
	}
	managedElasticLocationOneOf.ValidateString(ctx, req, resp)
}

func locationValidators() []validator.List {
	return []validator.List{
		listvalidator.ValueStringsAre(managedLocationStringValidator{}),
	}
}

func monitorConfigSchema() schema.Schema {
	return schema.Schema{
		MarkdownDescription: monitorDescription,
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
				MarkdownDescription: "The monitor's name.",
			},
			"space_id": schema.StringAttribute{
				MarkdownDescription: spaceIDDescription,
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
				Computed: true,
			},
			"namespace": schema.StringAttribute{
				MarkdownDescription: namespaceDescription,
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Computed: true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^[^*\\/?\"<>|\s,#:-]*$`),
						"namespace must not contain any of the following characters: *, \\, /, ?, \", <, >, |, whitespace, ,, #, :, or -",
					),
				},
			},
			"schedule": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "The monitor's schedule in minutes. Supported values are 1, 3, 5, 10, 15, 30, 60, 120 and 240.",
				Validators: []validator.Int64{
					int64validator.OneOf(1, 3, 5, 10, 15, 30, 60, 120, 240),
				},
				Computed:      true,
				PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
			},
			"locations": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "Where to deploy the monitor. Monitors can be deployed in multiple locations so that you can detect differences in availability and response times across those locations.",
				Validators:          locationValidators(),
			},
			"private_locations": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "These Private Locations refer to locations hosted and managed by you, whereas locations are hosted by Elastic. You can specify a Private Location using the location's name.",
			},
			"enabled": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "Whether the monitor is enabled. Default: `true`",
				Computed:            true,
				PlanModifiers:       []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
			},
			"tags": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "An array of tags.",
			},
			"labels": schema.MapAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "Key-value pairs of labels to associate with the monitor. Labels can be used for filtering and grouping monitors.",
			},
			"alert": monitorAlertConfigSchema(),
			"service_name": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The APM service name.",
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"timeout": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "The monitor timeout in seconds, monitor will fail if it doesn't complete within this time. Default: `16`",
				Computed:            true,
				PlanModifiers:       []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
			},
			"params":  jsonObjectSchema("Monitor parameters"),
			"http":    httpMonitorFieldsSchema(),
			"tcp":     tcpMonitorFieldsSchema(),
			"icmp":    icmpMonitorFieldsSchema(),
			"browser": browserMonitorFieldsSchema(),
			"retest_on_failure": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: retestOnFailureDescription,
			},
		},

		Blocks: map[string]schema.Block{
			"kibana_connection": providerschema.GetKbFWConnectionBlock(),
		}}
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
				Optional:      true,
				Computed:      true,
				PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
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
		Computed:      true,
		PlanModifiers: []planmodifier.Object{objectplanmodifier.UseStateForUnknown()},
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
			"ssl_certificate_authorities": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "The list of root certificates for verifications is required.",
			},
			"ssl_certificate": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Certificate.",
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"ssl_key": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Certificate key.",
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
				Sensitive:           true,
			},
			"ssl_key_passphrase": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Key passphrase.",
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
				Sensitive:           true,
			},
			"max_redirects": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "The maximum number of redirects to follow. Default: `0`",
				PlanModifiers:       []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
				Computed:            true,
			},
			"mode": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The mode of the monitor. Can be \"all\" or \"any\". If you're using a DNS-load balancer and want to ping every IP address for the specified hostname, you should use all.",
				Validators: []validator.String{
					stringvalidator.OneOf("any", "all"),
				},
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"ipv4": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "Whether to ping using the ipv4 protocol.",
				Computed:            true,
				PlanModifiers:       []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
			},
			"ipv6": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "Whether to ping using the ipv6 protocol.",
				Computed:            true,
				PlanModifiers:       []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
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
			"ssl_certificate_authorities": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "The list of root certificates for verifications is required.",
			},
			"ssl_certificate": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Certificate.",
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"ssl_key": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Certificate key.",
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
				Sensitive:           true,
			},
			"ssl_key_passphrase": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Key passphrase.",
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
				Sensitive:           true,
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
				MarkdownDescription: proxyURLDescription,
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"proxy_use_local_resolver": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: proxyUseLocalResolverDescription,
				Computed:            true,
				PlanModifiers:       []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
			},
		},
	}
}

func toNormalizedValue(jsObj map[string]any) (jsontypes.Normalized, error) {
	res, err := json.Marshal(jsObj)
	if err != nil {
		return jsontypes.NewNormalizedUnknown(), err
	}
	return jsontypes.NewNormalizedValue(string(res)), nil
}

func toJSONObject(v jsontypes.Normalized) (map[string]any, diag.Diagnostics) {
	if v.IsNull() {
		return nil, diag.Diagnostics{}
	}
	var res map[string]any
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

func (v *tfModelV0) toModelV0(ctx context.Context, api *kibanaoapi.SyntheticsMonitor, space string) (*tfModelV0, diag.Diagnostics) {
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

	var httpFields *tfHTTPMonitorFieldsV0
	var tcp *tfTCPMonitorFieldsV0
	var icmp *tfICMPMonitorFieldsV0
	var browser *tfBrowserMonitorFieldsV0

	switch mType := api.Type; mType {
	case kibanaoapi.SyntheticsMonitorTypeHTTP:
		httpFields = &tfHTTPMonitorFieldsV0{}
		if v.HTTP != nil {
			httpFields = v.HTTP
		}
		httpFields = httpFields.toTfHTTPMonitorFieldsV0(ctx, dg, api)
	case kibanaoapi.SyntheticsMonitorTypeTCP:
		tcp = &tfTCPMonitorFieldsV0{}
		if v.TCP != nil {
			tcp = v.TCP
		}
		tcp = tcp.toTfTCPMonitorFieldsV0(ctx, dg, api)
	case kibanaoapi.SyntheticsMonitorTypeICMP:
		icmp = &tfICMPMonitorFieldsV0{}
		if v.ICMP != nil {
			icmp = v.ICMP
		}
		icmp, err = icmp.toTfICMPMonitorFieldsV0(api)
	case kibanaoapi.SyntheticsMonitorTypeBrowser:
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

	resourceID := clients.CompositeID{
		ClusterID:  space,
		ResourceID: api.ID,
	}

	alertV0, dg := toTfAlertConfigV0(ctx, api.Alert)
	if dg.HasError() {
		return nil, dg
	}

	return &tfModelV0{
		ID:               types.StringValue(resourceID.String()),
		Name:             types.StringValue(api.Name),
		SpaceID:          types.StringValue(space),
		Namespace:        types.StringValue(api.Namespace),
		Schedule:         types.Int64Value(schedule),
		Locations:        v.Locations,
		PrivateLocations: synthetics.StringSliceValue(privateLocLabels),
		Enabled:          types.BoolPointerValue(api.Enabled),
		Tags:             synthetics.StringSliceValue(api.Tags),
		Labels:           typeutils.MapValueFrom(ctx, api.Labels, types.StringType, path.Root("labels"), &dg),
		Alert:            alertV0,
		APMServiceName:   types.StringValue(api.APMServiceName),
		TimeoutSeconds:   types.Int64Value(timeout),
		Params:           params,
		HTTP:             httpFields,
		TCP:              tcp,
		ICMP:             icmp,
		Browser:          browser,
		RetestOnFailure:  v.RetestOnFailure,
		KibanaConnection: v.KibanaConnection,
	}, dg
}

func (v *tfTCPMonitorFieldsV0) toTfTCPMonitorFieldsV0(ctx context.Context, dg diag.Diagnostics, api *kibanaoapi.SyntheticsMonitor) *tfTCPMonitorFieldsV0 {
	checkSend := v.CheckSend
	if api.CheckSend != "" {
		checkSend = types.StringValue(api.CheckSend)
	}
	checkReceive := v.CheckReceive
	if api.CheckReceive != "" {
		checkReceive = types.StringValue(api.CheckReceive)
	}
	sslCfg, dg := toTFSSLConfig(ctx, dg, api, "tcp")

	if dg.HasError() {
		return nil
	}
	return &tfTCPMonitorFieldsV0{
		Host:                  types.StringValue(api.Host),
		CheckSend:             checkSend,
		CheckReceive:          checkReceive,
		ProxyURL:              types.StringValue(api.ProxyURL),
		ProxyUseLocalResolver: types.BoolPointerValue(api.ProxyUseLocalResolver),
		tfSSLConfig:           sslCfg,
	}
}

func (v *tfICMPMonitorFieldsV0) toTfICMPMonitorFieldsV0(api *kibanaoapi.SyntheticsMonitor) (*tfICMPMonitorFieldsV0, error) {
	wait, err := stringToInt64(string(api.Wait))
	if err != nil {
		return nil, err
	}
	return &tfICMPMonitorFieldsV0{
		Host: types.StringValue(api.Host),
		Wait: types.Int64Value(wait),
	}, nil
}

func (v *tfBrowserMonitorFieldsV0) toTfBrowserMonitorFieldsV0(api *kibanaoapi.SyntheticsMonitor) (*tfBrowserMonitorFieldsV0, error) {

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
		syntheticsArgs = synthetics.StringSliceValue(api.SyntheticsArgs)
	}

	inlineScript := v.InlineScript
	if api.InlineScript != "" {
		inlineScript = types.StringValue(api.InlineScript)
	}

	return &tfBrowserMonitorFieldsV0{
		InlineScript:      inlineScript,
		Screenshots:       types.StringValue(api.Screenshots),
		SyntheticsArgs:    syntheticsArgs,
		IgnoreHTTPSErrors: types.BoolPointerValue(api.IgnoreHTTPSErrors),
		PlaywrightOptions: playwrightOptions,
	}, nil
}

func (v *tfHTTPMonitorFieldsV0) toTfHTTPMonitorFieldsV0(ctx context.Context, dg diag.Diagnostics, api *kibanaoapi.SyntheticsMonitor) *tfHTTPMonitorFieldsV0 {

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

	sslCfg, dg := toTFSSLConfig(ctx, dg, api, "http")
	if dg.HasError() {
		return nil
	}
	return &tfHTTPMonitorFieldsV0{
		URL:          types.StringValue(api.URL),
		MaxRedirects: types.Int64Value(maxRedirects),
		Mode:         types.StringValue(api.Mode),
		IPv4:         types.BoolPointerValue(api.Ipv4),
		IPv6:         types.BoolPointerValue(api.Ipv6),
		Username:     username,
		Password:     password,
		ProxyHeader:  proxyHeaders,
		ProxyURL:     types.StringValue(api.ProxyURL),
		Check:        v.Check,
		Response:     v.Response,
		tfSSLConfig:  sslCfg,
	}
}

func toTFSSLConfig(ctx context.Context, dg diag.Diagnostics, api *kibanaoapi.SyntheticsMonitor, p string) (tfSSLConfig, diag.Diagnostics) {
	sslSupportedProtocols := typeutils.SliceToListTypeString(ctx, api.SslSupportedProtocols, path.Root(p).AtName("ssl_supported_protocols"), &dg)
	return tfSSLConfig{
		SslVerificationMode:       types.StringValue(api.SslVerificationMode),
		SslSupportedProtocols:     sslSupportedProtocols,
		SslCertificateAuthorities: synthetics.StringSliceValue(api.SslCertificateAuthorities),
		SslCertificate:            types.StringValue(api.SslCertificate),
		SslKey:                    types.StringValue(api.SslKey),
		SslKeyPassphrase:          types.StringValue(api.SslKeyPassphrase),
	}, dg
}

func toTfAlertConfigV0(ctx context.Context, alert *kibanaoapi.SyntheticsMonitorAlert) (basetypes.ObjectValue, diag.Diagnostics) {

	dg := diag.Diagnostics{}

	alertAttributes := monitorAlertConfigSchema().GetType().(attr.TypeWithAttributeTypes).AttributeTypes()

	var emptyAttr = map[string]attr.Type(nil)

	if alert == nil {
		return basetypes.NewObjectNull(emptyAttr), dg
	}

	tfAlertConfig := tfAlertConfigV0{
		Status: toTfStatusConfigV0(alert.Status),
		TLS:    toTfStatusConfigV0(alert.TLS),
	}

	return types.ObjectValueFrom(ctx, alertAttributes, &tfAlertConfig)
}

func toTfStatusConfigV0(status *kibanaoapi.SyntheticsMonitorAlertStatus) *tfStatusConfigV0 {
	if status == nil {
		return nil
	}
	return &tfStatusConfigV0{
		Enabled: types.BoolPointerValue(status.Enabled),
	}
}

func (v *tfModelV0) toKibanaAPIRequest(ctx context.Context) (*kibanaoapi.SyntheticsMonitorRequest, diag.Diagnostics) {
	params, dg := toJSONObject(v.Params)
	if dg.HasError() {
		return nil, dg
	}

	labels := typeutils.MapTypeAs[string](ctx, v.Labels, path.Root("labels"), &dg)
	if dg.HasError() {
		return nil, dg
	}
	if labels == nil {
		labels = map[string]string{}
	}

	locations := Map[types.String, string](v.Locations, func(s types.String) string { return s.ValueString() })

	req := &kibanaoapi.SyntheticsMonitorRequest{
		Name:             v.Name.ValueString(),
		Schedule:         v.Schedule.ValueInt64(),
		Locations:        locations,
		PrivateLocations: synthetics.ValueStringSlice(v.PrivateLocations),
		Enabled:          v.Enabled.ValueBoolPointer(),
		Tags:             synthetics.ValueStringSlice(v.Tags),
		Labels:           labels,
		Alert:            toAPIAlertConfig(ctx, v.Alert),
		APMServiceName:   v.APMServiceName.ValueString(),
		TimeoutSeconds:   int(v.TimeoutSeconds.ValueInt64()),
		Namespace:        v.Namespace.ValueString(),
		Params:           params,
		RetestOnFailure:  v.RetestOnFailure.ValueBoolPointer(),
	}

	dg = v.populateTypeFields(ctx, req, dg)
	if dg.HasError() {
		return nil, dg
	}

	return req, dg
}

func (v *tfModelV0) populateTypeFields(ctx context.Context, req *kibanaoapi.SyntheticsMonitorRequest, dg diag.Diagnostics) diag.Diagnostics {
	switch {
	case v.HTTP != nil:
		return v.populateHTTPFields(ctx, req, dg)
	case v.TCP != nil:
		return v.populateTCPFields(ctx, req, dg)
	case v.ICMP != nil:
		v.populateICMPFields(req)
		return dg
	case v.Browser != nil:
		dg.Append(v.populateBrowserFields(req)...)
		return dg
	}
	dg.AddError("Unsupported monitor type config", "one of http,tcp,icmp,browser monitor fields is required")
	return dg
}

func toAPIAlertConfig(ctx context.Context, v basetypes.ObjectValue) *kibanaoapi.SyntheticsMonitorAlert {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	tfAlert := tfAlertConfigV0{}
	tfsdk.ValueAs(ctx, v, &tfAlert)
	return tfAlert.toAPIAlertConfig()
}

func tfInt64ToString(v types.Int64) string {
	res := ""
	if !v.IsUnknown() && !v.IsNull() { // handle omitempty case
		return strconv.FormatInt(v.ValueInt64(), 10)
	}
	return res
}

func toSSLConfig(ctx context.Context, dg diag.Diagnostics, v tfSSLConfig, p string) (*kibanaoapi.SyntheticsSSLConfig, diag.Diagnostics) {

	var ssl *kibanaoapi.SyntheticsSSLConfig
	if !v.SslSupportedProtocols.IsNull() && !v.SslSupportedProtocols.IsUnknown() {
		sslSupportedProtocols := typeutils.ListTypeToSliceString(ctx, v.SslSupportedProtocols, path.Root(p).AtName("ssl_supported_protocols"), &dg)
		if dg.HasError() {
			return nil, dg
		}
		ssl = &kibanaoapi.SyntheticsSSLConfig{}
		ssl.SupportedProtocols = sslSupportedProtocols
	}

	if !v.SslVerificationMode.IsNull() && !v.SslVerificationMode.IsUnknown() {
		if ssl == nil {
			ssl = &kibanaoapi.SyntheticsSSLConfig{}
		}
		ssl.VerificationMode = v.SslVerificationMode.ValueString()
	}

	certAuths := synthetics.ValueStringSlice(v.SslCertificateAuthorities)
	if len(certAuths) > 0 {
		if ssl == nil {
			ssl = &kibanaoapi.SyntheticsSSLConfig{}
		}
		ssl.CertificateAuthorities = certAuths
	}

	if !v.SslCertificate.IsUnknown() && !v.SslCertificate.IsNull() {
		if ssl == nil {
			ssl = &kibanaoapi.SyntheticsSSLConfig{}
		}
		ssl.Certificate = v.SslCertificate.ValueString()
	}

	if !v.SslKey.IsUnknown() && !v.SslKey.IsNull() {
		if ssl == nil {
			ssl = &kibanaoapi.SyntheticsSSLConfig{}
		}
		ssl.Key = v.SslKey.ValueString()
	}

	if !v.SslKeyPassphrase.IsUnknown() && !v.SslKeyPassphrase.IsNull() {
		if ssl == nil {
			ssl = &kibanaoapi.SyntheticsSSLConfig{}
		}
		ssl.KeyPassphrase = v.SslKeyPassphrase.ValueString()
	}
	return ssl, dg
}

func (v *tfModelV0) populateHTTPFields(ctx context.Context, req *kibanaoapi.SyntheticsMonitorRequest, dg diag.Diagnostics) diag.Diagnostics {
	h := v.HTTP
	proxyHeaders, d := toJSONObject(h.ProxyHeader)
	dg.Append(d...)
	if dg.HasError() {
		return dg
	}
	response, d := toJSONObject(h.Response)
	dg.Append(d...)
	if dg.HasError() {
		return dg
	}
	check, d := toJSONObject(h.Check)
	dg.Append(d...)
	if dg.HasError() {
		return dg
	}

	ssl, d := toSSLConfig(ctx, dg, h.tfSSLConfig, "http")
	dg.Append(d...)
	if dg.HasError() {
		return dg
	}

	req.Type = kibanaoapi.SyntheticsMonitorTypeHTTP
	req.URL = h.URL.ValueString()
	req.Ssl = ssl
	req.MaxRedirects = tfInt64ToString(h.MaxRedirects)
	req.Mode = h.Mode.ValueString()
	req.Ipv4 = h.IPv4.ValueBoolPointer()
	req.Ipv6 = h.IPv6.ValueBoolPointer()
	req.Username = h.Username.ValueString()
	req.Password = h.Password.ValueString()
	req.ProxyHeader = proxyHeaders
	req.ProxyURL = h.ProxyURL.ValueString()
	req.Response = response
	req.Check = check
	return dg
}

func (v *tfModelV0) populateTCPFields(ctx context.Context, req *kibanaoapi.SyntheticsMonitorRequest, dg diag.Diagnostics) diag.Diagnostics {
	tcp := v.TCP
	ssl, d := toSSLConfig(ctx, dg, tcp.tfSSLConfig, "tcp")
	dg.Append(d...)
	if dg.HasError() {
		return dg
	}

	req.Type = kibanaoapi.SyntheticsMonitorTypeTCP
	req.Host = tcp.Host.ValueString()
	req.CheckSend = tcp.CheckSend.ValueString()
	req.CheckReceive = tcp.CheckReceive.ValueString()
	req.ProxyURL = tcp.ProxyURL.ValueString()
	req.ProxyUseLocalResolver = tcp.ProxyUseLocalResolver.ValueBoolPointer()
	req.Ssl = ssl
	return dg
}

func (v *tfModelV0) populateICMPFields(req *kibanaoapi.SyntheticsMonitorRequest) {
	req.Type = kibanaoapi.SyntheticsMonitorTypeICMP
	req.Host = v.ICMP.Host.ValueString()
	req.Wait = tfInt64ToString(v.ICMP.Wait)
}

func (v *tfModelV0) populateBrowserFields(req *kibanaoapi.SyntheticsMonitorRequest) diag.Diagnostics {
	playwrightOptions, dg := toJSONObject(v.Browser.PlaywrightOptions)
	if dg.HasError() {
		return dg
	}

	req.Type = kibanaoapi.SyntheticsMonitorTypeBrowser
	req.InlineScript = v.Browser.InlineScript.ValueString()
	req.Screenshots = v.Browser.Screenshots.ValueString()
	req.SyntheticsArgs = synthetics.ValueStringSlice(v.Browser.SyntheticsArgs)
	req.IgnoreHTTPSErrors = v.Browser.IgnoreHTTPSErrors.ValueBoolPointer()
	req.PlaywrightOptions = playwrightOptions
	return dg
}

func Map[T, U any](ts []T, f func(T) U) []U {
	var us []U
	for _, v := range ts {
		us = append(us, f(v))
	}
	return us
}

func (v tfAlertConfigV0) toAPIAlertConfig() *kibanaoapi.SyntheticsMonitorAlert {
	var status *kibanaoapi.SyntheticsMonitorAlertStatus
	if v.Status != nil {
		status = v.Status.toAPIAlertStatus()
	}
	var tls *kibanaoapi.SyntheticsMonitorAlertStatus
	if v.TLS != nil {
		tls = v.TLS.toAPIAlertStatus()
	}
	return &kibanaoapi.SyntheticsMonitorAlert{
		Status: status,
		TLS:    tls,
	}
}

func (v tfStatusConfigV0) toAPIAlertStatus() *kibanaoapi.SyntheticsMonitorAlertStatus {
	return &kibanaoapi.SyntheticsMonitorAlertStatus{
		Enabled: v.Enabled.ValueBoolPointer(),
	}
}

func (v tfModelV0) enforceVersionConstraints(ctx context.Context, client *clients.KibanaScopedClient) diag.Diagnostics {
	if typeutils.IsKnown(v.Labels) {
		isSupported, sdkDiags := client.EnforceMinVersion(ctx, MinLabelsVersion)
		diags := diagutil.FrameworkDiagsFromSDK(sdkDiags)
		if diags.HasError() {
			return diags
		}

		if !isSupported {
			diags.AddAttributeError(
				path.Root("labels"),
				"Unsupported version for `labels` attribute",
				fmt.Sprintf("The `labels` attribute requires server version %s or higher. Either remove the `labels` attribute or upgrade your Elastic Stack installation.", MinLabelsVersion.String()),
			)
			return diags
		}
	}

	return nil
}
