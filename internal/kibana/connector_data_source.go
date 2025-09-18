package kibana

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibana_oapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func DataSourceConnector() *schema.Resource {
	var connectorSchema = map[string]*schema.Schema{
		"connector_id": {
			Description: "A UUID v1 or v4 randomly generated ID.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"space_id": {
			Description: "An identifier for the space. If space_id is not provided, the default space is used.",
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "default",
		},
		"name": {
			Description: "The name of the connector. While this name does not have to be unique, a distinctive name can help you identify a connector.",
			Type:        schema.TypeString,
			Required:    true,
		},
		"connector_type_id": {
			Description: "The ID of the connector type, e.g. `.index`.",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"config": {
			Description: "The configuration for the connector. Configuration properties vary depending on the connector type.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"is_deprecated": {
			Description: "Indicates whether the connector type is deprecated.",
			Type:        schema.TypeBool,
			Computed:    true,
		},
		"is_missing_secrets": {
			Description: "Indicates whether secrets are missing for the connector.",
			Type:        schema.TypeBool,
			Computed:    true,
		},
		"is_preconfigured": {
			Description: "Indicates whether it is a preconfigured connector.",
			Type:        schema.TypeBool,
			Computed:    true,
		},
	}

	return &schema.Resource{
		Description: "Search for a connector by name, space id, and type. Note, that this data source will fail if more than one connector shares the same name.",
		ReadContext: datasourceConnectorRead,
		Schema:      connectorSchema,
	}
}

func datasourceConnectorRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, diags := clients.NewApiClientFromSDKResource(d, meta)
	if diags.HasError() {
		return diags
	}
	oapiClient, err := client.GetKibanaOapiClient()
	if err != nil {
		return diag.FromErr(err)
	}
	connectorName := d.Get("name").(string)
	spaceId := d.Get("space_id").(string)
	connectorType := d.Get("connector_type_id").(string)

	foundConnectors, diags := kibana_oapi.SearchConnectors(ctx, oapiClient, connectorName, spaceId, connectorType)
	if diags.HasError() {
		return diags
	}

	if len(foundConnectors) == 0 {
		return diag.Errorf("error while creating elasticstack_kibana_action_connector datasource: connector with name [%s/%s] and type [%s] not found", spaceId, connectorName, connectorType)
	}

	if len(foundConnectors) > 1 {
		return diag.Errorf("error while creating elasticstack_kibana_action_connector datasource: multiple connectors found with name [%s/%s] and type [%s]", spaceId, connectorName, connectorType)
	}

	compositeID := &clients.CompositeId{ClusterId: spaceId, ResourceId: foundConnectors[0].ConnectorID}
	d.SetId(compositeID.String())

	return flattenActionConnector(foundConnectors[0], d)
}

func flattenActionConnector(connector *models.KibanaActionConnector, d *schema.ResourceData) diag.Diagnostics {
	if err := d.Set("connector_id", connector.ConnectorID); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("space_id", connector.SpaceID); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("name", connector.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("connector_type_id", connector.ConnectorTypeID); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("config", connector.ConfigJSON); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("is_deprecated", connector.IsDeprecated); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("is_missing_secrets", connector.IsMissingSecrets); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("is_preconfigured", connector.IsPreconfigured); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
