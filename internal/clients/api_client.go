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
	"github.com/elastic/terraform-provider-elasticstack/generated/alerting"
	"github.com/elastic/terraform-provider-elasticstack/generated/connectors"
	"github.com/elastic/terraform-provider-elasticstack/generated/slo"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/config"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibana_oapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/go-version"
	fwdiags "github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/logging"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/oapi-codegen/oapi-codegen/v2/pkg/securityprovider"
)

type CompositeId struct {
	ClusterId  string
	ResourceId string
}

const ServerlessFlavor = "serverless"

func CompositeIdFromStr(id string) (*CompositeId, diag.Diagnostics) {
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
	return &CompositeId{
			ClusterId:  idParts[0],
			ResourceId: idParts[1],
		},
		diags
}

func CompositeIdFromStrFw(id string) (*CompositeId, fwdiags.Diagnostics) {
	composite, diags := CompositeIdFromStr(id)
	return composite, utils.FrameworkDiagsFromSDK(diags)
}

func ResourceIDFromStr(id string) (string, diag.Diagnostics) {
	compID, diags := CompositeIdFromStr(id)
	if diags.HasError() {
		return "", diags
	}
	return compID.ResourceId, nil
}

func (c *CompositeId) String() string {
	return fmt.Sprintf("%s/%s", c.ClusterId, c.ResourceId)
}

type ApiClient struct {
	elasticsearch            *elasticsearch.Client
	elasticsearchClusterInfo *models.ClusterInfo
	kibana                   *kibana.Client
	kibanaOapi               *kibana_oapi.Client
	alerting                 alerting.AlertingAPI
	connectors               *connectors.Client
	slo                      slo.SloAPI
	kibanaConfig             kibana.Config
	fleet                    *fleet.Client
	version                  string
}

func NewApiClientFuncFromSDK(version string) func(context.Context, *schema.ResourceData) (interface{}, diag.Diagnostics) {
	return func(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
		return newApiClientFromSDK(d, version)
	}
}

func NewAcceptanceTestingClient() (*ApiClient, error) {
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

	kibanaHttpClient := kib.Client.GetClient()

	actionConnectors, err := buildConnectorsClient(cfg, kibanaHttpClient)
	if err != nil {
		return nil, fmt.Errorf("cannot create Kibana action connectors client: [%w]", err)
	}

	kibOapi, err := kibana_oapi.NewClient(*cfg.KibanaOapi)
	if err != nil {
		return nil, err
	}

	fleetClient, err := fleet.NewClient(*cfg.Fleet)
	if err != nil {
		return nil, err
	}

	return &ApiClient{
			elasticsearch: es,
			kibana:        kib,
			kibanaOapi:    kibOapi,
			alerting:      buildAlertingClient(cfg, kibanaHttpClient).AlertingAPI,
			slo:           buildSloClient(cfg, kibanaHttpClient).SloAPI,
			connectors:    actionConnectors,
			kibanaConfig:  *cfg.Kibana,
			fleet:         fleetClient,
			version:       version,
		},
		nil
}

func NewApiClientFromFramework(ctx context.Context, cfg config.ProviderConfiguration, version string) (*ApiClient, fwdiags.Diagnostics) {
	clientCfg, diags := config.NewFromFramework(ctx, cfg, version)
	if diags.HasError() {
		return nil, diags
	}

	client, err := newApiClientFromConfig(clientCfg, version)
	if err != nil {
		return nil, fwdiags.Diagnostics{
			fwdiags.NewErrorDiagnostic("Failed to create API client", err.Error()),
		}
	}

	return client, nil
}

func ConvertProviderData(providerData any) (*ApiClient, fwdiags.Diagnostics) {
	var diags fwdiags.Diagnostics

	if providerData == nil {
		return nil, diags
	}

	client, ok := providerData.(*ApiClient)
	if !ok {
		diags.AddError(
			"Unexpected Provider Data",
			fmt.Sprintf("Expected *ApiClient, got: %T. Please report this issue to the provider developers.", providerData),
		)

		return nil, diags
	}
	if client == nil {
		diags.AddError(
			"Unconfigured Client",
			"Expected configured client. Please report this issue to the provider developers.",
		)
	}
	return client, diags
}

func MaybeNewApiClientFromFrameworkResource(ctx context.Context, esConnList types.List, defaultClient *ApiClient) (*ApiClient, fwdiags.Diagnostics) {
	var esConns []config.ElasticsearchConnection
	if diags := esConnList.ElementsAs(ctx, &esConns, true); diags.HasError() {
		return nil, diags
	}

	if len(esConns) == 0 {
		return defaultClient, nil
	}

	cfg, diags := config.NewFromFramework(ctx, config.ProviderConfiguration{Elasticsearch: esConns}, defaultClient.version)
	if diags.HasError() {
		return nil, diags
	}

	esClient, err := buildEsClient(cfg)
	if err != nil {
		return nil, fwdiags.Diagnostics{fwdiags.NewErrorDiagnostic(err.Error(), err.Error())}
	}

	return &ApiClient{
		elasticsearch:            esClient,
		elasticsearchClusterInfo: defaultClient.elasticsearchClusterInfo,
		kibana:                   defaultClient.kibana,
		fleet:                    defaultClient.fleet,
		version:                  defaultClient.version,
	}, diags
}

func NewApiClientFromSDKResource(d *schema.ResourceData, meta interface{}) (*ApiClient, diag.Diagnostics) {
	defaultClient := meta.(*ApiClient)
	version := defaultClient.version
	resourceConfig, diags := config.NewFromSDKResource(d, version)
	if diags.HasError() {
		return nil, diags
	}

	if resourceConfig == nil {
		return defaultClient, nil
	}

	esClient, err := buildEsClient(*resourceConfig)
	if err != nil {
		return nil, diag.FromErr(err)
	}

	return &ApiClient{
		elasticsearch:            esClient,
		elasticsearchClusterInfo: defaultClient.elasticsearchClusterInfo,
		kibana:                   defaultClient.kibana,
		fleet:                    defaultClient.fleet,
		version:                  version,
	}, diags
}

func (a *ApiClient) GetESClient() (*elasticsearch.Client, error) {
	if a.elasticsearch == nil {
		return nil, errors.New("elasticsearch client not found")
	}

	return a.elasticsearch, nil
}

func (a *ApiClient) GetKibanaClient() (*kibana.Client, error) {
	if a.kibana == nil {
		return nil, errors.New("kibana client not found")
	}

	return a.kibana, nil
}

func (a *ApiClient) GetKibanaOapiClient() (*kibana_oapi.Client, error) {
	if a.kibanaOapi == nil {
		return nil, errors.New("kibana_oapi client not found")
	}

	return a.kibanaOapi, nil
}

func (a *ApiClient) GetAlertingClient() (alerting.AlertingAPI, error) {
	if a.alerting == nil {
		return nil, errors.New("alerting client not found")
	}

	return a.alerting, nil
}

func (a *ApiClient) GetKibanaConnectorsClient(ctx context.Context) (*connectors.Client, error) {
	if a.connectors == nil {
		return nil, errors.New("kibana action connector client not found")
	}

	return a.connectors, nil
}

func (a *ApiClient) GetSloClient() (slo.SloAPI, error) {
	if a.slo == nil {
		return nil, errors.New("slo client not found")
	}

	return a.slo, nil
}

func (a *ApiClient) GetFleetClient() (*fleet.Client, error) {
	if a.fleet == nil {
		return nil, errors.New("fleet client not found")
	}

	return a.fleet, nil
}

func (a *ApiClient) SetSloAuthContext(ctx context.Context) context.Context {
	if a.kibanaConfig.ApiKey != "" {
		return context.WithValue(ctx, slo.ContextAPIKeys, map[string]slo.APIKey{
			"apiKeyAuth": {
				Prefix: "ApiKey",
				Key:    a.kibanaConfig.ApiKey,
			}})
	} else {
		return context.WithValue(ctx, slo.ContextBasicAuth, slo.BasicAuth{
			UserName: a.kibanaConfig.Username,
			Password: a.kibanaConfig.Password,
		})
	}
}

func (a *ApiClient) SetAlertingAuthContext(ctx context.Context) context.Context {
	if a.kibanaConfig.ApiKey != "" {
		return context.WithValue(ctx, alerting.ContextAPIKeys, map[string]alerting.APIKey{
			"apiKeyAuth": {
				Prefix: "ApiKey",
				Key:    a.kibanaConfig.ApiKey,
			}})
	} else {
		return context.WithValue(ctx, alerting.ContextBasicAuth, alerting.BasicAuth{
			UserName: a.kibanaConfig.Username,
			Password: a.kibanaConfig.Password,
		})
	}
}

func (a *ApiClient) ID(ctx context.Context, resourceId string) (*CompositeId, diag.Diagnostics) {
	var diags diag.Diagnostics
	clusterId, diags := a.ClusterID(ctx)
	if diags.HasError() {
		return nil, diags
	}
	return &CompositeId{*clusterId, resourceId}, diags
}

func (a *ApiClient) serverInfo(ctx context.Context) (*models.ClusterInfo, diag.Diagnostics) {
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
	if diags := utils.CheckError(res, "Unable to connect to the Elasticsearch cluster"); diags.HasError() {
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

func (a *ApiClient) EnforceMinVersion(ctx context.Context, minVersion *version.Version) (bool, diag.Diagnostics) {
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

func (a *ApiClient) ServerVersion(ctx context.Context) (*version.Version, diag.Diagnostics) {
	if a.elasticsearch != nil {
		return a.versionFromElasticsearch(ctx)
	}

	return a.versionFromKibana()
}

func (a *ApiClient) versionFromKibana() (*version.Version, diag.Diagnostics) {
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

	vMap, ok := status["version"].(map[string]interface{})
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

func (a *ApiClient) versionFromElasticsearch(ctx context.Context) (*version.Version, diag.Diagnostics) {
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

func (a *ApiClient) ServerFlavor(ctx context.Context) (string, diag.Diagnostics) {
	if a.elasticsearch != nil {
		return a.flavorFromElasticsearch(ctx)
	}

	return a.flavorFromKibana()
}

func (a *ApiClient) flavorFromElasticsearch(ctx context.Context) (string, diag.Diagnostics) {
	info, diags := a.serverInfo(ctx)
	if diags.HasError() {
		return "", diags
	}

	return info.Version.BuildFlavor, nil
}

func (a *ApiClient) flavorFromKibana() (string, diag.Diagnostics) {
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

	vMap, ok := status["version"].(map[string]interface{})
	if !ok {
		return "", diag.Errorf("failed to get flavor from Kibana API")
	}

	serverFlavor, ok := vMap["build_flavor"].(string)
	if !ok {
		return "", diag.Errorf("failed to get build flavor from Kibana status")
	}

	return serverFlavor, nil
}

func (a *ApiClient) ClusterID(ctx context.Context) (*string, diag.Diagnostics) {
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
		var roundTripper http.RoundTripper = utils.NewDebugTransport("Kibana", transport)
		kib.Client.SetTransport(roundTripper)
	}

	return kib, nil
}

func buildKibanaOapiClient(cfg config.Client) (*kibana_oapi.Client, error) {
	client, err := kibana_oapi.NewClient(*cfg.KibanaOapi)
	if err != nil {
		return nil, fmt.Errorf("unable to create KibanaOapi client: %w", err)
	}

	return client, nil
}

func buildAlertingClient(cfg config.Client, httpClient *http.Client) *alerting.APIClient {
	alertingConfig := alerting.Configuration{
		Debug:     logging.IsDebugOrHigher(),
		UserAgent: cfg.UserAgent,
		Servers: alerting.ServerConfigurations{
			{
				URL: cfg.Kibana.Address,
			},
		},
		HTTPClient: httpClient,
	}
	return alerting.NewAPIClient(&alertingConfig)
}

func buildConnectorsClient(cfg config.Client, httpClient *http.Client) (*connectors.Client, error) {
	var authInterceptor connectors.ClientOption
	if cfg.Kibana.ApiKey != "" {
		apiKeyProvider, err := securityprovider.NewSecurityProviderApiKey(
			"header",
			"Authorization",
			"ApiKey "+cfg.Kibana.ApiKey,
		)
		if err != nil {
			return nil, fmt.Errorf("unable to create api key auth provider: %w", err)
		}
		authInterceptor = connectors.WithRequestEditorFn(apiKeyProvider.Intercept)
	} else {
		basicAuthProvider, err := securityprovider.NewSecurityProviderBasicAuth(cfg.Kibana.Username, cfg.Kibana.Password)
		if err != nil {
			return nil, fmt.Errorf("unable to create basic auth provider: %w", err)
		}
		authInterceptor = connectors.WithRequestEditorFn(basicAuthProvider.Intercept)
	}

	return connectors.NewClient(
		cfg.Kibana.Address,
		authInterceptor,
		connectors.WithHTTPClient(httpClient),
	)
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

func newApiClientFromSDK(d *schema.ResourceData, version string) (*ApiClient, diag.Diagnostics) {
	cfg, diags := config.NewFromSDK(d, version)
	if diags.HasError() {
		return nil, diags
	}

	client, err := newApiClientFromConfig(cfg, version)
	if err != nil {
		return nil, diag.FromErr(err)
	}

	return client, nil
}

func newApiClientFromConfig(cfg config.Client, version string) (*ApiClient, error) {
	client := &ApiClient{
		kibanaConfig: *cfg.Kibana,
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

		kibanaHttpClient := kibanaClient.Client.GetClient()
		connectorsClient, err := buildConnectorsClient(cfg, kibanaHttpClient)
		if err != nil {
			return nil, fmt.Errorf("cannot create Kibana connectors client: [%w]", err)
		}

		client.alerting = buildAlertingClient(cfg, kibanaHttpClient).AlertingAPI
		client.slo = buildSloClient(cfg, kibanaHttpClient).SloAPI
		client.connectors = connectorsClient
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
