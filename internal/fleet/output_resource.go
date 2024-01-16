package fleet

import (
	"context"

	fleetapi "github.com/elastic/terraform-provider-elasticstack/generated/fleet"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func ResourceOutput() *schema.Resource {
	outputSchema := map[string]*schema.Schema{
		"output_id": {
			Description: "Unique identifier of the output.",
			Type:        schema.TypeString,
			Computed:    true,
			Optional:    true,
		},
		"name": {
			Description: "The name of the output.",
			Type:        schema.TypeString,
			Required:    true,
		},
		"type": {
			Description:  "The output type.",
			Type:         schema.TypeString,
			Required:     true,
			ValidateFunc: validation.StringInSlice([]string{"elasticsearch", "logstash"}, false),
		},
		"hosts": {
			Description: "A list of hosts.",
			Type:        schema.TypeList,
			Optional:    true,
			MinItems:    1,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"ca_sha256": {
			Description: "Fingerprint of the Elasticsearch CA certificate.",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"ca_trusted_fingerprint": {
			Description: "Fingerprint of trusted CA.",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"default_integrations": {
			Description: "Make this output the default for agent integrations.",
			Type:        schema.TypeBool,
			Optional:    true,
		},
		"default_monitoring": {
			Description: "Make this output the default for agent monitoring.",
			Type:        schema.TypeBool,
			Optional:    true,
		},
		"ssl": {
			Description: "SSL configuration.",
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"certificate_authorities": {
						Description: "Server SSL certificate authorities.",
						Type:        schema.TypeList,
						Optional:    true,
						Elem: &schema.Schema{
							Type: schema.TypeString,
						},
					},
					"certificate": {
						Description: "Client SSL certificate.",
						Type:        schema.TypeString,
						Required:    true,
					},
					"key": {
						Description: "Client SSL certificate key.",
						Type:        schema.TypeString,
						Required:    true,
						Sensitive:   true,
					},
				},
			},
		},
		"config_yaml": {
			Description: "Advanced YAML configuration. YAML settings here will be added to the output section of each agent policy.",
			Type:        schema.TypeString,
			Optional:    true,
			Sensitive:   true,
		},
	}

	return &schema.Resource{
		Description: "Creates a new Fleet Output.",

		CreateContext: resourceOutputCreate,
		ReadContext:   resourceOutputRead,
		UpdateContext: resourceOutputUpdate,
		DeleteContext: resourceOutputDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: outputSchema,
	}
}

func resourceOutputCreateElasticsearch(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	fleetClient, diags := getFleetClient(d, meta)
	if diags.HasError() {
		return diags
	}

	reqData := fleetapi.OutputCreateRequestElasticsearch{
		Name: d.Get("name").(string),
		Type: fleetapi.OutputCreateRequestElasticsearchTypeElasticsearch,
	}

	var hosts []string
	if value := d.Get("hosts").([]interface{}); len(value) > 0 {
		for _, v := range value {
			if vStr, ok := v.(string); ok && vStr != "" {
				hosts = append(hosts, vStr)
			}
		}
	}
	if hosts != nil {
		reqData.Hosts = &hosts
	}
	if value := d.Get("default_integrations").(bool); value {
		reqData.IsDefault = &value
	}
	if value := d.Get("default_monitoring").(bool); value {
		reqData.IsDefaultMonitoring = &value
	}
	if value, ok := d.Get("ca_sha256").(string); ok && value != "" {
		reqData.CaSha256 = &value
	}
	if value, ok := d.Get("ca_trusted_fingerprint").(string); ok && value != "" {
		reqData.CaTrustedFingerprint = &value
	}
	if value, ok := d.Get("config_yaml").(string); ok && value != "" {
		reqData.ConfigYaml = &value
	}

	req := fleetapi.PostOutputsJSONRequestBody{}
	if err := req.FromOutputCreateRequestElasticsearch(reqData); err != nil {
		return diag.FromErr(err)
	}

	rawOutput, diags := fleet.CreateOutput(ctx, fleetClient, req)
	if diags.HasError() {
		return diags
	}

	output, err := rawOutput.AsOutputCreateRequestElasticsearch()
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(*output.Id)
	if err := d.Set("output_id", output.Id); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceOutputCreateLogstash(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	fleetClient, diags := getFleetClient(d, meta)
	if diags.HasError() {
		return diags
	}

	reqData := fleetapi.OutputCreateRequestLogstash{
		Name: d.Get("name").(string),
		Type: fleetapi.OutputCreateRequestLogstashTypeLogstash,
	}

	var hosts []string
	if value := d.Get("hosts").([]interface{}); len(value) > 0 {
		for _, v := range value {
			if vStr, ok := v.(string); ok && vStr != "" {
				hosts = append(hosts, vStr)
			}
		}
	}
	reqData.Hosts = hosts
	if value := d.Get("default_integrations").(bool); value {
		reqData.IsDefault = &value
	}
	if value := d.Get("default_monitoring").(bool); value {
		reqData.IsDefaultMonitoring = &value
	}
	if value, ok := d.Get("ca_sha256").(string); ok && value != "" {
		reqData.CaSha256 = &value
	}
	if value, ok := d.Get("ca_trusted_fingerprint").(string); ok && value != "" {
		reqData.CaTrustedFingerprint = &value
	}
	if value, ok := d.GetOk("ssl"); ok {
		ssl := value.([]interface{})[0].(map[string]interface{})
		reqData.Ssl = &struct {
			Certificate            *string   `json:"certificate,omitempty"`
			CertificateAuthorities *[]string `json:"certificate_authorities,omitempty"`
			Key                    *string   `json:"key,omitempty"`
		}{}
		if value, ok := ssl["certificate_authorities"].([]interface{}); ok {
			certs := make([]string, len(value))
			for i, v := range value {
				certs[i] = v.(string)
			}
			reqData.Ssl.CertificateAuthorities = &certs
		}
		if value, ok := ssl["certificate"].(string); ok {
			reqData.Ssl.Certificate = &value
		}
		if value, ok := ssl["key"].(string); ok {
			reqData.Ssl.Key = &value
		}
	}
	if value, ok := d.Get("config_yaml").(string); ok && value != "" {
		reqData.ConfigYaml = &value
	}

	req := fleetapi.PostOutputsJSONRequestBody{}
	if err := req.FromOutputCreateRequestLogstash(reqData); err != nil {
		return diag.FromErr(err)
	}

	rawOutput, diags := fleet.CreateOutput(ctx, fleetClient, req)
	if diags.HasError() {
		return diags
	}

	output, err := rawOutput.AsOutputCreateRequestElasticsearch()
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(*output.Id)
	if err := d.Set("output_id", output.Id); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceOutputCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	outputType := d.Get("type").(string)
	var diags diag.Diagnostics

	if id := d.Get("output_id").(string); id != "" {
		d.SetId(id)
	}

	switch outputType {
	case "elasticsearch":
		diags = resourceOutputCreateElasticsearch(ctx, d, meta)
	case "logstash":
		diags = resourceOutputCreateLogstash(ctx, d, meta)
	}
	if diags.HasError() {
		return diags
	}

	return resourceOutputRead(ctx, d, meta)
}

func resourceOutputUpdateElasticsearch(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	fleetClient, diags := getFleetClient(d, meta)
	if diags.HasError() {
		return diags
	}

	reqData := fleetapi.OutputUpdateRequestElasticsearch{
		Name: d.Get("name").(string),
		Type: fleetapi.OutputUpdateRequestElasticsearchTypeElasticsearch,
	}

	var hosts []string
	if value := d.Get("hosts").([]interface{}); len(value) > 0 {
		for _, v := range value {
			if vStr, ok := v.(string); ok && vStr != "" {
				hosts = append(hosts, vStr)
			}
		}
	}
	reqData.Hosts = hosts
	if value := d.Get("default_integrations").(bool); value {
		reqData.IsDefault = &value
	}
	if value := d.Get("default_monitoring").(bool); value {
		reqData.IsDefaultMonitoring = &value
	}
	if value, ok := d.Get("ca_sha256").(string); ok && value != "" {
		reqData.CaSha256 = &value
	}
	if value, ok := d.Get("config_yaml").(string); ok && value != "" {
		reqData.ConfigYaml = &value
	}

	req := fleetapi.UpdateOutputJSONRequestBody{}
	if err := req.FromOutputUpdateRequestElasticsearch(reqData); err != nil {
		return diag.FromErr(err)
	}

	_, diags = fleet.UpdateOutput(ctx, fleetClient, d.Id(), req)
	if diags.HasError() {
		return diags
	}

	return nil
}

func resourceOutputUpdateLogstash(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	fleetClient, diags := getFleetClient(d, meta)
	if diags.HasError() {
		return diags
	}

	reqData := fleetapi.OutputUpdateRequestLogstash{
		Name: d.Get("name").(string),
		Type: fleetapi.OutputUpdateRequestLogstashTypeLogstash,
	}

	var hosts []string
	if value := d.Get("hosts").([]interface{}); len(value) > 0 {
		for _, v := range value {
			if vStr, ok := v.(string); ok && vStr != "" {
				hosts = append(hosts, vStr)
			}
		}
	}
	if hosts != nil {
		reqData.Hosts = &hosts
	}
	if value := d.Get("default_integrations").(bool); value {
		reqData.IsDefault = &value
	}
	if value := d.Get("default_monitoring").(bool); value {
		reqData.IsDefaultMonitoring = &value
	}
	if value, ok := d.Get("ca_sha256").(string); ok && value != "" {
		reqData.CaSha256 = &value
	}
	if value, ok := d.GetOk("ssl"); ok {
		ssl := value.([]interface{})[0].(map[string]interface{})
		reqData.Ssl = &struct {
			Certificate            *string   `json:"certificate,omitempty"`
			CertificateAuthorities *[]string `json:"certificate_authorities,omitempty"`
			Key                    *string   `json:"key,omitempty"`
		}{}
		if value, ok := ssl["certificate_authorities"].([]interface{}); ok {
			certs := make([]string, len(value))
			for i, v := range value {
				certs[i] = v.(string)
			}
			reqData.Ssl.CertificateAuthorities = &certs
		}
		if value, ok := ssl["certificate"].(string); ok {
			reqData.Ssl.Certificate = &value
		}
		if value, ok := ssl["key"].(string); ok {
			reqData.Ssl.Key = &value
		}
	}
	if value, ok := d.Get("config_yaml").(string); ok && value != "" {
		reqData.ConfigYaml = &value
	}

	req := fleetapi.UpdateOutputJSONRequestBody{}
	if err := req.FromOutputUpdateRequestLogstash(reqData); err != nil {
		return diag.FromErr(err)
	}

	_, diags = fleet.UpdateOutput(ctx, fleetClient, d.Id(), req)
	if diags.HasError() {
		return diags
	}

	return nil
}

func resourceOutputUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	outputType := d.Get("type").(string)
	switch outputType {
	case "elasticsearch":
		diags = resourceOutputUpdateElasticsearch(ctx, d, meta)
	case "logstash":
		diags = resourceOutputUpdateLogstash(ctx, d, meta)
	}
	if diags.HasError() {
		return diags
	}

	return resourceOutputRead(ctx, d, meta)
}

func resourceOutputReadElasticsearch(d *schema.ResourceData, data fleetapi.OutputCreateRequestElasticsearch) diag.Diagnostics {
	if err := d.Set("type", "elasticsearch"); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("name", data.Name); err != nil {
		return diag.FromErr(err)
	}
	if data.Hosts != nil {
		if err := d.Set("hosts", *data.Hosts); err != nil {
			return diag.FromErr(err)
		}
	}
	if err := d.Set("default_integrations", data.IsDefault); err != nil {
		return diag.FromErr(err)
	}
	if data.IsDefaultMonitoring != nil {
		if err := d.Set("default_monitoring", *data.IsDefaultMonitoring); err != nil {
			return diag.FromErr(err)
		}
	}
	if data.CaSha256 != nil {
		if err := d.Set("ca_sha256", *data.CaSha256); err != nil {
			return diag.FromErr(err)
		}
	}
	if data.CaTrustedFingerprint != nil {
		if err := d.Set("ca_trusted_fingerprint", *data.CaTrustedFingerprint); err != nil {
			return diag.FromErr(err)
		}
	}
	if data.ConfigYaml != nil {
		if err := d.Set("config_yaml", *data.ConfigYaml); err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}

func resourceOutputReadLogstash(d *schema.ResourceData, data fleetapi.OutputCreateRequestLogstash) diag.Diagnostics {
	if err := d.Set("type", "logstash"); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("name", data.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("hosts", data.Hosts); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("default_integrations", data.IsDefault); err != nil {
		return diag.FromErr(err)
	}
	if data.IsDefaultMonitoring != nil {
		if err := d.Set("default_monitoring", *data.IsDefaultMonitoring); err != nil {
			return diag.FromErr(err)
		}
	}
	if data.CaSha256 != nil {
		if err := d.Set("ca_sha256", *data.CaSha256); err != nil {
			return diag.FromErr(err)
		}
	}
	if data.CaTrustedFingerprint != nil {
		if err := d.Set("ca_trusted_fingerprint", *data.CaTrustedFingerprint); err != nil {
			return diag.FromErr(err)
		}
	}
	if err := d.Set("ssl", flattenSslConfig(data)); err != nil {
		return diag.FromErr(err)
	}
	if data.ConfigYaml != nil {
		if err := d.Set("config_yaml", *data.ConfigYaml); err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}

func resourceOutputRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	fleetClient, diags := getFleetClient(d, meta)
	if diags.HasError() {
		return diags
	}

	rawOutput, diags := fleet.ReadOutput(ctx, fleetClient, d.Id())
	if diags.HasError() {
		return diags
	}
	// Not found.
	if rawOutput == nil {
		d.SetId("")
		return nil
	}

	output, err := rawOutput.ValueByDiscriminator()
	if err != nil {
		return diag.FromErr(err)
	}
	switch outputType := output.(type) {
	case fleetapi.OutputCreateRequestElasticsearch:
		diags = resourceOutputReadElasticsearch(d, outputType)
	case fleetapi.OutputCreateRequestLogstash:
		diags = resourceOutputReadLogstash(d, outputType)
	}
	if err := d.Set("output_id", d.Id()); err != nil {
		return diag.FromErr(err)
	}
	if diags.HasError() {
		return diags
	}

	return nil
}

func resourceOutputDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	fleetClient, diags := getFleetClient(d, meta)
	if diags.HasError() {
		return diags
	}

	if diags = fleet.DeleteOutput(ctx, fleetClient, d.Id()); diags.HasError() {
		return diags
	}
	d.SetId("")

	return diags
}

func flattenSslConfig(data fleetapi.OutputCreateRequestLogstash) []interface{} {
	if data.Ssl == nil {
		return []interface{}{}
	}

	ssl := make(map[string]interface{})
	if data.Ssl.CertificateAuthorities != nil {
		ssl["certificate_authorities"] = *data.Ssl.CertificateAuthorities
	}
	if data.Ssl.Certificate != nil {
		ssl["certificate"] = *data.Ssl.Certificate
	}
	if data.Ssl.Key != nil {
		ssl["key"] = *data.Ssl.Key
	}

	return []interface{}{ssl}
}
