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

package tools_test

import (
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/rancher-sandbox/ele-testhelpers/tools"
)

func createUser() {
	userName := "testuser"
	err := exec.Command("sudo", "useradd", userName).Run()
	Expect(err).To(Not(HaveOccurred()))

	// Use 'sed' instead of 'tools.Sed' because root access
	// is needed and it's easier with 'sudo'
	userShadow := userName + ":$6$X7HdGuscUQ.XW6dW$B8rTHpY2bZJKyPebFn20fuj0oiLj3D9L557tTBbZ2ZycuIT23UOnjxwgQEki3//wK0/RKmXeOYPHbYhregyfu1:19122:0:99999:7:::"
	regex := "s|^" + userName + ":.*|" + userShadow + "|"
	err = exec.Command("sudo", "sed", "-i", regex, "/etc/shadow").Run()
	Expect(err).To(Not(HaveOccurred()))
}

var _ = Describe("Tools tests", func() {
	It("Test tools helpers functions", func() {
		By("Testing GetFileFromURL function", func() {
			fileName := "check-file"

			err := tools.GetFileFromURL("https://raw.githubusercontent.com/rancher-sandbox/ele-testhelpers/main/README.md", fileName, false)
			Expect(err).To(Not(HaveOccurred()))

			out, err := exec.Command("file", fileName).CombinedOutput()
			Expect(err).To(Not(HaveOccurred()))
			Expect(string(out)).To(Equal(fileName + ": ASCII text\n"))

			// Check error handling
			err = tools.GetFileFromURL("http://web-site-does-not-exist.foo", fileName, true)
			Expect(err).To(HaveOccurred())
		})

		By("Testing GetFilesList function", func() {
			fileName := "README.md"

			file, err := tools.GetFilesList("..", fileName)
			Expect(err).To(Not(HaveOccurred()))
			Expect(len(file)).To(BeNumerically("==", 1))

			out, err := exec.Command("file", file[0]).CombinedOutput()
			Expect(err).To(Not(HaveOccurred()))
			Expect(string(out)).To(Equal("../" + fileName + ": ASCII text\n"))

			// Check error handling
			file, err = tools.GetFilesList("..", "foo")
			Expect(err).To(Not(HaveOccurred()))
			Expect(file).To(BeNil())
		})

		By("Testing Sed function", func() {
			fileName := "../README.md"
			value := "TEST"

			err := tools.Sed("#.*", value, fileName)
			Expect(err).To(Not(HaveOccurred()))

			out, err := exec.Command("sed", "-n", "1p", fileName).CombinedOutput()
			Expect(err).To(Not(HaveOccurred()))
			Expect(string(out)).To(Equal(value + "\n"))

			// Check error handling
			err = tools.Sed("#.*", value, "foo")
			Expect(err).To(HaveOccurred())
		})

		By("Testing RunSSH function", func() {
			userName := "testuser"
			testCmd := "uname -a"
			client := &tools.Client{
				Host:     "localhost:22",
				Username: userName,
				Password: "testpassword",
			}

			// Check connection without sshd started
			_, err := client.RunSSH(testCmd)
			Expect(err).To(HaveOccurred())

			// Start sshd
			err = exec.Command("sudo", "mkdir", "-p", "/run/sshd").Run()
			Expect(err).To(Not(HaveOccurred()))
			err = exec.Command("sudo", "ssh-keygen", "-A").Run()
			Expect(err).To(Not(HaveOccurred()))
			err = exec.Command("sudo", "/usr/sbin/sshd").Run()
			Expect(err).To(Not(HaveOccurred()))

			// Check connection without 'testuser' configured
			_, err = client.RunSSH(testCmd)
			Expect(err).To(HaveOccurred())

			// Add 'testuser'
			createUser()

			// Check a working connection
			_, err = client.RunSSH(testCmd)
			Expect(err).To(Not(HaveOccurred()))

			// Check a unknown command
			_, err = client.RunSSH("foo")
			Expect(err).To(HaveOccurred())
		})

		By("Testing SendFile function", func() {
			userName := "testuser"
			client := &tools.Client{
				Host:     "localhost:22",
				Username: userName,
				Password: "testpassword",
			}

			// Check a working copy
			err := client.SendFile("../README.md", "/tmp/file.tst", "0644")
			Expect(err).To(Not(HaveOccurred()))

			// Check a non-working copy (bad src)
			err = client.SendFile("README.md", "/tmp/badfile.tst", "0644")
			Expect(err).To(HaveOccurred())

			// Check a non-working copy (bad dst)
			err = client.SendFile("../README.md", "/badtmp/badfile.tst", "0644")
			Expect(err).To(HaveOccurred())

			// Remove 'testuser'
			err = exec.Command("sudo", "userdel", "-f", "-r", userName).Run()
			Expect(err).To(Not(HaveOccurred()))

			// Check a non-working copy (non-existent user)
			err = client.SendFile("../README.md", "/tmp/badfile.tst", "0644")
			Expect(err).To(HaveOccurred())
		})

		By("Testing GetFile function", func() {
			userName := "testuser"
			client := &tools.Client{
				Host:     "localhost:22",
				Username: userName,
				Password: "testpassword",
			}
			// Add 'testuser' back
			createUser()

			// Check a working copy
			err := client.GetFile("../README.md", "/tmp/file.tst", 0644)
			Expect(err).To(Not(HaveOccurred()))

			// Check a non-working copy (bad dst)
			err = client.GetFile("../README.md", "/badtmp/badfile.tst", 0644)
			Expect(err).To(HaveOccurred())

			// Remove 'testuser'
			err = exec.Command("sudo", "userdel", "-f", "-r", userName).Run()
			Expect(err).To(Not(HaveOccurred()))

			// Check a non-working copy (non-existent user)
			err = client.GetFile("../README.md", "/tmp/badfile.tst", 0644)
			Expect(err).To(HaveOccurred())
		})

		By("Testing HTTPShare function", func() {
			fileName := "/tmp/README.test"
			port := ":8000"

			// Start HTTP server
			tools.HTTPShare("../", port)

			// Check that we can download README.md
			err := tools.GetFileFromURL("http://localhost"+port+"/README.md", fileName, true)
			Expect(err).To(Not(HaveOccurred()))

			out, err := exec.Command("file", fileName).CombinedOutput()
			Expect(err).To(Not(HaveOccurred()))
			Expect(string(out)).To(Equal(fileName + ": ASCII text\n"))
		})

		By("Testing SetTimeout function", func() {
			value := 5
			timeoutScale := 7

			// TIMEOUT_SCALE is not defined, so 'timeout' should equal 'value'
			timeout := tools.SetTimeout(time.Duration(value))
			Expect(timeout).To(Equal(time.Duration(value)))

			// Defined TIMEOUT_SCALE
			os.Setenv("TIMEOUT_SCALE", strconv.Itoa(timeoutScale))

			// TIMEOUT_SCALE is defined, so 'timeout' should be increased
			timeout = tools.SetTimeout(time.Duration(value))
			Expect(timeout).To(Equal(time.Duration(value * timeoutScale)))
		})

		By("Testing AddDataToFile function", func() {
			srcFile := "/tmp/srcFile"
			dstFile := "/tmp/dstFile"
			srcValue := "My test content"
			addValue := "My added content"

			// ** Test with unexisting srcFile **
			err := tools.AddDataToFile(srcFile, dstFile, []byte(nil))
			Expect(err).To(HaveOccurred())

			// ** Test with added value **

			// Create srcFile
			err = exec.Command("bash", "-c", "echo -n '"+srcValue+"' > "+srcFile).Run()
			Expect(err).To(Not(HaveOccurred()))

			// Add content to file
			err = tools.AddDataToFile(srcFile, dstFile, []byte(addValue))
			Expect(err).To(Not(HaveOccurred()))

			// Check content
			out, err := exec.Command("cat", dstFile).CombinedOutput()
			Expect(err).To(Not(HaveOccurred()))
			Expect(out).To(Equal([]byte(srcValue + addValue)))

			// ** Test without added value **

			// Add content to file
			err = tools.AddDataToFile(srcFile, dstFile, []byte(nil))
			Expect(err).To(Not(HaveOccurred()))

			// Check content
			out, err = exec.Command("cat", dstFile).CombinedOutput()
			Expect(err).To(Not(HaveOccurred()))
			Expect(out).To(Equal([]byte(srcValue)))
		})

		By("Testing WriteFile function", func() {
			dstFile := "/tmp/dstFile"
			srcValue := "My test content"

			// ** Test with added value **

			// Add content to file
			err := tools.WriteFile(dstFile, []byte(srcValue))
			Expect(err).To(Not(HaveOccurred()))

			// Check content
			out, err := exec.Command("cat", dstFile).CombinedOutput()
			Expect(err).To(Not(HaveOccurred()))
			Expect(out).To(Equal([]byte(srcValue)))

			// ** Test without added value **

			// Add content to file
			err = tools.WriteFile(dstFile, []byte(nil))
			Expect(err).To(Not(HaveOccurred()))

			// Check content
			out, err = exec.Command("cat", dstFile).CombinedOutput()
			Expect(err).To(Not(HaveOccurred()))
			Expect(out).To(BeEmpty())
		})

		By("Testing TrimStringFromChar function", func() {
			stringToCheck := "myValueToCheck"
			separator := ":"
			stringToTrim := stringToCheck + separator + "myValueToRemove"

			out := tools.TrimStringFromChar(stringToTrim, separator)
			Expect(out).To(Not(BeEmpty()))
			Expect(out).To(Equal(stringToCheck))
		})

		By("Testing CreateTemp function", func() {
			// Create tmp file
			tmpFile, err := tools.CreateTemp("testFile")
			Expect(err).To(Not(HaveOccurred()))
			Expect(tmpFile).To(Not(BeEmpty()))

			// Check that tmp file exist
			err = exec.Command("test", "-f", tmpFile).Run()
			Expect(err).To(Not(HaveOccurred()))
		})

		By("Testing IsIPv4 function", func() {
			goodIPs := "192.168.122.2 10.68.250.156 1.2.3.4 255.255.255.255"
			for _, value := range strings.Fields(goodIPs) {
				Expect(tools.IsIPv4(value)).To(BeTrue())
			}

			badIPs := "255.255.256.255 E0925"
			for _, value := range strings.Fields(badIPs) {
				Expect(tools.IsIPv4(value)).To(BeFalse())
			}
		})
	})
})
