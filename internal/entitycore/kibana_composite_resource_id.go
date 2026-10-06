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

package entitycore

// KibanaCompositeResourceID marks models whose GetResourceID() field may encode
// a composite "<space>/<resource>" lookup key (for example Agent Builder
// skill_id). Only those models opt into parsing GetResourceID() as composite;
// plain identifiers that happen to contain "/" must not be reinterpreted.
type KibanaCompositeResourceID interface {
	UsesCompositeResourceID() bool
}

func usesCompositeResourceID(model kibanaIdentityModel) bool {
	c, ok := any(model).(KibanaCompositeResourceID)
	return ok && c.UsesCompositeResourceID()
}
