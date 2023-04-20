package kibana

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/generated/connectors"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
)

func CreateActionConnector(ctx context.Context, apiClient *clients.ApiClient, connectorOld models.KibanaActionConnector) (string, diag.Diagnostics) {
	client, ctxWithAuth, err := apiClient.GetKibanaActionConnectorClient(ctx)
	if err != nil {
		return "", diag.FromErr(err)
	}

	createProperties, err := createConnectorRequestBodyProperties(connectorOld)
	if err != nil {
		return "", diag.FromErr(err)
	}

	response, err := client.CreateConnector(ctxWithAuth, createProperties, connectors.CreateConnectorParams{KbnXSRF: "true", SpaceId: connectorOld.SpaceID})
	if err != nil {
		return "", diag.FromErr(fmt.Errorf("create connector failed: [%w]", err))
	}

	properties, ok := response.(*connectors.ConnectorResponseProperties)
	if !ok {
		return "", diag.FromErr(fmt.Errorf("failed to parse create response [%+v]", response))
	}

	connectorNew, err := actionConnectorToModel(connectorOld.SpaceID, *properties)
	if err != nil {
		return "", diag.FromErr(err)
	}

	return connectorNew.ConnectorID, nil
}

func UpdateActionConnector(ctx context.Context, apiClient *clients.ApiClient, connectorOld models.KibanaActionConnector) (string, diag.Diagnostics) {
	client, ctxWithAuth, err := apiClient.GetKibanaActionConnectorClient(ctx)
	if err != nil {
		return "", diag.FromErr(err)
	}

	updateProperties, err := updateConnectorRequestBodyProperties(connectorOld)
	if err != nil {
		return "", diag.FromErr(err)
	}

	response, err := client.UpdateConnector(
		ctxWithAuth,
		updateProperties,
		connectors.UpdateConnectorParams{
			KbnXSRF:     "true",
			ConnectorId: connectorOld.ConnectorID,
			SpaceId:     connectorOld.SpaceID,
		},
	)

	if err != nil {
		return "", diag.FromErr(err)
	}

	var properties *connectors.ConnectorResponseProperties

	switch resp := response.(type) {
	case *connectors.ConnectorResponseProperties:
		properties, _ = response.(*connectors.ConnectorResponseProperties)
	case *connectors.R400:
		return "", diag.Errorf("update failed with error [%s]: %s", resp.GetError().Value, resp.GetMessage().Value)
	case *connectors.R401:
		return "", diag.Errorf("update failed with error [%s]: %s", resp.GetError().Value, resp.GetMessage().Value)
	case *connectors.R404:
		return "", diag.Errorf("update failed with error [%s]: %s", resp.GetError().Value, resp.GetMessage().Value)
	default:
		return "", diag.Errorf("failed to parse update response %+v", response)
	}

	connectorNew, err := actionConnectorToModel(connectorOld.SpaceID, *properties)
	if err != nil {
		return "", diag.FromErr(err)
	}

	return connectorNew.ConnectorID, nil
}

func GetActionConnector(ctx context.Context, apiClient *clients.ApiClient, connectorID, spaceID string, connectorTypeID string) (*models.KibanaActionConnector, diag.Diagnostics) {
	client, ctxWithAuth, err := apiClient.GetKibanaActionConnectorClient(ctx)
	if err != nil {
		return nil, diag.FromErr(err)
	}

	response, err := client.GetConnector(ctxWithAuth, connectors.GetConnectorParams{ConnectorId: connectorID, SpaceId: spaceID})

	if err != nil {
		return nil, diag.FromErr(err)
	}

	properties, ok := response.(*connectors.ConnectorResponseProperties)
	if !ok {
		return nil, diag.FromErr(fmt.Errorf("failed to parse get response [%+v]", response))
	}

	connector, err := actionConnectorToModel(spaceID, *properties)
	if err != nil {
		return nil, diag.FromErr(err)
	}

	return connector, nil
}

func DeleteActionConnector(ctx context.Context, apiClient *clients.ApiClient, connectorID string, spaceID string) diag.Diagnostics {
	client, ctxWithAuth, err := apiClient.GetKibanaActionConnectorClient(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = client.DeleteConnector(ctxWithAuth, connectors.DeleteConnectorParams{KbnXSRF: "true", ConnectorId: connectorID, SpaceId: spaceID})

	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func createConnectorRequestBodyProperties(connector models.KibanaActionConnector) (connectors.CreateConnectorReq, error) {
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

	return connectors.CreateConnectorReq{}, fmt.Errorf("unknown connector type [%s]", connector.ConnectorTypeID)
}

func updateConnectorRequestBodyProperties(connector models.KibanaActionConnector) (connectors.UpdateConnectorReq, error) {
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

	return connectors.UpdateConnectorReq{}, fmt.Errorf("unknown connector type [%s]", connector.ConnectorTypeID)
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

func createConnectorRequestIndex(connector models.KibanaActionConnector) (connectors.CreateConnectorReq, error) {
	prefixError := "failed to compose create connector request for Index"

	config := &connectors.ConfigPropertiesIndex{}

	if err := config.UnmarshalJSON([]byte(connector.ConfigJSON)); err != nil {
		return connectors.CreateConnectorReq{}, fmt.Errorf("%s: failed to unmarshal [config]: %w", prefixError, err)
	}

	res := connectors.CreateConnectorReq{}

	res.SetCreateConnectorRequestIndex(
		connectors.CreateConnectorRequestIndex{
			Name:   connector.Name,
			Config: *config,
		},
	)

	return res, nil
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

func updateConnectorRequestIndex(connector models.KibanaActionConnector) (connectors.UpdateConnectorReq, error) {
	prefixError := "failed to compose update connector request for Index"

	config := &connectors.ConfigPropertiesIndex{}

	if err := config.UnmarshalJSON([]byte(connector.ConfigJSON)); err != nil {
		return connectors.UpdateConnectorReq{}, fmt.Errorf("%s: failed to unmarshal [config]: %w", prefixError, err)
	}

	res := connectors.UpdateConnectorReq{}

	res.SetUpdateConnectorRequestIndex(
		connectors.UpdateConnectorRequestIndex{
			Name:   connector.Name,
			Config: *config,
		},
	)

	return res, nil
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

func actionConnectorToModel(spaceID string, properties connectors.ConnectorResponseProperties) (*models.KibanaActionConnector, error) {
	switch properties.Type {

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

	case connectors.ConnectorResponsePropertiesIndexConnectorResponseProperties:
		resp, _ := properties.GetConnectorResponsePropertiesIndex()

		config, err := resp.Config.MarshalJSON()
		if err != nil {
			return nil, fmt.Errorf("failed to parse [config] in ConnectorResponsePropertiesCasesWebhook - [%w]", err)
		}

		connector := models.KibanaActionConnector{
			ConnectorID:      resp.ID,
			SpaceID:          spaceID,
			Name:             resp.Name,
			ConnectorTypeID:  string(connectors.ConnectorTypesDotIndex),
			IsDeprecated:     bool(resp.IsDeprecated.Or(connectors.IsDeprecated(false))),
			IsMissingSecrets: bool(resp.IsMissingSecrets.Or(connectors.IsMissingSecrets(false))),
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
