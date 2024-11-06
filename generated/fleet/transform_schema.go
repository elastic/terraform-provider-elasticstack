//go:build ignore
// +build ignore

package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"log"
	"maps"
	"os"
	"path"
	"reflect"
	"slices"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

func main() {
	_inFile := flag.String("i", "", "input file")
	_outFile := flag.String("o", "", "output file")
	flag.Parse()

	inFile := *_inFile
	outFile := *_outFile

	if inFile == "" || outFile == "" {
		flag.Usage()
		os.Exit(1)
	}

	outDir, _ := path.Split(outFile)
	if !pathExists(outDir) {
		if err := os.MkdirAll(outDir, 0755); err != nil {
			log.Fatalf("failed to create directory %q: %v", outDir, err)
		}
	}

	bytes, err := os.ReadFile(inFile)
	if err != nil {
		log.Fatalf("failed to read file %q: %v", inFile, err)
	}

	var schema Schema
	err = yaml.Unmarshal(bytes, &schema)
	if err != nil {
		log.Fatalf("failed to unmarshal schema from %q: %v", inFile, err)
	}

	// Run each transform
	for _, fn := range transformers {
		fn(&schema)
	}

	saveFile(schema, outFile)
}

// pathExists checks if path exists.
func pathExists(path string) bool {
	_, err := os.Stat(path)
	return !errors.Is(err, os.ErrNotExist)
}

// saveFile marshal and writes obj to path.
func saveFile(obj any, path string) {
	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(obj); err != nil {
		log.Fatalf("failed to marshal to file %q: %v", path, err)
	}

	if err := os.WriteFile(path, buf.Bytes(), 0664); err != nil {
		log.Fatalf("failed to write file %q: %v", path, err)
	}
}

// ============================================================================

type Schema struct {
	Paths      map[string]*Path `yaml:"paths"`
	Version    string           `yaml:"openapi"`
	Tags       []Map            `yaml:"tags,omitempty"`
	Servers    []Map            `yaml:"servers,omitempty"`
	Components Map              `yaml:"components,omitempty"`
	Security   []Map            `yaml:"security,omitempty"`
	Info       Map              `yaml:"info"`
}

func (s Schema) GetPath(path string) *Path {
	return s.Paths[path]
}

func (s Schema) MustGetPath(path string) *Path {
	p := s.GetPath(path)
	if p == nil {
		log.Panicf("Path not found: %q", path)
	}
	return p
}

// ============================================================================

type Path struct {
	Parameters []Map `yaml:"parameters,omitempty"`
	Get        Map   `yaml:"get,omitempty"`
	Post       Map   `yaml:"post,omitempty"`
	Put        Map   `yaml:"put,omitempty"`
	Delete     Map   `yaml:"delete,omitempty"`
}

func (p Path) Endpoints(yield func(key string, endpoint Map) bool) {
	if p.Get != nil {
		yield("get", p.Get)
	}
	if p.Post != nil {
		yield("post", p.Post)
	}
	if p.Put != nil {
		yield("put", p.Put)
	}
	if p.Delete != nil {
		yield("delete", p.Delete)
	}
}

func (p Path) GetEndpoint(method string) Map {
	switch method {
	case "get":
		return p.Get
	case "post":
		return p.Post
	case "put":
		return p.Put
	case "delete":
		return p.Delete
	default:
		log.Panicf("Unhandled method: %q", method)
	}
	return nil
}

func (p Path) MustGetEndpoint(method string) Map {
	endpoint := p.GetEndpoint(method)
	if endpoint == nil {
		log.Panicf("Method not found: %q", method)
	}
	return endpoint
}

func (p *Path) SetEndpoint(method string, endpoint Map) {
	switch method {
	case "get":
		p.Get = endpoint
	case "post":
		p.Post = endpoint
	case "put":
		p.Put = endpoint
	case "delete":
		p.Delete = endpoint
	default:
		log.Panicf("Invalid method %q", method)
	}
}

// ============================================================================

type Map map[string]any

func (m Map) Keys() []string {
	keys := slices.Collect(maps.Keys(m))
	slices.Sort(keys)
	return keys
}

func (m Map) Has(key string) bool {
	_, ok := m.Get(key)
	return ok
}

func (m Map) Get(key string) (any, bool) {
	rootKey, subKeys, found := strings.Cut(key, ".")
	if found {
		switch t := m[rootKey].(type) {
		case Map:
			return t.Get(subKeys)
		case map[string]any:
			return Map(t).Get(subKeys)
		case Slice:
			return t.Get(subKeys)
		case []any:
			return Slice(t).Get(subKeys)
		default:
			rootKey = key
		}
	}

	value, ok := m[rootKey]
	return value, ok
}

func (m Map) MustGet(key string) any {
	v, ok := m.Get(key)
	if !ok {
		log.Panicf("%q not found", key)
	}
	return v
}

func (m Map) GetSlice(key string) (Slice, bool) {
	value, ok := m.Get(key)
	if !ok {
		return nil, false
	}

	switch t := value.(type) {
	case Slice:
		return t, true
	case []any:
		return t, true
	}

	log.Panicf("%q is not a slice", key)
	return nil, false
}

func (m Map) MustGetSlice(key string) Slice {
	v, ok := m.GetSlice(key)
	if !ok {
		log.Panicf("%q not found", key)
	}
	return v
}

func (m Map) GetMap(key string) (Map, bool) {
	value, ok := m.Get(key)
	if !ok {
		return nil, false
	}

	switch t := value.(type) {
	case Map:
		return t, true
	case map[string]any:
		return t, true
	}

	log.Panicf("%q is not a map", key)
	return nil, false
}

func (m Map) MustGetMap(key string) Map {
	v, ok := m.GetMap(key)
	if !ok {
		log.Panicf("%q not found", key)
	}
	return v
}

func (m Map) Set(key string, value any) {
	rootKey, subKeys, found := strings.Cut(key, ".")
	if found {
		if v, ok := m[rootKey]; ok {
			switch t := v.(type) {
			case Slice:
				t.Set(subKeys, value)
			case []any:
				Slice(t).Set(subKeys, value)
			case Map:
				t.Set(subKeys, value)
			case map[string]any:
				Map(t).Set(subKeys, value)
			}
		} else {
			subMap := Map{}
			subMap.Set(subKeys, value)
			m[rootKey] = subMap
		}
	} else {
		m[rootKey] = value
	}
}

func (m Map) Move(src string, dst string) {
	value := m.MustGet(src)
	m.Set(dst, value)
	m.Delete(src)
}

func (m Map) Delete(key string) bool {
	rootKey, subKeys, found := strings.Cut(key, ".")
	if found {
		if v, ok := m[rootKey]; ok {
			switch t := v.(type) {
			case Slice:
				return t.Delete(subKeys)
			case []any:
				return Slice(t).Delete(subKeys)
			case Map:
				return t.Delete(subKeys)
			case map[string]any:
				return Map(t).Delete(subKeys)
			}
		}
	} else {
		delete(m, rootKey)
		return true
	}
	return false
}

func (m Map) MustDelete(key string) {
	if !m.Delete(key) {
		log.Panicf("%q not found", key)
	}
}

func (m Map) CreateRef(schema *Schema, name string, key string) Map {
	refTarget := m.MustGet(key) // Check the full path
	refPath := fmt.Sprintf("schemas.%s", name)
	refValue := Map{"$ref": fmt.Sprintf("#/components/schemas/%s", name)}

	// If the component schema already exists and is not the same, panic
	writeComponent := true
	if existing, ok := schema.Components.Get(refPath); ok {
		if reflect.DeepEqual(refTarget, existing) {
			writeComponent = false
		} else {
			log.Panicf("Component schema key already in use and not an exact duplicate: %q", refPath)
			return nil
		}
	}

	var parent any
	var childKey string
	// Get the parent of the refTarget
	i := strings.LastIndex(key, ".")
	if i == -1 {
		parent = m
		childKey = key
	} else {
		parent = m.MustGet(key[:i])
		childKey = key[i+1:]
	}

	doMap := func(target Map, key string) {
		if writeComponent {
			schema.Components.Set(refPath, target.MustGet(key))
		}
		target.Set(key, refValue)
	}

	doSlice := func(target Slice, key string) {
		index := target.atoi(key)
		if writeComponent {
			schema.Components.Set(refPath, target[index])
		}
		target[index] = refValue
	}

	switch t := parent.(type) {
	case map[string]any:
		doMap(Map(t), childKey)
	case Map:
		doMap(t, childKey)
	case []any:
		doSlice(Slice(t), childKey)
	case Slice:
		doSlice(t, childKey)
	default:
		log.Panicf("Cannot create a ref of target type %T at %q", parent, key)
	}

	return refValue
}

func (m Map) Iterate(iteratee func(key string, node Map)) {
	joinPath := func(existing string, next string) string {
		if existing == "" {
			return next
		} else {
			return fmt.Sprintf("%s.%s", existing, next)
		}
	}
	joinIndex := func(existing string, next int) string {
		if existing == "" {
			return fmt.Sprintf("%d", next)
		} else {
			return fmt.Sprintf("%s.%d", existing, next)
		}
	}

	var iterate func(key string, val any)
	iterate = func(key string, val any) {
		switch tval := val.(type) {
		case []any:
			iterate(key, Slice(tval))
		case Slice:
			for i, v := range tval {
				iterate(joinIndex(key, i), v)
			}
		case map[string]any:
			iterate(key, Map(tval))
		case Map:
			for _, k := range tval.Keys() {
				iterate(joinPath(key, k), tval[k])
			}
			iteratee(key, tval)
		}
	}

	iterate("", m)
}

// ============================================================================

type Slice []any

func (s Slice) Get(key string) (any, bool) {
	rootKey, subKeys, found := strings.Cut(key, ".")
	index := s.atoi(rootKey)

	if found {
		switch t := s[index].(type) {
		case Slice:
			return t.Get(subKeys)
		case []any:
			return Slice(t).Get(subKeys)
		case Map:
			return t.Get(subKeys)
		case map[string]any:
			return Map(t).Get(subKeys)
		}
	}

	value := s[index]
	return value, true
}

func (s Slice) GetMap(key string) (Map, bool) {
	value, ok := s.Get(key)
	if !ok {
		return nil, false
	}

	switch t := value.(type) {
	case Map:
		return t, true
	case map[string]any:
		return t, true
	}

	log.Panicf("%q is not a map", key)
	return nil, false
}

func (s Slice) MustGetMap(key string) Map {
	v, ok := s.GetMap(key)
	if !ok {
		log.Panicf("%q not found", key)
	}
	return v
}

func (s Slice) Set(key string, value any) {
	rootKey, subKeys, found := strings.Cut(key, ".")
	index := s.atoi(rootKey)
	if found {
		v := s[index]
		switch t := v.(type) {
		case Slice:
			t.Set(subKeys, value)
		case []any:
			Slice(t).Set(subKeys, value)
		case Map:
			t.Set(subKeys, value)
		case map[string]any:
			Map(t).Set(subKeys, value)
		}
	} else {
		s[index] = value
	}
}

func (s Slice) Delete(key string) bool {
	rootKey, subKeys, found := strings.Cut(key, ".")
	index := s.atoi(rootKey)
	if found {
		item := (s)[index]
		switch t := item.(type) {
		case Slice:
			return t.Delete(subKeys)
		case []any:
			return Slice(t).Delete(subKeys)
		case Map:
			return t.Delete(subKeys)
		case map[string]any:
			return Map(t).Delete(subKeys)
		}
	} else {
		log.Panicf("Unable to delete from slice directly")
		return true
	}
	return false
}

func (s Slice) Contains(value string) bool {
	for _, v := range s {
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

func (s Slice) atoi(key string) int {
	index, err := strconv.Atoi(key)
	if err != nil {
		log.Panicf("Failed to parse slice index key %q: %v", key, err)
	}
	if index < 0 || index >= len(s) {
		log.Panicf("Slice index is out of bounds (%d, target slice len: %d)", index, len(s))
	}
	return index
}

// ============================================================================

type TransformFunc func(schema *Schema)

var transformers = []TransformFunc{
	transformFilterPaths,
	transformRemoveKbnXsrf,
	transformRemoveApiVersionParam,
	transformSimplifyContentType,
	transformFleetPaths,
	// transformRemoveEnums,
	// transformAddGoPointersFlag,
	transformRemoveExamples,
	transformRemoveUnusedComponents,
}

// transformFilterPaths filters the paths in a schema down to a specified list
// of endpoints and methods.
func transformFilterPaths(schema *Schema) {
	var includePaths = map[string][]string{
		"/api/fleet/agent_policies":                      {"get", "post"},
		"/api/fleet/agent_policies/delete":               {"post"},
		"/api/fleet/agent_policies/{agentPolicyId}":      {"get", "put"},
		"/api/fleet/enrollment_api_keys":                 {"get"},
		"/api/fleet/epm/packages":                        {"get", "post"},
		"/api/fleet/epm/packages/{pkgName}/{pkgVersion}": {"get", "post", "delete"},
		"/api/fleet/fleet_server_hosts":                  {"get", "post"},
		"/api/fleet/fleet_server_hosts/{itemId}":         {"get", "put", "delete"},
		"/api/fleet/outputs":                             {"get", "post"},
		"/api/fleet/outputs/{outputId}":                  {"get", "put", "delete"},
		"/api/fleet/package_policies":                    {"get", "post"},
		"/api/fleet/package_policies/{packagePolicyId}":  {"get", "put", "delete"},
	}

	for path, pathInfo := range schema.Paths {
		if allowedMethods, ok := includePaths[path]; ok {
			// Filter out endpoints not if filter list
			for method := range pathInfo.Endpoints {
				if !slices.Contains(allowedMethods, method) {
					pathInfo.SetEndpoint(method, nil)
				}
			}
		} else {
			// Remove paths not in filter list.
			delete(schema.Paths, path)
		}
	}

	// Go through again, verify each entry exists
	for path, methods := range includePaths {
		pathInfo := schema.GetPath(path)
		if pathInfo == nil {
			log.Panicf("Missing path %q", path)
		}

		for _, method := range methods {
			endpoint := pathInfo.GetEndpoint(method)
			if endpoint == nil {
				log.Panicf("Missing method %q of %q", method, path)
			}
		}
	}
}

// transformRemoveKbnXsrf removes the kbn-xsrf header as it	is already applied
// in the client.
func transformRemoveKbnXsrf(schema *Schema) {
	removeKbnXsrf := func(node any) bool {
		param := node.(Map)
		if v, ok := param["name"]; ok {
			name := v.(string)
			if strings.HasSuffix(name, "kbn_xsrf") || strings.HasSuffix(name, "kbn-xsrf") {
				return true
			}
		}
		// Data_views_kbn_xsrf, Saved_objects_kbn_xsrf, etc
		if v, ok := param["$ref"]; ok {
			ref := v.(string)
			if strings.HasSuffix(ref, "kbn_xsrf") || strings.HasSuffix(ref, "kbn-xsrf") {
				return true
			}
		}
		return false
	}

	for _, pathInfo := range schema.Paths {
		for _, endpoint := range pathInfo.Endpoints {
			if params, ok := endpoint.GetSlice("parameters"); ok {
				params = slices.DeleteFunc(params, removeKbnXsrf)
				endpoint["parameters"] = params
			}
		}
	}
}

// transformRemoveApiVersionParam removes the Elastic API Version query
// parameter header.
func transformRemoveApiVersionParam(schema *Schema) {
	removeApiVersion := func(node any) bool {
		param := node.(Map)
		if name, ok := param["name"]; ok && name == "elastic-api-version" {
			return true
		}
		return false
	}

	for _, pathInfo := range schema.Paths {
		for _, endpoint := range pathInfo.Endpoints {
			if params, ok := endpoint.GetSlice("parameters"); ok {
				params = slices.DeleteFunc(params, removeApiVersion)
				endpoint["parameters"] = params
			}
		}
	}
}

// transformSimplifyContentType simplifies Content-Type headers such as
// 'application/json; Elastic-Api-Version=2023-10-31' by stripping everything
// after the ';'.
func transformSimplifyContentType(schema *Schema) {
	simplifyContentType := func(fields Map) {
		if content, ok := fields.GetMap("content"); ok {
			for key := range content {
				newKey, _, found := strings.Cut(key, ";")
				if found {
					content.Move(key, newKey)
				}
			}
		}
	}

	for _, pathInfo := range schema.Paths {
		for _, endpoint := range pathInfo.Endpoints {
			if req, ok := endpoint.GetMap("requestBody"); ok {
				simplifyContentType(req)
			}
			if resp, ok := endpoint.GetMap("responses"); ok {
				for code := range resp {
					simplifyContentType(resp.MustGetMap(code))
				}
			}
		}
	}

	if responses, ok := schema.Components.GetMap("responses"); ok {
		for key := range responses {
			resp := responses.MustGetMap(key)
			simplifyContentType(resp)
		}
	}
}

// transformFleetPaths fixes the fleet paths.
func transformFleetPaths(schema *Schema) {
	operationIds := map[string]map[string]string{
		"/api/fleet/agent_policies": {
			"get":  "get_agent_policies",
			"post": "create_agent_policy",
		},
		"/api/fleet/agent_policies/delete": {
			"post": "delete_agent_policy",
		},
		"/api/fleet/agent_policies/{agentPolicyId}": {
			"get": "get_agent_policy",
			"put": "update_agent_policy",
		},
		"/api/fleet/enrollment_api_keys": {
			"get": "get_enrollment_api_keys",
		},
		"/api/fleet/epm/packages": {
			"get":  "list_packages",
			"post": "install_package_by_upload",
		},
		"/api/fleet/epm/packages/{pkgName}/{pkgVersion}": {
			"get":    "get_package",
			"post":   "install_package",
			"delete": "delete_package",
		},
		"/api/fleet/fleet_server_hosts": {
			"get":  "get_fleet_server_hosts",
			"post": "create_fleet_server_host",
		},
		"/api/fleet/fleet_server_hosts/{itemId}": {
			"get":    "get_fleet_server_host",
			"put":    "update_fleet_server_host",
			"delete": "delete_fleet_server_host",
		},
		"/api/fleet/outputs": {
			"get":  "get_outputs",
			"post": "create_output",
		},
		"/api/fleet/outputs/{outputId}": {
			"get":    "get_output",
			"put":    "update_output",
			"delete": "delete_output",
		},
		"/api/fleet/package_policies": {
			"get":  "get_package_policies",
			"post": "create_package_policy",
		},
		"/api/fleet/package_policies/{packagePolicyId}": {
			"get":    "get_package_policy",
			"put":    "update_package_policy",
			"delete": "delete_package_policy",
		},
	}

	// Set each missing operationId
	for path, methods := range operationIds {
		pathInfo := schema.MustGetPath(path)
		for method, operationId := range methods {
			endpoint := pathInfo.GetEndpoint(method)
			endpoint.Set("operationId", operationId)
		}
	}

	// Fix OpenAPI error: set each missing description
	for _, pathInfo := range schema.Paths {
		for _, endpoint := range pathInfo.Endpoints {
			responses := endpoint.MustGetMap("responses")
			for code := range responses {
				response := responses.MustGetMap(code)
				if _, ok := response["description"]; !ok {
					response["description"] = ""
				}
			}
		}
	}

	// Agent policies
	// https://github.com/elastic/kibana/blob/main/x-pack/plugins/fleet/common/types/models/agent_policy.ts
	// https://github.com/elastic/kibana/blob/main/x-pack/plugins/fleet/common/types/rest_spec/agent_policy.ts

	agentPoliciesPath := schema.MustGetPath("/api/fleet/agent_policies")
	agentPolicyPath := schema.MustGetPath("/api/fleet/agent_policies/{agentPolicyId}")

	agentPoliciesPath.Get.CreateRef(schema, "agent_policy", "responses.200.content.application/json.schema.properties.items.items")
	agentPoliciesPath.Post.CreateRef(schema, "agent_policy", "responses.200.content.application/json.schema.properties.item")
	agentPolicyPath.Get.CreateRef(schema, "agent_policy", "responses.200.content.application/json.schema.properties.item")
	agentPolicyPath.Put.CreateRef(schema, "agent_policy", "responses.200.content.application/json.schema.properties.item")

	// See: https://github.com/elastic/kibana/issues/197155
	// [request body.keep_monitoring_alive]: expected value of type [boolean] but got [null]
	// [request body.supports_agentless]: expected value of type [boolean] but got [null]
	// [request body.overrides]: expected value of type [boolean] but got [null]
	for _, key := range []string{"keep_monitoring_alive", "supports_agentless", "overrides"} {
		agentPoliciesPath.Post.Set(fmt.Sprintf("requestBody.content.application/json.schema.properties.%s.x-omitempty", key), true)
		agentPolicyPath.Put.Set(fmt.Sprintf("requestBody.content.application/json.schema.properties.%s.x-omitempty", key), true)
	}

	// Enrollment api keys
	// https://github.com/elastic/kibana/blob/main/x-pack/plugins/fleet/common/types/models/enrollment_api_key.ts
	// https://github.com/elastic/kibana/blob/main/x-pack/plugins/fleet/common/types/rest_spec/enrollment_api_key.ts

	apiKeysPath := schema.MustGetPath("/api/fleet/enrollment_api_keys")
	apiKeysPath.Get.CreateRef(schema, "enrollment_api_key", "responses.200.content.application/json.schema.properties.items.items")

	// EPM
	// https://github.com/elastic/kibana/blob/main/x-pack/plugins/fleet/common/types/models/epm.ts
	// https://github.com/elastic/kibana/blob/main/x-pack/plugins/fleet/common/types/rest_spec/epm.ts

	packagesPath := schema.MustGetPath("/api/fleet/epm/packages")
	packagePath := schema.MustGetPath("/api/fleet/epm/packages/{pkgName}/{pkgVersion}")
	packagesPath.Get.CreateRef(schema, "package_list_item", "responses.200.content.application/json.schema.properties.items.items")
	packagePath.Get.CreateRef(schema, "package_info", "responses.200.content.application/json.schema.properties.item")

	// Server hosts
	// https://github.com/elastic/kibana/blob/main/x-pack/plugins/fleet/common/types/models/fleet_server_policy_config.ts
	// https://github.com/elastic/kibana/blob/main/x-pack/plugins/fleet/common/types/rest_spec/fleet_server_hosts.ts

	hostsPath := schema.MustGetPath("/api/fleet/fleet_server_hosts")
	hostPath := schema.MustGetPath("/api/fleet/fleet_server_hosts/{itemId}")

	hostsPath.Get.CreateRef(schema, "server_host", "responses.200.content.application/json.schema.properties.items.items")
	hostsPath.Post.CreateRef(schema, "server_host", "responses.200.content.application/json.schema.properties.item")
	hostPath.Get.CreateRef(schema, "server_host", "responses.200.content.application/json.schema.properties.item")
	hostPath.Put.CreateRef(schema, "server_host", "responses.200.content.application/json.schema.properties.item")

	// 8.6.2 regression
	// [request body.proxy_id]: definition for this key is missing
	// See: https://github.com/elastic/kibana/issues/197155
	hostsPath.Post.Set("requestBody.content.application/json.schema.properties.proxy_id.x-omitempty", true)
	hostPath.Put.Set("requestBody.content.application/json.schema.properties.proxy_id.x-omitempty", true)

	// Outputs
	// https://github.com/elastic/kibana/blob/main/x-pack/plugins/fleet/common/types/models/output.ts
	// https://github.com/elastic/kibana/blob/main/x-pack/plugins/fleet/common/types/rest_spec/output.ts

	outputByIdPath := schema.MustGetPath("/api/fleet/outputs/{outputId}")
	outputsPath := schema.MustGetPath("/api/fleet/outputs")

	outputsPath.Post.CreateRef(schema, "new_output_union", "requestBody.content.application/json.schema")
	outputByIdPath.Put.CreateRef(schema, "update_output_union", "requestBody.content.application/json.schema")
	outputsPath.Get.CreateRef(schema, "output_union", "responses.200.content.application/json.schema.properties.items.items")
	outputByIdPath.Get.CreateRef(schema, "output_union", "responses.200.content.application/json.schema.properties.item")
	outputsPath.Post.CreateRef(schema, "output_union", "responses.200.content.application/json.schema.properties.item")
	outputByIdPath.Put.CreateRef(schema, "output_union", "responses.200.content.application/json.schema.properties.item")

	for _, name := range []string{"output", "new_output", "update_output"} {
		// Ref each index in the anyOf union
		schema.Components.CreateRef(schema, fmt.Sprintf("%s_elasticsearch", name), fmt.Sprintf("schemas.%s_union.anyOf.0", name))
		schema.Components.CreateRef(schema, fmt.Sprintf("%s_remote_elasticsearch", name), fmt.Sprintf("schemas.%s_union.anyOf.1", name))
		schema.Components.CreateRef(schema, fmt.Sprintf("%s_logstash", name), fmt.Sprintf("schemas.%s_union.anyOf.2", name))
		schema.Components.CreateRef(schema, fmt.Sprintf("%s_kafka", name), fmt.Sprintf("schemas.%s_union.anyOf.3", name))

		// Extract child structs
		for _, typ := range []string{"elasticsearch", "remote_elasticsearch", "logstash", "kafka"} {
			schema.Components.CreateRef(schema, fmt.Sprintf("%s_shipper", name), fmt.Sprintf("schemas.%s_%s.properties.shipper", name, typ))
			schema.Components.CreateRef(schema, fmt.Sprintf("%s_ssl", name), fmt.Sprintf("schemas.%s_%s.properties.ssl", name, typ))
		}

		// Ideally just remove the "anyOf", however then we would need to make
		// refs for each of the "oneOf" options. So turn them into an "any" instead.
		// See: https://github.com/elastic/kibana/issues/197153
		/*
			anyOf:
			  - items: {}
			    type: array
			  - type: boolean
			  - type: number
			  - type: object
			  - type: string
			nullable: true
			oneOf:
			  - type: number
			  - not: {}
		*/

		props := schema.Components.MustGetMap(fmt.Sprintf("schemas.%s_kafka.properties", name))
		for _, key := range []string{"compression_level", "connection_type", "password", "username"} {
			props.Set(key, Map{})
		}
	}

	// Add the missing discriminator to the response union
	// See: https://github.com/elastic/kibana/issues/181994
	schema.Components.Set("schemas.output_union.discriminator", Map{
		"propertyName": "type",
		"mapping": Map{
			"elasticsearch":        "#/components/schemas/output_elasticsearch",
			"remote_elasticsearch": "#/components/schemas/output_remote_elasticsearch",
			"logstash":             "#/components/schemas/output_logstash",
			"kafka":                "#/components/schemas/output_kafka",
		},
	})

	for _, name := range []string{"new_output", "update_output"} {
		for _, typ := range []string{"elasticsearch", "remote_elasticsearch", "logstash", "kafka"} {
			// [request body.1.ca_sha256]: expected value of type [string] but got [null]"
			// See: https://github.com/elastic/kibana/issues/197155
			schema.Components.Set(fmt.Sprintf("schemas.%s_%s.properties.ca_sha256.x-omitempty", name, typ), true)

			// [request body.1.ca_trusted_fingerprint]: expected value of type [string] but got [null]
			// See: https://github.com/elastic/kibana/issues/197155
			schema.Components.Set(fmt.Sprintf("schemas.%s_%s.properties.ca_trusted_fingerprint.x-omitempty", name, typ), true)

			// 8.6.2 regression
			// [request body.proxy_id]: definition for this key is missing"
			// See: https://github.com/elastic/kibana/issues/197155
			schema.Components.Set(fmt.Sprintf("schemas.%s_%s.properties.proxy_id.x-omitempty", name, typ), true)
		}

		// [request body.1.shipper]: expected a plain object value, but found [null] instead
		// See: https://github.com/elastic/kibana/issues/197155
		schema.Components.Set(fmt.Sprintf("schemas.%s_shipper.x-omitempty", name), true)

		// [request body.1.ssl]: expected a plain object value, but found [null] instead
		// See: https://github.com/elastic/kibana/issues/197155
		schema.Components.Set(fmt.Sprintf("schemas.%s_ssl.x-omitempty", name), true)

	}

	for _, typ := range []string{"elasticsearch", "remote_elasticsearch", "logstash", "kafka"} {
		// strict_dynamic_mapping_exception: [1:345] mapping set to strict, dynamic introduction of [id] within [ingest-outputs] is not allowed"
		// See: https://github.com/elastic/kibana/issues/197155
		schema.Components.MustDelete(fmt.Sprintf("schemas.update_output_%s.properties.id", typ))
	}

	// Package policies
	// https://github.com/elastic/kibana/blob/main/x-pack/plugins/fleet/common/types/models/package_policy.ts
	// https://github.com/elastic/kibana/blob/main/x-pack/plugins/fleet/common/types/rest_spec/package_policy.ts

	epmPoliciesPath := schema.MustGetPath("/api/fleet/package_policies")
	epmPolicyPath := schema.MustGetPath("/api/fleet/package_policies/{packagePolicyId}")

	epmPoliciesPath.Get.CreateRef(schema, "package_policy", "responses.200.content.application/json.schema.properties.items.items")
	epmPoliciesPath.Post.CreateRef(schema, "package_policy", "responses.200.content.application/json.schema.properties.item")

	epmPoliciesPath.Post.Move("requestBody.content.application/json.schema.anyOf.1", "requestBody.content.application/json.schema") // anyOf.0 is the deprecated array format
	epmPolicyPath.Put.Move("requestBody.content.application/json.schema.anyOf.1", "requestBody.content.application/json.schema")    // anyOf.0 is the deprecated array format
	epmPoliciesPath.Post.CreateRef(schema, "package_policy_request", "requestBody.content.application/json.schema")
	epmPolicyPath.Put.CreateRef(schema, "package_policy_request", "requestBody.content.application/json.schema")

	epmPolicyPath.Get.CreateRef(schema, "package_policy", "responses.200.content.application/json.schema.properties.item")
	epmPolicyPath.Put.CreateRef(schema, "package_policy", "responses.200.content.application/json.schema.properties.item")

	schema.Components.CreateRef(schema, "package_policy_secret_ref", "schemas.package_policy.properties.secret_references.items")
	schema.Components.Move("schemas.package_policy.properties.inputs.anyOf.1", "schemas.package_policy.properties.inputs") // anyOf.0 is the deprecated array format

	schema.Components.CreateRef(schema, "package_policy_input", "schemas.package_policy.properties.inputs.additionalProperties")
	schema.Components.CreateRef(schema, "package_policy_input_stream", "schemas.package_policy_input.properties.streams.additionalProperties")

	schema.Components.CreateRef(schema, "package_policy_request_package", "schemas.package_policy_request.properties.package")
	schema.Components.CreateRef(schema, "package_policy_request_input", "schemas.package_policy_request.properties.inputs.additionalProperties")
	schema.Components.CreateRef(schema, "package_policy_request_input_stream", "schemas.package_policy_request_input.properties.streams.additionalProperties")

	// Simplify all of the vars
	schema.Components.Set("schemas.package_policy.properties.vars", Map{"type": "object"})
	schema.Components.Set("schemas.package_policy_input.properties.vars", Map{"type": "object"})
	schema.Components.Set("schemas.package_policy_input_stream.properties.vars", Map{"type": "object"})
	schema.Components.Set("schemas.package_policy_request.properties.vars", Map{"type": "object"})
	schema.Components.Set("schemas.package_policy_request_input.properties.vars", Map{"type": "object"})
	schema.Components.Set("schemas.package_policy_request_input_stream.properties.vars", Map{"type": "object"})

	// [request body.0.output_id]: expected value of type [string] but got [null]
	// [request body.1.output_id]: definition for this key is missing"
	// See: https://github.com/elastic/kibana/issues/197155
	schema.Components.Set("schemas.package_policy_request.properties.output_id.x-omitempty", true)
}

// transformRemoveEnums remove all enums.
func transformRemoveEnums(schema *Schema) {
	deleteEnumFn := func(key string, node Map) {
		if node.Has("enum") {
			delete(node, "enum")
		}
	}

	for _, pathInfo := range schema.Paths {
		for _, methInfo := range pathInfo.Endpoints {
			methInfo.Iterate(deleteEnumFn)
		}
	}
	schema.Components.Iterate(deleteEnumFn)
}

// transformRemoveExamples removes all examples.
func transformRemoveExamples(schema *Schema) {
	deleteExampleFn := func(key string, node Map) {
		if node.Has("example") {
			delete(node, "example")
		}
		if node.Has("examples") {
			delete(node, "examples")
		}
	}

	for _, pathInfo := range schema.Paths {
		for _, methInfo := range pathInfo.Endpoints {
			methInfo.Iterate(deleteExampleFn)
		}
	}
	schema.Components.Iterate(deleteExampleFn)
	schema.Components.Set("examples", Map{})
}

// transformAddOptionalPointersFlag adds a x-go-type-skip-optional-pointer
// flag to maps and arrays, since they are already nullable types.
func transformAddOptionalPointersFlag(schema *Schema) {
	addFlagFn := func(key string, node Map) {
		if node["type"] == "array" {
			node["x-go-type-skip-optional-pointer"] = true
		} else if node["type"] == "object" {
			if _, ok := node["properties"]; !ok {
				node["x-go-type-skip-optional-pointer"] = true
			}
		}
	}

	for _, pathInfo := range schema.Paths {
		for _, methInfo := range pathInfo.Endpoints {
			methInfo.Iterate(addFlagFn)
		}
	}
	schema.Components.Iterate(addFlagFn)
}

// transformRemoveUnusedComponents removes all unused schema components.
func transformRemoveUnusedComponents(schema *Schema) {
	var refs map[string]any
	collectRefsFn := func(key string, node Map) {
		if ref, ok := node["$ref"].(string); ok {
			i := strings.LastIndex(ref, "/")
			ref = ref[i+1:]
			refs[ref] = nil
		}
	}

	componentParams := schema.Components.MustGetMap("parameters")
	componentSchemas := schema.Components.MustGetMap("schemas")

	for {
		// Collect refs
		refs = make(map[string]any)
		for _, pathInfo := range schema.Paths {
			for _, methInfo := range pathInfo.Endpoints {
				methInfo.Iterate(collectRefsFn)
			}
		}
		schema.Components.Iterate(collectRefsFn)

		loop := false
		for key := range componentSchemas {
			if _, ok := refs[key]; !ok {
				delete(componentSchemas, key)
				loop = true
			}
		}
		for key := range componentParams {
			if _, ok := refs[key]; !ok {
				delete(componentParams, key)
				loop = true
			}
		}
		if !loop {
			break
		}
	}
}
