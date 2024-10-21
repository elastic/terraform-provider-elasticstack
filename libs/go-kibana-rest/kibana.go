package kibana

import (
	"crypto/tls"

	"github.com/disaster37/go-kibana-rest/v8/kbapi"
	"github.com/go-resty/resty/v2"
)

// Config contain the value to access on Kibana API
type Config struct {
	Address          string
	Username         string
	Password         string
	ApiKey           string
	DisableVerifySSL bool
	CAs              []string
}

// Client contain the REST client and the API specification
type Client struct {
	*kbapi.API
	Client *resty.Client
}

// NewDefaultClient init client with empty config
func NewDefaultClient() (*Client, error) {
	return NewClient(Config{})
}

// NewClient init client with custom config
func NewClient(cfg Config) (*Client, error) {
	if cfg.Address == "" {
		cfg.Address = "http://localhost:5601"
	}

	restyClient := resty.New().
		SetBaseURL(cfg.Address).
		SetHeader("kbn-xsrf", "true").
		SetHeader("Content-Type", "application/json").
		SetDisableWarn(true)

	if cfg.ApiKey != "" {
		restyClient.SetAuthScheme("ApiKey").SetAuthToken(cfg.ApiKey)
	} else {
		restyClient.SetBasicAuth(cfg.Username, cfg.Password)
	}

	for _, path := range cfg.CAs {
		restyClient.SetRootCertificate(path)
	}

	client := &Client{
		Client: restyClient,
		API:    kbapi.New(restyClient),
	}

	if cfg.DisableVerifySSL {
		client.Client.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
	}

	return client, nil

}
