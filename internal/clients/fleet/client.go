package fleet

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"os"
	"strings"

	fleetapi "github.com/elastic/terraform-provider-elasticstack/generated/fleet"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/logging"
)

// Config is the configuration for the fleet client.
type Config struct {
	URL      string
	Username string
	Password string
	APIKey   string
	Insecure bool
	CACerts  []string
}

// Client provides an API client for Elastic Fleet.
type Client struct {
	URL  string
	HTTP *http.Client
	API  *fleetapi.ClientWithResponses
}

// NewClient creates a new Elastic Fleet API client.
func NewClient(cfg Config) (*Client, error) {
	var caCertPool *x509.CertPool
	if len(cfg.CACerts) > 0 {
		caCertPool = x509.NewCertPool()
		for _, certFile := range cfg.CACerts {
			certData, err := os.ReadFile(certFile)
			if err != nil {
				return nil, fmt.Errorf("unable to open CA certificate file %q: %w", certFile, err)
			}
			_ = caCertPool.AppendCertsFromPEM(certData)
		}
	}

	var roundTripper http.RoundTripper = &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: cfg.Insecure,
			RootCAs:            caCertPool,
		},
	}

	if logging.IsDebugOrHigher() {
		roundTripper = utils.NewDebugTransport("Fleet", roundTripper)
	}

	httpClient := &http.Client{
		Transport: &transport{
			Config: cfg,
			next:   roundTripper,
		},
	}

	endpoint := cfg.URL
	if !strings.HasSuffix(endpoint, "/") {
		endpoint += "/"
	}

	fleetAPIClient, err := fleetapi.NewClientWithResponses(endpoint, fleetapi.WithHTTPClient(httpClient))
	if err != nil {
		return nil, fmt.Errorf("unable to create Fleet API client: %w", err)
	}

	return &Client{
		URL:  cfg.URL,
		HTTP: httpClient,
		API:  fleetAPIClient,
	}, nil
}

type transport struct {
	Config
	next http.RoundTripper
}

func (t *transport) RoundTrip(req *http.Request) (*http.Response, error) {
	switch req.Method {
	case "GET", "HEAD":
	default:
		// https://www.elastic.co/guide/en/kibana/current/api.html#api-request-headers
		req.Header.Add("kbn-xsrf", "true")
	}

	if t.Username != "" {
		req.SetBasicAuth(t.Username, t.Password)
	}

	if t.APIKey != "" {
		req.Header.Add("Authorization", "ApiKey "+t.APIKey)
	}

	return t.next.RoundTrip(req)
}
