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
	"encoding/json"
	"errors"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
)

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

// Users can omit optional fields in the config JSON.
// This helper remarshals the config so empty optional fields are included,
// avoiding plan diffs when Kibana returns fields that were omitted from
// the original configuration.
func remarshalConfig[T any](plan string) (string, error) {
	return connectorConfigWithDefaults[T](plan, nil)
}

// connectorConfigWithDefaults is the generic helper shared by all per-connector
// defaults functions. It unmarshals plan into T, calls setDefaults to fill any
// missing fields, then marshals back to JSON.
func connectorConfigWithDefaults[T any](plan string, setDefaults func(*T)) (string, error) {
	var config T
	if err := json.Unmarshal([]byte(plan), &config); err != nil {
		return "", err
	}
	if setDefaults != nil {
		setDefaults(&config)
	}
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
// otherwise the "Other" config is remarshaled without modification.
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
