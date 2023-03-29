package kibana

import (
	"context"

	"github.com/disaster37/go-kibana-rest/v8/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func ResourceSpace() *schema.Resource {
	apikeySchema := map[string]*schema.Schema{
		"id": {
			Description: "Internal identifier of the resource.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"space_id": {
			Description: "The space ID that is part of the Kibana URL when inside the space.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"name": {
			Description: "The display name for the space.",
			Type:        schema.TypeString,
			Required:    true,
		},
		"description": {
			Description: "The description for the space.",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"disabled_features": {
			Description: "The list of disabled features for the space. To get a list of available feature IDs, use the Features API (https://www.elastic.co/guide/en/kibana/master/features-api-get.html).",
			Type:        schema.TypeSet,
			Optional:    true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"initials": {
			Description:  "The initials shown in the space avatar. By default, the initials are automatically generated from the space name. Initials must be 1 or 2 characters.",
			Type:         schema.TypeString,
			Optional:     true,
			ValidateFunc: validation.StringLenBetween(1, 2),
		},
		"color": {
			Description: "The hexadecimal color code used in the space avatar. By default, the color is automatically generated from the space name.",
			Type:        schema.TypeString,
			Optional:    true,
		},
	}

	return &schema.Resource{
		Description: "Creates a Kibana space. See, https://www.elastic.co/guide/en/kibana/master/spaces-api-post.html",

		CreateContext: resourceSpaceUpsert,
		UpdateContext: resourceSpaceUpsert,
		ReadContext:   resourceSpaceRead,
		DeleteContext: resourceSpaceDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: apikeySchema,
	}
}

func resourceSpaceUpsert(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, diags := clients.NewApiClient(d, meta)
	if diags.HasError() {
		return diags
	}

	kibana, err := client.GetKibanaClient()
	if err != nil {
		return diag.FromErr(err)
	}

	space := kbapi.KibanaSpace{
		ID:   d.Get("space_id").(string),
		Name: d.Get("name").(string),
	}

	if description, ok := d.GetOk("description"); ok {
		space.Description = description.(string)
	}

	if disabledFeatures, ok := d.GetOk("disabled_features"); ok {
		space.DisabledFeatures = disabledFeatures.([]string)
	}

	if initials, ok := d.GetOk("initials"); ok {
		space.Initials = initials.(string)
	}

	if color, ok := d.GetOk("color"); ok {
		space.Color = color.(string)
	}

	var spaceResponse *kbapi.KibanaSpace

	if d.IsNewResource() {
		spaceResponse, err = kibana.KibanaSpaces.Create(&space)
		if err != nil {
			return diag.FromErr(err)
		}
	} else {
		spaceResponse, err = kibana.KibanaSpaces.Update(&space)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	id, diags := client.ID(ctx, spaceResponse.ID)
	if diags.HasError() {
		return diags
	}

	d.SetId(id.String())

	return resourceSpaceRead(ctx, d, meta)
}

func resourceSpaceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, diags := clients.NewApiClient(d, meta)
	if diags.HasError() {
		return diags
	}
	compId, diags := clients.CompositeIdFromStr(d.Id())
	if diags.HasError() {
		return diags
	}
	id := compId.ResourceId

	kibana, err := client.GetKibanaClient()
	if err != nil {
		return diag.FromErr(err)
	}

	space, err := kibana.KibanaSpaces.Get(id)
	if space == nil && err == nil {
		d.SetId("")
		return diags
	}
	if err != nil {
		return diag.FromErr(err)
	}

	// set the fields
	if err := d.Set("space_id", space.ID); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("name", space.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("description", space.Description); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("disabled_features", space.DisabledFeatures); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("initials", space.Initials); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("color", space.Color); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceSpaceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, diags := clients.NewApiClient(d, meta)
	if diags.HasError() {
		return diags
	}
	compId, diags := clients.CompositeIdFromStr(d.Id())
	if diags.HasError() {
		return diags
	}

	kibana, err := client.GetKibanaClient()
	if err != nil {
		return diag.FromErr(err)
	}

	err = kibana.KibanaSpaces.Delete(compId.ResourceId)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")
	return diags
}
