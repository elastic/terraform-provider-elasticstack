package scope

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
)

type frameworkResource struct {
	client *clients.APIClient
}

// This package lives under internal/fleet, so analyzer should flag violations.
func (r *frameworkResource) Read(_ context.Context) error {
	return r.client.GetFleetClient() // want "Kibana/Fleet client usage must use a helper-derived \\*clients.APIClient from clients.NewKibanaAPIClientFromSDKResource\\(\\.\\.\\.\\) or clients.MaybeNewKibanaAPIClientFromFrameworkResource\\(\\.\\.\\.\\)"
}

func sdkFleetOK(_ any) error {
	client, _ := clients.NewKibanaAPIClientFromSDKResource(nil, nil)
	return client.GetFleetClient()
}
