package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"reflect"

	"github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func CheckError(res *esapi.Response, errMsg string) diag.Diagnostics {
	var diags diag.Diagnostics

	if res.IsError() {
		body, err := io.ReadAll(res.Body)
		if err != nil {
			return diag.FromErr(err)
		}
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  errMsg,
			Detail:   fmt.Sprintf("Failed with: %s", body),
		})
		return diags
	}
	return diags
}

// Compares the JSON in two byte slices
func JSONBytesEqual(a, b []byte) (bool, error) {
	var j, j2 interface{}
	if err := json.Unmarshal(a, &j); err != nil {
		return false, err
	}
	if err := json.Unmarshal(b, &j2); err != nil {
		return false, err
	}
	return MapsEqual(j, j2), nil
}

func MapsEqual(m1, m2 interface{}) bool {
	return reflect.DeepEqual(m2, m1)
}

func DiffJsonSuppress(k, old, new string, d *schema.ResourceData) bool {
	result, _ := JSONBytesEqual([]byte(old), []byte(new))
	return result
}

// Returns the common connection schema for all the Elasticsearch resources,
// which defines the fields which can be used to configure the API access
func AddConnectionSchema(providedSchema map[string]*schema.Schema) {
	providedSchema["elasticsearch_connection"] = &schema.Schema{
		Description: "Used to establish connection to Elasticsearch server. Overrides environment variables if present",
		Type:        schema.TypeList,
		Optional:    true,
		MaxItems:    1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"username": {
					Description:  "A username to use for API authentication to Elasticsearch",
					Type:         schema.TypeString,
					Optional:     true,
					RequiredWith: []string{"elasticsearch_connection.0.password"},
				},
				"password": {
					Description:  "A password to use for API authentication to Elasticsearch",
					Type:         schema.TypeString,
					Optional:     true,
					Sensitive:    true,
					RequiredWith: []string{"elasticsearch_connection.0.username"},
				},
				"endpoints": {
					Description: "A list of endpoints where the terraform provider will point to. This must include the http(s) schema and port number.",
					Type:        schema.TypeList,
					Optional:    true,
					Sensitive:   true,
					Elem: &schema.Schema{
						Type: schema.TypeString,
					},
				},
			},
		},
	}
}
