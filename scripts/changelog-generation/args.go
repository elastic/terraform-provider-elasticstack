// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package main

// parseArgMap parses --key value pairs from a slice of CLI args.
// Keys are stored as-is (with hyphens preserved, no camelCase conversion).
func parseArgMap(args []string) map[string]string {
	m := make(map[string]string)
	for i := 0; i+1 < len(args); i += 2 {
		key := args[i]
		// Strip leading --
		if len(key) > 2 && key[0] == '-' && key[1] == '-' {
			key = key[2:]
		}
		m[key] = args[i+1]
	}
	return m
}

// argMapGet returns the value from the argMap if present, otherwise falls back
// to envVal, then to defaultVal.
func argMapGet(argMap map[string]string, key, envVal, defaultVal string) string {
	if v, ok := argMap[key]; ok {
		return v
	}
	if envVal != "" {
		return envVal
	}
	return defaultVal
}
