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

// Package health provides a health check service.
package health

import (
	"fmt"
	"net/http"

	"k8s.io/client-go/kubernetes"
)

// HealthzMetrics records request statistics for the /healthz endpoint.
type HealthzMetrics interface {
	IncHealthzRequests()
	IncHealthzSuccess()
	IncHealthzFailure()
}

// Handler returns a http Handler for a health endpoint. Requests to /healthz
// are recorded in m; requests to /health are not.
func Handler(client kubernetes.Interface, m HealthzMetrics) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		record := r.URL.Path == "/healthz"
		if record {
			m.IncHealthzRequests()
		}

		// Try and lookup Kubernetes server version as a quick and dirty check
		_, err := client.Discovery().ServerVersion()
		if err != nil {
			if record {
				m.IncHealthzFailure()
			}
			msg := fmt.Sprintf("Failed Kubernetes Check: %v", err)
			http.Error(w, msg, http.StatusServiceUnavailable)
			return
		}
		if record {
			m.IncHealthzSuccess()
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "OK")
	})
}
