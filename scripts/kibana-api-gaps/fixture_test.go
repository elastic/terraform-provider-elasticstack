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
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/tools/go/packages"
)

const fixtureKbapi = "fixture/kbapi"

var fixtureCache sync.Map

func loadFixture(t *testing.T, dir string) *loadedModule {
	t.Helper()
	if cached, ok := fixtureCache.Load(dir); ok {
		return cached.(*loadedModule)
	}
	mod, err := loadModule(dir, "./...")
	require.NoError(t, err)
	fixtureCache.Store(dir, mod)
	return mod
}

func (m *loadedModule) pkg(t *testing.T, path string) *packages.Package {
	t.Helper()
	pkg, ok := m.byPath[path]
	require.True(t, ok, path)
	return pkg
}
