package clients

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/disaster37/go-kibana-rest/v8"
	"github.com/elastic/go-elasticsearch/v7"
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
	version                  string
}

func NewApiClientFunc(version string) func(context.Context, *schema.ResourceData) (interface{}, diag.Diagnostics) {
	return func(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
		return newApiClient(d, version, true)
	}
}

func NewAcceptanceTestingClient() (*ApiClient, error) {
	config := elasticsearch.Config{
		Header: buildHeader("tf-acceptance-testing"),
	}

	if es := os.Getenv("ELASTICSEARCH_ENDPOINTS"); es != "" {
		endpoints := make([]string, 0)
		for _, e := range strings.Split(es, ",") {
			endpoints = append(endpoints, strings.TrimSpace(e))
		}
		config.Addresses = endpoints
	}

	if username := os.Getenv("ELASTICSEARCH_USERNAME"); username != "" {
		config.Username = username
		config.Password = os.Getenv("ELASTICSEARCH_PASSWORD")
	} else {
		config.APIKey = os.Getenv("ELASTICSEARCH_API_KEY")
	}

	es, err := elasticsearch.NewClient(config)
	if err != nil {
		return nil, err
	}

	return &ApiClient{es, nil, nil, "acceptance-testing"}, nil
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
	Username string
	Password string
	Header   http.Header
}

// Build base config from ES which can be shared for other resources
func buildBaseConfig(d *schema.ResourceData, version string, esKey string) BaseConfig {
	baseConfig := BaseConfig{}
	baseConfig.Header = buildHeader(version)

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

func buildHeader(version string) http.Header {
	return http.Header{"User-Agent": []string{fmt.Sprintf("elasticstack-terraform-provider/%s", version)}}
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

	es, err := elasticsearch.NewClient(config)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to create Elasticsearch client",
			Detail:   err.Error(),
		})
		return nil, diags
	}
	if logging.IsDebugOrHigher() {
		es.Transport = newDebugTransport("elasticsearch", es.Transport)
	}

	return es, diags
}

func buildKibanaClient(d *schema.ResourceData, baseConfig BaseConfig, useEnvAsDefault bool) (*kibana.Client, diag.Diagnostics) {
	var diags diag.Diagnostics

	kibConn, ok := d.GetOk("kibana")
	if !ok {
		return nil, diags
	}

	// Use ES details by default
	config := kibana.Config{
		Username: baseConfig.Username,
		Password: baseConfig.Password,
	}

	// if defined, then we only have a single entry
	if kib := kibConn.([]interface{})[0]; kib != nil {
		kibConfig := kib.(map[string]interface{})

		if useEnvAsDefault {
			if username := os.Getenv("KIBANA_USERNAME"); username != "" {
				config.Username = strings.TrimSpace(username)
			}
			if password := os.Getenv("KIBANA_PASSWORD"); password != "" {
				config.Password = strings.TrimSpace(password)
			}
		}

		if username, ok := kibConfig["username"]; ok {
			config.Username = username.(string)
		}
		if password, ok := kibConfig["password"]; ok {
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

	kib, err := kibana.NewClient(config)

	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to create Kibana client",
			Detail:   err.Error(),
		})
	}

	return kib, diags
}

const esKey string = "elasticsearch"

func newApiClient(d *schema.ResourceData, version string, useEnvAsDefault bool) (*ApiClient, diag.Diagnostics) {
	var diags diag.Diagnostics
	baseConfig := buildBaseConfig(d, version, esKey)

	esClient, diags := buildEsClient(d, baseConfig, useEnvAsDefault, esKey)
	if diags.HasError() {
		return nil, diags
	}

	kibanaClient, diags := buildKibanaClient(d, baseConfig, useEnvAsDefault)
	if diags.HasError() {
		return nil, diags
	}

	return &ApiClient{
		elasticsearch:            esClient,
		elasticsearchClusterInfo: nil,
		kibana:                   kibanaClient,
		version:                  version,
	}, diags
}
