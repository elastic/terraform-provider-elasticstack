//go:build ignore
// +build ignore

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

const (
	fleetSchemaURLTmpl = "https://raw.githubusercontent.com/elastic/kibana/%s/x-pack/plugins/fleet/common/openapi/bundled.json"
)

type Schema struct {
	Paths          map[string]*Path `json:"paths"`
	OpenAPIVersion string           `json:"openapi"`
	Tags           []any            `json:"tags,omitempty"`
	Servers        []any            `json:"servers,omitempty"`
	Components     Fields           `json:"components,omitempty"`
	Security       []any            `json:"security,omitempty"`
	Info           map[string]any   `json:"info"`
}

type Path struct {
	Parameters []Fields  `json:"parameters,omitempty"`
	Get        *Endpoint `json:"get,omitempty"`
	Post       *Endpoint `json:"post,omitempty"`
	Put        *Endpoint `json:"put,omitempty"`
	Delete     *Endpoint `json:"delete,omitempty"`
}

func (p *Path) GetEndpoint(method string) *Endpoint {
	switch strings.ToUpper(method) {
	case http.MethodGet:
		return p.Get
	case http.MethodPost:
		return p.Post
	case http.MethodPut:
		return p.Put
	case http.MethodDelete:
		return p.Delete
	}

	return nil
}

type Endpoint struct {
	Summary     string   `json:"summary,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	Responses   Fields   `json:"responses,omitempty"`
	RequestBody Fields   `json:"requestBody,omitempty"`
	OperationID string   `json:"operationId,omitempty"`
	Parameters  []Fields `json:"parameters,omitempty"`
	Deprecated  bool     `json:"deprecated,omitempty"`
}

type TransformFunc func(schema *Schema)

var transformers = []TransformFunc{
	transformFilterPaths,
	transformOutputTypeRequired,
	transformOutputResponseType,
}

// transformFilterPaths filters the paths in a schema down to
// a specified list of endpoints and methods.
func transformFilterPaths(schema *Schema) {
	var includePaths = map[string][]string{
		"/agent_policies":                 {"post"},
		"/agent_policies/{agentPolicyId}": {"get", "put"},
		"/agent_policies/delete":          {"post"},
		"/enrollment_api_keys":            {"get"},
		"/fleet_server_hosts":             {"post"},
		"/fleet_server_hosts/{itemId}":    {"get", "put", "delete"},
		"/outputs":                        {"post"},
		"/outputs/{outputId}":             {"get", "put", "delete"},
	}

	// filterKbnXsrfParameter filters out an entry if it is a kbn_xsrf parameter.
	// Returns a copy of the slice if it was modified, otherwise returns the original
	// slice if no match was found.
	filterKbnXsrfParameter := func(parameters []Fields) []Fields {
		removeIndex := -1

		for i, param := range parameters {
			if ref, ok := param["$ref"].(string); ok && ref == "#/components/parameters/kbn_xsrf" {
				removeIndex = i
				break
			}
		}
		if removeIndex != -1 {
			ret := make([]Fields, 0)
			ret = append(ret, parameters[:removeIndex]...)
			return append(ret, parameters[removeIndex+1:]...)
		}

		return parameters
	}

	for path, pathInfo := range schema.Paths {
		// Remove paths not in filter list.
		if _, exists := includePaths[path]; !exists {
			delete(schema.Paths, path)
			continue
		}

		// Filter out kbn-xsrf parameter (already set by API client).
		pathInfo.Parameters = filterKbnXsrfParameter(pathInfo.Parameters)

		// Filter out endpoints not if filter list, filter out kbn-xsrf
		// parameter in endpoint (already set by API client).
		allowedMethods := includePaths[path]
		filterEndpointFn := func(endpoint *Endpoint, method string) *Endpoint {
			if endpoint == nil {
				return nil
			}
			if !stringInSlice(method, allowedMethods) {
				return nil
			}

			endpoint.Parameters = filterKbnXsrfParameter(endpoint.Parameters)

			return endpoint
		}
		pathInfo.Get = filterEndpointFn(pathInfo.Get, "get")
		pathInfo.Post = filterEndpointFn(pathInfo.Post, "post")
		pathInfo.Put = filterEndpointFn(pathInfo.Put, "put")
		pathInfo.Delete = filterEndpointFn(pathInfo.Delete, "delete")
	}

	return
}

// transformOutputTypeRequired ensures that the type key is
// in the list of required keys for an output type.
func transformOutputTypeRequired(schema *Schema) {
	path := []string{
		"schemas.output_create_request_elasticsearch.required",
		"schemas.output_create_request_kafka.required",
		"schemas.output_create_request_logstash.required",
		"schemas.output_update_request_elasticsearch.required",
		"schemas.output_update_request_kafka.required",
		"schemas.output_update_request_logstash.required",
	}

	for _, v := range path {
		raw, ok := schema.Components.Get(v)
		if !ok {
			continue
		}
		required, ok := raw.([]any)
		if !ok {
			continue
		}

		if stringInAnySlice("type", required) {
			continue
		}

		required = append(required, "type")
		schema.Components.Set(v, required)
	}
}

// transformOutputTypeRequired ensures that the response object is wrapped
// in an `item` key/value pair. Remove once the following issue is closed:
// https://github.com/elastic/kibana/issues/167181
func transformOutputResponseType(schema *Schema) {
	methods := []string{http.MethodGet, http.MethodPut}
	for _, method := range methods {
		endpoint := schema.Paths["/outputs/{outputId}"].GetEndpoint(method)
		resSchema, ok := endpoint.Responses.GetFields("200.content.application/json.schema")
		if !ok {
			continue
		}
		ref, ok := resSchema.Get("$ref")
		if ok {
			resSchema.Set("type", "object")
			resSchema.Set("properties.item.$ref", ref)
			resSchema.Delete("$ref")
		}
	}
}

// downloadFile will download a file from url and return the
// bytes. If the request fails, or a non 200 error code is
// observed in the response, an error is returned instead.
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
	apiVersion := flag.String("v", "main", "api version")
	flag.Parse()

	if *outFile == "" {
		flag.Usage()
		os.Exit(1)
	}

	var err error
	var rawData []byte
	if *inFile != "" {
		rawData, err = os.ReadFile(*inFile)
	} else {
		rawData, err = downloadFile(fmt.Sprintf(fleetSchemaURLTmpl, *apiVersion))
	}
	if err != nil {
		log.Fatal(err)
	}

	var schema Schema
	if err = json.Unmarshal(rawData, &schema); err != nil {
		log.Fatal(err)
	}

	for _, fn := range transformers {
		fn(&schema)
	}

	outData, err := json.MarshalIndent(&schema, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	if err = os.WriteFile(*outFile, outData, 0664); err != nil {
		log.Fatal(err)
	}
}

// Fields wraps map[string]any with convenience functions for interacting
// with nested map values.
type Fields map[string]any

// Get will get the value at 'key' as the first returned
// parameter. The second parameter is a bool indicating
// if 'key' exists.
func (f Fields) Get(key string) (any, bool) {
	indexSliceFn := func(slice []any, key string) (any, string, bool) {
		indexStr, subKeys, _ := strings.Cut(key, ".")
		index, err := strconv.Atoi(indexStr)
		if err != nil {
			log.Printf("Failed to parse slice index key %q: %v", indexStr, err)
			return nil, "", false
		}

		if index < 0 || index >= len(slice) {
			log.Printf("Slice index is out of bounds (%d, target slice len: %d)", index, len(slice))
			return nil, "", false
		}

		return slice[index], subKeys, true
	}

	rootKey, subKeys, split := strings.Cut(key, ".")
	if split {
		switch t := f[rootKey].(type) {
		case Fields:
			return t.Get(subKeys)
		case map[string]any:
			return Fields(t).Get(subKeys)
		case []any:
			slicedValue, postSliceKeys, ok := indexSliceFn(t, subKeys)
			if !ok {
				return nil, false
			}
			if m, isMap := slicedValue.(map[string]any); ok && isMap {
				return Fields(m).Get(postSliceKeys)
			}
			return slicedValue, true

		default:
			rootKey = key
		}
	}

	value, ok := f[rootKey]
	return value, ok
}

// GetFields is like Get, but converts the found value to Fields.
// If the key is not found or the type conversion fails, the
// second return value will be false.
func (f Fields) GetFields(key string) (Fields, bool) {
	value, ok := f.Get(key)
	if !ok {
		return nil, false
	}

	switch t := value.(type) {
	case Fields:
		return t, true
	case map[string]any:
		return t, true
	}

	return nil, false
}

// Set will set key to the value of 'value'.
func (f Fields) Set(key string, value any) {
	rootKey, subKeys, split := strings.Cut(key, ".")
	if split {
		if v, ok := f[rootKey]; ok {
			switch t := v.(type) {
			case Fields:
				t.Set(subKeys, value)
			case map[string]any:
				Fields(t).Set(subKeys, value)
			}
		} else {
			subMap := Fields{}
			subMap.Set(subKeys, value)
			f[rootKey] = subMap
		}
	} else {
		f[rootKey] = value
	}
}

// Move will move the value from 'key' to 'target'. If 'key' does not
// exist, the operation is a no-op.
func (f Fields) Move(key, target string) {
	value, ok := f.Get(key)
	if !ok {
		return
	}

	f.Set(target, value)
	f.Delete(key)
}

// Delete will remove the key from the Fields. If key is nested,
// empty sub-keys will be removed as well.
func (f Fields) Delete(key string) {
	rootKey, subKeys, split := strings.Cut(key, ".")
	if split {
		if v, ok := f[rootKey]; ok {
			switch t := v.(type) {
			case Fields:
				t.Delete(subKeys)
			case map[string]any:
				Fields(t).Delete(subKeys)
			}
		}
	} else {
		delete(f, rootKey)
	}
}

// stringInSlice returns true if value is present in slice.
func stringInSlice(value string, slice []string) bool {
	for _, v := range slice {
		if value == v {
			return true
		}
	}

	return false
}

// stringInAnySlice returns true if value is present in slice.
func stringInAnySlice(value string, slice []any) bool {
	for _, v := range slice {
		s, ok := v.(string)
		if !ok {
			continue
		}
		if value == s {
			return true
		}
	}

	return false
}
