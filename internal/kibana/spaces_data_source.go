package kibana

import (
	"context"
	"fmt"

	"github.com/disaster37/go-kibana-rest/v8/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func DataSourceSpace() *schema.Resource {
	var spacesSchema = map[string]*schema.Schema{
		"search": {
			Description: "Search spaces by name.",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"spaces": {
			Description: "The list of spaces.",
			Type:        schema.TypeList,
			Computed:    true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"id": {
						Description: "Internal identifier of the resource.",
						Type:        schema.TypeString,
						Computed:    true,
					},
					"space_id": {
						Description: "The space ID that is part of the Kibana URL when inside the space.",
						Type:        schema.TypeString,
						Required:    true,
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
						Description: "The initials shown in the space avatar. By default, the initials are automatically generated from the space name. Initials must be 1 or 2 characters.",
						Type:        schema.TypeString,
						Computed:    true,
					},
					"color": {
						Description: "The hexadecimal color code used in the space avatar. By default, the color is automatically generated from the space name.",
						Type:        schema.TypeString,
						Computed:    true,
					},
					"image_url": {
						Description: "The data-URL encoded image to display in the space avatar.",
						Type:        schema.TypeString,
						Optional:    true,
					},
				},
			},
		},
	}

	return &schema.Resource{
		Description: "Search for spaces by name.",
		ReadContext: datasourceSpacesRead,
		Schema:      spacesSchema,
	}
}

func datasourceSpacesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, diags := clients.NewApiClientFromSDKResource(d, meta)
	if diags.HasError() {
		return diags
	}

	kibana, err := client.GetKibanaClient()
	if err != nil {
		return diag.FromErr(err)
	}

	spaceName := d.Get("search").(string)

	allSpaces, err := kibana.KibanaSpaces.List()
	if err != nil {
		return diag.FromErr(err)
	}

	foundSpaces := kbapi.KibanaSpaces{}
	for _, space := range allSpaces {
		if space.Name == spaceName {
			foundSpaces = append(foundSpaces, space)
		}
	}

	d.SetId(fmt.Sprintf("%d", schema.HashString(spaceName)))
	if err := d.Set("spaces", flattenSpaces(foundSpaces)); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func flattenSpaces(spaces kbapi.KibanaSpaces) []interface{} {
	spacesList := []interface{}{}

	for _, space := range spaces {
		values := map[string]interface{}{
			"id":                space.ID,
			"space_id":          space.ID,
			"name":              space.Name,
			"description":       space.Description,
			"disabled_features": space.DisabledFeatures,
			"initials":          space.Initials,
			"color":             space.Color,
			"image_url":         space.ImageURL,
		}

		spacesList = append(spacesList, values)
	}

	return spacesList
}
