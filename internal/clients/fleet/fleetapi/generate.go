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
)

const (
	fleetSchemaURLTmpl = "https://raw.githubusercontent.com/elastic/kibana/%s/x-pack/plugins/fleet/common/openapi/bundled.json"
)

type OpenAPISchema struct {
	Paths          map[string]*Path `json:"paths"`
	OpenAPIVersion string           `json:"openapi"`
	Tags           []any            `json:"tags,omitempty"`
	Servers        []any            `json:"servers,omitempty"`
	Components     map[string]any   `json:"components,omitempty"`
	Security       []any            `json:"security,omitempty"`
}

type Path struct {
	Parameters []map[string]any `json:"parameters,omitempty"`
	Get        *Endpoint        `json:"get,omitempty"`
	Post       *Endpoint        `json:"post,omitempty"`
	Put        *Endpoint        `json:"put,omitempty"`
	Delete     *Endpoint        `json:"delete,omitempty"`
}

type Endpoint struct {
	Summary     string           `json:"summary,omitempty"`
	Tags        []string         `json:"tags,omitempty"`
	Responses   map[string]any   `json:"responses,omitempty"`
	RequestBody map[string]any   `json:"requestBody,omitempty"`
	OperationID string           `json:"operationId,omitempty"`
	Parameters  []map[string]any `json:"parameters,omitempty"`
	Deprecated  bool             `json:"deprecated,omitempty"`
}

var includePaths = map[string][]string{
	"/agent_policies":                 {"post"},
	"/agent_policies/{agentPolicyId}": {"get", "put"},
	"/agent_policies/delete":          {"post"},
	"/enrollment_api_keys":            {"get"},
}

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

func stringInSlice(value string, slice []string) bool {
	for _, v := range slice {
		if v == value {
			return true
		}
	}

	return false
}

// filterKbnXsrfParameter filters out an entry if it is a kbn_xsrf parameter.
// Returns a copy of the slice if it was modified, otherwise returns the original
// slice if no match was found.
func filterKbnXsrfParameter(parameters []map[string]any) []map[string]any {
	removeIndex := -1

	for i, param := range parameters {
		if ref, ok := param["$ref"].(string); ok && ref == "#/components/parameters/kbn_xsrf" {
			removeIndex = i
			break
		}
	}
	if removeIndex != -1 {
		ret := make([]map[string]any, 0)
		ret = append(ret, parameters[:removeIndex]...)
		return append(ret, parameters[removeIndex+1:]...)
	}

	return parameters
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
	var schemaData []byte

	if *inFile != "" {
		schemaData, err = os.ReadFile(*inFile)
	} else {
		schemaData, err = downloadFile(fmt.Sprintf(fleetSchemaURLTmpl, *apiVersion))
	}
	if err != nil {
		log.Fatal(err)
	}

	var schema OpenAPISchema
	if err = json.Unmarshal(schemaData, &schema); err != nil {
		log.Fatal(err)
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

	outData, err := json.MarshalIndent(&schema, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	if err = os.WriteFile(*outFile, outData, 0664); err != nil {
		log.Fatal(err)
	}
}
