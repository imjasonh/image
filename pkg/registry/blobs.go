// Copyright 2018 Google LLC All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package registry

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"

	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/layout"
)

// Returns whether this url should be handled by the blob handler
// This is complicated because blob is indicated by the trailing path, not the leading path.
// https://github.com/opencontainers/distribution-spec/blob/master/spec.md#pulling-a-layer
// https://github.com/opencontainers/distribution-spec/blob/master/spec.md#pushing-a-layer
func isBlob(req *http.Request) bool {
	elem := strings.Split(req.URL.Path, "/")
	elem = elem[1:]
	if elem[len(elem)-1] == "" {
		elem = elem[:len(elem)-1]
	}
	if len(elem) < 3 {
		return false
	}
	return elem[len(elem)-2] == "blobs"
}

// blobs
type blobs struct {
	path layout.Path

	lock sync.Mutex
}

func (b *blobs) handle(resp http.ResponseWriter, req *http.Request) *regError {
	elem := strings.Split(req.URL.Path, "/")
	elem = elem[1:]
	if elem[len(elem)-1] == "" {
		elem = elem[:len(elem)-1]
	}
	// Must have a path of form /v2/{name}/blobs/{upload,sha256:}
	if len(elem) < 4 {
		return &regError{
			Status:  http.StatusBadRequest,
			Code:    "NAME_INVALID",
			Message: "blobs must be attached to a repo",
		}
	}
	target := elem[len(elem)-1]
	h, err := v1.NewHash(target)
	if err != nil {
		return &regError{
			Status:  http.StatusNotFound,
			Code:    "BLOB_UNKNOWN",
			Message: "Unknown blob",
		}
	}

	if req.Method == "HEAD" {
		b.lock.Lock()
		defer b.lock.Unlock()

		rc, err := b.path.Blob(h)
		if err != nil {
			return &regError{
				Status:  http.StatusNotFound,
				Code:    "BLOB_UNKNOWN",
				Message: "Unknown blob",
			}
		}
		defer rc.Close()
		size, err := io.Copy(ioutil.Discard, rc)
		if err != nil {
			return &regError{
				Status:  http.StatusNotFound,
				Code:    "BLOB_UNKNOWN",
				Message: "Unknown blob",
			}
		}

		resp.Header().Set("Content-Length", fmt.Sprint(size))
		resp.Header().Set("Docker-Content-Digest", target)
		resp.WriteHeader(http.StatusOK)
		return nil
	}

	if req.Method == "GET" {
		b.lock.Lock()
		defer b.lock.Unlock()

		b, err := b.path.Bytes(h)
		if err != nil {
			return &regError{
				Status:  http.StatusNotFound,
				Code:    "BLOB_UNKNOWN",
				Message: "Unknown blob",
			}
		}

		resp.Header().Set("Content-Length", fmt.Sprint(len(b)))
		resp.Header().Set("Docker-Content-Digest", target)
		resp.WriteHeader(http.StatusOK)
		io.Copy(resp, bytes.NewReader(b))
		return nil
	}

	return &regError{
		Status:  http.StatusBadRequest,
		Code:    "METHOD_UNKNOWN",
		Message: "We don't understand your method + url",
	}
}
