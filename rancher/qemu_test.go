/*
Copyright © 2022 - 2025 SUSE LLC

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
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/rancher-sandbox/ele-testhelpers/rancher"
)

var _ = Describe("Qemu helpers tests", func() {
	It("Test qemu helpers functions", func() {
		By("Testing GetHostNetConfig function", func() {
			dataFile := "assets/host-test.xml"
			vmName := "node02"
			regex := ".*name='" + vmName + "'.*"

			out, err := rancher.GetHostNetConfig(regex, dataFile)
			Expect(err).To(Not(HaveOccurred()))
			Expect(out.Name).To(Equal(vmName))
			Expect(out.Mac).To(Equal("52:54:00:00:00:02"))
			Expect(out.IP).To(Equal("192.168.122.12"))
		})
	})
})
