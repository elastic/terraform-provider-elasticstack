package kibana

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/elastic/terraform-provider-elasticstack/generated/connectors"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
)

func CreateConnector(ctx context.Context, apiClient *clients.ApiClient, connectorOld models.KibanaActionConnector) (string, diag.Diagnostics) {
	client, err := apiClient.GetKibanaConnectorsClient(ctx)
	if err != nil {
		return "", diag.FromErr(err)
	}

	body, err := createConnectorRequestBody(connectorOld)
	if err != nil {
		return "", diag.FromErr(err)
	}

	httpResp, err := client.CreateConnectorWithBody(ctx, connectorOld.SpaceID, &connectors.CreateConnectorParams{KbnXsrf: connectors.KbnXsrf("true")}, "application/json", body)

	if err != nil {
		return "", diag.Errorf("unable to create connector: [%v]", err)
	}

	defer httpResp.Body.Close()

	resp, err := connectors.ParseCreateConnectorResponse(httpResp)
	if err != nil {
		return "", diag.Errorf("unable to parse connector create response: [%v]", err)
	}

	if resp.JSON400 != nil {
		return "", diag.Errorf("%s: %s", *resp.JSON400.Error, *resp.JSON400.Message)
	}

	if resp.JSON401 != nil {
		return "", diag.Errorf("%s: %s", *resp.JSON401.Error, *resp.JSON401.Message)
	}

	if resp.JSON200 == nil {
		return "", diag.Errorf("%s: %s", resp.Status(), string(resp.Body))
	}

	connectorNew, err := connectorResponseToModel(connectorOld.SpaceID, *resp.JSON200)
	if err != nil {
		return "", diag.FromErr(err)
	}

	return connectorNew.ConnectorID, nil
}

func UpdateConnector(ctx context.Context, apiClient *clients.ApiClient, connectorOld models.KibanaActionConnector) (string, diag.Diagnostics) {
	client, err := apiClient.GetKibanaConnectorsClient(ctx)
	if err != nil {
		return "", diag.FromErr(err)
	}

	body, err := updateConnectorRequestBody(connectorOld)
	if err != nil {
		return "", diag.FromErr(err)
	}

	httpResp, err := client.UpdateConnectorWithBody(ctx, connectorOld.SpaceID, connectorOld.ConnectorID, &connectors.UpdateConnectorParams{KbnXsrf: connectors.KbnXsrf("true")}, "application/json", body)

	if err != nil {
		return "", diag.Errorf("unable to update connector: [%v]", err)
	}

	defer httpResp.Body.Close()

	resp, err := connectors.ParseCreateConnectorResponse(httpResp)
	if err != nil {
		return "", diag.Errorf("unable to parse connector update response: [%v]", err)
	}

	if resp.JSON400 != nil {
		return "", diag.Errorf("%s: %s", *resp.JSON400.Error, *resp.JSON400.Message)
	}

	if resp.JSON401 != nil {
		return "", diag.Errorf("%s: %s", *resp.JSON401.Error, *resp.JSON401.Message)
	}

	if resp.JSON200 == nil {
		return "", diag.Errorf("%s: %s", resp.Status(), string(resp.Body))
	}

	connectorNew, err := connectorResponseToModel(connectorOld.SpaceID, *resp.JSON200)
	if err != nil {
		return "", diag.FromErr(err)
	}

	return connectorNew.ConnectorID, nil
}

func GetConnector(ctx context.Context, apiClient *clients.ApiClient, connectorID, spaceID string, connectorTypeID string) (*models.KibanaActionConnector, diag.Diagnostics) {
	client, err := apiClient.GetKibanaConnectorsClient(ctx)
	if err != nil {
		return nil, diag.FromErr(err)
	}

	httpResp, err := client.GetConnector(ctx, spaceID, connectorID)

	if err != nil {
		return nil, diag.Errorf("unable to get connector: [%v]", err)
	}

	defer httpResp.Body.Close()

	resp, err := connectors.ParseGetConnectorResponse(httpResp)
	if err != nil {
		return nil, diag.Errorf("unable to parse connector get response: [%v]", err)
	}

	if resp.JSON401 != nil {
		return nil, diag.Errorf("%s: %s", *resp.JSON401.Error, *resp.JSON401.Message)
	}

	if resp.JSON404 != nil {
		return nil, nil
	}

	if resp.JSON200 == nil {
		return nil, diag.Errorf("%s: %s", resp.Status(), string(resp.Body))
	}

	connector, err := connectorResponseToModel(spaceID, *resp.JSON200)
	if err != nil {
		return nil, diag.Errorf("unable to convert response to model: %v", err)
	}

	return connector, nil
}

func DeleteConnector(ctx context.Context, apiClient *clients.ApiClient, connectorID string, spaceID string) diag.Diagnostics {
	client, err := apiClient.GetKibanaConnectorsClient(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	httpResp, err := client.DeleteConnector(ctx, spaceID, connectorID, &connectors.DeleteConnectorParams{KbnXsrf: "true"})

	if err != nil {
		return diag.Errorf("unable to delete connector: [%v]", err)
	}

	defer httpResp.Body.Close()

	resp, err := connectors.ParseDeleteConnectorResponse(httpResp)
	if err != nil {
		return diag.Errorf("unable to parse connector get response: [%v]", err)
	}

	if resp.JSON404 != nil {
		return diag.Errorf("%s: %s", *resp.JSON404.Error, *resp.JSON404.Message)
	}

	if resp.JSON401 != nil {
		return diag.Errorf("%s: %s", *resp.JSON401.Error, *resp.JSON401.Message)
	}

	if resp.StatusCode() != 200 && resp.StatusCode() != 204 {
		return diag.Errorf("failed to delete connector: got status [%v] [%s]", resp.StatusCode(), resp.Status())
	}

	return nil
}

func createConnectorRequestBody(connector models.KibanaActionConnector) (io.Reader, error) {
	switch connectors.ConnectorTypes(connector.ConnectorTypeID) {
	// case connectors.CASES_WEBHOOK_ConnectorTypes:
	// 	return createConnectorRequestCasesWebhook(connector)
	// case connectors.EMAIL_ConnectorTypes:
	//  	return createConnectorRequestEmail(connector)
	case connectors.ConnectorTypesDotIndex:
		return createConnectorRequestIndex(connector)
		// case connectors.JIRA_ConnectorTypes:
		// 	return createConnectorRequestJira(connector)
		// case connectors.OPSGENIE_ConnectorTypes:
		// 	return createConnectorRequestOpsgenie(connector)
		// case connectors.PAGERDUTY:
		// 	return createConnectorRequestPagerduty(connector)
		// case connectors.RESILIENT_ConnectorTypes:
		// 	return createConnectorRequestResilient(connector)
		// case connectors.SERVICENOW_ConnectorTypes:
		// 	return createConnectorRequestServicenow(connector)
		// case connectors.SERVICENOW_ITOM_ConnectorTypes:
		// 	return createConnectorRequestServicenowItom(connector)
		// case connectors.SERVICENOW_SIR:
		// 	return createConnectorRequestServicenowSir(connector)
		// case connectors.SERVER_LOG_ConnectorTypes:
		// 	return createConnectorRequestServerLog(connector)
		// case connectors.SLACK:
		// 	return createConnectorRequestSlack(connector)
		// case connectors.SWIMLANE_ConnectorTypes:
		// 	return createConnectorRequestSwimlane(connector)
		// case connectors.TEAMS:
		// 	return createConnectorRequestTeams(connector)
		// case connectors.TINES:
		// 	return createConnectorRequestTines(connector)
		// case connectors.WEBHOOK:
		// 	return createConnectorRequestWebhook(connector)
		// case connectors.XMATTERS:
		// 	return createConnectorRequestXmatters(connector)
	}

	return nil, fmt.Errorf("unknown connector type [%s]", connector.ConnectorTypeID)
}

func updateConnectorRequestBody(connector models.KibanaActionConnector) (io.Reader, error) {
	switch connectors.ConnectorTypes(connector.ConnectorTypeID) {
	// case connectors.CASES_WEBHOOK:
	// 	return updateConnectorRequestCasesWebhook(connector)
	// case connectors.EMAIL:
	// 	return updateConnectorRequestEmail(connector)
	case connectors.ConnectorTypesDotIndex:
		return updateConnectorRequestIndex(connector)
		// case connectors.JIRA:
		// 	return updateConnectorRequestJira(connector)
		// case connectors.OPSGENIE:
		// 	return updateConnectorRequestOpsgenie(connector)
		// case connectors.PAGERDUTY:
		// 	return updateConnectorRequestPagerduty(connector)
		// case connectors.RESILIENT:
		// 	return updateConnectorRequestResilient(connector)
		// case connectors.SERVICENOW:
		// 	return updateConnectorRequestServicenow(connector)
		// case connectors.SERVICENOW_ITOM:
		// 	return updateConnectorRequestServicenowItom(connector)
		// case connectors.SERVICENOW_SIR:
		// 	return updateConnectorRequestServicenowSir(connector)
		// case connectors.SERVER_LOG:
		// 	return updateConnectorRequestServerLog(connector)
		// case connectors.SLACK:
		// 	return updateConnectorRequestSlack(connector)
		// case connectors.SWIMLANE:
		// 	return updateConnectorRequestSwimlane(connector)
		// case connectors.TEAMS:
		// 	return updateConnectorRequestTeams(connector)
		// case connectors.TINES:
		// 	return updateConnectorRequestTines(connector)
		// case connectors.WEBHOOK:
		// 	return updateConnectorRequestWebhook(connector)
		// case connectors.XMATTERS:
		// 	return updateConnectorRequestXmatters(connector)
	}

	return nil, fmt.Errorf("unknown connector type [%s]", connector.ConnectorTypeID)
}

// func createConnectorRequestCasesWebhook(connector models.KibanaActionConnector) (connectors.CreateConnectorRequestBodyProperties, error) {
// 	prefixError := "failed to compose create connector request for CasesWebhook"

// 	config := connectors.NullableConfigPropertiesCasesWebhook{}
// 	if err := config.UnmarshalJSON([]byte(connector.ConfigJSON)); err != nil {
// 		return connectors.CreateConnectorRequestBodyProperties{}, fmt.Errorf("%s - failed to unmarshal [config] - %w", prefixError, err)
// 	}

// 	secrets := connectors.NullableSecretsPropertiesCasesWebhook{}
// 	if err := secrets.UnmarshalJSON([]byte(connector.SecretsJSON)); err != nil {
// 		return connectors.CreateConnectorRequestBodyProperties{}, fmt.Errorf("%s - failed to unmarshal [secrets]: %w", prefixError, err)
// 	}

// 	c := connectors.CreateConnectorRequestCasesWebhook{
// 		ConnectorTypeId: connector.ConnectorTypeID,
// 		Name:            connector.Name,
// 		Config:          *config.Get(),
// 		Secrets:         secrets.Get(),
// 	}

// 	return connectors.CreateConnectorRequestCasesWebhookAsCreateConnectorRequestBodyProperties(&c), nil
// }

// func createConnectorRequestEmail(connector models.KibanaActionConnector) (connectors.CreateConnectorRequestBodyProperties, error) {
// 	prefixError := "failed to compose create connector request for Email"

// 	var config map[string]interface{}
// 	if err := json.Unmarshal([]byte(connector.ConfigJSON), &config); err != nil {
// 		return connectors.CreateConnectorRequestBodyProperties{}, fmt.Errorf("%s - failed to unmarshal [config] - %w", prefixError, err)
// 	}

// 	var secrets map[string]interface{}
// 	if err := json.Unmarshal([]byte(connector.SecretsJSON), &secrets); err != nil {
// 		return connectors.CreateConnectorRequestBodyProperties{}, fmt.Errorf("%s - failed to unmarshal [secrets]: %w", prefixError, err)
// 	}

// 	c := connectors.CreateConnectorRequestEmail{
// 		ConnectorTypeId: connector.ConnectorTypeID,
// 		Name:            connector.Name,
// 		Config:          config,
// 		Secrets:         secrets,
// 	}

// 	return connectors.CreateConnectorRequestEmailAsCreateConnectorRequestBodyProperties(&c), nil
// }

func createConnectorRequestIndex(connector models.KibanaActionConnector) (io.Reader, error) {
	prefixError := "failed to create connector request for Index"

	request := connectors.CreateConnectorRequestIndex{
		ConnectorTypeId: connectors.CreateConnectorRequestIndexConnectorTypeIdDotIndex,
		Name:            connector.Name,
	}

	if err := json.Unmarshal([]byte(connector.ConfigJSON), &request.Config); err != nil {
		return nil, fmt.Errorf("%s: failed to unmarshal [CreateConnectorRequestIndex.Config]: %w", prefixError, err)
	}

	bt, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to marshal [CreateConnectorRequestIndex]: %w", prefixError, err)
	}

	return bytes.NewReader(bt), nil
}

// func createConnectorRequestJira(connector models.KibanaActionConnector) (connectors.CreateConnectorRequestBodyProperties, error) {
// 	prefixError := "failed to compose create connector request for Jira"

// 	config := connectors.NullableConfigPropertiesJira{}
// 	if err := config.UnmarshalJSON([]byte(connector.ConfigJSON)); err != nil {
// 		return connectors.CreateConnectorRequestBodyProperties{}, fmt.Errorf("%s - failed to unmarshal [config] - %w", prefixError, err)
// 	}

// 	secrets := connectors.NullableSecretsPropertiesJira{}
// 	if err := secrets.UnmarshalJSON([]byte(connector.SecretsJSON)); err != nil {
// 		return connectors.CreateConnectorRequestBodyProperties{}, fmt.Errorf("%s - failed to unmarshal [secrets]: %w", prefixError, err)
// 	}

// 	c := connectors.CreateConnectorRequestJira{
// 		ConnectorTypeId: connector.ConnectorTypeID,
// 		Name:            connector.Name,
// 		Config:          *config.Get(),
// 		Secrets:         *secrets.Get(),
// 	}

// 	return connectors.CreateConnectorRequestJiraAsCreateConnectorRequestBodyProperties(&c), nil
// }

// func createConnectorRequestOpsgenie(connector models.KibanaActionConnector) (connectors.CreateConnectorRequestBodyProperties, error) {
// 	prefixError := "failed to compose create connector request for Opsgenie"

// 	config := connectors.NullableConfigPropertiesOpsgenie{}
// 	if err := config.UnmarshalJSON([]byte(connector.ConfigJSON)); err != nil {
// 		return connectors.CreateConnectorRequestBodyProperties{}, fmt.Errorf("%s - failed to unmarshal [config] - %w", prefixError, err)
// 	}

// 	secrets := connectors.NullableSecretsPropertiesOpsgenie{}
// 	if err := secrets.UnmarshalJSON([]byte(connector.SecretsJSON)); err != nil {
// 		return connectors.CreateConnectorRequestBodyProperties{}, fmt.Errorf("%s - failed to unmarshal [secrets]: %w", prefixError, err)
// 	}

// 	c := connectors.CreateConnectorRequestOpsgenie{
// 		ConnectorTypeId: connector.ConnectorTypeID,
// 		Name:            connector.Name,
// 		Config:          *config.Get(),
// 		Secrets:         *secrets.Get(),
// 	}

// 	return connectors.CreateConnectorRequestOpsgenieAsCreateConnectorRequestBodyProperties(&c), nil
// }

// func createConnectorRequestPagerduty(connector models.KibanaActionConnector) (connectors.CreateConnectorRequestBodyProperties, error) {
// 	prefixError := "failed to compose create connector request for PagerDuty"

// 	var config map[string]interface{}
// 	if err := json.Unmarshal([]byte(connector.ConfigJSON), &config); err != nil {
// 		return connectors.CreateConnectorRequestBodyProperties{}, fmt.Errorf("%s - failed to unmarshal [config] - %w", prefixError, err)
// 	}

// 	var secrets map[string]interface{}
// 	if err := json.Unmarshal([]byte(connector.SecretsJSON), &secrets); err != nil {
// 		return connectors.CreateConnectorRequestBodyProperties{}, fmt.Errorf("%s - failed to unmarshal [secrets]: %w", prefixError, err)
// 	}

// 	c := connectors.CreateConnectorRequestPagerduty{
// 		ConnectorTypeId: connector.ConnectorTypeID,
// 		Name:            connector.Name,
// 		Config:          config,
// 		Secrets:         secrets,
// 	}

// 	return connectors.CreateConnectorRequestPagerdutyAsCreateConnectorRequestBodyProperties(&c), nil
// }

// func createConnectorRequestResilient(connector models.KibanaActionConnector) (connectors.CreateConnectorRequestBodyProperties, error) {
// 	prefixError := "failed to compose create connector request for Resilient"

// 	config := connectors.NullableConfigPropertiesResilient{}
// 	if err := config.UnmarshalJSON([]byte(connector.ConfigJSON)); err != nil {
// 		return connectors.CreateConnectorRequestBodyProperties{}, fmt.Errorf("%s - failed to unmarshal [config] - %w", prefixError, err)
// 	}

// 	secrets := connectors.NullableSecretsPropertiesResilient{}
// 	if err := secrets.UnmarshalJSON([]byte(connector.SecretsJSON)); err != nil {
// 		return connectors.CreateConnectorRequestBodyProperties{}, fmt.Errorf("%s - failed to unmarshal [secrets]: %w", prefixError, err)
// 	}

// 	c := connectors.CreateConnectorRequestResilient{
// 		ConnectorTypeId: connector.ConnectorTypeID,
// 		Name:            connector.Name,
// 		Config:          *config.Get(),
// 		Secrets:         *secrets.Get(),
// 	}

// 	return connectors.CreateConnectorRequestResilientAsCreateConnectorRequestBodyProperties(&c), nil
// }

// func createConnectorRequestServicenow(connector models.KibanaActionConnector) (connectors.CreateConnectorRequestBodyProperties, error) {
// 	prefixError := "failed to compose create connector request for Servicenow"

// 	config := connectors.NullableConfigPropertiesServicenow{}
// 	if err := config.UnmarshalJSON([]byte(connector.ConfigJSON)); err != nil {
// 		return connectors.CreateConnectorRequestBodyProperties{}, fmt.Errorf("%s - failed to unmarshal [config] - %w", prefixError, err)
// 	}

// 	secrets := connectors.NullableSecretsPropertiesServicenow{}
// 	if err := secrets.UnmarshalJSON([]byte(connector.SecretsJSON)); err != nil {
// 		return connectors.CreateConnectorRequestBodyProperties{}, fmt.Errorf("%s - failed to unmarshal [secrets]: %w", prefixError, err)
// 	}

// 	c := connectors.CreateConnectorRequestServicenow{
// 		ConnectorTypeId: connector.ConnectorTypeID,
// 		Name:            connector.Name,
// 		Config:          *config.Get(),
// 		Secrets:         *secrets.Get(),
// 	}

// 	return connectors.CreateConnectorRequestServicenowAsCreateConnectorRequestBodyProperties(&c), nil
// }

// func createConnectorRequestServicenowItom(connector models.KibanaActionConnector) (connectors.CreateConnectorRequestBodyProperties, error) {
// 	prefixError := "failed to compose create connector request for ServicenowItom"

// 	config := connectors.NullableConfigPropertiesServicenowItom{}
// 	if err := config.UnmarshalJSON([]byte(connector.ConfigJSON)); err != nil {
// 		return connectors.CreateConnectorRequestBodyProperties{}, fmt.Errorf("%s - failed to unmarshal [config] - %w", prefixError, err)
// 	}

// 	secrets := connectors.NullableSecretsPropertiesServicenow{}
// 	if err := secrets.UnmarshalJSON([]byte(connector.SecretsJSON)); err != nil {
// 		return connectors.CreateConnectorRequestBodyProperties{}, fmt.Errorf("%s - failed to unmarshal [secrets]: %w", prefixError, err)
// 	}

// 	c := connectors.CreateConnectorRequestServicenowItom{
// 		ConnectorTypeId: connector.ConnectorTypeID,
// 		Name:            connector.Name,
// 		Config:          *config.Get(),
// 		Secrets:         *secrets.Get(),
// 	}

// 	return connectors.CreateConnectorRequestServicenowItomAsCreateConnectorRequestBodyProperties(&c), nil
// }

// func createConnectorRequestServicenowSir(connector models.KibanaActionConnector) (connectors.CreateConnectorRequestBodyProperties, error) {
// 	prefixError := "failed to compose create connector request for ServicenowSir"

// 	config := connectors.NullableConfigPropertiesServicenow{}
// 	if err := config.UnmarshalJSON([]byte(connector.ConfigJSON)); err != nil {
// 		return connectors.CreateConnectorRequestBodyProperties{}, fmt.Errorf("%s - failed to unmarshal [config] - %w", prefixError, err)
// 	}

// 	secrets := connectors.NullableSecretsPropertiesServicenow{}
// 	if err := secrets.UnmarshalJSON([]byte(connector.SecretsJSON)); err != nil {
// 		return connectors.CreateConnectorRequestBodyProperties{}, fmt.Errorf("%s - failed to unmarshal [secrets]: %w", prefixError, err)
// 	}

// 	c := connectors.CreateConnectorRequestServicenowSir{
// 		ConnectorTypeId: connector.ConnectorTypeID,
// 		Name:            connector.Name,
// 		Config:          *config.Get(),
// 		Secrets:         *secrets.Get(),
// 	}

// 	return connectors.CreateConnectorRequestServicenowSirAsCreateConnectorRequestBodyProperties(&c), nil
// }

// func createConnectorRequestServerLog(connector models.KibanaActionConnector) (connectors.CreateConnectorRequestBodyProperties, error) {
// 	c := connectors.CreateConnectorRequestServerlog{
// 		ConnectorTypeId: connector.ConnectorTypeID,
// 		Name:            connector.Name,
// 	}

// 	return connectors.CreateConnectorRequestServerlogAsCreateConnectorRequestBodyProperties(&c), nil
// }

// func createConnectorRequestSlack(connector models.KibanaActionConnector) (connectors.CreateConnectorRequestBodyProperties, error) {
// 	prefixError := "failed to compose create connector request for Slack"

// 	var secrets map[string]interface{}
// 	if err := json.Unmarshal([]byte(connector.SecretsJSON), &secrets); err != nil {
// 		return connectors.CreateConnectorRequestBodyProperties{}, fmt.Errorf("%s - failed to unmarshal [secrets]: %w", prefixError, err)
// 	}

// 	c := connectors.CreateConnectorRequestSlack{
// 		ConnectorTypeId: connector.ConnectorTypeID,
// 		Name:            connector.Name,
// 		Secrets:         secrets,
// 	}

// 	return connectors.CreateConnectorRequestSlackAsCreateConnectorRequestBodyProperties(&c), nil
// }

// func createConnectorRequestSwimlane(connector models.KibanaActionConnector) (connectors.CreateConnectorRequestBodyProperties, error) {
// 	prefixError := "failed to compose create connector request for Swimlane"

// 	config := connectors.NullableConfigPropertiesSwimlane{}
// 	if err := config.UnmarshalJSON([]byte(connector.ConfigJSON)); err != nil {
// 		return connectors.CreateConnectorRequestBodyProperties{}, fmt.Errorf("%s - failed to unmarshal [config] - %w", prefixError, err)
// 	}

// 	secrets := connectors.NullableSecretsPropertiesSwimlane{}
// 	if err := secrets.UnmarshalJSON([]byte(connector.SecretsJSON)); err != nil {
// 		return connectors.CreateConnectorRequestBodyProperties{}, fmt.Errorf("%s - failed to unmarshal [secrets]: %w", prefixError, err)
// 	}

// 	c := connectors.CreateConnectorRequestSwimlane{
// 		ConnectorTypeId: connector.ConnectorTypeID,
// 		Name:            connector.Name,
// 		Config:          *config.Get(),
// 		Secrets:         *secrets.Get(),
// 	}

// 	return connectors.CreateConnectorRequestSwimlaneAsCreateConnectorRequestBodyProperties(&c), nil
// }

// func createConnectorRequestTeams(connector models.KibanaActionConnector) (connectors.CreateConnectorRequestBodyProperties, error) {
// 	prefixError := "failed to compose create connector request for Teams"

// 	var secrets map[string]interface{}
// 	if err := json.Unmarshal([]byte(connector.SecretsJSON), &secrets); err != nil {
// 		return connectors.CreateConnectorRequestBodyProperties{}, fmt.Errorf("%s - failed to unmarshal [secrets]: %w", prefixError, err)
// 	}

// 	c := connectors.CreateConnectorRequestTeams{
// 		ConnectorTypeId: connector.ConnectorTypeID,
// 		Name:            connector.Name,
// 		Secrets:         secrets,
// 	}

// 	return connectors.CreateConnectorRequestTeamsAsCreateConnectorRequestBodyProperties(&c), nil
// }

// func createConnectorRequestTines(connector models.KibanaActionConnector) (connectors.CreateConnectorRequestBodyProperties, error) {
// 	prefixError := "failed to compose create connector request for Tines"

// 	var config map[string]interface{}
// 	if err := json.Unmarshal([]byte(connector.ConfigJSON), &config); err != nil {
// 		return connectors.CreateConnectorRequestBodyProperties{}, fmt.Errorf("%s - failed to unmarshal [config] - %w", prefixError, err)
// 	}

// 	var secrets map[string]interface{}
// 	if err := json.Unmarshal([]byte(connector.SecretsJSON), &secrets); err != nil {
// 		return connectors.CreateConnectorRequestBodyProperties{}, fmt.Errorf("%s - failed to unmarshal [secrets]: %w", prefixError, err)
// 	}

// 	c := connectors.CreateConnectorRequestTines{
// 		ConnectorTypeId: connector.ConnectorTypeID,
// 		Name:            connector.Name,
// 		Config:          config,
// 		Secrets:         secrets,
// 	}

// 	return connectors.CreateConnectorRequestTinesAsCreateConnectorRequestBodyProperties(&c), nil
// }

// func createConnectorRequestWebhook(connector models.KibanaActionConnector) (connectors.CreateConnectorRequestBodyProperties, error) {
// 	prefixError := "failed to compose create connector request for Webhook"

// 	var config map[string]interface{}
// 	if err := json.Unmarshal([]byte(connector.ConfigJSON), &config); err != nil {
// 		return connectors.CreateConnectorRequestBodyProperties{}, fmt.Errorf("%s - failed to unmarshal [config] - %w", prefixError, err)
// 	}

// 	var secrets map[string]interface{}
// 	if err := json.Unmarshal([]byte(connector.SecretsJSON), &secrets); err != nil {
// 		return connectors.CreateConnectorRequestBodyProperties{}, fmt.Errorf("%s - failed to unmarshal [secrets]: %w", prefixError, err)
// 	}

// 	c := connectors.CreateConnectorRequestWebhook{
// 		ConnectorTypeId: connector.ConnectorTypeID,
// 		Name:            connector.Name,
// 		Config:          config,
// 		Secrets:         secrets,
// 	}

// 	return connectors.CreateConnectorRequestWebhookAsCreateConnectorRequestBodyProperties(&c), nil
// }

// func createConnectorRequestXmatters(connector models.KibanaActionConnector) (connectors.CreateConnectorRequestBodyProperties, error) {
// 	prefixError := "failed to compose create connector request for Xmatters"

// 	var config map[string]interface{}
// 	if err := json.Unmarshal([]byte(connector.ConfigJSON), &config); err != nil {
// 		return connectors.CreateConnectorRequestBodyProperties{}, fmt.Errorf("%s - failed to unmarshal [config] - %w", prefixError, err)
// 	}

// 	var secrets map[string]interface{}
// 	if err := json.Unmarshal([]byte(connector.SecretsJSON), &secrets); err != nil {
// 		return connectors.CreateConnectorRequestBodyProperties{}, fmt.Errorf("%s - failed to unmarshal [secrets]: %w", prefixError, err)
// 	}

// 	c := connectors.CreateConnectorRequestXmatters{
// 		ConnectorTypeId: connector.ConnectorTypeID,
// 		Name:            connector.Name,
// 		Config:          config,
// 		Secrets:         secrets,
// 	}

// 	return connectors.CreateConnectorRequestXmattersAsCreateConnectorRequestBodyProperties(&c), nil
// }

// func updateConnectorRequestCasesWebhook(connector models.KibanaActionConnector) (connectors.UpdateConnectorRequestBodyProperties, error) {
// 	prefixError := "failed to compose update connector request for CasesWebhook"

// 	nullableConfig := connectors.NullableConfigPropertiesCasesWebhook{}
// 	if err := nullableConfig.UnmarshalJSON([]byte(connector.ConfigJSON)); err != nil {
// 		return connectors.UpdateConnectorRequestBodyProperties{}, fmt.Errorf("%s - failed to unmarshal [config] - %w", prefixError, err)
// 	}

// 	nullableSecrets := connectors.NullableSecretsPropertiesCasesWebhook{}
// 	if err := nullableSecrets.UnmarshalJSON([]byte(connector.SecretsJSON)); err != nil {
// 		return connectors.UpdateConnectorRequestBodyProperties{}, fmt.Errorf("%s - failed to unmarshal [secrets]: %w", prefixError, err)
// 	}

// 	return connectors.UpdateConnectorRequestBodyProperties {
// 		 connectors.UpdateConnectorRequestCasesWebhook{
// 		Name:    connector.Name,
// 		Config:  *nullableConfig.Get(),
// 		Secrets: nullableSecrets.Get(),
// 	}

// 	return connectors.UpdateConnectorRequestCasesWebhookAsUpdateConnectorRequestBodyProperties(&c), nil
// }

func updateConnectorRequestIndex(connector models.KibanaActionConnector) (io.Reader, error) {
	prefixError := "failed to create connector request for Index"

	request := connectors.UpdateConnectorRequestIndex{
		Name: connector.Name,
	}

	if err := json.Unmarshal([]byte(connector.ConfigJSON), &request.Config); err != nil {
		return nil, fmt.Errorf("%s: failed to unmarshal [CreateConnectorRequestIndex.Config]: %w", prefixError, err)
	}

	bt, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to marshal [CreateConnectorRequestIndex]: %w", prefixError, err)
	}

	return bytes.NewReader(bt), nil
}

// func updateConnectorRequestJira(connector models.KibanaActionConnector) (connectors.UpdateConnectorRequestBodyProperties, error) {
// 	prefixError := "failed to compose update connector request for Jira"

// 	config := connectors.NullableConfigPropertiesJira{}
// 	if err := config.UnmarshalJSON([]byte(connector.ConfigJSON)); err != nil {
// 		return connectors.UpdateConnectorRequestBodyProperties{}, fmt.Errorf("%s - failed to unmarshal [config] - %w", prefixError, err)
// 	}

// 	secrets := connectors.NullableSecretsPropertiesJira{}
// 	if err := secrets.UnmarshalJSON([]byte(connector.SecretsJSON)); err != nil {
// 		return connectors.UpdateConnectorRequestBodyProperties{}, fmt.Errorf("%s - failed to unmarshal [secrets]: %w", prefixError, err)
// 	}

// 	c := connectors.UpdateConnectorRequestJira{
// 		Name:    connector.Name,
// 		Config:  *config.Get(),
// 		Secrets: *secrets.Get(),
// 	}

// 	return connectors.UpdateConnectorRequestJiraAsUpdateConnectorRequestBodyProperties(&c), nil
// }

// func updateConnectorRequestOpsgenie(connector models.KibanaActionConnector) (connectors.UpdateConnectorRequestBodyProperties, error) {
// 	prefixError := "failed to compose update connector request for Opsgenie"

// 	config := connectors.NullableConfigPropertiesOpsgenie{}
// 	if err := config.UnmarshalJSON([]byte(connector.ConfigJSON)); err != nil {
// 		return connectors.UpdateConnectorRequestBodyProperties{}, fmt.Errorf("%s - failed to unmarshal [config] - %w", prefixError, err)
// 	}

// 	secrets := connectors.NullableSecretsPropertiesOpsgenie{}
// 	if err := secrets.UnmarshalJSON([]byte(connector.SecretsJSON)); err != nil {
// 		return connectors.UpdateConnectorRequestBodyProperties{}, fmt.Errorf("%s - failed to unmarshal [secrets]: %w", prefixError, err)
// 	}

// 	c := connectors.UpdateConnectorRequestOpsgenie{
// 		Name:    connector.Name,
// 		Config:  *config.Get(),
// 		Secrets: *secrets.Get(),
// 	}

// 	return connectors.UpdateConnectorRequestOpsgenieAsUpdateConnectorRequestBodyProperties(&c), nil
// }

// func updateConnectorRequestResilient(connector models.KibanaActionConnector) (connectors.UpdateConnectorRequestBodyProperties, error) {
// 	prefixError := "failed to compose update connector request for Resilient"

// 	config := connectors.NullableConfigPropertiesResilient{}
// 	if err := config.UnmarshalJSON([]byte(connector.ConfigJSON)); err != nil {
// 		return connectors.UpdateConnectorRequestBodyProperties{}, fmt.Errorf("%s - failed to unmarshal [config] - %w", prefixError, err)
// 	}

// 	secrets := connectors.NullableSecretsPropertiesResilient{}
// 	if err := secrets.UnmarshalJSON([]byte(connector.SecretsJSON)); err != nil {
// 		return connectors.UpdateConnectorRequestBodyProperties{}, fmt.Errorf("%s - failed to unmarshal [secrets]: %w", prefixError, err)
// 	}

// 	c := connectors.UpdateConnectorRequestResilient{
// 		Name:    connector.Name,
// 		Config:  *config.Get(),
// 		Secrets: *secrets.Get(),
// 	}

// 	return connectors.UpdateConnectorRequestResilientAsUpdateConnectorRequestBodyProperties(&c), nil
// }

// func updateConnectorRequestServicenow(connector models.KibanaActionConnector) (connectors.UpdateConnectorRequestBodyProperties, error) {
// 	prefixError := "failed to compose update connector request for Servicenow"

// 	config := connectors.NullableConfigPropertiesServicenow{}
// 	if err := config.UnmarshalJSON([]byte(connector.ConfigJSON)); err != nil {
// 		return connectors.UpdateConnectorRequestBodyProperties{}, fmt.Errorf("%s - failed to unmarshal [config] - %w", prefixError, err)
// 	}

// 	secrets := connectors.NullableSecretsPropertiesServicenow{}
// 	if err := secrets.UnmarshalJSON([]byte(connector.SecretsJSON)); err != nil {
// 		return connectors.UpdateConnectorRequestBodyProperties{}, fmt.Errorf("%s - failed to unmarshal [secrets]: %w", prefixError, err)
// 	}

// 	c := connectors.UpdateConnectorRequestServicenow{
// 		Name:    connector.Name,
// 		Config:  *config.Get(),
// 		Secrets: *secrets.Get(),
// 	}

// 	return connectors.UpdateConnectorRequestServicenowAsUpdateConnectorRequestBodyProperties(&c), nil
// }

// func updateConnectorRequestServicenowItom(connector models.KibanaActionConnector) (connectors.UpdateConnectorRequestBodyProperties, error) {
// 	prefixError := "failed to compose update connector request for ServicenowItom"

// 	config := connectors.NullableConfigPropertiesServicenowItom{}
// 	if err := config.UnmarshalJSON([]byte(connector.ConfigJSON)); err != nil {
// 		return connectors.UpdateConnectorRequestBodyProperties{}, fmt.Errorf("%s - failed to unmarshal [config] - %w", prefixError, err)
// 	}

// 	secrets := connectors.NullableSecretsPropertiesServicenow{}
// 	if err := secrets.UnmarshalJSON([]byte(connector.SecretsJSON)); err != nil {
// 		return connectors.UpdateConnectorRequestBodyProperties{}, fmt.Errorf("%s - failed to unmarshal [secrets]: %w", prefixError, err)
// 	}

// 	c := connectors.UpdateConnectorRequestServicenowItom{
// 		Name:    connector.Name,
// 		Config:  *config.Get(),
// 		Secrets: *secrets.Get(),
// 	}

// 	return connectors.UpdateConnectorRequestServicenowItomAsUpdateConnectorRequestBodyProperties(&c), nil
// }

// func updateConnectorRequestServerLog(connector models.KibanaActionConnector) (connectors.UpdateConnectorRequestBodyProperties, error) {
// 	c := connectors.UpdateConnectorRequestServerlog{
// 		Name: connector.Name,
// 	}

// 	return connectors.UpdateConnectorRequestServerlogAsUpdateConnectorRequestBodyProperties(&c), nil
// }

// func updateConnectorRequestSwimlane(connector models.KibanaActionConnector) (connectors.UpdateConnectorRequestBodyProperties, error) {
// 	prefixError := "failed to compose update connector request for Swimlane"

// 	config := connectors.NullableConfigPropertiesSwimlane{}
// 	if err := config.UnmarshalJSON([]byte(connector.ConfigJSON)); err != nil {
// 		return connectors.UpdateConnectorRequestBodyProperties{}, fmt.Errorf("%s - failed to unmarshal [config] - %w", prefixError, err)
// 	}

// 	secrets := connectors.NullableSecretsPropertiesSwimlane{}
// 	if err := secrets.UnmarshalJSON([]byte(connector.SecretsJSON)); err != nil {
// 		return connectors.UpdateConnectorRequestBodyProperties{}, fmt.Errorf("%s - failed to unmarshal [secrets]: %w", prefixError, err)
// 	}

// 	c := connectors.UpdateConnectorRequestSwimlane{
// 		Name:    connector.Name,
// 		Config:  *config.Get(),
// 		Secrets: *secrets.Get(),
// 	}

// 	return connectors.UpdateConnectorRequestSwimlaneAsUpdateConnectorRequestBodyProperties(&c), nil
// }

func connectorResponseToModel(spaceID string, properties connectors.ConnectorResponseProperties) (*models.KibanaActionConnector, error) {
	discriminator, err := properties.Discriminator()
	if err != nil {
		return nil, err
	}
	switch discriminator {

	// case connectors.CASES_WEBHOOK_ConnectorTypes:
	// 	config, err := response.GetConfig().MarshalJSON()
	// 	if err != nil {
	// 		return nil, fmt.Errorf("failed to parse [config] in ConnectorResponsePropertiesCasesWebhook - [%w]", err)
	// 	}
	// 	// return responseToConnector(response, config, spaceID), nil

	// 	connector := models.KibanaActionConnector{
	// 		ConnectorID:      response.GetId(),
	// 		SpaceID:          spaceID,
	// 		Name:             response.GetName(),
	// 		ConnectorTypeID:  response.GetConnectorTypeId(),
	// 		IsDeprecated:     response.GetIsDeprecated(),
	// 		IsMissingSecrets: response.GetIsMissingSecrets(),
	// 		IsPreconfigured:  response.GetIsPreconfigured(),
	// 		ConfigJSON:       string(config),
	// 	}
	// 	return &connector, nil

	// case *connectors.ConnectorResponsePropertiesEmail:
	// 	config, err := json.Marshal(response.GetConfig())
	// 	if err != nil {
	// 		return nil, fmt.Errorf("failed to parse [config] in ConnectorResponsePropertiesEmail - [%w]", err)
	// 	}
	// 	// return responseToConnector(response, config, spaceID), nil
	// 	connector := models.KibanaActionConnector{
	// 		ConnectorID:      response.GetId(),
	// 		SpaceID:          spaceID,
	// 		Name:             response.GetName(),
	// 		ConnectorTypeID:  response.GetConnectorTypeId(),
	// 		IsDeprecated:     response.GetIsDeprecated(),
	// 		IsMissingSecrets: response.GetIsMissingSecrets(),
	// 		IsPreconfigured:  response.GetIsPreconfigured(),
	// 		ConfigJSON:       string(config),
	// 	}
	// 	return &connector, nil

	case string(connectors.ConnectorTypesDotIndex):
		resp, err := properties.AsConnectorResponsePropertiesIndex()
		if err != nil {
			return nil, err
		}

		config, err := json.Marshal(resp.Config)
		if err != nil {
			return nil, fmt.Errorf("unable to marshal config: %w", err)
		}

		isDeprecated := false
		isMissingSecrets := false

		if resp.IsDeprecated != nil {
			isDeprecated = *resp.IsDeprecated
		}

		if resp.IsMissingSecrets != nil {
			isMissingSecrets = *resp.IsMissingSecrets
		}

		connector := models.KibanaActionConnector{
			ConnectorID:      resp.Id,
			SpaceID:          spaceID,
			Name:             resp.Name,
			ConnectorTypeID:  string(connectors.ConnectorTypesDotIndex),
			IsDeprecated:     isDeprecated,
			IsMissingSecrets: isMissingSecrets,
			IsPreconfigured:  bool(resp.IsPreconfigured),
			ConfigJSON:       string(config),
		}

		return &connector, nil

		// case *connectors.ConnectorResponsePropertiesJira:
		// 	config, err := response.GetConfig().MarshalJSON()
		// 	if err != nil {
		// 		return nil, fmt.Errorf("failed to parse [config] in ConnectorResponsePropertiesCasesWebhook - [%w]", err)
		// 	}
		// 	// return responseToConnector(response, config, spaceID), nil
		// 	connector := models.KibanaActionConnector{
		// 		ConnectorID:      response.GetId(),
		// 		SpaceID:          spaceID,
		// 		Name:             response.GetName(),
		// 		ConnectorTypeID:  response.GetConnectorTypeId(),
		// 		IsDeprecated:     response.GetIsDeprecated(),
		// 		IsMissingSecrets: response.GetIsMissingSecrets(),
		// 		IsPreconfigured:  response.GetIsPreconfigured(),
		// 		ConfigJSON:       string(config),
		// 	}
		// 	return &connector, nil

		// case *connectors.ConnectorResponsePropertiesOpsgenie:
		// 	config, err := response.GetConfig().MarshalJSON()
		// 	if err != nil {
		// 		return nil, fmt.Errorf("failed to parse [config] in ConnectorResponsePropertiesCasesWebhook - [%w]", err)
		// 	}
		// 	// return responseToConnector(response, config, spaceID), nil
		// 	connector := models.KibanaActionConnector{
		// 		ConnectorID:      response.GetId(),
		// 		SpaceID:          spaceID,
		// 		Name:             response.GetName(),
		// 		ConnectorTypeID:  response.GetConnectorTypeId(),
		// 		IsDeprecated:     response.GetIsDeprecated(),
		// 		IsMissingSecrets: response.GetIsMissingSecrets(),
		// 		IsPreconfigured:  response.GetIsPreconfigured(),
		// 		ConfigJSON:       string(config),
		// 	}
		// 	return &connector, nil

		// case *connectors.ConnectorResponsePropertiesPagerduty:
		// 	config, err := json.Marshal(response.GetConfig())
		// 	if err != nil {
		// 		return nil, fmt.Errorf("failed to parse [config] in ConnectorResponsePropertiesCasesWebhook - [%w]", err)
		// 	}
		// 	// return responseToConnector(response, config, spaceID), nil
		// 	connector := models.KibanaActionConnector{
		// 		ConnectorID:      response.GetId(),
		// 		SpaceID:          spaceID,
		// 		Name:             response.GetName(),
		// 		ConnectorTypeID:  response.GetConnectorTypeId(),
		// 		IsDeprecated:     response.GetIsDeprecated(),
		// 		IsMissingSecrets: response.GetIsMissingSecrets(),
		// 		IsPreconfigured:  response.GetIsPreconfigured(),
		// 		ConfigJSON:       string(config),
		// 	}
		// 	return &connector, nil

		// case *connectors.ConnectorResponsePropertiesResilient:
		// 	config, err := response.GetConfig().MarshalJSON()
		// 	if err != nil {
		// 		return nil, fmt.Errorf("failed to parse [config] in ConnectorResponsePropertiesCasesWebhook - [%w]", err)
		// 	}
		// 	// return responseToConnector(response, config, spaceID), nil
		// 	connector := models.KibanaActionConnector{
		// 		ConnectorID:      response.GetId(),
		// 		SpaceID:          spaceID,
		// 		Name:             response.GetName(),
		// 		ConnectorTypeID:  response.GetConnectorTypeId(),
		// 		IsDeprecated:     response.GetIsDeprecated(),
		// 		IsMissingSecrets: response.GetIsMissingSecrets(),
		// 		IsPreconfigured:  response.GetIsPreconfigured(),
		// 		ConfigJSON:       string(config),
		// 	}
		// 	return &connector, nil

		// case *connectors.ConnectorResponsePropertiesServerlog:
		// 	config, err := json.Marshal(response.GetConfig())
		// 	if err != nil {
		// 		return nil, fmt.Errorf("failed to parse [config] in ConnectorResponsePropertiesCasesWebhook - [%w]", err)
		// 	}
		// 	// return responseToConnector(response, config, spaceID), nil
		// 	connector := models.KibanaActionConnector{
		// 		ConnectorID:      response.GetId(),
		// 		SpaceID:          spaceID,
		// 		Name:             response.GetName(),
		// 		ConnectorTypeID:  response.GetConnectorTypeId(),
		// 		IsDeprecated:     response.GetIsDeprecated(),
		// 		IsMissingSecrets: response.GetIsMissingSecrets(),
		// 		IsPreconfigured:  response.GetIsPreconfigured(),
		// 		ConfigJSON:       string(config),
		// 	}
		// 	return &connector, nil

		// case *connectors.ConnectorResponsePropertiesServicenow:
		// 	config, err := response.GetConfig().MarshalJSON()
		// 	if err != nil {
		// 		return nil, fmt.Errorf("failed to parse [config] in ConnectorResponsePropertiesCasesWebhook - [%w]", err)
		// 	}
		// 	// return responseToConnector(response, config, spaceID), nil
		// 	connector := models.KibanaActionConnector{
		// 		ConnectorID:      response.GetId(),
		// 		SpaceID:          spaceID,
		// 		Name:             response.GetName(),
		// 		ConnectorTypeID:  response.GetConnectorTypeId(),
		// 		IsDeprecated:     response.GetIsDeprecated(),
		// 		IsMissingSecrets: response.GetIsMissingSecrets(),
		// 		IsPreconfigured:  response.GetIsPreconfigured(),
		// 		ConfigJSON:       string(config),
		// 	}
		// 	return &connector, nil

		// case *connectors.ConnectorResponsePropertiesServicenowItom:
		// 	config, err := response.GetConfig().MarshalJSON()
		// 	if err != nil {
		// 		return nil, fmt.Errorf("failed to parse [config] in ConnectorResponsePropertiesCasesWebhook - [%w]", err)
		// 	}
		// 	// return responseToConnector(response, config, spaceID), nil
		// 	connector := models.KibanaActionConnector{
		// 		ConnectorID:      response.GetId(),
		// 		SpaceID:          spaceID,
		// 		Name:             response.GetName(),
		// 		ConnectorTypeID:  response.GetConnectorTypeId(),
		// 		IsDeprecated:     response.GetIsDeprecated(),
		// 		IsMissingSecrets: response.GetIsMissingSecrets(),
		// 		IsPreconfigured:  response.GetIsPreconfigured(),
		// 		ConfigJSON:       string(config),
		// 	}
		// 	return &connector, nil

		// case *connectors.ConnectorResponsePropertiesServicenowSir:
		// 	config, err := response.GetConfig().MarshalJSON()
		// 	if err != nil {
		// 		return nil, fmt.Errorf("failed to parse [config] in ConnectorResponsePropertiesCasesWebhook - [%w]", err)
		// 	}
		// 	// return responseToConnector(response, config, spaceID), nil
		// 	connector := models.KibanaActionConnector{
		// 		ConnectorID:      response.GetId(),
		// 		SpaceID:          spaceID,
		// 		Name:             response.GetName(),
		// 		ConnectorTypeID:  response.GetConnectorTypeId(),
		// 		IsDeprecated:     response.GetIsDeprecated(),
		// 		IsMissingSecrets: response.GetIsMissingSecrets(),
		// 		IsPreconfigured:  response.GetIsPreconfigured(),
		// 		ConfigJSON:       string(config),
		// 	}
		// 	return &connector, nil

		// case *connectors.ConnectorResponsePropertiesSlack:
		// 	// return responseToConnector(response, config, spaceID), nil
		// 	connector := models.KibanaActionConnector{
		// 		ConnectorID:      response.GetId(),
		// 		SpaceID:          spaceID,
		// 		Name:             response.GetName(),
		// 		ConnectorTypeID:  response.GetConnectorTypeId(),
		// 		IsDeprecated:     response.GetIsDeprecated(),
		// 		IsMissingSecrets: response.GetIsMissingSecrets(),
		// 		IsPreconfigured:  response.GetIsPreconfigured(),
		// 	}
		// 	return &connector, nil

		// case *connectors.ConnectorResponsePropertiesSwimlane:
		// 	config, err := response.GetConfig().MarshalJSON()
		// 	if err != nil {
		// 		return nil, fmt.Errorf("failed to parse [config] in ConnectorResponsePropertiesCasesWebhook - [%w]", err)
		// 	}
		// 	// return responseToConnector(response, config, spaceID), nil
		// 	connector := models.KibanaActionConnector{
		// 		ConnectorID:      response.GetId(),
		// 		SpaceID:          spaceID,
		// 		Name:             response.GetName(),
		// 		ConnectorTypeID:  response.GetConnectorTypeId(),
		// 		IsDeprecated:     response.GetIsDeprecated(),
		// 		IsMissingSecrets: response.GetIsMissingSecrets(),
		// 		IsPreconfigured:  response.GetIsPreconfigured(),
		// 		ConfigJSON:       string(config),
		// 	}
		// 	return &connector, nil

		// case *connectors.ConnectorResponsePropertiesTeams:
		// 	// return responseToConnector(response, config, spaceID), nil
		// 	connector := models.KibanaActionConnector{
		// 		ConnectorID:      response.GetId(),
		// 		SpaceID:          spaceID,
		// 		Name:             response.GetName(),
		// 		ConnectorTypeID:  response.GetConnectorTypeId(),
		// 		IsDeprecated:     response.GetIsDeprecated(),
		// 		IsMissingSecrets: response.GetIsMissingSecrets(),
		// 		IsPreconfigured:  response.GetIsPreconfigured(),
		// 	}
		// 	return &connector, nil

		// case *connectors.ConnectorResponsePropertiesTines:
		// 	config, err := json.Marshal(response.GetConfig())
		// 	if err != nil {
		// 		return nil, fmt.Errorf("failed to parse [config] in ConnectorResponsePropertiesCasesWebhook - [%w]", err)
		// 	}
		// 	// return responseToConnector(response, config, spaceID), nil
		// 	connector := models.KibanaActionConnector{
		// 		ConnectorID:      response.GetId(),
		// 		SpaceID:          spaceID,
		// 		Name:             response.GetName(),
		// 		ConnectorTypeID:  response.GetConnectorTypeId(),
		// 		IsDeprecated:     response.GetIsDeprecated(),
		// 		IsMissingSecrets: response.GetIsMissingSecrets(),
		// 		IsPreconfigured:  response.GetIsPreconfigured(),
		// 		ConfigJSON:       string(config),
		// 	}
		// 	return &connector, nil

		// case *connectors.ConnectorResponsePropertiesWebhook:
		// 	config, err := json.Marshal(response.GetConfig())
		// 	if err != nil {
		// 		return nil, fmt.Errorf("failed to parse [config] in ConnectorResponsePropertiesCasesWebhook - [%w]", err)
		// 	}
		// 	// return responseToConnector(response, config, spaceID), nil
		// 	connector := models.KibanaActionConnector{
		// 		ConnectorID:      response.GetId(),
		// 		SpaceID:          spaceID,
		// 		Name:             response.GetName(),
		// 		ConnectorTypeID:  response.GetConnectorTypeId(),
		// 		IsDeprecated:     response.GetIsDeprecated(),
		// 		IsMissingSecrets: response.GetIsMissingSecrets(),
		// 		IsPreconfigured:  response.GetIsPreconfigured(),
		// 		ConfigJSON:       string(config),
		// 	}
		// 	return &connector, nil

		// case *connectors.ConnectorResponsePropertiesXmatters:
		// 	config, err := json.Marshal(response.GetConfig())
		// 	if err != nil {
		// 		return nil, fmt.Errorf("failed to parse [config] in ConnectorResponsePropertiesCasesWebhook - [%w]", err)
		// 	}
		// 	// return responseToConnector(response, config, spaceID), nil
		// 	connector := models.KibanaActionConnector{
		// 		ConnectorID:      response.GetId(),
		// 		SpaceID:          spaceID,
		// 		Name:             response.GetName(),
		// 		ConnectorTypeID:  response.GetConnectorTypeId(),
		// 		IsDeprecated:     response.GetIsDeprecated(),
		// 		IsMissingSecrets: response.GetIsMissingSecrets(),
		// 		IsPreconfigured:  response.GetIsPreconfigured(),
		// 		ConfigJSON:       string(config),
		// 	}
		// 	return &connector, nil
	}

	return nil, fmt.Errorf("unknown connector type [%+v]", properties)
}

// func responseToConnector[T responseType](response T, config []byte, spaceID string) *models.KibanaActionConnector {
// 	return &models.KibanaActionConnector{
// 		ConnectorID:      response.GetId(),
// 		SpaceID:          spaceID,
// 		Name:             response.GetName(),
// 		ConnectorTypeID:  response.GetConnectorTypeId(),
// 		IsDeprecated:     response.GetIsDeprecated(),
// 		IsMissingSecrets: response.GetIsMissingSecrets(),
// 		IsPreconfigured:  response.GetIsPreconfigured(),
// 		ConfigJSON:       string(config),
// 	}
// }

// type responseType interface {
// 	*connectors.ConnectorResponsePropertiesCasesWebhook | *connectors.ConnectorResponsePropertiesEmail |
// 		*connectors.ConnectorResponsePropertiesIndex
// 	GetId() string
// 	GetName() string
// 	GetConnectorTypeId() string
// 	GetIsDeprecated() bool
// 	GetIsMissingSecrets() bool
// 	GetIsPreconfigured() bool
// }
