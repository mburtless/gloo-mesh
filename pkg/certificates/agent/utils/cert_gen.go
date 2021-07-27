package utils

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"strings"

	"github.com/rotisserie/eris"
	pkiutil "istio.io/istio/security/pkg/pki/util"
)

const (
	rsaKeySize = 4096
)

func GeneratePrivateKey(keySize int) ([]byte, error) {
	if keySize == 0 {
		keySize = rsaKeySize
	}
	priv, err := rsa.GenerateKey(rand.Reader, keySize)
	if err != nil {
		return nil, eris.Errorf("RSA key generation failed (%v)", err)
	}
	privKey := x509.MarshalPKCS1PrivateKey(priv)
	keyBlock := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privKey,
	}
	return pem.EncodeToMemory(keyBlock), nil
}

/*
	The reason for these constants stem from the golang pem package
	https://golang.org/pkg/encoding/pem/#Block

	a pem encoded block has the form:

	-----BEGIN Type-----
	Headers
	base64-encoded Bytes
	-----END Type-----

	The constants below are the BEGIN and END strings to instruct the encoder/decoder how to properly format the data
*/
const (
	certificateRequest = "CERTIFICATE REQUEST"
)

func GenerateCertificateSigningRequest(
	hosts []string,
	org, meshName string,
	privateKey []byte,
) (csr []byte, err error) {

	// Attempt to decode the key from the PEM format, currently only one format is supported (PKCS1)
	block, _ := pem.Decode(privateKey)
	priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, eris.Wrapf(err, "unable to decode private key, currently only supporting PKCS1 encrypted keys")
	}

	template, err := pkiutil.GenCSRTemplate(pkiutil.CertOptions{
		Host:          strings.Join(hosts, ","),
		Org:           org,
		SignerPrivPem: privateKey,
	})
	if err != nil {
		return nil, eris.Wrap(err, "CSR template creation failed")
	}

	// We add the cluster name to the new CSR template subject.
	// This is extremely important to identify this new certificate as different from it's parent,
	// and others it's communicating with.
	if meshName == "" {
		return nil, eris.New("meshName argument is required")
	}
	template.Subject.OrganizationalUnit = append(template.Subject.OrganizationalUnit, meshName)

	csr, err = x509.CreateCertificateRequest(rand.Reader, template, priv)
	if err != nil {
		return nil, eris.Wrap(err, "creating x509 certificate request")
	}

	// Encode the csr to PEM format before returning
	csrBlock := &pem.Block{
		Type:  certificateRequest,
		Bytes: csr,
	}
	csrByt := pem.EncodeToMemory(csrBlock)
	return csrByt, nil
}

/*
	AppendParentCerts appends 2 certs and/or cert chains together.

	child is the child cert/chain.
	parent is the parent cert/chain

	https://github.com/istio/istio/blob/5218a80f97cb61ff4a02989b7d9f8c4fda50780f/security/pkg/pki/util/generate_csr.go#L95

	Certificate chains are necessary to verify the authenticity of a certificate, in this case the authenticity of
	the generated Ca Certificate against the VirtualMesh root cert
*/
func AppendParentCerts(child, parent []byte) []byte {
	var childCopy []byte
	if len(child) > 0 {
		// Copy the input certificate
		childCopy = make([]byte, len(child))
		copy(childCopy, child)
	}
	if len(parent) > 0 {
		if len(childCopy) > 0 {
			// Append a newline after the last cert
			// Certs are very fooey, this is copy pasted from Mesh, plz do not touch
			// Love, eitan
			childCopy = []byte(strings.TrimSuffix(string(childCopy), "\n") + "\n")
		}
		childCopy = append(childCopy, parent...)
	}
	return childCopy
}
