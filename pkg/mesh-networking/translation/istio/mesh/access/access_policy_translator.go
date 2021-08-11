package access

import (
	"context"

	"github.com/rotisserie/eris"
	commonv1 "github.com/solo-io/gloo-mesh/pkg/api/common.mesh.gloo.solo.io/v1"
	discoveryv1 "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1"
	v1sets "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1/sets"
	"github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/input"
	"github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/output/istio"
	v1 "github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/v1"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/reporting"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/utils/metautils"
	"github.com/solo-io/skv2/contrib/pkg/sets"
	securityv1beta1spec "istio.io/api/security/v1beta1"
	"istio.io/api/type/v1beta1"
	securityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//go:generate mockgen -source ./access_policy_translator.go -destination mocks/access_policy_translator.go

// The access control translator translates a VirtualMesh EnforcementPolicy into Istio AuthorizationPolicies.
type Translator interface {
	// Returns nil if no AuthorizationPolicies are required for the mesh (i.e. because AccessPolicy enforcement is disabled).
	// Output resources will be added to the istio.Builder
	// Errors caused by invalid user config will be reported using the Reporter.
	Translate(
		in input.LocalSnapshot,
		mesh *discoveryv1.Mesh,
		virtualMesh *discoveryv1.MeshStatus_AppliedVirtualMesh,
		outputs istio.Builder,
		reporter reporting.Reporter,
	)
}

const (
	IngressGatewayAuthPolicyName      = "allow-ingress-gateway"
	GlobalAccessControlAuthPolicyName = "global-access-control"
)

type translator struct {
	ctx context.Context
}

func NewTranslator(ctx context.Context) Translator {
	return &translator{ctx}
}

func (t *translator) Translate(
	in input.LocalSnapshot,
	mesh *discoveryv1.Mesh,
	virtualMesh *discoveryv1.MeshStatus_AppliedVirtualMesh,
	outputs istio.Builder,
	reporter reporting.Reporter,
) {
	istioMesh := mesh.Spec.GetIstio()
	if istioMesh == nil {
		return
	}

	// Istio's default access enforcement policy is disabled.
	if virtualMesh.Spec.GlobalAccessPolicy == v1.VirtualMeshSpec_MESH_DEFAULT ||
		virtualMesh.Spec.GlobalAccessPolicy == v1.VirtualMeshSpec_DISABLED {
		return
	}
	clusterName := istioMesh.Installation.Cluster
	installationNamespace := istioMesh.Installation.Namespace
	globalAuthPolicy := buildGlobalAuthPolicy(installationNamespace, clusterName)
	ingressGatewayAuthPolicies, err := buildAuthPoliciesForIngressGateways(
		in.Destinations(),
		installationNamespace,
		clusterName,
		mesh.Status.GetAppliedEastWestIngressGateways(),
	)
	if err != nil {
		reporter.ReportVirtualMeshToMesh(mesh, virtualMesh.GetRef(), eris.Wrap(err, "creating AuthorizationPolicy for east west ingress gateways"))
	}

	// Append the VirtualMesh as a parent to each output resource
	metautils.AppendParent(t.ctx, globalAuthPolicy, virtualMesh.GetRef(), v1.VirtualMesh{}.GVK())
	for _, ap := range ingressGatewayAuthPolicies {
		metautils.AppendParent(t.ctx, ap, virtualMesh.GetRef(), v1.VirtualMesh{}.GVK())
	}

	outputs.AddAuthorizationPolicies(globalAuthPolicy)
	outputs.AddAuthorizationPolicies(ingressGatewayAuthPolicies...)
}

// Creates an AuthorizationPolicy that allows all traffic into the service
// which backs the Gateway used for multi cluster traffic.
func buildAuthPoliciesForIngressGateways(
	destinations v1sets.DestinationSet,
	installationNamespace string,
	clusterName string,
	ingressGateways []*commonv1.AppliedIngressGateway,
) ([]*securityv1beta1.AuthorizationPolicy, error) {
	var authPolicies []*securityv1beta1.AuthorizationPolicy
	for _, ingressGateway := range ingressGateways {
		destination, err := destinations.Find(ingressGateway.GetDestinationRef())
		if err != nil {
			return nil, err
		}

		ap := &securityv1beta1.AuthorizationPolicy{
			ObjectMeta: metav1.ObjectMeta{
				Name:        ingressGatewayAuthPolicyName(ingressGateway),
				Namespace:   installationNamespace,
				ClusterName: clusterName,
				Labels:      metautils.TranslatedObjectLabels(),
			},
			Spec: securityv1beta1spec.AuthorizationPolicy{
				Action: securityv1beta1spec.AuthorizationPolicy_ALLOW,
				// A single empty rule allows all traffic.
				// Reference: https://istio.io/docs/reference/config/security/authorization-policy/#AuthorizationPolicy
				Rules: []*securityv1beta1spec.Rule{{}},
				Selector: &v1beta1.WorkloadSelector{
					MatchLabels: destination.Spec.GetKubeService().GetWorkloadSelectorLabels(),
				},
			},
		}

		authPolicies = append(authPolicies, ap)
	}
	return authPolicies, nil
}

// Creates a global AuthorizationPolicy that denies all traffic within the Mesh unless explicitly allowed by GlooMesh AccessControl resources.
func buildGlobalAuthPolicy(
	installationNamespace,
	clusterName string,
) *securityv1beta1.AuthorizationPolicy {
	// The following config denies all traffic in the mesh because it defaults to an ALLOW rule that doesn't match any requests,
	// set to the installation namespace so it affects all namespaces,
	// thereby denying traffic unless explicitly allowed by the user through additional AuthorizationPolicies.
	// https://istio.io/docs/reference/config/security/authorization-policy/#AuthorizationPolicy
	return &securityv1beta1.AuthorizationPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:        GlobalAccessControlAuthPolicyName,
			Namespace:   installationNamespace,
			ClusterName: clusterName,
			Labels:      metautils.TranslatedObjectLabels(),
		},
		Spec: securityv1beta1spec.AuthorizationPolicy{},
	}
}

func ingressGatewayAuthPolicyName(ingressGateway *commonv1.AppliedIngressGateway) string {
	return IngressGatewayAuthPolicyName + "-" + sets.Key(ingressGateway.GetDestinationRef())
}
