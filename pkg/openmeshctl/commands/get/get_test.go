package get_test

import (
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	discoveryv1 "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1"
	"github.com/solo-io/gloo-mesh/pkg/openmeshctl/commands/get"
	mock "github.com/solo-io/gloo-mesh/pkg/openmeshctl/commands/get/mocks"
	"github.com/solo-io/gloo-mesh/pkg/openmeshctl/output"
	mock_output "github.com/solo-io/gloo-mesh/pkg/openmeshctl/output/mocks"
	mock_resource "github.com/solo-io/gloo-mesh/pkg/openmeshctl/resource/mocks"
	"github.com/solo-io/gloo-mesh/pkg/openmeshctl/test/util"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("Get Command", func() {
	const namespace = "gloo-mesh"

	var (
		ctrl      *gomock.Controller
		ctx       *mock.MockContext
		factory   *mock_resource.MockFactory
		formatter *mock_resource.MockFormatter
		printer   *mock_output.MockPrinter

		namespaceCall    *gomock.Call
		outputFormatCall *gomock.Call

		kubeClient    client.Client
		clientBuilder *fake.ClientBuilder
	)

	var expectTable = func(headers []string, objs []runtime.Object, includeNS, wide bool) {
		table := &output.Table{Headers: headers}
		formatter.EXPECT().ToTable(util.DiffEq(objs), includeNS, wide).Return(table)
		printer.EXPECT().PrintTable(util.DiffEq(table))
	}

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		ctx = mock.NewMockContext(ctrl)
		factory = mock_resource.NewMockFactory(ctrl)
		formatter = mock_resource.NewMockFormatter(ctrl)
		printer = mock_output.NewMockPrinter(ctrl)

		clientBuilder = util.NewClientBuilder()

		namespaceCall = ctx.EXPECT().Namespace()
		outputFormatCall = ctx.EXPECT().OutputFormat()

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

	Describe("Getting one resource by name", func() {
		var dest *discoveryv1.Destination

		BeforeEach(func() {
			dest = util.Destinations()[0]
			clientBuilder = clientBuilder.WithObjects(dest)
		})

		When("a namespace is selected", func() {
			type testCase struct {
				format       output.Format
				setupExpects func()
			}

			BeforeEach(func() {
				namespaceCall.Return(namespace).AnyTimes()
			})

			DescribeTable("output formats", func(tt testCase) {
				outputFormatCall.Return(tt.format).AnyTimes()
				tt.setupExpects()

				err := get.One(ctx, dest.GetName())
				Expect(err).ToNot(HaveOccurred())
			},
				Entry("Default", testCase{
					format: output.Default,
					setupExpects: func() {
						expectTable([]string{"foo", "bar"}, []runtime.Object{dest}, false, false)
					},
				}),
				Entry("Wide", testCase{
					format: output.Wide,
					setupExpects: func() {
						expectTable([]string{"foo", "bar", "baz"}, []runtime.Object{dest}, false, true)
					},
				}),
				Entry("YAML", testCase{
					format: output.YAML,
					setupExpects: func() {
						printer.EXPECT().PrintRaw(util.DiffEq(dest), output.YAML)
					},
				}),
				Entry("JSON", testCase{
					format: output.JSON,
					setupExpects: func() {
						printer.EXPECT().PrintRaw(util.DiffEq(dest), output.JSON)
					},
				}),
			)
		})

		When("no namespace is selected", func() {
			BeforeEach(func() {
				namespaceCall.Return("").AnyTimes()
				outputFormatCall.Times(0)
			})

			It("should return an error", func() {
				err := get.One(ctx, dest.GetName())
				Expect(err).To(MatchError("a resource cannot be retrieved by name across all namespaces"))
			})
		})
	})

	Describe("Getting all resources", func() {
		var (
			objs     []runtime.Object
			destList *discoveryv1.DestinationList
		)

		BeforeEach(func() {
			objs = util.DestinationObjects()
			destList = util.DestinationList()

			clientBuilder = clientBuilder.WithRuntimeObjects(objs...)
		})

		When("a namespace is selected", func() {
			type testCase struct {
				format output.Format
				setup  func()
			}

			BeforeEach(func() {
				namespaceCall.Return(namespace).AnyTimes()
			})

			DescribeTable("output format", func(tt testCase) {
				outputFormatCall.Return(tt.format).AnyTimes()
				tt.setup()

				err := get.All(ctx)
				Expect(err).ToNot(HaveOccurred())
			},
				Entry("Default", testCase{
					format: output.Default,
					setup: func() {
						expectTable([]string{"foo", "bar"}, objs, false, false)
					},
				}),
				Entry("Wide", testCase{
					format: output.Wide,
					setup: func() {
						expectTable([]string{"foo", "bar", "baz"}, objs, false, true)
					},
				}),
				Entry("YAML", testCase{
					format: output.YAML,
					setup: func() {
						printer.EXPECT().PrintRaw(util.DiffEq(destList), output.YAML)
					},
				}),
				Entry("JSON", testCase{
					format: output.JSON,
					setup: func() {
						printer.EXPECT().PrintRaw(util.DiffEq(destList), output.JSON)
					},
				}),
			)
		})

		When("all namespaces are selected", func() {
			BeforeEach(func() {
				objs[0].(*discoveryv1.Destination).SetNamespace("another-ns")
				destList.Items[0].SetNamespace("another-ns")
				clientBuilder = util.NewClientBuilder().WithLists(destList)

				namespaceCall.Return("").AnyTimes()
				outputFormatCall.Return(output.Default).AnyTimes()
				expectTable([]string{"foo", "bar"}, objs, true, false)
			})

			It("should return objects from all namespaces", func() {
				err := get.All(ctx)
				Expect(err).ToNot(HaveOccurred())
			})
		})
	})
})
