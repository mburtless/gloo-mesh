package access_test

import (
	"context"
	"strings"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v1beta1sets "github.com/solo-io/external-apis/pkg/api/istio/security.istio.io/v1beta1/sets"
	commonv1 "github.com/solo-io/gloo-mesh/pkg/api/common.mesh.gloo.solo.io/v1"
	discoveryv1 "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1"
	"github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/input"
	"github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/output/istio"
	networkingv1 "github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/v1"
	mock_reporting "github.com/solo-io/gloo-mesh/pkg/mesh-networking/reporting/mocks"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/istio/mesh/access"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/utils/metautils"
	v1 "github.com/solo-io/skv2/pkg/api/core.skv2.solo.io/v1"
	"github.com/solo-io/skv2/pkg/ezkube"
	securityv1beta1spec "istio.io/api/security/v1beta1"
	"istio.io/api/type/v1beta1"
	securityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("AccessPolicyTranslator", func() {
	var (
		translator   access.Translator
		mockReporter *mock_reporting.MockReporter
	)

	BeforeEach(func() {
		ctrl := gomock.NewController(GinkgoT())
		translator = access.NewTranslator(context.Background())
		mockReporter = mock_reporting.NewMockReporter(ctrl)
	})

	It("should translate an AuthorizationPolicy for the ingress gateway and in the installation namespace", func() {
		ingressDestination := &discoveryv1.Destination{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "istio-ingressgateway",
				Namespace: "istio-system",
			},
			Spec: discoveryv1.DestinationSpec{
				Type: &discoveryv1.DestinationSpec_KubeService_{
					KubeService: &discoveryv1.DestinationSpec_KubeService{
						Ref: &v1.ClusterObjectRef{
							Name:        "istio-ingressgateway",
							Namespace:   "istio-system",
							ClusterName: "cluster-name",
						},
						WorkloadSelectorLabels: map[string]string{"gateway": "selector"},
					},
				},
			},
		}

		in := input.NewInputLocalSnapshotManualBuilder("test").
			AddDestinations(discoveryv1.DestinationSlice{ingressDestination}).
			Build()

		mesh := &discoveryv1.Mesh{
			Spec: discoveryv1.MeshSpec{
				Type: &discoveryv1.MeshSpec_Istio_{
					Istio: &discoveryv1.MeshSpec_Istio{
						Installation: &discoveryv1.MeshInstallation{
							Namespace: "istio-system",
							Cluster:   "cluster-name",
						},
					},
				},
			},
			Status: discoveryv1.MeshStatus{
				AppliedVirtualMesh: &discoveryv1.MeshStatus_AppliedVirtualMesh{
					Ref: &v1.ObjectRef{
						Name:      "virtual-mesh",
						Namespace: "gloo-mesh",
					},
					Spec: &networkingv1.VirtualMeshSpec{
						GlobalAccessPolicy: networkingv1.VirtualMeshSpec_ENABLED,
					},
				},
				AppliedEastWestIngressGateways: []*commonv1.AppliedIngressGateway{
					{
						DestinationRef: ezkube.MakeObjectRef(ingressDestination),
					},
				},
			},
		}
		expectedAuthPolicies := v1beta1sets.NewAuthorizationPolicySet(
			&securityv1beta1.AuthorizationPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name:        access.IngressGatewayAuthPolicyName + "-istio-ingressgateway-istio-system",
					Namespace:   "istio-system",
					ClusterName: "cluster-name",
					Labels:      metautils.TranslatedObjectLabels(),
					Annotations: map[string]string{
						metautils.ParentLabelkey: `{"networking.mesh.gloo.solo.io/v1, Kind=VirtualMesh":[{"name":"virtual-mesh","namespace":"gloo-mesh"}]}`,
					},
				},
				Spec: securityv1beta1spec.AuthorizationPolicy{
					Action: securityv1beta1spec.AuthorizationPolicy_ALLOW,
					// A single empty rule allows all traffic.
					// Reference: https://istio.io/docs/reference/config/security/authorization-policy/#AuthorizationPolicy
					Rules: []*securityv1beta1spec.Rule{{}},
					Selector: &v1beta1.WorkloadSelector{
						MatchLabels: ingressDestination.Spec.GetKubeService().GetWorkloadSelectorLabels(),
					},
				},
			},
			&securityv1beta1.AuthorizationPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name:        access.GlobalAccessControlAuthPolicyName,
					Namespace:   "istio-system",
					ClusterName: "cluster-name",
					Labels:      metautils.TranslatedObjectLabels(),
					Annotations: map[string]string{
						metautils.ParentLabelkey: `{"networking.mesh.gloo.solo.io/v1, Kind=VirtualMesh":[{"name":"virtual-mesh","namespace":"gloo-mesh"}]}`,
					},
				},
				Spec: securityv1beta1spec.AuthorizationPolicy{},
			},
		)
		outputs := istio.NewBuilder(context.TODO(), "")
		translator.Translate(in, mesh, mesh.Status.AppliedVirtualMesh, outputs, nil)
		Expect(outputs.GetAuthorizationPolicies()).To(Equal(expectedAuthPolicies))
	})

	It("should translate hash a long AuthorizationPolicy name", func() {
		ingressDestination := &discoveryv1.Destination{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "a-quick-brown-istio-ingressgateway-jumped-over-the-lazy-red-dog",
				Namespace: "istio-system",
			},
			Spec: discoveryv1.DestinationSpec{
				Type: &discoveryv1.DestinationSpec_KubeService_{
					KubeService: &discoveryv1.DestinationSpec_KubeService{
						Ref: &v1.ClusterObjectRef{
							Name:        "istio-ingressgateway",
							Namespace:   "istio-system",
							ClusterName: "cluster-name",
						},
						WorkloadSelectorLabels: map[string]string{"gateway": "selector"},
					},
				},
			},
		}

		in := input.NewInputLocalSnapshotManualBuilder("test").
			AddDestinations(discoveryv1.DestinationSlice{ingressDestination}).
			Build()

		mesh := &discoveryv1.Mesh{
			Spec: discoveryv1.MeshSpec{
				Type: &discoveryv1.MeshSpec_Istio_{
					Istio: &discoveryv1.MeshSpec_Istio{
						Installation: &discoveryv1.MeshInstallation{
							Namespace: "istio-system",
							Cluster:   "cluster-name",
						},
					},
				},
			},
			Status: discoveryv1.MeshStatus{
				AppliedVirtualMesh: &discoveryv1.MeshStatus_AppliedVirtualMesh{
					Ref: &v1.ObjectRef{
						Name:      "virtual-mesh",
						Namespace: "gloo-mesh",
					},
					Spec: &networkingv1.VirtualMeshSpec{
						GlobalAccessPolicy: networkingv1.VirtualMeshSpec_ENABLED,
					},
				},
				AppliedEastWestIngressGateways: []*commonv1.AppliedIngressGateway{
					{
						DestinationRef: ezkube.MakeObjectRef(ingressDestination),
					},
				},
			},
		}
		expectedAuthPolicies := v1beta1sets.NewAuthorizationPolicySet(
			&securityv1beta1.AuthorizationPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name:        access.IngressGatewayAuthPolicyName + "-12556735161280984284",
					Namespace:   "istio-system",
					ClusterName: "cluster-name",
					Labels:      metautils.TranslatedObjectLabels(),
					Annotations: map[string]string{
						metautils.ParentLabelkey: `{"networking.mesh.gloo.solo.io/v1, Kind=VirtualMesh":[{"name":"virtual-mesh","namespace":"gloo-mesh"}]}`,
					},
				},
				Spec: securityv1beta1spec.AuthorizationPolicy{
					Action: securityv1beta1spec.AuthorizationPolicy_ALLOW,
					// A single empty rule allows all traffic.
					// Reference: https://istio.io/docs/reference/config/security/authorization-policy/#AuthorizationPolicy
					Rules: []*securityv1beta1spec.Rule{{}},
					Selector: &v1beta1.WorkloadSelector{
						MatchLabels: ingressDestination.Spec.GetKubeService().GetWorkloadSelectorLabels(),
					},
				},
			},
			&securityv1beta1.AuthorizationPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name:        access.GlobalAccessControlAuthPolicyName,
					Namespace:   "istio-system",
					ClusterName: "cluster-name",
					Labels:      metautils.TranslatedObjectLabels(),
					Annotations: map[string]string{
						metautils.ParentLabelkey: `{"networking.mesh.gloo.solo.io/v1, Kind=VirtualMesh":[{"name":"virtual-mesh","namespace":"gloo-mesh"}]}`,
					},
				},
				Spec: securityv1beta1spec.AuthorizationPolicy{},
			},
		)
		outputs := istio.NewBuilder(context.TODO(), "")
		translator.Translate(in, mesh, mesh.Status.AppliedVirtualMesh, outputs, nil)
		Expect(outputs.GetAuthorizationPolicies()).To(Equal(expectedAuthPolicies))
	})

	It("should not translate any AuthorizationPolicies", func() {
		mesh := &discoveryv1.Mesh{
			Spec: discoveryv1.MeshSpec{
				Type: &discoveryv1.MeshSpec_Istio_{
					Istio: &discoveryv1.MeshSpec_Istio{
						Installation: &discoveryv1.MeshInstallation{
							Namespace: "istio-system",
						},
					},
				},
			},
			Status: discoveryv1.MeshStatus{
				AppliedVirtualMesh: &discoveryv1.MeshStatus_AppliedVirtualMesh{
					Spec: &networkingv1.VirtualMeshSpec{
						GlobalAccessPolicy: networkingv1.VirtualMeshSpec_DISABLED,
					},
				},
			},
		}
		outputs := istio.NewBuilder(context.TODO(), "")
		translator.Translate(nil, mesh, mesh.Status.AppliedVirtualMesh, outputs, nil)
		Expect(outputs.GetAuthorizationPolicies().Length()).To(Equal(0))
	})

	It("should report error when ingress Destination is not found in input snapshot", func() {
		in := input.NewInputLocalSnapshotManualBuilder("test").Build()

		mesh := &discoveryv1.Mesh{
			Spec: discoveryv1.MeshSpec{
				Type: &discoveryv1.MeshSpec_Istio_{
					Istio: &discoveryv1.MeshSpec_Istio{
						Installation: &discoveryv1.MeshInstallation{
							Namespace: "istio-system",
						},
					},
				},
			},
			Status: discoveryv1.MeshStatus{
				AppliedVirtualMesh: &discoveryv1.MeshStatus_AppliedVirtualMesh{
					Ref: &v1.ObjectRef{
						Name:      "virtual-mesh",
						Namespace: "gloo-mesh",
					},
					Spec: &networkingv1.VirtualMeshSpec{
						GlobalAccessPolicy: networkingv1.VirtualMeshSpec_DISABLED,
					},
				},
				AppliedEastWestIngressGateways: []*commonv1.AppliedIngressGateway{
					{
						DestinationRef: &v1.ObjectRef{
							Name:      "ingress-dest",
							Namespace: "istio-system",
						},
					},
				},
			},
		}

		mockReporter.EXPECT().
			ReportVirtualMeshToMesh(mesh, mesh.Status.GetAppliedVirtualMesh().GetRef(), gomock.Any()).
			DoAndReturn(func(mesh *discoveryv1.Mesh, virtualMesh ezkube.ResourceId, err error) {
				Expect(strings.Contains(err.Error(), "creating AuthorizationPolicy for east west ingress gateways")).To(BeTrue())
			})

		outputs := istio.NewBuilder(context.TODO(), "")
		translator.Translate(in, mesh, mesh.Status.AppliedVirtualMesh, outputs, mockReporter)
		Expect(outputs.GetAuthorizationPolicies().Length()).To(Equal(0))
	})
})
