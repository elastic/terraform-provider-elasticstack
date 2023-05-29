package clients

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/deepmap/oapi-codegen/pkg/securityprovider"
	"github.com/disaster37/go-kibana-rest/v8"
	"github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/terraform-provider-elasticstack/generated/alerting"
	"github.com/elastic/terraform-provider-elasticstack/generated/connectors"
	"github.com/elastic/terraform-provider-elasticstack/generated/slo"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/logging"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type CompositeId struct {
	ClusterId  string
	ResourceId string
}

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
	alerting                 alerting.AlertingApi
	connectors               *connectors.Client
	slo                      slo.SlosApi
	kibanaConfig             kibana.Config
	fleet                    *fleet.Client
	version                  string
}

func NewApiClientFunc(version string) func(context.Context, *schema.ResourceData) (interface{}, diag.Diagnostics) {
	return func(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
		return newApiClient(d, version)
	}
}

func NewAcceptanceTestingClient() (*ApiClient, error) {
	ua := buildUserAgent("tf-acceptance-testing")
	baseConfig := BaseConfig{
		UserAgent: ua,
		Header:    http.Header{"User-Agent": []string{ua}},
		Username:  os.Getenv("ELASTICSEARCH_USERNAME"),
		Password:  os.Getenv("ELASTICSEARCH_PASSWORD"),
	}

	buildEsAccClient := func() (*elasticsearch.Client, error) {
		config := elasticsearch.Config{
			Header: baseConfig.Header,
		}

		if apiKey := os.Getenv("ELASTICSEARCH_API_KEY"); apiKey != "" {
			config.APIKey = apiKey
		} else {
			config.Username = baseConfig.Username
			config.Password = baseConfig.Password
		}

		if es := os.Getenv("ELASTICSEARCH_ENDPOINTS"); es != "" {
			endpoints := make([]string, 0)
			for _, e := range strings.Split(es, ",") {
				endpoints = append(endpoints, strings.TrimSpace(e))
			}
			config.Addresses = endpoints
		}

		if insecure := os.Getenv("ELASTICSEARCH_INSECURE"); insecure != "" {
			if insecureValue, _ := strconv.ParseBool(insecure); insecureValue {
				tlsClientConfig := ensureTLSClientConfig(&config)
				tlsClientConfig.InsecureSkipVerify = true
			}
		}

		return elasticsearch.NewClient(config)
	}

	kibanaConfig := kibana.Config{
		Username: baseConfig.Username,
		Password: baseConfig.Password,
		Address:  os.Getenv("KIBANA_ENDPOINT"),
	}
	if insecure := os.Getenv("KIBANA_INSECURE"); insecure != "" {
		if insecureValue, _ := strconv.ParseBool(insecure); insecureValue {
			kibanaConfig.DisableVerifySSL = true
		}
	}

	es, err := buildEsAccClient()
	if err != nil {
		return nil, err
	}

	kib, err := kibana.NewClient(kibanaConfig)
	if err != nil {
		return nil, err
	}

	actionConnectors, err := buildConnectorsClient(baseConfig, kibanaConfig)
	if err != nil {
		return nil, fmt.Errorf("cannot create Kibana action connectors client: [%w]", err)
	}
	fleetCfg := fleet.Config{
		URL:      kibanaConfig.Address,
		Username: kibanaConfig.Username,
		Password: kibanaConfig.Password,
		APIKey:   os.Getenv("FLEET_API_KEY"),
		Insecure: kibanaConfig.DisableVerifySSL,
	}
	if v := os.Getenv("FLEET_CA_CERTS"); v != "" {
		fleetCfg.CACerts = strings.Split(os.Getenv("FLEET_CA_CERTS"), ",")
	}

	fleetClient, err := fleet.NewClient(fleetCfg)
	if err != nil {
		return nil, err
	}

	return &ApiClient{
			elasticsearch: es,
			kibana:        kib,
			alerting:      buildAlertingClient(baseConfig, kibanaConfig).AlertingApi,
			slo:           buildSloClient(baseConfig, kibanaConfig).SlosApi,
			connectors:    actionConnectors,
			kibanaConfig:  kibanaConfig,
			fleet:         fleetClient,
			version:       "acceptance-testing",
		},
		nil
}

const esConnectionKey string = "elasticsearch_connection"

func NewApiClient(d *schema.ResourceData, meta interface{}) (*ApiClient, diag.Diagnostics) {
	defaultClient := meta.(*ApiClient)

	if _, ok := d.GetOk(esConnectionKey); !ok {
		return defaultClient, nil
	}

	version := defaultClient.version
	baseConfig := buildBaseConfig(d, version, esConnectionKey)

	esClient, diags := buildEsClient(d, baseConfig, false, esConnectionKey)
	if diags.HasError() {
		return nil, diags
	}

	return &ApiClient{
		elasticsearch:            esClient,
		elasticsearchClusterInfo: defaultClient.elasticsearchClusterInfo,
		kibana:                   defaultClient.kibana,
		fleet:                    defaultClient.fleet,
		version:                  version,
	}, diags
}

func ensureTLSClientConfig(config *elasticsearch.Config) *tls.Config {
	if config.Transport == nil {
		config.Transport = http.DefaultTransport.(*http.Transport)
	}
	if config.Transport.(*http.Transport).TLSClientConfig == nil {
		config.Transport.(*http.Transport).TLSClientConfig = &tls.Config{}
	}
	return config.Transport.(*http.Transport).TLSClientConfig
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

func (a *ApiClient) GetAlertingClient() (alerting.AlertingApi, error) {
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

func (a *ApiClient) GetSloClient() (slo.SlosApi, error) {
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

func (a *ApiClient) SetGeneratedClientAuthContext(ctx context.Context) context.Context {
	//I don't like that I'm using "alerting" here when the context is used for more than just alerting -- worth pulling these structs out somewhere else?
	return context.WithValue(ctx, alerting.ContextBasicAuth, alerting.BasicAuth{
		UserName: a.kibanaConfig.Username,
		Password: a.kibanaConfig.Password,
	})
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

func (a *ApiClient) ServerVersion(ctx context.Context) (*version.Version, diag.Diagnostics) {
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

type BaseConfig struct {
	Username  string
	Password  string
	UserAgent string
	Header    http.Header
}

// Build base config from ES which can be shared for other resources
func buildBaseConfig(d *schema.ResourceData, version string, esKey string) BaseConfig {
	baseConfig := BaseConfig{}
	baseConfig.UserAgent = buildUserAgent(version)
	baseConfig.Header = http.Header{"User-Agent": []string{baseConfig.UserAgent}}

	if esConn, ok := d.GetOk(esKey); ok {
		if resource := esConn.([]interface{})[0]; resource != nil {
			config := resource.(map[string]interface{})

			if username, ok := config["username"]; ok {
				baseConfig.Username = username.(string)
			}
			if password, ok := config["password"]; ok {
				baseConfig.Password = password.(string)
			}
		}
	}

	return baseConfig
}

func buildUserAgent(version string) string {
	return fmt.Sprintf("elasticstack-terraform-provider/%s", version)
}

func buildEsClient(d *schema.ResourceData, baseConfig BaseConfig, useEnvAsDefault bool, key string) (*elasticsearch.Client, diag.Diagnostics) {
	var diags diag.Diagnostics

	esConn, ok := d.GetOk(key)
	if !ok {
		return nil, diags
	}

	config := elasticsearch.Config{
		Header:   baseConfig.Header,
		Username: baseConfig.Username,
		Password: baseConfig.Password,
	}

	// if defined, then we only have a single entry
	if es := esConn.([]interface{})[0]; es != nil {
		esConfig := es.(map[string]interface{})

		if apikey, ok := esConfig["api_key"]; ok {
			config.APIKey = apikey.(string)
		}

		if useEnvAsDefault {
			if endpoints := os.Getenv("ELASTICSEARCH_ENDPOINTS"); endpoints != "" {
				var addrs []string
				for _, e := range strings.Split(endpoints, ",") {
					addrs = append(addrs, strings.TrimSpace(e))
				}
				config.Addresses = addrs
			}
		}

		if endpoints, ok := esConfig["endpoints"]; ok && len(endpoints.([]interface{})) > 0 {
			var addrs []string
			for _, e := range endpoints.([]interface{}) {
				addrs = append(addrs, e.(string))
			}
			config.Addresses = addrs
		}

		if insecure, ok := esConfig["insecure"]; ok && insecure.(bool) {
			tlsClientConfig := ensureTLSClientConfig(&config)
			tlsClientConfig.InsecureSkipVerify = true
		}

		if caFile, ok := esConfig["ca_file"]; ok && caFile.(string) != "" {
			caCert, err := os.ReadFile(caFile.(string))
			if err != nil {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Unable to read CA File",
					Detail:   err.Error(),
				})
				return nil, diags
			}
			config.CACert = caCert
		}
		if caData, ok := esConfig["ca_data"]; ok && caData.(string) != "" {
			config.CACert = []byte(caData.(string))
		}

		if certFile, ok := esConfig["cert_file"]; ok && certFile.(string) != "" {
			if keyFile, ok := esConfig["key_file"]; ok && keyFile.(string) != "" {
				cert, err := tls.LoadX509KeyPair(certFile.(string), keyFile.(string))
				if err != nil {
					diags = append(diags, diag.Diagnostic{
						Severity: diag.Error,
						Summary:  "Unable to read certificate or key file",
						Detail:   err.Error(),
					})
					return nil, diags
				}
				tlsClientConfig := ensureTLSClientConfig(&config)
				tlsClientConfig.Certificates = []tls.Certificate{cert}
			} else {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Unable to read key file",
					Detail:   "Path to key file has not been configured or is empty",
				})
				return nil, diags
			}
		}
		if certData, ok := esConfig["cert_data"]; ok && certData.(string) != "" {
			if keyData, ok := esConfig["key_data"]; ok && keyData.(string) != "" {
				cert, err := tls.X509KeyPair([]byte(certData.(string)), []byte(keyData.(string)))
				if err != nil {
					diags = append(diags, diag.Diagnostic{
						Severity: diag.Error,
						Summary:  "Unable to parse certificate or key",
						Detail:   err.Error(),
					})
					return nil, diags
				}
				tlsClientConfig := ensureTLSClientConfig(&config)
				tlsClientConfig.Certificates = []tls.Certificate{cert}
			} else {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Unable to parse key",
					Detail:   "Key data has not been configured or is empty",
				})
				return nil, diags
			}
		}
	}

	if logging.IsDebugOrHigher() {
		config.EnableDebugLogger = true
		config.Logger = &debugLogger{Name: "elasticsearch"}
	}

	es, err := elasticsearch.NewClient(config)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to create Elasticsearch client",
			Detail:   err.Error(),
		})
		return nil, diags
	}

	return es, diags
}

func buildKibanaConfig(d *schema.ResourceData, baseConfig BaseConfig) (kibana.Config, diag.Diagnostics) {
	var diags diag.Diagnostics

	kibConn, ok := d.GetOk("kibana")
	if !ok {
		return kibana.Config{}, diags
	}

	// Use ES details by default
	config := kibana.Config{
		Username: baseConfig.Username,
		Password: baseConfig.Password,
	}

	// if defined, then we only have a single entry
	if kib := kibConn.([]interface{})[0]; kib != nil {
		kibConfig := kib.(map[string]interface{})

		if username := os.Getenv("KIBANA_USERNAME"); username != "" {
			config.Username = strings.TrimSpace(username)
		}
		if password := os.Getenv("KIBANA_PASSWORD"); password != "" {
			config.Password = strings.TrimSpace(password)
		}
		if endpoint := os.Getenv("KIBANA_ENDPOINT"); endpoint != "" {
			config.Address = endpoint
		}
		if insecure := os.Getenv("KIBANA_INSECURE"); insecure != "" {
			if insecureValue, _ := strconv.ParseBool(insecure); insecureValue {
				config.DisableVerifySSL = true
			}
		}

		if username, ok := kibConfig["username"]; ok && username != "" {
			config.Username = username.(string)
		}
		if password, ok := kibConfig["password"]; ok && password != "" {
			config.Password = password.(string)
		}

		if endpoints, ok := kibConfig["endpoints"]; ok && len(endpoints.([]interface{})) > 0 {
			// We're curently limited by the API to a single endpoint
			if endpoint := endpoints.([]interface{})[0]; endpoint != nil {
				config.Address = endpoint.(string)
			}
		}

		if insecure, ok := kibConfig["insecure"]; ok && insecure.(bool) {
			config.DisableVerifySSL = true
		}
	}

	return config, nil
}

func buildKibanaClient(config kibana.Config) (*kibana.Client, diag.Diagnostics) {
	kib, err := kibana.NewClient(config)
	if err != nil {
		return nil, diag.FromErr(err)
	}

	if logging.IsDebugOrHigher() {
		kib.Client.SetDebug(true)
	}

	return kib, nil
}

func buildAlertingClient(baseConfig BaseConfig, config kibana.Config) *alerting.APIClient {
	alertingConfig := alerting.Configuration{
		UserAgent: baseConfig.UserAgent,
		Servers: alerting.ServerConfigurations{
			{
				URL: config.Address,
			},
		},
		Debug: logging.IsDebugOrHigher(),
	}
	return alerting.NewAPIClient(&alertingConfig)
}

func buildConnectorsClient(baseConfig BaseConfig, config kibana.Config) (*connectors.Client, error) {
	basicAuthProvider, err := securityprovider.NewSecurityProviderBasicAuth(config.Username, config.Password)
	if err != nil {
		return nil, fmt.Errorf("unable to create basic auth provider: %w", err)
	}
	return connectors.NewClient(
		config.Address,
		connectors.WithRequestEditorFn(basicAuthProvider.Intercept),
	)
}

func buildSloClient(baseConfig BaseConfig, config kibana.Config) *slo.APIClient {
	//again this is the same -- worth pulling this out into a common place?
	sloConfig := slo.Configuration{
		UserAgent: baseConfig.UserAgent,
		Servers: slo.ServerConfigurations{
			{
				URL: config.Address,
			},
		},
		Debug: logging.IsDebugOrHigher(),
	}
	return slo.NewAPIClient(&sloConfig)
}

func buildFleetClient(d *schema.ResourceData, kibanaCfg kibana.Config) (*fleet.Client, diag.Diagnostics) {
	var diags diag.Diagnostics

	// Order of precedence for config options:
	// 1 (highest): environment variables
	// 2: resource config
	// 3: kibana config

	// Set variables from kibana config.
	config := fleet.Config{
		URL:      kibanaCfg.Address,
		Username: kibanaCfg.Username,
		Password: kibanaCfg.Password,
		Insecure: kibanaCfg.DisableVerifySSL,
	}

	// Set variables from resource config.
	if fleetDataRaw, ok := d.GetOk("fleet"); ok {
		fleetData, ok := fleetDataRaw.([]interface{})[0].(map[string]any)
		if !ok {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to parse Fleet configuration",
				Detail:   "Fleet configuration data has not been configured correctly or is empty",
			})
			return nil, diags
		}
		if v, ok := fleetData["endpoint"].(string); ok && v != "" {
			config.URL = v
		}
		if v, ok := fleetData["username"].(string); ok && v != "" {
			config.Username = v
		}
		if v, ok := fleetData["password"].(string); ok && v != "" {
			config.Password = v
		}
		if v, ok := fleetData["api_key"].(string); ok && v != "" {
			config.APIKey = v
		}
		if v, ok := fleetData["ca_certs"].([]interface{}); ok && len(v) > 0 {
			for _, elem := range v {
				if vStr, elemOk := elem.(string); elemOk {
					config.CACerts = append(config.CACerts, vStr)
				}
			}
		}
		if v, ok := fleetData["insecure"].(bool); ok {
			config.Insecure = v
		}
	}

	if v := os.Getenv("FLEET_API_KEY"); v != "" {
		config.APIKey = v
	}
	if v := os.Getenv("FLEET_CA_CERTS"); v != "" {
		config.CACerts = strings.Split(v, ",")
	}

	client, err := fleet.NewClient(config)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to create Fleet client",
			Detail:   err.Error(),
		})
		return nil, diags
	}

	return client, diags
}

const esKey string = "elasticsearch"

func newApiClient(d *schema.ResourceData, version string) (*ApiClient, diag.Diagnostics) {
	baseConfig := buildBaseConfig(d, version, esKey)
	kibanaConfig, diags := buildKibanaConfig(d, baseConfig)
	if diags.HasError() {
		return nil, diags
	}

	esClient, diags := buildEsClient(d, baseConfig, true, esKey)
	if diags.HasError() {
		return nil, diags
	}

	kibanaClient, diags := buildKibanaClient(kibanaConfig)
	if diags.HasError() {
		return nil, diags
	}

	alertingClient := buildAlertingClient(baseConfig, kibanaConfig)
	sloClient := buildSloClient(baseConfig, kibanaConfig)

	connectorsClient, err := buildConnectorsClient(baseConfig, kibanaConfig)
	if err != nil {
		return nil, diag.FromErr(fmt.Errorf("cannot create Kibana connectors client: [%w]", err))
	}

	fleetClient, diags := buildFleetClient(d, kibanaConfig)
	if diags.HasError() {
		return nil, diags
	}

	return &ApiClient{
		elasticsearch:            esClient,
		elasticsearchClusterInfo: nil,
		kibana:                   kibanaClient,
		kibanaConfig:             kibanaConfig,
		alerting:                 alertingClient.AlertingApi,
		connectors:               connectorsClient,
		slo:                      sloClient.SlosApi,
		fleet:                    fleetClient,
		version:                  version,
	}, nil
}
