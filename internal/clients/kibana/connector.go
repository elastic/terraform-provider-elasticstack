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
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-log/tflog"
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

func GetConnector(ctx context.Context, apiClient *clients.ApiClient, connectorID, spaceID string) (*models.KibanaActionConnector, diag.Diagnostics) {
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

func SearchConnectors(ctx context.Context, apiClient *clients.ApiClient, connectorName, spaceID, connectorTypeID string) ([]*models.KibanaActionConnector, diag.Diagnostics) {
	client, err := apiClient.GetKibanaConnectorsClient(ctx)
	if err != nil {
		return nil, diag.FromErr(err)
	}

	httpResp, err := client.GetConnectors(ctx, spaceID)

	if err != nil {
		return nil, diag.Errorf("unable to get connectors: [%v]", err)
	}

	defer httpResp.Body.Close()

	resp, err := connectors.ParseGetConnectorsResponse(httpResp)
	if err != nil {
		return nil, diag.Errorf("unable to parse connectors get response: [%v]", err)
	}

	if resp.JSON401 != nil {
		return nil, diag.Errorf("%s: %s", *resp.JSON401.Error, *resp.JSON401.Message)
	}

	if resp.JSON200 == nil {
		return nil, diag.Errorf("%s: %s", resp.Status(), string(resp.Body))
	}

	foundConnectors := []*models.KibanaActionConnector{}
	for _, connector := range *resp.JSON200 {
		if connector.Name != connectorName {
			continue
		}

		if connectorTypeID != "" && string(connector.ConnectorTypeId) != connectorTypeID {
			continue
		}

		//this marshaling and unmarshaling business allows us to create a type with unexported fields.
		bytes, err := json.Marshal(connector)
		if err != nil {
			return nil, diag.Errorf("cannot marshal connector: %v", err)
		}

		var respProps connectors.ConnectorResponseProperties
		err = json.Unmarshal(bytes, &respProps)
		if err != nil {
			return nil, diag.Errorf("cannot unmarshal connector: %v", err)
		}

		c, err := connectorResponseToModel(spaceID, respProps)
		if err != nil {
			return nil, diag.Errorf("unable to convert response to model: %v", err)
		}

		foundConnectors = append(foundConnectors, c)
	}
	if len(foundConnectors) == 0 {
		tflog.Debug(ctx, fmt.Sprintf("no connectors found with name [%s/%s] and type [%s]", spaceID, connectorName, connectorTypeID))
	}

	return foundConnectors, nil
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

func ConnectorConfigWithDefaults(connectorTypeID, plan, backend, state string) (string, error) {
	switch connectors.ConnectorTypes(connectorTypeID) {

	case connectors.ConnectorTypesDotCasesWebhook:
		return connectorConfigWithDefaultsCasesWebhook(plan)

	case connectors.ConnectorTypesDotEmail:
		return connectorConfigWithDefaultsEmail(plan)

	case connectors.ConnectorTypesDotGemini:
		return connectorConfigWithDefaultsGemini(plan)

	case connectors.ConnectorTypesDotIndex:
		return connectorConfigWithDefaultsIndex(plan)

	case connectors.ConnectorTypesDotJira:
		return connectorConfigWithDefaultsJira(plan)

	case connectors.ConnectorTypesDotOpsgenie:
		return connectorConfigWithDefaultsOpsgenie(plan)

	case connectors.ConnectorTypesDotPagerduty:
		return connectorConfigWithDefaultsPagerduty(plan)

	case connectors.ConnectorTypesDotResilient:
		return connectorConfigWithDefaultsResilient(plan)

	case connectors.ConnectorTypesDotServicenow:
		return connectorConfigWithDefaultsServicenow(plan, backend)

	case connectors.ConnectorTypesDotServicenowItom:
		return connectorConfigWithDefaultsServicenowItom(plan)

	case connectors.ConnectorTypesDotServicenowSir:
		return connectorConfigWithDefaultsServicenowSir(plan, backend)

	case connectors.ConnectorTypesDotSwimlane:
		return connectorConfigWithDefaultsSwimlane(plan)

	case connectors.ConnectorTypesDotTines:
		return connectorConfigWithDefaultsTines(plan)

	case connectors.ConnectorTypesDotWebhook:
		return connectorConfigWithDefaultsWebhook(plan)

	case connectors.ConnectorTypesDotXmatters:
		return connectorConfigWithDefaultsXmatters(plan)
	}
	return plan, nil
}

// User can omit optonal fields in config JSON.
// The func adds empty optional fields to the diff.
// Otherwise plan command shows omitted fields as the diff,
// because backend returns all fields.
func connectorConfigWithDefaults[T any](plan string) (string, error) {
	var config T
	if err := json.Unmarshal([]byte(plan), &config); err != nil {
		return "", err
	}
	customJSON, err := json.Marshal(config)
	if err != nil {
		return "", err
	}
	return string(customJSON), nil
}

func connectorConfigWithDefaultsCasesWebhook(plan string) (string, error) {
	var custom connectors.ConfigPropertiesCasesWebhook
	if err := json.Unmarshal([]byte(plan), &custom); err != nil {
		return "", err
	}
	if custom.CreateIncidentMethod == nil {
		custom.CreateIncidentMethod = new(connectors.ConfigPropertiesCasesWebhookCreateIncidentMethod)
		*custom.CreateIncidentMethod = connectors.ConfigPropertiesCasesWebhookCreateIncidentMethodPost
	}
	if custom.HasAuth == nil {
		custom.HasAuth = utils.Pointer(true)
	}
	if custom.UpdateIncidentMethod == nil {
		custom.UpdateIncidentMethod = new(connectors.ConfigPropertiesCasesWebhookUpdateIncidentMethod)
		*custom.UpdateIncidentMethod = connectors.ConfigPropertiesCasesWebhookUpdateIncidentMethodPut
	}
	customJSON, err := json.Marshal(custom)
	if err != nil {
		return "", err
	}
	return string(customJSON), nil
}

func connectorConfigWithDefaultsEmail(plan string) (string, error) {
	var custom connectors.ConfigPropertiesEmail
	if err := json.Unmarshal([]byte(plan), &custom); err != nil {
		return "", err
	}
	if custom.HasAuth == nil {
		custom.HasAuth = utils.Pointer(true)
	}
	if custom.Service == nil {
		custom.Service = new(string)
		*custom.Service = "other"
	}
	customJSON, err := json.Marshal(custom)
	if err != nil {
		return "", err
	}
	return string(customJSON), nil
}

func connectorConfigWithDefaultsGemini(plan string) (string, error) {
	return plan, nil
}

func connectorConfigWithDefaultsIndex(plan string) (string, error) {
	var custom connectors.ConfigPropertiesIndex
	if err := json.Unmarshal([]byte(plan), &custom); err != nil {
		return "", err
	}
	if custom.Refresh == nil {
		custom.Refresh = utils.Pointer(false)
	}
	customJSON, err := json.Marshal(custom)
	if err != nil {
		return "", err
	}
	return string(customJSON), nil
}

func connectorConfigWithDefaultsJira(plan string) (string, error) {
	return connectorConfigWithDefaults[connectors.ConfigPropertiesJira](plan)
}

func connectorConfigWithDefaultsOpsgenie(plan string) (string, error) {
	return plan, nil
}

func connectorConfigWithDefaultsPagerduty(plan string) (string, error) {
	return connectorConfigWithDefaults[connectors.ConfigPropertiesPagerduty](plan)
}

func connectorConfigWithDefaultsResilient(plan string) (string, error) {
	return plan, nil
}

func connectorConfigWithDefaultsServicenow(plan, backend string) (string, error) {
	var planConfig connectors.ConfigPropertiesServicenow
	if err := json.Unmarshal([]byte(plan), &planConfig); err != nil {
		return "", err
	}
	var backendConfig connectors.ConfigPropertiesServicenow
	if err := json.Unmarshal([]byte(backend), &backendConfig); err != nil {
		return "", err
	}
	if planConfig.IsOAuth == nil && backendConfig.IsOAuth != nil && !*backendConfig.IsOAuth {
		planConfig.IsOAuth = utils.Pointer(false)
	}
	if planConfig.UsesTableApi == nil {
		planConfig.UsesTableApi = utils.Pointer(true)
	}
	customJSON, err := json.Marshal(planConfig)
	if err != nil {
		return "", err
	}
	return string(customJSON), nil
}

func connectorConfigWithDefaultsServicenowItom(plan string) (string, error) {
	var custom connectors.ConfigPropertiesServicenowItom
	if err := json.Unmarshal([]byte(plan), &custom); err != nil {
		return "", err
	}
	if custom.IsOAuth == nil {
		custom.IsOAuth = utils.Pointer(false)
	}
	customJSON, err := json.Marshal(custom)
	if err != nil {
		return "", err
	}
	return string(customJSON), nil
}

func connectorConfigWithDefaultsServicenowSir(plan, backend string) (string, error) {
	return connectorConfigWithDefaultsServicenow(plan, backend)
}

func connectorConfigWithDefaultsSwimlane(plan string) (string, error) {
	var custom connectors.ConfigPropertiesSwimlane
	if err := json.Unmarshal([]byte(plan), &custom); err != nil {
		return "", err
	}
	if custom.Mappings == nil {
		custom.Mappings = &connectors.ConfigPropertiesSwimlaneMappings{}
	}
	customJSON, err := json.Marshal(custom)
	if err != nil {
		return "", err
	}
	return string(customJSON), nil
}

func connectorConfigWithDefaultsTines(plan string) (string, error) {
	return plan, nil
}

func connectorConfigWithDefaultsWebhook(plan string) (string, error) {
	return plan, nil
}

func connectorConfigWithDefaultsXmatters(plan string) (string, error) {
	var custom connectors.ConfigPropertiesXmatters
	if err := json.Unmarshal([]byte(plan), &custom); err != nil {
		return "", err
	}
	if custom.UsesBasic == nil {
		custom.UsesBasic = utils.Pointer(true)
	}
	customJSON, err := json.Marshal(custom)
	if err != nil {
		return "", err
	}
	return string(customJSON), nil
}

func createConnectorRequestBody(connector models.KibanaActionConnector) (io.Reader, error) {
	switch connectors.ConnectorTypes(connector.ConnectorTypeID) {

	case connectors.ConnectorTypesDotCasesWebhook:
		return createConnectorRequestCasesWebhook(connector)

	case connectors.ConnectorTypesDotEmail:
		return createConnectorRequestEmail(connector)

	case connectors.ConnectorTypesDotGemini:
		return createConnectorRequestGemini(connector)

	case connectors.ConnectorTypesDotIndex:
		return createConnectorRequestIndex(connector)

	case connectors.ConnectorTypesDotJira:
		return createConnectorRequestJira(connector)

	case connectors.ConnectorTypesDotOpsgenie:
		return createConnectorRequestOpsgenie(connector)

	case connectors.ConnectorTypesDotPagerduty:
		return createConnectorRequestPagerduty(connector)

	case connectors.ConnectorTypesDotResilient:
		return createConnectorRequestResilient(connector)

	case connectors.ConnectorTypesDotServicenow:
		return createConnectorRequestServicenow(connector)

	case connectors.ConnectorTypesDotServicenowItom:
		return createConnectorRequestServicenowItom(connector)

	case connectors.ConnectorTypesDotServicenowSir:
		return createConnectorRequestServicenowSir(connector)

	case connectors.ConnectorTypesDotServerLog:
		return createConnectorRequestServerLog(connector)

	case connectors.ConnectorTypesDotSlack:
		return createConnectorRequestSlack(connector)

	case connectors.ConnectorTypesDotSlackApi:
		return createConnectorRequestSlackApi(connector)

	case connectors.ConnectorTypesDotSwimlane:
		return createConnectorRequestSwimlane(connector)

	case connectors.ConnectorTypesDotTeams:
		return createConnectorRequestTeams(connector)

	case connectors.ConnectorTypesDotTines:
		return createConnectorRequestTines(connector)

	case connectors.ConnectorTypesDotWebhook:
		return createConnectorRequestWebhook(connector)

	case connectors.ConnectorTypesDotXmatters:
		return createConnectorRequestXmatters(connector)
	}

	return nil, fmt.Errorf("unknown connector type [%s]", connector.ConnectorTypeID)
}

func updateConnectorRequestBody(connector models.KibanaActionConnector) (io.Reader, error) {
	switch connectors.ConnectorTypes(connector.ConnectorTypeID) {

	case connectors.ConnectorTypesDotCasesWebhook:
		return updateConnectorRequestCasesWebhook(connector)

	case connectors.ConnectorTypesDotEmail:
		return updateConnectorRequestEmail(connector)

	case connectors.ConnectorTypesDotGemini:
		return updateConnectorRequestGemini(connector)

	case connectors.ConnectorTypesDotIndex:
		return updateConnectorRequestIndex(connector)

	case connectors.ConnectorTypesDotJira:
		return updateConnectorRequestJira(connector)

	case connectors.ConnectorTypesDotOpsgenie:
		return updateConnectorRequestOpsgenie(connector)

	case connectors.ConnectorTypesDotPagerduty:
		return updateConnectorRequestPagerduty(connector)

	case connectors.ConnectorTypesDotResilient:
		return updateConnectorRequestResilient(connector)

	case connectors.ConnectorTypesDotServicenow:
		return updateConnectorRequestServicenow(connector)

	case connectors.ConnectorTypesDotServicenowItom:
		return updateConnectorRequestServicenowItom(connector)

	case connectors.ConnectorTypesDotServicenowSir:
		return updateConnectorRequestServicenowSir(connector)

	case connectors.ConnectorTypesDotServerLog:
		return updateConnectorRequestServerlog(connector)

	case connectors.ConnectorTypesDotSlack:
		return updateConnectorRequestSlack(connector)

	case connectors.ConnectorTypesDotSlackApi:
		return updateConnectorRequestSlackApi(connector)

	case connectors.ConnectorTypesDotSwimlane:
		return updateConnectorRequestSwimlane(connector)

	case connectors.ConnectorTypesDotTeams:
		return updateConnectorRequestTeams(connector)

	case connectors.ConnectorTypesDotTines:
		return updateConnectorRequestTines(connector)

	case connectors.ConnectorTypesDotWebhook:
		return updateConnectorRequestWebhook(connector)

	case connectors.ConnectorTypesDotXmatters:
		return updateConnectorRequestXmatters(connector)
	}

	return nil, fmt.Errorf("unknown connector type [%s]", connector.ConnectorTypeID)
}

func marshalConnectorRequest[C, S, R any](connector models.KibanaActionConnector, config *C, secrets *S, request *R) (io.Reader, error) {
	if config != nil && len(connector.ConfigJSON) > 0 {
		if err := json.Unmarshal([]byte(connector.ConfigJSON), config); err != nil {
			return nil, fmt.Errorf("failed to unmarshal [config] attribute: %w", err)
		}
	}

	if secrets != nil && len(connector.SecretsJSON) > 0 {
		if err := json.Unmarshal([]byte(connector.SecretsJSON), secrets); err != nil {
			return nil, fmt.Errorf("failed to unmarshal [secrets] attribute: %w", err)
		}
	}

	bt, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	return bytes.NewReader(bt), nil
}

func createConnectorRequestCasesWebhook(connector models.KibanaActionConnector) (io.Reader, error) {
	request := connectors.CreateConnectorRequestCasesWebhook{
		ConnectorTypeId: connectors.DotCasesWebhook,
		Name:            connector.Name,
	}

	return marshalConnectorRequest(connector, &request.Config, &request.Secrets, &request)
}

func createConnectorRequestEmail(connector models.KibanaActionConnector) (io.Reader, error) {
	request := connectors.CreateConnectorRequestEmail{
		ConnectorTypeId: connectors.CreateConnectorRequestEmailConnectorTypeIdDotEmail,
		Name:            connector.Name,
	}

	return marshalConnectorRequest(connector, &request.Config, &request.Secrets, &request)
}

func createConnectorRequestGemini(connector models.KibanaActionConnector) (io.Reader, error) {
	request := connectors.CreateConnectorRequestGemini{
		ConnectorTypeId: connectors.CreateConnectorRequestGeminiConnectorTypeIdDotGemini,
		Name:            connector.Name,
	}

	return marshalConnectorRequest(connector, &request.Config, &request.Secrets, &request)
}

func createConnectorRequestIndex(connector models.KibanaActionConnector) (io.Reader, error) {
	request := connectors.CreateConnectorRequestIndex{
		ConnectorTypeId: connectors.CreateConnectorRequestIndexConnectorTypeIdDotIndex,
		Name:            connector.Name,
	}

	return marshalConnectorRequest[connectors.ConfigPropertiesIndex, any](connector, &request.Config, nil, &request)
}

func createConnectorRequestJira(connector models.KibanaActionConnector) (io.Reader, error) {
	request := connectors.CreateConnectorRequestJira{
		ConnectorTypeId: connectors.CreateConnectorRequestJiraConnectorTypeIdDotJira,
		Name:            connector.Name,
	}

	return marshalConnectorRequest(connector, &request.Config, &request.Secrets, &request)
}

func createConnectorRequestOpsgenie(connector models.KibanaActionConnector) (io.Reader, error) {
	request := connectors.CreateConnectorRequestOpsgenie{
		ConnectorTypeId: connectors.CreateConnectorRequestOpsgenieConnectorTypeIdDotOpsgenie,
		Name:            connector.Name,
	}

	return marshalConnectorRequest(connector, &request.Config, &request.Secrets, &request)
}

func createConnectorRequestPagerduty(connector models.KibanaActionConnector) (io.Reader, error) {
	request := connectors.CreateConnectorRequestPagerduty{
		ConnectorTypeId: connectors.CreateConnectorRequestPagerdutyConnectorTypeIdDotPagerduty,
		Name:            connector.Name,
	}

	return marshalConnectorRequest(connector, &request.Config, &request.Secrets, &request)
}

func createConnectorRequestResilient(connector models.KibanaActionConnector) (io.Reader, error) {
	request := connectors.CreateConnectorRequestResilient{
		ConnectorTypeId: connectors.CreateConnectorRequestResilientConnectorTypeIdDotResilient,
		Name:            connector.Name,
	}

	return marshalConnectorRequest(connector, &request.Config, &request.Secrets, &request)
}

func createConnectorRequestServicenow(connector models.KibanaActionConnector) (io.Reader, error) {
	request := connectors.CreateConnectorRequestServicenow{
		ConnectorTypeId: connectors.CreateConnectorRequestServicenowConnectorTypeIdDotServicenow,
		Name:            connector.Name,
	}

	return marshalConnectorRequest(connector, &request.Config, &request.Secrets, &request)
}

func createConnectorRequestServicenowItom(connector models.KibanaActionConnector) (io.Reader, error) {
	request := connectors.CreateConnectorRequestServicenowItom{
		ConnectorTypeId: connectors.CreateConnectorRequestServicenowItomConnectorTypeIdDotServicenowItom,
		Name:            connector.Name,
	}

	return marshalConnectorRequest(connector, &request.Config, &request.Secrets, &request)
}

func createConnectorRequestServicenowSir(connector models.KibanaActionConnector) (io.Reader, error) {
	request := connectors.CreateConnectorRequestServicenowSir{
		ConnectorTypeId: connectors.CreateConnectorRequestServicenowSirConnectorTypeIdDotServicenowSir,
		Name:            connector.Name,
	}

	return marshalConnectorRequest(connector, &request.Config, &request.Secrets, &request)
}

func createConnectorRequestServerLog(connector models.KibanaActionConnector) (io.Reader, error) {
	request := connectors.CreateConnectorRequestServerlog{
		ConnectorTypeId: connectors.CreateConnectorRequestServerlogConnectorTypeIdDotServerLog,
		Name:            connector.Name,
	}

	return marshalConnectorRequest[any, any](connector, nil, nil, &request)
}

func createConnectorRequestSlack(connector models.KibanaActionConnector) (io.Reader, error) {
	request := connectors.CreateConnectorRequestSlack{
		ConnectorTypeId: connectors.CreateConnectorRequestSlackConnectorTypeIdDotSlack,
		Name:            connector.Name,
	}

	return marshalConnectorRequest[any](connector, nil, &request.Secrets, &request)
}

func createConnectorRequestSlackApi(connector models.KibanaActionConnector) (io.Reader, error) {
	request := connectors.CreateConnectorRequestSlackApi{
		ConnectorTypeId: connectors.CreateConnectorRequestSlackApiConnectorTypeIdDotSlackApi,
		Name:            connector.Name,
	}

	return marshalConnectorRequest[any](connector, nil, &request.Secrets, &request)
}

func createConnectorRequestSwimlane(connector models.KibanaActionConnector) (io.Reader, error) {
	request := connectors.CreateConnectorRequestSwimlane{
		ConnectorTypeId: connectors.CreateConnectorRequestSwimlaneConnectorTypeIdDotSwimlane,
		Name:            connector.Name,
	}

	return marshalConnectorRequest(connector, &request.Config, &request.Secrets, &request)
}

func createConnectorRequestTeams(connector models.KibanaActionConnector) (io.Reader, error) {
	request := connectors.CreateConnectorRequestTeams{
		ConnectorTypeId: connectors.CreateConnectorRequestTeamsConnectorTypeIdDotTeams,
		Name:            connector.Name,
	}

	return marshalConnectorRequest[any](connector, nil, &request.Secrets, &request)
}

func createConnectorRequestTines(connector models.KibanaActionConnector) (io.Reader, error) {
	request := connectors.CreateConnectorRequestTines{
		ConnectorTypeId: connectors.CreateConnectorRequestTinesConnectorTypeIdDotTines,
		Name:            connector.Name,
	}

	return marshalConnectorRequest(connector, &request.Config, &request.Secrets, &request)
}

func createConnectorRequestWebhook(connector models.KibanaActionConnector) (io.Reader, error) {
	request := connectors.CreateConnectorRequestWebhook{
		ConnectorTypeId: connectors.CreateConnectorRequestWebhookConnectorTypeIdDotWebhook,
		Name:            connector.Name,
	}

	return marshalConnectorRequest(connector, &request.Config, &request.Secrets, &request)
}

func createConnectorRequestXmatters(connector models.KibanaActionConnector) (io.Reader, error) {
	request := connectors.CreateConnectorRequestXmatters{
		ConnectorTypeId: connectors.CreateConnectorRequestXmattersConnectorTypeIdDotXmatters,
		Name:            connector.Name,
	}

	return marshalConnectorRequest(connector, &request.Config, &request.Secrets, &request)
}

func updateConnectorRequestCasesWebhook(connector models.KibanaActionConnector) (io.Reader, error) {
	request := connectors.UpdateConnectorRequestCasesWebhook{
		Name: connector.Name,
	}

	return marshalConnectorRequest(connector, &request.Config, &request.Secrets, &request)
}

func updateConnectorRequestEmail(connector models.KibanaActionConnector) (io.Reader, error) {
	request := connectors.UpdateConnectorRequestEmail{
		Name: connector.Name,
	}

	return marshalConnectorRequest(connector, &request.Config, &request.Secrets, &request)
}

func updateConnectorRequestGemini(connector models.KibanaActionConnector) (io.Reader, error) {
	request := connectors.UpdateConnectorRequestGemini{
		Name: connector.Name,
	}

	return marshalConnectorRequest(connector, &request.Config, &request.Secrets, &request)
}

func updateConnectorRequestIndex(connector models.KibanaActionConnector) (io.Reader, error) {
	request := connectors.UpdateConnectorRequestIndex{
		Name: connector.Name,
	}

	return marshalConnectorRequest[connectors.ConfigPropertiesIndex, any](connector, &request.Config, nil, &request)
}

func updateConnectorRequestJira(connector models.KibanaActionConnector) (io.Reader, error) {
	request := connectors.UpdateConnectorRequestJira{
		Name: connector.Name,
	}

	return marshalConnectorRequest(connector, &request.Config, &request.Secrets, &request)
}

func updateConnectorRequestOpsgenie(connector models.KibanaActionConnector) (io.Reader, error) {
	request := connectors.UpdateConnectorRequestOpsgenie{
		Name: connector.Name,
	}

	return marshalConnectorRequest(connector, &request.Config, &request.Secrets, &request)
}

func updateConnectorRequestPagerduty(connector models.KibanaActionConnector) (io.Reader, error) {
	request := connectors.UpdateConnectorRequestPagerduty{
		Name: connector.Name,
	}

	return marshalConnectorRequest(connector, &request.Config, &request.Secrets, &request)
}

func updateConnectorRequestResilient(connector models.KibanaActionConnector) (io.Reader, error) {
	request := connectors.UpdateConnectorRequestResilient{
		Name: connector.Name,
	}

	return marshalConnectorRequest(connector, &request.Config, &request.Secrets, &request)
}

func updateConnectorRequestServicenow(connector models.KibanaActionConnector) (io.Reader, error) {
	request := connectors.UpdateConnectorRequestServicenow{
		Name: connector.Name,
	}

	return marshalConnectorRequest(connector, &request.Config, &request.Secrets, &request)
}

func updateConnectorRequestServicenowItom(connector models.KibanaActionConnector) (io.Reader, error) {
	request := connectors.UpdateConnectorRequestServicenowItom{
		Name: connector.Name,
	}

	return marshalConnectorRequest(connector, &request.Config, &request.Secrets, &request)
}

func updateConnectorRequestServicenowSir(connector models.KibanaActionConnector) (io.Reader, error) {
	request := connectors.UpdateConnectorRequestServicenowSir{
		Name: connector.Name,
	}

	return marshalConnectorRequest(connector, &request.Config, &request.Secrets, &request)
}

func updateConnectorRequestServerlog(connector models.KibanaActionConnector) (io.Reader, error) {
	request := connectors.UpdateConnectorRequestServerlog{
		Name: connector.Name,
	}

	return marshalConnectorRequest[any, any](connector, nil, nil, &request)
}

func updateConnectorRequestSlack(connector models.KibanaActionConnector) (io.Reader, error) {
	request := connectors.UpdateConnectorRequestSlack{
		Name: connector.Name,
	}

	return marshalConnectorRequest[any](connector, nil, &request.Secrets, &request)
}

func updateConnectorRequestSlackApi(connector models.KibanaActionConnector) (io.Reader, error) {
	request := connectors.UpdateConnectorRequestSlackApi{
		Name: connector.Name,
	}

	return marshalConnectorRequest[any](connector, nil, &request.Secrets, &request)
}

func updateConnectorRequestSwimlane(connector models.KibanaActionConnector) (io.Reader, error) {
	request := connectors.UpdateConnectorRequestSwimlane{
		Name: connector.Name,
	}

	return marshalConnectorRequest(connector, &request.Config, &request.Secrets, &request)
}

func updateConnectorRequestTeams(connector models.KibanaActionConnector) (io.Reader, error) {
	request := connectors.UpdateConnectorRequestTeams{
		Name: connector.Name,
	}

	return marshalConnectorRequest[any](connector, nil, &request.Secrets, &request)
}

func updateConnectorRequestTines(connector models.KibanaActionConnector) (io.Reader, error) {
	request := connectors.UpdateConnectorRequestTines{
		Name: connector.Name,
	}

	return marshalConnectorRequest(connector, &request.Config, &request.Secrets, &request)
}

func updateConnectorRequestWebhook(connector models.KibanaActionConnector) (io.Reader, error) {
	request := connectors.UpdateConnectorRequestWebhook{
		Name: connector.Name,
	}

	return marshalConnectorRequest(connector, &request.Config, &request.Secrets, &request)
}

func updateConnectorRequestXmatters(connector models.KibanaActionConnector) (io.Reader, error) {
	request := connectors.UpdateConnectorRequestXmatters{
		Name: connector.Name,
	}

	return marshalConnectorRequest(connector, &request.Config, &request.Secrets, &request)
}

func connectorResponseToModel(spaceID string, properties connectors.ConnectorResponseProperties) (*models.KibanaActionConnector, error) {
	discriminator, err := properties.Discriminator()
	if err != nil {
		return nil, err
	}

	switch connectors.ConnectorTypes(discriminator) {

	case connectors.ConnectorTypesDotCasesWebhook:
		return connectorResponseToModelCasesWebhook(discriminator, spaceID, properties)

	case connectors.ConnectorTypesDotEmail:
		return connectorResponseToModelEmail(discriminator, spaceID, properties)

	case connectors.ConnectorTypesDotGemini:
		return connectorResponseToModelGemini(discriminator, spaceID, properties)

	case connectors.ConnectorTypesDotIndex:
		return connectorResponseToModelIndex(discriminator, spaceID, properties)

	case connectors.ConnectorTypesDotJira:
		return connectorResponseToModelJira(discriminator, spaceID, properties)

	case connectors.ConnectorTypesDotOpsgenie:
		return connectorResponseToModelOpsgenie(discriminator, spaceID, properties)

	case connectors.ConnectorTypesDotPagerduty:
		return connectorResponseToModelPagerduty(discriminator, spaceID, properties)

	case connectors.ConnectorTypesDotResilient:
		return connectorResponseToModelResilient(discriminator, spaceID, properties)

	case connectors.ConnectorTypesDotServerLog:
		return connectorResponseToModelServerlog(discriminator, spaceID, properties)

	case connectors.ConnectorTypesDotServicenow:
		return connectorResponseToModelServicenow(discriminator, spaceID, properties)

	case connectors.ConnectorTypesDotServicenowItom:
		return connectorResponseToModelServicenowItom(discriminator, spaceID, properties)

	case connectors.ConnectorTypesDotServicenowSir:
		return connectorResponseToModelServicenowSir(discriminator, spaceID, properties)

	case connectors.ConnectorTypesDotSlack:
		return connectorResponseToModelSlack(discriminator, spaceID, properties)

	case connectors.ConnectorTypesDotSlackApi:
		return connectorResponseToModelSlackApi(discriminator, spaceID, properties)

	case connectors.ConnectorTypesDotSwimlane:
		return connectorResponseToModelSwimlane(discriminator, spaceID, properties)

	case connectors.ConnectorTypesDotTeams:
		return connectorResponseToModelTeams(discriminator, spaceID, properties)

	case connectors.ConnectorTypesDotTines:
		return connectorResponseToModelTines(discriminator, spaceID, properties)

	case connectors.ConnectorTypesDotWebhook:
		return connectorResponseToModelWebhook(discriminator, spaceID, properties)

	case connectors.ConnectorTypesDotXmatters:
		return connectorResponseToModelXmatters(discriminator, spaceID, properties)
	}

	return nil, fmt.Errorf("unknown connector type [%s]", discriminator)
}

func connectorResponseToModelCasesWebhook(discriminator, spaceID string, properties connectors.ConnectorResponseProperties) (*models.KibanaActionConnector, error) {
	resp, err := properties.AsConnectorResponsePropertiesCasesWebhook()
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
		ConnectorTypeID:  discriminator,
		IsDeprecated:     isDeprecated,
		IsMissingSecrets: isMissingSecrets,
		IsPreconfigured:  bool(resp.IsPreconfigured),
		ConfigJSON:       string(config),
	}

	return &connector, nil
}

func connectorResponseToModelEmail(discriminator, spaceID string, properties connectors.ConnectorResponseProperties) (*models.KibanaActionConnector, error) {
	resp, err := properties.AsConnectorResponsePropertiesEmail()
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
		ConnectorTypeID:  discriminator,
		IsDeprecated:     isDeprecated,
		IsMissingSecrets: isMissingSecrets,
		IsPreconfigured:  bool(resp.IsPreconfigured),
		ConfigJSON:       string(config),
	}

	return &connector, nil
}

func connectorResponseToModelGemini(discriminator, spaceID string, properties connectors.ConnectorResponseProperties) (*models.KibanaActionConnector, error) {
	resp, err := properties.AsConnectorResponsePropertiesGemini()
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
		ConnectorTypeID:  discriminator,
		IsDeprecated:     isDeprecated,
		IsMissingSecrets: isMissingSecrets,
		IsPreconfigured:  bool(resp.IsPreconfigured),
		ConfigJSON:       string(config),
	}

	return &connector, nil
}

func connectorResponseToModelIndex(discriminator, spaceID string, properties connectors.ConnectorResponseProperties) (*models.KibanaActionConnector, error) {
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
		ConnectorTypeID:  discriminator,
		IsDeprecated:     isDeprecated,
		IsMissingSecrets: isMissingSecrets,
		IsPreconfigured:  bool(resp.IsPreconfigured),
		ConfigJSON:       string(config),
	}

	return &connector, nil
}

func connectorResponseToModelJira(discriminator, spaceID string, properties connectors.ConnectorResponseProperties) (*models.KibanaActionConnector, error) {
	resp, err := properties.AsConnectorResponsePropertiesJira()
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
		ConnectorTypeID:  discriminator,
		IsDeprecated:     isDeprecated,
		IsMissingSecrets: isMissingSecrets,
		IsPreconfigured:  bool(resp.IsPreconfigured),
		ConfigJSON:       string(config),
	}

	return &connector, nil
}

func connectorResponseToModelOpsgenie(discriminator, spaceID string, properties connectors.ConnectorResponseProperties) (*models.KibanaActionConnector, error) {
	resp, err := properties.AsConnectorResponsePropertiesOpsgenie()
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
		ConnectorTypeID:  discriminator,
		IsDeprecated:     isDeprecated,
		IsMissingSecrets: isMissingSecrets,
		IsPreconfigured:  bool(resp.IsPreconfigured),
		ConfigJSON:       string(config),
	}

	return &connector, nil
}

func connectorResponseToModelPagerduty(discriminator, spaceID string, properties connectors.ConnectorResponseProperties) (*models.KibanaActionConnector, error) {
	resp, err := properties.AsConnectorResponsePropertiesPagerduty()
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
		ConnectorTypeID:  discriminator,
		IsDeprecated:     isDeprecated,
		IsMissingSecrets: isMissingSecrets,
		IsPreconfigured:  bool(resp.IsPreconfigured),
		ConfigJSON:       string(config),
	}

	return &connector, nil
}

func connectorResponseToModelResilient(discriminator, spaceID string, properties connectors.ConnectorResponseProperties) (*models.KibanaActionConnector, error) {
	resp, err := properties.AsConnectorResponsePropertiesResilient()
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
		ConnectorTypeID:  discriminator,
		IsDeprecated:     isDeprecated,
		IsMissingSecrets: isMissingSecrets,
		IsPreconfigured:  bool(resp.IsPreconfigured),
		ConfigJSON:       string(config),
	}

	return &connector, nil
}

func connectorResponseToModelServerlog(discriminator, spaceID string, properties connectors.ConnectorResponseProperties) (*models.KibanaActionConnector, error) {
	resp, err := properties.AsConnectorResponsePropertiesServerlog()
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
		ConnectorTypeID:  discriminator,
		IsDeprecated:     isDeprecated,
		IsMissingSecrets: isMissingSecrets,
		IsPreconfigured:  bool(resp.IsPreconfigured),
		ConfigJSON:       string(config),
	}

	return &connector, nil
}

func connectorResponseToModelServicenow(discriminator, spaceID string, properties connectors.ConnectorResponseProperties) (*models.KibanaActionConnector, error) {
	resp, err := properties.AsConnectorResponsePropertiesServicenow()
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
		ConnectorTypeID:  discriminator,
		IsDeprecated:     isDeprecated,
		IsMissingSecrets: isMissingSecrets,
		IsPreconfigured:  bool(resp.IsPreconfigured),
		ConfigJSON:       string(config),
	}

	return &connector, nil
}

func connectorResponseToModelServicenowItom(discriminator, spaceID string, properties connectors.ConnectorResponseProperties) (*models.KibanaActionConnector, error) {
	resp, err := properties.AsConnectorResponsePropertiesServicenowItom()
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
		ConnectorTypeID:  discriminator,
		IsDeprecated:     isDeprecated,
		IsMissingSecrets: isMissingSecrets,
		IsPreconfigured:  bool(resp.IsPreconfigured),
		ConfigJSON:       string(config),
	}

	return &connector, nil
}

func connectorResponseToModelServicenowSir(discriminator, spaceID string, properties connectors.ConnectorResponseProperties) (*models.KibanaActionConnector, error) {
	resp, err := properties.AsConnectorResponsePropertiesServicenowSir()
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
		ConnectorTypeID:  discriminator,
		IsDeprecated:     isDeprecated,
		IsMissingSecrets: isMissingSecrets,
		IsPreconfigured:  bool(resp.IsPreconfigured),
		ConfigJSON:       string(config),
	}

	return &connector, nil
}

func connectorResponseToModelSlack(discriminator, spaceID string, properties connectors.ConnectorResponseProperties) (*models.KibanaActionConnector, error) {
	resp, err := properties.AsConnectorResponsePropertiesSlack()
	if err != nil {
		return nil, err
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
		ConnectorTypeID:  discriminator,
		IsDeprecated:     isDeprecated,
		IsMissingSecrets: isMissingSecrets,
		IsPreconfigured:  bool(resp.IsPreconfigured),
	}

	return &connector, nil
}

func connectorResponseToModelSlackApi(discriminator, spaceID string, properties connectors.ConnectorResponseProperties) (*models.KibanaActionConnector, error) {
	resp, err := properties.AsConnectorResponsePropertiesSlackApi()
	if err != nil {
		return nil, err
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
		ConnectorTypeID:  discriminator,
		IsDeprecated:     isDeprecated,
		IsMissingSecrets: isMissingSecrets,
		IsPreconfigured:  bool(resp.IsPreconfigured),
	}

	return &connector, nil
}

func connectorResponseToModelSwimlane(discriminator, spaceID string, properties connectors.ConnectorResponseProperties) (*models.KibanaActionConnector, error) {
	resp, err := properties.AsConnectorResponsePropertiesSwimlane()
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
		ConnectorTypeID:  discriminator,
		IsDeprecated:     isDeprecated,
		IsMissingSecrets: isMissingSecrets,
		IsPreconfigured:  bool(resp.IsPreconfigured),
		ConfigJSON:       string(config),
	}

	return &connector, nil
}

func connectorResponseToModelTeams(discriminator, spaceID string, properties connectors.ConnectorResponseProperties) (*models.KibanaActionConnector, error) {
	resp, err := properties.AsConnectorResponsePropertiesTeams()
	if err != nil {
		return nil, err
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
		ConnectorTypeID:  discriminator,
		IsDeprecated:     isDeprecated,
		IsMissingSecrets: isMissingSecrets,
		IsPreconfigured:  bool(resp.IsPreconfigured),
	}

	return &connector, nil
}

func connectorResponseToModelTines(discriminator, spaceID string, properties connectors.ConnectorResponseProperties) (*models.KibanaActionConnector, error) {
	resp, err := properties.AsConnectorResponsePropertiesTines()
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
		ConnectorTypeID:  discriminator,
		IsDeprecated:     isDeprecated,
		IsMissingSecrets: isMissingSecrets,
		IsPreconfigured:  bool(resp.IsPreconfigured),
		ConfigJSON:       string(config),
	}

	return &connector, nil
}

func connectorResponseToModelWebhook(discriminator, spaceID string, properties connectors.ConnectorResponseProperties) (*models.KibanaActionConnector, error) {
	resp, err := properties.AsConnectorResponsePropertiesWebhook()
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
		ConnectorTypeID:  discriminator,
		IsDeprecated:     isDeprecated,
		IsMissingSecrets: isMissingSecrets,
		IsPreconfigured:  bool(resp.IsPreconfigured),
		ConfigJSON:       string(config),
	}

	return &connector, nil
}

func connectorResponseToModelXmatters(discriminator, spaceID string, properties connectors.ConnectorResponseProperties) (*models.KibanaActionConnector, error) {
	resp, err := properties.AsConnectorResponsePropertiesXmatters()
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
		ConnectorTypeID:  discriminator,
		IsDeprecated:     isDeprecated,
		IsMissingSecrets: isMissingSecrets,
		IsPreconfigured:  bool(resp.IsPreconfigured),
		ConfigJSON:       string(config),
	}

	return &connector, nil
}
