// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package kibanaoapi

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
		func(_ context.Context, req *http.Request) error {
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
		return "", diagutil.ReportUnknownHTTPError(resp.StatusCode(), resp.Body)
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
		return "", diagutil.ReportUnknownHTTPError(resp.StatusCode(), resp.Body)
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
		return nil, diagutil.ReportUnknownHTTPError(resp.StatusCode(), resp.Body)
	}
}

func SearchConnectors(ctx context.Context, client *Client, connectorName, spaceID, connectorTypeID string) ([]*models.KibanaActionConnector, sdkdiag.Diagnostics) {
	resp, err := client.API.GetActionsConnectorsWithResponse(ctx, spaceID)
	if err != nil {
		return nil, sdkdiag.Errorf("unable to get connectors: [%v]", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, diagutil.SDKDiagsFromFramework(diagutil.ReportUnknownHTTPError(resp.StatusCode(), resp.Body))
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
		return diagutil.ReportUnknownHTTPError(resp.StatusCode(), resp.Body)
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
	".bedrock": {
		defaults:        connectorConfigWithDefaultsBedrock,
		remarshalConfig: remarshalConfig[kbapi.BedrockConfig],
	},
	".gen-ai": {
		defaults:        connectorConfigWithDefaultsGenAi,
		remarshalConfig: remarshalConfigGenAi,
	},
	".gemini": {
		remarshalConfig: remarshalConfig[kbapi.GeminiConfig],
	},
	".index": {
		defaults:        connectorConfigWithDefaultsIndex,
		remarshalConfig: remarshalConfig[kbapi.IndexConfig],
	},
	".jira": {
		remarshalConfig: remarshalConfig[kbapi.JiraConfig],
	},
	".opsgenie": {
		remarshalConfig: remarshalConfig[kbapi.OpsgenieConfig],
	},
	".pagerduty": {
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
		defaults:        connectorConfigWithDefaultsServicenow,
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

// connectorConfigWithDefaults is the generic helper shared by all per-connector
// defaults functions. It unmarshals plan into T, calls setDefaults to fill any
// missing fields, then marshals back to JSON.
func connectorConfigWithDefaults[T any](plan string, setDefaults func(*T)) (string, error) {
	var config T
	if err := json.Unmarshal([]byte(plan), &config); err != nil {
		return "", err
	}
	setDefaults(&config)
	customJSON, err := json.Marshal(config)
	if err != nil {
		return "", err
	}
	return string(customJSON), nil
}

// parseGenAiAPIProvider extracts the apiProvider field from a Gen AI connector config JSON.
func parseGenAiAPIProvider(plan string) (string, error) {
	var configMap map[string]any
	if err := json.Unmarshal([]byte(plan), &configMap); err != nil {
		return "", err
	}
	apiProvider, ok := configMap["apiProvider"].(string)
	if !ok {
		return "", errors.New("apiProvider is required for .gen-ai connector type")
	}
	return apiProvider, nil
}

// dispatchGenAiConfig is the single dispatch point for .gen-ai connectors.
// It selects the appropriate per-provider handler based on the apiProvider field.
// When setOtherDefaults is non-nil it is applied to the "Other" provider config;
// otherwise the "Other" config is remarshed without modification.
func dispatchGenAiConfig(plan string, setOtherDefaults func(*kbapi.GenaiOpenaiOtherConfig)) (string, error) {
	apiProvider, err := parseGenAiAPIProvider(plan)
	if err != nil {
		return "", err
	}
	switch apiProvider {
	case "OpenAI":
		return remarshalConfig[kbapi.GenaiOpenaiConfig](plan)
	case "Azure OpenAI":
		return remarshalConfig[kbapi.GenaiAzureConfig](plan)
	case "Other":
		if setOtherDefaults != nil {
			return connectorConfigWithDefaults(plan, setOtherDefaults)
		}
		return remarshalConfig[kbapi.GenaiOpenaiOtherConfig](plan)
	default:
		return "", fmt.Errorf("unsupported apiProvider %q for .gen-ai connector type, must be one of: OpenAI, Azure OpenAI, Other", apiProvider)
	}
}

func remarshalConfigGenAi(plan string) (string, error) {
	return dispatchGenAiConfig(plan, nil)
}

func connectorConfigWithDefaultsBedrock(plan string) (string, error) {
	return connectorConfigWithDefaults(plan, func(c *kbapi.BedrockConfig) {
		if c.DefaultModel == nil {
			c.DefaultModel = new("us.anthropic.claude-sonnet-4-5-20250929-v1:0")
		}
	})
}

func connectorConfigWithDefaultsGenAi(plan string) (string, error) {
	return dispatchGenAiConfig(plan, func(c *kbapi.GenaiOpenaiOtherConfig) {
		if c.VerificationMode == nil {
			c.VerificationMode = new(kbapi.GenaiOpenaiOtherConfigVerificationModeFull)
		}
	})
}

func connectorConfigWithDefaultsCasesWebhook(plan string) (string, error) {
	return connectorConfigWithDefaults(plan, func(c *kbapi.CasesWebhookConfig) {
		if c.AuthType == nil {
			authType := kbapi.WebhookAuthenticationBasic
			c.AuthType = &authType
		}
		if c.CreateIncidentMethod == nil {
			c.CreateIncidentMethod = new(kbapi.CasesWebhookConfigCreateIncidentMethodPost)
		}
		if c.HasAuth == nil {
			c.HasAuth = new(true)
		}
		if c.UpdateIncidentMethod == nil {
			c.UpdateIncidentMethod = new(kbapi.CasesWebhookConfigUpdateIncidentMethodPut)
		}
		if c.CreateCommentMethod == nil {
			c.CreateCommentMethod = new(kbapi.CasesWebhookConfigCreateCommentMethodPut)
		}
	})
}

func connectorConfigWithDefaultsEmail(plan string) (string, error) {
	return connectorConfigWithDefaults(plan, func(c *kbapi.EmailConfig) {
		if c.HasAuth == nil {
			c.HasAuth = new(true)
		}
		if c.Service == nil {
			c.Service = new(kbapi.EmailConfigService("other"))
		}
	})
}

func connectorConfigWithDefaultsIndex(plan string) (string, error) {
	return connectorConfigWithDefaults(plan, func(c *kbapi.IndexConfig) {
		if c.Refresh == nil {
			c.Refresh = new(false)
		}
	})
}

func connectorConfigWithDefaultsServicenow(plan string) (string, error) {
	return connectorConfigWithDefaults(plan, func(c *kbapi.ServicenowConfig) {
		if c.IsOAuth == nil {
			c.IsOAuth = new(false)
		}
		if c.UsesTableApi == nil {
			c.UsesTableApi = new(true)
		}
	})
}

func connectorConfigWithDefaultsServicenowItom(plan string) (string, error) {
	return connectorConfigWithDefaults(plan, func(c *kbapi.ServicenowItomConfig) {
		if c.IsOAuth == nil {
			c.IsOAuth = new(false)
		}
	})
}

func connectorConfigWithDefaultsSwimlane(plan string) (string, error) {
	return connectorConfigWithDefaults(plan, func(c *kbapi.SwimlaneConfig) {
		if c.Mappings == nil {
			c.Mappings = &struct {
				AlertIdConfig *struct { //nolint:revive // var-naming: API struct field
					FieldType string "json:\"fieldType\""
					Id        string "json:\"id\"" //nolint:revive // var-naming: API struct field
					Key       string "json:\"key\""
					Name      string "json:\"name\""
				} "json:\"alertIdConfig,omitempty\""
				CaseIdConfig *struct { //nolint:revive // var-naming: API struct field
					FieldType string "json:\"fieldType\""
					Id        string "json:\"id\"" //nolint:revive // var-naming: API struct field
					Key       string "json:\"key\""
					Name      string "json:\"name\""
				} "json:\"caseIdConfig,omitempty\""
				CaseNameConfig *struct {
					FieldType string "json:\"fieldType\""
					Id        string "json:\"id\"" //nolint:revive // var-naming: API struct field
					Key       string "json:\"key\""
					Name      string "json:\"name\""
				} "json:\"caseNameConfig,omitempty\""
				CommentsConfig *struct {
					FieldType string "json:\"fieldType\""
					Id        string "json:\"id\"" //nolint:revive // var-naming: API struct field
					Key       string "json:\"key\""
					Name      string "json:\"name\""
				} "json:\"commentsConfig,omitempty\""
				DescriptionConfig *struct {
					FieldType string "json:\"fieldType\""
					Id        string "json:\"id\"" //nolint:revive // var-naming: API struct field
					Key       string "json:\"key\""
					Name      string "json:\"name\""
				} "json:\"descriptionConfig,omitempty\""
				RuleNameConfig *struct {
					FieldType string "json:\"fieldType\""
					Id        string "json:\"id\"" //nolint:revive // var-naming: API struct field
					Key       string "json:\"key\""
					Name      string "json:\"name\""
				} "json:\"ruleNameConfig,omitempty\""
				SeverityConfig *struct {
					FieldType string "json:\"fieldType\""
					Id        string "json:\"id\"" //nolint:revive // var-naming: API struct field
					Key       string "json:\"key\""
					Name      string "json:\"name\""
				} "json:\"severityConfig,omitempty\""
			}{}
		}
	})
}

func connectorConfigWithDefaultsXmatters(plan string) (string, error) {
	return connectorConfigWithDefaults(plan, func(c *kbapi.XmattersConfig) {
		if c.UsesBasic == nil {
			c.UsesBasic = new(true)
		}
	})
}

func unmarshalConnectorFields(configJSON, secretsJSON string, configDest, secretsDest *map[string]any) error {
	if len(configJSON) > 0 {
		if err := json.Unmarshal([]byte(configJSON), configDest); err != nil {
			return fmt.Errorf("failed to unmarshal [config] attribute: %w", err)
		}
	}
	if len(secretsJSON) > 0 {
		if err := json.Unmarshal([]byte(secretsJSON), secretsDest); err != nil {
			return fmt.Errorf("failed to unmarshal [secrets] attribute: %w", err)
		}
	}
	return nil
}

func createConnectorRequestBody(connector models.KibanaActionConnector) (kbapi.PostActionsConnectorIdJSONRequestBody, error) {
	req := kbapi.PostActionsConnectorIdJSONRequestBody{
		ConnectorTypeId: connector.ConnectorTypeID,
		Name:            connector.Name,
		Config:          &kbapi.CreateConnectorConfig{},
		Secrets:         &kbapi.CreateConnectorSecrets{},
	}

	if err := unmarshalConnectorFields(connector.ConfigJSON, connector.SecretsJSON, &req.Config.AdditionalProperties, &req.Secrets.AdditionalProperties); err != nil {
		return kbapi.PostActionsConnectorIdJSONRequestBody{}, err
	}

	return req, nil
}

func updateConnectorRequestBody(connector models.KibanaActionConnector) (kbapi.PutActionsConnectorIdJSONRequestBody, error) {
	req := kbapi.PutActionsConnectorIdJSONRequestBody{
		Name:    connector.Name,
		Config:  &kbapi.UpdateConnectorConfig{},
		Secrets: &kbapi.UpdateConnectorSecrets{},
	}

	if err := unmarshalConnectorFields(connector.ConfigJSON, connector.SecretsJSON, &req.Config.AdditionalProperties, &req.Secrets.AdditionalProperties); err != nil {
		return kbapi.PutActionsConnectorIdJSONRequestBody{}, err
	}

	return req, nil
}
