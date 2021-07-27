package utils

import (
	"context"

	corev1clients "github.com/solo-io/external-apis/pkg/api/k8s/core/v1"
	networkingv1 "github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/v1"
	skcorev1 "github.com/solo-io/skv2/pkg/api/core.skv2.solo.io/v1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func SetupProvidedCA(
	ctx context.Context,
	secret *corev1.Secret,
	dyn client.Client,
	vm *networkingv1.VirtualMesh,
) error {
	secretClient := corev1clients.NewSecretClient(dyn)
	if err := secretClient.UpsertSecret(ctx, secret); err != nil {
		return err
	}
	vm.Spec.MtlsConfig.TrustModel = &networkingv1.VirtualMeshSpec_MTLSConfig_Shared{
		Shared: &networkingv1.SharedTrust{
			CertificateAuthority: &networkingv1.SharedTrust_RootCertificateAuthority{
				RootCertificateAuthority: &networkingv1.RootCertificateAuthority{
					CaSource: &networkingv1.RootCertificateAuthority_Secret{
						Secret: &skcorev1.ObjectRef{
							Name:      secret.Name,
							Namespace: secret.Namespace,
						},
					},
				},
			},
		},
	}
	return nil
}
