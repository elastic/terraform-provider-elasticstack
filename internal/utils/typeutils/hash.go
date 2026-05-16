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

package typeutils

import (
	"crypto/sha1"
	"fmt"
)

// StringToHash computes a SHA-1 hash of s and returns the hex-encoded result as a pointer.
func StringToHash(s string) (*string, error) {
	h := sha1.New()
	_, err := h.Write([]byte(s))
	if err != nil {
		return nil, err
	}
	bs := h.Sum(nil)
	hash := fmt.Sprintf("%x", bs)
	return &hash, nil
}
