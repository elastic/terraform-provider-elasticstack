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

func TestConnectorConfigWithDefaults(t *testing.T) {
	tests := []struct {
		name            string
		connectorTypeID string
		planConfig      string
		expectedError   bool
		errorContains   string
		validateResult  func(t *testing.T, result string)
	}{
		{
			name:            "bedrock connector with valid config and explicit defaultModel",
			connectorTypeID: ".bedrock",
			planConfig:      `{"apiUrl":"https://bedrock.us-east-1.amazonaws.com","defaultModel":"anthropic.claude-v2"}`,
			expectedError:   false,
			validateResult: func(t *testing.T, result string) {
				expected := `{"apiUrl":"https://bedrock.us-east-1.amazonaws.com","defaultModel":"anthropic.claude-v2"}`
				require.JSONEq(t, expected, result)
			},
		},
		{
			name:            "bedrock connector without defaultModel gets default value",
			connectorTypeID: ".bedrock",
			planConfig:      `{"apiUrl":"https://bedrock.us-east-1.amazonaws.com"}`,
			expectedError:   false,
			validateResult: func(t *testing.T, result string) {
				expected := `{"apiUrl":"https://bedrock.us-east-1.amazonaws.com","defaultModel":"us.anthropic.claude-sonnet-4-5-20250929-v1:0"}`
				require.JSONEq(t, expected, result)
			},
		},
		{
			name:            "gen-ai connector with OpenAI provider with defaultModel",
			connectorTypeID: ".gen-ai",
			planConfig:      `{"apiProvider":"OpenAI","apiUrl":"https://api.openai.com/v1","defaultModel":"gpt-4"}`,
			expectedError:   false,
			validateResult: func(t *testing.T, result string) {
				expected := `{"apiProvider":"OpenAI","apiUrl":"https://api.openai.com/v1","defaultModel":"gpt-4"}`
				require.JSONEq(t, expected, result)
			},
		},
		{
			name:            "gen-ai connector with Azure provider",
			connectorTypeID: ".gen-ai",
			planConfig:      `{"apiProvider":"Azure OpenAI","apiUrl":"https://my-resource.openai.azure.com/openai/deployments/my-deployment"}`,
			expectedError:   false,
			validateResult: func(t *testing.T, result string) {
				expected := `{"apiProvider":"Azure OpenAI","apiUrl":"https://my-resource.openai.azure.com/openai/deployments/my-deployment"}`
				require.JSONEq(t, expected, result)
			},
		},
		{
			name:            "gen-ai connector with Other provider and explicit verificationMode",
			connectorTypeID: ".gen-ai",
			planConfig:      `{"apiProvider":"Other","apiUrl":"https://custom-llm.example.com/v1","defaultModel":"custom-model","verificationMode":"none"}`,
			expectedError:   false,
			validateResult: func(t *testing.T, result string) {
				expected := `{"apiProvider":"Other","apiUrl":"https://custom-llm.example.com/v1","defaultModel":"custom-model","verificationMode":"none"}`
				require.JSONEq(t, expected, result)
			},
		},
		{
			name:            "gen-ai connector with Other provider without verificationMode gets default",
			connectorTypeID: ".gen-ai",
			planConfig:      `{"apiProvider":"Other","apiUrl":"https://custom-llm.example.com/v1","defaultModel":"custom-model"}`,
			expectedError:   false,
			validateResult: func(t *testing.T, result string) {
				expected := `{"apiProvider":"Other","apiUrl":"https://custom-llm.example.com/v1","defaultModel":"custom-model","verificationMode":"full"}`
				require.JSONEq(t, expected, result)
			},
		},
		{
			name:            "gen-ai connector with OpenAI provider without defaultModel",
			connectorTypeID: ".gen-ai",
			planConfig:      `{"apiProvider":"OpenAI","apiUrl":"https://api.openai.com/v1"}`,
			expectedError:   false,
			validateResult: func(t *testing.T, result string) {
				// Verify no verificationMode is added (that's only for Other provider)
				expected := `{"apiProvider":"OpenAI","apiUrl":"https://api.openai.com/v1"}`
				require.JSONEq(t, expected, result)
			},
		},
		{
			name:            "gemini connector with valid config",
			connectorTypeID: ".gemini",
			planConfig:      `{"apiUrl":"https://us-central1-aiplatform.googleapis.com","gcpProjectId":"my-project","gcpRegion":"us-central1","defaultModel":"gemini-pro"}`,
			expectedError:   false,
			validateResult: func(t *testing.T, result string) {
				expected := `{"apiUrl":"https://us-central1-aiplatform.googleapis.com","gcpProjectId":"my-project","gcpRegion":"us-central1","defaultModel":"gemini-pro"}`
				require.JSONEq(t, expected, result)
			},
		},
		{
			name:            "gen-ai OpenAI connector silently filters unknown fields",
			connectorTypeID: ".gen-ai",
			planConfig:      `{"apiProvider":"OpenAI","apiUrl":"https://api.openai.com/v1","defaultModel":"gpt-4","unknownField":"should-be-filtered"}`,
			expectedError:   false,
			validateResult: func(t *testing.T, result string) {
				// Unknown field should be filtered out
				expected := `{"apiProvider":"OpenAI","apiUrl":"https://api.openai.com/v1","defaultModel":"gpt-4"}`
				require.JSONEq(t, expected, result)
			},
		},
		{
			name:            "gen-ai Azure connector silently filters invalid PKI fields",
			connectorTypeID: ".gen-ai",
			planConfig:      `{"apiProvider":"Azure OpenAI","apiUrl":"https://my.openai.azure.com","certificateData":"invalid-for-azure"}`,
			expectedError:   false,
			validateResult: func(t *testing.T, result string) {
				// certificateData is not valid for Azure provider, should be filtered
				expected := `{"apiProvider":"Azure OpenAI","apiUrl":"https://my.openai.azure.com"}`
				require.JSONEq(t, expected, result)
			},
		},
		{
			name:            "gen-ai Other connector allows PKI fields",
			connectorTypeID: ".gen-ai",
			planConfig:      `{"apiProvider":"Other","apiUrl":"https://custom.com","defaultModel":"custom","certificateData":"pem-data","verificationMode":"full"}`,
			expectedError:   false,
			validateResult: func(t *testing.T, result string) {
				expected := `{"apiProvider":"Other","apiUrl":"https://custom.com","defaultModel":"custom","certificateData":"pem-data","verificationMode":"full"}`
				require.JSONEq(t, expected, result)
			},
		},
		{
			name:            "gen-ai connector without apiProvider returns error",
			connectorTypeID: ".gen-ai",
			planConfig:      `{"apiUrl":"https://api.example.com"}`,
			expectedError:   true,
			errorContains:   "apiProvider is required",
		},
		{
			name:            "gen-ai connector with unknown apiProvider returns error",
			connectorTypeID: ".gen-ai",
			planConfig:      `{"apiProvider":"UnknownProvider","apiUrl":"https://api.example.com"}`,
			expectedError:   true,
			errorContains:   "unsupported apiProvider",
		},
		{
			name:            "unknown connector type returns error",
			connectorTypeID: ".unknown-type",
			planConfig:      `{"key":"value"}`,
			expectedError:   true,
			errorContains:   "unknown connector type ID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := kibana_oapi.ConnectorConfigWithDefaults(tt.connectorTypeID, tt.planConfig)

			if tt.expectedError {
				require.Error(t, err)
				if tt.errorContains != "" {
					require.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				require.NoError(t, err)
				require.NotEmpty(t, result)
				if tt.validateResult != nil {
					tt.validateResult(t, result)
				}
			}
		})
	}
}
