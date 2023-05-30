package fleet

import (
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func getFleetClient(d *schema.ResourceData, meta interface{}) (*fleet.Client, diag.Diagnostics) {
	client, diags := clients.NewApiClient(d, meta)
	if diags.HasError() {
		return nil, diags
	}
	fleetClient, err := client.GetFleetClient()
	if err != nil {
		return nil, diag.FromErr(err)
	}

	return fleetClient, nil
}
