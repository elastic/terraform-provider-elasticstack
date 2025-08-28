package kibana_oapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
)

func CreateConnector(ctx context.Context, client *Client, connectorOld models.KibanaActionConnector) (string, diag.Diagnostics) {
	body, err := createConnectorRequestBody(connectorOld)
	if err != nil {
		return "", diag.FromErr(err)
	}

	resp, err := client.API.PostActionsConnectorIdWithResponse(
		ctx, connectorOld.SpaceID, connectorOld.ConnectorID, body,
		// When there isn't an explicit connector ID the request path will include a trailing slash
		// Kibana 8.7 and lower return a 404 for such request paths, whilst 8.8+ correctly handle then empty ID parameter
		// This request editor ensures that the trailing slash is removed allowing all supported
		// Stack versions to correctly create connectors without an explicit ID
		func(ctx context.Context, req *http.Request) error {
			if connectorOld.ConnectorID == "" {
				req.URL.Path = strings.TrimRight(req.URL.Path, "/")
			}
			return nil
		},
	)
	if err != nil {
		return "", diag.FromErr(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		return resp.JSON200.Id, nil
	default:
		return "", reportUnknownErrorSDK(resp.StatusCode(), resp.Body)
	}
}

func UpdateConnector(ctx context.Context, client *Client, connectorOld models.KibanaActionConnector) (string, diag.Diagnostics) {
	body, err := updateConnectorRequestBody(connectorOld)
	if err != nil {
		return "", diag.FromErr(err)
	}

	resp, err := client.API.PutActionsConnectorIdWithResponse(ctx, connectorOld.SpaceID, connectorOld.ConnectorID, body)
	if err != nil {
		return "", diag.Errorf("unable to update connector: [%v]", err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		return resp.JSON200.Id, nil
	default:
		return "", reportUnknownErrorSDK(resp.StatusCode(), resp.Body)
	}
}

func GetConnector(ctx context.Context, client *Client, connectorID, spaceID string) (*models.KibanaActionConnector, diag.Diagnostics) {
	resp, err := client.API.GetActionsConnectorIdWithResponse(ctx, spaceID, connectorID)
	if err != nil {
		return nil, diag.Errorf("unable to get connector: [%v]", err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		return ConnectorResponseToModel(spaceID, resp.JSON200)
	case http.StatusNotFound:
		return nil, nil
	default:
		return nil, reportUnknownErrorSDK(resp.StatusCode(), resp.Body)
	}
}

func SearchConnectors(ctx context.Context, client *Client, connectorName, spaceID, connectorTypeID string) ([]*models.KibanaActionConnector, diag.Diagnostics) {
	resp, err := client.API.GetActionsConnectorsWithResponse(ctx, spaceID)
	if err != nil {
		return nil, diag.Errorf("unable to get connectors: [%v]", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, reportUnknownErrorSDK(resp.StatusCode(), resp.Body)
	}

	foundConnectors := []*models.KibanaActionConnector{}
	for _, connector := range *resp.JSON200 {
		if connector.Name != connectorName {
			continue
		}

		if connectorTypeID != "" && connector.ConnectorTypeId != connectorTypeID {
			continue
		}

		c, diags := ConnectorResponseToModel(spaceID, &connector)
		if diags.HasError() {
			return nil, diags
		}

		foundConnectors = append(foundConnectors, c)
	}
	if len(foundConnectors) == 0 {
		tflog.Debug(ctx, fmt.Sprintf("no connectors found with name [%s/%s] and type [%s]", spaceID, connectorName, connectorTypeID))
	}

	return foundConnectors, nil
}

func ConnectorResponseToModel(spaceID string, connector *kbapi.ConnectorResponse) (*models.KibanaActionConnector, diag.Diagnostics) {
	if connector == nil {
		return nil, diag.Errorf("connector response is nil")
	}

	var configJSON []byte
	if connector.Config != nil {
		configMap := *connector.Config
		for k, v := range configMap {
			if v == nil {
				delete(configMap, k)
			}
		}

		var err error
		configJSON, err = json.Marshal(configMap)
		if err != nil {
			return nil, diag.Errorf("unable to marshal config: %v", err)
		}

		// If we have a specific config type, marshal into and out of that to
		// remove any extra fields Kibana may have returned.
		handler, ok := connectorConfigHandlers[connector.ConnectorTypeId]
		if ok {
			configJSONString, err := handler.remarshalConfig(string(configJSON))
			if err != nil {
				return nil, diag.Errorf("failed to remarshal config: %v", err)
			}

			configJSON = []byte(configJSONString)
		}
	}

	model := &models.KibanaActionConnector{
		ConnectorID:     connector.Id,
		SpaceID:         spaceID,
		Name:            connector.Name,
		ConfigJSON:      string(configJSON),
		ConnectorTypeID: connector.ConnectorTypeId,
		IsDeprecated:    connector.IsDeprecated,
		IsPreconfigured: connector.IsPreconfigured,
	}

	if connector.IsMissingSecrets != nil {
		model.IsMissingSecrets = *connector.IsMissingSecrets
	}

	return model, nil
}

func DeleteConnector(ctx context.Context, client *Client, connectorID string, spaceID string) diag.Diagnostics {
	resp, err := client.API.DeleteActionsConnectorIdWithResponse(ctx, spaceID, connectorID)
	if err != nil {
		return diag.Errorf("unable to delete connector: [%v]", err)
	}

	if resp.StatusCode() != http.StatusOK && resp.StatusCode() != http.StatusNoContent {
		return reportUnknownErrorSDK(resp.StatusCode(), resp.Body)
	}

	return nil
}

type connectorConfigHandler struct {
	defaults        func(plan, backend string) (string, error)
	remarshalConfig func(config string) (string, error)
}

var connectorConfigHandlers = map[string]connectorConfigHandler{
	".cases-webhook": {
		defaults:        connectorConfigWithDefaultsCasesWebhook,
		remarshalConfig: remarshalConfig[kbapi.CasesWebhookConfig],
	},
	".email": {
		defaults:        connectorConfigWithDefaultsEmail,
		remarshalConfig: remarshalConfig[kbapi.EmailConfig],
	},
	".gemini": {
		defaults:        connectorConfigWithDefaultsGemini,
		remarshalConfig: remarshalConfig[kbapi.GeminiConfig],
	},
	".index": {
		defaults:        connectorConfigWithDefaultsIndex,
		remarshalConfig: remarshalConfig[kbapi.IndexConfig],
	},
	".jira": {
		defaults:        connectorConfigWithDefaultsJira,
		remarshalConfig: remarshalConfig[kbapi.JiraConfig],
	},
	".opsgenie": {
		defaults:        connectorConfigWithDefaultsOpsgenie,
		remarshalConfig: remarshalConfig[kbapi.OpsgenieConfig],
	},
	".pagerduty": {
		defaults:        connectorConfigWithDefaultsPagerduty,
		remarshalConfig: remarshalConfig[kbapi.PagerdutyConfig],
	},
	".resilient": {
		defaults:        connectorConfigWithDefaultsResilient,
		remarshalConfig: remarshalConfig[kbapi.ResilientConfig],
	},
	".servicenow": {
		defaults:        connectorConfigWithDefaultsServicenow,
		remarshalConfig: remarshalConfig[kbapi.ServicenowConfig],
	},
	".servicenow-itom": {
		defaults:        connectorConfigWithDefaultsServicenowItom,
		remarshalConfig: remarshalConfig[kbapi.ServicenowItomConfig],
	},
	".servicenow-sir": {
		defaults:        connectorConfigWithDefaultsServicenowSir,
		remarshalConfig: remarshalConfig[kbapi.ServicenowConfig],
	},
	".swimlane": {
		defaults:        connectorConfigWithDefaultsSwimlane,
		remarshalConfig: remarshalConfig[kbapi.SwimlaneConfig],
	},
	".tines": {
		defaults:        connectorConfigWithDefaultsTines,
		remarshalConfig: remarshalConfig[kbapi.TinesConfig],
	},
	".webhook": {
		defaults:        connectorConfigWithDefaultsWebhook,
		remarshalConfig: remarshalConfig[kbapi.WebhookConfig],
	},
	".xmatters": {
		defaults:        connectorConfigWithDefaultsXmatters,
		remarshalConfig: remarshalConfig[kbapi.XmattersConfig],
	},
}

func ConnectorConfigWithDefaults(connectorTypeID, plan, backend, state string) (string, error) {
	handler, ok := connectorConfigHandlers[connectorTypeID]
	if !ok {
		return plan, errors.New("unknown connector type ID: " + connectorTypeID)
	}

	return handler.defaults(plan, backend)
}

// User can omit optonal fields in config JSON.
// The func adds empty optional fields to the diff.
// Otherwise plan command shows omitted fields as the diff,
// because backend returns all fields.
func remarshalConfig[T any](plan string) (string, error) {
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

func connectorConfigWithDefaultsCasesWebhook(plan, _ string) (string, error) {
	var custom kbapi.CasesWebhookConfig
	if err := json.Unmarshal([]byte(plan), &custom); err != nil {
		return "", err
	}
	if custom.CreateIncidentMethod == nil {
		custom.CreateIncidentMethod = utils.Pointer(kbapi.CasesWebhookConfigCreateIncidentMethodPost)
	}
	if custom.HasAuth == nil {
		custom.HasAuth = utils.Pointer(true)
	}
	if custom.UpdateIncidentMethod == nil {
		custom.UpdateIncidentMethod = utils.Pointer(kbapi.CasesWebhookConfigUpdateIncidentMethodPut)
	}
	customJSON, err := json.Marshal(custom)
	if err != nil {
		return "", err
	}
	return string(customJSON), nil
}

func connectorConfigWithDefaultsEmail(plan, _ string) (string, error) {
	var custom kbapi.EmailConfig
	if err := json.Unmarshal([]byte(plan), &custom); err != nil {
		return "", err
	}
	if custom.HasAuth == nil {
		custom.HasAuth = utils.Pointer(true)
	}
	if custom.Service == nil {
		custom.Service = utils.Pointer(kbapi.EmailConfigService("other"))
	}
	customJSON, err := json.Marshal(custom)
	if err != nil {
		return "", err
	}
	return string(customJSON), nil
}

func connectorConfigWithDefaultsGemini(plan, _ string) (string, error) {
	return plan, nil
}

func connectorConfigWithDefaultsIndex(plan, _ string) (string, error) {
	var custom kbapi.IndexConfig
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

func connectorConfigWithDefaultsJira(plan, _ string) (string, error) {
	return remarshalConfig[kbapi.JiraConfig](plan)
}

func connectorConfigWithDefaultsOpsgenie(plan, _ string) (string, error) {
	return plan, nil
}

func connectorConfigWithDefaultsPagerduty(plan, _ string) (string, error) {
	return remarshalConfig[kbapi.PagerdutyConfig](plan)
}

func connectorConfigWithDefaultsResilient(plan, _ string) (string, error) {
	return plan, nil
}

func connectorConfigWithDefaultsServicenow(plan, backend string) (string, error) {
	var planConfig kbapi.ServicenowConfig
	if err := json.Unmarshal([]byte(plan), &planConfig); err != nil {
		return "", err
	}
	var backendConfig kbapi.ServicenowConfig
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

func connectorConfigWithDefaultsServicenowItom(plan, _ string) (string, error) {
	var custom kbapi.ServicenowItomConfig
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

func connectorConfigWithDefaultsSwimlane(plan, _ string) (string, error) {
	var custom kbapi.SwimlaneConfig
	if err := json.Unmarshal([]byte(plan), &custom); err != nil {
		return "", err
	}
	if custom.Mappings == nil {
		custom.Mappings = &struct {
			AlertIdConfig *struct {
				FieldType string "json:\"fieldType\""
				Id        string "json:\"id\""
				Key       string "json:\"key\""
				Name      string "json:\"name\""
			} "json:\"alertIdConfig,omitempty\""
			CaseIdConfig *struct {
				FieldType string "json:\"fieldType\""
				Id        string "json:\"id\""
				Key       string "json:\"key\""
				Name      string "json:\"name\""
			} "json:\"caseIdConfig,omitempty\""
			CaseNameConfig *struct {
				FieldType string "json:\"fieldType\""
				Id        string "json:\"id\""
				Key       string "json:\"key\""
				Name      string "json:\"name\""
			} "json:\"caseNameConfig,omitempty\""
			CommentsConfig *struct {
				FieldType string "json:\"fieldType\""
				Id        string "json:\"id\""
				Key       string "json:\"key\""
				Name      string "json:\"name\""
			} "json:\"commentsConfig,omitempty\""
			DescriptionConfig *struct {
				FieldType string "json:\"fieldType\""
				Id        string "json:\"id\""
				Key       string "json:\"key\""
				Name      string "json:\"name\""
			} "json:\"descriptionConfig,omitempty\""
			RuleNameConfig *struct {
				FieldType string "json:\"fieldType\""
				Id        string "json:\"id\""
				Key       string "json:\"key\""
				Name      string "json:\"name\""
			} "json:\"ruleNameConfig,omitempty\""
			SeverityConfig *struct {
				FieldType string "json:\"fieldType\""
				Id        string "json:\"id\""
				Key       string "json:\"key\""
				Name      string "json:\"name\""
			} "json:\"severityConfig,omitempty\""
		}{}
	}
	customJSON, err := json.Marshal(custom)
	if err != nil {
		return "", err
	}
	return string(customJSON), nil
}

func connectorConfigWithDefaultsTines(plan, _ string) (string, error) {
	return plan, nil
}

func connectorConfigWithDefaultsWebhook(plan, _ string) (string, error) {
	return plan, nil
}

func connectorConfigWithDefaultsXmatters(plan, _ string) (string, error) {
	var custom kbapi.XmattersConfig
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

func createConnectorRequestBody(connector models.KibanaActionConnector) (kbapi.PostActionsConnectorIdJSONRequestBody, error) {
	req := kbapi.PostActionsConnectorIdJSONRequestBody{
		ConnectorTypeId: connector.ConnectorTypeID,
		Name:            connector.Name,
		Config:          &kbapi.CreateConnectorConfig{},
		Secrets:         &kbapi.CreateConnectorSecrets{},
	}

	if len(connector.ConfigJSON) > 0 {
		if err := json.Unmarshal([]byte(connector.ConfigJSON), &req.Config.AdditionalProperties); err != nil {
			return kbapi.PostActionsConnectorIdJSONRequestBody{}, fmt.Errorf("failed to unmarshal [config] attribute: %w", err)
		}
	}

	if len(connector.SecretsJSON) > 0 {
		if err := json.Unmarshal([]byte(connector.SecretsJSON), &req.Secrets.AdditionalProperties); err != nil {
			return kbapi.PostActionsConnectorIdJSONRequestBody{}, fmt.Errorf("failed to unmarshal [secrets] attribute: %w", err)
		}
	}

	return req, nil
}

func updateConnectorRequestBody(connector models.KibanaActionConnector) (kbapi.PutActionsConnectorIdJSONRequestBody, error) {
	req := kbapi.PutActionsConnectorIdJSONRequestBody{
		Name:    connector.Name,
		Config:  &kbapi.UpdateConnectorConfig{},
		Secrets: &kbapi.UpdateConnectorSecrets{},
	}

	if len(connector.ConfigJSON) > 0 {
		if err := json.Unmarshal([]byte(connector.ConfigJSON), &req.Config.AdditionalProperties); err != nil {
			return kbapi.PutActionsConnectorIdJSONRequestBody{}, fmt.Errorf("failed to unmarshal [config] attribute: %w", err)
		}
	}

	if len(connector.SecretsJSON) > 0 {
		if err := json.Unmarshal([]byte(connector.SecretsJSON), &req.Secrets.AdditionalProperties); err != nil {
			return kbapi.PutActionsConnectorIdJSONRequestBody{}, fmt.Errorf("failed to unmarshal [secrets] attribute: %w", err)
		}
	}

	return req, nil
}
