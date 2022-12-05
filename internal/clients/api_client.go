package clients

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/elastic/go-elasticsearch/v7"
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
	es      *elasticsearch.Client
	version string
}

func NewApiClientFunc(version string) func(context.Context, *schema.ResourceData) (interface{}, diag.Diagnostics) {
	return func(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
		var diags diag.Diagnostics
		config := elasticsearch.Config{}
		config.Header = http.Header{"User-Agent": []string{fmt.Sprintf("elasticstack-terraform-provider/%s", version)}}

		if v, ok := d.GetOk("elasticsearch"); ok {
			// if defined we must have only one entry
			if esc := v.([]interface{})[0]; esc != nil {
				esConfig := esc.(map[string]interface{})
				if username, ok := esConfig["username"]; ok {
					config.Username = username.(string)
				}
				if password, ok := esConfig["password"]; ok {
					config.Password = password.(string)
				}
				if apikey, ok := esConfig["api_key"]; ok {
					config.APIKey = apikey.(string)
				}

				// default endpoints taken from Env if set
				if es := os.Getenv("ELASTICSEARCH_ENDPOINTS"); es != "" {
					endpoints := make([]string, 0)
					for _, e := range strings.Split(es, ",") {
						endpoints = append(endpoints, strings.TrimSpace(e))
					}
					config.Addresses = endpoints
				}
				// setting endpoints from config block if provided
				if eps, ok := esConfig["endpoints"]; ok && len(eps.([]interface{})) > 0 {
					endpoints := make([]string, 0)
					for _, e := range eps.([]interface{}) {
						endpoints = append(endpoints, e.(string))
					}
					config.Addresses = endpoints
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
		}

		es, err := elasticsearch.NewClient(config)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to create Elasticsearch client",
				Detail:   err.Error(),
			})
		}
		if logging.IsDebugOrHigher() {
			es.Transport = newDebugTransport("elasticsearch", es.Transport)
		}
		return &ApiClient{es, version}, diags
	}
}

func NewAcceptanceTestingClient() (*ApiClient, error) {
	config := elasticsearch.Config{}
	config.Header = http.Header{"User-Agent": []string{"elasticstack-terraform-provider/tf-acceptance-testing"}}

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

	return &ApiClient{es, "acceptance-testing"}, nil
}

func NewApiClient(d *schema.ResourceData, meta interface{}) (*ApiClient, error) {
	defaultClient := meta.(*ApiClient)
	// if the config provided let's use it
	if esConn, ok := d.GetOk("elasticsearch_connection"); ok {
		config := elasticsearch.Config{}
		config.Header = http.Header{"User-Agent": []string{fmt.Sprintf("elasticstack-terraform-provider/%s", defaultClient.version)}}

		// there is always only 1 connection per resource
		conn := esConn.([]interface{})[0].(map[string]interface{})

		if u := conn["username"]; u != nil {
			config.Username = u.(string)
		}
		if p := conn["password"]; p != nil {
			config.Password = p.(string)
		}
		if k := conn["api_key"]; k != nil {
			config.APIKey = k.(string)
		}
		if endpoints := conn["endpoints"]; endpoints != nil {
			var addrs []string
			for _, e := range endpoints.([]interface{}) {
				addrs = append(addrs, e.(string))
			}
			config.Addresses = addrs
		}
		if insecure := conn["insecure"]; insecure.(bool) {
			tlsClientConfig := ensureTLSClientConfig(&config)
			tlsClientConfig.InsecureSkipVerify = true
		}
		if caFile, ok := conn["ca_file"]; ok && caFile.(string) != "" {
			caCert, err := os.ReadFile(caFile.(string))
			if err != nil {
				return nil, fmt.Errorf("Unable to read ca_file: %w", err)
			}
			config.CACert = caCert
		}
		if caData, ok := conn["ca_data"]; ok && caData.(string) != "" {
			config.CACert = []byte(caData.(string))
		}

		if certFile, ok := conn["cert_file"]; ok && certFile.(string) != "" {
			if keyFile, ok := conn["key_file"]; ok && keyFile.(string) != "" {
				cert, err := tls.LoadX509KeyPair(certFile.(string), keyFile.(string))
				if err != nil {
					return nil, fmt.Errorf("Unable to read certificate or key file: %w", err)
				}
				tlsClientConfig := ensureTLSClientConfig(&config)
				tlsClientConfig.Certificates = []tls.Certificate{cert}
			} else {
				return nil, fmt.Errorf("Unable to read key file: Path to key file has not been configured or is empty")
			}
		}
		if certData, ok := conn["cert_data"]; ok && certData.(string) != "" {
			if keyData, ok := conn["key_data"]; ok && keyData.(string) != "" {
				cert, err := tls.X509KeyPair([]byte(certData.(string)), []byte(keyData.(string)))
				if err != nil {
					return nil, fmt.Errorf("Unable to parse certificate or key: %w", err)
				}
				tlsClientConfig := ensureTLSClientConfig(&config)
				tlsClientConfig.Certificates = []tls.Certificate{cert}
			} else {
				return nil, fmt.Errorf("Unable to parse key: Key data has not been configured or is empty")
			}
		}

		es, err := elasticsearch.NewClient(config)
		if err != nil {
			return nil, fmt.Errorf("Unable to create Elasticsearch client")
		}
		if logging.IsDebugOrHigher() {
			es.Transport = newDebugTransport("elasticsearch", es.Transport)
		}
		return &ApiClient{es, defaultClient.version}, nil
	} else { // or return the default client
		return defaultClient, nil
	}
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

func (a *ApiClient) GetESClient() *elasticsearch.Client {
	return a.es
}

func (a *ApiClient) ID(ctx context.Context, resourceId string) (*CompositeId, diag.Diagnostics) {
	var diags diag.Diagnostics
	clusterId, diags := a.ClusterID(ctx)
	if diags.HasError() {
		return nil, diags
	}
	return &CompositeId{*clusterId, resourceId}, diags
}

func (a *ApiClient) serverInfo(ctx context.Context) (map[string]interface{}, diag.Diagnostics) {
	var diags diag.Diagnostics
	res, err := a.es.Info(a.es.Info.WithContext(ctx))
	if err != nil {
		return nil, diag.FromErr(err)
	}
	defer res.Body.Close()
	if diags := utils.CheckError(res, "Unable to connect to the Elasticsearch cluster"); diags.HasError() {
		return nil, diags
	}

	info := make(map[string]interface{})
	if err := json.NewDecoder(res.Body).Decode(&info); err != nil {
		return nil, diag.FromErr(err)
	}

	return info, diags
}

func (a *ApiClient) ServerVersion(ctx context.Context) (*version.Version, diag.Diagnostics) {
	info, diags := a.serverInfo(ctx)
	if diags.HasError() {
		return nil, diags
	}

	rawVersion := info["version"].(map[string]interface{})["number"].(string)
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

	if uuid := info["cluster_uuid"].(string); uuid != "" && uuid != "_na_" {
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
