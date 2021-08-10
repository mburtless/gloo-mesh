package istio

import (
	"context"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rotisserie/eris"
	discoveryv1 "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1"
	"github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/input"
	"github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/output/istio"
	mock_istio_output "github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/output/istio/mocks"
	"github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/output/local"
	mock_local_output "github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/output/local/mocks"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/reporting"
	mock_reporting "github.com/solo-io/gloo-mesh/pkg/mesh-networking/reporting/mocks"
	mock_destination "github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/istio/destination/mocks"
	mock_extensions "github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/istio/extensions/mocks"
	mock_istio "github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/istio/internal/mocks"
	mock_mesh "github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/istio/mesh/mocks"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/utils/metautils"
	"github.com/solo-io/go-utils/contextutils"
	v1 "github.com/solo-io/skv2/pkg/api/core.skv2.solo.io/v1"
	multiclusterv1alpha1 "github.com/solo-io/skv2/pkg/api/multicluster.solo.io/v1alpha1"
	"github.com/solo-io/skv2/pkg/resource"
	security_istio_io_v1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var _ = Describe("IstioNetworkingTranslator", func() {
	var (
		ctrl                      *gomock.Controller
		ctx                       context.Context
		ctxWithValue              context.Context
		mockIstioExtender         *mock_extensions.MockIstioExtender
		mockReporter              *mock_reporting.MockReporter
		mockIstioOutputs          *mock_istio_output.MockBuilder
		mockLocalOutputs          *mock_local_output.MockBuilder
		mockDestinationTranslator *mock_destination.MockTranslator
		mockMeshTranslator        *mock_mesh.MockTranslator
		mockDependencyFactory     *mock_istio.MockDependencyFactory
		translator                *istioTranslator
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		ctx = context.TODO()
		ctxWithValue = contextutils.WithLogger(context.TODO(), "istio-translator-0")
		mockIstioExtender = mock_extensions.NewMockIstioExtender(ctrl)
		mockReporter = mock_reporting.NewMockReporter(ctrl)
		mockDestinationTranslator = mock_destination.NewMockTranslator(ctrl)
		mockMeshTranslator = mock_mesh.NewMockTranslator(ctrl)
		mockDependencyFactory = mock_istio.NewMockDependencyFactory(ctrl)
		mockIstioOutputs = mock_istio_output.NewMockBuilder(ctrl)
		mockLocalOutputs = mock_local_output.NewMockBuilder(ctrl)
		translator = &istioTranslator{
			dependencies:            mockDependencyFactory,
			extender:                mockIstioExtender,
			translationOutputsCache: newPreservedTranslationOutputs(),
		}
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("should call the individual translators for each discovery output object", func() {
		in := input.NewInputLocalSnapshotManualBuilder("").
			AddKubernetesClusters([]*multiclusterv1alpha1.KubernetesCluster{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "kube-cluster",
						Namespace: "namespace",
					},
				},
			}).
			AddMeshes([]*discoveryv1.Mesh{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "mesh-1",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "mesh-2",
					},
				},
			}).
			AddWorkloads([]*discoveryv1.Workload{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:   "mesh-workload-1",
						Labels: metautils.TranslatedObjectLabels(),
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:   "mesh-workload-2",
						Labels: metautils.TranslatedObjectLabels(),
					},
				},
			}).
			AddDestinations([]*discoveryv1.Destination{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:   "traffic-target-1",
						Labels: metautils.TranslatedObjectLabels(),
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:   "traffic-target-2",
						Labels: metautils.TranslatedObjectLabels(),
					},
				},
			}).Build()

		contextMatcher := gomock.Any()
		mockDependencyFactory.
			EXPECT().
			MakeDestinationTranslator(contextMatcher, nil, in.KubernetesClusters(), in.Destinations()).
			Return(mockDestinationTranslator)

		for _, destination := range in.Destinations().List() {
			mockDestinationTranslator.
				EXPECT().
				Translate(in, destination, mockIstioOutputs, mockReporter)
		}

		mockDependencyFactory.
			EXPECT().
			MakeMeshTranslator(ctxWithValue, in.Secrets(), in.Workloads()).
			Return(mockMeshTranslator)
		for _, mesh := range in.Meshes().List() {
			perMeshIstioOuputs := gomock.AssignableToTypeOf(istio.NewBuilder(nil, ""))
			perMeshLocalOuputs := gomock.AssignableToTypeOf(local.NewBuilder(nil, ""))
			mockMeshTranslator.
				EXPECT().
				Translate(
					in,
					mesh,
					perMeshIstioOuputs, // a new istio builder is constructed for each mesh
					perMeshLocalOuputs, // a new local builder is constructed for each mesh
					gomock.AssignableToTypeOf(&reportInterceptor{}), // a new report interceptor is constructed for each mesh.
				)

			// each mesh output will be merged with the final outputs will
			mockIstioOutputs.EXPECT().Merge(perMeshIstioOuputs)
			mockLocalOutputs.EXPECT().Merge(perMeshLocalOuputs)

		}

		mockIstioExtender.EXPECT().PatchOutputs(contextMatcher, in, mockIstioOutputs)

		translator.Translate(ctx, in, nil, mockIstioOutputs, mockLocalOutputs, mockReporter)
	})

	It("should preserve the outputs for a mesh when a successive translation results in an error", func() {
		in := input.NewInputLocalSnapshotManualBuilder("").
			AddKubernetesClusters([]*multiclusterv1alpha1.KubernetesCluster{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "kube-cluster",
						Namespace: "namespace",
					},
				},
			}).
			AddMeshes([]*discoveryv1.Mesh{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "mesh-1",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "mesh-2",
					},
				},
			}).Build()

		contextMatcher := gomock.Any()
		mockDependencyFactory.
			EXPECT().
			MakeDestinationTranslator(contextMatcher, nil, in.KubernetesClusters(), in.Destinations()).
			Return(mockDestinationTranslator)

		mockDependencyFactory.
			EXPECT().
			MakeMeshTranslator(ctxWithValue, in.Secrets(), in.Workloads()).
			Return(mockMeshTranslator)

		for _, mesh := range in.Meshes().List() {
			perMeshIstioOuputs := gomock.AssignableToTypeOf(istio.NewBuilder(nil, ""))
			perMeshLocalOuputs := gomock.AssignableToTypeOf(local.NewBuilder(nil, ""))
			mockMeshTranslator.
				EXPECT().
				Translate(
					in,
					mesh,
					perMeshIstioOuputs, // a new istio builder is constructed for each mesh
					perMeshLocalOuputs, // a new local builder is constructed for each mesh
					gomock.AssignableToTypeOf(&reportInterceptor{}), // a new report interceptor is constructed for each mesh.
				)

			// each mesh output will be merged with the final outputs will
			mockIstioOutputs.EXPECT().Merge(perMeshIstioOuputs)
			mockLocalOutputs.EXPECT().Merge(perMeshLocalOuputs)

		}

		mockIstioExtender.EXPECT().PatchOutputs(contextMatcher, in, mockIstioOutputs)

		translator.Translate(ctx, in, nil, mockIstioOutputs, mockLocalOutputs, mockReporter)
	})

	It("should preserve the outputs for a mesh when a successive translation results in an error", func() {

		in := input.NewInputLocalSnapshotManualBuilder("").
			AddKubernetesClusters([]*multiclusterv1alpha1.KubernetesCluster{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "kube-cluster",
						Namespace: "namespace",
					},
				},
			}).Build()

		// we expect these outputs to be preserverd
		preservedIstioOutput := &security_istio_io_v1beta1.AuthorizationPolicy{
			ObjectMeta: metav1.ObjectMeta{
				Name: "authpolicy-1",
			},
		}
		preservedLocalOutput := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name: "authpolicy-1",
			},
		}

		mesh := &discoveryv1.Mesh{
			ObjectMeta: metav1.ObjectMeta{
				Name: "mesh-1",
			},
		}

		// first, translation should succeed so we cache outputs
		successTranslator := &testMeshTranslator{tx: func(
			in input.LocalSnapshot,
			mesh *discoveryv1.Mesh,
			istioOutputs istio.Builder,
			localOutputs local.Builder,
			reporter reporting.Reporter,
		) {
			// "translate" successfully
			istioOutputs.AddAuthorizationPolicies(preservedIstioOutput)
			localOutputs.AddSecrets(preservedLocalOutput)
		}}

		actualIstio, actualLocal := translator.translateMesh(
			ctx,
			in,
			mesh,
			successTranslator,
			mockReporter,
		)

		// assert expected outputs
		var actualIstioOutput resource.TypedObject
		actualIstio.ForEachObject(func(_ string, _ schema.GroupVersionKind, obj resource.TypedObject) {
			actualIstioOutput = obj
		})
		Expect(actualIstioOutput).To(Equal(preservedIstioOutput))

		var actualLocalOutput resource.TypedObject
		actualLocal.ForEachObject(func(_ string, _ schema.GroupVersionKind, obj resource.TypedObject) {
			actualLocalOutput = obj
		})
		Expect(actualLocalOutput).To(Equal(preservedLocalOutput))

		// next, translation should fail, and we use cached outputs
		translationErr := eris.Errorf("oopsie doopsie")
		virtualMeshRef := &v1.ObjectRef{}
		failureTranslator := &testMeshTranslator{tx: func(
			in input.LocalSnapshot,
			mesh *discoveryv1.Mesh,
			istioOutputs istio.Builder,
			localOutputs local.Builder,
			reporter reporting.Reporter,
		) {
			// "translate" should call reporter and fail
			reporter.ReportVirtualMeshToMesh(mesh, virtualMeshRef, translationErr)
		}}

		// expect underlying reporter to be called for mesh
		mockReporter.EXPECT().ReportVirtualMeshToMesh(mesh, virtualMeshRef, translationErr)

		actualIstio, actualLocal = translator.translateMesh(
			ctx,
			in,
			mesh,
			failureTranslator,
			mockReporter,
		)

		// assert expected outputs are preserved
		var actualPreservedIstioOutput resource.TypedObject
		actualIstio.ForEachObject(func(_ string, _ schema.GroupVersionKind, obj resource.TypedObject) {
			actualPreservedIstioOutput = obj
		})
		Expect(actualPreservedIstioOutput).To(Equal(preservedIstioOutput))

		var actualPresrvedLocalOutput resource.TypedObject
		actualLocal.ForEachObject(func(_ string, _ schema.GroupVersionKind, obj resource.TypedObject) {
			actualPresrvedLocalOutput = obj
		})
		Expect(actualPresrvedLocalOutput).To(Equal(preservedLocalOutput))
	})
})

type testMeshTranslator struct {
	tx func(in input.LocalSnapshot, mesh *discoveryv1.Mesh, istioOutputs istio.Builder, localOutputs local.Builder, reporter reporting.Reporter)
}

func (t *testMeshTranslator) Translate(in input.LocalSnapshot, mesh *discoveryv1.Mesh, istioOutputs istio.Builder, localOutputs local.Builder, reporter reporting.Reporter) {
	t.tx(in, mesh, istioOutputs, localOutputs, reporter)
}
