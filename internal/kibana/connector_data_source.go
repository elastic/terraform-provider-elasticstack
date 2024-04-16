package kibana

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func DataSourceConnector() *schema.Resource {
	return &schema.Resource{
		Description: "Retrieve a specific connector.",
		ReadContext: dataSourceConnectorRead,
		Schema:      connectorSchema,
	}
}

func dataSourceConnectorRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	connectorName := d.Get("name").(string)
	d.Set("name", connectorName)
	spaceId := d.Get("space_id").(string)
	d.SetId(spaceId)

	return resourceConnectorsRead(ctx, d, meta)
}
