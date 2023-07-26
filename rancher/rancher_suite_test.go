package rancher_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestRancher(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "rancher test Suite")
}
