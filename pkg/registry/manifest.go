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
	"log"
	"net/http"
	"strings"
	"sync"

	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/layout"
)

type catalog struct {
	Repos []string `json:"repositories"`
}

type listTags struct {
	Name string   `json:"name"`
	Tags []string `json:"tags"`
}

type manifest struct {
	contentType string
	blob        []byte
}

type manifests struct {
	path layout.Path

	lock sync.Mutex
	log  *log.Logger
}

func isManifest(req *http.Request) bool {
	elems := strings.Split(req.URL.Path, "/")
	elems = elems[1:]
	if len(elems) < 4 {
		return false
	}
	return elems[len(elems)-2] == "manifests"
}

// https://github.com/opencontainers/distribution-spec/blob/master/spec.md#pulling-an-image-manifest
// https://github.com/opencontainers/distribution-spec/blob/master/spec.md#pushing-an-image
func (m *manifests) handle(resp http.ResponseWriter, req *http.Request) *regError {
	elem := strings.Split(req.URL.Path, "/")
	elem = elem[1:]
	target := elem[len(elem)-1]
	h, err := v1.NewHash(target)
	if err != nil {
		return &regError{
			Status:  http.StatusNotFound,
			Code:    "BLOB_UNKNOWN",
			Message: "Unknown blob",
		}
	}

	if req.Method == "GET" {
		m.lock.Lock()
		defer m.lock.Unlock()

		img, err := m.path.Image(h)
		if err != nil {
			return &regError{
				Status:  http.StatusNotFound,
				Code:    "NAME_UNKNOWN",
				Message: "Unknown name",
			}
		}
		mt, _ := img.MediaType()
		size, _ := img.Size()
		d, _ := img.Digest()
		resp.Header().Set("Docker-Content-Digest", d.String())
		resp.Header().Set("Content-Type", string(mt))
		resp.Header().Set("Content-Length", fmt.Sprintf("%d", size))
		rm, _ := img.RawManifest()
		io.Copy(resp, bytes.NewReader(rm))
		return nil
	}

	if req.Method == "HEAD" {
		m.lock.Lock()
		defer m.lock.Unlock()

		img, err := m.path.Image(h)
		if err != nil {
			return &regError{
				Status:  http.StatusNotFound,
				Code:    "NAME_UNKNOWN",
				Message: "Unknown name",
			}
		}
		mt, _ := img.MediaType()
		size, _ := img.Size()
		d, _ := img.Digest()
		resp.Header().Set("Docker-Content-Digest", d.String())
		resp.Header().Set("Content-Type", string(mt))
		resp.Header().Set("Content-Length", fmt.Sprintf("%d", size))
		return nil
	}

	return &regError{
		Status:  http.StatusBadRequest,
		Code:    "METHOD_UNKNOWN",
		Message: "We don't understand your method + url",
	}
}
