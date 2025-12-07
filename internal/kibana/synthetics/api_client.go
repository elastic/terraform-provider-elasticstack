package synthetics

import (
	"github.com/disaster37/go-kibana-rest/v8"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibana_oapi"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// ESApiClient interface provides access to the underlying API client
type ESApiClient interface {
	GetClient() *clients.ApiClient
}

// GetKibanaClient returns a configured Kibana client for the given ESApiClient
func GetKibanaClient(c ESApiClient, dg diag.Diagnostics) *kibana.Client {
	client := c.GetClient()
	if client == nil {
		dg.AddError(
			"Unconfigured Client",
			"Expected configured client. Please report this issue to the provider developers.",
		)
		return nil
	}

	kibanaClient, err := client.GetKibanaClient()
	if err != nil {
		dg.AddError("unable to get kibana client", err.Error())
		return nil
	}
	return kibanaClient
}

// GetKibanaOAPIClient returns a configured Kibana OpenAPI client for the given ESApiClient
func GetKibanaOAPIClient(c ESApiClient, dg diag.Diagnostics) *kibana_oapi.Client {
	client := c.GetClient()
	if client == nil {
		dg.AddError(
			"Unconfigured Client",
			"Expected configured client. Please report this issue to the provider developers.",
		)
		return nil
	}

	kibanaClient, err := client.GetKibanaOapiClient()
	if err != nil {
		dg.AddError("unable to get kibana oapi client", err.Error())
		return nil
	}
	return kibanaClient
}
