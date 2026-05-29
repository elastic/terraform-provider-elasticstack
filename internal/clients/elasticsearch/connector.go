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

package elasticsearch

import (
	"context"

	getconnector "github.com/elastic/go-elasticsearch/v8/typedapi/connector/get"
	"github.com/elastic/go-elasticsearch/v8/typedapi/connector/post"
	"github.com/elastic/go-elasticsearch/v8/typedapi/connector/put"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	fwdiag "github.com/hashicorp/terraform-plugin-framework/diag"
)

// CreateConnectorBody holds the narrow create/update envelope fields for POST /_connector
// or PUT /_connector/{id}. All fields are optional except ServiceType.
type CreateConnectorBody struct {
	Name        *string
	Description *string
	IndexName   *string
	IsNative    *bool
	Language    *string
	ServiceType string
}

func createConnectorRequest(body CreateConnectorBody) *post.Request {
	serviceType := body.ServiceType
	return &post.Request{
		Name:        body.Name,
		Description: body.Description,
		IndexName:   body.IndexName,
		IsNative:    body.IsNative,
		Language:    body.Language,
		ServiceType: &serviceType,
	}
}

func createConnectorPutRequest(body CreateConnectorBody) *put.Request {
	serviceType := body.ServiceType
	return &put.Request{
		Name:        body.Name,
		Description: body.Description,
		IndexName:   body.IndexName,
		IsNative:    body.IsNative,
		Language:    body.Language,
		ServiceType: &serviceType,
	}
}

// CreateConnector creates a new connector. When connectorID is non-empty, it
// uses PUT /_connector/{id}; otherwise it uses POST /_connector and returns
// the auto-generated id. Returns the connector id on success.
func CreateConnector(
	ctx context.Context,
	apiClient *clients.ElasticsearchScopedClient,
	connectorID string,
	body CreateConnectorBody,
) (string, fwdiag.Diagnostics) {
	typedClient := apiClient.GetESClient()

	if connectorID != "" {
		res, err := typedClient.Connector.Put().ConnectorId(connectorID).Request(createConnectorPutRequest(body)).Do(ctx)
		if err != nil {
			return "", diagutil.FrameworkDiagFromError(err)
		}
		return res.Id, nil
	}

	res, err := typedClient.Connector.Post().Request(createConnectorRequest(body)).Do(ctx)
	if err != nil {
		return "", diagutil.FrameworkDiagFromError(err)
	}
	return res.Id, nil
}

// GetConnector returns the connector by id. Returns nil, nil on 404.
func GetConnector(
	ctx context.Context,
	apiClient *clients.ElasticsearchScopedClient,
	connectorID string,
) (*getconnector.Response, fwdiag.Diagnostics) {
	typedClient := apiClient.GetESClient()

	res, err := typedClient.Connector.Get(connectorID).Do(ctx)
	if err != nil {
		if IsNotFoundElasticsearchError(err) {
			return nil, nil
		}
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	return res, nil
}

// DeleteConnector deletes the connector by id. Returns nil on 404 (idempotent).
func DeleteConnector(
	ctx context.Context,
	apiClient *clients.ElasticsearchScopedClient,
	connectorID string,
) fwdiag.Diagnostics {
	typedClient := apiClient.GetESClient()

	_, err := typedClient.Connector.Delete(connectorID).Do(ctx)
	if err != nil {
		if IsNotFoundElasticsearchError(err) {
			return nil
		}
		return diagutil.FrameworkDiagFromError(err)
	}

	return nil
}
