/*
Copyright © 2022 - 2024 SUSE LLC

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

package http_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/rancher-sandbox/ele-testhelpers/http"
	"golang.org/x/net/context"
)

var _ = Describe("HTTP Server", func() {
	ctx, cancel := context.WithCancel(context.Background())
	Context("Server", func() {
		It("serves correctly pages and closes when context is closed", func() {
			Server(ctx, ":9099", "foobar")

			Eventually(func() string {
				str, _ := GetInsecure("http://localhost:9099")
				//	f, _ := k.GetPodNames("default", "")
				return str
			}, 10*time.Second, 1*time.Second).Should(Equal("foobar"))

			cancel()

			Eventually(func() string {
				str, _ := GetInsecure("http://localhost:9099")
				//	f, _ := k.GetPodNames("default", "")
				return str
			}, 10*time.Second, 1*time.Second).Should(Equal(""))
		})
	})
})
