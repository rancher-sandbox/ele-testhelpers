package tools_test

import (
	"os/exec"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/rancher-sandbox/ele-testhelpers/tools"
)

var _ = Describe("Tools tests", func() {
	It("Test tools helper functions", func() {
		By("Checking GetFileFromURL function", func() {
			fileName := "check-file"
			err := tools.GetFileFromURL("https://raw.githubusercontent.com/rancher-sandbox/ele-testhelpers/main/README.md", fileName, false)
			Expect(err).NotTo(HaveOccurred())
			out, err := exec.Command("file", fileName).CombinedOutput()
			Expect(err).NotTo(HaveOccurred())
			Expect(string(out)).To(Equal(fileName + ": ASCII text\n"))

			// Check error handling
			err = tools.GetFileFromURL("http://web-site-does-not-exist.foo", fileName, true)
			Expect(err).To(HaveOccurred())
		})

		By("Checking GetFiles function", func() {
			fileName := "README.md"
			file, err := tools.GetFiles("..", fileName)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(file)).To(BeNumerically("==", 1))
			out, err := exec.Command("file", file[0]).CombinedOutput()
			Expect(err).NotTo(HaveOccurred())
			Expect(string(out)).To(Equal("../" + fileName + ": ASCII text\n"))

			// Check error handling
			file, err = tools.GetFiles("..", "foo")
			Expect(err).Should(HaveOccurred())
			Expect(file).To(BeNil())
		})

		By("Checking Sed function", func() {
			fileName := "../README.md"
			value := "TEST"
			err := tools.Sed("#.*", value, fileName)
			Expect(err).NotTo(HaveOccurred())
			out, err := exec.Command("sed", "-n", "1p", fileName).CombinedOutput()
			Expect(err).NotTo(HaveOccurred())
			Expect(string(out)).To(Equal(value + "\n"))

			// Check error handling
			err = tools.Sed("#.*", value, "foo")
			Expect(err).To(HaveOccurred())
		})
	})
})
