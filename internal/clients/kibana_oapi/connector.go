package kibana_oapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	fwdiag "github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	sdkdiag "github.com/hashicorp/terraform-plugin-sdk/v2/diag"
)

func CreateConnector(ctx context.Context, client *Client, connectorOld models.KibanaActionConnector) (string, fwdiag.Diagnostics) {
	body, err := createConnectorRequestBody(connectorOld)
	if err != nil {
		return "", fwdiag.Diagnostics{fwdiag.NewErrorDiagnostic("Failed to create connector request body", err.Error())}
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
		return "", fwdiag.Diagnostics{fwdiag.NewErrorDiagnostic("HTTP request failed", err.Error())}
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		return resp.JSON200.Id, nil
	default:
		return "", reportUnknownError(resp.StatusCode(), resp.Body)
	}
}

func UpdateConnector(ctx context.Context, client *Client, connectorOld models.KibanaActionConnector) (string, fwdiag.Diagnostics) {
	body, err := updateConnectorRequestBody(connectorOld)
	if err != nil {
		return "", fwdiag.Diagnostics{fwdiag.NewErrorDiagnostic("Failed to create update request body", err.Error())}
	}

	resp, err := client.API.PutActionsConnectorIdWithResponse(ctx, connectorOld.SpaceID, connectorOld.ConnectorID, body)
	if err != nil {
		return "", fwdiag.Diagnostics{fwdiag.NewErrorDiagnostic("Unable to update connector", err.Error())}
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		return resp.JSON200.Id, nil
	default:
		return "", reportUnknownError(resp.StatusCode(), resp.Body)
	}
}

func GetConnector(ctx context.Context, client *Client, connectorID, spaceID string) (*models.KibanaActionConnector, fwdiag.Diagnostics) {
	resp, err := client.API.GetActionsConnectorIdWithResponse(ctx, spaceID, connectorID)
	if err != nil {
		return nil, fwdiag.Diagnostics{fwdiag.NewErrorDiagnostic("Unable to get connector", err.Error())}
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		return ConnectorResponseToModel(spaceID, resp.JSON200)
	case http.StatusNotFound:
		return nil, nil
	default:
		return nil, reportUnknownError(resp.StatusCode(), resp.Body)
	}
}

func SearchConnectors(ctx context.Context, client *Client, connectorName, spaceID, connectorTypeID string) ([]*models.KibanaActionConnector, sdkdiag.Diagnostics) {
	resp, err := client.API.GetActionsConnectorsWithResponse(ctx, spaceID)
	if err != nil {
		return nil, sdkdiag.Errorf("unable to get connectors: [%v]", err)
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

		c, fwDiags := ConnectorResponseToModel(spaceID, &connector)
		if fwDiags.HasError() {
			return nil, diagutil.SDKDiagsFromFramework(fwDiags)
		}

		foundConnectors = append(foundConnectors, c)
	}
	if len(foundConnectors) == 0 {
		tflog.Debug(ctx, fmt.Sprintf("no connectors found with name [%s/%s] and type [%s]", spaceID, connectorName, connectorTypeID))
	}

	return foundConnectors, nil
}

func ConnectorResponseToModel(spaceID string, connector *kbapi.ConnectorResponse) (*models.KibanaActionConnector, fwdiag.Diagnostics) {
	if connector == nil {
		return nil, fwdiag.Diagnostics{fwdiag.NewErrorDiagnostic("Invalid connector response", "connector response is nil")}
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
			return nil, fwdiag.Diagnostics{fwdiag.NewErrorDiagnostic("Unable to marshal config", err.Error())}
		}

		// If we have a specific config type, marshal into and out of that to
		// remove any extra fields Kibana may have returned.
		handler, ok := connectorConfigHandlers[connector.ConnectorTypeId]
		if ok {
			configJSONString, err := handler.remarshalConfig(string(configJSON))
			if err != nil {
				return nil, fwdiag.Diagnostics{fwdiag.NewErrorDiagnostic("Failed to remarshal config", err.Error())}
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

func DeleteConnector(ctx context.Context, client *Client, connectorID string, spaceID string) fwdiag.Diagnostics {
	resp, err := client.API.DeleteActionsConnectorIdWithResponse(ctx, spaceID, connectorID)
	if err != nil {
		return fwdiag.Diagnostics{fwdiag.NewErrorDiagnostic("Unable to delete connector", err.Error())}
	}

	if resp.StatusCode() != http.StatusOK && resp.StatusCode() != http.StatusNoContent {
		return reportUnknownError(resp.StatusCode(), resp.Body)
	}

	return nil
}

type connectorConfigHandler struct {
	defaults        func(plan string) (string, error)
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
		remarshalConfig: remarshalConfig[kbapi.OpsgenieConfig],
	},
	".pagerduty": {
		defaults:        connectorConfigWithDefaultsPagerduty,
		remarshalConfig: remarshalConfig[kbapi.PagerdutyConfig],
	},
	".resilient": {
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
	".slack_api": {
		remarshalConfig: remarshalConfig[kbapi.SlackApiConfig],
	},
	".swimlane": {
		defaults:        connectorConfigWithDefaultsSwimlane,
		remarshalConfig: remarshalConfig[kbapi.SwimlaneConfig],
	},
	".tines": {
		remarshalConfig: remarshalConfig[kbapi.TinesConfig],
	},
	".webhook": {
		remarshalConfig: remarshalConfig[kbapi.WebhookConfig],
	},
	".xmatters": {
		defaults:        connectorConfigWithDefaultsXmatters,
		remarshalConfig: remarshalConfig[kbapi.XmattersConfig],
	},
}

func ConnectorConfigWithDefaults(connectorTypeID, plan string) (string, error) {
	handler, ok := connectorConfigHandlers[connectorTypeID]
	if !ok {
		return plan, errors.New("unknown connector type ID: " + connectorTypeID)
	}

	if handler.defaults == nil {
		return plan, nil
	}

	return handler.defaults(plan)
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

func connectorConfigWithDefaultsCasesWebhook(plan string) (string, error) {
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

func connectorConfigWithDefaultsEmail(plan string) (string, error) {
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

func connectorConfigWithDefaultsIndex(plan string) (string, error) {
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

func connectorConfigWithDefaultsJira(plan string) (string, error) {
	return remarshalConfig[kbapi.JiraConfig](plan)
}

func connectorConfigWithDefaultsPagerduty(plan string) (string, error) {
	return remarshalConfig[kbapi.PagerdutyConfig](plan)
}

func connectorConfigWithDefaultsServicenow(plan string) (string, error) {
	var planConfig kbapi.ServicenowConfig
	if err := json.Unmarshal([]byte(plan), &planConfig); err != nil {
		return "", err
	}
	if planConfig.IsOAuth == nil {
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

func connectorConfigWithDefaultsServicenowSir(plan string) (string, error) {
	return connectorConfigWithDefaultsServicenow(plan)
}

func connectorConfigWithDefaultsSwimlane(plan string) (string, error) {
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

func connectorConfigWithDefaultsXmatters(plan string) (string, error) {
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
