package tfsdkutils

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	schemautil "github.com/elastic/terraform-provider-elasticstack/internal/utils"
)

func DiffJSONSuppress(_, old, newValue string, _ *schema.ResourceData) bool {
	result, _ := schemautil.JSONBytesEqual([]byte(old), []byte(newValue))
	return result
}

func DiffIndexSettingSuppress(_, old, newValue string, _ *schema.ResourceData) bool {
	var o, n map[string]any
	if err := json.Unmarshal([]byte(old), &o); err != nil {
		return false
	}
	if err := json.Unmarshal([]byte(newValue), &n); err != nil {
		return false
	}
	return reflect.DeepEqual(normalizeIndexSettings(flattenMap(o)), normalizeIndexSettings(flattenMap(n)))
}

func normalizeIndexSettings(m map[string]any) map[string]any {
	out := make(map[string]any, len(m))
	for k, v := range m {
		if strings.HasPrefix(k, "index.") {
			out[k] = fmt.Sprintf("%v", v)
			continue
		}
		out[fmt.Sprintf("index.%s", k)] = fmt.Sprintf("%v", v)
	}
	return out
}

// Flattens the multilevel map, and concatenates keys together with dot "."
// # Examples
// map of form:
//
//	map := map[string]interface{}{
//	        "index": map[string]interface{}{
//	                "key": 1
//	        }
//	}
//
// becomes:
//
//	map := map[string]interface{}{
//	        "index.key": 1
//	}
func flattenMap(m map[string]any) map[string]any {
	out := make(map[string]any)

	var flattener func(string, map[string]any, map[string]any)
	flattener = func(k string, src, dst map[string]any) {
		if len(k) > 0 {
			k += "."
		}
		for key, v := range src {
			switch inner := v.(type) {
			case map[string]any:
				flattener(k+key, inner, dst)
			default:
				dst[k+key] = v
			}
		}
	}
	flattener("", m, out)
	return out
}
