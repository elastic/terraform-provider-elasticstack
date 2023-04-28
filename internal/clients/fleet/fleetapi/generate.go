//go:build ignore
// +build ignore

package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

const (
	fleetOpenapiURL = "https://raw.githubusercontent.com/elastic/kibana/main/x-pack/plugins/fleet/common/openapi/bundled.json"
)

type OpenAPI struct {
	Other map[string]any            `yaml:",inline"`
	Paths map[string]map[string]any `yaml:"paths"`
}

type Path struct {
	Other      map[string]any   `yaml:",inline"`
	Parameters []map[string]any `yaml:"parameters,omitempty"`
}

var includePaths = map[string][]string{}

func downloadFile(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: HTTP %v: %v", resp.StatusCode, resp.Status)
	}

	return io.ReadAll(resp.Body)
}

func main() {
	outFile := flag.String("o", "", "output file")
	inFile := flag.String("i", "", "input file")
	flag.Parse()

	if *outFile == "" {
		flag.Usage()
		os.Exit(1)
	}

	var err error
	var openapiYAML []byte
	if *inFile != "" {
		openapiYAML, err = os.ReadFile(*inFile)
	} else {
		openapiYAML, err = downloadFile(fleetOpenapiURL)
	}
	if err != nil {
		log.Fatal(err)
	}

	var openapi M
	if err := json.Unmarshal(openapiYAML, &openapi); err != nil {
		log.Fatal(err)
	}

	openapiPaths := openapi["paths"].(map[string]any)
	for path, pathData := range openapiPaths {
		pathDataMap := pathData.(map[string]any)
		removeParametersKbnXsrf(pathDataMap)

		if allowedOps, found := includePaths[path]; found {
		nextOp:
			for op := range pathDataMap {
				for _, allowedOp := range allowedOps {
					if op == allowedOp || op == "parameters" {
						continue nextOp
					}
				}

				delete(pathDataMap, op)
			}
			continue
		}

		delete(openapiPaths, path)
	}

	f, err := os.Create(*outFile)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	openapi.Put("components.schemas.package_policy_request_input",
		openapi.Get("components.schemas.package_policy_request.properties.inputs.additionalProperties"))
	openapi.Put("components.schemas.package_policy_request_input_stream",
		openapi.Get("components.schemas.package_policy_request.properties.inputs.additionalProperties.properties.streams.additionalProperties"))

	openapi.Put("components.schemas.package_policy_request.properties.inputs.additionalProperties.properties.streams.additionalProperties", map[string]any{
		"$ref": "#/components/schemas/package_policy_request_input_stream",
	})
	openapi.Put("components.schemas.package_policy_request.properties.inputs.additionalProperties", map[string]any{
		"$ref": "#/components/schemas/package_policy_request_input",
	})

	// Add "policy_template" to input.
	openapi.Put("components.schemas.new_package_policy.properties.inputs.items.properties.policy_template", map[string]any{
		"type": "string",
	})
	openapi.Append("components.schemas.new_package_policy.properties.inputs.items.required", "policy_template")

	// Add properties to "streams".
	openapi.Put("components.schemas.new_package_policy.properties.inputs.items.properties.streams.items.properties", map[string]any{
		"enabled": M{
			"type": "boolean",
		},
		"vars": M{
			"type": "object",
		},
		"id": M{
			"type": "string",
		},
		"compiled_stream": M{
			"type": "object",
		},
		"data_stream": M{
			"type": "object",
			"properties": M{
				"type": M{
					"type": "string",
				},
				"dataset": M{
					"type": "string",
				},
			},
		},
	})

	// TODO: https://github.com/elastic/kibana/issues/151525
	openapi.Delete("components.schemas.new_agent_policy.properties.unenroll_timeout.nullable")
	openapi.Delete("components.schemas.new_agent_policy.properties.inactivity_timeout.nullable")

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err = enc.Encode(openapi); err != nil {
		log.Fatal(err)
	}
}

func removeParametersKbnXsrf(m map[string]any) {
	for k, v := range m {
		if k == "parameters" {
			if a, ok := v.([]interface{}); ok {
				n := 0
				for _, x := range a {
					if keep(x) {
						a[n] = x
						n++
					}
				}
				a = a[:n]

				m[k] = a
			}

			continue
		}

		switch tmp := v.(type) {
		case map[string]any:
			removeParametersKbnXsrf(tmp)
		case []any:
			for _, item := range tmp {
				if nestedMap, ok := item.(map[string]any); ok {
					removeParametersKbnXsrf(nestedMap)
				}
			}
		}
	}
}

func keep(v interface{}) bool {
	m, ok := v.(map[string]any)
	if !ok {
		return true
	}

	if ref, ok := m["$ref"].(string); ok && ref == "#/components/parameters/kbn_xsrf" {
		return false
	}

	return true
}

type M map[string]any

// Delete deletes the given key from the map.
func (m M) Delete(key string) error {
	k, d, _, found, err := mapFind(key, m, false)
	if err != nil {
		return err
	}
	if !found {
		return ErrKeyNotFound
	}

	delete(d, k)
	return nil
}

func (m M) Get(key string) interface{} {
	_, _, v, found, err := mapFind(key, m, false)
	if err != nil {
		panic(err)
	}
	if !found {
		return nil
	}
	return v
}

func (m M) Append(key string, value any) {
	v := m.Get(key)
	if v == nil {
		panic(key + " not found for append.")
	}
	list, ok := v.([]interface{})
	if !ok {
		panic(fmt.Errorf("%v is not a list (got %T)", key, v))
	}
	list = append(list, value)
	m.Put(key, list)
}

func (m M) Put(key string, value interface{}) (interface{}, error) {
	// XXX `safemapstr.Put` mimics this implementation, both should be updated to have similar behavior
	k, d, old, _, err := mapFind(key, m, true)
	if err != nil {
		return nil, err
	}

	d[k] = value
	return old, nil
}

func mapFind(
	key string,
	data M,
	createMissing bool,
) (subKey string, subMap M, oldValue interface{}, present bool, err error) {
	for {
		// Fast path, key is present as is.
		if v, exists := data[key]; exists {
			return key, data, v, true, nil
		}

		idx := strings.IndexRune(key, '.')
		if idx < 0 {
			return key, data, nil, false, nil
		}

		k := key[:idx]
		d, exists := data[k]
		if !exists {
			if createMissing {
				d = M{}
				data[k] = d
			} else {
				return "", nil, nil, false, ErrKeyNotFound
			}
		}

		v, err := toMapStr(d)
		if err != nil {
			return "", nil, nil, false, err
		}

		// advance to sub-map
		key = key[idx+1:]
		data = v
	}
}

func toMapStr(v interface{}) (M, error) {
	m, ok := tryToMapStr(v)
	if !ok {
		return nil, fmt.Errorf("expected map but type is %T", v)
	}
	return m, nil
}

func tryToMapStr(v interface{}) (M, bool) {
	switch m := v.(type) {
	case M:
		return m, true
	case map[string]interface{}:
		return M(m), true
	default:
		return nil, false
	}
}

// ErrKeyNotFound indicates that the specified key was not found.
var ErrKeyNotFound = errors.New("key not found")
