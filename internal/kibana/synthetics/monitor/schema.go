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
	"math"
	"os"
	"regexp"
	"strconv"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
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

// unionToInt64 converts a string-or-float32 union type to int64 using the provided accessor callbacks.
// The float variant must represent a whole number; fractional values are rejected.
func unionToInt64(asStr func() (string, error), asFloat func() (float32, error), fieldName string) (int64, error) {
	if s, err := asStr(); err == nil {
		return stringToInt64(s)
	}

	if n, err := asFloat(); err == nil {
		if math.Trunc(float64(n)) != float64(n) {
			return 0, fmt.Errorf("%s must be a whole number, got %v", fieldName, n)
		}
		return int64(n), nil
	}

	return 0, fmt.Errorf("%s has unsupported type", fieldName)
}

func syntheticsMonitorTimeoutToInt64(v *kbapi.SyntheticsMonitor_Timeout) (int64, error) {
	if v == nil {
		return 0, nil
	}
	return unionToInt64(v.AsSyntheticsMonitorTimeout0, v.AsSyntheticsMonitorTimeout1, "timeout")
}

func syntheticsMonitorWaitToInt64(v *kbapi.SyntheticsMonitor_Wait) (int64, error) {
	if v == nil {
		return 0, nil
	}
	return unionToInt64(v.AsSyntheticsMonitorWait0, v.AsSyntheticsMonitorWait1, "wait")
}

func syntheticsMonitorMaxRedirectsToInt64(v *kbapi.SyntheticsMonitor_MaxRedirects) (int64, error) {
	if v == nil {
		return 0, nil
	}
	return unionToInt64(v.AsSyntheticsMonitorMaxRedirects0, v.AsSyntheticsMonitorMaxRedirects1, "max_redirects")
}

func (v *tfModelV0) toModelV0(ctx context.Context, api *kbapi.SyntheticsMonitor, space string) (*tfModelV0, diag.Diagnostics) {
	var (
		schedule int64
		timeout  int64
		err      error
	)
	dg := diag.Diagnostics{}
	if api.Schedule != nil && api.Schedule.Number != nil {
		schedule, err = stringToInt64(*api.Schedule.Number)
		if err != nil {
			dg.AddError("Failed to convert schedule to int64", err.Error())
			return nil, dg
		}
	}

	var privateLocLabels []string
	if api.Locations != nil {
		for _, l := range *api.Locations {
			if l.IsServiceManaged != nil && *l.IsServiceManaged {
				continue
			}
			if l.Label != nil {
				privateLocLabels = append(privateLocLabels, *l.Label)
			}
		}
	}

	if api.Timeout != nil {
		timeout, err = syntheticsMonitorTimeoutToInt64(api.Timeout)
		if err != nil {
			dg.AddError("Failed to convert timeout to int64", err.Error())
			return nil, dg
		}
	}

	var httpFields *tfHTTPMonitorFieldsV0
	var tcp *tfTCPMonitorFieldsV0
	var icmp *tfICMPMonitorFieldsV0
	var browser *tfBrowserMonitorFieldsV0
	monitorType := kbapi.SyntheticsMonitorType("")
	if api.Type != nil {
		monitorType = *api.Type
	}

	switch monitorType {
	case kbapi.SyntheticsMonitorTypeHttp:
		httpFields = &tfHTTPMonitorFieldsV0{}
		if v.HTTP != nil {
			httpFields = v.HTTP
		}
		httpFields = httpFields.toTfHTTPMonitorFieldsV0(ctx, dg, api)
	case kbapi.SyntheticsMonitorTypeTcp:
		tcp = &tfTCPMonitorFieldsV0{}
		if v.TCP != nil {
			tcp = v.TCP
		}
		tcp = tcp.toTfTCPMonitorFieldsV0(ctx, dg, api)
	case kbapi.SyntheticsMonitorTypeIcmp:
		icmp = &tfICMPMonitorFieldsV0{}
		if v.ICMP != nil {
			icmp = v.ICMP
		}
		icmp, err = icmp.toTfICMPMonitorFieldsV0(api)
	case kbapi.SyntheticsMonitorTypeBrowser:
		browser = &tfBrowserMonitorFieldsV0{}
		if v.Browser != nil {
			browser = v.Browser
		}
		browser, err = browser.toTfBrowserMonitorFieldsV0(api)
	default:
		err = fmt.Errorf("unsupported monitor type: %s", monitorType)
	}

	if err != nil {
		dg.AddError("Failed to convert monitor fields", err.Error())
		return nil, dg
	}

	params := v.Params
	if api.Params != nil {
		params, err = toNormalizedValue(*api.Params)
		if err != nil {
			dg.AddError("Failed to parse params", err.Error())
			return nil, dg
		}
	}

	resourceID := clients.CompositeID{
		ClusterID:  space,
		ResourceID: typeutils.Deref(api.Id),
	}

	alertV0, dg := toTfAlertConfigV0(ctx, api.Alert)
	if dg.HasError() {
		return nil, dg
	}

	return &tfModelV0{
		ID:        types.StringValue(resourceID.String()),
		Name:      types.StringPointerValue(api.Name),
		SpaceID:   types.StringValue(space),
		Namespace: types.StringPointerValue(api.Namespace),
		Schedule:  types.Int64Value(schedule),
		// Locations (managed/service-managed) are preserved from prior state per REQ-015:
		// the Kibana API returns location objects with both ID and Label, but the provider
		// spec requires using the practitioner-configured identifiers (from state) rather
		// than re-deriving from the API response. Private locations are re-derived from
		// the API response where IsServiceManaged == false.
		Locations:        v.Locations,
		PrivateLocations: synthetics.StringSliceValue(privateLocLabels),
		Enabled:          types.BoolPointerValue(api.Enabled),
		Tags:             synthetics.StringSliceValue(typeutils.Deref(api.Tags)),
		Labels:           typeutils.MapValueFrom(ctx, typeutils.Deref(api.Labels), types.StringType, path.Root("labels"), &dg),
		Alert:            alertV0,
		APMServiceName:   types.StringPointerValue(api.ServiceName),
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

func (v *tfTCPMonitorFieldsV0) toTfTCPMonitorFieldsV0(ctx context.Context, dg diag.Diagnostics, api *kbapi.SyntheticsMonitor) *tfTCPMonitorFieldsV0 {
	checkSend := v.CheckSend
	if api.CheckSend != nil {
		checkSend = types.StringPointerValue(api.CheckSend)
	}
	checkReceive := v.CheckReceive
	if api.CheckReceive != nil {
		checkReceive = types.StringPointerValue(api.CheckReceive)
	}
	proxyURL := types.StringValue("")
	if !v.ProxyURL.IsNull() && !v.ProxyURL.IsUnknown() {
		proxyURL = v.ProxyURL
	}
	if api.ProxyUrl != nil {
		proxyURL = types.StringPointerValue(api.ProxyUrl)
	}
	sslCfg, dg := toTFSSLConfig(ctx, dg, api, "tcp")

	if dg.HasError() {
		return nil
	}
	return &tfTCPMonitorFieldsV0{
		Host:                  types.StringPointerValue(api.Host),
		CheckSend:             checkSend,
		CheckReceive:          checkReceive,
		ProxyURL:              proxyURL,
		ProxyUseLocalResolver: types.BoolPointerValue(api.ProxyUseLocalResolver),
		tfSSLConfig:           sslCfg,
	}
}

func (v *tfICMPMonitorFieldsV0) toTfICMPMonitorFieldsV0(
	api *kbapi.SyntheticsMonitor,
) (*tfICMPMonitorFieldsV0, error) {
	var wait int64
	if api.Wait != nil {
		var err error
		wait, err = syntheticsMonitorWaitToInt64(api.Wait)
		if err != nil {
			return nil, err
		}
	}
	return &tfICMPMonitorFieldsV0{
		Host: types.StringPointerValue(api.Host),
		Wait: types.Int64Value(wait),
	}, nil
}

func (v *tfBrowserMonitorFieldsV0) toTfBrowserMonitorFieldsV0(api *kbapi.SyntheticsMonitor) (*tfBrowserMonitorFieldsV0, error) {
	var err error
	playwrightOptions := v.PlaywrightOptions
	if api.PlaywrightOptions != nil {
		playwrightOptions, err = toNormalizedValue(*api.PlaywrightOptions)
		if err != nil {
			return nil, err
		}
	}

	syntheticsArgs := v.SyntheticsArgs
	if api.SyntheticsArgs != nil {
		syntheticsArgs = synthetics.StringSliceValue(*api.SyntheticsArgs)
	}

	inlineScript := v.InlineScript
	if api.InlineScript != nil {
		inlineScript = types.StringPointerValue(api.InlineScript)
	}

	return &tfBrowserMonitorFieldsV0{
		InlineScript:      inlineScript,
		Screenshots:       types.StringPointerValue(api.Screenshots),
		SyntheticsArgs:    syntheticsArgs,
		IgnoreHTTPSErrors: types.BoolPointerValue(api.IgnoreHttpsErrors),
		PlaywrightOptions: playwrightOptions,
	}, nil
}

func (v *tfHTTPMonitorFieldsV0) toTfHTTPMonitorFieldsV0(ctx context.Context, dg diag.Diagnostics, api *kbapi.SyntheticsMonitor) *tfHTTPMonitorFieldsV0 {
	var err error
	proxyHeaders := v.ProxyHeader
	if api.ProxyHeaders != nil {
		proxyHeaders, err = toNormalizedValue(*api.ProxyHeaders)
		if err != nil {
			dg.AddError("Failed to parse proxy_headers", err.Error())
			return nil
		}
	}

	username := v.Username
	if api.Username != nil {
		username = types.StringPointerValue(api.Username)
	}
	password := v.Password
	if api.Password != nil {
		password = types.StringPointerValue(api.Password)
	}
	proxyURL := types.StringValue("")
	if !v.ProxyURL.IsNull() && !v.ProxyURL.IsUnknown() {
		proxyURL = v.ProxyURL
	}
	if api.ProxyUrl != nil {
		proxyURL = types.StringPointerValue(api.ProxyUrl)
	}

	maxRedirects := int64(0)
	if api.MaxRedirects != nil {
		maxRedirects, err = syntheticsMonitorMaxRedirectsToInt64(api.MaxRedirects)
		if err != nil {
			dg.AddError("Failed to parse max_redirects", err.Error())
			return nil
		}
	}

	sslCfg, dg := toTFSSLConfig(ctx, dg, api, "http")
	if dg.HasError() {
		return nil
	}
	return &tfHTTPMonitorFieldsV0{
		URL:          types.StringPointerValue(api.Url),
		MaxRedirects: types.Int64Value(maxRedirects),
		Mode:         types.StringPointerValue(api.Mode),
		IPv4:         types.BoolPointerValue(api.Ipv4),
		IPv6:         types.BoolPointerValue(api.Ipv6),
		Username:     username,
		Password:     password,
		ProxyHeader:  proxyHeaders,
		ProxyURL:     proxyURL,
		Check:        v.Check,
		Response:     v.Response,
		tfSSLConfig:  sslCfg,
	}
}

func toTFSSLConfig(ctx context.Context, dg diag.Diagnostics, api *kbapi.SyntheticsMonitor, p string) (tfSSLConfig, diag.Diagnostics) {
	sslSupportedProtocols := typeutils.SliceToListTypeString(ctx, typeutils.Deref(api.SslSupportedProtocols), path.Root(p).AtName("ssl_supported_protocols"), &dg)
	return tfSSLConfig{
		SslVerificationMode:       types.StringPointerValue(api.SslVerificationMode),
		SslSupportedProtocols:     sslSupportedProtocols,
		SslCertificateAuthorities: synthetics.StringSliceValue(typeutils.Deref(api.SslCertificateAuthorities)),
		SslCertificate:            types.StringPointerValue(api.SslCertificate),
		SslKey:                    types.StringPointerValue(api.SslKey),
		SslKeyPassphrase:          types.StringPointerValue(api.SslKeyPassphrase),
	}, dg
}

func toTfAlertConfigV0(ctx context.Context, alert *kbapi.SyntheticsMonitorAlert) (basetypes.ObjectValue, diag.Diagnostics) {
	dg := diag.Diagnostics{}

	alertAttributes := monitorAlertConfigSchema().GetType().(attr.TypeWithAttributeTypes).AttributeTypes()

	var emptyAttr = map[string]attr.Type(nil)

	if alert == nil {
		return basetypes.NewObjectNull(emptyAttr), dg
	}

	tfAlertConfig := tfAlertConfigV0{
		Status: toTfStatusConfigV0(alert.Status),
		TLS:    toTfStatusConfigV0(alert.Tls),
	}

	return types.ObjectValueFrom(ctx, alertAttributes, &tfAlertConfig)
}

func toTfStatusConfigV0(status *kbapi.SyntheticsMonitorAlertStatus) *tfStatusConfigV0 {
	if status == nil {
		return nil
	}
	return &tfStatusConfigV0{
		Enabled: types.BoolPointerValue(status.Enabled),
	}
}

func toAPIAlertConfig(ctx context.Context, v basetypes.ObjectValue) (*kbapi.SyntheticsMonitorAlert, diag.Diagnostics) {
	if v.IsNull() || v.IsUnknown() {
		return nil, nil
	}
	tfAlert := tfAlertConfigV0{}
	dg := tfsdk.ValueAs(ctx, v, &tfAlert)
	if dg.HasError() {
		return nil, dg
	}
	return tfAlert.toAPIAlertConfig(), dg
}

func toSSLConfig(ctx context.Context, dg diag.Diagnostics, v tfSSLConfig, p string) (*kbapi.SyntheticsSslConfig, diag.Diagnostics) {
	var ssl *kbapi.SyntheticsSslConfig
	if !v.SslSupportedProtocols.IsNull() && !v.SslSupportedProtocols.IsUnknown() {
		sslSupportedProtocols := typeutils.ListTypeToSliceString(ctx, v.SslSupportedProtocols, path.Root(p).AtName("ssl_supported_protocols"), &dg)
		if dg.HasError() {
			return nil, dg
		}
		ssl = &kbapi.SyntheticsSslConfig{
			SupportedProtocols: typeutils.SliceNilIfEmpty(sslSupportedProtocols),
		}
	}

	if !v.SslVerificationMode.IsNull() && !v.SslVerificationMode.IsUnknown() {
		if ssl == nil {
			ssl = &kbapi.SyntheticsSslConfig{}
		}
		value := v.SslVerificationMode.ValueString()
		ssl.VerificationMode = &value
	}

	certAuths := synthetics.ValueStringSlice(v.SslCertificateAuthorities)
	if len(certAuths) > 0 {
		if ssl == nil {
			ssl = &kbapi.SyntheticsSslConfig{}
		}
		ssl.CertificateAuthorities = typeutils.SliceNilIfEmpty(certAuths)
	}

	if !v.SslCertificate.IsUnknown() && !v.SslCertificate.IsNull() {
		if ssl == nil {
			ssl = &kbapi.SyntheticsSslConfig{}
		}
		value := v.SslCertificate.ValueString()
		ssl.Certificate = &value
	}

	if !v.SslKey.IsUnknown() && !v.SslKey.IsNull() {
		if ssl == nil {
			ssl = &kbapi.SyntheticsSslConfig{}
		}
		value := v.SslKey.ValueString()
		ssl.Key = &value
	}

	if !v.SslKeyPassphrase.IsUnknown() && !v.SslKeyPassphrase.IsNull() {
		if ssl == nil {
			ssl = &kbapi.SyntheticsSslConfig{}
		}
		value := v.SslKeyPassphrase.ValueString()
		ssl.KeyPassphrase = &value
	}
	return ssl, dg
}

func (v *tfModelV0) toKibanaAPIRequest(ctx context.Context) (*kbapi.SyntheticsMonitorRequest, diag.Diagnostics) {
	params, dg := toJSONObject(v.Params)
	if dg.HasError() {
		return nil, dg
	}

	labels, locations, alert, dg := v.monitorRequestCommon(ctx, dg)
	if dg.HasError() {
		return nil, dg
	}

	req := &kbapi.SyntheticsMonitorRequest{}
	switch {
	case v.HTTP != nil:
		httpReq, httpDg := v.newHTTPMonitorRequest(ctx, labels, locations, params, alert)
		dg.Append(httpDg...)
		if dg.HasError() {
			return nil, dg
		}
		if err := req.FromSyntheticsHttpMonitorFields(*httpReq); err != nil {
			return nil, diagutil.FrameworkDiagFromError(err)
		}
	case v.TCP != nil:
		tcpReq, tcpDg := v.newTCPMonitorRequest(ctx, labels, locations, params, alert)
		dg.Append(tcpDg...)
		if dg.HasError() {
			return nil, dg
		}
		if err := req.FromSyntheticsTcpMonitorFields(*tcpReq); err != nil {
			return nil, diagutil.FrameworkDiagFromError(err)
		}
	case v.ICMP != nil:
		icmpReq := v.newICMPMonitorRequest(labels, locations, params, alert)
		if err := req.FromSyntheticsIcmpMonitorFields(*icmpReq); err != nil {
			return nil, diagutil.FrameworkDiagFromError(err)
		}
	case v.Browser != nil:
		browserReq, browserDg := v.newBrowserMonitorRequest(labels, locations, params, alert)
		dg.Append(browserDg...)
		if dg.HasError() {
			return nil, dg
		}
		if err := req.FromSyntheticsBrowserMonitorFields(*browserReq); err != nil {
			return nil, diagutil.FrameworkDiagFromError(err)
		}
	default:
		dg.AddError("Unsupported monitor type config", "one of http,tcp,icmp,browser monitor fields is required")
	}

	if dg.HasError() {
		return nil, dg
	}

	return req, dg
}

func (v *tfModelV0) monitorRequestCommon(ctx context.Context, dg diag.Diagnostics) (map[string]string, []string, *kbapi.SyntheticsMonitorAlert, diag.Diagnostics) {
	labels := typeutils.MapTypeAs[string](ctx, v.Labels, path.Root("labels"), &dg)
	if dg.HasError() {
		return nil, nil, nil, dg
	}
	if labels == nil {
		labels = map[string]string{}
	}

	locations := make([]string, 0, len(v.Locations))
	for _, location := range v.Locations {
		locations = append(locations, location.ValueString())
	}

	alert, alertDg := toAPIAlertConfig(ctx, v.Alert)
	dg.Append(alertDg...)
	if dg.HasError() {
		return nil, nil, nil, dg
	}

	return labels, locations, alert, dg
}

func (v *tfModelV0) newHTTPMonitorRequest(
	ctx context.Context,
	labels map[string]string,
	locations []string,
	params map[string]any,
	alert *kbapi.SyntheticsMonitorAlert,
) (*kbapi.SyntheticsHttpMonitorFields, diag.Diagnostics) {
	h := v.HTTP

	proxyHeaders, dg := toJSONObject(h.ProxyHeader)
	if dg.HasError() {
		return nil, dg
	}
	response, d := toJSONObject(h.Response)
	dg.Append(d...)
	if dg.HasError() {
		return nil, dg
	}
	check, d := toJSONObject(h.Check)
	dg.Append(d...)
	if dg.HasError() {
		return nil, dg
	}

	ssl, d := toSSLConfig(ctx, dg, h.tfSSLConfig, "http")
	dg.Append(d...)
	if dg.HasError() {
		return nil, dg
	}

	req := &kbapi.SyntheticsHttpMonitorFields{
		AdditionalProperties: map[string]any{},
		Alert:                alert,
		Enabled:              v.Enabled.ValueBoolPointer(),
		Ipv4:                 h.IPv4.ValueBoolPointer(),
		Ipv6:                 h.IPv6.ValueBoolPointer(),
		Labels:               typeutils.MapRef(labels),
		Locations:            typeutils.SliceNilIfEmpty(locations),
		MaxRedirects:         int64ToSyntheticsHTTPMonitorFieldsMaxRedirects(h.MaxRedirects),
		Mode:                 stringEnumPtr[kbapi.SyntheticsHttpMonitorFieldsMode](h.Mode),
		Name:                 v.Name.ValueString(),
		Namespace:            typeutils.NonEmptyStringPointerValue(v.Namespace),
		Params:               typeutils.MapRef(params),
		Password:             typeutils.NonEmptyStringPointerValue(h.Password),
		PrivateLocations:     typeutils.SliceNilIfEmpty(synthetics.ValueStringSlice(v.PrivateLocations)),
		ProxyHeaders:         typeutils.MapRef(proxyHeaders),
		ProxyUrl:             typeutils.NonEmptyStringPointerValue(h.ProxyURL),
		Response:             typeutils.MapRef(response),
		RetestOnFailure:      v.RetestOnFailure.ValueBoolPointer(),
		Schedule:             int64ToFloat32Ptr(v.Schedule),
		ServiceName:          typeutils.NonEmptyStringPointerValue(v.APMServiceName),
		Ssl:                  ssl,
		Tags:                 typeutils.SliceNilIfEmpty(synthetics.ValueStringSlice(v.Tags)),
		Timeout:              int64ToFloat32Ptr(v.TimeoutSeconds),
		Type:                 kbapi.SyntheticsHttpMonitorFieldsType(kbapi.SyntheticsMonitorTypeHttp),
		Url:                  h.URL.ValueString(),
		Username:             typeutils.NonEmptyStringPointerValue(h.Username),
	}

	if check != nil {
		req.AdditionalProperties["check"] = check
	}

	return req, dg
}

func (v *tfModelV0) newTCPMonitorRequest(
	ctx context.Context,
	labels map[string]string,
	locations []string,
	params map[string]any,
	alert *kbapi.SyntheticsMonitorAlert,
) (*kbapi.SyntheticsTcpMonitorFields, diag.Diagnostics) {
	ssl, dg := toSSLConfig(ctx, nil, v.TCP.tfSSLConfig, "tcp")
	if dg.HasError() {
		return nil, dg
	}

	additionalProperties := map[string]any{}
	if !v.TCP.CheckSend.IsNull() && !v.TCP.CheckSend.IsUnknown() {
		additionalProperties["check.send"] = v.TCP.CheckSend.ValueString()
	}
	if !v.TCP.CheckReceive.IsNull() && !v.TCP.CheckReceive.IsUnknown() {
		additionalProperties["check.receive"] = v.TCP.CheckReceive.ValueString()
	}

	return &kbapi.SyntheticsTcpMonitorFields{
		Alert:                 alert,
		Enabled:               v.Enabled.ValueBoolPointer(),
		Host:                  v.TCP.Host.ValueString(),
		Labels:                typeutils.MapRef(labels),
		Locations:             typeutils.SliceNilIfEmpty(locations),
		Name:                  v.Name.ValueString(),
		Namespace:             typeutils.NonEmptyStringPointerValue(v.Namespace),
		Params:                typeutils.MapRef(params),
		PrivateLocations:      typeutils.SliceNilIfEmpty(synthetics.ValueStringSlice(v.PrivateLocations)),
		ProxyUrl:              typeutils.NonEmptyStringPointerValue(v.TCP.ProxyURL),
		ProxyUseLocalResolver: v.TCP.ProxyUseLocalResolver.ValueBoolPointer(),
		RetestOnFailure:       v.RetestOnFailure.ValueBoolPointer(),
		Schedule:              int64ToFloat32Ptr(v.Schedule),
		ServiceName:           typeutils.NonEmptyStringPointerValue(v.APMServiceName),
		Ssl:                   ssl,
		Tags:                  typeutils.SliceNilIfEmpty(synthetics.ValueStringSlice(v.Tags)),
		Timeout:               int64ToFloat32Ptr(v.TimeoutSeconds),
		Type:                  kbapi.SyntheticsTcpMonitorFieldsType(kbapi.SyntheticsMonitorTypeTcp),
		AdditionalProperties:  additionalProperties,
	}, dg
}

func (v *tfModelV0) newICMPMonitorRequest(labels map[string]string, locations []string, params map[string]any, alert *kbapi.SyntheticsMonitorAlert) *kbapi.SyntheticsIcmpMonitorFields {
	return &kbapi.SyntheticsIcmpMonitorFields{
		Alert:            alert,
		Enabled:          v.Enabled.ValueBoolPointer(),
		Host:             v.ICMP.Host.ValueString(),
		Labels:           typeutils.MapRef(labels),
		Locations:        typeutils.SliceNilIfEmpty(locations),
		Name:             v.Name.ValueString(),
		Namespace:        typeutils.NonEmptyStringPointerValue(v.Namespace),
		Params:           typeutils.MapRef(params),
		PrivateLocations: typeutils.SliceNilIfEmpty(synthetics.ValueStringSlice(v.PrivateLocations)),
		RetestOnFailure:  v.RetestOnFailure.ValueBoolPointer(),
		Schedule:         int64ToFloat32Ptr(v.Schedule),
		ServiceName:      typeutils.NonEmptyStringPointerValue(v.APMServiceName),
		Tags:             typeutils.SliceNilIfEmpty(synthetics.ValueStringSlice(v.Tags)),
		Timeout:          int64ToFloat32Ptr(v.TimeoutSeconds),
		Type:             kbapi.SyntheticsIcmpMonitorFieldsType(kbapi.SyntheticsMonitorTypeIcmp),
		Wait:             int64ToSyntheticsIcmpMonitorFieldsWait(v.ICMP.Wait),
	}
}

func (v *tfModelV0) newBrowserMonitorRequest(
	labels map[string]string,
	locations []string,
	params map[string]any,
	alert *kbapi.SyntheticsMonitorAlert,
) (*kbapi.SyntheticsBrowserMonitorFields, diag.Diagnostics) {
	playwrightOptions, dg := toJSONObject(v.Browser.PlaywrightOptions)
	if dg.HasError() {
		return nil, dg
	}

	return &kbapi.SyntheticsBrowserMonitorFields{
		Alert:             alert,
		Enabled:           v.Enabled.ValueBoolPointer(),
		IgnoreHttpsErrors: v.Browser.IgnoreHTTPSErrors.ValueBoolPointer(),
		InlineScript:      v.Browser.InlineScript.ValueString(),
		Labels:            typeutils.MapRef(labels),
		Locations:         typeutils.SliceNilIfEmpty(locations),
		Name:              v.Name.ValueString(),
		Namespace:         typeutils.NonEmptyStringPointerValue(v.Namespace),
		Params:            typeutils.MapRef(params),
		PlaywrightOptions: typeutils.MapRef(playwrightOptions),
		PrivateLocations:  typeutils.SliceNilIfEmpty(synthetics.ValueStringSlice(v.PrivateLocations)),
		RetestOnFailure:   v.RetestOnFailure.ValueBoolPointer(),
		Schedule:          int64ToFloat32Ptr(v.Schedule),
		Screenshots:       stringEnumPtr[kbapi.SyntheticsBrowserMonitorFieldsScreenshots](v.Browser.Screenshots),
		ServiceName:       typeutils.NonEmptyStringPointerValue(v.APMServiceName),
		SyntheticsArgs:    typeutils.SliceNilIfEmpty(synthetics.ValueStringSlice(v.Browser.SyntheticsArgs)),
		Tags:              typeutils.SliceNilIfEmpty(synthetics.ValueStringSlice(v.Tags)),
		Timeout:           int64ToFloat32Ptr(v.TimeoutSeconds),
		Type:              kbapi.SyntheticsBrowserMonitorFieldsType(kbapi.SyntheticsMonitorTypeBrowser),
	}, dg
}

func (v tfAlertConfigV0) toAPIAlertConfig() *kbapi.SyntheticsMonitorAlert {
	var status *kbapi.SyntheticsMonitorAlertStatus
	if v.Status != nil {
		status = v.Status.toAPIAlertStatus()
	}
	var tls *kbapi.SyntheticsMonitorAlertStatus
	if v.TLS != nil {
		tls = v.TLS.toAPIAlertStatus()
	}
	return &kbapi.SyntheticsMonitorAlert{
		Status: status,
		Tls:    tls,
	}
}

func (v tfStatusConfigV0) toAPIAlertStatus() *kbapi.SyntheticsMonitorAlertStatus {
	return &kbapi.SyntheticsMonitorAlertStatus{
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
