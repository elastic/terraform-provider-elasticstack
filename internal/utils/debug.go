package utils

import (
	"bytes"
	"encoding/json"
	"strings"
)

// PrettyPrintJSONLines iterates through a []byte line-by-line,
// transforming any lines that are complete json into pretty-printed json.
func PrettyPrintJSONLines(b []byte) string {
	parts := strings.Split(string(b), "\n")
	for i, p := range parts {
		if b := []byte(p); json.Valid(b) {
			var out bytes.Buffer
			if err := json.Indent(&out, b, "", " "); err != nil {
				continue
			}
			parts[i] = out.String()
		}
		// Mask Authorization header value
		if strings.Contains(strings.ToLower(p), "authorization:") {
			kv := strings.Split(p, ": ")
			if len(kv) != 2 {
				continue
			}
			kv[1] = strings.Repeat("*", len(kv[1]))
			parts[i] = strings.Join(kv, ": ")
		}
	}
	return strings.Join(parts, "\n")
}
