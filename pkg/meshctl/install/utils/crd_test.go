package utils_test

import (
	_ "embed"
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/solo-io/gloo-mesh/pkg/meshctl/install/utils"
	"github.com/solo-io/skv2/pkg/crdutils"
)

//go:embed examplemanifest.yaml
var exampleManifest []byte

var _ = Describe("Crd", func() {
	It("should get crd metadata from deployment", func() {
		result := &crdutils.CRDMetadata{
			CRDS: []crdutils.CRDAnnotations{
				{
					Name: "destinations.discovery.mesh.gloo.solo.io",
					Hash: "7e30f8d386339cbb",
				},
			},
			Version: "1.1.0",
		}
		fmt.Fprintln(GinkgoWriter, "input:\n", string(exampleManifest))
		md, err := GetCrdMetadataFromManifests("enterprise-networking", exampleManifest)
		Expect(err).NotTo(HaveOccurred())
		Expect(md).To(Equal(result))
	})

})
