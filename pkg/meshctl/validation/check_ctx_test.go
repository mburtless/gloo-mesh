package validation_test

import (
	"context"
	"os"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	mock_appsv1 "github.com/solo-io/external-apis/pkg/api/k8s/apps/v1/mocks"
	. "github.com/solo-io/gloo-mesh/pkg/meshctl/validation"
	"github.com/solo-io/skv2/pkg/crdutils"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("CheckCtx", func() {

	var (
		ctx  context.Context
		ctrl *gomock.Controller

		gmInstallNamespace = "gloo-mesh"

		mockAppsClientset *mock_appsv1.MockClientset

		mockDeploymentClient *mock_appsv1.MockDeploymentClient
	)

	BeforeEach(func() {
		// suppress stdout output to make test results easier to read
		os.Stdout, _ = os.Open(os.DevNull)

		ctrl, ctx = gomock.WithContext(context.Background(), GinkgoT())

		mockAppsClientset = mock_appsv1.NewMockClientset(ctrl)
		mockDeploymentClient = mock_appsv1.NewMockDeploymentClient(ctrl)

		mockAppsClientset.EXPECT().Deployments().Return(mockDeploymentClient)
	})

	It("should get metadata from deployment", func() {
		deploymentName := "enterprise-networking"

		mockDeploymentClient.EXPECT().
			GetDeployment(ctx, client.ObjectKey{
				Namespace: gmInstallNamespace,
				Name:      deploymentName,
			}).
			Return(
				&appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "enterprise-networking",
						Namespace: "gloo-mesh",
						Annotations: map[string]string{
							crdutils.CRDMetadataKey: `{"crds":[{"name":"destinations.discovery.mesh.gloo.solo.io","hash":"7e30f8d386339cbb"}],"version":"1.1.0"}`,
						},
					},
				}, nil,
			)

		testctx := NewTestCheckContext(gmInstallNamespace, 0, 0, nil, nil, mockAppsClientset, nil, nil, nil, nil, false, nil)

		md, err := testctx.CRDMetadata(ctx, deploymentName)
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
