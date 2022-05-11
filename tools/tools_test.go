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
			Expect(err).NotTo(HaveOccurred())
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

		By("Checking RunSSH function", func() {
			userName := "testuser"
			userPassword := "testpassword"
			server := "localhost"
			testCmd := "uname -a"

			// Check connection without sshd started
			_, err := tools.RunSSH(userName, userPassword, server, testCmd)
			Expect(err).To(HaveOccurred())

			// Start sshd
			err = exec.Command("sudo", "mkdir", "-p", "/run/sshd").Run()
			Expect(err).NotTo(HaveOccurred())
			err = exec.Command("sudo", "ssh-keygen", "-A").Run()
			Expect(err).NotTo(HaveOccurred())
			err = exec.Command("sudo", "/usr/sbin/sshd").Run()
			Expect(err).NotTo(HaveOccurred())

			// Check connection without 'testuser' configured
			_, err = tools.RunSSH(userName, userPassword, server, testCmd)
			Expect(err).To(HaveOccurred())

			// Add 'testuser'
			err = exec.Command("sudo", "useradd", userName).Run()
			Expect(err).NotTo(HaveOccurred())

			// Use 'sed' instead of 'tools.Sed' because root access
			// is needed and it's easier with 'sudo'
			userShadow := userName + ":$6$X7HdGuscUQ.XW6dW$B8rTHpY2bZJKyPebFn20fuj0oiLj3D9L557tTBbZ2ZycuIT23UOnjxwgQEki3//wK0/RKmXeOYPHbYhregyfu1:19122:0:99999:7:::"
			regex := "s|^" + userName + ":.*|" + userShadow + "|"
			err = exec.Command("sudo", "sed", "-i", regex, "/etc/shadow").Run()
			Expect(err).NotTo(HaveOccurred())

			// Check a working connection
			_, err = tools.RunSSH(userName, userPassword, server, testCmd)
			Expect(err).NotTo(HaveOccurred())

			// Check a unknown command
			_, err = tools.RunSSH(userName, userPassword, server, "foo")
			Expect(err).To(HaveOccurred())
		})
	})
})
