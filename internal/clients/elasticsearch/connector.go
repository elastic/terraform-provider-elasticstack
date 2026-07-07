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
	"encoding/json"

	getconnector "github.com/elastic/go-elasticsearch/v9/typedapi/connector/get"
	"github.com/elastic/go-elasticsearch/v9/typedapi/connector/post"
	"github.com/elastic/go-elasticsearch/v9/typedapi/connector/put"
	"github.com/elastic/go-elasticsearch/v9/typedapi/connector/syncjobget"
	"github.com/elastic/go-elasticsearch/v9/typedapi/connector/syncjobpost"
	"github.com/elastic/go-elasticsearch/v9/typedapi/connector/updateapikeyid"
	"github.com/elastic/go-elasticsearch/v9/typedapi/connector/updateconfiguration"
	"github.com/elastic/go-elasticsearch/v9/typedapi/connector/updatefeatures"
	"github.com/elastic/go-elasticsearch/v9/typedapi/connector/updateindexname"
	"github.com/elastic/go-elasticsearch/v9/typedapi/connector/updatename"
	"github.com/elastic/go-elasticsearch/v9/typedapi/connector/updatenative"
	"github.com/elastic/go-elasticsearch/v9/typedapi/connector/updatepipeline"
	"github.com/elastic/go-elasticsearch/v9/typedapi/connector/updatescheduling"
	"github.com/elastic/go-elasticsearch/v9/typedapi/connector/updateservicetype"
	"github.com/elastic/go-elasticsearch/v9/typedapi/types"
	"github.com/elastic/go-elasticsearch/v9/typedapi/types/enums/syncjobtriggermethod"
	"github.com/elastic/go-elasticsearch/v9/typedapi/types/enums/syncjobtype"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	fwdiag "github.com/hashicorp/terraform-plugin-framework/diag"
)

// CreateConnectorBody is the narrow body for POST /_connector or PUT /_connector/{id}.
// It carries only the envelope fields the connector create API accepts directly;
// pipeline, scheduling, features, configuration, and api_key_id are written via
// the corresponding partial-update wrappers (UpdateConnectorPipeline, etc.).
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

// UpdateConnectorName updates the connector name and description.
func UpdateConnectorName(
	ctx context.Context,
	apiClient *clients.ElasticsearchScopedClient,
	connectorID string,
	name, description *string,
) fwdiag.Diagnostics {
	typedClient := apiClient.GetESClient()

	_, err := typedClient.Connector.UpdateName(connectorID).Request(&updatename.Request{
		Name:        name,
		Description: description,
	}).Do(ctx)
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}

	return nil
}

// UpdateConnectorIndexName updates the connector index name.
func UpdateConnectorIndexName(
	ctx context.Context,
	apiClient *clients.ElasticsearchScopedClient,
	connectorID string,
	indexName *string,
) fwdiag.Diagnostics {
	typedClient := apiClient.GetESClient()

	_, err := typedClient.Connector.UpdateIndexName(connectorID).Request(&updateindexname.Request{
		IndexName: indexName,
	}).Do(ctx)
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}

	return nil
}

// UpdateConnectorServiceType updates the connector service type.
func UpdateConnectorServiceType(
	ctx context.Context,
	apiClient *clients.ElasticsearchScopedClient,
	connectorID string,
	serviceType string,
) fwdiag.Diagnostics {
	typedClient := apiClient.GetESClient()

	_, err := typedClient.Connector.UpdateServiceType(connectorID).Request(&updateservicetype.Request{
		ServiceType: serviceType,
	}).Do(ctx)
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}

	return nil
}

// UpdateConnectorNative updates the connector is_native flag.
func UpdateConnectorNative(
	ctx context.Context,
	apiClient *clients.ElasticsearchScopedClient,
	connectorID string,
	isNative bool,
) fwdiag.Diagnostics {
	typedClient := apiClient.GetESClient()

	_, err := typedClient.Connector.UpdateNative(connectorID).Request(&updatenative.Request{
		IsNative: isNative,
	}).Do(ctx)
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}

	return nil
}

// UpdateConnectorPipeline replaces the connector's pipeline configuration via
// PUT /_connector/{id}/_pipeline. The upstream API performs a full replace,
// so callers must pass an API-complete IngestPipelineParams; sparse values
// will overwrite server state with empty fields.
func UpdateConnectorPipeline(
	ctx context.Context,
	apiClient *clients.ElasticsearchScopedClient,
	connectorID string,
	pipeline types.IngestPipelineParams,
) fwdiag.Diagnostics {
	typedClient := apiClient.GetESClient()

	_, err := typedClient.Connector.UpdatePipeline(connectorID).Request(&updatepipeline.Request{
		Pipeline: pipeline,
	}).Do(ctx)
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}

	return nil
}

// UpdateConnectorScheduling replaces the connector's scheduling configuration via
// PUT /_connector/{id}/_scheduling. The upstream API performs a full replace,
// so callers must pass an API-complete SchedulingConfiguration; sparse values
// will overwrite server state with empty fields.
func UpdateConnectorScheduling(
	ctx context.Context,
	apiClient *clients.ElasticsearchScopedClient,
	connectorID string,
	scheduling types.SchedulingConfiguration,
) fwdiag.Diagnostics {
	typedClient := apiClient.GetESClient()

	_, err := typedClient.Connector.UpdateScheduling(connectorID).Request(&updatescheduling.Request{
		Scheduling: scheduling,
	}).Do(ctx)
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}

	return nil
}

// UpdateConnectorFeatures replaces the connector's features configuration via
// PUT /_connector/{id}/_features. The upstream API performs a full replace,
// so callers must pass an API-complete ConnectorFeatures; sparse values
// will overwrite server state with empty fields.
func UpdateConnectorFeatures(
	ctx context.Context,
	apiClient *clients.ElasticsearchScopedClient,
	connectorID string,
	features types.ConnectorFeatures,
) fwdiag.Diagnostics {
	typedClient := apiClient.GetESClient()

	_, err := typedClient.Connector.UpdateFeatures(connectorID).Request(&updatefeatures.Request{
		Features: features,
	}).Do(ctx)
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}

	return nil
}

// UpdateConnectorAPIKeyID updates the connector API key id and secret id.
func UpdateConnectorAPIKeyID(
	ctx context.Context,
	apiClient *clients.ElasticsearchScopedClient,
	connectorID string,
	apiKeyID, apiKeySecretID *string,
) fwdiag.Diagnostics {
	typedClient := apiClient.GetESClient()

	_, err := typedClient.Connector.UpdateApiKeyId(connectorID).Request(&updateapikeyid.Request{
		ApiKeyId:       apiKeyID,
		ApiKeySecretId: apiKeySecretID,
	}).Do(ctx)
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}

	return nil
}

// UpdateConnectorConfiguration updates connector configuration values.
func UpdateConnectorConfiguration(
	ctx context.Context,
	apiClient *clients.ElasticsearchScopedClient,
	connectorID string,
	values map[string]json.RawMessage,
) fwdiag.Diagnostics {
	typedClient := apiClient.GetESClient()

	_, err := typedClient.Connector.UpdateConfiguration(connectorID).Request(&updateconfiguration.Request{
		Values: values,
	}).Do(ctx)
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}

	return nil
}

// CreateSyncJob creates a sync job for the given connector and returns its id.
// jobType and triggerMethod must be one of the canonical enum constants from
// the syncjobtype and syncjobtriggermethod packages (e.g. syncjobtype.Full,
// syncjobtriggermethod.Ondemand). The wrapper does not apply defaults;
// the action layer is responsible for that per REQ-SYNC-001.
func CreateSyncJob(
	ctx context.Context,
	apiClient *clients.ElasticsearchScopedClient,
	connectorID string,
	jobType syncjobtype.SyncJobType,
	triggerMethod syncjobtriggermethod.SyncJobTriggerMethod,
) (string, fwdiag.Diagnostics) {
	typedClient := apiClient.GetESClient()

	res, err := typedClient.Connector.SyncJobPost().Request(&syncjobpost.Request{
		Id:            connectorID,
		JobType:       &jobType,
		TriggerMethod: &triggerMethod,
	}).Do(ctx)
	if err != nil {
		return "", diagutil.FrameworkDiagFromError(err)
	}

	return res.Id, nil
}

// GetSyncJob returns the sync job document for polling.
func GetSyncJob(
	ctx context.Context,
	apiClient *clients.ElasticsearchScopedClient,
	syncJobID string,
) (*syncjobget.Response, fwdiag.Diagnostics) {
	typedClient := apiClient.GetESClient()

	res, err := typedClient.Connector.SyncJobGet(syncJobID).Do(ctx)
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	return res, nil
}
