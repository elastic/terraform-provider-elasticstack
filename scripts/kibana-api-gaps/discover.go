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

import (
	"cmp"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"slices"
	"sort"
	"strings"
)

// Group is the set of unused request fields beneath one top-level request field of a resource.
type Group struct {
	Key         string   `json:"key"`
	Resource    string   `json:"resource"`
	Field       string   `json:"field"`
	Pointer     string   `json:"pointer"`
	Fingerprint string   `json:"fingerprint"`
	Members     []Member `json:"members"`
	// query is true for a query parameter group and false for a request body property group.
	query bool
}

// Discovery is the deterministic result of analysing every in-scope resource.
type Discovery struct {
	Groups        []Group        `json:"groups"`
	SkippedShapes []SkippedShape `json:"skipped_shapes"`
	// Operations holds, per analysed resource, the number of API operations it can reach.
	Operations map[string]int `json:"operations"`
	// NoOperations lists registered resources that reach no Kibana API operation.
	NoOperations []string `json:"no_operations"`
}

type discoverConfig struct {
	module    *loadedModule
	kbapiPath string
	oas       *oasDoc
	resources []Resource
}

func discover(cfg discoverConfig) (*Discovery, error) {
	opIdx, err := buildOpIndex(cfg.module.byPath[cfg.kbapiPath])
	if err != nil {
		return nil, err
	}
	out := &Discovery{
		Groups:        []Group{},
		SkippedShapes: []SkippedShape{},
		Operations:    make(map[string]int),
		NoOperations:  []string{},
	}
	defined := buildDefinedTypes(cfg.module.byPath[cfg.kbapiPath])
	groups := make(map[string]*groupBuild)
	for _, res := range cfg.resources {
		reach := computeReach(cfg.module, cfg.kbapiPath, res.PkgPath)
		ops := reachedOperations(reach, opIdx)
		if len(ops) == 0 {
			out.NoOperations = append(out.NoOperations, res.Name)
			continue
		}
		out.Operations[res.Name] = len(ops)
		w := newWalker(cfg.oas, collectUsage(reach), defined, res.Name)
		for _, op := range ops {
			props, err := operationGaps(cfg.oas, w, op)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", res.Name, err)
			}
			for _, prop := range props {
				addGroup(groups, res.Name, prop)
			}
		}
		out.SkippedShapes = append(out.SkippedShapes, w.skipped...)
	}
	for _, build := range groups {
		group, err := build.finish(cfg.oas)
		if err != nil {
			return nil, err
		}
		out.Groups = append(out.Groups, group)
	}
	out.Groups = mergeGroups(out.Groups)
	for _, g := range out.Groups {
		if err := checkMarkerSafe(g); err != nil {
			return nil, err
		}
	}
	sort.Strings(out.NoOperations)
	return out, nil
}

func reachedOperations(reach *reachResult, idx *opIndex) []opSpec {
	byKey := make(map[string]opSpec)
	for name := range reach.Methods {
		spec, ok := idx.method(name)
		if !ok {
			continue
		}
		key := operationKey(spec.Method, spec.Path)
		if existing, seen := byKey[key]; seen && !existing.RawBody {
			continue
		}
		byKey[key] = spec
	}
	keys := make([]string, 0, len(byKey))
	for key := range byKey {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	ops := make([]opSpec, 0, len(keys))
	for _, key := range keys {
		ops = append(ops, byKey[key])
	}
	return ops
}

func operationGaps(doc *oasDoc, w *walker, spec opSpec) ([]propGap, error) {
	op, ok := doc.operation(spec.Method, spec.Path)
	if !ok {
		return nil, fmt.Errorf("operation %s %s of the generated client is missing from the OpenAPI document", spec.Method, spec.Path)
	}
	name := spec.Method + " " + spec.Path
	props, err := w.queryGaps(op.Query, spec.ParamsType, name)
	if err != nil {
		return nil, err
	}
	if op.Body == nil {
		return props, nil
	}
	if spec.BodyType == nil {
		shape := "request body without a JSON request body type in the client"
		if spec.RawBody {
			shape = "raw request body: the client method takes an opaque io.Reader and no typed body is known"
		}
		w.skip(op.Body.pointer, shape)
		return props, nil
	}
	bodyProps, err := w.bodyGaps(*op.Body, spec.BodyType)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", name, err)
	}
	return append(props, bodyProps...), nil
}

// groupBuild accumulates a group across operations; required is OR-ed so the result does not
// depend on operation order.
type groupBuild struct {
	group    Group
	ref      schemaRef
	required bool
}

func addGroup(groups map[string]*groupBuild, resource string, prop propGap) {
	if len(prop.members) == 0 {
		return
	}
	key := resource + "#" + prop.ref.pointer
	build, ok := groups[key]
	if !ok {
		build = &groupBuild{
			group: Group{Key: key, Resource: resource, Field: prop.name, Pointer: prop.ref.pointer, query: prop.query},
			ref:   prop.ref,
		}
		groups[key] = build
	}
	build.required = build.required || prop.required
	for _, member := range prop.members {
		if !slices.Contains(build.group.Members, member) {
			build.group.Members = append(build.group.Members, member)
		}
	}
}

func (b *groupBuild) finish(doc *oasDoc) (Group, error) {
	structure, err := doc.fingerprint(b.ref)
	if err != nil {
		return Group{}, err
	}
	sum := sha256.Sum256([]byte(structure + "|" + strings.ToLower(fmt.Sprint(b.required))))
	group := b.group
	group.Fingerprint = hex.EncodeToString(sum[:])
	slices.SortFunc(group.Members, func(a, b Member) int {
		return cmp.Or(strings.Compare(a.Field, b.Field), strings.Compare(a.Pointer, b.Pointer), strings.Compare(a.XState, b.XState))
	})
	return group, nil
}

// mergeGroups collapses the groups of one resource that share a top-level field name and kind (request-body property or query parameter), such as the
// same field in a create and an update body or in every variant of a union body, into a single
// group. The merged group is keyed by the smallest member location so its identity does not
// depend on input order; its fingerprint covers every constituent, so a structural change to
// any of them, or a new constituent, expires the verdict. A group with no siblings is returned
// unchanged.
func mergeGroups(groups []Group) []Group {
	type id struct {
		resource, field string
		query           bool
	}
	byID := make(map[id][]Group)
	for _, g := range groups {
		key := id{g.Resource, g.Field, g.query}
		byID[key] = append(byID[key], g)
	}
	merged := make([]Group, 0, len(byID))
	for _, siblings := range byID {
		merged = append(merged, mergeSiblings(siblings))
	}
	sort.Slice(merged, func(i, j int) bool { return merged[i].Key < merged[j].Key })
	return merged
}

func mergeSiblings(siblings []Group) Group {
	if len(siblings) == 1 {
		return siblings[0]
	}
	slices.SortFunc(siblings, func(a, b Group) int { return strings.Compare(a.Pointer, b.Pointer) })
	out := siblings[0]
	out.Members = nil
	var structure strings.Builder
	for _, g := range siblings {
		structure.WriteString(g.Pointer + "=" + g.Fingerprint + "\n")
		for _, m := range g.Members {
			if !slices.Contains(out.Members, m) {
				out.Members = append(out.Members, m)
			}
		}
	}
	slices.SortFunc(out.Members, func(a, b Member) int {
		return cmp.Or(strings.Compare(a.Field, b.Field), strings.Compare(a.Pointer, b.Pointer), strings.Compare(a.XState, b.XState))
	})
	sum := sha256.Sum256([]byte(structure.String()))
	out.Fingerprint = hex.EncodeToString(sum[:])
	return out
}
