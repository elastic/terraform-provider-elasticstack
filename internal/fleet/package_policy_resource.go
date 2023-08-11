package fleet

import (
	"context"
	"encoding/json"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	fleetapi "github.com/elastic/terraform-provider-elasticstack/generated/fleet"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
)

func ResourcePackagePolicy() *schema.Resource {
	packagePolicySchema := map[string]*schema.Schema{
		"policy_id": {
			Description: "Unique identifier of the package policy.",
			Type:        schema.TypeString,
			Computed:    true,
			Optional:    true,
			ForceNew:    true,
		},
		"name": {
			Description: "The name of the package policy.",
			Type:        schema.TypeString,
			Required:    true,
		},
		"namespace": {
			Description: "The namespace of the package policy.",
			Type:        schema.TypeString,
			Required:    true,
		},
		"agent_policy_id": {
			Description: "ID of the agent policy.",
			Type:        schema.TypeString,
			Required:    true,
		},
		"description": {
			Description: "The description of the package policy.",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"enabled": {
			Description: "Enable the package policy.",
			Type:        schema.TypeBool,
			Default:     true,
			Optional:    true,
		},
		"force": {
			Description: "Force operations, such as creation and deletion, to occur.",
			Type:        schema.TypeBool,
			Optional:    true,
		},
		"package_name": {
			Description: "The name of the package.",
			Type:        schema.TypeString,
			Required:    true,
		},
		"package_version": {
			Description: "The version of the package.",
			Type:        schema.TypeString,
			Required:    true,
		},
		"input": {
			Type:     schema.TypeList,
			Required: true,
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
					"config_json": {
						Description: "Input configuration as JSON.",
						Type:        schema.TypeString,
						Computed:    true,
						Optional:    true,
						Sensitive:   true,
					},
					"processors_json": {
						Description: "Input processors as JSON.",
						Type:        schema.TypeString,
						Computed:    true,
						Optional:    true,
						Sensitive:   true,
					},
				},
			},
		},
	}

	return &schema.Resource{
		Description: "Creates a new Fleet Package Policy. See https://www.elastic.co/guide/en/fleet/current/agent-policy.html",

		CreateContext: resourcePackagePolicyCreate,
		ReadContext:   resourcePackagePolicyRead,
		UpdateContext: resourcePackagePolicyUpdate,
		DeleteContext: resourcePackagePolicyDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: packagePolicySchema,
	}
}

func resourcePackagePolicyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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
	req.Package.Name = d.Get("package_name").(string)
	req.Package.Version = d.Get("package_version").(string)

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

	return resourcePackagePolicyRead(ctx, d, meta)
}

func resourcePackagePolicyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	fleetClient, diags := getFleetClient(d, meta)
	if diags.HasError() {
		return diags
	}

	req := fleetapi.UpdatePackagePolicyJSONRequestBody{
		PolicyId: d.Get("agent_policy_id").(string),
		Name:     d.Get("name").(string),
	}
	req.Package.Name = d.Get("package_name").(string)
	req.Package.Version = d.Get("package_version").(string)

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

	return resourcePackagePolicyRead(ctx, d, meta)
}

func resourcePackagePolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	fleetClient, diags := getFleetClient(d, meta)
	if diags.HasError() {
		return diags
	}

	tflog.Info(ctx, "Package policy ID is: "+d.Id())

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
	if err := d.Set("package_name", pkgPolicy.Package.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("package_version", pkgPolicy.Package.Version); err != nil {
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

	var newInputs []any
	seen := map[string]struct{}{}

	// Range over existing inputs first.
	for _, v := range d.Get("input").([]interface{}) {
		oldInputData := v.(map[string]any)
		inputID := oldInputData["input_id"].(string)

		newInputData, ok := pkgPolicy.Inputs[inputID]
		if !ok {
			continue
		}

		inputMap := map[string]any{
			"input_id": inputID,
			"enabled":  newInputData.Enabled,
		}

		if newInputData.Streams != nil {
			data, err := json.Marshal(*newInputData.Streams)
			if err != nil {
				return diag.FromErr(err)
			}
			inputMap["streams_json"] = string(data)
		}
		if newInputData.Vars != nil {
			data, err := json.Marshal(*newInputData.Vars)
			if err != nil {
				return diag.FromErr(err)
			}
			inputMap["vars_json"] = string(data)
		}
		if newInputData.Config != nil {
			data, err := json.Marshal(*newInputData.Config)
			if err != nil {
				return diag.FromErr(err)
			}
			inputMap["config_json"] = string(data)
		}
		if newInputData.Processors != nil {
			data, err := json.Marshal(*newInputData.Processors)
			if err != nil {
				return diag.FromErr(err)
			}
			inputMap["processors_json"] = string(data)
		}

		newInputs = append(newInputs, inputMap)
		seen[inputID] = struct{}{}
	}

	// Handle any new inputs.
	for inputID, input := range pkgPolicy.Inputs {
		if _, exists := seen[inputID]; exists {
			continue
		}

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
		if input.Config != nil {
			data, err := json.Marshal(*input.Config)
			if err != nil {
				return diag.FromErr(err)
			}
			inputMap["config_json"] = string(data)
		}
		if input.Processors != nil {
			data, err := json.Marshal(*input.Processors)
			if err != nil {
				return diag.FromErr(err)
			}
			inputMap["processors_json"] = string(data)
		}

		newInputs = append(newInputs, inputMap)
	}
	if err := d.Set("input", newInputs); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourcePackagePolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	fleetClient, diags := getFleetClient(d, meta)
	if diags.HasError() {
		return diags
	}

	id := d.Get("policy_id").(string)
	d.SetId(id)

	force := d.Get("force").(bool)

	if diags = fleet.DeletePackagePolicy(ctx, fleetClient, id, force); diags.HasError() {
		return diags
	}
	d.SetId("")

	return diags
}
