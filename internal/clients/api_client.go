package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type CompositeId struct {
	ClusterId  string
	ResourceId string
}

func CompositeIdFromStr(id string) (*CompositeId, diag.Diagnostics) {
	var diags diag.Diagnostics
	idParts := strings.Split(id, "/")
	if len(idParts) != 2 {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Wrong resource ID.",
			Detail:   "Resource ID must have following format: <cluster_uuid>/<resource identifier>",
		})
		return nil, diags
	}
	return &CompositeId{
			ClusterId:  idParts[0],
			ResourceId: idParts[1],
		},
		diags
}

func (c *CompositeId) String() string {
	return fmt.Sprintf("%s/%s", c.ClusterId, c.ResourceId)
}

type ApiClient struct {
	*elasticsearch.Client
}

func NewApiClientFunc(version string, p *schema.Provider) func(context.Context, *schema.ResourceData) (interface{}, diag.Diagnostics) {
	return func(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
		var diags diag.Diagnostics
		config := elasticsearch.Config{}

		if username, ok := d.GetOk("username"); ok {
			config.Username = username.(string)
			config.Password = d.Get("password").(string)
		}
		config.Addresses = []string{d.Get("url").(string)}

		es, err := elasticsearch.NewClient(config)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to create Elasticsearch client",
				Detail:   err.Error(),
			})
		}

		return &ApiClient{es}, diags
	}
}

func NewApiClient(d *schema.ResourceData, meta interface{}) (*ApiClient, error) {
	// if the config provided let's use it
	if esConn, ok := d.GetOk("elasticsearch_connection"); ok {
		config := elasticsearch.Config{}
		// there is always only 1 connection per resource
		conn := esConn.([]interface{})[0].(map[string]interface{})

		if u := conn["username"]; u != nil {
			config.Username = u.(string)
		}
		if p := conn["password"]; p != nil {
			config.Password = p.(string)
		}
		if endpoints := conn["endpoints"]; endpoints != nil {
			var addrs []string
			for _, e := range endpoints.([]interface{}) {
				addrs = append(addrs, e.(string))
			}
			config.Addresses = addrs
		}

		es, err := elasticsearch.NewClient(config)
		if err != nil {
			return nil, fmt.Errorf("Unable to create Elasticsearch client")
		}
		return &ApiClient{es}, nil
	} else { // or return the default client
		return meta.(*ApiClient), nil
	}
}

func (a *ApiClient) ID(resourceId string) (*CompositeId, diag.Diagnostics) {
	var diags diag.Diagnostics
	clusterId, err := a.ClusterID()
	if err != nil {
		return nil, diag.FromErr(err)
	}
	log.Printf("[TRACE] cluster UUID: %s", clusterId)
	return &CompositeId{clusterId, resourceId}, diags
}

func (a *ApiClient) ClusterID() (string, error) {
	res, err := a.Info()
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	if diags := utils.CheckError(res, "Unable to connect to the Elasticsearch cluster"); diags.HasError() {
		return "", fmt.Errorf("Unable to get cluster info")
	}

	info := make(map[string]interface{})
	if err := json.NewDecoder(res.Body).Decode(&info); err != nil {
		return "", err
	}
	if uuid := info["cluster_uuid"].(string); uuid != "" && uuid != "_na_" {
		log.Printf("[TRACE] cluster UUID: %s", uuid)
		return uuid, nil
	}

	return "", fmt.Errorf("Unable to get cluster uuid")
}
