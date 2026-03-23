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

package securityuser

import _ "embed"

//go:embed resource-description.md
var userResourceDescription string

//go:embed descriptions/username.md
var usernameDescription string

//go:embed descriptions/password_hash.md
var passwordHashDescription string

//go:embed descriptions/password_wo.md
var passwordWriteOnlyDescription string

//go:embed descriptions/password_wo_version.md
var passwordWriteOnlyVersionDescription string
