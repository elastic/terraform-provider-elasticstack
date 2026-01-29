//go:build ignore
// +build ignore

package main

import (
	"bytes"
	_ "embed"
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
	Patch      Map   `yaml:"patch,omitempty"`
	Post       Map   `yaml:"post,omitempty"`
	Put        Map   `yaml:"put,omitempty"`
	Delete     Map   `yaml:"delete,omitempty"`
}

func (p Path) Endpoints(yield func(key string, endpoint Map) bool) {
	if p.Get != nil {
		if !yield("get", p.Get) {
			return
		}
	}
	if p.Post != nil {
		if !yield("post", p.Post) {
			return
		}
	}
	if p.Put != nil {
		if !yield("put", p.Put) {
			return
		}
	}
	if p.Patch != nil {
		if !yield("patch", p.Patch) {
			return
		}
	}
	if p.Delete != nil {
		if !yield("delete", p.Delete) {
			return
		}
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
	case "patch":
		return p.Patch
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
	case "patch":
		p.Patch = endpoint
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
	mergeDashboardsSchema,
	transformRemoveKbnXsrf,
	transformRemoveApiVersionParam,
	transformSimplifyContentType,
	transformAddMisingDescriptions,
	transformKibanaPaths,
	transformFleetPaths,
	removeBrokenDiscriminator,
	fixPutSecurityRoleName,
	fixGetSpacesParams,
	fixGetSyntheticsMonitorsParams,
	fixGetMaintenanceWindowFindParams,
	fixGetStreamsAttachmentTypesParams,
	fixSecurityAPIPageSize,
	fixSecurityExceptionListItems,
	removeDuplicateOneOfRefs,
	fixDashboardPanelItemRefs,
	transformRemoveExamples,
	transformRemoveUnusedComponents,
	transformOmitEmptyNullable,
}

//go:embed dashboards.yaml
var dashboardsYaml string

func mergeDashboardsSchema(schema *Schema) {
	var dashboardsSchema Schema
	err := yaml.Unmarshal([]byte(dashboardsYaml), &dashboardsSchema)
	if err != nil {
		log.Fatalf("failed to unmarshal schema from dashboards.yaml: %v", err)
	}

	// Merge paths
	for path, pathInfo := range dashboardsSchema.Paths {
		// Only add the path if it doesn't already exist
		if _, ok := schema.Paths[path]; !ok {
			schema.Paths[path] = pathInfo
		}
	}

	// Merge component schemas
	dashboardSchemas := dashboardsSchema.Components.MustGetMap("schemas")
	schemaSchemas := schema.Components.MustGetMap("schemas")
	for key, schemaInfo := range dashboardSchemas {
		// Only add the schema if it doesn't already exist
		if _, ok := schemaSchemas[key]; !ok {
			schemaSchemas[key] = schemaInfo
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

// transformAddMisingDescriptions adds descriptions to each path missing one.
func transformAddMisingDescriptions(schema *Schema) {
	for _, pathInfo := range schema.Paths {
		for _, endpoint := range pathInfo.Endpoints {
			responses, ok := endpoint.GetMap("responses")
			if !ok {
				continue
			}

			for code := range responses {
				response := responses.MustGetMap(code)
				if _, ok := response["description"]; !ok {
					response["description"] = ""
				}
			}
		}
	}
}

// transformKibanaPaths fixes the Kibana paths.
func transformKibanaPaths(schema *Schema) {
	// Convert any paths needing it to /s/{spaceId} variants
	spaceIdPaths := []string{
		"/api/data_views",
		"/api/data_views/data_view",
		"/api/data_views/data_view/{viewId}",
		"/api/maintenance_window",
		"/api/maintenance_window/{id}",
		"/api/actions/connector/{id}",
		"/api/actions/connectors",
		"/api/data_views/default",
		"/api/detection_engine/rules",
		"/api/exception_lists",
		"/api/exception_lists/items",
		"/api/lists",
		"/api/lists/index",
		"/api/lists/items",
	}

	// Add a spaceId parameter if not already present
	if _, ok := schema.Components.Get("parameters.spaceId"); !ok {
		schema.Components.Set("parameters.spaceId", Map{
			"in":          "path",
			"name":        "spaceId",
			"description": "An identifier for the space. If `/s/` and the identifier are omitted from the path, the default space is used.",
			"required":    true,
			"schema":      Map{"type": "string", "example": "default"},
		})
	}

	for _, path := range spaceIdPaths {
		pathInfo := schema.Paths[path]
		schema.Paths[fmt.Sprintf("/s/{spaceId}%s", path)] = pathInfo
		delete(schema.Paths, path)

		// Add the spaceId parameter
		param := Map{"$ref": "#/components/parameters/spaceId"}
		for _, endpoint := range pathInfo.Endpoints {
			if params, ok := endpoint.GetSlice("parameters"); ok {
				params = append(params, param)
				endpoint.Set("parameters", params)
			} else {
				params = Slice{param}
				endpoint.Set("parameters", params)
			}
		}
	}

	// Connectors
	// Can be removed when https://github.com/elastic/kibana/issues/230149 is addressed.
	connectorPath := schema.MustGetPath("/s/{spaceId}/api/actions/connector/{id}")
	connectorsPath := schema.MustGetPath("/s/{spaceId}/api/actions/connectors")

	connectorPath.Post.CreateRef(schema, "create_connector_config", "requestBody.content.application/json.schema.properties.config")
	connectorPath.Post.CreateRef(schema, "create_connector_secrets", "requestBody.content.application/json.schema.properties.secrets")

	connectorPath.Put.CreateRef(schema, "update_connector_config", "requestBody.content.application/json.schema.properties.config")
	connectorPath.Put.CreateRef(schema, "update_connector_secrets", "requestBody.content.application/json.schema.properties.secrets")

	connectorPath.Get.CreateRef(schema, "connector_response", "responses.200.content.application/json.schema")
	connectorsPath.Get.Set("responses.200.content.application/json.schema", Map{
		"type": "array",
		"items": Map{
			"$ref": "#/components/schemas/connector_response",
		},
	})

	// Data views
	// https://github.com/elastic/kibana/blob/main/src/plugins/data_views/server/rest_api_routes/schema.ts

	dataViewsPath := schema.MustGetPath("/s/{spaceId}/api/data_views")

	dataViewsPath.Get.CreateRef(schema, "get_data_views_response_item", "responses.200.content.application/json.schema.properties.data_view.items")

	sytheticsParamsPath := schema.MustGetPath("/api/synthetics/params")
	sytheticsParamsPath.Post.CreateRef(schema, "create_param_response", "responses.200.content.application/json.schema")

	schema.Components.CreateRef(schema, "Data_views_data_view_response_object_inner", "schemas.Data_views_data_view_response_object.properties.data_view")
	schema.Components.CreateRef(schema, "Data_views_sourcefilter_item", "schemas.Data_views_sourcefilters.items")
	schema.Components.CreateRef(schema, "Data_views_runtimefieldmap_script", "schemas.Data_views_runtimefieldmap.properties.script")

	schema.Components.Set("schemas.Data_views_fieldformats.additionalProperties", Map{
		"$ref": "#/components/schemas/Data_views_fieldformat",
	})
	schema.Components.Set("schemas.Data_views_fieldformat", Map{
		"type": "object",
		"properties": Map{
			"id":     Map{"type": "string"},
			"params": Map{"$ref": "#/components/schemas/Data_views_fieldformat_params"},
		},
	})
	schema.Components.Set("schemas.Data_views_fieldformat_params", Map{
		"type": "object",
		"properties": Map{
			"pattern":                Map{"type": "string"},
			"urlTemplate":            Map{"type": "string"},
			"labelTemplate":          Map{"type": "string"},
			"inputFormat":            Map{"type": "string"},
			"outputFormat":           Map{"type": "string"},
			"outputPrecision":        Map{"type": "integer"},
			"includeSpaceWithSuffix": Map{"type": "boolean"},
			"useShortSuffix":         Map{"type": "boolean"},
			"timezone":               Map{"type": "string"},
			"fieldType":              Map{"type": "string"},
			"colors": Map{
				"type":  "array",
				"items": Map{"$ref": "#/components/schemas/Data_views_fieldformat_params_color"},
			},
			"fieldLength": Map{"type": "integer"},
			"transform":   Map{"type": "string"},
			"lookupEntries": Map{
				"type":  "array",
				"items": Map{"$ref": "#/components/schemas/Data_views_fieldformat_params_lookup"},
			},
			"unknownKeyValue": Map{"type": "string"},
			"type":            Map{"type": "string"},
			"width":           Map{"type": "integer"},
			"height":          Map{"type": "integer"},
		},
	})
	schema.Components.Set("schemas.Data_views_fieldformat_params_color", Map{
		"type": "object",
		"properties": Map{
			"range":      Map{"type": "string"},
			"regex":      Map{"type": "string"},
			"text":       Map{"type": "string"},
			"background": Map{"type": "string"},
		},
	})
	schema.Components.Set("schemas.Data_views_fieldformat_params_lookup", Map{
		"type": "object",
		"properties": Map{
			"key":   Map{"type": "string"},
			"value": Map{"type": "string"},
		},
	})

	schema.Components.CreateRef(schema, "Data_views_create_data_view_request_object_inner", "schemas.Data_views_create_data_view_request_object.properties.data_view")
	schema.Components.CreateRef(schema, "Data_views_update_data_view_request_object_inner", "schemas.Data_views_update_data_view_request_object.properties.data_view")

	schema.Components.Set("schemas.Security_Detections_API_RuleResponse.discriminator", Map{
		"mapping": Map{
			"eql":              "#/components/schemas/Security_Detections_API_EqlRule",
			"esql":             "#/components/schemas/Security_Detections_API_EsqlRule",
			"machine_learning": "#/components/schemas/Security_Detections_API_MachineLearningRule",
			"new_terms":        "#/components/schemas/Security_Detections_API_NewTermsRule",
			"query":            "#/components/schemas/Security_Detections_API_QueryRule",
			"saved_query":      "#/components/schemas/Security_Detections_API_SavedQueryRule",
			"threat_match":     "#/components/schemas/Security_Detections_API_ThreatMatchRule",
			"threshold":        "#/components/schemas/Security_Detections_API_ThresholdRule",
		},
		"propertyName": "type",
	})

	schema.Components.Set("schemas.Security_Detections_API_RuleCreateProps.discriminator", Map{
		"mapping": Map{
			"eql":              "#/components/schemas/Security_Detections_API_EqlRuleCreateProps",
			"esql":             "#/components/schemas/Security_Detections_API_EsqlRuleCreateProps",
			"machine_learning": "#/components/schemas/Security_Detections_API_MachineLearningRuleCreateProps",
			"new_terms":        "#/components/schemas/Security_Detections_API_NewTermsRuleCreateProps",
			"query":            "#/components/schemas/Security_Detections_API_QueryRuleCreateProps",
			"saved_query":      "#/components/schemas/Security_Detections_API_SavedQueryRuleCreateProps",
			"threat_match":     "#/components/schemas/Security_Detections_API_ThreatMatchRuleCreateProps",
			"threshold":        "#/components/schemas/Security_Detections_API_ThresholdRuleCreateProps",
		},
		"propertyName": "type",
	})

	schema.Components.Set("schemas.Security_Detections_API_RuleUpdateProps.discriminator", Map{
		"mapping": Map{
			"eql":              "#/components/schemas/Security_Detections_API_EqlRuleUpdateProps",
			"esql":             "#/components/schemas/Security_Detections_API_EsqlRuleUpdateProps",
			"machine_learning": "#/components/schemas/Security_Detections_API_MachineLearningRuleUpdateProps",
			"new_terms":        "#/components/schemas/Security_Detections_API_NewTermsRuleUpdateProps",
			"query":            "#/components/schemas/Security_Detections_API_QueryRuleUpdateProps",
			"saved_query":      "#/components/schemas/Security_Detections_API_SavedQueryRuleUpdateProps",
			"threat_match":     "#/components/schemas/Security_Detections_API_ThreatMatchRuleUpdateProps",
			"threshold":        "#/components/schemas/Security_Detections_API_ThresholdRuleUpdateProps",
		},
		"propertyName": "type",
	})

	schema.Components.Set("schemas.Security_Detections_API_ResponseAction.discriminator", Map{
		"mapping": Map{
			".osquery":  "#/components/schemas/Security_Detections_API_OsqueryResponseAction",
			".endpoint": "#/components/schemas/Security_Detections_API_EndpointResponseAction",
		},
		"propertyName": "action_type_id",
	})
	schema.Components.Delete("schemas.Security_Exceptions_API_ExceptionListItemExpireTime.format")

}

func removeBrokenDiscriminator(schema *Schema) {
	brokenDiscriminatorPaths := map[string]string{
		"/api/detection_engine/rules/preview": "post",
		"/api/synthetics/monitors":            "post",
		"/api/synthetics/monitors/{id}":       "put",
	}

	brokenDiscriminatorComponents := []string{
		"Security_AI_Assistant_API_KnowledgeBaseEntryCreateProps",
		"Security_AI_Assistant_API_KnowledgeBaseEntryResponse",
		"Security_AI_Assistant_API_KnowledgeBaseEntryUpdateProps",
		"Security_AI_Assistant_API_KnowledgeBaseEntryUpdateRouteProps",
		"Security_Detections_API_RuleSource",
		"Security_Endpoint_Exceptions_API_ExceptionListItemEntry",
		"Security_Exceptions_API_ExceptionListItemEntry",
		"Security_Endpoint_Management_API_ActionDetailsResponse",
	}

	for _, component := range brokenDiscriminatorComponents {
		schema.Components.Delete(fmt.Sprintf("schemas.%s.discriminator", component))
	}

	for path, method := range brokenDiscriminatorPaths {
		schema.MustGetPath(path).MustGetEndpoint(method).Delete("requestBody.content.application/json.schema.discriminator")
	}
}

func fixPutSecurityRoleName(schema *Schema) {
	putEndpoint := schema.MustGetPath("/api/security/role/{name}").MustGetEndpoint("put")
	putEndpoint.Delete("requestBody.content.application/json.schema.properties.kibana.items.properties.base.anyOf")
	putEndpoint.Move("requestBody.content.application/json.schema.properties.kibana.items.properties.spaces.anyOf.1", "requestBody.content.application/json.schema.properties.kibana.items.properties.spaces")

	postEndpoint := schema.MustGetPath("/api/security/roles").MustGetEndpoint("post")
	postEndpoint.Move("requestBody.content.application/json.schema.properties.roles.additionalProperties", "requestBody.content.application/json.schema.properties.roles")
	postEndpoint.Delete("requestBody.content.application/json.schema.properties.roles.properties.kibana.items.properties.base.anyOf")
	postEndpoint.Move("requestBody.content.application/json.schema.properties.roles.properties.kibana.items.properties.spaces.anyOf.1", "requestBody.content.application/json.schema.properties.roles.properties.kibana.items.properties.spaces")
}

func fixGetSpacesParams(schema *Schema) {
	schema.MustGetPath("/api/spaces/space").MustGetEndpoint("get").Delete("parameters.1.schema.anyOf")
}

func fixGetSyntheticsMonitorsParams(schema *Schema) {
	schema.MustGetPath("/api/synthetics/monitors").MustGetEndpoint("get").Set("parameters.12.schema.oneOf.1.x-go-type", "[]GetSyntheticMonitorsParamsUseLogicalAndFor0")
}

func fixGetMaintenanceWindowFindParams(schema *Schema) {
	schema.MustGetPath("/api/maintenance_window/_find").MustGetEndpoint("get").Set("parameters.2.schema.anyOf.1.x-go-type", "[]GetMaintenanceWindowFindParamsStatus0")
}

func fixGetStreamsAttachmentTypesParams(schema *Schema) {
	schema.MustGetPath("/api/streams/{streamName}/attachments").MustGetEndpoint("get").Set("parameters.2.schema.anyOf.1.x-go-type", "[]GetStreamsStreamnameAttachmentsParamsAttachmentTypes0")
}

func fixSecurityAPIPageSize(schema *Schema) {
	apiPageSize := schema.Components.MustGetMap("schemas.Security_Endpoint_Management_API_ApiPageSize")
	schema.Components.Set("schemas.Security_Endpoint_Management_API_ApiPageSize", apiPageSize.MustGetMap("allOf.0"))
}

func fixDashboardPanelItemRefs(schema *Schema) {
	dashboardsPath := schema.MustGetPath("/api/dashboards")
	dashboardPath := schema.MustGetPath("/api/dashboards/{id}")

	dashboardsPath.Post.CreateRef(schema, "dashboard_panel_item", "requestBody.content.application/json.schema.properties.data.properties.panels.items.anyOf.0")
	dashboardsPath.Post.CreateRef(schema, "dashboard_panel_section", "requestBody.content.application/json.schema.properties.data.properties.panels.items.anyOf.1")
	dashboardsPath.Post.CreateRef(schema, "dashboard_panels", "requestBody.content.application/json.schema.properties.data.properties.panels")

	dashboardPath.Put.CreateRef(schema, "dashboard_panel_item", "requestBody.content.application/json.schema.properties.data.properties.panels.items.anyOf.0")
	dashboardPath.Put.CreateRef(schema, "dashboard_panel_section", "requestBody.content.application/json.schema.properties.data.properties.panels.items.anyOf.1")
	dashboardPath.Put.CreateRef(schema, "dashboard_panels", "requestBody.content.application/json.schema.properties.data.properties.panels")

	dashboardPath.Get.CreateRef(schema, "dashboard_panel_item", "responses.200.content.application/json.schema.properties.data.properties.panels.items.anyOf.0")
	dashboardPath.Get.CreateRef(schema, "dashboard_panel_section", "responses.200.content.application/json.schema.properties.data.properties.panels.items.anyOf.1")
	dashboardPath.Get.CreateRef(schema, "dashboard_panels", "responses.200.content.application/json.schema.properties.data.properties.panels")

	schema.Components.CreateRef(schema, "dashboard_panel_item", "schemas.dashboard_panel_section.properties.panels.items")
}

func fixSecurityExceptionListItems(schema *Schema) {
	exceptionListItems := schema.MustGetPath("/s/{spaceId}/api/exception_lists/items")

	putExceptionListItem := exceptionListItems.MustGetEndpoint("put")
	putExceptionListItem.CreateRef(schema, "Security_Exceptions_API_UpdateExceptionListItem", "requestBody.content.application/json.schema")

	postExceptionListItem := exceptionListItems.MustGetEndpoint("post")
	postExceptionListItem.CreateRef(schema, "Security_Exceptions_API_CreateExceptionListItem", "requestBody.content.application/json.schema")
}

func removeDuplicateOneOfRefs(schema *Schema) {
	componentSchemas := schema.Components.MustGetMap("schemas")
	componentSchemas.Iterate(removeDuplicateOneOfRefsFromNode)
}

// https://github.com/elastic/kibana/issues/244264
func removeDuplicateOneOfRefsFromNode(key string, node Map) {
	maybeOneOf, hasOneOf := node.GetSlice("oneOf")
	if hasOneOf {
		// Check for duplicate $ref entries
		seenRefs := map[string]bool{}
		newOneOf := Slice{}
		for _, item := range maybeOneOf {
			itemMap, ok := item.(Map)
			if !ok {
				newOneOf = append(newOneOf, item)
				continue
			}
			refValue, hasRef := itemMap["$ref"]
			if hasRef {
				refStr, ok := refValue.(string)
				if !ok {
					newOneOf = append(newOneOf, item)
					continue
				}
				if _, seen := seenRefs[refStr]; seen {
					// Duplicate found, skip it
					continue
				}
				seenRefs[refStr] = true
			}
			newOneOf = append(newOneOf, item)
		}
		node["oneOf"] = newOneOf
	}

	properties, hasProperties := node.GetMap("properties")
	if !hasProperties {
		return
	}

	properties.Iterate(removeDuplicateOneOfRefsFromNode)
}

// transformFleetPaths fixes the fleet paths.
func transformFleetPaths(schema *Schema) {
	// Agent policies
	// https://github.com/elastic/kibana/blob/main/x-pack/plugins/fleet/common/types/models/agent_policy.ts
	// https://github.com/elastic/kibana/blob/main/x-pack/plugins/fleet/common/types/rest_spec/agent_policy.ts

	agentPoliciesPath := schema.MustGetPath("/api/fleet/agent_policies")
	agentPolicyPath := schema.MustGetPath("/api/fleet/agent_policies/{agentPolicyId}")

	agentPoliciesPath.Get.CreateRef(schema, "agent_policy", "responses.200.content.application/json.schema.properties.items.items")
	agentPoliciesPath.Post.CreateRef(schema, "agent_policy", "responses.200.content.application/json.schema.properties.item")
	agentPolicyPath.Get.CreateRef(schema, "agent_policy", "responses.200.content.application/json.schema.properties.item")
	agentPolicyPath.Put.CreateRef(schema, "agent_policy", "responses.200.content.application/json.schema.properties.item")

	schema.Components.CreateRef(schema, "agent_policy_global_data_tags_item", "schemas.agent_policy.properties.global_data_tags.items")

	// Define the value types for the GlobalDataTags
	agentPoliciesPath.Post.Set("requestBody.content.application/json.schema.properties.global_data_tags.items.$ref", "#/components/schemas/agent_policy_global_data_tags_item")
	agentPolicyPath.Put.Set("requestBody.content.application/json.schema.properties.global_data_tags.items.$ref", "#/components/schemas/agent_policy_global_data_tags_item")

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
		kafkaComponent := fmt.Sprintf("%s_kafka", name)
		schema.Components.CreateRef(schema, fmt.Sprintf("%s_elasticsearch", name), fmt.Sprintf("schemas.%s_union.anyOf.0", name))
		schema.Components.CreateRef(schema, fmt.Sprintf("%s_remote_elasticsearch", name), fmt.Sprintf("schemas.%s_union.anyOf.1", name))
		schema.Components.CreateRef(schema, fmt.Sprintf("%s_logstash", name), fmt.Sprintf("schemas.%s_union.anyOf.2", name))
		schema.Components.CreateRef(schema, kafkaComponent, fmt.Sprintf("schemas.%s_union.anyOf.3", name))

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

		// https://github.com/elastic/kibana/issues/197153
		kafkaRequiredName := fmt.Sprintf("schemas.%s.required", kafkaComponent)
		props := schema.Components.MustGetMap(fmt.Sprintf("schemas.%s.properties", kafkaComponent))
		required := schema.Components.MustGetSlice(kafkaRequiredName)
		for key, apiType := range map[string]string{"compression_level": "integer", "connection_type": "string", "password": "string", "username": "string"} {
			props.Set(key, Map{
				"type": apiType,
			})
			required = slices.DeleteFunc(required, func(item any) bool {
				itemStr, ok := item.(string)
				if !ok {
					return false
				}

				return itemStr == key
			})
		}
		schema.Components.Set(kafkaRequiredName, required)
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
}

func setAllXOmitEmpty(key string, node Map) {
	maybeNullable, hasNullable := node.Get("nullable")
	isNullable, ok := maybeNullable.(bool)
	if hasNullable && ok && isNullable {
		node.Set("x-omitempty", true)
	}

	properties, hasProperties := node.GetMap("properties")
	if !hasProperties {
		return
	}

	properties.Iterate(setAllXOmitEmpty)
}

func transformOmitEmptyNullable(schema *Schema) {
	componentSchemas := schema.Components.MustGetMap("schemas")
	componentSchemas.Iterate(setAllXOmitEmpty)

	for _, pathInfo := range schema.Paths {
		for _, methInfo := range pathInfo.Endpoints {
			requestBody, ok := methInfo.GetMap("requestBody.content.application/json.schema.properties")
			if ok {
				requestBody.Iterate(setAllXOmitEmpty)
			}
		}
	}
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
