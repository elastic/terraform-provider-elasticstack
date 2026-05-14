package contracttest

import "strings"

// tfAttrToAPICamel guesses a camelCase JSON key for a terraform snake_case attribute name.
func tfAttrToAPICamel(s string) string {
	parts := strings.Split(s, "_")
	for i := range parts {
		if parts[i] == "" {
			continue
		}
		if i == 0 {
			parts[i] = strings.ToLower(parts[i])
			continue
		}
		parts[i] = strings.ToUpper(parts[i][:1]) + strings.ToLower(parts[i][1:])
	}
	return strings.Join(parts, "")
}
