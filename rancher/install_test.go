/*
Copyright © 2022 - 2026 SUSE LLC

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package rancher_test

import (
	"net/http"
	"net/http/httptest"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/rancher-sandbox/ele-testhelpers/rancher"
)

const fakeIndex = `apiVersion: v1
entries:
  rancher:
  - name: rancher
    version: 2.15.3-aaaaaaa-head
    created: "2026-09-24T09:59:15Z"
  - name: rancher
    version: 2.15.3-bbbbbbb-head
    created: "2026-10-02T09:52:41Z"
  - name: rancher
    version: 2.15.2-ccccccc-head
    created: "2026-09-30T10:00:00Z"
  - name: rancher
    version: 2.1.0
    created: "2026-10-04T10:00:00Z"
  - name: rancher
    version: 2.16.0-ddddddd-head
    created: "2026-10-03T10:00:00Z"
`

var _ = Describe("Install helpers tests", func() {
	It("Test fetchPrimeHeadVersion function", func() {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/index.yaml" {
				http.NotFound(w, r)
				return
			}
			_, _ = w.Write([]byte(fakeIndex))
		}))
		defer server.Close()

		By("Returning the most recently created matching version", func() {
			version, err := rancher.FetchPrimeHeadVersion("2.15", server.URL)
			Expect(err).To(Not(HaveOccurred()))
			Expect(version).To(Equal("2.15.3-bbbbbbb-head"))
		})

		By("Not matching a longer minor version", func() {
			version, err := rancher.FetchPrimeHeadVersion("2.1", server.URL)
			Expect(err).To(Not(HaveOccurred()))
			Expect(version).To(Equal("2.1.0"))
		})

		By("Returning an error when no version matches", func() {
			_, err := rancher.FetchPrimeHeadVersion("2.99", server.URL)
			Expect(err).To(HaveOccurred())
		})

		By("Returning an error when index.yaml cannot be fetched", func() {
			_, err := rancher.FetchPrimeHeadVersion("2.15", server.URL+"/missing")
			Expect(err).To(HaveOccurred())
		})
	})
})
