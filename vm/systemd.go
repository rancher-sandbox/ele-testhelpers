package vm

import (
	"fmt"

	. "github.com/onsi/gomega" //nolint:revive
)

func SystemdUnitIsStarted(s string, st *SUT) {
	out, _ := st.Command(fmt.Sprintf("systemctl status %s", s))

	Expect(out).To(And(
		ContainSubstring(fmt.Sprintf("%s.service; enabled", s)),
		ContainSubstring("status=0/SUCCESS"),
	))
}

func SystemdUnitIsActive(s string, st *SUT) {
	out, _ := st.Command(fmt.Sprintf("systemctl is-active %s", s))

	Expect(out).To(Equal("active\n"))
}
