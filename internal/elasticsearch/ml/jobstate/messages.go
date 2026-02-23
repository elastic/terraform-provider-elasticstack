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

package jobstate

const (
	createTimeoutErrorMessage = "The operation to create the ML job state timed out after %s. " +
		"You may need to allocate more free memory within ML nodes by either closing other jobs, " +
		"or increasing the overall ML memory. You may retry the operation."

	updateTimeoutErrorMessage = "The operation to update the ML job state timed out after %s. " +
		"You may need to allocate more free memory within ML nodes by either closing other jobs, " +
		"or increasing the overall ML memory. You may retry the operation."
)
