package translation_test

import (
	"context"
	"crypto/x509"
	"encoding/pem"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	corev1clients "github.com/solo-io/external-apis/pkg/api/k8s/core/v1/mocks"
	certificatesv1 "github.com/solo-io/gloo-mesh/pkg/api/certificates.mesh.gloo.solo.io/v1"
	"github.com/solo-io/gloo-mesh/pkg/certificates/common/secrets"
	"github.com/solo-io/gloo-mesh/pkg/certificates/issuer/translation"
	"github.com/solo-io/skv2/pkg/ezkube"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("CertIssueTranslator", func() {
	var (
		ctrl *gomock.Controller
		ctx  context.Context

		mockSecretClient *corev1clients.MockSecretClient
	)

	BeforeEach(func() {
		ctrl, ctx = gomock.WithContext(context.Background(), GinkgoT())
		mockSecretClient = corev1clients.NewMockSecretClient(ctrl)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("will not process an issuedCert with no signing cert", func() {
		translator := translation.NewTranslator(mockSecretClient)

		issuedCert := &certificatesv1.IssuedCertificate{
			Spec: certificatesv1.IssuedCertificateSpec{
				CertificateAuthority: &certificatesv1.IssuedCertificateSpec_GlooMeshCa{
					GlooMeshCa: &certificatesv1.RootCertificateAuthority{
						CertificateAuthority: &certificatesv1.RootCertificateAuthority_SigningCertificateSecret{
							SigningCertificateSecret: nil,
						},
					},
				},
			},
		}

		output, err := translator.Translate(ctx, &certificatesv1.CertificateRequest{}, issuedCert)
		Expect(err).NotTo(HaveOccurred())
		Expect(output).To(BeNil())
	})

	DescribeTable(
		"It will sign a CSR when signing cert is present",
		func(issuedCertFactory func(secret *corev1.Secret) *certificatesv1.IssuedCertificate) {

			translator := translation.NewTranslator(mockSecretClient)

			rootCaData := &secrets.CAData{
				CaPrivateKey: []byte(caKey),
				CaCert:       []byte(caCert),
				RootCert:     []byte(caCert),
				CertChain:    []byte(caCert),
			}

			secret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "i'm a",
					Namespace: "secret",
				},
				Data: rootCaData.ToSecretData(),
			}

			certRequest := &certificatesv1.CertificateRequest{
				Spec: certificatesv1.CertificateRequestSpec{
					CertificateSigningRequest: []byte(csr),
				},
			}

			issuedCert := issuedCertFactory(secret)

			mockSecretClient.EXPECT().
				GetSecret(gomock.Any(), ezkube.MakeClientObjectKey(secret)).
				Return(secret, nil)

			output, err := translator.Translate(ctx, certRequest, issuedCert)
			Expect(err).NotTo(HaveOccurred())

			translatedCertByt, _ := pem.Decode(output.SignedCertificate)
			translatedCert, err := x509.ParseCertificate(translatedCertByt.Bytes)
			Expect(err).NotTo(HaveOccurred())
			Expect(translatedCert.Subject.OrganizationalUnit).To(ConsistOf("mesh-name"))

			precomputedCertByt, _ := pem.Decode([]byte(outputCert))
			precomputedCert, err := x509.ParseCertificate(precomputedCertByt.Bytes)
			Expect(err).NotTo(HaveOccurred())

			Expect(precomputedCert.Extensions).To(Equal(translatedCert.Extensions))
			Expect(precomputedCert.Issuer).To(Equal(translatedCert.Issuer))
			Expect(precomputedCert.Subject.Organization).To(Equal(translatedCert.Subject.Organization))
			Expect(precomputedCert.Subject.OrganizationalUnit).To(Equal(translatedCert.Subject.OrganizationalUnit))
			Expect(err).NotTo(HaveOccurred())
		},
		Entry("Standard Config", func(secret *corev1.Secret) *certificatesv1.IssuedCertificate {
			return &certificatesv1.IssuedCertificate{
				Spec: certificatesv1.IssuedCertificateSpec{
					CertificateAuthority: &certificatesv1.IssuedCertificateSpec_GlooMeshCa{
						GlooMeshCa: &certificatesv1.RootCertificateAuthority{
							CertificateAuthority: &certificatesv1.RootCertificateAuthority_SigningCertificateSecret{
								SigningCertificateSecret: ezkube.MakeObjectRef(secret),
							},
						},
					},
				},
			}
		}),
		Entry("Deprecated Config", func(secret *corev1.Secret) *certificatesv1.IssuedCertificate {
			return &certificatesv1.IssuedCertificate{
				Spec: certificatesv1.IssuedCertificateSpec{
					SigningCertificateSecret: ezkube.MakeObjectRef(secret),
				},
			}
		}),
	)
})

const (
	caKey = `-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEA1Y0WSk8KvoeS9tIhVzuNJfrQEloydi+ae+f7YNIN8Fak+Mny
lTdzECUQnNaRvWwCMota5mQMqNqu9EGmkgl/KKxzMa7IbQmCY5XS5h+QrkElAx77
HByxEt5DRW+41dxBGTWXt8bpTZ5lOCFNWAIiWLLTMwHg+w1tNJd33gmPVsjobnoQ
HlY0RUR6gZy+QuVymyYlKIpvTC+YYVSR78gKEGvZ1oA4xbVoTabgt2UZOaDJ9W/x
lv4hhMx7h8kK5JBGa0e5jV2Kd5vqUJGNaiBcwMYdmaObW7zSP/xRRbtIb0J4AwA0
ayG7uxLYFoOM1w52ZsAo+9UPrBS6GDCPX3hikwIDAQABAoIBAE5lVAiFieE2LhqU
O48lmoSV1erW1+2RPjo8iIkbs+hGNpvqrzZeO8xyfu3Ey43pZ8kcZYtssUUPuuiK
bVbxS5An9sYHbyawNgDPELRQZDHEUo0Zw3+nfM37cGC+SfGgwPk7Nm5OBHntKyV2
/EjRx8AeLfBswSYI4M1MycFmawulW+mw730HiQlO2e6JSQXIAwDONxkEaweMOEVB
o1ZVWDT0liYDuSH5h5hNXxFMRHAe70WMnsq6GfRQvlGka2oW5EvIRF6nJfv3riqq
eXKvum8WqX0CBYahIiqlh+s3BeG2mw3BJiyVWXzmDh6jb13EkKuCP+gnY44dxcnJ
0bCF5jECgYEA3wbW5Be8x+vLg315Vn3oxE0HCXABNCTQvyROOar7tCeI0M4j645e
UW1a5wDjofAex0TG5mak9oqFS/ChK3R0jxMY8Dy8lIDoQcAg3r+ztTlpAlFRM3pT
n2UOmIY6FFyiX/BMGSRNSOjwT61oA+lW3I/7pwoB4FqN2LMt8vCusukCgYEA9R+d
B85M1QhRjMb2tZJ55pL5UyqC/seoDqYO53KenkMRPvPN8NE0nmKEA9tZqaJHFEYQ
mqlqyr2OT5qIFWwlkLDi/g+RMgcyudny2K7x3JWUEb3R0fhYSExT5X5952D3W/zT
Ejl1tx6frPmBBiTYrXhKmEkQNhSYiLijXN+U5BsCgYEAzy6aV//ZJltcjoT0QC3t
GtZ3kAPVimwc40PFy3qUIrKLPXYSFlQGOFx/EpNX42qeHP0+THDUFBdwZrBd+HFR
ikvyYdH6WXY6zEHAB01MkzCG5VlHNqwPnMYTPguLTrkTOk6PUtfPV8jU3R+4vdF5
GKJE49K/FXzpwoIJUGLX12kCgYBRJeYWb3WAEQDuWe/SrGsuqflgTvKO5gn8z3yf
opJgUlOjQ5Mp5hhFVtfdbwB/5/kf/RICIZP5CkfSkpX6gZLuE6ER+pVWuotQe5ap
pUDshZg/R1fu6whO5vXfQ8DqmG9LRKeboOoXdUvnN7I/FnOk+e23/HghbzAQExAB
7wKbgQKBgQCRrawCRhBO2s8+QsvV3S746o0IR+b1zxztuupkA7Oe7ui2J1X9W7Ry
UT8oobs321PK3hgqT8sP6gPl5G8XNesJsh2FtefUMkSGSeT1/jF9ubXsMc5JsCGf
ygnbMUIDpnsPxCJ7E8Muy8xw8IdH7g4q8FqBz8jdvAtFT5hYnqx08w==
-----END RSA PRIVATE KEY-----
`
	caCert = `-----BEGIN CERTIFICATE-----
MIIDYjCCAkqgAwIBAgIQOPQEaPEyYbFnCzIslYjADTANBgkqhkiG9w0BAQsFADAP
MQ0wCwYDVQQKEwRST09UMB4XDTIxMDcyMTE0MzcwMloXDTQxMDcxNjE0MzcwMlow
FDESMBAGA1UEChMJZ2xvby1tZXNoMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIB
CgKCAQEA1Y0WSk8KvoeS9tIhVzuNJfrQEloydi+ae+f7YNIN8Fak+MnylTdzECUQ
nNaRvWwCMota5mQMqNqu9EGmkgl/KKxzMa7IbQmCY5XS5h+QrkElAx77HByxEt5D
RW+41dxBGTWXt8bpTZ5lOCFNWAIiWLLTMwHg+w1tNJd33gmPVsjobnoQHlY0RUR6
gZy+QuVymyYlKIpvTC+YYVSR78gKEGvZ1oA4xbVoTabgt2UZOaDJ9W/xlv4hhMx7
h8kK5JBGa0e5jV2Kd5vqUJGNaiBcwMYdmaObW7zSP/xRRbtIb0J4AwA0ayG7uxLY
FoOM1w52ZsAo+9UPrBS6GDCPX3hikwIDAQABo4G0MIGxMA4GA1UdDwEB/wQEAwIC
BDAPBgNVHRMBAf8EBTADAQH/MB0GA1UdDgQWBBQMsBklV0mDZvXfKWicfz4c05O4
fjAfBgNVHSMEGDAWgBTDa7HcQdxT5r+Lit2AJa5RpJF0VTBOBgNVHREBAf8ERDBC
hkBzcGlmZmU6Ly9jbHVzdGVyLmxvY2FsL25zL2lzdGlvLXN5c3RlbS9zYS9pc3Rp
b2Qtc2VydmljZS1hY2NvdW50MA0GCSqGSIb3DQEBCwUAA4IBAQA5VZX/C21Un1Rn
6mO+g4xLXni15QuxWKG6jUb+994GLixyDxQ2+0M77p9ixwYUPE+JbQJ/gM9ThFkT
CQ5efsa0KxnWMlLDmjj4Y+n7ucA0blJ9QOoxemA6DlxP9vaKgdpwtJJK4BcWkhR/
XMyeMF2JXBDeHmyKYS/Ctk0IjlukzW2bRVQxZ397dBChlCmGhErqb2SJVVBi2GAM
UNsRUVt8Zata0gp4dQHJjZ4q5hRePym8qeSOPQly7hpbGbUyZyTGymWhgwEcN+L8
D7s0kk4Xd/LRVtUjMO8p/AhAxJlWVomKwYWnRhGkJc6N4iFq9Zkt3hoMRZFBGgNK
DNGJapXp
-----END CERTIFICATE-----
`
	csr = `-----BEGIN CERTIFICATE REQUEST-----
MIIE0zCCArsCAQAwKDESMBAGA1UEChMJZ2xvby1tZXNoMRIwEAYDVQQLEwltZXNo
LW5hbWUwggIiMA0GCSqGSIb3DQEBAQUAA4ICDwAwggIKAoICAQDC0faGHKmGnreC
0Bv64wRjDb4NoUy7AoeS/WY8gXbw5wXRVMmnMrmgiQbVvMHd+6K7R2KYJ0b6W/Zk
TV3RlTFYJYnVEjmalQvhpLI7u7omwqks8zbyOX+1JqbRUaXG/VCgt1O59UUlVx6+
paxok4Wtj1bE8WYmMfSlgf1sNYYniCay9rrY2HyIZCDu44oZLwgGxZng3kcAD/95
MHPaV63O9bp2uuZaGcNYSaUBRvfCqGkwd3oJKfBhQqRIdET9PiDO7i4CmW1Yse43
5K1nnFZh/97aWSTEJ0L68chOXyzZ1axCZlHPuihveK5eI8TsYOmWdAPxgcuANSQt
MMC57VVZi60m7gaqvQmhFdMPcLYp8iAl8r0WbZf0iVItTRdTb2qHX8zb6O+5AsXU
VscpFYUCyL/CNabKTqW7F07GspGNQnqfop1NKw2UGTvmtzqnYihb/oSUkNKmG2PF
PYCeezFm72MXA6FJpq3CrvvAbcKCiBJv8BMIHIFWklIT7p+1ViOCoxmcg6t6/Ptb
HP0vLXrLRj7FMAsI4PYUb8g+eYNhtAM6+rERDX19qGeaPHJnBKU7Uvep14qztPbB
fFhhCaGgZ02DdyalJV045cgB/2blvyyKH4QaDC8MPwCr6+flghMc0+dQ355N22V4
6b86rhgzy5qvU9kkX0xqR/BclGUHBQIDAQABoGYwZAYJKoZIhvcNAQkOMVcwVTBT
BgNVHREBAf8ESTBHhkVzcGlmZmU6Ly9jdXN0b20tZG9tYWluL25zL2lzdGlvLXN5
c3RlbS9zYS9pc3Rpby1waWxvdC1zZXJ2aWNlLWFjY291bnQwDQYJKoZIhvcNAQEL
BQADggIBAIu5anFcAFq74RZpuzBXY13CHC044MVxZoqiEkCTZbzdVRpZuiNoFqtK
sp0ZmhmNnp+E7AVIhUZ8a6rRgkC3Sr23gctc/JW1Ir6IsClly2Nf9brRD9qH6cQw
AGURRx234wCAzsjQj/SZV8WfXIC8A3R+4U7nQWBpDlIZ4gP/FTdQ6vojwCNAbCV1
su2P6A5v12n3plFM52TsB2K9rZlGG7GvY4FOQ6vWPUwMjyLC8AzBrws5UodmZY6x
hbHPUpp/myHKBZLutebVgY92A1PKjyMaJk0GTejhk974tyBL+d0sa1IZ/QErOkO9
6vo3sxH340jmhcCRvMk539hLvi2Enh+BULCTcQxF4GPyyHVH8DMdtH2XvlvK2cL5
NfGzpQXnyNytGAhYWl1tAEAtV7TZ1zPwP4pvkaX4FlTYsOu35dk0PL559kAokudA
qViJytIaMm1+GB2IWrskL93dOodN06FOdLie8sLhF6vbfS4bbtu+sT8VLJ+HjsZ8
bYITEzQ/K3pvrq0LlFnOlFFCBiNthZ6QhMneaMu9mNqAJG8V9MvffQhdjK6yH7DI
u4rg3XJdr2VJ7k8l8QPDSV226XhTjozJ4YjyfA9CHH+W5Ihero8VnCt4h11TSoWG
jM8iBeWq0DXwvTlf3eFlZPfGl/M1t5mNEtijRH0DP9YhTT7xcirp
-----END CERTIFICATE REQUEST-----


`
	outputCert = `-----BEGIN CERTIFICATE-----
MIIEOjCCAyKgAwIBAgIRAMCbgti2hWWkrTL1XwevheUwDQYJKoZIhvcNAQELBQAw
FDESMBAGA1UEChMJZ2xvby1tZXNoMB4XDTIxMDcyMTE1MDEwMloXDTIyMDcyMTE1
MDEwMlowKDESMBAGA1UEChMJZ2xvby1tZXNoMRIwEAYDVQQLEwltZXNoLW5hbWUw
ggIiMA0GCSqGSIb3DQEBAQUAA4ICDwAwggIKAoICAQDC0faGHKmGnreC0Bv64wRj
Db4NoUy7AoeS/WY8gXbw5wXRVMmnMrmgiQbVvMHd+6K7R2KYJ0b6W/ZkTV3RlTFY
JYnVEjmalQvhpLI7u7omwqks8zbyOX+1JqbRUaXG/VCgt1O59UUlVx6+paxok4Wt
j1bE8WYmMfSlgf1sNYYniCay9rrY2HyIZCDu44oZLwgGxZng3kcAD/95MHPaV63O
9bp2uuZaGcNYSaUBRvfCqGkwd3oJKfBhQqRIdET9PiDO7i4CmW1Yse435K1nnFZh
/97aWSTEJ0L68chOXyzZ1axCZlHPuihveK5eI8TsYOmWdAPxgcuANSQtMMC57VVZ
i60m7gaqvQmhFdMPcLYp8iAl8r0WbZf0iVItTRdTb2qHX8zb6O+5AsXUVscpFYUC
yL/CNabKTqW7F07GspGNQnqfop1NKw2UGTvmtzqnYihb/oSUkNKmG2PFPYCeezFm
72MXA6FJpq3CrvvAbcKCiBJv8BMIHIFWklIT7p+1ViOCoxmcg6t6/PtbHP0vLXrL
Rj7FMAsI4PYUb8g+eYNhtAM6+rERDX19qGeaPHJnBKU7Uvep14qztPbBfFhhCaGg
Z02DdyalJV045cgB/2blvyyKH4QaDC8MPwCr6+flghMc0+dQ355N22V46b86rhgz
y5qvU9kkX0xqR/BclGUHBQIDAQABo3MwcTAOBgNVHQ8BAf8EBAMCAgQwDwYDVR0T
AQH/BAUwAwEB/zAdBgNVHQ4EFgQUD8cpu3VlAGuBRoCuUqrPiaRV3mMwHwYDVR0j
BBgwFoAUDLAZJVdJg2b13ylonH8+HNOTuH4wDgYDVR0RAQH/BAQwAoIAMA0GCSqG
SIb3DQEBCwUAA4IBAQA6IxN7cZMIqQfSMRF7k6Hz+Oi+FJ+xoqxMxr/01cyCD1Li
GgPj5PLLOrjZ6HbyLYK3oOssx8o6JXBzz14qLaDUY5QshBuLXZQAagRRqkxlSMqe
kcYYclqah14HhdZz5432c+PBlL315rICUJrZXGH5lb8RNsc9KOjh0+ntKXM5tj31
wraBGqWf5SF0UzhMHyTv0CbJXK0ENHaAJBjwb8XOw2MZs1fnvthRL+1q/IlPYzzU
etm0fvcGElKvDQFg7NVe1qZuFWTLKYkHPGpTS0tVxx17k5QoRiB2rECg//9J9+TW
9HnznasuHeB4oMwBmLbm6wEmBAQIUZW+PscsCvY5
-----END CERTIFICATE-----

`
)
