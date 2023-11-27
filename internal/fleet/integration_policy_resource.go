package fleet

import (
	"context"
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	fleetapi "github.com/elastic/terraform-provider-elasticstack/generated/fleet"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
)

func ResourceIntegrationPolicy() *schema.Resource {
	packagePolicySchema := map[string]*schema.Schema{
		"policy_id": {
			Description: "Unique identifier of the integration policy.",
			Type:        schema.TypeString,
			Computed:    true,
			Optional:    true,
			ForceNew:    true,
		},
		"name": {
			Description: "The name of the integration policy.",
			Type:        schema.TypeString,
			Required:    true,
		},
		"namespace": {
			Description: "The namespace of the integration policy.",
			Type:        schema.TypeString,
			Required:    true,
		},
		"agent_policy_id": {
			Description: "ID of the agent policy.",
			Type:        schema.TypeString,
			Required:    true,
		},
		"description": {
			Description: "The description of the integration policy.",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"enabled": {
			Description: "Enable the integration policy.",
			Type:        schema.TypeBool,
			Optional:    true,
			Computed:    true,
		},
		"force": {
			Description: "Force operations, such as creation and deletion, to occur.",
			Type:        schema.TypeBool,
			Optional:    true,
		},
		"integration_name": {
			Description: "The name of the integration package.",
			Type:        schema.TypeString,
			Required:    true,
		},
		"integration_version": {
			Description: "The version of the integration package.",
			Type:        schema.TypeString,
			Required:    true,
		},
		"input": {
			Type:     schema.TypeList,
			Optional: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"input_id": {
						Description: "The identifier of the input.",
						Type:        schema.TypeString,
						Required:    true,
					},
					"enabled": {
						Description: "Enable the input.",
						Type:        schema.TypeBool,
						Default:     true,
						Optional:    true,
					},
					"streams_json": {
						Description:  "Input streams as JSON.",
						Type:         schema.TypeString,
						ValidateFunc: validation.StringIsJSON,
						Optional:     true,
						Computed:     true,
						Sensitive:    true,
					},
					"vars_json": {
						Description:  "Input variables as JSON.",
						Type:         schema.TypeString,
						ValidateFunc: validation.StringIsJSON,
						Computed:     true,
						Optional:     true,
						Sensitive:    true,
					},
				},
			},
		},
		"vars_json": {
			Description:  "Integration-level variables as JSON.",
			Type:         schema.TypeString,
			ValidateFunc: validation.StringIsJSON,
			Computed:     true,
			Optional:     true,
			Sensitive:    true,
		},
	}

	return &schema.Resource{
		Description: "Creates a new Fleet Integration Policy. See https://www.elastic.co/guide/en/fleet/current/add-integration-to-policy.html",

		CreateContext: resourceIntegrationPolicyCreate,
		ReadContext:   resourceIntegrationPolicyRead,
		UpdateContext: resourceIntegrationPolicyUpdate,
		DeleteContext: resourceIntegrationPolicyDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: packagePolicySchema,
	}
}

func resourceIntegrationPolicyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	fleetClient, diags := getFleetClient(d, meta)
	if diags.HasError() {
		return diags
	}

	if id := d.Get("policy_id").(string); id != "" {
		d.SetId(id)
	}

	req := fleetapi.CreatePackagePolicyJSONRequestBody{
		PolicyId: d.Get("agent_policy_id").(string),
		Name:     d.Get("name").(string),
	}
	req.Package.Name = d.Get("integration_name").(string)
	req.Package.Version = d.Get("integration_version").(string)

	if value := d.Get("policy_id").(string); value != "" {
		req.Id = &value
	}
	if value := d.Get("namespace").(string); value != "" {
		req.Namespace = &value
	}
	if value := d.Get("description").(string); value != "" {
		req.Description = &value
	}
	if value := d.Get("force").(bool); value {
		req.Force = &value
	}
	if varsRaw, _ := d.Get("vars_json").(string); varsRaw != "" {
		vars := map[string]interface{}{}
		if err := json.Unmarshal([]byte(varsRaw), &vars); err != nil {
			panic(err)
		}
		req.Vars = &vars
	}

	values := d.Get("input").([]interface{})
	if len(values) > 0 {
		inputMap := map[string]fleetapi.PackagePolicyRequestInput{}

		for _, v := range values {
			var input fleetapi.PackagePolicyRequestInput

			inputData := v.(map[string]interface{})
			inputID := inputData["input_id"].(string)

			enabled, _ := inputData["enabled"].(bool)
			input.Enabled = &enabled

			if streamsRaw, _ := inputData["streams_json"].(string); streamsRaw != "" {
				streams := map[string]fleetapi.PackagePolicyRequestInputStream{}
				if err := json.Unmarshal([]byte(streamsRaw), &streams); err != nil {
					panic(err)
				}
				input.Streams = &streams
			}
			if varsRaw, _ := inputData["vars_json"].(string); varsRaw != "" {
				vars := map[string]interface{}{}
				if err := json.Unmarshal([]byte(varsRaw), &vars); err != nil {
					panic(err)
				}
				input.Vars = &vars
			}

			inputMap[inputID] = input
		}

		req.Inputs = &inputMap
	}

	obj, diags := fleet.CreatePackagePolicy(ctx, fleetClient, req)
	if diags.HasError() {
		return diags
	}

	d.SetId(obj.Id)
	if err := d.Set("policy_id", obj.Id); err != nil {
		return diag.FromErr(err)
	}

	return resourceIntegrationPolicyRead(ctx, d, meta)
}

func resourceIntegrationPolicyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	fleetClient, diags := getFleetClient(d, meta)
	if diags.HasError() {
		return diags
	}

	req := fleetapi.UpdatePackagePolicyJSONRequestBody{
		PolicyId: d.Get("agent_policy_id").(string),
		Name:     d.Get("name").(string),
	}
	req.Package.Name = d.Get("integration_name").(string)
	req.Package.Version = d.Get("integration_version").(string)

	if value := d.Get("policy_id").(string); value != "" {
		req.Id = &value
	}
	if value := d.Get("namespace").(string); value != "" {
		req.Namespace = &value
	}
	if value := d.Get("description").(string); value != "" {
		req.Description = &value
	}
	if value := d.Get("force").(bool); value {
		req.Force = &value
	}
	if varsRaw, _ := d.Get("vars_json").(string); varsRaw != "" {
		vars := map[string]interface{}{}
		if err := json.Unmarshal([]byte(varsRaw), &vars); err != nil {
			panic(err)
		}
		req.Vars = &vars
	}

	if values := d.Get("input").([]interface{}); len(values) > 0 {
		inputMap := map[string]fleetapi.PackagePolicyRequestInput{}

		for _, v := range values {
			var input fleetapi.PackagePolicyRequestInput

			inputData := v.(map[string]interface{})
			inputID := inputData["input_id"].(string)

			enabled, _ := inputData["enabled"].(bool)
			input.Enabled = &enabled

			if streamsRaw, _ := inputData["streams_json"].(string); streamsRaw != "" {
				streams := map[string]fleetapi.PackagePolicyRequestInputStream{}
				if err := json.Unmarshal([]byte(streamsRaw), &streams); err != nil {
					panic(err)
				}
				input.Streams = &streams
			}
			if varsRaw, _ := inputData["vars_json"].(string); varsRaw != "" {
				vars := map[string]interface{}{}
				if err := json.Unmarshal([]byte(varsRaw), &vars); err != nil {
					panic(err)
				}
				input.Vars = &vars
			}

			inputMap[inputID] = input
		}

		req.Inputs = &inputMap
	}

	_, diags = fleet.UpdatePackagePolicy(ctx, fleetClient, d.Id(), req)
	if diags.HasError() {
		return diags
	}

	return resourceIntegrationPolicyRead(ctx, d, meta)
}

func resourceIntegrationPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	fleetClient, diags := getFleetClient(d, meta)
	if diags.HasError() {
		return diags
	}

	pkgPolicy, diags := fleet.ReadPackagePolicy(ctx, fleetClient, d.Id())
	if diags.HasError() {
		return diags
	}

	// Not found.
	if pkgPolicy == nil {
		d.SetId("")
		return nil
	}

	if err := d.Set("name", pkgPolicy.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("namespace", pkgPolicy.Namespace); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("integration_name", pkgPolicy.Package.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("integration_version", pkgPolicy.Package.Version); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("agent_policy_id", pkgPolicy.PolicyId); err != nil {
		return diag.FromErr(err)
	}
	if pkgPolicy.Description != nil {
		if err := d.Set("description", *pkgPolicy.Description); err != nil {
			return diag.FromErr(err)
		}
	}
	if pkgPolicy.Vars != nil {

		vars := make(map[string]any, len(*pkgPolicy.Vars))
		// Var values are wrapped in a type/value struct and need
		// to be extracted. The only applies to reading values back
		// from the API, sending var values does not use this format.
		for k, v := range *pkgPolicy.Vars {
			wrappedTypeValue, _ := v.(map[string]any)
			if wrappedValue, ok := wrappedTypeValue["value"]; ok {
				vars[k] = wrappedValue
			}
		}

		data, err := json.Marshal(vars)
		if err != nil {
			return diag.FromErr(err)
		}
		if err = d.Set("vars_json", string(data)); err != nil {
			return diag.FromErr(err)
		}
	}

	var inputs []any
	for inputID, input := range pkgPolicy.Inputs {
		inputMap := map[string]any{
			"input_id": inputID,
			"enabled":  input.Enabled,
		}

		if input.Streams != nil {
			data, err := json.Marshal(*input.Streams)
			if err != nil {
				return diag.FromErr(err)
			}
			inputMap["streams_json"] = string(data)
		}
		if input.Vars != nil {
			data, err := json.Marshal(*input.Vars)
			if err != nil {
				return diag.FromErr(err)
			}
			inputMap["vars_json"] = string(data)
		}

		inputs = append(inputs, inputMap)
	}
	if err := d.Set("input", inputs); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceIntegrationPolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	fleetClient, diags := getFleetClient(d, meta)
	if diags.HasError() {
		return diags
	}

	force := d.Get("force").(bool)

	if diags = fleet.DeletePackagePolicy(ctx, fleetClient, d.Id(), force); diags.HasError() {
		return diags
	}
	d.SetId("")

	return diags
}
