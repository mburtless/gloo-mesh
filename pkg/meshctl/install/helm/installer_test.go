package helm_test

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/solo-io/gloo-mesh/pkg/meshctl/install/gloomesh"
	. "github.com/solo-io/gloo-mesh/pkg/meshctl/install/helm"
)

var _ = Describe("Installer", func() {

	// NOTE: This test pulls a chart from a bucket, so needs to have internet to work.
	It("should get rendered manifests from installer", func() {
		gmeInstaller := Installer{
			ChartUri:    fmt.Sprintf(gloomesh.GlooMeshEnterpriseChartUriTemplate, "1.0.0"),
			KubeContext: "nonexistent-kube-context",
			Namespace:   "invalid namespace",
			ReleaseName: "invalid release name",
			Values:      map[string]string{"licenseKey": "not a license key"},
			Verbose:     true,
		}
		manifests, err := gmeInstaller.GetRenderedManifests(context.Background())
		Expect(err).NotTo(HaveOccurred())
		// make sure that we see something that we expect
		Expect(string(manifests)).To(ContainSubstring("enterprise-networking"))
	})
})
