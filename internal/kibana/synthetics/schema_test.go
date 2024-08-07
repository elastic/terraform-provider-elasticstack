package synthetics

import (
	"encoding/json"
	"github.com/disaster37/go-kibana-rest/v8/kbapi"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"testing"
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
	apiMonitorHTTP := &kbapi.SyntheticsMonitor{
		Id:        "test-id-http",
		Name:      "test-name-http",
		Namespace: "default",
		Schedule:  &kbapi.MonitorScheduleConfig{Number: "5", Unit: "m"},
		Locations: []kbapi.MonitorLocationConfig{
			{Label: "us_east", IsServiceManaged: true},
			{Label: "test private location", IsServiceManaged: false},
		},
		Enabled:               tBool,
		Tags:                  []string{"tag1", "tag2"},
		Alert:                 &kbapi.MonitorAlertConfig{Status: &kbapi.SyntheticsStatusConfig{Enabled: tBool}, Tls: &kbapi.SyntheticsStatusConfig{Enabled: fBool}},
		APMServiceName:        "test-service-http",
		Timeout:               json.Number("30"),
		Params:                kbapi.JsonObject{"param1": "value1"},
		Type:                  kbapi.Http,
		Url:                   "https://example.com",
		SslVerificationMode:   "full",
		SslSupportedProtocols: []string{"TLSv1.2", "TLSv1.3"},
		MaxRedirects:          "5",
		Mode:                  kbapi.HttpMonitorMode("all"),
		Ipv4:                  tBool,
		Ipv6:                  fBool,
		Username:              "user",
		Password:              "pass",
		ProxyHeaders:          kbapi.JsonObject{"header1": "value1"},
		ProxyUrl:              "https://proxy.com",
	}

	expectedModelHTTP := &tfModelV0{
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
			URL:                   types.StringValue("https://example.com"),
			SslVerificationMode:   types.StringValue("full"),
			SslSupportedProtocols: []types.String{types.StringValue("TLSv1.2"), types.StringValue("TLSv1.3")},
			MaxRedirects:          types.StringValue("5"),
			Mode:                  types.StringValue("all"),
			IPv4:                  types.BoolPointerValue(tBool),
			IPv6:                  types.BoolPointerValue(fBool),
			Username:              types.StringValue("user"),
			Password:              types.StringValue("pass"),
			ProxyHeader:           jsontypes.NewNormalizedValue(`{"header1":"value1"}`),
			ProxyURL:              types.StringValue("https://proxy.com"),
		},
	}

	modelHTTP, err := toModelV0(apiMonitorHTTP)
	assert.NoError(t, err)
	assert.Equal(t, expectedModelHTTP, modelHTTP)

	// Test case for TCP fields
	apiMonitorTCP := &kbapi.SyntheticsMonitor{
		Id:        "test-id-tcp",
		Name:      "test-name-tcp",
		Namespace: "default",
		Schedule:  &kbapi.MonitorScheduleConfig{Number: "5", Unit: "m"},
		Locations: []kbapi.MonitorLocationConfig{
			{Label: "test private location", IsServiceManaged: false},
		},
		Enabled:               tBool,
		Tags:                  nil,
		Alert:                 &kbapi.MonitorAlertConfig{Status: &kbapi.SyntheticsStatusConfig{Enabled: tBool}},
		APMServiceName:        "test-service-tcp",
		Timeout:               json.Number("30"),
		Params:                kbapi.JsonObject{"param1": "value1"},
		Type:                  kbapi.Tcp,
		Host:                  "example.com:9200",
		SslVerificationMode:   "full",
		SslSupportedProtocols: []string{"TLSv1.2", "TLSv1.3"},
		CheckSend:             "hello",
		CheckReceive:          "world",
		ProxyUrl:              "http://proxy.com",
		ProxyUseLocalResolver: tBool,
	}

	expectedModelTCP := &tfModelV0{
		ID:               types.StringValue("test-id-tcp"),
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
			Host:                  types.StringValue("example.com:9200"),
			SslVerificationMode:   types.StringValue("full"),
			SslSupportedProtocols: []types.String{types.StringValue("TLSv1.2"), types.StringValue("TLSv1.3")},
			CheckSend:             types.StringValue("hello"),
			CheckReceive:          types.StringValue("world"),
			ProxyURL:              types.StringValue("http://proxy.com"),
			ProxyUseLocalResolver: types.BoolPointerValue(tBool),
		},
	}

	modelTCP, err := toModelV0(apiMonitorTCP)
	assert.NoError(t, err)
	assert.Equal(t, expectedModelTCP, modelTCP)
}

/*
func TestToKibanaAPIRequest(t *testing.T) {
	// Test case for HTTP monitor config
	modelHTTP := &tfModelV0{
		ID:               types.StringValue("test-id-http"),
		Name:             types.StringValue("test-name-http"),
		SpaceID:          types.StringValue("default"),
		Schedule:         types.Int64Value(5),
		Locations:        []types.String{types.StringValue("us_east")},
		PrivateLocations: []types.String{},
		//Enabled:          types.BoolPointerValue(true),
		Tags: []types.String{types.StringValue("tag1"), types.StringValue("tag2")},
		//Alert:            &tfAlertConfigV0{Status: &tfStatusConfigV0{Enabled: types.BoolPointerValue(true)}},
		APMServiceName: types.StringValue("test-service-http"),
		TimeoutSeconds: types.Int64Value(30),
		Params:         jsontypes.NewNormalizedValue(`{"param1":"value1"}`),
		HTTP: &tfHTTPMonitorFieldsV0{
			URL:                   types.StringValue("http://example.com"),
			SslVerificationMode:   types.StringValue("full"),
			SslSupportedProtocols: []types.String{types.StringValue("TLSv1.2"), types.StringValue("TLSv1.3")},
			MaxRedirects:          types.StringValue("5"),
			Mode:                  types.StringValue("all"),
			//IPv4:                  types.BoolPointerValue(true),
			//IPv6:                  types.BoolPointerValue(false),
			Username:    types.StringValue("user"),
			Password:    types.StringValue("pass"),
			ProxyHeader: jsontypes.NewNormalizedValue(`{"header1":"value1"}`),
			ProxyURL:    types.StringValue("http://proxy.com"),
			Response:    jsontypes.NewNormalizedValue(`{"response1":"value1"}`),
			Check:       jsontypes.NewNormalizedValue(`{"check1":"value1"}`),
		},
	}

	expectedAPIRequestHTTP := &kibanaAPIRequest{
		fields: &kbapi.HTTPMonitorFields{
			//URL:                   "http://example.com",
			SslVerificationMode:   "full",
			SslSupportedProtocols: []string{"TLSv1.2", "TLSv1.3"},
			MaxRedirects:          "5",
			Mode:                  "all",
			//IPv4:                  true,
			//IPv6:                  false,
			Username:    "user",
			Password:    "pass",
			ProxyHeader: kbapi.JsonObject{"header1": "value1"},
			//ProxyURL:              "http://proxy.com",
			Response: kbapi.JsonObject{"response1": "value1"},
			Check:    kbapi.JsonObject{"check1": "value1"},
		},
		config: kbapi.SyntheticsMonitorConfig{
			Name:     "test-name-http",
			Schedule: kbapi.MonitorSchedule(5),
			//Locations:        []kbapi.MonitorLocation{{Label: "us_east", IsServiceManaged: true}},
			PrivateLocations: []string{},
			//Enabled:          true,
			Tags: []string{"tag1", "tag2"},
			//Alert:          &kbapi.MonitorAlertConfig{Status: &kbapi.SyntheticsStatusConfig{Enabled: true}},
			APMServiceName: "test-service-http",
			TimeoutSeconds: 30,
			Params:         kbapi.JsonObject{"param1": "value1"},
		},
	}

	apiRequestHTTP, dg := modelHTTP.toKibanaAPIRequest()
	assert.False(t, dg.HasError())
	assert.Equal(t, expectedAPIRequestHTTP, apiRequestHTTP)

	// Test case for TCP monitor config
	modelTCP := &tfModelV0{
		ID:               types.StringValue("test-id-tcp"),
		Name:             types.StringValue("test-name-tcp"),
		SpaceID:          types.StringValue("default"),
		Schedule:         types.Int64Value(5),
		Locations:        []types.String{types.StringValue("us_east")},
		PrivateLocations: []types.String{},
		//Enabled:          types.BoolPointerValue(true),
		Tags: []types.String{types.StringValue("tag1"), types.StringValue("tag2")},
		//Alert:            &tfAlertConfigV0{Status: &tfStatusConfigV0{Enabled: types.BoolPointerValue(true)}},
		APMServiceName: types.StringValue("test-service-tcp"),
		TimeoutSeconds: types.Int64Value(30),
		Params:         jsontypes.NewNormalizedValue(`{"param1":"value1"}`),
		TCP: &tfTCPMonitorFieldsV0{
			Host:                  types.StringValue("example.com:9200"),
			SslVerificationMode:   types.StringValue("full"),
			SslSupportedProtocols: []types.String{types.StringValue("TLSv1.2"), types.StringValue("TLSv1.3")},
			CheckSend:             types.StringValue("hello"),
			CheckReceive:          types.StringValue("world"),
			ProxyURL:              types.StringValue("http://proxy.com"),
			ProxyUseLocalResolver: types.BoolPointerValue(true),
		},
	}

	expectedAPIRequestTCP := &kibanaAPIRequest{
		fields: &kbapi.TCPMonitorFields{
			Host:                  "example.com:9200",
			SslVerificationMode:   "full",
			SslSupportedProtocols: []string{"TLSv1.2", "TLSv1.3"},
			CheckSend:             "hello",
			CheckReceive:          "world",
			//ProxyURL:              "http://proxy.com",
			//ProxyUseLocalResolver: true,
		},
		config: kbapi.SyntheticsMonitorConfig{
			Name:     "test-name-tcp",
			Schedule: kbapi.MonitorSchedule(5),
			//Locations:        []kbapi.MonitorLocation{{Label: "us_east", IsServiceManaged: true}},
			PrivateLocations: []string{},
			//Enabled:          true,
			Tags: []string{"tag1", "tag2"},
			//Alert:          &kbapi.MonitorAlertConfig{Status: &kbapi.SyntheticsStatusConfig{Enabled: true}},
			APMServiceName: "test-service-tcp",
			TimeoutSeconds: 30,
			Params:         kbapi.JsonObject{"param1": "value1"},
		},
	}

	apiRequestTCP, dg := modelTCP.toKibanaAPIRequest()
	assert.False(t, dg.HasError())
	assert.Equal(t, expectedAPIRequestTCP, apiRequestTCP)
}
*/
