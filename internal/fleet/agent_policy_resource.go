package fleet

import (
	"context"
	"fmt"

	fleetapi "github.com/elastic/terraform-provider-elasticstack/generated/fleet"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	monitorLogs    = "logs"
	monitorMetrics = "metrics"
)

var minVersionGlobalDataTags = version.Must(version.NewVersion("8.15.0"))

func ResourceAgentPolicy() *schema.Resource {
	agentPolicySchema := map[string]*schema.Schema{
		"policy_id": {
			Description: "Unique identifier of the agent policy.",
			Type:        schema.TypeString,
			Computed:    true,
			Optional:    true,
			ForceNew:    true,
		},
		"name": {
			Description: "The name of the agent policy.",
			Type:        schema.TypeString,
			Required:    true,
		},
		"namespace": {
			Description: "The namespace of the agent policy.",
			Type:        schema.TypeString,
			Required:    true,
		},
		"description": {
			Description: "The description of the agent policy.",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"data_output_id": {
			Description: "The identifier for the data output.",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"monitoring_output_id": {
			Description: "The identifier for monitoring output.",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"fleet_server_host_id": {
			Description: "The identifier for the Fleet server host.",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"download_source_id": {
			Description: "The identifier for the Elastic Agent binary download server.",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"sys_monitoring": {
			Description: "Enable collection of system logs and metrics.",
			Type:        schema.TypeBool,
			Optional:    true,
		},
		"monitor_logs": {
			Description: "Enable collection of agent logs.",
			Type:        schema.TypeBool,
			Optional:    true,
		},
		"monitor_metrics": {
			Description: "Enable collection of agent metrics.",
			Type:        schema.TypeBool,
			Optional:    true,
		},
		"skip_destroy": {
			Description: "Set to true if you do not wish the agent policy to be deleted at destroy time, and instead just remove the agent policy from the Terraform state.",
			Type:        schema.TypeBool,
			Optional:    true,
		},
		"global_data_tags": {
			Description: "User-defined data tags that are added to all inputs.",
			Type:        schema.TypeMap,
			Optional:    true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
	}

	return &schema.Resource{
		Description: "Creates a new Fleet Agent Policy. See https://www.elastic.co/guide/en/fleet/current/agent-policy.html",

		CreateContext: resourceAgentPolicyCreate,
		ReadContext:   resourceAgentPolicyRead,
		UpdateContext: resourceAgentPolicyUpdate,
		DeleteContext: resourceAgentPolicyDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: agentPolicySchema,
	}
}

func resourceAgentPolicyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	fleetClient, diags := getFleetClient(d, meta)
	if diags.HasError() {
		return diags
	}

	apiClient, diags := clients.NewApiClientFromSDKResource(d, meta)
	if diags.HasError() {
		return diags
	}

	serverVersion, diags := apiClient.ServerVersion(ctx)
	if diags.HasError() {
		return diags
	}

	if id := d.Get("policy_id").(string); id != "" {
		d.SetId(id)
	}

	req := fleetapi.AgentPolicyCreateRequest{
		Name:      d.Get("name").(string),
		Namespace: d.Get("namespace").(string),
	}

	if value := d.Get("policy_id").(string); value != "" {
		req.Id = &value
	}
	if value := d.Get("description").(string); value != "" {
		req.Description = &value
	}
	if value := d.Get("data_output_id").(string); value != "" {
		req.DataOutputId = &value
	}
	if value := d.Get("download_source_id").(string); value != "" {
		req.DownloadSourceId = &value
	}
	if value := d.Get("fleet_server_host_id").(string); value != "" {
		req.FleetServerHostId = &value
	}
	if value := d.Get("monitoring_output_id").(string); value != "" {
		req.MonitoringOutputId = &value
	}

	monitoringValues := make([]fleetapi.AgentPolicyCreateRequestMonitoringEnabled, 0, 2)
	if value := d.Get("monitor_logs").(bool); value {
		monitoringValues = append(monitoringValues, monitorLogs)
	}
	if value := d.Get("monitor_metrics").(bool); value {
		monitoringValues = append(monitoringValues, monitorMetrics)
	}
	req.MonitoringEnabled = &monitoringValues

	if tags, ok := d.GetOk("global_data_tags"); ok {
		tagMap := tags.(map[string]interface{})

		if len(tagMap) > 0 && serverVersion.LessThan(minVersionGlobalDataTags) {
			return diag.FromErr(fmt.Errorf("'global_data_tags' is supported only for Elasticsearch v%s and above", minVersionGlobalDataTags.String()))
		}

		globalDataTags := make([]fleetapi.GlobalDataTag, 0, len(tagMap))
		for key, value := range tagMap {
			globalDataTags = append(globalDataTags, fleetapi.GlobalDataTag{
				Name:  key,
				Value: value.(string),
			})
		}
		req.GlobalDataTags = globalDataTags
	} else {
		req.GlobalDataTags = make([]fleetapi.GlobalDataTag, 0) // Ensure it's an empty array
	}

	policy, diags := fleet.CreateAgentPolicy(ctx, fleetClient, req)
	if diags.HasError() {
		return diags
	}

	d.SetId(policy.Id)
	if err := d.Set("policy_id", policy.Id); err != nil {
		return diag.FromErr(err)
	}

	return resourceAgentPolicyRead(ctx, d, meta)
}

func resourceAgentPolicyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	fleetClient, diags := getFleetClient(d, meta)
	if diags.HasError() {
		return diags
	}

	apiClient, diags := clients.NewApiClientFromSDKResource(d, meta)
	if diags.HasError() {
		return diags
	}

	serverVersion, diags := apiClient.ServerVersion(ctx)
	if diags.HasError() {
		return diags
	}

	req := fleetapi.AgentPolicyUpdateRequest{
		Name:      d.Get("name").(string),
		Namespace: d.Get("namespace").(string),
	}

	if value := d.Get("description").(string); value != "" {
		req.Description = &value
	}
	if value := d.Get("data_output_id").(string); value != "" {
		req.DataOutputId = &value
	}
	if value := d.Get("download_source_id").(string); value != "" {
		req.DownloadSourceId = &value
	}
	if value := d.Get("fleet_server_host_id").(string); value != "" {
		req.FleetServerHostId = &value
	}
	if value := d.Get("monitoring_output_id").(string); value != "" {
		req.MonitoringOutputId = &value
	}

	monitoringValues := make([]fleetapi.AgentPolicyUpdateRequestMonitoringEnabled, 0, 2)
	if value := d.Get("monitor_logs").(bool); value {
		monitoringValues = append(monitoringValues, monitorLogs)
	}
	if value := d.Get("monitor_metrics").(bool); value {
		monitoringValues = append(monitoringValues, monitorMetrics)
	}
	req.MonitoringEnabled = &monitoringValues

	if tags, ok := d.GetOk("global_data_tags"); ok {
		tagMap := tags.(map[string]interface{})

		if len(tagMap) > 0 && serverVersion.LessThan(minVersionGlobalDataTags) {
			return diag.FromErr(fmt.Errorf("'global_data_tags' is supported only for Elasticsearch v%s and above", minVersionGlobalDataTags.String()))
		}

		globalDataTags := make([]fleetapi.GlobalDataTag, 0, len(tagMap))
		for key, value := range tagMap {
			globalDataTags = append(globalDataTags, fleetapi.GlobalDataTag{
				Name:  key,
				Value: value.(string),
			})
		}
		req.GlobalDataTags = globalDataTags
	} else {
		req.GlobalDataTags = make([]fleetapi.GlobalDataTag, 0) // Ensure it's an empty array
	}

	_, diags = fleet.UpdateAgentPolicy(ctx, fleetClient, d.Id(), req)
	if diags.HasError() {
		return diags
	}

	return resourceAgentPolicyRead(ctx, d, meta)
}

func resourceAgentPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	fleetClient, diags := getFleetClient(d, meta)
	if diags.HasError() {
		return diags
	}

	apiClient, diags := clients.NewApiClientFromSDKResource(d, meta)
	if diags.HasError() {
		return diags
	}

	serverVersion, diags := apiClient.ServerVersion(ctx)
	if diags.HasError() {
		return diags
	}

	agentPolicy, diags := fleet.ReadAgentPolicy(ctx, fleetClient, d.Id())
	if diags.HasError() {
		return diags
	}

	// Not found.
	if agentPolicy == nil {
		d.SetId("")
		return nil
	}

	if err := d.Set("name", agentPolicy.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("namespace", agentPolicy.Namespace); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("policy_id", agentPolicy.Id); err != nil {
		return diag.FromErr(err)
	}
	if agentPolicy.Description != nil {
		if err := d.Set("description", *agentPolicy.Description); err != nil {
			return diag.FromErr(err)
		}
	}
	if agentPolicy.DataOutputId != nil {
		if err := d.Set("data_output_id", *agentPolicy.DataOutputId); err != nil {
			return diag.FromErr(err)
		}
	}
	if agentPolicy.DownloadSourceId != nil {
		if err := d.Set("download_source_id", *agentPolicy.DownloadSourceId); err != nil {
			return diag.FromErr(err)
		}
	}
	if agentPolicy.FleetServerHostId != nil {
		if err := d.Set("fleet_server_host_id", *agentPolicy.FleetServerHostId); err != nil {
			return diag.FromErr(err)
		}
	}
	if agentPolicy.MonitoringOutputId != nil {
		if err := d.Set("monitoring_output_id", *agentPolicy.MonitoringOutputId); err != nil {
			return diag.FromErr(err)
		}
	}
	if agentPolicy.MonitoringEnabled != nil {
		for _, v := range *agentPolicy.MonitoringEnabled {
			switch v {
			case monitorLogs:
				if err := d.Set("monitor_logs", true); err != nil {
					return diag.FromErr(err)
				}
			case monitorMetrics:
				if err := d.Set("monitor_metrics", true); err != nil {
					return diag.FromErr(err)

				}
			}
		}
	}

	if agentPolicy.GlobalDataTags != nil {

		globalDataTags := make(map[string]string, len(agentPolicy.GlobalDataTags))
		for _, tag := range agentPolicy.GlobalDataTags {
			globalDataTags[tag.Name] = tag.Value.(string)
		}

		if err := d.Set("global_data_tags", globalDataTags); err != nil {
			return diag.FromErr(err)
		}
	} else {
		if err := d.Set("global_data_tags", nil); err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}

func resourceAgentPolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	if d.Get("skip_destroy").(bool) {
		tflog.Debug(ctx, "Skipping destroy of Agent Policy", map[string]interface{}{"policy_id": d.Id()})
		return nil
	}

	fleetClient, diags := getFleetClient(d, meta)
	if diags.HasError() {
		return diags
	}

	if diags = fleet.DeleteAgentPolicy(ctx, fleetClient, d.Id()); diags.HasError() {
		return diags
	}
	d.SetId("")

	return diags
}
