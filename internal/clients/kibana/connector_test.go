package kibana

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/connectors"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/stretchr/testify/require"
)

func Test_connectorResponseToModel(t *testing.T) {
	type testCase struct {
		name          string
		spaceId       string
		response      connectors.ConnectorResponseProperties
		expectedModel *models.KibanaActionConnector
		expectedError error
	}

	generator := func(connectorTypeID string, config any, propertiesGenerator func(*connectors.ConnectorResponseProperties) error) testCase {
		return testCase{
			name:    fmt.Sprintf("it should parse empty [%s] connector", connectorTypeID),
			spaceId: "test",
			response: func() connectors.ConnectorResponseProperties {
				var properties connectors.ConnectorResponseProperties
				err := propertiesGenerator(&properties)
				require.Nil(t, err)
				return properties
			}(),
			expectedModel: &models.KibanaActionConnector{
				SpaceID:         "test",
				ConnectorTypeID: connectorTypeID,
				ConfigJSON: func() string {
					if config == nil {
						return ""
					}
					byt, err := json.Marshal(config)
					require.Nil(t, err)
					return string(byt)
				}(),
			},
		}
	}
	tests := []testCase{
		{
			name: "it should fail if discriminator is unknown",
			response: func() connectors.ConnectorResponseProperties {
				discriminator := struct {
					Discriminator string `json:"connector_type_id"`
				}{"unknown-value"}
				byt, err := json.Marshal(discriminator)
				require.Nil(t, err)
				var resp connectors.ConnectorResponseProperties
				err = resp.UnmarshalJSON(byt)
				require.Nil(t, err)
				return resp
			}(),
			expectedError: func() error { return fmt.Errorf("unknown connector type [unknown-value]") }(),
		},
		generator(".cases-webhook", connectors.ConfigPropertiesCasesWebhook{}, func(props *connectors.ConnectorResponseProperties) error {
			return props.FromConnectorResponsePropertiesCasesWebhook(connectors.ConnectorResponsePropertiesCasesWebhook{})
		}),
		generator(".email", connectors.ConfigPropertiesEmail{}, func(props *connectors.ConnectorResponseProperties) error {
			return props.FromConnectorResponsePropertiesEmail(connectors.ConnectorResponsePropertiesEmail{})
		}),
		generator(".gemini", connectors.ConfigPropertiesGemini{}, func(props *connectors.ConnectorResponseProperties) error {
			return props.FromConnectorResponsePropertiesGemini(connectors.ConnectorResponsePropertiesGemini{})
		}),
		generator(".index", connectors.ConfigPropertiesIndex{}, func(props *connectors.ConnectorResponseProperties) error {
			return props.FromConnectorResponsePropertiesIndex(connectors.ConnectorResponsePropertiesIndex{})
		}),
		generator(".jira", connectors.ConfigPropertiesJira{}, func(props *connectors.ConnectorResponseProperties) error {
			return props.FromConnectorResponsePropertiesJira(connectors.ConnectorResponsePropertiesJira{})
		}),
		generator(".opsgenie", connectors.ConfigPropertiesOpsgenie{}, func(props *connectors.ConnectorResponseProperties) error {
			return props.FromConnectorResponsePropertiesOpsgenie(connectors.ConnectorResponsePropertiesOpsgenie{})
		}),
		generator(".pagerduty", connectors.ConfigPropertiesPagerduty{}, func(props *connectors.ConnectorResponseProperties) error {
			return props.FromConnectorResponsePropertiesPagerduty(connectors.ConnectorResponsePropertiesPagerduty{})
		}),
		generator(".resilient", connectors.ConfigPropertiesResilient{}, func(props *connectors.ConnectorResponseProperties) error {
			return props.FromConnectorResponsePropertiesResilient(connectors.ConnectorResponsePropertiesResilient{})
		}),
		generator(".server-log", map[string]interface{}{}, func(props *connectors.ConnectorResponseProperties) error {
			return props.FromConnectorResponsePropertiesServerlog(connectors.ConnectorResponsePropertiesServerlog{
				Config: &map[string]interface{}{},
			})
		}),
		generator(".servicenow", connectors.ConfigPropertiesServicenow{}, func(props *connectors.ConnectorResponseProperties) error {
			return props.FromConnectorResponsePropertiesServicenow(connectors.ConnectorResponsePropertiesServicenow{})
		}),
		generator(".servicenow-itom", connectors.ConfigPropertiesServicenowItom{}, func(props *connectors.ConnectorResponseProperties) error {
			return props.FromConnectorResponsePropertiesServicenowItom(connectors.ConnectorResponsePropertiesServicenowItom{})
		}),
		generator(".servicenow-sir", connectors.ConfigPropertiesServicenow{}, func(props *connectors.ConnectorResponseProperties) error {
			return props.FromConnectorResponsePropertiesServicenowSir(connectors.ConnectorResponsePropertiesServicenowSir{})
		}),
		generator(".slack", nil, func(props *connectors.ConnectorResponseProperties) error {
			return props.FromConnectorResponsePropertiesSlack(connectors.ConnectorResponsePropertiesSlack{})
		}),
		generator(".slack_api", nil, func(props *connectors.ConnectorResponseProperties) error {
			return props.FromConnectorResponsePropertiesSlackApi(connectors.ConnectorResponsePropertiesSlackApi{})
		}),
		generator(".swimlane", connectors.ConfigPropertiesSwimlane{}, func(props *connectors.ConnectorResponseProperties) error {
			return props.FromConnectorResponsePropertiesSwimlane(connectors.ConnectorResponsePropertiesSwimlane{})
		}),
		generator(".teams", nil, func(props *connectors.ConnectorResponseProperties) error {
			return props.FromConnectorResponsePropertiesTeams(connectors.ConnectorResponsePropertiesTeams{})
		}),
		generator(".tines", connectors.ConfigPropertiesTines{}, func(props *connectors.ConnectorResponseProperties) error {
			return props.FromConnectorResponsePropertiesTines(connectors.ConnectorResponsePropertiesTines{})
		}),
		generator(".webhook", connectors.ConfigPropertiesWebhook{}, func(props *connectors.ConnectorResponseProperties) error {
			return props.FromConnectorResponsePropertiesWebhook(connectors.ConnectorResponsePropertiesWebhook{})
		}),
		generator(".xmatters", connectors.ConfigPropertiesXmatters{}, func(props *connectors.ConnectorResponseProperties) error {
			return props.FromConnectorResponsePropertiesXmatters(connectors.ConnectorResponsePropertiesXmatters{})
		}),
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model, err := connectorResponseToModel(tt.spaceId, tt.response)

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

	connector, diags := SearchConnectors(context.Background(), apiClient, "my-connector", "default", "")
	require.Nil(t, diags)
	require.NotNil(t, connector)

	mockResponses = append(mockResponses, getConnectorsResponse)
	failConnector, diags := SearchConnectors(context.Background(), apiClient, "failwhale", "default", "")
	require.Nil(t, diags)
	require.Empty(t, failConnector)

	mockResponses = append(mockResponses, getConnectorsResponse)
	dupConnector, diags := SearchConnectors(context.Background(), apiClient, "doubledup-connector", "default", "")
	require.Nil(t, diags)
	require.Len(t, dupConnector, 2)

	mockResponses = append(mockResponses, getConnectorsResponse)
	wrongConnectorType, diags := SearchConnectors(context.Background(), apiClient, "my-connector", "default", ".slack")
	require.Nil(t, diags)
	require.Empty(t, wrongConnectorType)

	mockResponses = append(mockResponses, getConnectorsResponse)
	successConnector, diags := SearchConnectors(context.Background(), apiClient, "my-connector", "default", ".index")
	require.Nil(t, diags)
	require.Len(t, successConnector, 1)

	mockResponses = append(mockResponses, emptyConnectorsResponse)
	emptyConnector, diags := SearchConnectors(context.Background(), apiClient, "my-connector", "default", "")
	require.Nil(t, diags)
	require.Empty(t, emptyConnector)

	httpStatus = http.StatusBadGateway
	mockResponses = append(mockResponses, emptyConnectorsResponse)
	fail, diags := SearchConnectors(context.Background(), apiClient, "my-connector", "default", "")
	require.NotNil(t, diags)
	require.Nil(t, fail)

}
