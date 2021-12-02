package describe_test

import (
	"io"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	discoveryv1 "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1"
	"github.com/solo-io/gloo-mesh/pkg/openmeshctl/commands/describe"
	mock "github.com/solo-io/gloo-mesh/pkg/openmeshctl/commands/describe/mocks"
	"github.com/solo-io/gloo-mesh/pkg/openmeshctl/output"
	mock_output "github.com/solo-io/gloo-mesh/pkg/openmeshctl/output/mocks"
	mock_resource "github.com/solo-io/gloo-mesh/pkg/openmeshctl/resource/mocks"
	"github.com/solo-io/gloo-mesh/pkg/openmeshctl/test/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("Describe", func() {
	const namespace = "gloo-mesh"

	var (
		ctrl      *gomock.Controller
		ctx       *mock.MockContext
		factory   *mock_resource.MockFactory
		formatter *mock_resource.MockFormatter
		printer   *mock_output.MockPrinter

		namespaceCall *gomock.Call

		kubeClient    client.Client
		clientBuilder *fake.ClientBuilder
	)

	var expectObject = func(obj metav1.Object) {
		summary := &output.Summary{Meta: obj}
		formatter.EXPECT().ToSummary(util.DiffEq(obj)).Return(summary)
		printer.EXPECT().PrintSummary(util.DiffEq(summary))
	}

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		ctx = mock.NewMockContext(ctrl)
		factory = mock_resource.NewMockFactory(ctrl)
		formatter = mock_resource.NewMockFormatter(ctrl)
		printer = mock_output.NewMockPrinter(ctrl)
		clientBuilder = util.NewClientBuilder()

		namespaceCall = ctx.EXPECT().Namespace()
		ctx.EXPECT().Out().Return(io.Discard).AnyTimes()
		ctx.EXPECT().Factory().Return(factory).AnyTimes()
		ctx.EXPECT().Formatter().Return(formatter).AnyTimes()
		ctx.EXPECT().Printer().Return(printer).AnyTimes()
		factory.EXPECT().New().Return(&discoveryv1.Destination{}).AnyTimes()
		factory.EXPECT().NewList().Return(&discoveryv1.DestinationList{}).AnyTimes()
	})

	JustBeforeEach(func() {
		kubeClient = clientBuilder.Build()
		ctx.EXPECT().KubeClient().Return(kubeClient, nil).AnyTimes()
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Describe("Describing one resource by name", func() {
		var dest *discoveryv1.Destination

		BeforeEach(func() {
			dest = util.Destinations()[0]
			clientBuilder = clientBuilder.WithObjects(dest)
		})

		When("a namespace is selected", func() {
			BeforeEach(func() {
				namespaceCall.Return(dest.GetNamespace()).AnyTimes()
				expectObject(dest)
			})

			It("should print the resource summary", func() {
				err := describe.One(ctx, dest.GetName())
				Expect(err).ToNot(HaveOccurred())
			})
		})

		When("no namespace is selected", func() {
			BeforeEach(func() {
				namespaceCall.Return("").AnyTimes()
			})

			It("should return an error", func() {
				err := describe.One(ctx, dest.GetName())
				Expect(err).To(MatchError("a resource cannot be retrieved by name across all namespaces"))
			})
		})
	})

	Describe("Describing all resources", func() {
		var dests []*discoveryv1.Destination

		BeforeEach(func() {
			dests = util.Destinations()
			clientBuilder = clientBuilder.WithRuntimeObjects(util.DestinationObjects()...)
		})

		When("a namespace is selected", func() {
			BeforeEach(func() {
				namespaceCall.Return(namespace).AnyTimes()

				for _, dest := range dests {
					expectObject(dest)
				}
			})

			It("should print summaries for all resources", func() {
				err := describe.All(ctx)
				Expect(err).ToNot(HaveOccurred())
			})
		})

		When("all namespaces are selected", func() {
			BeforeEach(func() {
				objs := util.DestinationObjects()
				objs[0].(*discoveryv1.Destination).SetNamespace("another-ns")
				dests[0].SetNamespace("another-ns")
				clientBuilder = util.NewClientBuilder().WithRuntimeObjects(objs...)

				namespaceCall.Return("").AnyTimes()
				for _, dest := range dests {
					expectObject(dest)
				}
			})

			It("should print summaries for all resources", func() {
				err := describe.All(ctx)
				Expect(err).ToNot(HaveOccurred())
			})
		})
	})
})
