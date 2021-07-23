package secrets

import (
	"crypto/tls"
	"crypto/x509"

	"github.com/rotisserie/eris"

	"istio.io/istio/security/pkg/pki/util"
)

const (
	/*
		TODO(ilackarms): document the expected structure of secrets (required for VirtualMeshes  using a user-provided root CA)
	*/
	// CaCertID is the CA certificate chain file.
	CaCertID = "ca-cert.pem"
	// CaPrivateKeyID is the private key file of CA.
	CaPrivateKeyID = "ca-key.pem"
	// CertChainID is the ID/name for the certificate chain file.
	CertChainID = "cert-chain.pem"
	// RootCertID is the ID/name for the CA root certificate file.
	RootCertID = "root-cert.pem"
)

// The intermediate CA derived from the root CA of the MeshGroup
type CAData struct {
	RootCert     []byte
	CertChain    []byte
	CaCert       []byte
	CaPrivateKey []byte
}

// Verify the CA Data
// Copied from https://github.com/istio/istio/blob/943ba0765876590d5c6da89d5df034fe0ea0808a/security/pkg/pki/util/keycertbundle.go#L264
func (d CAData) Verify() error {
	// Verify the cert can be verified from the root cert through the cert chain.
	rcp := x509.NewCertPool()
	rcp.AppendCertsFromPEM(d.RootCert)

	icp := x509.NewCertPool()
	icp.AppendCertsFromPEM(d.CertChain)

	opts := x509.VerifyOptions{
		Intermediates: icp,
		Roots:         rcp,
	}
	cert, err := util.ParsePemEncodedCertificate(d.CaCert)
	if err != nil {
		return eris.Wrap(err, "failed to parse cert PEM")
	}
	chains, err := cert.Verify(opts)

	if len(chains) == 0 || err != nil {
		return eris.Wrap(
			err,
			"failed to verify the cert with the provided root chain and cert pool",
		)
	}

	// Verify that the key can be correctly parsed.
	if _, err = util.ParsePemEncodedKey(d.CaPrivateKey); err != nil {
		return eris.Wrap(err, "failed to parse private key PEM")
	}

	// Verify the cert and key match.
	if _, err := tls.X509KeyPair(d.CaCert, d.CaPrivateKey); err != nil {
		return eris.Wrap(err, "the cert does not match the key")
	}

	return nil
}

func (d CAData) ToSecretData() map[string][]byte {
	return map[string][]byte{
		CertChainID:    d.CertChain,
		RootCertID:     d.RootCert,
		CaCertID:       d.CaCert,
		CaPrivateKeyID: d.CaPrivateKey,
	}
}

func CADataFromSecretData(data map[string][]byte) CAData {
	caKey := data[CaPrivateKeyID]
	caCert := data[CaCertID]
	certChain := data[CertChainID]
	rootCert := data[RootCertID]
	return CAData{
		RootCert:     rootCert,
		CertChain:    certChain,
		CaCert:       caCert,
		CaPrivateKey: caKey,
	}
}
