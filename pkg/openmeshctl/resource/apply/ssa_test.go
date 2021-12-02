package apply_test

import (
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	networkingv1 "github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/v1"
	"github.com/solo-io/gloo-mesh/pkg/openmeshctl/resource/apply"
	mock "github.com/solo-io/gloo-mesh/pkg/openmeshctl/resource/apply/mocks"
	"github.com/solo-io/gloo-mesh/pkg/openmeshctl/test/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("Server Side Apply", func() {
	const namespace = "gloo-mesh"

	var (
		ctrl *gomock.Controller
		ctx  *mock.MockContext
		// Use a generated gomock instead of the built in fake here because the
		// fake type does not support the server side apply patch type.
		dynamicClient *mock.MockInterface
		nsResourceReq *mock.MockNamespaceableResourceInterface
		resourceReq   *mock.MockResourceInterface

		accessPolicy *networkingv1.AccessPolicy
		ssa          apply.ApplyFunc
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		ctx = mock.NewMockContext(ctrl)
		dynamicClient = mock.NewMockInterface(ctrl)
		nsResourceReq = mock.NewMockNamespaceableResourceInterface(ctrl)
		resourceReq = mock.NewMockResourceInterface(ctrl)

		accessPolicy = util.AccessPolicy()

		ctx.EXPECT().ToRESTMapper().Return(util.NewRESTMapper(), nil).AnyTimes()
		ctx.EXPECT().DynamicClient().Return(dynamicClient, nil).AnyTimes()
		dynamicClient.EXPECT().Resource(schema.GroupVersionResource{
			Group:    "networking.mesh.gloo.solo.io",
			Version:  "v1",
			Resource: "accesspolicies",
		}).Return(nsResourceReq)

		ssa = apply.ServerSideApplier(&networkingv1.AccessPolicyGVK)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	JustBeforeEach(func() {
		// Allow changes to the access policy object before creating the mock
		resourceReq.EXPECT().Patch(
			ctx, accessPolicy.GetName(), types.ApplyPatchType,
			util.MustMarshalJSON(accessPolicy), metav1.PatchOptions{FieldManager: "meshctl"},
		)
	})

	When("the object has a namespace set", func() {
		BeforeEach(func() {
			nsResourceReq.EXPECT().Namespace(accessPolicy.GetNamespace()).Return(resourceReq)
		})

		It("should not error", func() {
			err := ssa.Apply(ctx, accessPolicy)
			Expect(err).ToNot(HaveOccurred())
		})
	})

	When("the object doesn't have a namespace set", func() {
		BeforeEach(func() {
			ctx.EXPECT().Namespace().Return(namespace)
			nsResourceReq.EXPECT().Namespace(namespace).Return(resourceReq)

			accessPolicy.SetNamespace("")
		})

		It("should use the context namespace", func() {
			err := ssa.Apply(ctx, accessPolicy)
			Expect(err).ToNot(HaveOccurred())
		})
	})
})
