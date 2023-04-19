package kibana

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/elastic/terraform-provider-elasticstack/generated/kibanaactions"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
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

	responseProperties, httpRes, err := client.CreateConnector(ctxWithAuth, createProperties, "true", connectorOld.SpaceID)
	if err != nil {
		var swagErr kibanaactions.GenericSwaggerError
		if errors.As(err, &swagErr) {
			return "", diag.FromErr(fmt.Errorf("%s", string(swagErr.Body())))
		}
		return "", diag.FromErr(err)
	}
	defer httpRes.Body.Close()

	// if diags := utils.CheckHttpError(httpRes, "Unabled to create action connector"); diags.HasError() {
	// 	return "", diag.FromErr(err)
	// }

	// if err != nil {
	// 	return "", diag.FromErr(err)
	// }

	connectorNew, err := actionConnectorToModel(connectorOld.SpaceID, responseProperties, connectorOld.ConnectorID)
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

	responseProperties, httpRes, err := client.UpdateConnector(ctxWithAuth, updateProperties, "true", connectorOld.ConnectorID, connectorOld.SpaceID)
	if err != nil && httpRes == nil {
		return "", diag.FromErr(err)
	}
	defer httpRes.Body.Close()

	if diags := utils.CheckHttpError(httpRes, "Unabled to update action connector"); diags.HasError() {
		return "", diags
	}

	connectorNew, err := actionConnectorToModel(connectorOld.SpaceID, responseProperties, connectorOld.ConnectorTypeID)
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

	properties, res, err := client.GetConnector(ctxWithAuth, connectorID, spaceID)
	if err != nil && res == nil {
		return nil, diag.FromErr(err)
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	if diags := utils.CheckHttpError(res, "Unabled to get action connector"); diags.HasError() {
		return nil, diags
	}

	connector, err := actionConnectorToModel(spaceID, properties, connectorTypeID)
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

	res, err := client.DeleteConnector(ctxWithAuth, "true", connectorID, spaceID)
	if err != nil && res == nil {
		return diag.FromErr(err)
	}
	defer res.Body.Close()

	return utils.CheckHttpError(res, "Unabled to delete action connector")
}

func createConnectorRequestBodyProperties(connector models.KibanaActionConnector) (kibanaactions.CreateConnectorRequestBodyProperties, error) {
	switch kibanaactions.ConnectorTypes(connector.ConnectorTypeID) {
	// case kibanaactions.CASES_WEBHOOK_ConnectorTypes:
	// 	return createConnectorRequestCasesWebhook(connector)
	// case kibanaactions.EMAIL_ConnectorTypes:
	//  	return createConnectorRequestEmail(connector)
	case kibanaactions.INDEX_ConnectorTypes:
		return createConnectorRequestIndex(connector)
		// case kibanaactions.JIRA_ConnectorTypes:
		// 	return createConnectorRequestJira(connector)
		// case kibanaactions.OPSGENIE_ConnectorTypes:
		// 	return createConnectorRequestOpsgenie(connector)
		// case kibanaactions.PAGERDUTY:
		// 	return createConnectorRequestPagerduty(connector)
		// case kibanaactions.RESILIENT_ConnectorTypes:
		// 	return createConnectorRequestResilient(connector)
		// case kibanaactions.SERVICENOW_ConnectorTypes:
		// 	return createConnectorRequestServicenow(connector)
		// case kibanaactions.SERVICENOW_ITOM_ConnectorTypes:
		// 	return createConnectorRequestServicenowItom(connector)
		// case kibanaactions.SERVICENOW_SIR:
		// 	return createConnectorRequestServicenowSir(connector)
		// case kibanaactions.SERVER_LOG_ConnectorTypes:
		// 	return createConnectorRequestServerLog(connector)
		// case kibanaactions.SLACK:
		// 	return createConnectorRequestSlack(connector)
		// case kibanaactions.SWIMLANE_ConnectorTypes:
		// 	return createConnectorRequestSwimlane(connector)
		// case kibanaactions.TEAMS:
		// 	return createConnectorRequestTeams(connector)
		// case kibanaactions.TINES:
		// 	return createConnectorRequestTines(connector)
		// case kibanaactions.WEBHOOK:
		// 	return createConnectorRequestWebhook(connector)
		// case kibanaactions.XMATTERS:
		// 	return createConnectorRequestXmatters(connector)
	}

	return kibanaactions.CreateConnectorRequestBodyProperties{}, fmt.Errorf("unknown connector type [%s]", connector.ConnectorTypeID)
}

func updateConnectorRequestBodyProperties(connector models.KibanaActionConnector) (kibanaactions.UpdateConnectorRequestBodyProperties, error) {
	switch kibanaactions.ConnectorTypes(connector.ConnectorTypeID) {
	// case kibanaactions.CASES_WEBHOOK:
	// 	return updateConnectorRequestCasesWebhook(connector)
	// case kibanaactions.EMAIL:
	// 	return updateConnectorRequestEmail(connector)
	case kibanaactions.INDEX_ConnectorTypes:
		return updateConnectorRequestIndex(connector)
		// case kibanaactions.JIRA:
		// 	return updateConnectorRequestJira(connector)
		// case kibanaactions.OPSGENIE:
		// 	return updateConnectorRequestOpsgenie(connector)
		// case kibanaactions.PAGERDUTY:
		// 	return updateConnectorRequestPagerduty(connector)
		// case kibanaactions.RESILIENT:
		// 	return updateConnectorRequestResilient(connector)
		// case kibanaactions.SERVICENOW:
		// 	return updateConnectorRequestServicenow(connector)
		// case kibanaactions.SERVICENOW_ITOM:
		// 	return updateConnectorRequestServicenowItom(connector)
		// case kibanaactions.SERVICENOW_SIR:
		// 	return updateConnectorRequestServicenowSir(connector)
		// case kibanaactions.SERVER_LOG:
		// 	return updateConnectorRequestServerLog(connector)
		// case kibanaactions.SLACK:
		// 	return updateConnectorRequestSlack(connector)
		// case kibanaactions.SWIMLANE:
		// 	return updateConnectorRequestSwimlane(connector)
		// case kibanaactions.TEAMS:
		// 	return updateConnectorRequestTeams(connector)
		// case kibanaactions.TINES:
		// 	return updateConnectorRequestTines(connector)
		// case kibanaactions.WEBHOOK:
		// 	return updateConnectorRequestWebhook(connector)
		// case kibanaactions.XMATTERS:
		// 	return updateConnectorRequestXmatters(connector)
	}

	return kibanaactions.UpdateConnectorRequestBodyProperties{}, fmt.Errorf("unknown connector type [%s]", connector.ConnectorTypeID)
}

// func createConnectorRequestCasesWebhook(connector models.KibanaActionConnector) (kibanaactions.CreateConnectorRequestBodyProperties, error) {
// 	prefixError := "failed to compose create connector request for CasesWebhook"

// 	config := kibanaactions.NullableConfigPropertiesCasesWebhook{}
// 	if err := config.UnmarshalJSON([]byte(connector.ConfigJSON)); err != nil {
// 		return kibanaactions.CreateConnectorRequestBodyProperties{}, fmt.Errorf("%s - failed to unmarshal [config] - %w", prefixError, err)
// 	}

// 	secrets := kibanaactions.NullableSecretsPropertiesCasesWebhook{}
// 	if err := secrets.UnmarshalJSON([]byte(connector.SecretsJSON)); err != nil {
// 		return kibanaactions.CreateConnectorRequestBodyProperties{}, fmt.Errorf("%s - failed to unmarshal [secrets]: %w", prefixError, err)
// 	}

// 	c := kibanaactions.CreateConnectorRequestCasesWebhook{
// 		ConnectorTypeId: connector.ConnectorTypeID,
// 		Name:            connector.Name,
// 		Config:          *config.Get(),
// 		Secrets:         secrets.Get(),
// 	}

// 	return kibanaactions.CreateConnectorRequestCasesWebhookAsCreateConnectorRequestBodyProperties(&c), nil
// }

// func createConnectorRequestEmail(connector models.KibanaActionConnector) (kibanaactions.CreateConnectorRequestBodyProperties, error) {
// 	prefixError := "failed to compose create connector request for Email"

// 	var config map[string]interface{}
// 	if err := json.Unmarshal([]byte(connector.ConfigJSON), &config); err != nil {
// 		return kibanaactions.CreateConnectorRequestBodyProperties{}, fmt.Errorf("%s - failed to unmarshal [config] - %w", prefixError, err)
// 	}

// 	var secrets map[string]interface{}
// 	if err := json.Unmarshal([]byte(connector.SecretsJSON), &secrets); err != nil {
// 		return kibanaactions.CreateConnectorRequestBodyProperties{}, fmt.Errorf("%s - failed to unmarshal [secrets]: %w", prefixError, err)
// 	}

// 	c := kibanaactions.CreateConnectorRequestEmail{
// 		ConnectorTypeId: connector.ConnectorTypeID,
// 		Name:            connector.Name,
// 		Config:          config,
// 		Secrets:         secrets,
// 	}

// 	return kibanaactions.CreateConnectorRequestEmailAsCreateConnectorRequestBodyProperties(&c), nil
// }

func createConnectorRequestIndex(connector models.KibanaActionConnector) (kibanaactions.CreateConnectorRequestBodyProperties, error) {
	prefixError := "failed to compose create connector request for Index"

	var config kibanaactions.ConfigPropertiesIndex

	if err := json.Unmarshal([]byte(connector.ConfigJSON), &config); err != nil {
		return kibanaactions.CreateConnectorRequestBodyProperties{}, fmt.Errorf("%s - failed to unmarshal [config] - %w", prefixError, err)
	}

	return kibanaactions.CreateConnectorRequestBodyProperties{
		CreateConnectorRequestIndex: kibanaactions.CreateConnectorRequestIndex{
			ConnectorTypeId: connector.ConnectorTypeID,
			Name:            connector.Name,
			Config:          &config,
		},
	}, nil
}

// func createConnectorRequestJira(connector models.KibanaActionConnector) (kibanaactions.CreateConnectorRequestBodyProperties, error) {
// 	prefixError := "failed to compose create connector request for Jira"

// 	config := kibanaactions.NullableConfigPropertiesJira{}
// 	if err := config.UnmarshalJSON([]byte(connector.ConfigJSON)); err != nil {
// 		return kibanaactions.CreateConnectorRequestBodyProperties{}, fmt.Errorf("%s - failed to unmarshal [config] - %w", prefixError, err)
// 	}

// 	secrets := kibanaactions.NullableSecretsPropertiesJira{}
// 	if err := secrets.UnmarshalJSON([]byte(connector.SecretsJSON)); err != nil {
// 		return kibanaactions.CreateConnectorRequestBodyProperties{}, fmt.Errorf("%s - failed to unmarshal [secrets]: %w", prefixError, err)
// 	}

// 	c := kibanaactions.CreateConnectorRequestJira{
// 		ConnectorTypeId: connector.ConnectorTypeID,
// 		Name:            connector.Name,
// 		Config:          *config.Get(),
// 		Secrets:         *secrets.Get(),
// 	}

// 	return kibanaactions.CreateConnectorRequestJiraAsCreateConnectorRequestBodyProperties(&c), nil
// }

// func createConnectorRequestOpsgenie(connector models.KibanaActionConnector) (kibanaactions.CreateConnectorRequestBodyProperties, error) {
// 	prefixError := "failed to compose create connector request for Opsgenie"

// 	config := kibanaactions.NullableConfigPropertiesOpsgenie{}
// 	if err := config.UnmarshalJSON([]byte(connector.ConfigJSON)); err != nil {
// 		return kibanaactions.CreateConnectorRequestBodyProperties{}, fmt.Errorf("%s - failed to unmarshal [config] - %w", prefixError, err)
// 	}

// 	secrets := kibanaactions.NullableSecretsPropertiesOpsgenie{}
// 	if err := secrets.UnmarshalJSON([]byte(connector.SecretsJSON)); err != nil {
// 		return kibanaactions.CreateConnectorRequestBodyProperties{}, fmt.Errorf("%s - failed to unmarshal [secrets]: %w", prefixError, err)
// 	}

// 	c := kibanaactions.CreateConnectorRequestOpsgenie{
// 		ConnectorTypeId: connector.ConnectorTypeID,
// 		Name:            connector.Name,
// 		Config:          *config.Get(),
// 		Secrets:         *secrets.Get(),
// 	}

// 	return kibanaactions.CreateConnectorRequestOpsgenieAsCreateConnectorRequestBodyProperties(&c), nil
// }

// func createConnectorRequestPagerduty(connector models.KibanaActionConnector) (kibanaactions.CreateConnectorRequestBodyProperties, error) {
// 	prefixError := "failed to compose create connector request for PagerDuty"

// 	var config map[string]interface{}
// 	if err := json.Unmarshal([]byte(connector.ConfigJSON), &config); err != nil {
// 		return kibanaactions.CreateConnectorRequestBodyProperties{}, fmt.Errorf("%s - failed to unmarshal [config] - %w", prefixError, err)
// 	}

// 	var secrets map[string]interface{}
// 	if err := json.Unmarshal([]byte(connector.SecretsJSON), &secrets); err != nil {
// 		return kibanaactions.CreateConnectorRequestBodyProperties{}, fmt.Errorf("%s - failed to unmarshal [secrets]: %w", prefixError, err)
// 	}

// 	c := kibanaactions.CreateConnectorRequestPagerduty{
// 		ConnectorTypeId: connector.ConnectorTypeID,
// 		Name:            connector.Name,
// 		Config:          config,
// 		Secrets:         secrets,
// 	}

// 	return kibanaactions.CreateConnectorRequestPagerdutyAsCreateConnectorRequestBodyProperties(&c), nil
// }

// func createConnectorRequestResilient(connector models.KibanaActionConnector) (kibanaactions.CreateConnectorRequestBodyProperties, error) {
// 	prefixError := "failed to compose create connector request for Resilient"

// 	config := kibanaactions.NullableConfigPropertiesResilient{}
// 	if err := config.UnmarshalJSON([]byte(connector.ConfigJSON)); err != nil {
// 		return kibanaactions.CreateConnectorRequestBodyProperties{}, fmt.Errorf("%s - failed to unmarshal [config] - %w", prefixError, err)
// 	}

// 	secrets := kibanaactions.NullableSecretsPropertiesResilient{}
// 	if err := secrets.UnmarshalJSON([]byte(connector.SecretsJSON)); err != nil {
// 		return kibanaactions.CreateConnectorRequestBodyProperties{}, fmt.Errorf("%s - failed to unmarshal [secrets]: %w", prefixError, err)
// 	}

// 	c := kibanaactions.CreateConnectorRequestResilient{
// 		ConnectorTypeId: connector.ConnectorTypeID,
// 		Name:            connector.Name,
// 		Config:          *config.Get(),
// 		Secrets:         *secrets.Get(),
// 	}

// 	return kibanaactions.CreateConnectorRequestResilientAsCreateConnectorRequestBodyProperties(&c), nil
// }

// func createConnectorRequestServicenow(connector models.KibanaActionConnector) (kibanaactions.CreateConnectorRequestBodyProperties, error) {
// 	prefixError := "failed to compose create connector request for Servicenow"

// 	config := kibanaactions.NullableConfigPropertiesServicenow{}
// 	if err := config.UnmarshalJSON([]byte(connector.ConfigJSON)); err != nil {
// 		return kibanaactions.CreateConnectorRequestBodyProperties{}, fmt.Errorf("%s - failed to unmarshal [config] - %w", prefixError, err)
// 	}

// 	secrets := kibanaactions.NullableSecretsPropertiesServicenow{}
// 	if err := secrets.UnmarshalJSON([]byte(connector.SecretsJSON)); err != nil {
// 		return kibanaactions.CreateConnectorRequestBodyProperties{}, fmt.Errorf("%s - failed to unmarshal [secrets]: %w", prefixError, err)
// 	}

// 	c := kibanaactions.CreateConnectorRequestServicenow{
// 		ConnectorTypeId: connector.ConnectorTypeID,
// 		Name:            connector.Name,
// 		Config:          *config.Get(),
// 		Secrets:         *secrets.Get(),
// 	}

// 	return kibanaactions.CreateConnectorRequestServicenowAsCreateConnectorRequestBodyProperties(&c), nil
// }

// func createConnectorRequestServicenowItom(connector models.KibanaActionConnector) (kibanaactions.CreateConnectorRequestBodyProperties, error) {
// 	prefixError := "failed to compose create connector request for ServicenowItom"

// 	config := kibanaactions.NullableConfigPropertiesServicenowItom{}
// 	if err := config.UnmarshalJSON([]byte(connector.ConfigJSON)); err != nil {
// 		return kibanaactions.CreateConnectorRequestBodyProperties{}, fmt.Errorf("%s - failed to unmarshal [config] - %w", prefixError, err)
// 	}

// 	secrets := kibanaactions.NullableSecretsPropertiesServicenow{}
// 	if err := secrets.UnmarshalJSON([]byte(connector.SecretsJSON)); err != nil {
// 		return kibanaactions.CreateConnectorRequestBodyProperties{}, fmt.Errorf("%s - failed to unmarshal [secrets]: %w", prefixError, err)
// 	}

// 	c := kibanaactions.CreateConnectorRequestServicenowItom{
// 		ConnectorTypeId: connector.ConnectorTypeID,
// 		Name:            connector.Name,
// 		Config:          *config.Get(),
// 		Secrets:         *secrets.Get(),
// 	}

// 	return kibanaactions.CreateConnectorRequestServicenowItomAsCreateConnectorRequestBodyProperties(&c), nil
// }

// func createConnectorRequestServicenowSir(connector models.KibanaActionConnector) (kibanaactions.CreateConnectorRequestBodyProperties, error) {
// 	prefixError := "failed to compose create connector request for ServicenowSir"

// 	config := kibanaactions.NullableConfigPropertiesServicenow{}
// 	if err := config.UnmarshalJSON([]byte(connector.ConfigJSON)); err != nil {
// 		return kibanaactions.CreateConnectorRequestBodyProperties{}, fmt.Errorf("%s - failed to unmarshal [config] - %w", prefixError, err)
// 	}

// 	secrets := kibanaactions.NullableSecretsPropertiesServicenow{}
// 	if err := secrets.UnmarshalJSON([]byte(connector.SecretsJSON)); err != nil {
// 		return kibanaactions.CreateConnectorRequestBodyProperties{}, fmt.Errorf("%s - failed to unmarshal [secrets]: %w", prefixError, err)
// 	}

// 	c := kibanaactions.CreateConnectorRequestServicenowSir{
// 		ConnectorTypeId: connector.ConnectorTypeID,
// 		Name:            connector.Name,
// 		Config:          *config.Get(),
// 		Secrets:         *secrets.Get(),
// 	}

// 	return kibanaactions.CreateConnectorRequestServicenowSirAsCreateConnectorRequestBodyProperties(&c), nil
// }

// func createConnectorRequestServerLog(connector models.KibanaActionConnector) (kibanaactions.CreateConnectorRequestBodyProperties, error) {
// 	c := kibanaactions.CreateConnectorRequestServerlog{
// 		ConnectorTypeId: connector.ConnectorTypeID,
// 		Name:            connector.Name,
// 	}

// 	return kibanaactions.CreateConnectorRequestServerlogAsCreateConnectorRequestBodyProperties(&c), nil
// }

// func createConnectorRequestSlack(connector models.KibanaActionConnector) (kibanaactions.CreateConnectorRequestBodyProperties, error) {
// 	prefixError := "failed to compose create connector request for Slack"

// 	var secrets map[string]interface{}
// 	if err := json.Unmarshal([]byte(connector.SecretsJSON), &secrets); err != nil {
// 		return kibanaactions.CreateConnectorRequestBodyProperties{}, fmt.Errorf("%s - failed to unmarshal [secrets]: %w", prefixError, err)
// 	}

// 	c := kibanaactions.CreateConnectorRequestSlack{
// 		ConnectorTypeId: connector.ConnectorTypeID,
// 		Name:            connector.Name,
// 		Secrets:         secrets,
// 	}

// 	return kibanaactions.CreateConnectorRequestSlackAsCreateConnectorRequestBodyProperties(&c), nil
// }

// func createConnectorRequestSwimlane(connector models.KibanaActionConnector) (kibanaactions.CreateConnectorRequestBodyProperties, error) {
// 	prefixError := "failed to compose create connector request for Swimlane"

// 	config := kibanaactions.NullableConfigPropertiesSwimlane{}
// 	if err := config.UnmarshalJSON([]byte(connector.ConfigJSON)); err != nil {
// 		return kibanaactions.CreateConnectorRequestBodyProperties{}, fmt.Errorf("%s - failed to unmarshal [config] - %w", prefixError, err)
// 	}

// 	secrets := kibanaactions.NullableSecretsPropertiesSwimlane{}
// 	if err := secrets.UnmarshalJSON([]byte(connector.SecretsJSON)); err != nil {
// 		return kibanaactions.CreateConnectorRequestBodyProperties{}, fmt.Errorf("%s - failed to unmarshal [secrets]: %w", prefixError, err)
// 	}

// 	c := kibanaactions.CreateConnectorRequestSwimlane{
// 		ConnectorTypeId: connector.ConnectorTypeID,
// 		Name:            connector.Name,
// 		Config:          *config.Get(),
// 		Secrets:         *secrets.Get(),
// 	}

// 	return kibanaactions.CreateConnectorRequestSwimlaneAsCreateConnectorRequestBodyProperties(&c), nil
// }

// func createConnectorRequestTeams(connector models.KibanaActionConnector) (kibanaactions.CreateConnectorRequestBodyProperties, error) {
// 	prefixError := "failed to compose create connector request for Teams"

// 	var secrets map[string]interface{}
// 	if err := json.Unmarshal([]byte(connector.SecretsJSON), &secrets); err != nil {
// 		return kibanaactions.CreateConnectorRequestBodyProperties{}, fmt.Errorf("%s - failed to unmarshal [secrets]: %w", prefixError, err)
// 	}

// 	c := kibanaactions.CreateConnectorRequestTeams{
// 		ConnectorTypeId: connector.ConnectorTypeID,
// 		Name:            connector.Name,
// 		Secrets:         secrets,
// 	}

// 	return kibanaactions.CreateConnectorRequestTeamsAsCreateConnectorRequestBodyProperties(&c), nil
// }

// func createConnectorRequestTines(connector models.KibanaActionConnector) (kibanaactions.CreateConnectorRequestBodyProperties, error) {
// 	prefixError := "failed to compose create connector request for Tines"

// 	var config map[string]interface{}
// 	if err := json.Unmarshal([]byte(connector.ConfigJSON), &config); err != nil {
// 		return kibanaactions.CreateConnectorRequestBodyProperties{}, fmt.Errorf("%s - failed to unmarshal [config] - %w", prefixError, err)
// 	}

// 	var secrets map[string]interface{}
// 	if err := json.Unmarshal([]byte(connector.SecretsJSON), &secrets); err != nil {
// 		return kibanaactions.CreateConnectorRequestBodyProperties{}, fmt.Errorf("%s - failed to unmarshal [secrets]: %w", prefixError, err)
// 	}

// 	c := kibanaactions.CreateConnectorRequestTines{
// 		ConnectorTypeId: connector.ConnectorTypeID,
// 		Name:            connector.Name,
// 		Config:          config,
// 		Secrets:         secrets,
// 	}

// 	return kibanaactions.CreateConnectorRequestTinesAsCreateConnectorRequestBodyProperties(&c), nil
// }

// func createConnectorRequestWebhook(connector models.KibanaActionConnector) (kibanaactions.CreateConnectorRequestBodyProperties, error) {
// 	prefixError := "failed to compose create connector request for Webhook"

// 	var config map[string]interface{}
// 	if err := json.Unmarshal([]byte(connector.ConfigJSON), &config); err != nil {
// 		return kibanaactions.CreateConnectorRequestBodyProperties{}, fmt.Errorf("%s - failed to unmarshal [config] - %w", prefixError, err)
// 	}

// 	var secrets map[string]interface{}
// 	if err := json.Unmarshal([]byte(connector.SecretsJSON), &secrets); err != nil {
// 		return kibanaactions.CreateConnectorRequestBodyProperties{}, fmt.Errorf("%s - failed to unmarshal [secrets]: %w", prefixError, err)
// 	}

// 	c := kibanaactions.CreateConnectorRequestWebhook{
// 		ConnectorTypeId: connector.ConnectorTypeID,
// 		Name:            connector.Name,
// 		Config:          config,
// 		Secrets:         secrets,
// 	}

// 	return kibanaactions.CreateConnectorRequestWebhookAsCreateConnectorRequestBodyProperties(&c), nil
// }

// func createConnectorRequestXmatters(connector models.KibanaActionConnector) (kibanaactions.CreateConnectorRequestBodyProperties, error) {
// 	prefixError := "failed to compose create connector request for Xmatters"

// 	var config map[string]interface{}
// 	if err := json.Unmarshal([]byte(connector.ConfigJSON), &config); err != nil {
// 		return kibanaactions.CreateConnectorRequestBodyProperties{}, fmt.Errorf("%s - failed to unmarshal [config] - %w", prefixError, err)
// 	}

// 	var secrets map[string]interface{}
// 	if err := json.Unmarshal([]byte(connector.SecretsJSON), &secrets); err != nil {
// 		return kibanaactions.CreateConnectorRequestBodyProperties{}, fmt.Errorf("%s - failed to unmarshal [secrets]: %w", prefixError, err)
// 	}

// 	c := kibanaactions.CreateConnectorRequestXmatters{
// 		ConnectorTypeId: connector.ConnectorTypeID,
// 		Name:            connector.Name,
// 		Config:          config,
// 		Secrets:         secrets,
// 	}

// 	return kibanaactions.CreateConnectorRequestXmattersAsCreateConnectorRequestBodyProperties(&c), nil
// }

// func updateConnectorRequestCasesWebhook(connector models.KibanaActionConnector) (kibanaactions.UpdateConnectorRequestBodyProperties, error) {
// 	prefixError := "failed to compose update connector request for CasesWebhook"

// 	nullableConfig := kibanaactions.NullableConfigPropertiesCasesWebhook{}
// 	if err := nullableConfig.UnmarshalJSON([]byte(connector.ConfigJSON)); err != nil {
// 		return kibanaactions.UpdateConnectorRequestBodyProperties{}, fmt.Errorf("%s - failed to unmarshal [config] - %w", prefixError, err)
// 	}

// 	nullableSecrets := kibanaactions.NullableSecretsPropertiesCasesWebhook{}
// 	if err := nullableSecrets.UnmarshalJSON([]byte(connector.SecretsJSON)); err != nil {
// 		return kibanaactions.UpdateConnectorRequestBodyProperties{}, fmt.Errorf("%s - failed to unmarshal [secrets]: %w", prefixError, err)
// 	}

// 	return kibanaactions.UpdateConnectorRequestBodyProperties {
// 		 kibanaactions.UpdateConnectorRequestCasesWebhook{
// 		Name:    connector.Name,
// 		Config:  *nullableConfig.Get(),
// 		Secrets: nullableSecrets.Get(),
// 	}

// 	return kibanaactions.UpdateConnectorRequestCasesWebhookAsUpdateConnectorRequestBodyProperties(&c), nil
// }

func updateConnectorRequestIndex(connector models.KibanaActionConnector) (kibanaactions.UpdateConnectorRequestBodyProperties, error) {
	prefixError := "failed to compose update connector request for Index"

	var config kibanaactions.ConfigPropertiesIndex
	if err := json.Unmarshal([]byte(connector.ConfigJSON), &config); err != nil {
		return kibanaactions.UpdateConnectorRequestBodyProperties{}, fmt.Errorf("%s - failed to unmarshal [config] - %w", prefixError, err)
	}

	return kibanaactions.UpdateConnectorRequestBodyProperties{
		UpdateConnectorRequestIndex: kibanaactions.UpdateConnectorRequestIndex{
			Name:   connector.Name,
			Config: &config,
		},
	}, nil
}

// func updateConnectorRequestJira(connector models.KibanaActionConnector) (kibanaactions.UpdateConnectorRequestBodyProperties, error) {
// 	prefixError := "failed to compose update connector request for Jira"

// 	config := kibanaactions.NullableConfigPropertiesJira{}
// 	if err := config.UnmarshalJSON([]byte(connector.ConfigJSON)); err != nil {
// 		return kibanaactions.UpdateConnectorRequestBodyProperties{}, fmt.Errorf("%s - failed to unmarshal [config] - %w", prefixError, err)
// 	}

// 	secrets := kibanaactions.NullableSecretsPropertiesJira{}
// 	if err := secrets.UnmarshalJSON([]byte(connector.SecretsJSON)); err != nil {
// 		return kibanaactions.UpdateConnectorRequestBodyProperties{}, fmt.Errorf("%s - failed to unmarshal [secrets]: %w", prefixError, err)
// 	}

// 	c := kibanaactions.UpdateConnectorRequestJira{
// 		Name:    connector.Name,
// 		Config:  *config.Get(),
// 		Secrets: *secrets.Get(),
// 	}

// 	return kibanaactions.UpdateConnectorRequestJiraAsUpdateConnectorRequestBodyProperties(&c), nil
// }

// func updateConnectorRequestOpsgenie(connector models.KibanaActionConnector) (kibanaactions.UpdateConnectorRequestBodyProperties, error) {
// 	prefixError := "failed to compose update connector request for Opsgenie"

// 	config := kibanaactions.NullableConfigPropertiesOpsgenie{}
// 	if err := config.UnmarshalJSON([]byte(connector.ConfigJSON)); err != nil {
// 		return kibanaactions.UpdateConnectorRequestBodyProperties{}, fmt.Errorf("%s - failed to unmarshal [config] - %w", prefixError, err)
// 	}

// 	secrets := kibanaactions.NullableSecretsPropertiesOpsgenie{}
// 	if err := secrets.UnmarshalJSON([]byte(connector.SecretsJSON)); err != nil {
// 		return kibanaactions.UpdateConnectorRequestBodyProperties{}, fmt.Errorf("%s - failed to unmarshal [secrets]: %w", prefixError, err)
// 	}

// 	c := kibanaactions.UpdateConnectorRequestOpsgenie{
// 		Name:    connector.Name,
// 		Config:  *config.Get(),
// 		Secrets: *secrets.Get(),
// 	}

// 	return kibanaactions.UpdateConnectorRequestOpsgenieAsUpdateConnectorRequestBodyProperties(&c), nil
// }

// func updateConnectorRequestResilient(connector models.KibanaActionConnector) (kibanaactions.UpdateConnectorRequestBodyProperties, error) {
// 	prefixError := "failed to compose update connector request for Resilient"

// 	config := kibanaactions.NullableConfigPropertiesResilient{}
// 	if err := config.UnmarshalJSON([]byte(connector.ConfigJSON)); err != nil {
// 		return kibanaactions.UpdateConnectorRequestBodyProperties{}, fmt.Errorf("%s - failed to unmarshal [config] - %w", prefixError, err)
// 	}

// 	secrets := kibanaactions.NullableSecretsPropertiesResilient{}
// 	if err := secrets.UnmarshalJSON([]byte(connector.SecretsJSON)); err != nil {
// 		return kibanaactions.UpdateConnectorRequestBodyProperties{}, fmt.Errorf("%s - failed to unmarshal [secrets]: %w", prefixError, err)
// 	}

// 	c := kibanaactions.UpdateConnectorRequestResilient{
// 		Name:    connector.Name,
// 		Config:  *config.Get(),
// 		Secrets: *secrets.Get(),
// 	}

// 	return kibanaactions.UpdateConnectorRequestResilientAsUpdateConnectorRequestBodyProperties(&c), nil
// }

// func updateConnectorRequestServicenow(connector models.KibanaActionConnector) (kibanaactions.UpdateConnectorRequestBodyProperties, error) {
// 	prefixError := "failed to compose update connector request for Servicenow"

// 	config := kibanaactions.NullableConfigPropertiesServicenow{}
// 	if err := config.UnmarshalJSON([]byte(connector.ConfigJSON)); err != nil {
// 		return kibanaactions.UpdateConnectorRequestBodyProperties{}, fmt.Errorf("%s - failed to unmarshal [config] - %w", prefixError, err)
// 	}

// 	secrets := kibanaactions.NullableSecretsPropertiesServicenow{}
// 	if err := secrets.UnmarshalJSON([]byte(connector.SecretsJSON)); err != nil {
// 		return kibanaactions.UpdateConnectorRequestBodyProperties{}, fmt.Errorf("%s - failed to unmarshal [secrets]: %w", prefixError, err)
// 	}

// 	c := kibanaactions.UpdateConnectorRequestServicenow{
// 		Name:    connector.Name,
// 		Config:  *config.Get(),
// 		Secrets: *secrets.Get(),
// 	}

// 	return kibanaactions.UpdateConnectorRequestServicenowAsUpdateConnectorRequestBodyProperties(&c), nil
// }

// func updateConnectorRequestServicenowItom(connector models.KibanaActionConnector) (kibanaactions.UpdateConnectorRequestBodyProperties, error) {
// 	prefixError := "failed to compose update connector request for ServicenowItom"

// 	config := kibanaactions.NullableConfigPropertiesServicenowItom{}
// 	if err := config.UnmarshalJSON([]byte(connector.ConfigJSON)); err != nil {
// 		return kibanaactions.UpdateConnectorRequestBodyProperties{}, fmt.Errorf("%s - failed to unmarshal [config] - %w", prefixError, err)
// 	}

// 	secrets := kibanaactions.NullableSecretsPropertiesServicenow{}
// 	if err := secrets.UnmarshalJSON([]byte(connector.SecretsJSON)); err != nil {
// 		return kibanaactions.UpdateConnectorRequestBodyProperties{}, fmt.Errorf("%s - failed to unmarshal [secrets]: %w", prefixError, err)
// 	}

// 	c := kibanaactions.UpdateConnectorRequestServicenowItom{
// 		Name:    connector.Name,
// 		Config:  *config.Get(),
// 		Secrets: *secrets.Get(),
// 	}

// 	return kibanaactions.UpdateConnectorRequestServicenowItomAsUpdateConnectorRequestBodyProperties(&c), nil
// }

// func updateConnectorRequestServerLog(connector models.KibanaActionConnector) (kibanaactions.UpdateConnectorRequestBodyProperties, error) {
// 	c := kibanaactions.UpdateConnectorRequestServerlog{
// 		Name: connector.Name,
// 	}

// 	return kibanaactions.UpdateConnectorRequestServerlogAsUpdateConnectorRequestBodyProperties(&c), nil
// }

// func updateConnectorRequestSwimlane(connector models.KibanaActionConnector) (kibanaactions.UpdateConnectorRequestBodyProperties, error) {
// 	prefixError := "failed to compose update connector request for Swimlane"

// 	config := kibanaactions.NullableConfigPropertiesSwimlane{}
// 	if err := config.UnmarshalJSON([]byte(connector.ConfigJSON)); err != nil {
// 		return kibanaactions.UpdateConnectorRequestBodyProperties{}, fmt.Errorf("%s - failed to unmarshal [config] - %w", prefixError, err)
// 	}

// 	secrets := kibanaactions.NullableSecretsPropertiesSwimlane{}
// 	if err := secrets.UnmarshalJSON([]byte(connector.SecretsJSON)); err != nil {
// 		return kibanaactions.UpdateConnectorRequestBodyProperties{}, fmt.Errorf("%s - failed to unmarshal [secrets]: %w", prefixError, err)
// 	}

// 	c := kibanaactions.UpdateConnectorRequestSwimlane{
// 		Name:    connector.Name,
// 		Config:  *config.Get(),
// 		Secrets: *secrets.Get(),
// 	}

// 	return kibanaactions.UpdateConnectorRequestSwimlaneAsUpdateConnectorRequestBodyProperties(&c), nil
// }

func actionConnectorToModel(spaceID string, properties kibanaactions.ConnectorResponseProperties, connectorTypeID string) (*models.KibanaActionConnector, error) {
	switch kibanaactions.ConnectorTypes(connectorTypeID) {

	// case kibanaactions.CASES_WEBHOOK_ConnectorTypes:
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

	// case *kibanaactions.ConnectorResponsePropertiesEmail:
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

	case kibanaactions.INDEX_ConnectorTypes:
		config, err := json.Marshal(properties.ConnectorResponsePropertiesIndex.Config)
		if err != nil {
			return nil, fmt.Errorf("failed to parse [config] in ConnectorResponsePropertiesCasesWebhook - [%w]", err)
		}
		connector := models.KibanaActionConnector{
			ConnectorID:      properties.ConnectorResponsePropertiesIndex.Id,
			SpaceID:          spaceID,
			Name:             properties.ConnectorResponsePropertiesIndex.Name,
			ConnectorTypeID:  connectorTypeID,
			IsDeprecated:     properties.ConnectorResponsePropertiesIndex.IsDeprecated,
			IsMissingSecrets: properties.ConnectorResponsePropertiesIndex.IsMissingSecrets,
			IsPreconfigured:  properties.ConnectorResponsePropertiesIndex.IsPreconfigured,
			ConfigJSON:       string(config),
		}
		return &connector, nil

		// case *kibanaactions.ConnectorResponsePropertiesJira:
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

		// case *kibanaactions.ConnectorResponsePropertiesOpsgenie:
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

		// case *kibanaactions.ConnectorResponsePropertiesPagerduty:
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

		// case *kibanaactions.ConnectorResponsePropertiesResilient:
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

		// case *kibanaactions.ConnectorResponsePropertiesServerlog:
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

		// case *kibanaactions.ConnectorResponsePropertiesServicenow:
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

		// case *kibanaactions.ConnectorResponsePropertiesServicenowItom:
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

		// case *kibanaactions.ConnectorResponsePropertiesServicenowSir:
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

		// case *kibanaactions.ConnectorResponsePropertiesSlack:
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

		// case *kibanaactions.ConnectorResponsePropertiesSwimlane:
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

		// case *kibanaactions.ConnectorResponsePropertiesTeams:
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

		// case *kibanaactions.ConnectorResponsePropertiesTines:
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

		// case *kibanaactions.ConnectorResponsePropertiesWebhook:
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

		// case *kibanaactions.ConnectorResponsePropertiesXmatters:
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
// 	*kibanaactions.ConnectorResponsePropertiesCasesWebhook | *kibanaactions.ConnectorResponsePropertiesEmail |
// 		*kibanaactions.ConnectorResponsePropertiesIndex
// 	GetId() string
// 	GetName() string
// 	GetConnectorTypeId() string
// 	GetIsDeprecated() bool
// 	GetIsMissingSecrets() bool
// 	GetIsPreconfigured() bool
// }
