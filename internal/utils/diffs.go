package utils

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func DiffJsonSuppress(k, old, new string, d *schema.ResourceData) bool {
	result, _ := JSONBytesEqual([]byte(old), []byte(new))
	return result
}

func DiffIndexSettingSuppress(k, old, new string, d *schema.ResourceData) bool {
	var o, n map[string]interface{}
	if err := json.Unmarshal([]byte(old), &o); err != nil {
		return false
	}
	if err := json.Unmarshal([]byte(new), &n); err != nil {
		return false
	}
	return MapsEqual(NormalizeIndexSettings(FlattenMap(o)), NormalizeIndexSettings(FlattenMap(n)))
}

func NormalizeIndexSettings(m map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{}, len(m))
	for k, v := range m {
		if strings.HasPrefix(k, "index.") {
			out[k] = fmt.Sprintf("%v", v)
			continue
		}
		out[fmt.Sprintf("index.%s", k)] = fmt.Sprintf("%v", v)
	}
	return out
}

func DiffNullMapEntriesSuppress(key, old, new string, d *schema.ResourceData) bool {
	var oldMap, newMap map[string]interface{}
	if err := json.Unmarshal([]byte(old), &oldMap); err != nil {
		return false
	}
	if err := json.Unmarshal([]byte(new), &newMap); err != nil {
		return false
	}
	for _, m := range [...]map[string]interface{}{oldMap, newMap} {
		for k, v := range m {
			if v == nil {
				delete(m, k)
			}
		}
	}
	return MapsEqual(oldMap, newMap)
}
