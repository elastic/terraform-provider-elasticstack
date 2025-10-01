package kibana_oapi_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibana_oapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	fwdiag "github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/stretchr/testify/require"
)

func Test_connectorResponseToModel(t *testing.T) {
	type testCase struct {
		name          string
		spaceId       string
		response      *kbapi.ConnectorResponse
		expectedModel *models.KibanaActionConnector
		expectedError fwdiag.Diagnostics
	}
	tests := []testCase{
		{
			name:          "should return an error diag when response is nil",
			spaceId:       "default",
			response:      nil,
			expectedModel: nil,
			expectedError: fwdiag.Diagnostics{fwdiag.NewErrorDiagnostic("Invalid connector response", "connector response is nil")},
		},
		{
			name:    "should map valid connector response to model",
			spaceId: "default",
			response: &kbapi.ConnectorResponse{
				Id:               "test-id",
				ConnectorTypeId:  ".slack",
				Name:             "test-connector",
				IsPreconfigured:  false,
				IsDeprecated:     false,
				IsMissingSecrets: func() *bool { b := false; return &b }(),
				Config: func() *map[string]interface{} {
					m := map[string]interface{}{"webhookUrl": "https://hooks.slack.com/services/xxx"}
					return &m
				}(),
			},
			expectedModel: &models.KibanaActionConnector{
				ConnectorID:      "test-id",
				SpaceID:          "default",
				Name:             "test-connector",
				ConnectorTypeID:  ".slack",
				ConfigJSON:       `{"webhookUrl":"https://hooks.slack.com/services/xxx"}`,
				IsDeprecated:     false,
				IsMissingSecrets: false,
				IsPreconfigured:  false,
			},
			expectedError: nil,
		},
		{
			name:    "should handle empty config",
			spaceId: "default",
			response: &kbapi.ConnectorResponse{
				Id:               "empty-id",
				ConnectorTypeId:  ".webhook",
				Name:             "empty-connector",
				IsPreconfigured:  false,
				IsDeprecated:     false,
				IsMissingSecrets: func() *bool { b := false; return &b }(),
				Config:           nil,
			},
			expectedModel: &models.KibanaActionConnector{
				ConnectorID:      "empty-id",
				SpaceID:          "default",
				Name:             "empty-connector",
				ConnectorTypeID:  ".webhook",
				ConfigJSON:       "",
				IsDeprecated:     false,
				IsMissingSecrets: false,
				IsPreconfigured:  false,
			},
			expectedError: nil,
		},
		{
			name:    "should handle missing optional fields",
			spaceId: "default",
			response: &kbapi.ConnectorResponse{
				Id:              "missing-fields",
				ConnectorTypeId: ".webhook",
				Name:            "missing-connector",
			},
			expectedModel: &models.KibanaActionConnector{
				ConnectorID:      "missing-fields",
				SpaceID:          "default",
				Name:             "missing-connector",
				ConnectorTypeID:  ".webhook",
				ConfigJSON:       "",
				IsDeprecated:     false,
				IsMissingSecrets: false,
				IsPreconfigured:  false,
			},
			expectedError: nil,
		},
		{
			name:    "should handle non-default spaceId",
			spaceId: "custom-space",
			response: &kbapi.ConnectorResponse{
				Id:               "custom-id",
				ConnectorTypeId:  ".webhook",
				Name:             "custom-connector",
				IsPreconfigured:  true,
				IsDeprecated:     true,
				IsMissingSecrets: func() *bool { b := true; return &b }(),
				Config: func() *map[string]interface{} {
					m := map[string]interface{}{"url": "https://example.com"}
					return &m
				}(),
			},
			expectedModel: &models.KibanaActionConnector{
				ConnectorID:      "custom-id",
				SpaceID:          "custom-space",
				Name:             "custom-connector",
				ConnectorTypeID:  ".webhook",
				ConfigJSON:       `{"url":"https://example.com"}`,
				IsDeprecated:     true,
				IsMissingSecrets: true,
				IsPreconfigured:  true,
			},
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model, err := kibana_oapi.ConnectorResponseToModel(tt.spaceId, tt.response)

			if tt.expectedError == nil {
				require.Nil(t, err)
				require.Equal(t, tt.expectedModel, model)
			} else {
				require.Equal(t, tt.expectedError, err)
			}
		})
	}
}

func TestGetConnectorByName(t *testing.T) {
	const getConnectorsResponse = `[
		{
			"id": "c55b6eb0-6bad-11eb-9f3b-611eebc6c3ad",
			"connector_type_id": ".index",
			"name": "my-connector",
			"config": {
			"index": "test-index",
			"refresh": false,
			"executionTimeField": null
			},
			"is_preconfigured": false,
			"is_deprecated": false,
			"is_missing_secrets": false,
			"referenced_by_count": 3
		},
		{
			"id": "d55b6eb0-6bad-11eb-9f3b-611eebc6c3ad",
			"connector_type_id": ".index",
			"name": "doubledup-connector",
			"config": {
				"index": "test-index",
				"refresh": false,
				"executionTimeField": null
			},
			"is_preconfigured": false,
			"is_deprecated": false,
			"is_missing_secrets": false,
			"referenced_by_count": 3
		  },
		  {
			"id": "855b6eb0-6bad-11eb-9f3b-611eebc6c3ad",
			"connector_type_id": ".index",
			"name": "doubledup-connector",
			"config": {
			  "index": "test-index",
			  "refresh": false,
			  "executionTimeField": null
			},
			"is_preconfigured": false,
			"is_deprecated": false,
			"is_missing_secrets": false,
			"referenced_by_count": 0
		  }
	  ]`

	const emptyConnectorsResponse = `[]`

	var requests []*http.Request
	var mockResponses []string
	var httpStatus int
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		requests = append(requests, req)

		if len(mockResponses) > 0 {
			r := []byte(mockResponses[0])
			rw.Header().Add("X-Elastic-Product", "Elasticsearch")
			rw.Header().Add("Content-Type", "application/json")
			rw.WriteHeader(httpStatus)
			_, err := rw.Write(r)
			require.NoError(t, err)
			mockResponses = mockResponses[1:]
		} else {
			t.Fatalf("Unexpected request: %s %s", req.Method, req.URL.Path)
		}
	}))
	defer server.Close()

	httpStatus = http.StatusOK
	mockResponses = append(mockResponses, getConnectorsResponse)

	err := os.Setenv("ELASTICSEARCH_URL", server.URL)
	require.NoError(t, err)
	err = os.Setenv("KIBANA_ENDPOINT", server.URL)
	require.NoError(t, err)

	apiClient, err := clients.NewAcceptanceTestingClient()
	require.NoError(t, err)

	oapiClient, err := apiClient.GetKibanaOapiClient()
	require.NoError(t, err)

	connector, diags := kibana_oapi.SearchConnectors(context.Background(), oapiClient, "my-connector", "default", "")
	require.Nil(t, diags)
	require.NotNil(t, connector)

	mockResponses = append(mockResponses, getConnectorsResponse)
	failConnector, diags := kibana_oapi.SearchConnectors(context.Background(), oapiClient, "failwhale", "default", "")
	require.Nil(t, diags)
	require.Empty(t, failConnector)

	mockResponses = append(mockResponses, getConnectorsResponse)
	dupConnector, diags := kibana_oapi.SearchConnectors(context.Background(), oapiClient, "doubledup-connector", "default", "")
	require.Nil(t, diags)
	require.Len(t, dupConnector, 2)

	mockResponses = append(mockResponses, getConnectorsResponse)
	wrongConnectorType, diags := kibana_oapi.SearchConnectors(context.Background(), oapiClient, "my-connector", "default", ".slack")
	require.Nil(t, diags)
	require.Empty(t, wrongConnectorType)

	mockResponses = append(mockResponses, getConnectorsResponse)
	successConnector, diags := kibana_oapi.SearchConnectors(context.Background(), oapiClient, "my-connector", "default", ".index")
	require.Nil(t, diags)
	require.Len(t, successConnector, 1)

	mockResponses = append(mockResponses, emptyConnectorsResponse)
	emptyConnector, diags := kibana_oapi.SearchConnectors(context.Background(), oapiClient, "my-connector", "default", "")
	require.Nil(t, diags)
	require.Empty(t, emptyConnector)

	httpStatus = http.StatusBadGateway
	mockResponses = append(mockResponses, emptyConnectorsResponse)
	fail, diags := kibana_oapi.SearchConnectors(context.Background(), oapiClient, "my-connector", "default", "")
	require.NotNil(t, diags)
	require.Nil(t, fail)
}
