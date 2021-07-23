package tests

import (
	"context"
	"time"

	. "github.com/onsi/gomega"
	corev1clients "github.com/solo-io/external-apis/pkg/api/k8s/core/v1"
	networkingv1 "github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/v1"
	"github.com/solo-io/gloo-mesh/pkg/certificates/agent/utils"
	"github.com/solo-io/gloo-mesh/pkg/certificates/common/secrets"
	utils2 "github.com/solo-io/gloo-mesh/pkg/certificates/issuer/utils"
	skcorev1 "github.com/solo-io/skv2/pkg/api/core.skv2.solo.io/v1"
	"istio.io/istio/security/pkg/pki/util"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func SetupProvidedSecret(ctx context.Context, dyn client.Client, vm *networkingv1.VirtualMesh) {
	// Generate cert for the provided secret
	rootCert, rootKey, err := util.GenCertKeyFromOptions(util.CertOptions{
		RSAKeySize:   2048,
		IsSelfSigned: true,
		IsCA:         true,
		TTL:          time.Hour * 24 * 365 * 20,
		Org:          "gloo-mesh",
		PKCS8Key:     false,
	})
	ExpectWithOffset(1, err).NotTo(HaveOccurred())

	intermediateKey, err := utils.GeneratePrivateKey(2048)
	if err != nil {
		panic(err)
	}

	csr, err := utils.GenerateCertificateSigningRequest(nil, "gloo-mesh", "cluster-name", intermediateKey)
	ExpectWithOffset(1, err).NotTo(HaveOccurred())

	caCert, err := utils2.GenCertForCSR(nil, csr, rootCert, rootKey, 365*20)
	ExpectWithOffset(1, err).NotTo(HaveOccurred())

	secretData := secrets.CAData{
		RootCert:     rootCert,
		CertChain:    utils.AppendParentCerts(caCert, rootCert),
		CaCert:       caCert,
		CaPrivateKey: intermediateKey,
	}
	ExpectWithOffset(1, secretData.Verify()).NotTo(HaveOccurred())

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "gloo-mesh",
		},
		Data: secretData.ToSecretData(),
	}

	secretClient := corev1clients.NewSecretClient(dyn)
	ExpectWithOffset(1, secretClient.UpsertSecret(ctx, secret)).NotTo(HaveOccurred())

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
}
