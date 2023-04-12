package kibana

import (
	"context"
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

	req := client.CreateConnector(ctxWithAuth, connectorOld.SpaceID).KbnXsrf("true").CreateConnectorRequestBodyProperties(createProperties)

	responseProperties, httpRes, err := req.Execute()
	if err != nil && httpRes == nil {
		return "", diag.FromErr(err)
	}
	defer httpRes.Body.Close()

	if diags := utils.CheckHttpError(httpRes, "Unabled to create action connector"); diags.HasError() {
		return "", diag.FromErr(err)
	}

	connectorNew, err := actionConnectorToModel(connectorOld.SpaceID, responseProperties)
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

	req := client.UpdateConnector(ctxWithAuth, connectorOld.ConnectorID, connectorOld.SpaceID).KbnXsrf("true").UpdateConnectorRequestBodyProperties(updateProperties)

	responseProperties, httpRes, err := req.Execute()
	if err != nil && httpRes == nil {
		return "", diag.FromErr(err)
	}
	defer httpRes.Body.Close()

	if diags := utils.CheckHttpError(httpRes, "Unabled to update action connector"); diags.HasError() {
		return "", diags
	}

	connectorNew, err := actionConnectorToModel(connectorOld.SpaceID, responseProperties)
	if err != nil {
		return "", diag.FromErr(err)
	}

	return connectorNew.ConnectorID, nil
}

func GetActionConnector(ctx context.Context, apiClient *clients.ApiClient, connectorID, spaceID string) (*models.KibanaActionConnector, diag.Diagnostics) {
	client, ctxWithAuth, err := apiClient.GetKibanaActionConnectorClient(ctx)
	if err != nil {
		return nil, diag.FromErr(err)
	}

	req := client.GetConnector(ctxWithAuth, connectorID, spaceID)

	properties, res, err := req.Execute()
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

	connector, err := actionConnectorToModel(spaceID, properties)
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

	req := client.DeleteConnector(ctxWithAuth, connectorID, spaceID).KbnXsrf("true")
	res, err := req.Execute()
	if err != nil && res == nil {
		return diag.FromErr(err)
	}
	defer res.Body.Close()

	return utils.CheckHttpError(res, "Unabled to delete action connector")
}

func createConnectorRequestBodyProperties(connector models.KibanaActionConnector) (kibanaactions.CreateConnectorRequestBodyProperties, error) {
	connectorType, err := kibanaactions.NewConnectorTypesFromValue(connector.ConnectorTypeID)
	if err != nil {
		return kibanaactions.CreateConnectorRequestBodyProperties{}, err
	}

	switch *connectorType {
	case kibanaactions.CASES_WEBHOOK:
		return createConnectorRequestCasesWebhook(connector)
	// case kibanaactions.EMAIL:
	// 	return createConnectorRequestEmail(connector)
	case kibanaactions.INDEX:
		return createConnectorRequestIndex(connector)
	case kibanaactions.JIRA:
		return createConnectorRequestJira(connector)
	case kibanaactions.OPSGENIE:
		return createConnectorRequestOpsgenie(connector)
	// case kibanaactions.PAGERDUTY:
	// 	return createConnectorRequestPagerduty(connector)
	case kibanaactions.RESILIENT:
		return createConnectorRequestResilient(connector)
	case kibanaactions.SERVICENOW:
		return createConnectorRequestServicenow(connector)
	case kibanaactions.SERVICENOW_ITOM:
		return createConnectorRequestServicenowItom(connector)
	// case kibanaactions.SERVICENOW_SIR:
	// 	return createConnectorRequestServicenowSir(connector)
	case kibanaactions.SERVER_LOG:
		return createConnectorRequestServerLog(connector)
	// case kibanaactions.SLACK:
	// 	return createConnectorRequestSlack(connector)
	case kibanaactions.SWIMLANE:
		return createConnectorRequestSwimlane(connector)
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
	connectorType, err := kibanaactions.NewConnectorTypesFromValue(connector.ConnectorTypeID)
	if err != nil {
		return kibanaactions.UpdateConnectorRequestBodyProperties{}, err
	}

	switch *connectorType {
	case kibanaactions.CASES_WEBHOOK:
		return updateConnectorRequestCasesWebhook(connector)
	// case kibanaactions.EMAIL:
	// 	return updateConnectorRequestEmail(connector)
	case kibanaactions.INDEX:
		return updateConnectorRequestIndex(connector)
	case kibanaactions.JIRA:
		return updateConnectorRequestJira(connector)
	case kibanaactions.OPSGENIE:
		return updateConnectorRequestOpsgenie(connector)
	// case kibanaactions.PAGERDUTY:
	// 	return updateConnectorRequestPagerduty(connector)
	case kibanaactions.RESILIENT:
		return updateConnectorRequestResilient(connector)
	case kibanaactions.SERVICENOW:
		return updateConnectorRequestServicenow(connector)
	case kibanaactions.SERVICENOW_ITOM:
		return updateConnectorRequestServicenowItom(connector)
	// case kibanaactions.SERVICENOW_SIR:
	// 	return updateConnectorRequestServicenowSir(connector)
	case kibanaactions.SERVER_LOG:
		return updateConnectorRequestServerLog(connector)
	// case kibanaactions.SLACK:
	// 	return updateConnectorRequestSlack(connector)
	case kibanaactions.SWIMLANE:
		return updateConnectorRequestSwimlane(connector)
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

func createConnectorRequestCasesWebhook(connector models.KibanaActionConnector) (kibanaactions.CreateConnectorRequestBodyProperties, error) {
	prefixError := "failed to compose create connector request for CasesWebhook"

	config := kibanaactions.NullableConfigPropertiesCasesWebhook{}
	if err := config.UnmarshalJSON([]byte(connector.ConfigJSON)); err != nil {
		return kibanaactions.CreateConnectorRequestBodyProperties{}, fmt.Errorf("%s - failed to unmarshal [config] - %w", prefixError, err)
	}

	secrets := kibanaactions.NullableSecretsPropertiesCasesWebhook{}
	if err := secrets.UnmarshalJSON([]byte(connector.SecretsJSON)); err != nil {
		return kibanaactions.CreateConnectorRequestBodyProperties{}, fmt.Errorf("%s - failed to unmarshal [secrets]: %w", prefixError, err)
	}

	c := kibanaactions.CreateConnectorRequestCasesWebhook{
		ConnectorTypeId: connector.ConnectorTypeID,
		Name:            connector.Name,
		Config:          *config.Get(),
		Secrets:         secrets.Get(),
	}

	return kibanaactions.CreateConnectorRequestCasesWebhookAsCreateConnectorRequestBodyProperties(&c), nil
}

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

	config := kibanaactions.NullableConfigPropertiesIndex{}
	if err := config.UnmarshalJSON([]byte(connector.ConfigJSON)); err != nil {
		return kibanaactions.CreateConnectorRequestBodyProperties{}, fmt.Errorf("%s - failed to unmarshal [config] - %w", prefixError, err)
	}

	c := kibanaactions.CreateConnectorRequestIndex{
		ConnectorTypeId: connector.ConnectorTypeID,
		Name:            connector.Name,
		Config:          *config.Get(),
	}

	return kibanaactions.CreateConnectorRequestIndexAsCreateConnectorRequestBodyProperties(&c), nil
}

func createConnectorRequestJira(connector models.KibanaActionConnector) (kibanaactions.CreateConnectorRequestBodyProperties, error) {
	prefixError := "failed to compose create connector request for Jira"

	config := kibanaactions.NullableConfigPropertiesJira{}
	if err := config.UnmarshalJSON([]byte(connector.ConfigJSON)); err != nil {
		return kibanaactions.CreateConnectorRequestBodyProperties{}, fmt.Errorf("%s - failed to unmarshal [config] - %w", prefixError, err)
	}

	secrets := kibanaactions.NullableSecretsPropertiesJira{}
	if err := secrets.UnmarshalJSON([]byte(connector.SecretsJSON)); err != nil {
		return kibanaactions.CreateConnectorRequestBodyProperties{}, fmt.Errorf("%s - failed to unmarshal [secrets]: %w", prefixError, err)
	}

	c := kibanaactions.CreateConnectorRequestJira{
		ConnectorTypeId: connector.ConnectorTypeID,
		Name:            connector.Name,
		Config:          *config.Get(),
		Secrets:         *secrets.Get(),
	}

	return kibanaactions.CreateConnectorRequestJiraAsCreateConnectorRequestBodyProperties(&c), nil
}

func createConnectorRequestOpsgenie(connector models.KibanaActionConnector) (kibanaactions.CreateConnectorRequestBodyProperties, error) {
	prefixError := "failed to compose create connector request for Opsgenie"

	config := kibanaactions.NullableConfigPropertiesOpsgenie{}
	if err := config.UnmarshalJSON([]byte(connector.ConfigJSON)); err != nil {
		return kibanaactions.CreateConnectorRequestBodyProperties{}, fmt.Errorf("%s - failed to unmarshal [config] - %w", prefixError, err)
	}

	secrets := kibanaactions.NullableSecretsPropertiesOpsgenie{}
	if err := secrets.UnmarshalJSON([]byte(connector.SecretsJSON)); err != nil {
		return kibanaactions.CreateConnectorRequestBodyProperties{}, fmt.Errorf("%s - failed to unmarshal [secrets]: %w", prefixError, err)
	}

	c := kibanaactions.CreateConnectorRequestOpsgenie{
		ConnectorTypeId: connector.ConnectorTypeID,
		Name:            connector.Name,
		Config:          *config.Get(),
		Secrets:         *secrets.Get(),
	}

	return kibanaactions.CreateConnectorRequestOpsgenieAsCreateConnectorRequestBodyProperties(&c), nil
}

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

func createConnectorRequestResilient(connector models.KibanaActionConnector) (kibanaactions.CreateConnectorRequestBodyProperties, error) {
	prefixError := "failed to compose create connector request for Resilient"

	config := kibanaactions.NullableConfigPropertiesResilient{}
	if err := config.UnmarshalJSON([]byte(connector.ConfigJSON)); err != nil {
		return kibanaactions.CreateConnectorRequestBodyProperties{}, fmt.Errorf("%s - failed to unmarshal [config] - %w", prefixError, err)
	}

	secrets := kibanaactions.NullableSecretsPropertiesResilient{}
	if err := secrets.UnmarshalJSON([]byte(connector.SecretsJSON)); err != nil {
		return kibanaactions.CreateConnectorRequestBodyProperties{}, fmt.Errorf("%s - failed to unmarshal [secrets]: %w", prefixError, err)
	}

	c := kibanaactions.CreateConnectorRequestResilient{
		ConnectorTypeId: connector.ConnectorTypeID,
		Name:            connector.Name,
		Config:          *config.Get(),
		Secrets:         *secrets.Get(),
	}

	return kibanaactions.CreateConnectorRequestResilientAsCreateConnectorRequestBodyProperties(&c), nil
}

func createConnectorRequestServicenow(connector models.KibanaActionConnector) (kibanaactions.CreateConnectorRequestBodyProperties, error) {
	prefixError := "failed to compose create connector request for Servicenow"

	config := kibanaactions.NullableConfigPropertiesServicenow{}
	if err := config.UnmarshalJSON([]byte(connector.ConfigJSON)); err != nil {
		return kibanaactions.CreateConnectorRequestBodyProperties{}, fmt.Errorf("%s - failed to unmarshal [config] - %w", prefixError, err)
	}

	secrets := kibanaactions.NullableSecretsPropertiesServicenow{}
	if err := secrets.UnmarshalJSON([]byte(connector.SecretsJSON)); err != nil {
		return kibanaactions.CreateConnectorRequestBodyProperties{}, fmt.Errorf("%s - failed to unmarshal [secrets]: %w", prefixError, err)
	}

	c := kibanaactions.CreateConnectorRequestServicenow{
		ConnectorTypeId: connector.ConnectorTypeID,
		Name:            connector.Name,
		Config:          *config.Get(),
		Secrets:         *secrets.Get(),
	}

	return kibanaactions.CreateConnectorRequestServicenowAsCreateConnectorRequestBodyProperties(&c), nil
}

func createConnectorRequestServicenowItom(connector models.KibanaActionConnector) (kibanaactions.CreateConnectorRequestBodyProperties, error) {
	prefixError := "failed to compose create connector request for ServicenowItom"

	config := kibanaactions.NullableConfigPropertiesServicenowItom{}
	if err := config.UnmarshalJSON([]byte(connector.ConfigJSON)); err != nil {
		return kibanaactions.CreateConnectorRequestBodyProperties{}, fmt.Errorf("%s - failed to unmarshal [config] - %w", prefixError, err)
	}

	secrets := kibanaactions.NullableSecretsPropertiesServicenow{}
	if err := secrets.UnmarshalJSON([]byte(connector.SecretsJSON)); err != nil {
		return kibanaactions.CreateConnectorRequestBodyProperties{}, fmt.Errorf("%s - failed to unmarshal [secrets]: %w", prefixError, err)
	}

	c := kibanaactions.CreateConnectorRequestServicenowItom{
		ConnectorTypeId: connector.ConnectorTypeID,
		Name:            connector.Name,
		Config:          *config.Get(),
		Secrets:         *secrets.Get(),
	}

	return kibanaactions.CreateConnectorRequestServicenowItomAsCreateConnectorRequestBodyProperties(&c), nil
}

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

func createConnectorRequestServerLog(connector models.KibanaActionConnector) (kibanaactions.CreateConnectorRequestBodyProperties, error) {
	c := kibanaactions.CreateConnectorRequestServerlog{
		ConnectorTypeId: connector.ConnectorTypeID,
		Name:            connector.Name,
	}

	return kibanaactions.CreateConnectorRequestServerlogAsCreateConnectorRequestBodyProperties(&c), nil
}

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

func createConnectorRequestSwimlane(connector models.KibanaActionConnector) (kibanaactions.CreateConnectorRequestBodyProperties, error) {
	prefixError := "failed to compose create connector request for Swimlane"

	config := kibanaactions.NullableConfigPropertiesSwimlane{}
	if err := config.UnmarshalJSON([]byte(connector.ConfigJSON)); err != nil {
		return kibanaactions.CreateConnectorRequestBodyProperties{}, fmt.Errorf("%s - failed to unmarshal [config] - %w", prefixError, err)
	}

	secrets := kibanaactions.NullableSecretsPropertiesSwimlane{}
	if err := secrets.UnmarshalJSON([]byte(connector.SecretsJSON)); err != nil {
		return kibanaactions.CreateConnectorRequestBodyProperties{}, fmt.Errorf("%s - failed to unmarshal [secrets]: %w", prefixError, err)
	}

	c := kibanaactions.CreateConnectorRequestSwimlane{
		ConnectorTypeId: connector.ConnectorTypeID,
		Name:            connector.Name,
		Config:          *config.Get(),
		Secrets:         *secrets.Get(),
	}

	return kibanaactions.CreateConnectorRequestSwimlaneAsCreateConnectorRequestBodyProperties(&c), nil
}

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

func updateConnectorRequestCasesWebhook(connector models.KibanaActionConnector) (kibanaactions.UpdateConnectorRequestBodyProperties, error) {
	prefixError := "failed to compose update connector request for CasesWebhook"

	nullableConfig := kibanaactions.NullableConfigPropertiesCasesWebhook{}
	if err := nullableConfig.UnmarshalJSON([]byte(connector.ConfigJSON)); err != nil {
		return kibanaactions.UpdateConnectorRequestBodyProperties{}, fmt.Errorf("%s - failed to unmarshal [config] - %w", prefixError, err)
	}

	nullableSecrets := kibanaactions.NullableSecretsPropertiesCasesWebhook{}
	if err := nullableSecrets.UnmarshalJSON([]byte(connector.SecretsJSON)); err != nil {
		return kibanaactions.UpdateConnectorRequestBodyProperties{}, fmt.Errorf("%s - failed to unmarshal [secrets]: %w", prefixError, err)
	}

	c := kibanaactions.UpdateConnectorRequestCasesWebhook{
		Name:    connector.Name,
		Config:  *nullableConfig.Get(),
		Secrets: nullableSecrets.Get(),
	}

	return kibanaactions.UpdateConnectorRequestCasesWebhookAsUpdateConnectorRequestBodyProperties(&c), nil
}

func updateConnectorRequestIndex(connector models.KibanaActionConnector) (kibanaactions.UpdateConnectorRequestBodyProperties, error) {
	prefixError := "failed to compose update connector request for Index"

	config := kibanaactions.NullableConfigPropertiesIndex{}
	if err := config.UnmarshalJSON([]byte(connector.ConfigJSON)); err != nil {
		return kibanaactions.UpdateConnectorRequestBodyProperties{}, fmt.Errorf("%s - failed to unmarshal [config] - %w", prefixError, err)
	}

	c := kibanaactions.UpdateConnectorRequestIndex{
		Name:   connector.Name,
		Config: *config.Get(),
	}

	return kibanaactions.UpdateConnectorRequestIndexAsUpdateConnectorRequestBodyProperties(&c), nil
}

func updateConnectorRequestJira(connector models.KibanaActionConnector) (kibanaactions.UpdateConnectorRequestBodyProperties, error) {
	prefixError := "failed to compose update connector request for Jira"

	config := kibanaactions.NullableConfigPropertiesJira{}
	if err := config.UnmarshalJSON([]byte(connector.ConfigJSON)); err != nil {
		return kibanaactions.UpdateConnectorRequestBodyProperties{}, fmt.Errorf("%s - failed to unmarshal [config] - %w", prefixError, err)
	}

	secrets := kibanaactions.NullableSecretsPropertiesJira{}
	if err := secrets.UnmarshalJSON([]byte(connector.SecretsJSON)); err != nil {
		return kibanaactions.UpdateConnectorRequestBodyProperties{}, fmt.Errorf("%s - failed to unmarshal [secrets]: %w", prefixError, err)
	}

	c := kibanaactions.UpdateConnectorRequestJira{
		Name:    connector.Name,
		Config:  *config.Get(),
		Secrets: *secrets.Get(),
	}

	return kibanaactions.UpdateConnectorRequestJiraAsUpdateConnectorRequestBodyProperties(&c), nil
}

func updateConnectorRequestOpsgenie(connector models.KibanaActionConnector) (kibanaactions.UpdateConnectorRequestBodyProperties, error) {
	prefixError := "failed to compose update connector request for Opsgenie"

	config := kibanaactions.NullableConfigPropertiesOpsgenie{}
	if err := config.UnmarshalJSON([]byte(connector.ConfigJSON)); err != nil {
		return kibanaactions.UpdateConnectorRequestBodyProperties{}, fmt.Errorf("%s - failed to unmarshal [config] - %w", prefixError, err)
	}

	secrets := kibanaactions.NullableSecretsPropertiesOpsgenie{}
	if err := secrets.UnmarshalJSON([]byte(connector.SecretsJSON)); err != nil {
		return kibanaactions.UpdateConnectorRequestBodyProperties{}, fmt.Errorf("%s - failed to unmarshal [secrets]: %w", prefixError, err)
	}

	c := kibanaactions.UpdateConnectorRequestOpsgenie{
		Name:    connector.Name,
		Config:  *config.Get(),
		Secrets: *secrets.Get(),
	}

	return kibanaactions.UpdateConnectorRequestOpsgenieAsUpdateConnectorRequestBodyProperties(&c), nil
}

func updateConnectorRequestResilient(connector models.KibanaActionConnector) (kibanaactions.UpdateConnectorRequestBodyProperties, error) {
	prefixError := "failed to compose update connector request for Resilient"

	config := kibanaactions.NullableConfigPropertiesResilient{}
	if err := config.UnmarshalJSON([]byte(connector.ConfigJSON)); err != nil {
		return kibanaactions.UpdateConnectorRequestBodyProperties{}, fmt.Errorf("%s - failed to unmarshal [config] - %w", prefixError, err)
	}

	secrets := kibanaactions.NullableSecretsPropertiesResilient{}
	if err := secrets.UnmarshalJSON([]byte(connector.SecretsJSON)); err != nil {
		return kibanaactions.UpdateConnectorRequestBodyProperties{}, fmt.Errorf("%s - failed to unmarshal [secrets]: %w", prefixError, err)
	}

	c := kibanaactions.UpdateConnectorRequestResilient{
		Name:    connector.Name,
		Config:  *config.Get(),
		Secrets: *secrets.Get(),
	}

	return kibanaactions.UpdateConnectorRequestResilientAsUpdateConnectorRequestBodyProperties(&c), nil
}

func updateConnectorRequestServicenow(connector models.KibanaActionConnector) (kibanaactions.UpdateConnectorRequestBodyProperties, error) {
	prefixError := "failed to compose update connector request for Servicenow"

	config := kibanaactions.NullableConfigPropertiesServicenow{}
	if err := config.UnmarshalJSON([]byte(connector.ConfigJSON)); err != nil {
		return kibanaactions.UpdateConnectorRequestBodyProperties{}, fmt.Errorf("%s - failed to unmarshal [config] - %w", prefixError, err)
	}

	secrets := kibanaactions.NullableSecretsPropertiesServicenow{}
	if err := secrets.UnmarshalJSON([]byte(connector.SecretsJSON)); err != nil {
		return kibanaactions.UpdateConnectorRequestBodyProperties{}, fmt.Errorf("%s - failed to unmarshal [secrets]: %w", prefixError, err)
	}

	c := kibanaactions.UpdateConnectorRequestServicenow{
		Name:    connector.Name,
		Config:  *config.Get(),
		Secrets: *secrets.Get(),
	}

	return kibanaactions.UpdateConnectorRequestServicenowAsUpdateConnectorRequestBodyProperties(&c), nil
}

func updateConnectorRequestServicenowItom(connector models.KibanaActionConnector) (kibanaactions.UpdateConnectorRequestBodyProperties, error) {
	prefixError := "failed to compose update connector request for ServicenowItom"

	config := kibanaactions.NullableConfigPropertiesServicenowItom{}
	if err := config.UnmarshalJSON([]byte(connector.ConfigJSON)); err != nil {
		return kibanaactions.UpdateConnectorRequestBodyProperties{}, fmt.Errorf("%s - failed to unmarshal [config] - %w", prefixError, err)
	}

	secrets := kibanaactions.NullableSecretsPropertiesServicenow{}
	if err := secrets.UnmarshalJSON([]byte(connector.SecretsJSON)); err != nil {
		return kibanaactions.UpdateConnectorRequestBodyProperties{}, fmt.Errorf("%s - failed to unmarshal [secrets]: %w", prefixError, err)
	}

	c := kibanaactions.UpdateConnectorRequestServicenowItom{
		Name:    connector.Name,
		Config:  *config.Get(),
		Secrets: *secrets.Get(),
	}

	return kibanaactions.UpdateConnectorRequestServicenowItomAsUpdateConnectorRequestBodyProperties(&c), nil
}

func updateConnectorRequestServerLog(connector models.KibanaActionConnector) (kibanaactions.UpdateConnectorRequestBodyProperties, error) {
	c := kibanaactions.UpdateConnectorRequestServerlog{
		Name: connector.Name,
	}

	return kibanaactions.UpdateConnectorRequestServerlogAsUpdateConnectorRequestBodyProperties(&c), nil
}

func updateConnectorRequestSwimlane(connector models.KibanaActionConnector) (kibanaactions.UpdateConnectorRequestBodyProperties, error) {
	prefixError := "failed to compose update connector request for Swimlane"

	config := kibanaactions.NullableConfigPropertiesSwimlane{}
	if err := config.UnmarshalJSON([]byte(connector.ConfigJSON)); err != nil {
		return kibanaactions.UpdateConnectorRequestBodyProperties{}, fmt.Errorf("%s - failed to unmarshal [config] - %w", prefixError, err)
	}

	secrets := kibanaactions.NullableSecretsPropertiesSwimlane{}
	if err := secrets.UnmarshalJSON([]byte(connector.SecretsJSON)); err != nil {
		return kibanaactions.UpdateConnectorRequestBodyProperties{}, fmt.Errorf("%s - failed to unmarshal [secrets]: %w", prefixError, err)
	}

	c := kibanaactions.UpdateConnectorRequestSwimlane{
		Name:    connector.Name,
		Config:  *config.Get(),
		Secrets: *secrets.Get(),
	}

	return kibanaactions.UpdateConnectorRequestSwimlaneAsUpdateConnectorRequestBodyProperties(&c), nil
}

func actionConnectorToModel(spaceID string, properties *kibanaactions.ConnectorResponseProperties) (*models.KibanaActionConnector, error) {
	instance := properties.GetActualInstance()
	commonProps, ok := instance.(connectorCommon)
	if !ok {
		return nil, fmt.Errorf("failed parse common connector properties")
	}
	name, ok := commonProps.GetNameOk()
	if !ok {
		return nil, fmt.Errorf("failed parse connector name")
	}
	typeId, ok := commonProps.GetNameOk()
	if !ok {
		return nil, fmt.Errorf("failed parse connector type id")
	}
	id, ok := commonProps.GetIdOk()
	if !ok {
		return nil, fmt.Errorf("failed parse connector id")
	}
	connector := models.KibanaActionConnector{
		ConnectorID:      *id,
		SpaceID:          spaceID,
		Name:             *name,
		ConnectorTypeID:  *typeId,
		IsDeprecated:     commonProps.GetIsDeprecated(),
		IsMissingSecrets: commonProps.GetIsMissingSecrets(),
		IsPreconfigured:  commonProps.GetIsPreconfigured(),
	}
	return &connector, nil
}

type connectorCommon interface {
	GetIdOk() (*string, bool)
	GetConnectorTypeIdOk() (*string, bool)
	GetNameOk() (*string, bool)
	GetIsDeprecated() bool
	GetIsMissingSecrets() bool
	GetIsPreconfigured() bool
}
