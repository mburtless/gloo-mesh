package utils

import (
	"crypto"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"strings"
	"time"

	pkiutil "istio.io/istio/security/pkg/pki/util"
)

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
	certificate = "CERTIFICATE"
)

func GenCertForCSR(
	hosts []string, csrPem, signingCert, privateKey []byte, ttlDays uint32,
) ([]byte, error) {

	// Default to 1 year
	ttl := time.Until(time.Now().AddDate(1, 0, 0))
	if ttlDays > 0 {
		ttl = time.Hour * 24 * time.Duration(ttlDays)
	}

	// The following three function calls allow the input byte arrays to be PEM encoded, so that the caller does not
	// need to pre decode the data.
	cert, err := pkiutil.ParsePemEncodedCertificate(signingCert)
	if err != nil {
		return nil, err
	}
	csr, err := pkiutil.ParsePemEncodedCSR(csrPem)
	if err != nil {
		return nil, err
	}
	key, err := pkiutil.ParsePemEncodedKey(privateKey)
	if err != nil {
		return nil, err
	}

	newCertBytes, err := genCertFromCSR(
		csr,
		cert,
		csr.PublicKey,
		key,
		hosts,
		ttl,
		true,
	)
	if err != nil {
		return nil, err
	}
	// This block is the go way to encode the cert into the PEM format before returning it
	block := &pem.Block{
		Type:  certificate,
		Bytes: newCertBytes,
	}
	return pem.EncodeToMemory(block), nil
}

// genCertFromCSR generates a X.509 certificate with the given CSR.
// Copied from https://github.com/istio/istio/blob/b976960cab61b400860f0266dd09c009b31ee5e3/security/pkg/pki/util/generate_cert.go#L223
// We could not use the version in istio because we need access to the certificate template in the function
// genCertTemplateFromCSR below.
// We need access to this because we need to ensure the subject of the new certificate is different from it's parent,
// and we ensure this by adding the mesh name to the OU list in the subject.
func genCertFromCSR(csr *x509.CertificateRequest, signingCert *x509.Certificate, publicKey interface{},
	signingKey crypto.PrivateKey, subjectIDs []string, ttl time.Duration, isCA bool) (cert []byte, err error) {
	tmpl, err := genCertTemplateFromCSR(csr, subjectIDs, ttl, isCA)
	if err != nil {
		return nil, err
	}
	return x509.CreateCertificate(rand.Reader, tmpl, signingCert, publicKey, signingKey)
}

// genCertTemplateFromCSR generates a certificate template with the given CSR.
// The NotBefore value of the cert is set to current time.
// Copied from https://github.com/istio/istio/blob/b976960cab61b400860f0266dd09c009b31ee5e3/security/pkg/pki/util/generate_cert.go#L223
// Added the ability to maintain the subject from the CertificateRequest template.
func genCertTemplateFromCSR(csr *x509.CertificateRequest, subjectIDs []string, ttl time.Duration, isCA bool) (
	*x509.Certificate, error) {
	subjectIDsInString := strings.Join(subjectIDs, ",")
	var keyUsage x509.KeyUsage
	extKeyUsages := []x509.ExtKeyUsage{}
	if isCA {
		// If the cert is a CA cert, the private key is allowed to sign other certificates.
		keyUsage = x509.KeyUsageCertSign
	} else {
		// Otherwise the private key is allowed for digital signature and key encipherment.
		keyUsage = x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment
		// For now, we do not differentiate non-CA certs to be used on client auth or server auth.
		extKeyUsages = append(extKeyUsages, x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth)
	}

	// Build cert extensions with the subjectIDs.
	ext, err := pkiutil.BuildSubjectAltNameExtension(subjectIDsInString)
	if err != nil {
		return nil, err
	}
	exts := []pkix.Extension{*ext}

	// WE NEED TO MAINTAIN THE SUBJECT FROM THE CSR
	// Istio normally clears this here: https://github.com/istio/istio/blob/b976960cab61b400860f0266dd09c009b31ee5e3/security/pkg/pki/util/generate_cert.go#L287
	subject := csr.Subject

	// Dual use mode if common name in CSR is not empty.
	// In this case, set CN as determined by DualUseCommonName(subjectIDsInString).
	if len(csr.Subject.CommonName) != 0 {
		if cn, err := pkiutil.DualUseCommonName(subjectIDsInString); err != nil {
			// log and continue
			//log.Errorf("dual-use failed for cert template - omitting CN (%v)", err)
		} else {
			subject.CommonName = cn
		}
	}

	now := time.Now()

	serialNum, err := genSerialNum()
	if err != nil {
		return nil, err
	}
	// SignatureAlgorithm will use the default algorithm.
	// See https://golang.org/src/crypto/x509/x509.go?s=5131:5158#L1965 .
	return &x509.Certificate{
		SerialNumber:          serialNum,
		Subject:               subject,
		NotBefore:             now,
		NotAfter:              now.Add(ttl),
		KeyUsage:              keyUsage,
		ExtKeyUsage:           extKeyUsages,
		IsCA:                  isCA,
		BasicConstraintsValid: true,
		ExtraExtensions:       exts}, nil
}

func genSerialNum() (*big.Int, error) {
	serialNumLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNum, err := rand.Int(rand.Reader, serialNumLimit)
	if err != nil {
		return nil, fmt.Errorf("serial number generation failure (%v)", err)
	}
	return serialNum, nil
}
