// Copyright Project Contour Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package health

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/version"
	fakediscovery "k8s.io/client-go/discovery/fake"
	"k8s.io/client-go/kubernetes/fake"
	k8stesting "k8s.io/client-go/testing"
)

type fakeHealthzMetrics struct {
	requests  int
	successes int
	failures  int
}

func (f *fakeHealthzMetrics) IncHealthzRequests() { f.requests++ }
func (f *fakeHealthzMetrics) IncHealthzSuccess()  { f.successes++ }
func (f *fakeHealthzMetrics) IncHealthzFailure()  { f.failures++ }

func TestHandler(t *testing.T) {
	tests := map[string]struct {
		path          string
		configure     func(*fake.Clientset)
		wantStatus    int
		wantBody      string
		wantRequests  int
		wantSuccesses int
		wantFailures  int
	}{
		"healthz success": {
			path: "/healthz",
			configure: func(c *fake.Clientset) {
				c.Discovery().(*fakediscovery.FakeDiscovery).FakedServerVersion = &version.Info{Major: "1", Minor: "30"}
			},
			wantStatus:    http.StatusOK,
			wantBody:      "OK\n",
			wantRequests:  1,
			wantSuccesses: 1,
		},
		"healthz failure": {
			path: "/healthz",
			configure: func(c *fake.Clientset) {
				c.PrependReactor("get", "version", func(k8stesting.Action) (bool, runtime.Object, error) {
					return true, nil, errors.New("kube api unavailable")
				})
			},
			wantStatus:   http.StatusServiceUnavailable,
			wantBody:     "Failed Kubernetes Check: kube api unavailable\n",
			wantRequests: 1,
			wantFailures: 1,
		},
		"health success is not counted": {
			path: "/health",
			configure: func(c *fake.Clientset) {
				c.Discovery().(*fakediscovery.FakeDiscovery).FakedServerVersion = &version.Info{Major: "1", Minor: "30"}
			},
			wantStatus: http.StatusOK,
			wantBody:   "OK\n",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			client := fake.NewSimpleClientset()
			tc.configure(client)

			metrics := &fakeHealthzMetrics{}
			recorder := httptest.NewRecorder()
			Handler(client, metrics).ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, tc.path, nil))

			assert.Equal(t, tc.wantStatus, recorder.Code)
			assert.Equal(t, tc.wantBody, recorder.Body.String())
			assert.Equal(t, tc.wantRequests, metrics.requests)
			assert.Equal(t, tc.wantSuccesses, metrics.successes)
			assert.Equal(t, tc.wantFailures, metrics.failures)
		})
	}
}
