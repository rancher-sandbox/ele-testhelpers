package vm_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/rancher-sandbox/ele-testhelpers/vm"
	"os"
)

var _ = Describe("VM tests", func() {
	Describe("elementalCmd tests", func() {
		var sut *vm.SUT

		BeforeEach(func() {
			sut = vm.NewSUT()
		})
		Describe("With no args", func() {
			It("Sets default values", func() {
				cmd := sut.ElementalCmd()
				Expect(cmd).To(Equal("elemental --debug --logfile /tmp/elemental.log"))
			})
			It("Allows overriding default args via env var", func() {
				_ = os.Setenv("ELEMENTAL_CMD_ARGS", "--logfile /boot/vmlinuz")
				defer func() {
					_ = os.Unsetenv("ELEMENTAL_CMD_ARGS")
				}()
				cmd := sut.ElementalCmd()
				Expect(cmd).To(Equal("elemental --logfile /boot/vmlinuz"))
			})
		})
		Describe("With args", func() {
			It("Properly appends one arg to the default values", func() {
				cmd := sut.ElementalCmd("arg1")
				Expect(cmd).To(Equal("elemental --debug --logfile /tmp/elemental.log arg1"))
			})
			It("Properly appends n args to the default values", func() {
				cmd := sut.ElementalCmd("arg1", "arg2", "arg3")
				Expect(cmd).To(Equal("elemental --debug --logfile /tmp/elemental.log arg1 arg2 arg3"))
			})
			It("Allows overriding default args via env var with one arg", func() {
				_ = os.Setenv("ELEMENTAL_CMD_ARGS", "--logfile /boot/vmlinuz")
				defer func() {
					_ = os.Unsetenv("ELEMENTAL_CMD_ARGS")
				}()
				cmd := sut.ElementalCmd("arg1")
				Expect(cmd).To(Equal("elemental --logfile /boot/vmlinuz arg1"))
			})
			It("Allows overriding default args via env var with n args", func() {
				_ = os.Setenv("ELEMENTAL_CMD_ARGS", "--logfile /boot/vmlinuz")
				defer func() {
					_ = os.Unsetenv("ELEMENTAL_CMD_ARGS")
				}()
				cmd := sut.ElementalCmd("arg1", "arg2", "arg3")
				Expect(cmd).To(Equal("elemental --logfile /boot/vmlinuz arg1 arg2 arg3"))
			})
		})
	})
})
