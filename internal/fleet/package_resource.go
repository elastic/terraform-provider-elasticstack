package fleet

import (
	"context"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
)

func getPackageID(name, version string) string {
	hash, _ := utils.StringToHash(name + version)

	return *hash
}

func ResourcePackage() *schema.Resource {
	packageSchema := map[string]*schema.Schema{
		"name": {
			Description: "The package name.",
			Type:        schema.TypeString,
			Required:    true,
		},
		"version": {
			Description: "The package version.",
			Type:        schema.TypeString,
			Required:    true,
		},
		"force": {
			Description: "Set to true to force the requested action.",
			Type:        schema.TypeBool,
			Optional:    true,
		},
		"skip_destroy": {
			Description: "Set to true if you do not wish the package to be uninstalled at destroy time, and instead just remove the package from the Terraform state.",
			Type:        schema.TypeBool,
			Optional:    true,
		},
	}

	return &schema.Resource{
		Description: "Manage installation of a Fleet package.",

		CreateContext: resourcePackageInstall,
		ReadContext:   resourcePackageRead,
		UpdateContext: resourcePackageInstall,
		DeleteContext: resourcePackageDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: packageSchema,
	}
}

func resourcePackageInstall(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	fleetClient, diags := getFleetClient(d, meta)
	if diags.HasError() {
		return diags
	}

	name := d.Get("name").(string)
	version := d.Get("version").(string)
	force := d.Get("force").(bool)

	d.SetId(getPackageID(name, version))

	if diags = fleet.InstallPackage(ctx, fleetClient, name, version, force); diags.HasError() {
		return diags
	}

	return resourcePackageRead(ctx, d, meta)
}

func resourcePackageRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	fleetClient, diags := getFleetClient(d, meta)
	if diags.HasError() {
		return diags
	}

	name := d.Get("name").(string)
	version := d.Get("version").(string)

	d.SetId(getPackageID(name, version))

	if diags = fleet.ReadPackage(ctx, fleetClient, name, version); diags.HasError() {
		return diags
	}

	return nil
}

func resourcePackageDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	version := d.Get("version").(string)
	force := d.Get("force").(bool)

	if d.Get("skip_destroy").(bool) {
		tflog.Debug(ctx, "Skipping uninstall of Package", map[string]interface{}{"name": name, "version": version})
		return nil
	}

	d.SetId(getPackageID(name, version))

	fleetClient, diags := getFleetClient(d, meta)
	if diags.HasError() {
		return diags
	}

	if diags = fleet.Uninstall(ctx, fleetClient, name, version, force); diags.HasError() {
		return diags
	}
	d.SetId("")

	return diags
}
