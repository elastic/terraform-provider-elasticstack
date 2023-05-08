package kibana

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"

	// "github.com/google/go-cmp/cmp"

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

func ConnectorConfigWithDefaults(connectorTypeID, proposed, backend, local string) (string, error) {
	switch connectors.ConnectorTypes(connectorTypeID) {
	case connectors.ConnectorTypesDotEmail:
		return connectorEmailConfigWithDefaults(proposed)
	case connectors.ConnectorTypesDotIndex:
		return connectorIndexConfigWithDefaults(proposed)
	}
	return proposed, nil
}

func connectorEmailConfigWithDefaults(proposed string) (string, error) {
	var custom connectors.ConfigPropertiesEmail
	if err := json.Unmarshal([]byte(proposed), &custom); err != nil {
		return "", err
	}
	if custom.HasAuth == nil {
		custom.HasAuth = new(bool)
		*custom.HasAuth = true
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

func connectorIndexConfigWithDefaults(proposed string) (string, error) {
	var custom connectors.ConfigPropertiesIndex
	if err := json.Unmarshal([]byte(proposed), &custom); err != nil {
		return "", err
	}
	if custom.Refresh == nil {
		custom.Refresh = new(bool)
		*custom.Refresh = false
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

func createConnectorRequestCasesWebhook(connector models.KibanaActionConnector) (io.Reader, error) {
	prefixError := "failed to create CasesWebhook connector request"

	request := connectors.CreateConnectorRequestCasesWebhook{
		ConnectorTypeId: connectors.DotCasesWebhook,
		Name:            connector.Name,
	}

	if err := json.Unmarshal([]byte(connector.ConfigJSON), &request.Config); err != nil {
		return nil, fmt.Errorf("%s: failed to unmarshal [config] attribute: %w", prefixError, err)
	}

	if err := json.Unmarshal([]byte(connector.SecretsJSON), &request.Secrets); err != nil {
		return nil, fmt.Errorf("%s: failed to unmarshal [secrets] attribute: %w", prefixError, err)
	}

	bt, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to marshal request: %w", prefixError, err)
	}

	return bytes.NewReader(bt), nil
}

func createConnectorRequestEmail(connector models.KibanaActionConnector) (io.Reader, error) {
	prefixError := "failed to create Email connector request"

	request := connectors.CreateConnectorRequestEmail{
		ConnectorTypeId: connectors.CreateConnectorRequestEmailConnectorTypeIdDotEmail,
		Name:            connector.Name,
	}

	if err := json.Unmarshal([]byte(connector.ConfigJSON), &request.Config); err != nil {
		return nil, fmt.Errorf("%s: failed to unmarshal [config] attribute: %w", prefixError, err)
	}

	if err := json.Unmarshal([]byte(connector.SecretsJSON), &request.Secrets); err != nil {
		return nil, fmt.Errorf("%s: failed to unmarshal [secrets] attribute: %w", prefixError, err)
	}

	bt, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to marshal request: %w", prefixError, err)
	}

	return bytes.NewReader(bt), nil
}

func createConnectorRequestIndex(connector models.KibanaActionConnector) (io.Reader, error) {
	prefixError := "failed to create Index connector request"

	request := connectors.CreateConnectorRequestIndex{
		ConnectorTypeId: connectors.CreateConnectorRequestIndexConnectorTypeIdDotIndex,
		Name:            connector.Name,
	}

	if err := json.Unmarshal([]byte(connector.ConfigJSON), &request.Config); err != nil {
		return nil, fmt.Errorf("%s: failed to unmarshal [config] attribute: %w", prefixError, err)
	}

	bt, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to marshal request: %w", prefixError, err)
	}

	return bytes.NewReader(bt), nil
}

func createConnectorRequestJira(connector models.KibanaActionConnector) (io.Reader, error) {
	prefixError := "failed to create Jira connector request"

	request := connectors.CreateConnectorRequestJira{
		ConnectorTypeId: connectors.CreateConnectorRequestJiraConnectorTypeIdDotJira,
		Name:            connector.Name,
	}

	if err := json.Unmarshal([]byte(connector.ConfigJSON), &request.Config); err != nil {
		return nil, fmt.Errorf("%s: failed to unmarshal [config] attribute: %w", prefixError, err)
	}

	if err := json.Unmarshal([]byte(connector.SecretsJSON), &request.Secrets); err != nil {
		return nil, fmt.Errorf("%s: failed to unmarshal [secrets] attribute: %w", prefixError, err)
	}

	bt, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to marshal request: %w", prefixError, err)
	}

	return bytes.NewReader(bt), nil
}

func createConnectorRequestOpsgenie(connector models.KibanaActionConnector) (io.Reader, error) {
	prefixError := "failed to create Opsgenie connector request"

	request := connectors.CreateConnectorRequestOpsgenie{
		ConnectorTypeId: connectors.CreateConnectorRequestOpsgenieConnectorTypeIdDotOpsgenie,
		Name:            connector.Name,
	}

	if err := json.Unmarshal([]byte(connector.ConfigJSON), &request.Config); err != nil {
		return nil, fmt.Errorf("%s: failed to unmarshal [config] attribute: %w", prefixError, err)
	}

	if err := json.Unmarshal([]byte(connector.SecretsJSON), &request.Secrets); err != nil {
		return nil, fmt.Errorf("%s: failed to unmarshal [secrets] attribute: %w", prefixError, err)
	}

	bt, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to marshal request: %w", prefixError, err)
	}

	return bytes.NewReader(bt), nil
}

func createConnectorRequestPagerduty(connector models.KibanaActionConnector) (io.Reader, error) {
	prefixError := "failed to create Pagerduty connector request"

	request := connectors.CreateConnectorRequestPagerduty{
		ConnectorTypeId: connectors.CreateConnectorRequestPagerdutyConnectorTypeIdDotPagerduty,
		Name:            connector.Name,
	}

	if err := json.Unmarshal([]byte(connector.ConfigJSON), &request.Config); err != nil {
		return nil, fmt.Errorf("%s: failed to unmarshal [config] attribute: %w", prefixError, err)
	}

	if err := json.Unmarshal([]byte(connector.SecretsJSON), &request.Secrets); err != nil {
		return nil, fmt.Errorf("%s: failed to unmarshal [secrets] attribute: %w", prefixError, err)
	}

	bt, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to marshal request: %w", prefixError, err)
	}

	return bytes.NewReader(bt), nil
}

func createConnectorRequestResilient(connector models.KibanaActionConnector) (io.Reader, error) {
	prefixError := "failed to create Resilient connector request"

	request := connectors.CreateConnectorRequestResilient{
		ConnectorTypeId: connectors.CreateConnectorRequestResilientConnectorTypeIdDotResilient,
		Name:            connector.Name,
	}

	if err := json.Unmarshal([]byte(connector.ConfigJSON), &request.Config); err != nil {
		return nil, fmt.Errorf("%s: failed to unmarshal [config] attribute: %w", prefixError, err)
	}

	if err := json.Unmarshal([]byte(connector.SecretsJSON), &request.Secrets); err != nil {
		return nil, fmt.Errorf("%s: failed to unmarshal [secrets] attribute: %w", prefixError, err)
	}

	bt, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to marshal request: %w", prefixError, err)
	}

	return bytes.NewReader(bt), nil
}

func createConnectorRequestServicenow(connector models.KibanaActionConnector) (io.Reader, error) {
	prefixError := "failed to create Servicenow connector request"

	request := connectors.CreateConnectorRequestServicenow{
		ConnectorTypeId: connectors.CreateConnectorRequestServicenowConnectorTypeIdDotServicenow,
		Name:            connector.Name,
	}

	if err := json.Unmarshal([]byte(connector.ConfigJSON), &request.Config); err != nil {
		return nil, fmt.Errorf("%s: failed to unmarshal [config] attribute: %w", prefixError, err)
	}

	if err := json.Unmarshal([]byte(connector.SecretsJSON), &request.Secrets); err != nil {
		return nil, fmt.Errorf("%s: failed to unmarshal [secrets] attribute: %w", prefixError, err)
	}

	bt, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to marshal request: %w", prefixError, err)
	}

	return bytes.NewReader(bt), nil
}

func createConnectorRequestServicenowItom(connector models.KibanaActionConnector) (io.Reader, error) {
	prefixError := "failed to create ServicenowItom connector request"

	request := connectors.CreateConnectorRequestServicenowItom{
		ConnectorTypeId: connectors.CreateConnectorRequestServicenowItomConnectorTypeIdDotServicenowItom,
		Name:            connector.Name,
	}

	if err := json.Unmarshal([]byte(connector.ConfigJSON), &request.Config); err != nil {
		return nil, fmt.Errorf("%s: failed to unmarshal [config] attribute: %w", prefixError, err)
	}

	if err := json.Unmarshal([]byte(connector.SecretsJSON), &request.Secrets); err != nil {
		return nil, fmt.Errorf("%s: failed to unmarshal [secrets] attribute: %w", prefixError, err)
	}

	bt, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to marshal request: %w", prefixError, err)
	}

	return bytes.NewReader(bt), nil
}

func createConnectorRequestServicenowSir(connector models.KibanaActionConnector) (io.Reader, error) {
	prefixError := "failed to create ServicenowSir connector request"

	request := connectors.CreateConnectorRequestServicenowSir{
		ConnectorTypeId: connectors.CreateConnectorRequestServicenowSirConnectorTypeIdDotServicenowSir,
		Name:            connector.Name,
	}

	if err := json.Unmarshal([]byte(connector.ConfigJSON), &request.Config); err != nil {
		return nil, fmt.Errorf("%s: failed to unmarshal [config] attribute: %w", prefixError, err)
	}

	if err := json.Unmarshal([]byte(connector.SecretsJSON), &request.Secrets); err != nil {
		return nil, fmt.Errorf("%s: failed to unmarshal [secrets] attribute: %w", prefixError, err)
	}

	bt, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to marshal request: %w", prefixError, err)
	}

	return bytes.NewReader(bt), nil
}

func createConnectorRequestServerLog(connector models.KibanaActionConnector) (io.Reader, error) {
	prefixError := "failed to create Serverlog connector request"

	request := connectors.CreateConnectorRequestServerlog{
		ConnectorTypeId: connectors.CreateConnectorRequestServerlogConnectorTypeIdDotServerLog,
		Name:            connector.Name,
	}

	bt, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to marshal request: %w", prefixError, err)
	}

	return bytes.NewReader(bt), nil
}

func createConnectorRequestSlack(connector models.KibanaActionConnector) (io.Reader, error) {
	prefixError := "failed to create Slack connector request"

	request := connectors.CreateConnectorRequestSlack{
		ConnectorTypeId: connectors.CreateConnectorRequestSlackConnectorTypeIdDotSlack,
		Name:            connector.Name,
	}

	if err := json.Unmarshal([]byte(connector.SecretsJSON), &request.Secrets); err != nil {
		return nil, fmt.Errorf("%s: failed to unmarshal [secrets] attribute: %w", prefixError, err)
	}

	bt, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to marshal request: %w", prefixError, err)
	}

	return bytes.NewReader(bt), nil
}

func createConnectorRequestSwimlane(connector models.KibanaActionConnector) (io.Reader, error) {
	prefixError := "failed to create Swimlane connector request"

	request := connectors.CreateConnectorRequestSwimlane{
		ConnectorTypeId: connectors.CreateConnectorRequestSwimlaneConnectorTypeIdDotSwimlane,
		Name:            connector.Name,
	}

	if err := json.Unmarshal([]byte(connector.ConfigJSON), &request.Config); err != nil {
		return nil, fmt.Errorf("%s: failed to unmarshal [config] attribute: %w", prefixError, err)
	}

	if err := json.Unmarshal([]byte(connector.SecretsJSON), &request.Secrets); err != nil {
		return nil, fmt.Errorf("%s: failed to unmarshal [secrets] attribute: %w", prefixError, err)
	}

	bt, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to marshal request: %w", prefixError, err)
	}

	return bytes.NewReader(bt), nil
}

func createConnectorRequestTeams(connector models.KibanaActionConnector) (io.Reader, error) {
	prefixError := "failed to create Teams connector request"

	request := connectors.CreateConnectorRequestTeams{
		ConnectorTypeId: connectors.CreateConnectorRequestTeamsConnectorTypeIdDotTeams,
		Name:            connector.Name,
	}

	if err := json.Unmarshal([]byte(connector.SecretsJSON), &request.Secrets); err != nil {
		return nil, fmt.Errorf("%s: failed to unmarshal [secrets] attribute: %w", prefixError, err)
	}

	bt, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to marshal request: %w", prefixError, err)
	}

	return bytes.NewReader(bt), nil
}

func createConnectorRequestTines(connector models.KibanaActionConnector) (io.Reader, error) {
	prefixError := "failed to create Tines connector request"

	request := connectors.CreateConnectorRequestTines{
		ConnectorTypeId: connectors.CreateConnectorRequestTinesConnectorTypeIdDotTines,
		Name:            connector.Name,
	}

	if err := json.Unmarshal([]byte(connector.ConfigJSON), &request.Config); err != nil {
		return nil, fmt.Errorf("%s: failed to unmarshal [config] attribute: %w", prefixError, err)
	}

	if err := json.Unmarshal([]byte(connector.SecretsJSON), &request.Secrets); err != nil {
		return nil, fmt.Errorf("%s: failed to unmarshal [secrets] attribute: %w", prefixError, err)
	}

	bt, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to marshal request: %w", prefixError, err)
	}

	return bytes.NewReader(bt), nil
}

func createConnectorRequestWebhook(connector models.KibanaActionConnector) (io.Reader, error) {
	prefixError := "failed to create Webhook connector request"

	request := connectors.CreateConnectorRequestWebhook{
		ConnectorTypeId: connectors.CreateConnectorRequestWebhookConnectorTypeIdDotWebhook,
		Name:            connector.Name,
	}

	if err := json.Unmarshal([]byte(connector.ConfigJSON), &request.Config); err != nil {
		return nil, fmt.Errorf("%s: failed to unmarshal [config] attribute: %w", prefixError, err)
	}

	if err := json.Unmarshal([]byte(connector.SecretsJSON), &request.Secrets); err != nil {
		return nil, fmt.Errorf("%s: failed to unmarshal [secrets] attribute: %w", prefixError, err)
	}

	bt, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to marshal request: %w", prefixError, err)
	}

	return bytes.NewReader(bt), nil
}

func createConnectorRequestXmatters(connector models.KibanaActionConnector) (io.Reader, error) {
	prefixError := "failed to create Xmatters connector request"

	request := connectors.CreateConnectorRequestXmatters{
		ConnectorTypeId: connectors.CreateConnectorRequestXmattersConnectorTypeIdDotXmatters,
		Name:            connector.Name,
	}

	if err := json.Unmarshal([]byte(connector.ConfigJSON), &request.Config); err != nil {
		return nil, fmt.Errorf("%s: failed to unmarshal [config] attribute: %w", prefixError, err)
	}

	if err := json.Unmarshal([]byte(connector.SecretsJSON), &request.Secrets); err != nil {
		return nil, fmt.Errorf("%s: failed to unmarshal [secrets] attribute: %w", prefixError, err)
	}

	bt, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to marshal request: %w", prefixError, err)
	}

	return bytes.NewReader(bt), nil
}

func updateConnectorRequestCasesWebhook(connector models.KibanaActionConnector) (io.Reader, error) {
	prefixError := "failed to create CasesWebhook connector request"

	request := connectors.UpdateConnectorRequestCasesWebhook{
		Name: connector.Name,
	}

	if err := json.Unmarshal([]byte(connector.ConfigJSON), &request.Config); err != nil {
		return nil, fmt.Errorf("%s: failed to unmarshal [config] attribute: %w", prefixError, err)
	}

	if err := json.Unmarshal([]byte(connector.SecretsJSON), &request.Secrets); err != nil {
		return nil, fmt.Errorf("%s: failed to unmarshal [secrets] attribute: %w", prefixError, err)
	}

	bt, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to marshal request: %w", prefixError, err)
	}

	return bytes.NewReader(bt), nil
}

func updateConnectorRequestEmail(connector models.KibanaActionConnector) (io.Reader, error) {
	prefixError := "failed to create Email connector request"

	request := connectors.UpdateConnectorRequestEmail{
		Name: connector.Name,
	}

	if err := json.Unmarshal([]byte(connector.ConfigJSON), &request.Config); err != nil {
		return nil, fmt.Errorf("%s: failed to unmarshal [config] attribute: %w", prefixError, err)
	}

	if err := json.Unmarshal([]byte(connector.SecretsJSON), &request.Secrets); err != nil {
		return nil, fmt.Errorf("%s: failed to unmarshal [secrets] attribute: %w", prefixError, err)
	}

	bt, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to marshal request: %w", prefixError, err)
	}

	return bytes.NewReader(bt), nil
}

func updateConnectorRequestIndex(connector models.KibanaActionConnector) (io.Reader, error) {
	prefixError := "failed to create Index connector request"

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

func updateConnectorRequestJira(connector models.KibanaActionConnector) (io.Reader, error) {
	prefixError := "failed to create Jira connector request"

	request := connectors.UpdateConnectorRequestJira{
		Name: connector.Name,
	}

	if err := json.Unmarshal([]byte(connector.ConfigJSON), &request.Config); err != nil {
		return nil, fmt.Errorf("%s: failed to unmarshal [config] attribute: %w", prefixError, err)
	}

	if err := json.Unmarshal([]byte(connector.SecretsJSON), &request.Secrets); err != nil {
		return nil, fmt.Errorf("%s: failed to unmarshal [secrets] attribute: %w", prefixError, err)
	}

	bt, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to marshal request: %w", prefixError, err)
	}

	return bytes.NewReader(bt), nil
}

func updateConnectorRequestOpsgenie(connector models.KibanaActionConnector) (io.Reader, error) {
	prefixError := "failed to create Opsgenie connector request"

	request := connectors.UpdateConnectorRequestOpsgenie{
		Name: connector.Name,
	}

	if err := json.Unmarshal([]byte(connector.ConfigJSON), &request.Config); err != nil {
		return nil, fmt.Errorf("%s: failed to unmarshal [config] attribute: %w", prefixError, err)
	}

	if err := json.Unmarshal([]byte(connector.SecretsJSON), &request.Secrets); err != nil {
		return nil, fmt.Errorf("%s: failed to unmarshal [secrets] attribute: %w", prefixError, err)
	}

	bt, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to marshal request: %w", prefixError, err)
	}

	return bytes.NewReader(bt), nil
}

func updateConnectorRequestPagerduty(connector models.KibanaActionConnector) (io.Reader, error) {
	prefixError := "failed to create Pagerduty connector request"

	request := connectors.UpdateConnectorRequestPagerduty{
		Name: connector.Name,
	}

	if err := json.Unmarshal([]byte(connector.ConfigJSON), &request.Config); err != nil {
		return nil, fmt.Errorf("%s: failed to unmarshal [config] attribute: %w", prefixError, err)
	}

	if err := json.Unmarshal([]byte(connector.SecretsJSON), &request.Secrets); err != nil {
		return nil, fmt.Errorf("%s: failed to unmarshal [secrets] attribute: %w", prefixError, err)
	}

	bt, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to marshal request: %w", prefixError, err)
	}

	return bytes.NewReader(bt), nil
}

func updateConnectorRequestResilient(connector models.KibanaActionConnector) (io.Reader, error) {
	prefixError := "failed to create Resilient connector request"

	request := connectors.UpdateConnectorRequestResilient{
		Name: connector.Name,
	}

	if err := json.Unmarshal([]byte(connector.ConfigJSON), &request.Config); err != nil {
		return nil, fmt.Errorf("%s: failed to unmarshal [config] attribute: %w", prefixError, err)
	}

	if err := json.Unmarshal([]byte(connector.SecretsJSON), &request.Secrets); err != nil {
		return nil, fmt.Errorf("%s: failed to unmarshal [secrets] attribute: %w", prefixError, err)
	}

	bt, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to marshal request: %w", prefixError, err)
	}

	return bytes.NewReader(bt), nil
}

func updateConnectorRequestServicenow(connector models.KibanaActionConnector) (io.Reader, error) {
	prefixError := "failed to create Servicenow connector request"

	request := connectors.UpdateConnectorRequestServicenow{
		Name: connector.Name,
	}

	if err := json.Unmarshal([]byte(connector.ConfigJSON), &request.Config); err != nil {
		return nil, fmt.Errorf("%s: failed to unmarshal [config] attribute: %w", prefixError, err)
	}

	if err := json.Unmarshal([]byte(connector.SecretsJSON), &request.Secrets); err != nil {
		return nil, fmt.Errorf("%s: failed to unmarshal [secrets] attribute: %w", prefixError, err)
	}

	bt, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to marshal request: %w", prefixError, err)
	}

	return bytes.NewReader(bt), nil
}

func updateConnectorRequestServicenowItom(connector models.KibanaActionConnector) (io.Reader, error) {
	prefixError := "failed to create ServicenowItom connector request"

	request := connectors.UpdateConnectorRequestServicenowItom{
		Name: connector.Name,
	}

	if err := json.Unmarshal([]byte(connector.ConfigJSON), &request.Config); err != nil {
		return nil, fmt.Errorf("%s: failed to unmarshal [config] attribute: %w", prefixError, err)
	}

	if err := json.Unmarshal([]byte(connector.SecretsJSON), &request.Secrets); err != nil {
		return nil, fmt.Errorf("%s: failed to unmarshal [secrets] attribute: %w", prefixError, err)
	}

	bt, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to marshal request: %w", prefixError, err)
	}

	return bytes.NewReader(bt), nil
}

func updateConnectorRequestServicenowSir(connector models.KibanaActionConnector) (io.Reader, error) {
	prefixError := "failed to create ServicenowSir connector request"

	request := connectors.UpdateConnectorRequestServicenowSir{
		Name: connector.Name,
	}

	if err := json.Unmarshal([]byte(connector.ConfigJSON), &request.Config); err != nil {
		return nil, fmt.Errorf("%s: failed to unmarshal [config] attribute: %w", prefixError, err)
	}

	if err := json.Unmarshal([]byte(connector.SecretsJSON), &request.Secrets); err != nil {
		return nil, fmt.Errorf("%s: failed to unmarshal [secrets] attribute: %w", prefixError, err)
	}

	bt, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to marshal request: %w", prefixError, err)
	}

	return bytes.NewReader(bt), nil
}

func updateConnectorRequestServerlog(connector models.KibanaActionConnector) (io.Reader, error) {
	prefixError := "failed to create Serverlog connector request"

	request := connectors.UpdateConnectorRequestServerlog{
		Name: connector.Name,
	}

	bt, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to marshal request: %w", prefixError, err)
	}

	return bytes.NewReader(bt), nil
}

func updateConnectorRequestSlack(connector models.KibanaActionConnector) (io.Reader, error) {
	prefixError := "failed to create Slack connector request"

	request := connectors.UpdateConnectorRequestSlack{
		Name: connector.Name,
	}

	if err := json.Unmarshal([]byte(connector.SecretsJSON), &request.Secrets); err != nil {
		return nil, fmt.Errorf("%s: failed to unmarshal [secrets] attribute: %w", prefixError, err)
	}

	bt, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to marshal request: %w", prefixError, err)
	}

	return bytes.NewReader(bt), nil
}

func updateConnectorRequestSwimlane(connector models.KibanaActionConnector) (io.Reader, error) {
	prefixError := "failed to create Swimlane connector request"

	request := connectors.UpdateConnectorRequestSwimlane{
		Name: connector.Name,
	}

	if err := json.Unmarshal([]byte(connector.ConfigJSON), &request.Config); err != nil {
		return nil, fmt.Errorf("%s: failed to unmarshal [config] attribute: %w", prefixError, err)
	}

	if err := json.Unmarshal([]byte(connector.SecretsJSON), &request.Secrets); err != nil {
		return nil, fmt.Errorf("%s: failed to unmarshal [secrets] attribute: %w", prefixError, err)
	}

	bt, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to marshal request: %w", prefixError, err)
	}

	return bytes.NewReader(bt), nil
}

func updateConnectorRequestTeams(connector models.KibanaActionConnector) (io.Reader, error) {
	prefixError := "failed to create Teams connector request"

	request := connectors.UpdateConnectorRequestTeams{
		Name: connector.Name,
	}

	if err := json.Unmarshal([]byte(connector.SecretsJSON), &request.Secrets); err != nil {
		return nil, fmt.Errorf("%s: failed to unmarshal [secrets] attribute: %w", prefixError, err)
	}

	bt, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to marshal request: %w", prefixError, err)
	}

	return bytes.NewReader(bt), nil
}

func updateConnectorRequestTines(connector models.KibanaActionConnector) (io.Reader, error) {
	prefixError := "failed to create Tines connector request"

	request := connectors.UpdateConnectorRequestTines{
		Name: connector.Name,
	}

	if err := json.Unmarshal([]byte(connector.ConfigJSON), &request.Config); err != nil {
		return nil, fmt.Errorf("%s: failed to unmarshal [config] attribute: %w", prefixError, err)
	}

	if err := json.Unmarshal([]byte(connector.SecretsJSON), &request.Secrets); err != nil {
		return nil, fmt.Errorf("%s: failed to unmarshal [secrets] attribute: %w", prefixError, err)
	}

	bt, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to marshal request: %w", prefixError, err)
	}

	return bytes.NewReader(bt), nil
}

func updateConnectorRequestWebhook(connector models.KibanaActionConnector) (io.Reader, error) {
	prefixError := "failed to create Webhook connector request"

	request := connectors.UpdateConnectorRequestWebhook{
		Name: connector.Name,
	}

	if err := json.Unmarshal([]byte(connector.ConfigJSON), &request.Config); err != nil {
		return nil, fmt.Errorf("%s: failed to unmarshal [config] attribute: %w", prefixError, err)
	}

	if err := json.Unmarshal([]byte(connector.SecretsJSON), &request.Secrets); err != nil {
		return nil, fmt.Errorf("%s: failed to unmarshal [secrets] attribute: %w", prefixError, err)
	}

	bt, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to marshal request: %w", prefixError, err)
	}

	return bytes.NewReader(bt), nil
}

func updateConnectorRequestXmatters(connector models.KibanaActionConnector) (io.Reader, error) {
	prefixError := "failed to create Xmatters connector request"

	request := connectors.UpdateConnectorRequestXmatters{
		Name: connector.Name,
	}

	if err := json.Unmarshal([]byte(connector.ConfigJSON), &request.Config); err != nil {
		return nil, fmt.Errorf("%s: failed to unmarshal [config] attribute: %w", prefixError, err)
	}

	if err := json.Unmarshal([]byte(connector.SecretsJSON), &request.Secrets); err != nil {
		return nil, fmt.Errorf("%s: failed to unmarshal [secrets] attribute: %w", prefixError, err)
	}

	bt, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to marshal request: %w", prefixError, err)
	}

	return bytes.NewReader(bt), nil
}

func connectorResponseToModel(spaceID string, properties connectors.ConnectorResponseProperties) (*models.KibanaActionConnector, error) {
	discriminator, err := properties.Discriminator()
	if err != nil {
		return nil, err
	}
	switch connectors.ConnectorTypes(discriminator) {

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

	case connectors.ConnectorTypesDotEmail:
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

	case connectors.ConnectorTypesDotIndex:
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

	return nil, fmt.Errorf("unknown connector type [%s]", discriminator)
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
