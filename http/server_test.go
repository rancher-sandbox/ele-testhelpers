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
