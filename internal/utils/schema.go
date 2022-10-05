package utils

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

func ExpandStringSet(set *schema.Set) []string {
	var strs []string
	for _, v := range set.List() {
		strs = append(strs, v.(string))
	}
	return strs
}
