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

// Package fileutil provides shared helpers for computing and comparing
// file checksums, used by resources that detect drift in a file referenced
// by path (e.g. package archives, investigation guide content).
package fileutil

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
)

// SHA256HexDigest returns the hex-encoded SHA-256 digest of the file at path.
// It streams the file contents rather than loading the whole file into
// memory, so it is safe to use on large files.
func SHA256HexDigest(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// FileChecksumDrifted computes the current SHA-256 digest of the file at path
// and reports whether it differs from priorChecksum. hasPriorState must be
// false when the resource is being created (there is no prior state to
// compare against); in that case changed is true when the digest is computed
// successfully. Otherwise, changed reports whether newChecksum != priorChecksum.
func FileChecksumDrifted(path string, priorChecksum string, hasPriorState bool) (newChecksum string, changed bool, err error) {
	newChecksum, err = SHA256HexDigest(path)
	if err != nil {
		return "", false, err
	}
	return newChecksum, !hasPriorState || newChecksum != priorChecksum, nil
}
