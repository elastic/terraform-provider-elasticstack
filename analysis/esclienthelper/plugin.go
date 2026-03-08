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

package esclienthelper

import (
	"github.com/golangci/plugin-module-register/register"
	"golang.org/x/tools/go/analysis"
)

func init() {
	register.Plugin("esclienthelper", New)
}

type plugin struct {
	cfg Config
}

func New(settings any) (register.LinterPlugin, error) {
	cfg, err := register.DecodeSettings[Config](settings)
	if err != nil {
		return nil, err
	}

	return &plugin{cfg: cfg}, nil
}

func (p *plugin) BuildAnalyzers() ([]*analysis.Analyzer, error) {
	return []*analysis.Analyzer{NewAnalyzer(p.cfg)}, nil
}

func (p *plugin) GetLoadMode() string {
	return register.LoadModeTypesInfo
}
