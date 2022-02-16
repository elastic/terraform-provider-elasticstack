package utils

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strings"

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

// Flattens the multilevel map, and concatenates keys together with dot "."
// # Exmaples
// map of form:
//     map := map[string]interface{}{
//             "index": map[string]interface{}{
//                     "key": 1
//             }
//     }
// becomes:
//     map := map[string]interface{}{
//             "index.key": 1
//     }
func FlattenMap(m map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{})

	var flattener func(string, map[string]interface{}, map[string]interface{})
	flattener = func(k string, src, dst map[string]interface{}) {
		if len(k) > 0 {
			k += "."
		}
		for key, v := range src {
			switch inner := v.(type) {
			case map[string]interface{}:
				flattener(k+key, inner, dst)
			default:
				dst[k+key] = v
			}
		}
	}
	flattener("", m, out)
	return out
}

func MergeSchemaMaps(maps ...map[string]*schema.Schema) map[string]*schema.Schema {
	result := make(map[string]*schema.Schema)
	for _, m := range maps {
		for k, v := range m {
			result[k] = v
		}
	}
	return result
}

func IsEmpty(v interface{}) bool {
	switch t := v.(type) {
	case int, int8, int16, int32, int64, float32, float64:
		if t == 0 {
			return true
		}
	case string:
		if strings.TrimSpace(t) == "" {
			return true
		}
	case []interface{}:
		if len(t) == 0 {
			return true
		}
	case map[interface{}]interface{}:
		if len(t) == 0 {
			return true
		}
	case nil:
		return true
	}
	return false
}

// Returns the common connection schema for all the Elasticsearch resources,
// which defines the fields which can be used to configure the API access
func AddConnectionSchema(providedSchema map[string]*schema.Schema) {
	providedSchema["elasticsearch_connection"] = &schema.Schema{
		Description: "Used to establish connection to Elasticsearch server. Overrides environment variables if present.",
		Type:        schema.TypeList,
		Optional:    true,
		MaxItems:    1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"username": {
					Description:  "A username to use for API authentication to Elasticsearch.",
					Type:         schema.TypeString,
					Optional:     true,
					RequiredWith: []string{"elasticsearch_connection.0.password"},
				},
				"password": {
					Description:  "A password to use for API authentication to Elasticsearch.",
					Type:         schema.TypeString,
					Optional:     true,
					Sensitive:    true,
					RequiredWith: []string{"elasticsearch_connection.0.username"},
				},
				"endpoints": {
					Description: "A list of endpoints the Terraform provider will point to. They must include the http(s) schema and port number.",
					Type:        schema.TypeList,
					Optional:    true,
					Sensitive:   true,
					Elem: &schema.Schema{
						Type: schema.TypeString,
					},
				},
				"insecure": {
					Description: "Disable TLS certificate validation",
					Type:        schema.TypeBool,
					Optional:    true,
					Default:     false,
				},
				"ca_file": {
					Description: "Path to a custom Certificate Authority certificate",
					Type:        schema.TypeString,
					Optional:    true,
				},
			},
		},
	}
}

func StringToHash(s string) (*string, error) {
	h := sha1.New()
	_, err := h.Write([]byte(s))
	if err != nil {
		return nil, err
	}
	bs := h.Sum(nil)
	hash := fmt.Sprintf("%x", bs)
	return &hash, nil
}
