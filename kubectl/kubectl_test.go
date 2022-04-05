package kubectl_test

import (
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/rancher-sandbox/ele-testhelpers/kubectl"
)

var _ = Describe("MachineRegistration e2e tests", func() {
	k := New()
	Context("registration", func() {
		It("creates a machine registration resource and a URL attaching CA certificate", func() {
			err := k.ApplyYAML("default", "test", &Pod{
				APIVersion: "v1",
				Kind:       "Pod",
				Metadata:   Metadata{Name: "test", Namespace: "default"},
				Spec: PodSpec{
					Containers: []Container{
						{Name: "test",
							Image:   "alpine",
							Command: []string{"/bin/sh", "-c"},
							Args:    []string{"sleep", "3600"},
						},
					},
				},
			})

			Expect(err).ShouldNot(HaveOccurred())

			defer func() {
				err := k.Delete("pod", "test")
				if err != nil {
					fmt.Fprintf(GinkgoWriter, "Error while deleting test pod: %v\n", err)
				}
			}()
			Eventually(func() []string {
				f, _ := k.GetPodNames("default", "")
				return f
			}, 10*time.Second, 1*time.Second).Should(ContainElement("test"))

		})

	})

})
