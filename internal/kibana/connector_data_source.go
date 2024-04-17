package kibana

import (
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
		ReadContext: resourceConnectorsRead,
		Schema:      connectorSchema,
	}
}
