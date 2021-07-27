package data

import (
	"context"
	"time"

	"github.com/solo-io/gloo-mesh/pkg/certificates/agent/utils"
	"github.com/solo-io/gloo-mesh/pkg/certificates/common/secrets"
	utils2 "github.com/solo-io/gloo-mesh/pkg/certificates/issuer/utils"
	"github.com/solo-io/skv2/pkg/ezkube"
	"istio.io/istio/security/pkg/pki/util"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func BuildRootCa(ctx context.Context, secretMeta ezkube.ResourceId) (*corev1.Secret, error) {
	// Generate cert for the provided secret
	rootCert, rootKey, err := util.GenCertKeyFromOptions(util.CertOptions{
		RSAKeySize:   2048,
		IsSelfSigned: true,
		IsCA:         true,
		TTL:          time.Hour * 24 * 365 * 20,
		Org:          "gloo-mesh",
		PKCS8Key:     false,
	})
	if err != nil {
		return nil, err
	}

	intermediateKey, err := utils.GeneratePrivateKey(2048)
	if err != nil {
		panic(err)
	}

	csr, err := utils.GenerateCertificateSigningRequest(nil, "gloo-mesh", "cluster-name", intermediateKey)
	if err != nil {
		return nil, err
	}

	caCert, err := utils2.GenCertForCSR(ctx, nil, csr, rootCert, rootKey, 365*20)
	if err != nil {
		return nil, err
	}

	secretData := secrets.CAData{
		RootCert:     rootCert,
		CertChain:    utils.AppendParentCerts(caCert, rootCert),
		CaCert:       caCert,
		CaPrivateKey: intermediateKey,
	}

	if err := secretData.Verify(); err != nil {
		return nil, err
	}

	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretMeta.GetName(),
			Namespace: secretMeta.GetNamespace(),
		},
		Data: secretData.ToSecretData(),
	}, nil
}
