package synthetics

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/disaster37/go-kibana-rest/v8/kbapi"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

var (
	fBool = boolPointer(false)
	tBool = boolPointer(true)
)

func boolPointer(v bool) *bool {
	var res = new(bool)
	*res = v
	return res
}

func TestToModelV0(t *testing.T) {
	testcases := []struct {
		name     string
		input    kbapi.SyntheticsMonitor
		expected tfModelV0
	}{
		{
			name: "HTTP monitor empty data",
			input: kbapi.SyntheticsMonitor{
				Type: kbapi.Http,
			},
			expected: tfModelV0{
				ID:             types.StringValue("/"),
				Name:           types.StringValue(""),
				SpaceID:        types.StringValue(""),
				Schedule:       types.Int64Value(0),
				APMServiceName: types.StringValue(""),
				TimeoutSeconds: types.Int64Value(0),
				Params:         jsontypes.NewNormalizedValue("null"),
				HTTP: &tfHTTPMonitorFieldsV0{
					URL:                   types.StringValue(""),
					SslVerificationMode:   types.StringValue(""),
					MaxRedirects:          types.Int64Value(0),
					Mode:                  types.StringValue(""),
					Username:              types.StringValue(""),
					Password:              types.StringValue(""),
					ProxyHeader:           jsontypes.NewNormalizedValue("null"),
					ProxyURL:              types.StringValue(""),
					Response:              jsontypes.NewNormalizedValue("null"),
					Check:                 jsontypes.NewNormalizedValue("null"),
					SslSupportedProtocols: types.ListNull(types.StringType),
				},
			},
		},
		{
			name: "TCP monitor empty data",
			input: kbapi.SyntheticsMonitor{
				Type: kbapi.Tcp,
			},
			expected: tfModelV0{
				ID:             types.StringValue("/"),
				Name:           types.StringValue(""),
				SpaceID:        types.StringValue(""),
				Schedule:       types.Int64Value(0),
				APMServiceName: types.StringValue(""),
				TimeoutSeconds: types.Int64Value(0),
				Params:         jsontypes.NewNormalizedValue("null"),
				TCP: &tfTCPMonitorFieldsV0{
					Host:                  types.StringValue(""),
					SslVerificationMode:   types.StringValue(""),
					CheckSend:             types.StringValue(""),
					CheckReceive:          types.StringValue(""),
					ProxyURL:              types.StringValue(""),
					SslSupportedProtocols: types.ListNull(types.StringType),
				},
			},
		},
		{
			name: "ICMP monitor empty data",
			input: kbapi.SyntheticsMonitor{
				Type: kbapi.Icmp,
			},
			expected: tfModelV0{
				ID:             types.StringValue("/"),
				Name:           types.StringValue(""),
				SpaceID:        types.StringValue(""),
				Schedule:       types.Int64Value(0),
				APMServiceName: types.StringValue(""),
				TimeoutSeconds: types.Int64Value(0),
				Params:         jsontypes.NewNormalizedValue("null"),
				ICMP: &tfICMPMonitorFieldsV0{
					Host: types.StringValue(""),
					Wait: types.Int64Value(0),
				},
			},
		},
		{
			name: "Browser monitor empty data",
			input: kbapi.SyntheticsMonitor{
				Type: kbapi.Browser,
			},
			expected: tfModelV0{
				ID:             types.StringValue("/"),
				Name:           types.StringValue(""),
				SpaceID:        types.StringValue(""),
				Schedule:       types.Int64Value(0),
				APMServiceName: types.StringValue(""),
				TimeoutSeconds: types.Int64Value(0),
				Params:         jsontypes.NewNormalizedValue("null"),
				Browser: &tfBrowserMonitorFieldsV0{
					InlineScript:      types.StringValue(""),
					Screenshots:       types.StringValue(""),
					PlaywrightOptions: jsontypes.NewNormalizedValue("null"),
				},
			},
		},
		{
			name: "HTTP monitor",
			input: kbapi.SyntheticsMonitor{
				Id:             "test-id-http",
				Name:           "test-name-http",
				Namespace:      "default",
				Enabled:        tBool,
				Alert:          &kbapi.MonitorAlertConfig{Status: &kbapi.SyntheticsStatusConfig{Enabled: tBool}, Tls: &kbapi.SyntheticsStatusConfig{Enabled: fBool}},
				Schedule:       &kbapi.MonitorScheduleConfig{Number: "5", Unit: "m"},
				Tags:           []string{"tag1", "tag2"},
				APMServiceName: "test-service-http",
				Timeout:        json.Number("30"),
				Locations: []kbapi.MonitorLocationConfig{
					{Label: "North America - US East", Id: "us_east", IsServiceManaged: true},
					{Label: "test private location", IsServiceManaged: false},
				},
				Origin:                      "origin",
				Params:                      kbapi.JsonObject{"param1": "value1"},
				MaxAttempts:                 3,
				Revision:                    1,
				Ui:                          kbapi.JsonObject{"is_tls_enabled": false},
				Type:                        kbapi.Http,
				Url:                         "https://example.com",
				Mode:                        kbapi.HttpMonitorMode("all"),
				MaxRedirects:                "5",
				Ipv4:                        tBool,
				Ipv6:                        fBool,
				Username:                    "user",
				Password:                    "pass",
				ProxyHeaders:                kbapi.JsonObject{"header1": "value1"},
				ProxyUrl:                    "https://proxy.com",
				CheckResponseBodyPositive:   []string{"foo", "bar"},
				CheckResponseStatus:         []string{"200", "201"},
				ResponseIncludeBody:         "always",
				ResponseIncludeHeaders:      true,
				ResponseIncludeBodyMaxBytes: "1024",
				CheckRequestBody:            kbapi.JsonObject{"type": "text", "value": "name=first&email=someemail%40someemailprovider.com"},
				CheckRequestHeaders:         kbapi.JsonObject{"Content-Type": "application/x-www-form-urlencoded"},
				CheckRequestMethod:          "POST",
				SslVerificationMode:         "full",
				SslSupportedProtocols:       []string{"TLSv1.2", "TLSv1.3"},
			},
			expected: tfModelV0{
				ID:               types.StringValue("default/test-id-http"),
				Name:             types.StringValue("test-name-http"),
				SpaceID:          types.StringValue("default"),
				Schedule:         types.Int64Value(5),
				Locations:        []types.String{types.StringValue("us_east")},
				PrivateLocations: []types.String{types.StringValue("test private location")},
				Enabled:          types.BoolPointerValue(tBool),
				Tags:             []types.String{types.StringValue("tag1"), types.StringValue("tag2")},
				Alert:            &tfAlertConfigV0{Status: &tfStatusConfigV0{Enabled: types.BoolPointerValue(tBool)}, TLS: &tfStatusConfigV0{Enabled: types.BoolPointerValue(fBool)}},
				APMServiceName:   types.StringValue("test-service-http"),
				TimeoutSeconds:   types.Int64Value(30),
				Params:           jsontypes.NewNormalizedValue(`{"param1":"value1"}`),
				HTTP: &tfHTTPMonitorFieldsV0{
					URL:                 types.StringValue("https://example.com"),
					SslVerificationMode: types.StringValue("full"),
					SslSupportedProtocols: types.ListValueMust(types.StringType, []attr.Value{
						types.StringValue("TLSv1.2"), types.StringValue("TLSv1.3"),
					}),
					MaxRedirects: types.Int64Value(5),
					Mode:         types.StringValue("all"),
					IPv4:         types.BoolPointerValue(tBool),
					IPv6:         types.BoolPointerValue(fBool),
					Username:     types.StringValue("user"),
					Password:     types.StringValue("pass"),
					ProxyHeader:  jsontypes.NewNormalizedValue(`{"header1":"value1"}`),
					ProxyURL:     types.StringValue("https://proxy.com"),
				},
			},
		},
		{
			name: "TCP monitor",
			input: kbapi.SyntheticsMonitor{
				Id:             "test-id-tcp",
				Name:           "test-name-tcp",
				Namespace:      "default",
				Enabled:        tBool,
				Alert:          &kbapi.MonitorAlertConfig{Status: &kbapi.SyntheticsStatusConfig{Enabled: tBool}},
				Schedule:       &kbapi.MonitorScheduleConfig{Number: "5", Unit: "m"},
				Tags:           nil,
				APMServiceName: "test-service-tcp",
				Timeout:        json.Number("30"),
				Locations: []kbapi.MonitorLocationConfig{
					{Label: "test private location", IsServiceManaged: false},
				},
				Origin:                "origin",
				Params:                kbapi.JsonObject{"param1": "value1"},
				MaxAttempts:           3,
				Revision:              1,
				Ui:                    kbapi.JsonObject{"is_tls_enabled": false},
				Type:                  kbapi.Tcp,
				SslVerificationMode:   "full",
				SslSupportedProtocols: []string{"TLSv1.2", "TLSv1.3"},
				ProxyUrl:              "http://proxy.com",
				Host:                  "example.com:9200",
				CheckSend:             "hello",
				CheckReceive:          "world",
				ProxyUseLocalResolver: tBool,
			},
			expected: tfModelV0{
				ID:               types.StringValue("default/test-id-tcp"),
				Name:             types.StringValue("test-name-tcp"),
				SpaceID:          types.StringValue("default"),
				Schedule:         types.Int64Value(5),
				Locations:        nil,
				PrivateLocations: []types.String{types.StringValue("test private location")},
				Enabled:          types.BoolPointerValue(tBool),
				Tags:             nil,
				Alert:            &tfAlertConfigV0{Status: &tfStatusConfigV0{Enabled: types.BoolPointerValue(tBool)}},
				APMServiceName:   types.StringValue("test-service-tcp"),
				TimeoutSeconds:   types.Int64Value(30),
				Params:           jsontypes.NewNormalizedValue(`{"param1":"value1"}`),
				TCP: &tfTCPMonitorFieldsV0{
					Host:                types.StringValue("example.com:9200"),
					SslVerificationMode: types.StringValue("full"),
					SslSupportedProtocols: types.ListValueMust(types.StringType, []attr.Value{
						types.StringValue("TLSv1.2"), types.StringValue("TLSv1.3"),
					}),
					CheckSend:             types.StringValue("hello"),
					CheckReceive:          types.StringValue("world"),
					ProxyURL:              types.StringValue("http://proxy.com"),
					ProxyUseLocalResolver: types.BoolPointerValue(tBool),
				},
			},
		},
		{
			name: "ICMP monitor",
			input: kbapi.SyntheticsMonitor{
				Id:             "test-id-icmp",
				Name:           "test-name-icmp",
				Namespace:      "default",
				Enabled:        tBool,
				Alert:          &kbapi.MonitorAlertConfig{Status: &kbapi.SyntheticsStatusConfig{Enabled: tBool}},
				Schedule:       &kbapi.MonitorScheduleConfig{Number: "5", Unit: "m"},
				Tags:           nil,
				APMServiceName: "test-service-tcp",
				Timeout:        json.Number("30"),
				Locations: []kbapi.MonitorLocationConfig{
					{Label: "test private location", IsServiceManaged: false},
				},
				Origin:                "origin",
				Params:                kbapi.JsonObject{"param1": "value1"},
				MaxAttempts:           3,
				Revision:              1,
				Ui:                    kbapi.JsonObject{"is_tls_enabled": false},
				Type:                  kbapi.Icmp,
				SslVerificationMode:   "full",
				SslSupportedProtocols: []string{"TLSv1.2", "TLSv1.3"},
				ProxyUrl:              "http://proxy.com",
				Host:                  "example.com:9200",
				Wait:                  "30",
			},
			expected: tfModelV0{
				ID:               types.StringValue("default/test-id-icmp"),
				Name:             types.StringValue("test-name-icmp"),
				SpaceID:          types.StringValue("default"),
				Schedule:         types.Int64Value(5),
				Locations:        nil,
				PrivateLocations: []types.String{types.StringValue("test private location")},
				Enabled:          types.BoolPointerValue(tBool),
				Tags:             nil,
				Alert:            &tfAlertConfigV0{Status: &tfStatusConfigV0{Enabled: types.BoolPointerValue(tBool)}},
				APMServiceName:   types.StringValue("test-service-tcp"),
				TimeoutSeconds:   types.Int64Value(30),
				Params:           jsontypes.NewNormalizedValue(`{"param1":"value1"}`),
				ICMP: &tfICMPMonitorFieldsV0{
					Host: types.StringValue("example.com:9200"),
					Wait: types.Int64Value(30),
				},
			},
		},
		{
			name: "Browser monitor",
			input: kbapi.SyntheticsMonitor{
				Id:             "test-id-browser",
				Name:           "test-name-browser",
				Namespace:      "default",
				Enabled:        tBool,
				Alert:          &kbapi.MonitorAlertConfig{Status: &kbapi.SyntheticsStatusConfig{Enabled: tBool}},
				Schedule:       &kbapi.MonitorScheduleConfig{Number: "5", Unit: "m"},
				Tags:           nil,
				APMServiceName: "test-service-tcp",
				Timeout:        json.Number("30"),
				Locations: []kbapi.MonitorLocationConfig{
					{Label: "test private location", IsServiceManaged: false},
				},
				Origin:                "origin",
				Params:                kbapi.JsonObject{"param1": "value1"},
				MaxAttempts:           3,
				Revision:              1,
				Ui:                    kbapi.JsonObject{"is_tls_enabled": false},
				Type:                  kbapi.Browser,
				SslVerificationMode:   "full",
				SslSupportedProtocols: []string{"TLSv1.2", "TLSv1.3"},
				ProxyUrl:              "http://proxy.com",
				Screenshots:           "off",
				IgnoreHttpsErrors:     tBool,
				InlineScript:          `step('Go to https://google.com.co', () => page.goto('https://www.google.com'))`,
				SyntheticsArgs:        []string{"--no-sandbox", "--disable-setuid-sandbox"},
				PlaywrightOptions: map[string]interface{}{
					"ignoreHTTPSErrors": false,
					"httpCredentials": map[string]interface{}{
						"username": "test",
						"password": "test",
					},
				},
			},
			expected: tfModelV0{
				ID:               types.StringValue("default/test-id-browser"),
				Name:             types.StringValue("test-name-browser"),
				SpaceID:          types.StringValue("default"),
				Schedule:         types.Int64Value(5),
				Locations:        nil,
				PrivateLocations: []types.String{types.StringValue("test private location")},
				Enabled:          types.BoolPointerValue(tBool),
				Tags:             nil,
				Alert:            &tfAlertConfigV0{Status: &tfStatusConfigV0{Enabled: types.BoolPointerValue(tBool)}},
				APMServiceName:   types.StringValue("test-service-tcp"),
				TimeoutSeconds:   types.Int64Value(30),
				Params:           jsontypes.NewNormalizedValue(`{"param1":"value1"}`),
				Browser: &tfBrowserMonitorFieldsV0{
					Screenshots:       types.StringValue("off"),
					IgnoreHttpsErrors: types.BoolPointerValue(tBool),
					InlineScript:      types.StringValue(`step('Go to https://google.com.co', () => page.goto('https://www.google.com'))`),
					SyntheticsArgs:    []types.String{types.StringValue("--no-sandbox"), types.StringValue("--disable-setuid-sandbox")},
					PlaywrightOptions: jsontypes.NewNormalizedValue(`{"httpCredentials":{"password":"test","username":"test"},"ignoreHTTPSErrors":false}`),
				},
			},
		},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			model, diag := tt.expected.toModelV0(ctx, &tt.input)
			assert.False(t, diag.HasError())
			assert.Equal(t, &tt.expected, model)
		})
	}
}

func TestToKibanaAPIRequest(t *testing.T) {
	testcases := []struct {
		name     string
		input    tfModelV0
		expected kibanaAPIRequest
	}{
		{
			name: "Empty HTTP monitor",
			input: tfModelV0{
				HTTP: &tfHTTPMonitorFieldsV0{},
			},
			expected: kibanaAPIRequest{
				fields: kbapi.HTTPMonitorFields{},
				config: kbapi.SyntheticsMonitorConfig{},
			},
		},
		{
			name: "Empty TCP monitor",
			input: tfModelV0{
				TCP: &tfTCPMonitorFieldsV0{},
			},
			expected: kibanaAPIRequest{
				fields: kbapi.TCPMonitorFields{},
				config: kbapi.SyntheticsMonitorConfig{},
			},
		},
		{
			name: "Empty ICMP monitor",
			input: tfModelV0{
				ICMP: &tfICMPMonitorFieldsV0{},
			},
			expected: kibanaAPIRequest{
				fields: kbapi.ICMPMonitorFields{},
				config: kbapi.SyntheticsMonitorConfig{},
			},
		},
		{
			name: "Empty Browser monitor",
			input: tfModelV0{
				Browser: &tfBrowserMonitorFieldsV0{},
			},
			expected: kibanaAPIRequest{
				fields: kbapi.BrowserMonitorFields{},
				config: kbapi.SyntheticsMonitorConfig{},
			},
		},
		{
			name: "HTTP monitor",
			input: tfModelV0{
				ID:               types.StringValue("test-id-http"),
				Name:             types.StringValue("test-name-http"),
				SpaceID:          types.StringValue("default"),
				Schedule:         types.Int64Value(5),
				Locations:        []types.String{types.StringValue("us_east")},
				PrivateLocations: []types.String{types.StringValue("test private location")},
				Enabled:          types.BoolPointerValue(tBool),
				Tags:             []types.String{types.StringValue("tag1"), types.StringValue("tag2")},
				Alert:            &tfAlertConfigV0{Status: &tfStatusConfigV0{Enabled: types.BoolPointerValue(tBool)}, TLS: &tfStatusConfigV0{Enabled: types.BoolPointerValue(fBool)}},
				APMServiceName:   types.StringValue("test-service-http"),
				TimeoutSeconds:   types.Int64Value(30),
				Params:           jsontypes.NewNormalizedValue(`{"param1":"value1"}`),
				HTTP: &tfHTTPMonitorFieldsV0{
					URL:                 types.StringValue("https://example.com"),
					SslVerificationMode: types.StringValue("full"),
					SslSupportedProtocols: types.ListValueMust(types.StringType, []attr.Value{
						types.StringValue("TLSv1.2"), types.StringValue("TLSv1.3"),
					}),
					MaxRedirects: types.Int64Value(5),
					Mode:         types.StringValue("all"),
					IPv4:         types.BoolPointerValue(tBool),
					IPv6:         types.BoolPointerValue(fBool),
					Username:     types.StringValue("user"),
					Password:     types.StringValue("pass"),
					ProxyHeader:  jsontypes.NewNormalizedValue(`{"header1":"value1"}`),
					ProxyURL:     types.StringValue("https://proxy.com"),
					Response:     jsontypes.NewNormalizedValue(`{"response1":"value1"}`),
					Check:        jsontypes.NewNormalizedValue(`{"check1":"value1"}`),
				},
			},
			expected: kibanaAPIRequest{
				config: kbapi.SyntheticsMonitorConfig{
					Name:             "test-name-http",
					Schedule:         kbapi.MonitorSchedule(5),
					Locations:        []kbapi.MonitorLocation{"us_east"},
					PrivateLocations: []string{"test private location"},
					Enabled:          tBool,
					Tags:             []string{"tag1", "tag2"},
					Alert:            &kbapi.MonitorAlertConfig{Status: &kbapi.SyntheticsStatusConfig{Enabled: tBool}, Tls: &kbapi.SyntheticsStatusConfig{Enabled: fBool}},
					APMServiceName:   "test-service-http",
					Namespace:        "default",
					TimeoutSeconds:   30,
					Params:           kbapi.JsonObject{"param1": "value1"},
				},
				fields: kbapi.HTTPMonitorFields{
					Url:                   "https://example.com",
					SslVerificationMode:   "full",
					SslSupportedProtocols: []string{"TLSv1.2", "TLSv1.3"},
					MaxRedirects:          "5",
					Mode:                  "all",
					Ipv4:                  tBool,
					Ipv6:                  fBool,
					Username:              "user",
					Password:              "pass",
					ProxyHeader:           kbapi.JsonObject{"header1": "value1"},
					ProxyUrl:              "https://proxy.com",
					Response:              kbapi.JsonObject{"response1": "value1"},
					Check:                 kbapi.JsonObject{"check1": "value1"},
				},
			},
		},
		{
			name: "TCP monitor",
			input: tfModelV0{
				ID:               types.StringValue("test-id-tcp"),
				Name:             types.StringValue("test-name-tcp"),
				SpaceID:          types.StringValue("default"),
				Schedule:         types.Int64Value(5),
				Locations:        []types.String{types.StringValue("us_east")},
				PrivateLocations: nil,
				Enabled:          types.BoolPointerValue(tBool),
				Tags:             []types.String{types.StringValue("tag1"), types.StringValue("tag2")},
				Alert:            &tfAlertConfigV0{Status: &tfStatusConfigV0{Enabled: types.BoolPointerValue(tBool)}},
				APMServiceName:   types.StringValue("test-service-tcp"),
				TimeoutSeconds:   types.Int64Value(30),
				Params:           jsontypes.NewNormalizedValue(`{"param1":"value1"}`),
				TCP: &tfTCPMonitorFieldsV0{
					Host:                types.StringValue("example.com:9200"),
					SslVerificationMode: types.StringValue("full"),
					SslSupportedProtocols: types.ListValueMust(types.StringType, []attr.Value{
						types.StringValue("TLSv1.2"), types.StringValue("TLSv1.3"),
					}),
					CheckSend:             types.StringValue("hello"),
					CheckReceive:          types.StringValue("world"),
					ProxyURL:              types.StringValue("http://proxy.com"),
					ProxyUseLocalResolver: types.BoolPointerValue(tBool),
				},
			},
			expected: kibanaAPIRequest{
				config: kbapi.SyntheticsMonitorConfig{
					Name:             "test-name-tcp",
					Schedule:         kbapi.MonitorSchedule(5),
					Locations:        []kbapi.MonitorLocation{"us_east"},
					PrivateLocations: nil,
					Enabled:          tBool,
					Tags:             []string{"tag1", "tag2"},
					Alert:            &kbapi.MonitorAlertConfig{Status: &kbapi.SyntheticsStatusConfig{Enabled: tBool}},
					APMServiceName:   "test-service-tcp",
					Namespace:        "default",
					TimeoutSeconds:   30,
					Params:           kbapi.JsonObject{"param1": "value1"},
				},
				fields: kbapi.TCPMonitorFields{
					Host:                  "example.com:9200",
					SslVerificationMode:   "full",
					SslSupportedProtocols: []string{"TLSv1.2", "TLSv1.3"},
					CheckSend:             "hello",
					CheckReceive:          "world",
					ProxyUrl:              "http://proxy.com",
					ProxyUseLocalResolver: tBool,
				},
			},
		},
		{
			name: "ICMP monitor",
			input: tfModelV0{
				ID:               types.StringValue("test-id-icmp"),
				Name:             types.StringValue("test-name-icmp"),
				SpaceID:          types.StringValue("default"),
				Schedule:         types.Int64Value(5),
				Locations:        []types.String{types.StringValue("us_east")},
				PrivateLocations: nil,
				Enabled:          types.BoolPointerValue(tBool),
				Tags:             []types.String{types.StringValue("tag1"), types.StringValue("tag2")},
				Alert:            &tfAlertConfigV0{Status: &tfStatusConfigV0{Enabled: types.BoolPointerValue(tBool)}},
				APMServiceName:   types.StringValue("test-service-tcp"),
				TimeoutSeconds:   types.Int64Value(30),
				Params:           jsontypes.NewNormalizedValue(`{"param1":"value1"}`),
				ICMP: &tfICMPMonitorFieldsV0{
					Host: types.StringValue("example.com:9200"),
					Wait: types.Int64Value(30),
				},
			},
			expected: kibanaAPIRequest{
				config: kbapi.SyntheticsMonitorConfig{
					Name:             "test-name-icmp",
					Schedule:         kbapi.MonitorSchedule(5),
					Locations:        []kbapi.MonitorLocation{"us_east"},
					PrivateLocations: nil,
					Enabled:          tBool,
					Tags:             []string{"tag1", "tag2"},
					Alert:            &kbapi.MonitorAlertConfig{Status: &kbapi.SyntheticsStatusConfig{Enabled: tBool}},
					APMServiceName:   "test-service-tcp",
					Namespace:        "default",
					TimeoutSeconds:   30,
					Params:           kbapi.JsonObject{"param1": "value1"},
				},
				fields: kbapi.ICMPMonitorFields{
					Host: "example.com:9200",
					Wait: "30",
				},
			},
		},
		{
			name: "Browser monitor",
			input: tfModelV0{
				ID:               types.StringValue("test-id-browser"),
				Name:             types.StringValue("test-name-browser"),
				SpaceID:          types.StringValue("default"),
				Schedule:         types.Int64Value(5),
				Locations:        []types.String{types.StringValue("us_east")},
				PrivateLocations: nil,
				Enabled:          types.BoolPointerValue(tBool),
				Tags:             []types.String{types.StringValue("tag1"), types.StringValue("tag2")},
				Alert:            &tfAlertConfigV0{Status: &tfStatusConfigV0{Enabled: types.BoolPointerValue(tBool)}},
				APMServiceName:   types.StringValue("test-service-tcp"),
				TimeoutSeconds:   types.Int64Value(30),
				Params:           jsontypes.NewNormalizedValue(`{"param1":"value1"}`),
				Browser: &tfBrowserMonitorFieldsV0{
					Screenshots:       types.StringValue("off"),
					IgnoreHttpsErrors: types.BoolPointerValue(tBool),
					InlineScript:      types.StringValue(`step('Go to https://google.com.co', () => page.goto('https://www.google.com'))`),
					SyntheticsArgs:    []types.String{types.StringValue("--no-sandbox"), types.StringValue("--disable-setuid-sandbox")},
					PlaywrightOptions: jsontypes.NewNormalizedValue(`{"httpCredentials":{"password":"test","username":"test"},"ignoreHTTPSErrors":false}`),
				},
			},
			expected: kibanaAPIRequest{
				config: kbapi.SyntheticsMonitorConfig{
					Name:             "test-name-browser",
					Schedule:         kbapi.MonitorSchedule(5),
					Locations:        []kbapi.MonitorLocation{"us_east"},
					PrivateLocations: nil,
					Enabled:          tBool,
					Tags:             []string{"tag1", "tag2"},
					Alert:            &kbapi.MonitorAlertConfig{Status: &kbapi.SyntheticsStatusConfig{Enabled: tBool}},
					APMServiceName:   "test-service-tcp",
					Namespace:        "default",
					TimeoutSeconds:   30,
					Params:           kbapi.JsonObject{"param1": "value1"},
				},
				fields: kbapi.BrowserMonitorFields{
					Screenshots:       "off",
					IgnoreHttpsErrors: tBool,
					InlineScript:      `step('Go to https://google.com.co', () => page.goto('https://www.google.com'))`,
					SyntheticsArgs:    []string{"--no-sandbox", "--disable-setuid-sandbox"},
					PlaywrightOptions: map[string]interface{}{
						"ignoreHTTPSErrors": false,
						"httpCredentials": map[string]interface{}{
							"username": "test",
							"password": "test",
						},
					},
				},
			},
		},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			apiRequest, dg := tt.input.toKibanaAPIRequest(context.Background())
			assert.False(t, dg.HasError(), dg.Errors())
			assert.Equal(t, &tt.expected, apiRequest)
		})
	}
}

func TestToModelV0MergeAttributes(t *testing.T) {

	testcases := []struct {
		name     string
		input    kbapi.SyntheticsMonitor
		state    tfModelV0
		expected tfModelV0
	}{
		{
			name: "HTTP monitor",
			state: tfModelV0{
				HTTP: &tfHTTPMonitorFieldsV0{
					ProxyHeader: jsontypes.NewNormalizedValue(`{"header1":"value1"}`),
					Username:    types.StringValue("test"),
					Password:    types.StringValue("password"),
					Check:       jsontypes.NewNormalizedValue(`{"check1":"value1"}`),
					Response:    jsontypes.NewNormalizedValue(`{"response1":"value1"}`),
				},
				Params:          jsontypes.NewNormalizedValue(`{"param1":"value1"}`),
				RetestOnFailure: types.BoolValue(true),
			},
			input: kbapi.SyntheticsMonitor{
				Type: kbapi.Http,
			},
			expected: tfModelV0{
				ID:              types.StringValue("/"),
				Name:            types.StringValue(""),
				SpaceID:         types.StringValue(""),
				Schedule:        types.Int64Value(0),
				APMServiceName:  types.StringValue(""),
				TimeoutSeconds:  types.Int64Value(0),
				Params:          jsontypes.NewNormalizedValue(`{"param1":"value1"}`),
				RetestOnFailure: types.BoolValue(true),
				HTTP: &tfHTTPMonitorFieldsV0{
					URL:                   types.StringValue(""),
					SslVerificationMode:   types.StringValue(""),
					SslSupportedProtocols: types.ListNull(types.StringType),
					MaxRedirects:          types.Int64Value(0),
					Mode:                  types.StringValue(""),
					ProxyURL:              types.StringValue(""),
					ProxyHeader:           jsontypes.NewNormalizedValue(`{"header1":"value1"}`),
					Username:              types.StringValue("test"),
					Password:              types.StringValue("password"),
					Check:                 jsontypes.NewNormalizedValue(`{"check1":"value1"}`),
					Response:              jsontypes.NewNormalizedValue(`{"response1":"value1"}`),
				},
			},
		},
		{
			name: "TCP monitor",
			state: tfModelV0{
				TCP: &tfTCPMonitorFieldsV0{
					CheckSend:    types.StringValue("hello"),
					CheckReceive: types.StringValue("world"),
				},
			},
			input: kbapi.SyntheticsMonitor{
				Type: kbapi.Tcp,
			},
			expected: tfModelV0{
				ID:             types.StringValue("/"),
				Name:           types.StringValue(""),
				SpaceID:        types.StringValue(""),
				Schedule:       types.Int64Value(0),
				APMServiceName: types.StringValue(""),
				TimeoutSeconds: types.Int64Value(0),
				TCP: &tfTCPMonitorFieldsV0{
					Host:                  types.StringValue(""),
					SslVerificationMode:   types.StringValue(""),
					CheckSend:             types.StringValue("hello"),
					CheckReceive:          types.StringValue("world"),
					ProxyURL:              types.StringValue(""),
					SslSupportedProtocols: types.ListNull(types.StringType),
				},
			},
		},
		{
			name: "Browser monitor",
			state: tfModelV0{
				Browser: &tfBrowserMonitorFieldsV0{
					InlineScript:   types.StringValue("aaa"),
					SyntheticsArgs: []types.String{types.StringValue("aaa"), types.StringValue("bbb")},
				},
			},
			input: kbapi.SyntheticsMonitor{
				Type: kbapi.Browser,
			},
			expected: tfModelV0{
				ID:             types.StringValue("/"),
				Name:           types.StringValue(""),
				SpaceID:        types.StringValue(""),
				Schedule:       types.Int64Value(0),
				APMServiceName: types.StringValue(""),
				TimeoutSeconds: types.Int64Value(0),
				Browser: &tfBrowserMonitorFieldsV0{
					InlineScript:   types.StringValue("aaa"),
					SyntheticsArgs: []types.String{types.StringValue("aaa"), types.StringValue("bbb")},
					Screenshots:    types.StringValue(""),
				},
			},
		},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			actual, diag := tt.state.toModelV0(ctx, &tt.input)
			assert.False(t, diag.HasError())
			assert.NotNil(t, actual)
			assert.Equal(t, &tt.expected, actual)
		})
	}
}
