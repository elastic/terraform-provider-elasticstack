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

package clients

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/disaster37/go-kibana-rest/v8"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/terraform-provider-elasticstack/generated/slo"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/config"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	kibanaoapi "github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/debugutils"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/hashicorp/go-version"
	fwdiags "github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/logging"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type CompositeID struct {
	ClusterID  string
	ResourceID string
}

const ServerlessFlavor = "serverless"

func CompositeIDFromStr(id string) (*CompositeID, diag.Diagnostics) {
	var diags diag.Diagnostics
	idParts := strings.Split(id, "/")
	if len(idParts) != 2 {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Wrong resource ID.",
			Detail:   "Resource ID must have following format: <cluster_uuid>/<resource identifier>",
		})
		return nil, diags
	}
	return &CompositeID{
			ClusterID:  idParts[0],
			ResourceID: idParts[1],
		},
		diags
}

func CompositeIDFromStrFw(id string) (*CompositeID, fwdiags.Diagnostics) {
	composite, diags := CompositeIDFromStr(id)
	return composite, diagutil.FrameworkDiagsFromSDK(diags)
}

func ResourceIDFromStr(id string) (string, diag.Diagnostics) {
	compID, diags := CompositeIDFromStr(id)
	if diags.HasError() {
		return "", diags
	}
	return compID.ResourceID, nil
}

func (c *CompositeID) String() string {
	return fmt.Sprintf("%s/%s", c.ClusterID, c.ResourceID)
}

// apiClient is the internal broad client that holds all configured service
// clients built from the provider configuration block. It is unexported;
// external code must use scoped clients (KibanaScopedClient,
// ElasticsearchScopedClient) obtained through ProviderClientFactory or the
// acceptance-testing helper constructors.
type apiClient struct {
	elasticsearch            *elasticsearch.Client
	elasticsearchClusterInfo *models.ClusterInfo
	kibana                   *kibana.Client
	kibanaOapi               *kibanaoapi.Client
	slo                      slo.SloAPI
	kibanaConfig             kibana.Config
	fleet                    *fleet.Client
	version                  string
}

func NewAPIClientFuncFromSDK(version string) func(context.Context, *schema.ResourceData) (any, diag.Diagnostics) {
	return func(_ context.Context, d *schema.ResourceData) (any, diag.Diagnostics) {
		client, diags := newAPIClientFromSDK(d, version)
		if diags.HasError() {
			return nil, diags
		}
		return NewProviderClientFactory(client), diags
	}
}

func newAcceptanceTestingClient() (*apiClient, error) {
	version := "tf-acceptance-testing"
	cfg := config.NewFromEnv(version)

	es, err := elasticsearch.NewClient(*cfg.Elasticsearch)
	if err != nil {
		return nil, err
	}

	kib, err := kibana.NewClient(*cfg.Kibana)
	if err != nil {
		return nil, err
	}

	kibanaHTTPClient := kib.Client.GetClient()

	kibOapi, err := kibanaoapi.NewClient(*cfg.KibanaOapi)
	if err != nil {
		return nil, err
	}

	fleetClient, err := fleet.NewClient(*cfg.Fleet)
	if err != nil {
		return nil, err
	}

	return &apiClient{
			elasticsearch: es,
			kibana:        kib,
			kibanaOapi:    kibOapi,
			slo:           buildSloClient(cfg, kibanaHTTPClient).SloAPI,
			kibanaConfig:  *cfg.Kibana,
			fleet:         fleetClient,
			version:       version,
		},
		nil
}

func newAPIClientFromFramework(ctx context.Context, cfg config.ProviderConfiguration, version string) (*apiClient, fwdiags.Diagnostics) {
	clientCfg, diags := config.NewFromFramework(ctx, cfg, version)
	if diags.HasError() {
		return nil, diags
	}

	client, err := newAPIClientFromConfig(clientCfg, version)
	if err != nil {
		return nil, fwdiags.Diagnostics{
			fwdiags.NewErrorDiagnostic("Failed to create API client", err.Error()),
		}
	}

	return client, nil
}

func (a *apiClient) GetESClient() (*elasticsearch.Client, error) {
	if a.elasticsearch == nil {
		return nil, errors.New("elasticsearch client not found")
	}

	return a.elasticsearch, nil
}

func (a *apiClient) GetKibanaClient() (*kibana.Client, error) {
	if a.kibana == nil {
		return nil, errors.New("kibana client not found")
	}

	return a.kibana, nil
}

func (a *apiClient) GetKibanaOapiClient() (*kibanaoapi.Client, error) {
	if a.kibanaOapi == nil {
		return nil, errors.New("kibanaoapi client not found")
	}

	return a.kibanaOapi, nil
}

func (a *apiClient) GetSloClient() (slo.SloAPI, error) {
	if a.slo == nil {
		return nil, errors.New("slo client not found")
	}

	return a.slo, nil
}

func (a *apiClient) GetFleetClient() (*fleet.Client, error) {
	if a.fleet == nil {
		return nil, errors.New("fleet client not found")
	}

	return a.fleet, nil
}

func (a *apiClient) SetSloAuthContext(ctx context.Context) context.Context {
	if a.kibanaConfig.ApiKey != "" {
		return context.WithValue(ctx, slo.ContextAPIKeys, map[string]slo.APIKey{
			"apiKeyAuth": {
				Prefix: "ApiKey",
				Key:    a.kibanaConfig.ApiKey,
			}})
	}

	return context.WithValue(ctx, slo.ContextBasicAuth, slo.BasicAuth{
		UserName: a.kibanaConfig.Username,
		Password: a.kibanaConfig.Password,
	})
}

func (a *apiClient) ID(ctx context.Context, resourceID string) (*CompositeID, diag.Diagnostics) {
	var diags diag.Diagnostics
	clusterID, diags := a.ClusterID(ctx)
	if diags.HasError() {
		return nil, diags
	}
	return &CompositeID{*clusterID, resourceID}, diags
}

func (a *apiClient) serverInfo(ctx context.Context) (*models.ClusterInfo, diag.Diagnostics) {
	if a.elasticsearchClusterInfo != nil {
		return a.elasticsearchClusterInfo, nil
	}

	var diags diag.Diagnostics
	esClient, err := a.GetESClient()
	if err != nil {
		return nil, diag.FromErr(err)
	}
	res, err := esClient.Info(esClient.Info.WithContext(ctx))
	if err != nil {
		return nil, diag.FromErr(err)
	}
	defer res.Body.Close()
	if diags := diagutil.CheckError(res, "Unable to connect to the Elasticsearch cluster"); diags.HasError() {
		return nil, diags
	}

	info := models.ClusterInfo{}
	if err := json.NewDecoder(res.Body).Decode(&info); err != nil {
		return nil, diag.FromErr(err)
	}
	// cache info
	a.elasticsearchClusterInfo = &info

	return &info, diags
}

func (a *apiClient) EnforceMinVersion(ctx context.Context, minVersion *version.Version) (bool, diag.Diagnostics) {
	flavor, diags := a.ServerFlavor(ctx)
	if diags.HasError() {
		return false, diags
	}

	if flavor == ServerlessFlavor {
		return true, nil
	}

	serverVersion, diags := a.ServerVersion(ctx)
	if diags.HasError() {
		return false, diags
	}

	return serverVersion.GreaterThanOrEqual(minVersion), nil
}

type MinVersionEnforceable interface {
	EnforceMinVersion(ctx context.Context, minVersion *version.Version) (bool, diag.Diagnostics)
}

func (a *apiClient) ServerVersion(ctx context.Context) (*version.Version, diag.Diagnostics) {
	if a.elasticsearch != nil {
		return a.versionFromElasticsearch(ctx)
	}

	return a.versionFromKibana()
}

func (a *apiClient) versionFromKibana() (*version.Version, diag.Diagnostics) {
	kibClient, err := a.GetKibanaClient()
	if err != nil {
		return nil, diag.Errorf("failed to get version from Kibana API: %s, "+
			"please ensure a working 'kibana' endpoint is configured", err.Error())
	}

	status, err := kibClient.KibanaStatus.Get()
	if err != nil {
		return nil, diag.Errorf("failed to get version from Kibana API: %s, "+
			"Please ensure a working 'kibana' endpoint is configured", err.Error())
	}

	vMap, ok := status["version"].(map[string]any)
	if !ok {
		return nil, diag.Errorf("failed to get version from Kibana API")
	}

	rawVersion, ok := vMap["number"].(string)
	if !ok {
		return nil, diag.Errorf("failed to get version number from Kibana status")
	}

	serverVersion, err := version.NewVersion(rawVersion)
	if err != nil {
		return nil, diag.FromErr(err)
	}

	return serverVersion, nil
}

func (a *apiClient) versionFromElasticsearch(ctx context.Context) (*version.Version, diag.Diagnostics) {
	info, diags := a.serverInfo(ctx)
	if diags.HasError() {
		return nil, diags
	}

	rawVersion := info.Version.Number
	serverVersion, err := version.NewVersion(rawVersion)
	if err != nil {
		return nil, diag.FromErr(err)
	}

	return serverVersion, nil
}

func (a *apiClient) ServerFlavor(ctx context.Context) (string, diag.Diagnostics) {
	if a.elasticsearch != nil {
		return a.flavorFromElasticsearch(ctx)
	}

	return a.flavorFromKibana()
}

func (a *apiClient) flavorFromElasticsearch(ctx context.Context) (string, diag.Diagnostics) {
	info, diags := a.serverInfo(ctx)
	if diags.HasError() {
		return "", diags
	}

	return info.Version.BuildFlavor, nil
}

func (a *apiClient) flavorFromKibana() (string, diag.Diagnostics) {
	kibClient, err := a.GetKibanaClient()
	if err != nil {
		return "", diag.Errorf("failed to get flavor from Kibana API: %s, "+
			"please ensure a working 'kibana' endpoint is configured", err.Error())
	}

	status, err := kibClient.KibanaStatus.Get()
	if err != nil {
		return "", diag.Errorf("failed to get flavor from Kibana API: %s, "+
			"Please ensure a working 'kibana' endpoint is configured", err.Error())
	}

	vMap, ok := status["version"].(map[string]any)
	if !ok {
		return "", diag.Errorf("failed to get flavor from Kibana API")
	}

	serverFlavor, ok := vMap["build_flavor"].(string)
	if !ok {
		// build_flavor field is not present in older Kibana versions (pre-serverless)
		// Default to empty string to indicate traditional/stateful deployment
		return "", nil
	}

	return serverFlavor, nil
}

func (a *apiClient) ClusterID(ctx context.Context) (*string, diag.Diagnostics) {
	info, diags := a.serverInfo(ctx)
	if diags.HasError() {
		return nil, diags
	}

	if uuid := info.ClusterUUID; uuid != "" && uuid != "_na_" {
		tflog.Trace(ctx, fmt.Sprintf("cluster UUID: %s", uuid))
		return &uuid, diags
	}

	diags = append(diags, diag.Diagnostic{
		Severity: diag.Error,
		Summary:  "Unable to get cluster UUID",
		Detail: `Unable to get cluster UUID.
		There might be a problem with permissions or cluster is still starting up and UUID has not been populated yet.`,
	})
	return nil, diags
}

func buildEsClient(cfg config.Client) (*elasticsearch.Client, error) {
	if cfg.Elasticsearch == nil {
		return nil, nil
	}

	es, err := elasticsearch.NewClient(*cfg.Elasticsearch)
	if err != nil {
		return nil, fmt.Errorf("unable to create Elasticsearch client: %w", err)
	}

	return es, nil
}

func buildKibanaClient(cfg config.Client) (*kibana.Client, error) {
	if cfg.Kibana == nil {
		return nil, nil
	}

	kib, err := kibana.NewClient(*cfg.Kibana)

	if err != nil {
		return nil, err
	}

	if logging.IsDebugOrHigher() {
		// It is required to set debug mode even if we re-use the http client within the OpenAPI generated clients
		// some of the clients are not relying on the OpenAPI generated clients and are using the http client directly
		kib.Client.SetDebug(true)
		transport, err := kib.Client.Transport()
		if err != nil {
			return nil, err
		}
		roundTripper := debugutils.NewDebugTransport("Kibana", transport)
		kib.Client.SetTransport(roundTripper)
	}

	return kib, nil
}

func buildKibanaOapiClient(cfg config.Client) (*kibanaoapi.Client, error) {
	client, err := kibanaoapi.NewClient(*cfg.KibanaOapi)
	if err != nil {
		return nil, fmt.Errorf("unable to create KibanaOapi client: %w", err)
	}

	return client, nil
}

func buildSloClient(cfg config.Client, httpClient *http.Client) *slo.APIClient {
	sloConfig := slo.Configuration{
		Debug:     logging.IsDebugOrHigher(),
		UserAgent: cfg.UserAgent,
		Servers: slo.ServerConfigurations{
			{
				URL: cfg.Kibana.Address,
			},
		},
		HTTPClient: httpClient,
	}
	return slo.NewAPIClient(&sloConfig)
}

func buildFleetClient(cfg config.Client) (*fleet.Client, error) {
	client, err := fleet.NewClient(*cfg.Fleet)
	if err != nil {
		return nil, fmt.Errorf("unable to create Fleet client: %w", err)
	}

	return client, nil
}

func newAPIClientFromSDK(d *schema.ResourceData, version string) (*apiClient, diag.Diagnostics) {
	cfg, diags := config.NewFromSDK(d, version)
	if diags.HasError() {
		return nil, diags
	}

	client, err := newAPIClientFromConfig(cfg, version)
	if err != nil {
		return nil, diag.FromErr(err)
	}

	return client, nil
}

func newAPIClientFromConfig(cfg config.Client, version string) (*apiClient, error) {
	var kibanaConfig kibana.Config
	if cfg.Kibana != nil {
		kibanaConfig = *cfg.Kibana
	}
	client := &apiClient{
		kibanaConfig: kibanaConfig,
		version:      version,
	}

	if cfg.Elasticsearch != nil {
		esClient, err := buildEsClient(cfg)
		if err != nil {
			return nil, err
		}
		client.elasticsearch = esClient
	}

	if cfg.Kibana != nil {
		kibanaClient, err := buildKibanaClient(cfg)
		if err != nil {
			return nil, err
		}
		client.kibana = kibanaClient

		kibanaOapiClient, err := buildKibanaOapiClient(cfg)
		if err != nil {
			return nil, err
		}
		client.kibanaOapi = kibanaOapiClient

		kibanaHTTPClient := kibanaClient.Client.GetClient()

		client.slo = buildSloClient(cfg, kibanaHTTPClient).SloAPI
	}

	if cfg.Fleet != nil {
		fleetClient, err := buildFleetClient(cfg)
		if err != nil {
			return nil, err
		}

		client.fleet = fleetClient
	}

	return client, nil
}
