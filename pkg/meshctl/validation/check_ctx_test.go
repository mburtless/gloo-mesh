package validation_test

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/solo-io/gloo-mesh/pkg/meshctl/validation"
	"github.com/solo-io/skv2/pkg/crdutils"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("CheckCtx", func() {
	It("should get metadata from deployment", func() {
		cli := fake.NewFakeClient(&appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "enterprise-networking",
				Namespace: "gloo-mesh",
				Annotations: map[string]string{
					crdutils.CRDMetadataKey: `{"crds":[{"name":"destinations.discovery.mesh.gloo.solo.io","hash":"7e30f8d386339cbb"}],"version":"1.1.0"}`,
				},
			},
		})

		testctx := NewTestCheckContext(cli, "gloo-mesh", 0, 0, nil, false, nil)
		md, err := testctx.CRDMetadata(context.Background(), "enterprise-networking")
		Expect(err).NotTo(HaveOccurred())
		result := &crdutils.CRDMetadata{
			CRDS: []crdutils.CRDAnnotations{
				{
					Name: "destinations.discovery.mesh.gloo.solo.io",
					Hash: "7e30f8d386339cbb",
				},
			},
			Version: "1.1.0",
		}
		Expect(md).To(Equal(result))
	})
})
