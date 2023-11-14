package fleet

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
)

func DataSourcePackage() *schema.Resource {
	packageSchema := map[string]*schema.Schema{
		"name": {
			Description: "The package name.",
			Type:        schema.TypeString,
			Required:    true,
		},
		"prerelease": {
			Description: "Include prerelease packages.",
			Type:        schema.TypeBool,
			Optional:    true,
		},
		"version": {
			Description: "The package version.",
			Type:        schema.TypeString,
			Computed:    true,
		},
	}

	return &schema.Resource{
		Description: "Retrieves the latest version of an integration package in Fleet.",

		ReadContext: dataSourcePackageRead,

		Schema: packageSchema,
	}
}

func dataSourcePackageRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	fleetClient, diags := getFleetClient(d, meta)
	if diags.HasError() {
		return diags
	}

	pkgName := d.Get("name").(string)
	if d.Id() == "" {
		hash, err := utils.StringToHash(pkgName)
		if err != nil {
			return diag.FromErr(err)
		}
		d.SetId(*hash)
	}

	prerelease := d.Get("prerelease").(bool)
	allPackages, diags := fleet.AllPackages(ctx, fleetClient, prerelease)
	if diags.HasError() {
		return diags
	}

	for _, v := range allPackages {
		if v.Name != pkgName {
			continue
		}

		if err := d.Set("version", v.Version); err != nil {
			return diag.FromErr(err)
		}
		break
	}

	return diags
}
