package kibana

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Schema defines the schema for the data source.
func DataSourceExportSavedObjects() *schema.Resource {
	var savedObjectSchema = map[string]*schema.Schema{
		"id": {
			Description: "Identifier for the data source.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"space_id": {
			Description: "An identifier for the space. If space_id is not provided, the default space is used.",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"objects": {
			Description: "JSON-encoded list of objects to export. Each object should have 'type' and 'id' fields.",
			Type:        schema.TypeString,
			Required:    true,
		},
		"exclude_export_details": {
			Description: "Do not add export details. Defaults to true.",
			Type:        schema.TypeBool,
			Optional:    true,
		},
		"include_references_deep": {
			Description: "Include references to other saved objects recursively. Defaults to true.",
			Type:        schema.TypeBool,
			Optional:    true,
		},
		"exported_objects": {
			Description: "The exported objects in NDJSON format.",
			Type:        schema.TypeString,
			Computed:    true,
		},
	}
	return &schema.Resource{
		Description: "Export Kibana saved objects. This data source allows you to export saved objects from Kibana and store the result in the Terraform state.",
		ReadContext: datasourceExportSavedObjectRead,
		Schema:      savedObjectSchema,
	}
}

func datasourceExportSavedObjectRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, diags := clients.NewApiClientFromSDKResource(d, meta)
	if diags.HasError() {
		return diags
	}
	oapiClient, err := client.GetKibanaOapiClient()
	if err != nil {
		return diag.FromErr(err)
	}

	spaceId := d.Get("space_id").(string)
	if spaceId == "" {
		spaceId = "default"
	}

	var objectsList kbapi.PostSavedObjectsExportJSONBodyHasReference1
	objectsJSON := d.Get("objects").(string)

	var rawObjects []map[string]interface{}
	if err := json.Unmarshal([]byte(objectsJSON), &rawObjects); err != nil {
		return diag.Errorf("Invalid objects JSON: %v", err)
	}

	for _, obj := range rawObjects {
		id, ok := obj["id"].(string)
		if !ok {
			return diag.Errorf("Object missing 'id' field")
		}
		objType, ok := obj["type"].(string)
		if !ok {
			return diag.Errorf("Object missing 'type' field")
		}
		objectsList = append(objectsList, struct {
			Id   string `json:"id"`
			Type string `json:"type"`
		}{
			Id:   id,
			Type: objType,
		})
	}

	excludeExportDetails := true
	if val, ok := d.GetOk("exclude_export_details"); ok {
		excludeExportDetails = val.(bool)
	}

	includeReferencesDeep := true
	if val, ok := d.GetOk("include_references_deep"); ok {
		includeReferencesDeep = val.(bool)
	}

	body := kbapi.PostSavedObjectsExportJSONRequestBody{
		ExcludeExportDetails:  &excludeExportDetails,
		IncludeReferencesDeep: &includeReferencesDeep,
		Objects:               &objectsList,
	}

	resp, err := oapiClient.API.PostSavedObjectsExportWithResponse(ctx, body)
	if err != nil {
		return diag.Errorf("unable to export saved objects: [%v]", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return diag.Diagnostics{
			diag.Diagnostic{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("Unexpected status code from server: got HTTP %d", resp.StatusCode()),
				Detail:   string(resp.Body),
			},
		}
	}

	// Set the results
	compositeID := &clients.CompositeId{ClusterId: spaceId, ResourceId: "export"}
	d.SetId(compositeID.String())

	if err := d.Set("exported_objects", string(resp.Body)); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("space_id", spaceId); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("exclude_export_details", excludeExportDetails); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("include_references_deep", includeReferencesDeep); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
