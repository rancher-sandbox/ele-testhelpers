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
