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

package registry_test

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/google/go-containerregistry/pkg/registry"
)

const (
	weirdIndex = `{
  "manifests": [
	  {
			"digest":"sha256:2c26b46b68ffc68ff99b453c1d30413413422d706483bfa0f98a5e886266e7ae",
			"mediaType":"application/vnd.oci.image.layer.nondistributable.v1.tar+gzip"
		},{
			"digest":"sha256:2c26b46b68ffc68ff99b453c1d30413413422d706483bfa0f98a5e886266e7ae",
			"mediaType":"application/xml"
		},{
			"digest":"sha256:2c26b46b68ffc68ff99b453c1d30413413422d706483bfa0f98a5e886266e7ae",
			"mediaType":"application/vnd.oci.image.manifest.v1+json"
		}
	]
}`
)

func sha256String(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}

func TestCalls(t *testing.T) {
	tcs := []struct {
		Description string

		// Request / setup
		URL           string
		Digests       map[string]string
		Manifests     map[string]string
		BlobStream    map[string]string
		RequestHeader map[string]string

		// Response
		Code   int
		Header map[string]string
		Method string
		Body   string
	}{
		{
			Description: "/v2 returns 200",
			Method:      "GET",
			URL:         "/v2",
			Code:        http.StatusOK,
			Header:      map[string]string{"Docker-Distribution-API-Version": "registry/2.0"},
		},
		{
			Description: "/v2/ returns 200",
			Method:      "GET",
			URL:         "/v2/",
			Code:        http.StatusOK,
			Header:      map[string]string{"Docker-Distribution-API-Version": "registry/2.0"},
		},
		{
			Description: "/v2/bad returns 404",
			Method:      "GET",
			URL:         "/v2/bad",
			Code:        http.StatusNotFound,
			Header:      map[string]string{"Docker-Distribution-API-Version": "registry/2.0"},
		},
		{
			Description: "GET non existent blob",
			Method:      "GET",
			URL:         "/v2/foo/blobs/sha256:asd",
			Code:        http.StatusNotFound,
		},
		{
			Description: "HEAD non existent blob",
			Method:      "HEAD",
			URL:         "/v2/foo/blobs/sha256:asd",
			Code:        http.StatusNotFound,
		},
		{
			Description: "bad blob verb",
			Method:      "FOO",
			URL:         "/v2/foo/blobs/sha256:asd",
			Code:        http.StatusBadRequest,
		},
		{
			Description: "GET containerless blob",
			Digests:     map[string]string{"sha256:2c26b46b68ffc68ff99b453c1d30413413422d706483bfa0f98a5e886266e7ae": "foo"},
			Method:      "GET",
			URL:         "/v2/foo/blobs/sha256:2c26b46b68ffc68ff99b453c1d30413413422d706483bfa0f98a5e886266e7ae",
			Code:        http.StatusOK,
			Header:      map[string]string{"Docker-Content-Digest": "sha256:2c26b46b68ffc68ff99b453c1d30413413422d706483bfa0f98a5e886266e7ae"},
		},
		{
			Description: "GET blob",
			Digests:     map[string]string{"sha256:2c26b46b68ffc68ff99b453c1d30413413422d706483bfa0f98a5e886266e7ae": "foo"},
			Method:      "GET",
			URL:         "/v2/foo/blobs/sha256:2c26b46b68ffc68ff99b453c1d30413413422d706483bfa0f98a5e886266e7ae",
			Code:        http.StatusOK,
			Header:      map[string]string{"Docker-Content-Digest": "sha256:2c26b46b68ffc68ff99b453c1d30413413422d706483bfa0f98a5e886266e7ae"},
		},
		{
			Description: "HEAD blob",
			Digests:     map[string]string{"sha256:2c26b46b68ffc68ff99b453c1d30413413422d706483bfa0f98a5e886266e7ae": "foo"},
			Method:      "HEAD",
			URL:         "/v2/foo/blobs/sha256:2c26b46b68ffc68ff99b453c1d30413413422d706483bfa0f98a5e886266e7ae",
			Code:        http.StatusOK,
			Header: map[string]string{
				"Content-Length":        "3",
				"Docker-Content-Digest": "sha256:2c26b46b68ffc68ff99b453c1d30413413422d706483bfa0f98a5e886266e7ae",
			},
		},
		{
			Description: "blob url with no container",
			Method:      "GET",
			URL:         "/v2/blobs/sha256:asd",
			Code:        http.StatusBadRequest,
		},
		{
			Description: "get missing manifest",
			Method:      "GET",
			URL:         "/v2/foo/manifests/latest",
			Code:        http.StatusNotFound,
		},
		{
			Description: "head missing manifest",
			Method:      "HEAD",
			URL:         "/v2/foo/manifests/latest",
			Code:        http.StatusNotFound,
		},
		{
			Description: "get missing manifest good container",
			Manifests:   map[string]string{"foo/manifests/latest": "foo"},
			Method:      "GET",
			URL:         "/v2/foo/manifests/bar",
			Code:        http.StatusNotFound,
		},
		{
			Description: "head missing manifest good container",
			Manifests:   map[string]string{"foo/manifests/latest": "foo"},
			Method:      "HEAD",
			URL:         "/v2/foo/manifests/bar",
			Code:        http.StatusNotFound,
		},
		{
			Description: "get manifest by tag",
			Manifests:   map[string]string{"foo/manifests/latest": "foo"},
			Method:      "GET",
			URL:         "/v2/foo/manifests/latest",
			Code:        http.StatusOK,
		},
		{
			Description: "get manifest by digest",
			Manifests:   map[string]string{"foo/manifests/latest": "foo"},
			Method:      "GET",
			URL:         "/v2/foo/manifests/sha256:" + sha256String("foo"),
			Code:        http.StatusOK,
		},
		{
			Description: "head manifest",
			Manifests:   map[string]string{"foo/manifests/latest": "foo"},
			Method:      "HEAD",
			URL:         "/v2/foo/manifests/latest",
			Code:        http.StatusOK,
		},
		{
			Description: "bad manifest method",
			Method:      "BAR",
			URL:         "/v2/foo/manifests/latest",
			Code:        http.StatusBadRequest,
		},
	}

	for _, tc := range tcs {

		var logger *log.Logger
		testf := func(t *testing.T) {

			r := registry.New()
			if logger != nil {
				r = registry.New(registry.Logger(logger))
			}
			s := httptest.NewServer(r)
			defer s.Close()

			for manifest, contents := range tc.Manifests {
				u, err := url.Parse(s.URL + "/v2/" + manifest)
				if err != nil {
					t.Fatalf("Error parsing %q: %v", s.URL+"/v2", err)
				}
				req := &http.Request{
					Method: "PUT",
					URL:    u,
					Body:   ioutil.NopCloser(strings.NewReader(contents)),
				}
				resp, err := s.Client().Do(req)
				if err != nil {
					t.Fatalf("Error uploading manifest: %v", err)
				}
				if resp.StatusCode != http.StatusCreated {
					t.Fatalf("Error uploading manifest got status: %d", resp.StatusCode)
				}
				t.Logf("created manifest with digest %v", resp.Header.Get("Docker-Content-Digest"))
			}

			for digest, contents := range tc.Digests {
				u, err := url.Parse(fmt.Sprintf("%s/v2/foo/blobs/uploads/1?digest=%s", s.URL, digest))
				if err != nil {
					t.Fatalf("Error parsing %q: %v", s.URL+tc.URL, err)
				}
				req := &http.Request{
					Method: "PUT",
					URL:    u,
					Body:   ioutil.NopCloser(strings.NewReader(contents)),
				}
				resp, err := s.Client().Do(req)
				if err != nil {
					t.Fatalf("Error uploading digest: %v", err)
				}
				if resp.StatusCode != http.StatusCreated {
					t.Fatalf("Error uploading digest got status: %d", resp.StatusCode)
				}
			}

			for upload, contents := range tc.BlobStream {
				u, err := url.Parse(fmt.Sprintf("%s/v2/foo/blobs/uploads/%s", s.URL, upload))
				if err != nil {
					t.Fatalf("Error parsing %q: %v", s.URL+tc.URL, err)
				}
				req := &http.Request{
					Method: "PATCH",
					URL:    u,
					Body:   ioutil.NopCloser(strings.NewReader(contents)),
				}
				resp, err := s.Client().Do(req)
				if err != nil {
					t.Fatalf("Error streaming blob: %v", err)
				}
				if resp.StatusCode != http.StatusNoContent {
					t.Fatalf("Error streaming blob: %d", resp.StatusCode)
				}

			}

			u, err := url.Parse(s.URL + tc.URL)
			if err != nil {
				t.Fatalf("Error parsing %q: %v", s.URL+tc.URL, err)
			}
			req := &http.Request{
				Method: tc.Method,
				URL:    u,
				Body:   ioutil.NopCloser(strings.NewReader(tc.Body)),
				Header: map[string][]string{},
			}
			for k, v := range tc.RequestHeader {
				req.Header.Set(k, v)
			}
			resp, err := s.Client().Do(req)
			if err != nil {
				t.Fatalf("Error getting %q: %v", tc.URL, err)
			}
			if resp.StatusCode != tc.Code {
				t.Errorf("Incorrect status code, got %d, want %d", resp.StatusCode, tc.Code)
			}

			for k, v := range tc.Header {
				r := resp.Header.Get(k)
				if r != v {
					t.Errorf("Incorrect header %q received, got %q, want %q", k, r, v)
				}
			}
		}
		t.Run(tc.Description, testf)
		logger = log.New(ioutil.Discard, "", log.Ldate)
		t.Run(tc.Description+" - custom log", testf)
	}
}
