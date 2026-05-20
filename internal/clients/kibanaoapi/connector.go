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
	"fmt"
	"net/http"
	"strings"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	fwdiag "github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// ConnectorResponse mirrors the fields we need from the Kibana connector API responses.
type ConnectorResponse struct {
	Config           *map[string]*any
	ConnectorTypeID  string
	ID               string
	IsDeprecated     bool
	IsMissingSecrets *bool
	IsPreconfigured  bool
	Name             string
}

func CreateConnector(ctx context.Context, client *Client, connector models.KibanaActionConnector) (string, fwdiag.Diagnostics) {
	body, err := createConnectorRequestBody(connector)
	if err != nil {
		return "", fwdiag.Diagnostics{fwdiag.NewErrorDiagnostic("Failed to create connector request body", err.Error())}
	}

	resp, err := client.API.PostActionsConnectorIdWithResponse(
		ctx, connector.SpaceID, connector.ConnectorID, body,
		// When there isn't an explicit connector ID the request path will include a trailing slash
		// Kibana 8.7 and lower return a 404 for such request paths, whilst 8.8+ correctly handle then empty ID parameter
		// This request editor ensures that the trailing slash is removed allowing all supported
		// Stack versions to correctly create connectors without an explicit ID
		func(_ context.Context, req *http.Request) error {
			if connector.ConnectorID == "" {
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

func UpdateConnector(ctx context.Context, client *Client, connector models.KibanaActionConnector) (string, fwdiag.Diagnostics) {
	body, err := updateConnectorRequestBody(connector)
	if err != nil {
		return "", fwdiag.Diagnostics{fwdiag.NewErrorDiagnostic("Failed to create update request body", err.Error())}
	}

	resp, err := client.API.PutActionsConnectorIdWithResponse(ctx, connector.SpaceID, connector.ConnectorID, body)
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
		cr := ConnectorResponse{
			Config:           resp.JSON200.Config,
			ConnectorTypeID:  resp.JSON200.ConnectorTypeId,
			ID:               resp.JSON200.Id,
			IsDeprecated:     resp.JSON200.IsDeprecated,
			IsMissingSecrets: resp.JSON200.IsMissingSecrets,
			IsPreconfigured:  resp.JSON200.IsPreconfigured,
			Name:             resp.JSON200.Name,
		}
		return ConnectorResponseToModel(spaceID, &cr)
	case http.StatusNotFound:
		return nil, nil
	default:
		return nil, diagutil.ReportUnknownHTTPError(resp.StatusCode(), resp.Body)
	}
}

func SearchConnectors(ctx context.Context, client *Client, connectorName, spaceID, connectorTypeID string) ([]*models.KibanaActionConnector, fwdiag.Diagnostics) {
	resp, err := client.API.GetActionsConnectorsWithResponse(ctx, spaceID)
	if err != nil {
		return nil, fwdiag.Diagnostics{
			fwdiag.NewErrorDiagnostic("Unable to get connectors", err.Error()),
		}
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, diagutil.ReportUnknownHTTPError(resp.StatusCode(), resp.Body)
	}

	foundConnectors := []*models.KibanaActionConnector{}
	for _, connector := range *resp.JSON200 {
		if connector.Name != connectorName {
			continue
		}

		if connectorTypeID != "" && connector.ConnectorTypeId != connectorTypeID {
			continue
		}

		cr := ConnectorResponse{
			Config:           connector.Config,
			ConnectorTypeID:  connector.ConnectorTypeId,
			ID:               connector.Id,
			IsDeprecated:     connector.IsDeprecated,
			IsMissingSecrets: connector.IsMissingSecrets,
			IsPreconfigured:  connector.IsPreconfigured,
			Name:             connector.Name,
		}
		c, fwDiags := ConnectorResponseToModel(spaceID, &cr)
		if fwDiags.HasError() {
			return nil, fwDiags
		}

		foundConnectors = append(foundConnectors, c)
	}
	if len(foundConnectors) == 0 {
		tflog.Debug(ctx, fmt.Sprintf("no connectors found with name [%s/%s] and type [%s]", spaceID, connectorName, connectorTypeID))
	}

	return foundConnectors, nil
}

func ConnectorResponseToModel(spaceID string, connector *ConnectorResponse) (*models.KibanaActionConnector, fwdiag.Diagnostics) {
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
		handler, ok := connectorConfigHandlers[connector.ConnectorTypeID]
		if ok {
			configJSONString, err := handler.remarshalConfig(string(configJSON))
			if err != nil {
				return nil, fwdiag.Diagnostics{fwdiag.NewErrorDiagnostic("Failed to remarshal config", err.Error())}
			}

			configJSON = []byte(configJSONString)
		}
	}

	model := &models.KibanaActionConnector{
		ConnectorID:     connector.ID,
		SpaceID:         spaceID,
		Name:            connector.Name,
		ConfigJSON:      string(configJSON),
		ConnectorTypeID: connector.ConnectorTypeID,
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
		Config:          &kbapi.PostActionsConnectorIdJSONBody_Config{},
		Secrets:         &kbapi.PostActionsConnectorIdJSONBody_Secrets{},
	}

	if err := unmarshalConnectorFields(connector.ConfigJSON, connector.SecretsJSON, &req.Config.AdditionalProperties, &req.Secrets.AdditionalProperties); err != nil {
		return kbapi.PostActionsConnectorIdJSONRequestBody{}, err
	}

	return req, nil
}

func updateConnectorRequestBody(connector models.KibanaActionConnector) (kbapi.PutActionsConnectorIdJSONRequestBody, error) {
	req := kbapi.PutActionsConnectorIdJSONRequestBody{
		Name:    connector.Name,
		Config:  &kbapi.PutActionsConnectorIdJSONBody_Config{},
		Secrets: &kbapi.PutActionsConnectorIdJSONBody_Secrets{},
	}

	if err := unmarshalConnectorFields(connector.ConfigJSON, connector.SecretsJSON, &req.Config.AdditionalProperties, &req.Secrets.AdditionalProperties); err != nil {
		return kbapi.PutActionsConnectorIdJSONRequestBody{}, err
	}

	return req, nil
}
