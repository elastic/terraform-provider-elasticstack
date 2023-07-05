package kibana

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibana"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func ResourceActionConnector() *schema.Resource {
	apikeySchema := map[string]*schema.Schema{
		"connector_id": {
			Description: "A UUID v1 or v4 to use instead of a randomly generated ID.",
			Type:        schema.TypeString,
			Computed:    true,
			Optional:    true,
			ForceNew:    true,
		},
		"space_id": {
			Description: "An identifier for the space. If space_id is not provided, the default space is used.",
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "default",
			ForceNew:    true,
		},
		"name": {
			Description: "The name of the connector. While this name does not have to be unique, a distinctive name can help you identify a connector.",
			Type:        schema.TypeString,
			Required:    true,
		},
		"connector_type_id": {
			Description: "The ID of the connector type, e.g. `.index`.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"config": {
			Description:  "The configuration for the connector. Configuration properties vary depending on the connector type.",
			Type:         schema.TypeString,
			Optional:     true,
			Computed:     true,
			ValidateFunc: validation.StringIsJSON,
		},
		"secrets": {
			Description:      "The secrets configuration for the connector. Secrets configuration properties vary depending on the connector type.",
			Type:             schema.TypeString,
			Optional:         true,
			DiffSuppressFunc: utils.DiffJsonSuppress,
			ValidateFunc:     validation.StringIsJSON,
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
		Description: "Creates a Kibana action connector. See https://www.elastic.co/guide/en/kibana/current/action-types.html",

		CreateContext: resourceConnectorCreate,
		UpdateContext: resourceConnectorUpdate,
		ReadContext:   resourceConnectorRead,
		DeleteContext: resourceConnectorDelete,
		CustomizeDiff: connectorCustomizeDiff,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: apikeySchema,
	}
}

func connectorCustomizeDiff(ctx context.Context, rd *schema.ResourceDiff, in interface{}) error {
	if !rd.HasChange("config") {
		return nil
	}
	oldVal, newVal := rd.GetChange("config")
	oldJSON := oldVal.(string)
	newJSON := newVal.(string)
	if oldJSON == newJSON {
		return nil
	}
	oldVal, newVal = rd.GetChange("connector_type_id")
	oldTypeID := oldVal.(string)
	newTypeID := newVal.(string)
	if oldTypeID != newTypeID {
		return nil
	}

	rawState := rd.GetRawState()
	if !rawState.IsKnown() || rawState.IsNull() {
		return nil
	}

	state := rawState.GetAttr("config")
	if !state.IsKnown() || state.IsNull() {
		return nil
	}

	stateJSON := state.AsString()

	customJSON, err := kibana.ConnectorConfigWithDefaults(oldTypeID, newJSON, oldJSON, stateJSON)
	if err != nil {
		return err
	}
	return rd.SetNew("config", string(customJSON))
}

func resourceConnectorCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, diags := clients.NewApiClient(d, meta)
	if diags.HasError() {
		return diags
	}

	connectorOld, diags := expandActionConnector(d)
	if diags.HasError() {
		return diags
	}

	connectorID, diags := kibana.CreateConnector(ctx, client, connectorOld)

	if diags.HasError() {
		return diags
	}

	compositeID := &clients.CompositeId{ClusterId: connectorOld.SpaceID, ResourceId: connectorID}
	d.SetId(compositeID.String())

	return resourceConnectorRead(ctx, d, meta)
}

func resourceConnectorUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, diags := clients.NewApiClient(d, meta)
	if diags.HasError() {
		return diags
	}

	connectorOld, diags := expandActionConnector(d)
	if diags.HasError() {
		return diags
	}

	compositeIDold, diags := clients.CompositeIdFromStr(d.Id())
	if diags.HasError() {
		return diags
	}
	connectorOld.ConnectorID = compositeIDold.ResourceId

	connectorID, diags := kibana.UpdateConnector(ctx, client, connectorOld)

	if diags.HasError() {
		return diags
	}

	compositeIDnew := &clients.CompositeId{ClusterId: connectorOld.SpaceID, ResourceId: connectorID}
	d.SetId(compositeIDnew.String())

	return resourceConnectorRead(ctx, d, meta)
}

func resourceConnectorRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, diags := clients.NewApiClient(d, meta)
	if diags.HasError() {
		return diags
	}

	compositeID, diags := clients.CompositeIdFromStr(d.Id())
	if diags.HasError() {
		return diags
	}

	connector, diags := kibana.GetConnector(ctx, client, compositeID.ResourceId, compositeID.ClusterId)
	if connector == nil && diags == nil {
		d.SetId("")
		return diags
	}
	if diags.HasError() {
		return diags
	}

	return flattenActionConnector(connector, d)
}

func resourceConnectorDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, diags := clients.NewApiClient(d, meta)
	if diags.HasError() {
		return diags
	}

	compositeID, diags := clients.CompositeIdFromStr(d.Id())
	if diags.HasError() {
		return diags
	}

	spaceId := d.Get("space_id").(string)

	if diags := kibana.DeleteConnector(ctx, client, compositeID.ResourceId, spaceId); diags.HasError() {
		return diags
	}

	d.SetId("")
	return nil
}

func expandActionConnector(d *schema.ResourceData) (models.KibanaActionConnector, diag.Diagnostics) {
	var diags diag.Diagnostics

	connector := models.KibanaActionConnector{
		SpaceID:         d.Get("space_id").(string),
		Name:            d.Get("name").(string),
		ConnectorTypeID: d.Get("connector_type_id").(string),
	}

	connector.ConfigJSON = d.Get("config").(string)
	connector.SecretsJSON = d.Get("secrets").(string)

	return connector, diags
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
