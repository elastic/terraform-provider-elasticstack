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
	"encoding/json"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	fBool = new(false)
	tBool = new(true)
)

func toAlertObject(t *testing.T, v tfAlertConfigV0) basetypes.ObjectValue {
	t.Helper()
	alertAttributes := monitorAlertConfigSchema().GetType().(attr.TypeWithAttributeTypes).AttributeTypes()
	from, dg := types.ObjectValueFrom(context.Background(), alertAttributes, &v)
	if dg.HasError() {
		t.Fatalf("Failed to create Alert object: %v", dg)
	}
	return from
}

func mustJSON(t *testing.T, v any) string {
	t.Helper()
	b, err := json.Marshal(v)
	require.NoError(t, err)
	return string(b)
}

func mustTimeoutString(t *testing.T, v string) *kbapi.SyntheticsMonitor_Timeout {
	t.Helper()
	timeout := &kbapi.SyntheticsMonitor_Timeout{}
	require.NoError(t, timeout.FromSyntheticsMonitorTimeout0(v))
	return timeout
}

func mustTimeoutNumber(t *testing.T, v float32) *kbapi.SyntheticsMonitor_Timeout {
	t.Helper()
	timeout := &kbapi.SyntheticsMonitor_Timeout{}
	require.NoError(t, timeout.FromSyntheticsMonitorTimeout1(kbapi.SyntheticsMonitorTimeout1(v)))
	return timeout
}

func mustWaitString(t *testing.T, v string) *kbapi.SyntheticsMonitor_Wait {
	t.Helper()
	wait := &kbapi.SyntheticsMonitor_Wait{}
	require.NoError(t, wait.FromSyntheticsMonitorWait0(v))
	return wait
}

func mustWaitNumber(t *testing.T, v float32) *kbapi.SyntheticsMonitor_Wait {
	t.Helper()
	wait := &kbapi.SyntheticsMonitor_Wait{}
	require.NoError(t, wait.FromSyntheticsMonitorWait1(kbapi.SyntheticsMonitorWait1(v)))
	return wait
}

func mustMaxRedirectsString(t *testing.T, v string) *kbapi.SyntheticsMonitor_MaxRedirects {
	t.Helper()
	maxRedirects := &kbapi.SyntheticsMonitor_MaxRedirects{}
	require.NoError(t, maxRedirects.FromSyntheticsMonitorMaxRedirects0(v))
	return maxRedirects
}

func mustMaxRedirectsNumber(t *testing.T, v float32) *kbapi.SyntheticsMonitor_MaxRedirects {
	t.Helper()
	maxRedirects := &kbapi.SyntheticsMonitor_MaxRedirects{}
	require.NoError(t, maxRedirects.FromSyntheticsMonitorMaxRedirects1(kbapi.SyntheticsMonitorMaxRedirects1(v)))
	return maxRedirects
}

func TestLabelsFieldConversion(t *testing.T) {
	testcases := []struct {
		name     string
		input    kbapi.SyntheticsMonitor
		expected types.Map
	}{
		{
			name: "monitor with nil labels",
			input: kbapi.SyntheticsMonitor{
				Type: new(kbapi.SyntheticsMonitorTypeHttp),
			},
			expected: types.MapNull(types.StringType),
		},
		{
			name: "monitor with empty labels",
			input: kbapi.SyntheticsMonitor{
				Type:   new(kbapi.SyntheticsMonitorTypeHttp),
				Labels: &map[string]string{},
			},
			expected: types.MapValueMust(types.StringType, map[string]attr.Value{}),
		},
		{
			name: "monitor with labels",
			input: kbapi.SyntheticsMonitor{
				Type: new(kbapi.SyntheticsMonitorTypeHttp),
				Labels: &map[string]string{
					"environment": "production",
					"team":        "platform",
					"service":     "web-app",
				},
			},
			expected: types.MapValueMust(types.StringType, map[string]attr.Value{
				"environment": types.StringValue("production"),
				"team":        types.StringValue("platform"),
				"service":     types.StringValue("web-app"),
			}),
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			model := &tfModelV0{}
			result, diags := model.toModelV0(context.Background(), &tc.input, "default")
			assert.False(t, diags.HasError())
			assert.Equal(t, tc.expected, result.Labels)
		})
	}
}

func TestToModelV0HTTP(t *testing.T) {
	api := kbapi.SyntheticsMonitor{
		Id:        new("test-id-http"),
		Name:      new("test-name-http"),
		Namespace: new("default"),
		Enabled:   tBool,
		Alert: &kbapi.SyntheticsMonitorAlert{
			Status: &kbapi.SyntheticsMonitorAlertStatus{Enabled: tBool},
			Tls:    &kbapi.SyntheticsMonitorAlertStatus{Enabled: fBool},
		},
		Schedule: &kbapi.SyntheticsMonitorSchedule{
			Number: new("5"),
			Unit:   new("m"),
		},
		Tags:        &[]string{"tag1", "tag2"},
		ServiceName: new("test-service-http"),
		Timeout:     mustTimeoutString(t, "30"),
		Locations: &[]kbapi.SyntheticsLocationConfig{
			{Id: new("us-east4-a"), Label: new("North America - US East"), IsServiceManaged: tBool},
			{Label: new("test private location"), IsServiceManaged: fBool},
		},
		Params:                    &map[string]any{"param1": "value1"},
		Type:                      new(kbapi.SyntheticsMonitorTypeHttp),
		Url:                       new("https://example.com"),
		Mode:                      new("all"),
		MaxRedirects:              mustMaxRedirectsString(t, "5"),
		Ipv4:                      tBool,
		Ipv6:                      fBool,
		Username:                  new("user"),
		Password:                  new("pass"),
		ProxyHeaders:              &map[string]any{"header1": "value1"},
		ProxyUrl:                  new("https://proxy.com"),
		SslVerificationMode:       new("full"),
		SslSupportedProtocols:     &[]string{"TLSv1.2", "TLSv1.3"},
		SslCertificateAuthorities: &[]string{"cert1", "cert2"},
		SslCertificate:            new("cert"),
		SslKey:                    new("key"),
		SslKeyPassphrase:          new("passphrase"),
	}

	model, diags := (&tfModelV0{}).toModelV0(context.Background(), &api, "default")
	assert.False(t, diags.HasError(), diags)
	require.NotNil(t, model)

	assert.Equal(t, types.StringValue("default/test-id-http"), model.ID)
	assert.Equal(t, types.StringValue("test-name-http"), model.Name)
	assert.Equal(t, types.Int64Value(5), model.Schedule)
	assert.Equal(t, []types.String{types.StringValue("test private location")}, model.PrivateLocations)
	assert.Equal(t, types.StringValue("https://example.com"), model.HTTP.URL)
	assert.Equal(t, types.StringValue("https://proxy.com"), model.HTTP.ProxyURL)
	assert.Equal(t, types.StringValue("full"), model.HTTP.SslVerificationMode)
	assert.Equal(t, types.StringValue("test-service-http"), model.APMServiceName)
	assert.Equal(t, types.Int64Value(30), model.TimeoutSeconds)
	assert.Equal(t, jsontypes.NewNormalizedValue(`{"param1":"value1"}`), model.Params)
}

func TestToKibanaAPIRequest(t *testing.T) {
	testcases := []struct {
		name         string
		input        tfModelV0
		expectedJSON string
	}{
		{
			name:         "empty HTTP monitor",
			input:        tfModelV0{HTTP: &tfHTTPMonitorFieldsV0{}},
			expectedJSON: `{"labels":{},"name":"","type":"http","url":""}`,
		},
		{
			name:         "empty TCP monitor",
			input:        tfModelV0{TCP: &tfTCPMonitorFieldsV0{}},
			expectedJSON: `{"host":"","labels":{},"name":"","type":"tcp"}`,
		},
		{
			name:         "empty ICMP monitor",
			input:        tfModelV0{ICMP: &tfICMPMonitorFieldsV0{}},
			expectedJSON: `{"host":"","labels":{},"name":"","type":"icmp"}`,
		},
		{
			name:         "empty Browser monitor",
			input:        tfModelV0{Browser: &tfBrowserMonitorFieldsV0{}},
			expectedJSON: `{"inline_script":"","labels":{},"name":"","type":"browser"}`,
		},
		{
			name: "HTTP monitor",
			input: tfModelV0{
				Name:             types.StringValue("test-name-http"),
				Namespace:        types.StringValue("default-3"),
				Schedule:         types.Int64Value(5),
				Locations:        []types.String{types.StringValue("us_east")},
				PrivateLocations: []types.String{types.StringValue("test private location")},
				Enabled:          types.BoolPointerValue(tBool),
				Tags:             []types.String{types.StringValue("tag1"), types.StringValue("tag2")},
				Alert:            toAlertObject(t, tfAlertConfigV0{Status: &tfStatusConfigV0{Enabled: types.BoolPointerValue(tBool)}, TLS: &tfStatusConfigV0{Enabled: types.BoolPointerValue(fBool)}}),
				APMServiceName:   types.StringValue("test-service-http"),
				TimeoutSeconds:   types.Int64Value(30),
				Params:           jsontypes.NewNormalizedValue(`{"param1":"value1"}`),
				HTTP: &tfHTTPMonitorFieldsV0{
					URL:          types.StringValue("https://example.com"),
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
					tfSSLConfig: tfSSLConfig{
						SslVerificationMode:       types.StringValue("full"),
						SslSupportedProtocols:     types.ListValueMust(types.StringType, []attr.Value{types.StringValue("TLSv1.2"), types.StringValue("TLSv1.3")}),
						SslCertificateAuthorities: []types.String{types.StringValue("cert1"), types.StringValue("cert2")},
						SslCertificate:            types.StringValue("cert"),
						SslKey:                    types.StringValue("key"),
						SslKeyPassphrase:          types.StringValue("passphrase"),
					},
				},
			},
			expectedJSON: `{
				"alert":{"status":{"enabled":true},"tls":{"enabled":false}},
				"enabled":true,
				"ipv4":true,
				"ipv6":false,
				"labels":{},
				"locations":["us_east"],
				"max_redirects":"5",
				"mode":"all",
				"name":"test-name-http",
				"namespace":"default-3",
				"params":{"param1":"value1"},
				"password":"pass",
				"private_locations":["test private location"],
				"proxy_headers":{"header1":"value1"},
				"proxy_url":"https://proxy.com",
				"response":{"response1":"value1"},
				"schedule":5,
				"service.name":"test-service-http",
				"ssl":{
					"certificate":"cert",
					"certificate_authorities":["cert1","cert2"],
					"key":"key",
					"key_passphrase":"passphrase",
					"supported_protocols":["TLSv1.2","TLSv1.3"],
					"verification_mode":"full"
				},
				"tags":["tag1","tag2"],
				"timeout":30,
				"type":"http",
				"url":"https://example.com",
				"username":"user",
				"check":{"check1":"value1"}
			}`,
		},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			apiRequest, dg := tt.input.toKibanaAPIRequest(context.Background())
			assert.False(t, dg.HasError(), dg.Errors())
			assert.JSONEq(t, tt.expectedJSON, mustJSON(t, apiRequest))
		})
	}
}

func TestToModelV0TimeoutNumber(t *testing.T) {
	api := kbapi.SyntheticsMonitor{
		Type:    new(kbapi.SyntheticsMonitorTypeHttp),
		Timeout: mustTimeoutNumber(t, 30),
	}

	model, diags := (&tfModelV0{}).toModelV0(context.Background(), &api, "default")
	assert.False(t, diags.HasError(), diags)
	require.NotNil(t, model)
	assert.Equal(t, types.Int64Value(30), model.TimeoutSeconds)
}

func TestToModelV0ICMPWaitString(t *testing.T) {
	api := kbapi.SyntheticsMonitor{
		Type: new(kbapi.SyntheticsMonitorTypeIcmp),
		Wait: mustWaitString(t, "2"),
	}

	model, diags := (&tfModelV0{}).toModelV0(context.Background(), &api, "default")
	assert.False(t, diags.HasError(), diags)
	require.NotNil(t, model)
	require.NotNil(t, model.ICMP)
	assert.Equal(t, types.Int64Value(2), model.ICMP.Wait)
}

func TestToModelV0ICMPWaitNumber(t *testing.T) {
	api := kbapi.SyntheticsMonitor{
		Type: new(kbapi.SyntheticsMonitorTypeIcmp),
		Wait: mustWaitNumber(t, 2),
	}

	model, diags := (&tfModelV0{}).toModelV0(context.Background(), &api, "default")
	assert.False(t, diags.HasError(), diags)
	require.NotNil(t, model)
	require.NotNil(t, model.ICMP)
	assert.Equal(t, types.Int64Value(2), model.ICMP.Wait)
}

func TestToModelV0HTTPMaxRedirectsNumber(t *testing.T) {
	api := kbapi.SyntheticsMonitor{
		Type:         new(kbapi.SyntheticsMonitorTypeHttp),
		MaxRedirects: mustMaxRedirectsNumber(t, 5),
	}

	model, diags := (&tfModelV0{}).toModelV0(context.Background(), &api, "default")
	assert.False(t, diags.HasError(), diags)
	require.NotNil(t, model)
	require.NotNil(t, model.HTTP)
	assert.Equal(t, types.Int64Value(5), model.HTTP.MaxRedirects)
}

func TestToKibanaAPIRequestICMPWait(t *testing.T) {
	input := tfModelV0{
		Name: types.StringValue("test-icmp-monitor"),
		ICMP: &tfICMPMonitorFieldsV0{
			Host: types.StringValue("example.com"),
			Wait: types.Int64Value(10),
		},
	}

	result, diags := input.toKibanaAPIRequest(context.Background())
	assert.False(t, diags.HasError(), diags)
	assert.JSONEq(t, `{"host":"example.com","labels":{},"name":"test-icmp-monitor","type":"icmp","wait":"10"}`, mustJSON(t, result))
}

func TestToModelV0MergeAttributes(t *testing.T) {
	state := tfModelV0{
		HTTP: &tfHTTPMonitorFieldsV0{
			ProxyHeader: jsontypes.NewNormalizedValue(`{"header1":"value1"}`),
			Username:    types.StringValue("test"),
			Password:    types.StringValue("password"),
			Check:       jsontypes.NewNormalizedValue(`{"check1":"value1"}`),
			Response:    jsontypes.NewNormalizedValue(`{"response1":"value1"}`),
		},
		Params:          jsontypes.NewNormalizedValue(`{"param1":"value1"}`),
		RetestOnFailure: types.BoolValue(true),
	}
	input := kbapi.SyntheticsMonitor{
		Type: new(kbapi.SyntheticsMonitorTypeHttp),
	}

	actual, diags := state.toModelV0(context.Background(), &input, "")
	assert.False(t, diags.HasError())
	require.NotNil(t, actual)
	assert.Equal(t, state.HTTP.ProxyHeader, actual.HTTP.ProxyHeader)
	assert.Equal(t, state.HTTP.Username, actual.HTTP.Username)
	assert.Equal(t, state.HTTP.Password, actual.HTTP.Password)
	assert.Equal(t, state.HTTP.Check, actual.HTTP.Check)
	assert.Equal(t, state.HTTP.Response, actual.HTTP.Response)
	assert.Equal(t, state.Params, actual.Params)
	assert.Equal(t, state.RetestOnFailure, actual.RetestOnFailure)
}

func TestToKibanaAPIRequestLabels(t *testing.T) {
	input := tfModelV0{
		Name:   types.StringValue("test-monitor"),
		Labels: types.MapValueMust(types.StringType, map[string]attr.Value{"environment": types.StringValue("production"), "team": types.StringValue("platform")}),
		HTTP:   &tfHTTPMonitorFieldsV0{},
	}

	result, diags := input.toKibanaAPIRequest(context.Background())
	assert.False(t, diags.HasError())

	var payload map[string]any
	require.NoError(t, json.Unmarshal([]byte(mustJSON(t, result)), &payload))
	assert.Equal(t, "test-monitor", payload["name"])
	assert.Equal(t, map[string]any{"environment": "production", "team": "platform"}, payload["labels"])
}

func TestAlertConfigRoundTrip(t *testing.T) {
	alert := tfAlertConfigV0{
		Status: &tfStatusConfigV0{Enabled: types.BoolPointerValue(tBool)},
		TLS:    &tfStatusConfigV0{Enabled: types.BoolPointerValue(fBool)},
	}

	apiAlert := alert.toAPIAlertConfig()
	require.NotNil(t, apiAlert)
	require.NotNil(t, apiAlert.Status)
	require.NotNil(t, apiAlert.Tls)
	assert.True(t, *apiAlert.Status.Enabled)
	assert.False(t, *apiAlert.Tls.Enabled)

	tfAlert, diags := toTfAlertConfigV0(context.Background(), apiAlert)
	assert.False(t, diags.HasError())
	assert.False(t, tfAlert.IsNull())
}

func TestUnionToInt64(t *testing.T) {
	strVariant := func(s string) func() (string, error) {
		return func() (string, error) { return s, nil }
	}
	floatVariant := func(n float32) func() (float32, error) {
		return func() (float32, error) { return n, nil }
	}
	errVariant := func() (string, error) { return "", assert.AnError }
	errFloatVariant := func() (float32, error) { return 0, assert.AnError }

	tests := []struct {
		name      string
		asStr     func() (string, error)
		asFloat   func() (float32, error)
		fieldName string
		want      int64
		wantErr   bool
	}{
		{
			name:      "string integer",
			asStr:     strVariant("42"),
			asFloat:   errFloatVariant,
			fieldName: "timeout",
			want:      42,
		},
		{
			name:      "string empty",
			asStr:     strVariant(""),
			asFloat:   errFloatVariant,
			fieldName: "timeout",
			want:      0,
		},
		{
			name:      "float whole number",
			asStr:     errVariant,
			asFloat:   floatVariant(30),
			fieldName: "wait",
			want:      30,
		},
		{
			name:      "float fractional rejected",
			asStr:     errVariant,
			asFloat:   floatVariant(1.5),
			fieldName: "timeout",
			wantErr:   true,
		},
		{
			name:      "both fail",
			asStr:     errVariant,
			asFloat:   errFloatVariant,
			fieldName: "max_redirects",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := unionToInt64(tt.asStr, tt.asFloat, tt.fieldName)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}
