package checks_test

import (
	"context"
	"os"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	mock_appsv1 "github.com/solo-io/external-apis/pkg/api/k8s/apps/v1/mocks"
	mock_corev1 "github.com/solo-io/external-apis/pkg/api/k8s/core/v1/mocks"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/validation"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/validation/checks"
	mock_checks "github.com/solo-io/gloo-mesh/pkg/meshctl/validation/checks/mocks"
	"github.com/solo-io/go-utils/grpcutils"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("Agent Checks", func() {

	var (
		ctx             context.Context
		ctrl            *gomock.Controller
		mockRelayDialer *mock_checks.MockRelayDialer

		mockCoreClientset *mock_corev1.MockClientset
		mockAppsClientset *mock_appsv1.MockClientset

		mockDeploymentClient *mock_appsv1.MockDeploymentClient
		mockSecretClient     *mock_corev1.MockSecretClient

		validClientCertSecret = &corev1.Secret{
			Data: map[string][]byte{
				"ca.crt":  []byte("-----BEGIN CERTIFICATE-----\nMIIFTzCCAzegAwIBAgIUMTk4dspR5TTCymDu3CI+iVFbtXkwDQYJKoZIhvcNAQEL\nBQAwKzEUMBIGA1UEAwwLKi5nbG9vLW1lc2gxEzARBgNVBAoMCnJlbGF5LXJvb3Qw\nHhcNMjEwNzIyMTYxNzA3WhcNMzEwNzIwMTYxNzA3WjArMRQwEgYDVQQDDAsqLmds\nb28tbWVzaDETMBEGA1UECgwKcmVsYXktcm9vdDCCAiIwDQYJKoZIhvcNAQEBBQAD\nggIPADCCAgoCggIBAMzYVOnZSUYWG+Ysd7PRNid1sx+EtMeqy3XSVoibDH4VfeAg\nVlxb+cIjE2dr1COwDnHwMW43Xvrx+fJRD5BhZPt3lzKIyzxDw/irGzaANtbLbKQf\nUWSiXULsj6Fdx2LwC+trM/M8NO7blWpQEPgyRcy2XtH+wWgXw61VwKAMTNdTgytD\niM1Yn0b6VA+xn+bbmGMRHNFp92YY+4FeOPIcQcTpqc+xsPWPpEr+kfmZ3uSXvmEd\nlmAkqpvrTJaUrIUQXwyZa8I1GOchOkLs8SBZLaBm+RZneIkVET5UR2ltOKyOdntM\nm4V/ulK3UrVr/AdfheQMVfiV6QK5Tmfpe9gX58M8BdZ+dNFJQE+y3+3Su292PZUc\nB3P6KjH3C+/fpgwjc2xl5762it+p6r+Xq89VqOXJDJO5xZbMLNZOUhxydZHLntvZ\nNeCZeE7vWoM/DfKzmqaPnFXnoi5hYboH1w/k+avOrDJawxxD1EqKnPK/Ml/9476n\nA6KD5EENQmgtFFlebtH+bg4yAgwO2sM8SJJi9vWNE/jwmwjDzRBhdZBhp9b3SuXN\n1TDDqg8TllyDUimCuUbg+UlQbKdP1ZCGetFHdpKNKkPw5G+Q4S5Emaz8BcIVIqgd\nuVqSS0nE1CveN9HQ+sq99TdJFEA+ioayWMjl/IFcBpmrQuwxJCV9OS0kERYDAgMB\nAAGjazBpMB0GA1UdDgQWBBQu37nyof7dzdvdqO/+Kkiqm2rsCDAfBgNVHSMEGDAW\ngBQu37nyof7dzdvdqO/+Kkiqm2rsCDAPBgNVHRMBAf8EBTADAQH/MBYGA1UdEQQP\nMA2CCyouZ2xvby1tZXNoMA0GCSqGSIb3DQEBCwUAA4ICAQAGFEXw4UKMLsTRZWKN\nDW8ts2gClbVTWxLk8YMatDHOCo1XrM4tX0FQCO11d1guCaO1k6UoJJ+LfH0bYCpY\nZbFQGfEO167AmecBZFhZK6+aFP9UKOO3CODcAsMTLg5TBIKNE8gn4ppyJ/fnPtaj\nbRw0yFs9bMhpGsAIs/wH/3D0ld8sFjAYxt3Y6FHchljAOFREgmXUgbSZxc/bfRvG\nnWokM/eTPtq6ZKRLtZzX9CgnntWcPU38tPPndC+DQPNsFUm4i6/aJ6Wt5GFznl+x\n2txIusOx+cdQhkRqXJ2hZFF/xszZN2wPWGHOz0wr7B4k20Y1EAK1tiVFHSI2TtSJ\nwM60yXuK2dwtg7rV8DNWCmU0+HyqCWwcvvz2WGCvH/v2S/z8J4gox8IdkxAtxKXP\nWMt31ZsSdpwrcB0UN5zpMkFzAV7iMOHCuMxx7E0ieqaD5WdDOIqGumt9q0vIJbw2\nS20spvz+U2j6XbBz4YPzrfB7BqA2mPMzAd1Nt7/ZW44V3Oq4uN+cs3ZtV94wBExB\nwd+JK1NyxTpHrOP+ZOyOTZsdfl34/sOpyprX0YCX2iTeBoxZ/FP9VZMzSm5ZXGPb\nI8C0srZyOFL2VGEotboatO/1zfPuz3VtkRdJJflzSIy7dFHDu/9kuEYCFz2TXWE3\n1sQZPZprvxx9hVgIH3cbqHXCfQ==\n-----END CERTIFICATE-----\n"),
				"tls.crt": []byte("-----BEGIN CERTIFICATE-----\nMIIDMjCCAhqgAwIBAgIQa00pgMw18k7pg5I/SvqafDANBgkqhkiG9w0BAQsFADAy\nMRQwEgYDVQQDDAsqLmdsb28tbWVzaDEaMBgGA1UECgwRcmVsYXktdGxzLXNpZ25p\nbmcwHhcNMjEwNzIyMTYyMTA5WhcNMjIwNzIyMTYyMTA5WjAZMRcwFQYDVQQDEw5y\nZW1vdGUtY2x1c3RlcjCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBAKdJ\nxkidxF1iq9omTky5qc5ho5tgTuxZFY2yZGwIoGv+igpkuw0LTcCvarOI5dqF57/h\nOAPx4qVOHUKlDyxHWlSH02VYueALrgNxcsP9sFzliiQARhYCdSllTaZEww/adMF6\nwQJEQ+ffX0kkSTCh1EyL01MDU+z44hsXTsnSUA5ZAj8K7Wm78l21xFgsCqz5UQox\ncEHTqW6lTE2UPtZP+SF+YQHXQ8iThYDRWJdpjBOBBGCWsPILqzfQEcVW3UcxQH0Y\n9Ai1lztrOtvmJNCnekVKfW7Q01srFjCY8mLhQq3rmIDXztPXR3Jx8nM+IOd8ybUJ\nDvG46z0eIbmz4oHJD/MCAwEAAaNdMFswDgYDVR0PAQH/BAQDAgWgMB0GA1UdJQQW\nMBQGCCsGAQUFBwMBBggrBgEFBQcDAjAMBgNVHRMBAf8EAjAAMBwGA1UdEQEB/wQS\nMBCCDnJlbW90ZS1jbHVzdGVyMA0GCSqGSIb3DQEBCwUAA4IBAQAztvBT23uIqqXt\nUXIPnM+1rOm7ecyt9mXEnqXBAR4pm0WEwQV3Bx/OL/pqh5ubGxNrZRqVfVd80otN\nvV8FvD0KrM5kIUcfJ+cgG30xcL+8FwCC1tVVYBsDNl8PKtnrlq+bYeLBsyFgHuag\nb0rl5zqvM/mFsAY0u+Pnvfsuqq9JxFkDB7yUOe+q0xecztV0e+1LFXK2uP5mBoZK\noovUiQqJK9PZ2nuk/eQVTiUtMWQZDbhutROEBvE14IwJic+6Nbtuey2r+vah4U2p\na0eGpyYu5nedlqMPyQ7RbayecjFe+i4AxaxLgPn/LzsalghNWV0qGXE6/9KkkqOb\nGT7II8Fa\n-----END CERTIFICATE-----\n\n-----BEGIN CERTIFICATE-----\nMIIELzCCAhegAwIBAgIBADANBgkqhkiG9w0BAQsFADArMRQwEgYDVQQDDAsqLmds\nb28tbWVzaDETMBEGA1UECgwKcmVsYXktcm9vdDAeFw0yMTA3MjIxNjE3MDhaFw0z\nMTA3MjAxNjE3MDhaMDIxFDASBgNVBAMMCyouZ2xvby1tZXNoMRowGAYDVQQKDBFy\nZWxheS10bHMtc2lnbmluZzCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEB\nAMkxqOhc9TFOTSu7CpJZzBoX7rUFD2fp914sEmYbGI+RbkEQaznGvdK5XeHvVJu5\n7uz6r/1bjbriVuk/25UY4mYMgn8DaWDH435DdzKa6N7j2l3eWlvGUcDwa2qSAFZn\nOTBa7q/CVJiHag1Avh87r+/uPl8a7DSlXGAn+F8b09wABA6IWbuoN3CHKCSTFmMb\nQR6NlDQRuUIKUaVFDMFIW4GGWu6ZW+lDLoX8jGh4Zmd8qcCLoqoy0j70daN3pAfj\nAzI9+qJj1TpYZ9ZAbuR8EBBntm/BLuxrG2AjgqwgGwHL8CZJ0XPTpnh3ws2eGJPC\n2fJqE11o23k+WZvZYaovPXsCAwEAAaNXMFUwDwYDVR0TAQH/BAUwAwEB/zALBgNV\nHQ8EBAMCAuQwHQYDVR0lBBYwFAYIKwYBBQUHAwIGCCsGAQUFBwMBMBYGA1UdEQQP\nMA2CCyouZ2xvby1tZXNoMA0GCSqGSIb3DQEBCwUAA4ICAQAkOSH3W/3swy570V/v\nHPD8AYETXfEyfmOgOvnTbxutBDYW+DQ6heQDxWTkWNnp3s433AsBRckvQGwyV5nJ\nljEFz0WQDKwqjySfJ0el2eBste5PSeFsFv21JYEFfwOFS2KGUevDZ0e4kD4yAqyU\nigovcb+TAhRDnNNgdCgDxVWlA3Nq2IrsDbvZrMD47CnKKyApFSmlbKWF5lzSoLLD\nGerHnLFpzOp99EGoeube48Jk8gtSjIwggzWCntrO4fu1nsEJPP10HpAA3lxaetJ5\nwSYPf0eTTdvEGktJpQKwm0fBG6qxNA8AVnhyIwPheoFL5OgckrPg/aM7/onxGGgX\nONAH11H5AuhTVzRhRTySkGUiVVhG9/x+Y29mg+pTeu+W+x8p6xzM6rz2BXhazg9Z\noPgRO3X86QL6Ngi+imRrkHBVuK3C2u4qPbyrPBf1Tl0QYAuiJ2xA/fwEOtoMpgJk\n75roqD1BR4SqNmX+Kz+k6YgI5xccQlUQUdjunus9PO2zNrQWcP9JGe+vBU9E3uu6\nJjrRpZYnMl98F3ReVN/dhqTQhgmD2SehOmREmUNG+oH2pDjvx15ILOQqYWTR4P5M\n4DrsyYWmvuG7JV0kxyMeWAcK7OUIhYgpzX6/GsfcYgBRuZaYNtt3Tssyy1aHkuwh\npS47T0UMXYWhGuGV4oQFOJN6eQ==\n-----END CERTIFICATE-----\n\n-----BEGIN CERTIFICATE-----\nMIIFTzCCAzegAwIBAgIUMTk4dspR5TTCymDu3CI+iVFbtXkwDQYJKoZIhvcNAQEL\nBQAwKzEUMBIGA1UEAwwLKi5nbG9vLW1lc2gxEzARBgNVBAoMCnJlbGF5LXJvb3Qw\nHhcNMjEwNzIyMTYxNzA3WhcNMzEwNzIwMTYxNzA3WjArMRQwEgYDVQQDDAsqLmds\nb28tbWVzaDETMBEGA1UECgwKcmVsYXktcm9vdDCCAiIwDQYJKoZIhvcNAQEBBQAD\nggIPADCCAgoCggIBAMzYVOnZSUYWG+Ysd7PRNid1sx+EtMeqy3XSVoibDH4VfeAg\nVlxb+cIjE2dr1COwDnHwMW43Xvrx+fJRD5BhZPt3lzKIyzxDw/irGzaANtbLbKQf\nUWSiXULsj6Fdx2LwC+trM/M8NO7blWpQEPgyRcy2XtH+wWgXw61VwKAMTNdTgytD\niM1Yn0b6VA+xn+bbmGMRHNFp92YY+4FeOPIcQcTpqc+xsPWPpEr+kfmZ3uSXvmEd\nlmAkqpvrTJaUrIUQXwyZa8I1GOchOkLs8SBZLaBm+RZneIkVET5UR2ltOKyOdntM\nm4V/ulK3UrVr/AdfheQMVfiV6QK5Tmfpe9gX58M8BdZ+dNFJQE+y3+3Su292PZUc\nB3P6KjH3C+/fpgwjc2xl5762it+p6r+Xq89VqOXJDJO5xZbMLNZOUhxydZHLntvZ\nNeCZeE7vWoM/DfKzmqaPnFXnoi5hYboH1w/k+avOrDJawxxD1EqKnPK/Ml/9476n\nA6KD5EENQmgtFFlebtH+bg4yAgwO2sM8SJJi9vWNE/jwmwjDzRBhdZBhp9b3SuXN\n1TDDqg8TllyDUimCuUbg+UlQbKdP1ZCGetFHdpKNKkPw5G+Q4S5Emaz8BcIVIqgd\nuVqSS0nE1CveN9HQ+sq99TdJFEA+ioayWMjl/IFcBpmrQuwxJCV9OS0kERYDAgMB\nAAGjazBpMB0GA1UdDgQWBBQu37nyof7dzdvdqO/+Kkiqm2rsCDAfBgNVHSMEGDAW\ngBQu37nyof7dzdvdqO/+Kkiqm2rsCDAPBgNVHRMBAf8EBTADAQH/MBYGA1UdEQQP\nMA2CCyouZ2xvby1tZXNoMA0GCSqGSIb3DQEBCwUAA4ICAQAGFEXw4UKMLsTRZWKN\nDW8ts2gClbVTWxLk8YMatDHOCo1XrM4tX0FQCO11d1guCaO1k6UoJJ+LfH0bYCpY\nZbFQGfEO167AmecBZFhZK6+aFP9UKOO3CODcAsMTLg5TBIKNE8gn4ppyJ/fnPtaj\nbRw0yFs9bMhpGsAIs/wH/3D0ld8sFjAYxt3Y6FHchljAOFREgmXUgbSZxc/bfRvG\nnWokM/eTPtq6ZKRLtZzX9CgnntWcPU38tPPndC+DQPNsFUm4i6/aJ6Wt5GFznl+x\n2txIusOx+cdQhkRqXJ2hZFF/xszZN2wPWGHOz0wr7B4k20Y1EAK1tiVFHSI2TtSJ\nwM60yXuK2dwtg7rV8DNWCmU0+HyqCWwcvvz2WGCvH/v2S/z8J4gox8IdkxAtxKXP\nWMt31ZsSdpwrcB0UN5zpMkFzAV7iMOHCuMxx7E0ieqaD5WdDOIqGumt9q0vIJbw2\nS20spvz+U2j6XbBz4YPzrfB7BqA2mPMzAd1Nt7/ZW44V3Oq4uN+cs3ZtV94wBExB\nwd+JK1NyxTpHrOP+ZOyOTZsdfl34/sOpyprX0YCX2iTeBoxZ/FP9VZMzSm5ZXGPb\nI8C0srZyOFL2VGEotboatO/1zfPuz3VtkRdJJflzSIy7dFHDu/9kuEYCFz2TXWE3\n1sQZPZprvxx9hVgIH3cbqHXCfQ==\n-----END CERTIFICATE-----\n"),
				"tls.key": []byte("-----BEGIN RSA PRIVATE KEY-----\nMIIEowIBAAKCAQEAp0nGSJ3EXWKr2iZOTLmpzmGjm2BO7FkVjbJkbAiga/6KCmS7\nDQtNwK9qs4jl2oXnv+E4A/HipU4dQqUPLEdaVIfTZVi54AuuA3Fyw/2wXOWKJABG\nFgJ1KWVNpkTDD9p0wXrBAkRD599fSSRJMKHUTIvTUwNT7PjiGxdOydJQDlkCPwrt\nabvyXbXEWCwKrPlRCjFwQdOpbqVMTZQ+1k/5IX5hAddDyJOFgNFYl2mME4EEYJaw\n8gurN9ARxVbdRzFAfRj0CLWXO2s62+Yk0Kd6RUp9btDTWysWMJjyYuFCreuYgNfO\n09dHcnHycz4g53zJtQkO8bjrPR4hubPigckP8wIDAQABAoIBAEV0ofjAWCkVsJhQ\nUy4T3+aqL01xfRMwIXzFVmBsbH6qHhIWpHrT+KJJspl7+0LxWbkW/zTUFu/fMNLc\nySHvNhfmlOR9JB9behI/5hBtoe3P97zeyDqXJqHbR5QC9KH+Z391QfF4+mCpI1yw\nzhp89jIZ09GhjhTTeL0avkGZKlfp/IOB7CYr4m39uUJBcrf02rWpRVvySAzZKtEZ\n5F4zWdI0VyBcxrsn3znfyzBN3wPRsFB5+yJK8H1AHr3Q4VNMDBmNTIDJjMHszW6f\nSYbQ8cS8KwEEBfA2AVMt657XbkL//cLokxHyaIFXCXrw8ms5seK8uRj+RqvTpTVf\n702jzBECgYEAwXwPSO1R7XkR+I+WZVEpUory5QEIObHEoRWN20OF/zq6CEfJ6K2y\n4H5es1KnPBIBovAcVMFrLO8Bqo9h/Fqse8rPRDzQ7CS8Ufjq0QNc3EEVlgT200Ty\nnF1WoIO3gE6j7BCUOMmb7Y7pgzK8fo9TRNlAzEsoSO3o/tGbC6s6h78CgYEA3Vbl\ncTDnGzX2ulc6BMiFIqhgqu3Hd1kQFv90nyVarPu6J37zOmcA/JpwVwc3/34TZA2m\nfR3PQ5f2RpCGQ4pDoYNdZ+9cr1iaGsZfrMjiSt5saljei4QEHRX4lru9iNiDVXp/\nQPcU2Rw7UP69Aa/An+568ceOb4W8MTdqx3gMpM0CgYEAjbUOAMyHx5R3nAOWFAh4\nalMICL9TxeWz7IK7zc5Lkp3xaGUjtP2a2B7VvyKXB0Ds3+hZ551toJBAOSogitHi\nKBxm50Rfg8R4BNV5LbH3zf0BEUn7eMqzoeAetRsjR57RIfEWjezi/f9AeW2sbkkM\npI01jyqwi5Frp03e75HuIUMCgYAXGdKolaoJNQCjQidUCHmcvGYacOa4lhsPy2mo\nkoV8OGmdZaqNFeMMejHvY1l82PO9JY+Sz2GqdFnH052vvuaAHO3KwzixNFYhJUMn\nDXBQ0BYQo2XWudiUEI75bG7DsZVDfp15clBCuKeYNH4VhvpbttAuG93J1fNmT5pd\nZzIqoQKBgE5IRKdTvEqeCR1EU4ompmjE4cRCCj1ANCiePntZB+i65VUC02KqiJ12\nSFQ+xqeDufXrSV8gYo3DW2LrdZ+w1NT6t8HSYcRiiX7OKTUZgaRMyowRqPIbjUJm\nkbRLOpKyJ6CToAMOftWKc2QBgjC80J0BS0C2pDSygbTZ/FaMDwnX\n-----END RSA PRIVATE KEY-----\n"),
			},
		}
	)

	BeforeEach(func() {
		// suppress stdout output to make test results easier to read
		os.Stdout, _ = os.Open(os.DevNull)

		ctrl, ctx = gomock.WithContext(context.Background(), GinkgoT())
		mockRelayDialer = mock_checks.NewMockRelayDialer(ctrl)

		mockCoreClientset = mock_corev1.NewMockClientset(ctrl)
		mockAppsClientset = mock_appsv1.NewMockClientset(ctrl)

		mockDeploymentClient = mock_appsv1.NewMockDeploymentClient(ctrl)
		mockSecretClient = mock_corev1.NewMockSecretClient(ctrl)

		mockAppsClientset.EXPECT().Deployments().Return(mockDeploymentClient)
		mockCoreClientset.EXPECT().Secrets().Return(mockSecretClient)
	})

	var buildCheckContext = func(relayDialer checks.RelayDialer) checks.CheckContext {
		checkContext := validation.NewTestCheckContext("gloo-mesh", 0, 0, &checks.AgentParams{
			RelayServerAddress: "relay-server-address",
			RelayAuthority:     "relay-authority",
			RootCertSecretRef: client.ObjectKey{
				Name:      "relay-root-tls-secret",
				Namespace: "gloo-mesh",
			},
			ClientCertSecretRef: client.ObjectKey{
				Name:      "relay-client-tls-secret",
				Namespace: "gloo-mesh",
			},
		}, relayDialer, mockAppsClientset, mockCoreClientset, nil, nil, nil, false, nil)
		return checkContext
	}

	It("pre install relay connectivity check without client cert should succeed", func() {
		checkCtx := buildCheckContext(mockRelayDialer)

		agentParams := checkCtx.Context().AgentParams

		mockSecretClient.EXPECT().GetSecret(ctx, client.ObjectKey{
			Name:      "relay-client-tls-secret",
			Namespace: "gloo-mesh",
		}).Return(nil, errors.NewNotFound(schema.GroupResource{}, ""))

		mockRelayDialer.EXPECT().
			DialIdentity(
				gomock.Any(),
				grpcutils.DialOpts{
					Address:   agentParams.RelayServerAddress,
					Authority: agentParams.RelayAuthority,
				},
				mockSecretClient,
				agentParams.RootCertSecretRef,
			).
			Return(nil)

		Expect(checks.NewRelayConnectivityCheck(true).Run(ctx, checkCtx).IsSuccess()).To(BeTrue())
	})

	It("pre install relay connectivity check with client cert should succeed", func() {
		checkCtx := buildCheckContext(mockRelayDialer)

		mockSecretClient.EXPECT().GetSecret(ctx, client.ObjectKey{
			Name:      "relay-client-tls-secret",
			Namespace: "gloo-mesh",
		}).Return(validClientCertSecret, nil)

		mockRelayDialer.EXPECT().
			DialServer(gomock.Any(), gomock.Any()).
			DoAndReturn(func(
				ctx context.Context,
				relayServerDialOpts grpcutils.DialOpts,
			) {
				Expect(relayServerDialOpts.Address).To(Equal(checkCtx.Context().AgentParams.RelayServerAddress))
				Expect(relayServerDialOpts.Authority).To(Equal(checkCtx.Context().AgentParams.RelayAuthority))
				Expect(relayServerDialOpts.ExtraOptions).To(HaveLen(1))
			}).Return(nil)

		Expect(checks.NewRelayConnectivityCheck(true).Run(ctx, checkCtx).IsSuccess()).To(BeTrue())
	})

	It("pre install relay connectivity check with insecure mode should directly dial relay server without dialing identity server", func() {
		checkCtx := buildCheckContext(mockRelayDialer)

		agentParams := checkCtx.Context().AgentParams
		agentParams.Insecure = true

		mockRelayDialer.EXPECT().
			DialServer(
				gomock.Any(),
				grpcutils.DialOpts{
					Address:   agentParams.RelayServerAddress,
					Authority: agentParams.RelayAuthority,
					Insecure:  true,
				},
			).
			Return(nil)

		Expect(checks.NewRelayConnectivityCheck(true).Run(ctx, checkCtx).IsSuccess()).To(BeTrue())
	})

	It("pre install relay connectivity check should fail for non existent root cert", func() {
		checkCtx := buildCheckContext(mockRelayDialer)

		agentParams := checkCtx.Context().AgentParams

		mockSecretClient.EXPECT().GetSecret(ctx, client.ObjectKey{
			Name:      "relay-client-tls-secret",
			Namespace: "gloo-mesh",
		}).Return(nil, errors.NewNotFound(schema.GroupResource{}, ""))

		mockRelayDialer.EXPECT().
			DialIdentity(
				gomock.Any(),
				grpcutils.DialOpts{
					Address:   agentParams.RelayServerAddress,
					Authority: agentParams.RelayAuthority,
				},
				mockSecretClient,
				agentParams.RootCertSecretRef,
			).
			Return(errors.NewNotFound(schema.GroupResource{}, ""))

		Expect(checks.NewRelayConnectivityCheck(true).Run(ctx, checkCtx).IsFailure()).To(BeTrue())
	})

	It("pre install relay connectivity check should handle dial time out", func() {
		checkCtx := buildCheckContext(mockRelayDialer)

		agentParams := checkCtx.Context().AgentParams

		mockSecretClient.EXPECT().GetSecret(ctx, client.ObjectKey{
			Name:      "relay-client-tls-secret",
			Namespace: "gloo-mesh",
		}).Return(nil, errors.NewNotFound(schema.GroupResource{}, ""))

		mockRelayDialer.EXPECT().
			DialIdentity(
				gomock.Any(),
				grpcutils.DialOpts{
					Address:   agentParams.RelayServerAddress,
					Authority: agentParams.RelayAuthority,
				},
				mockSecretClient,
				agentParams.RootCertSecretRef,
			).
			Return(context.DeadlineExceeded)

		Expect(checks.NewRelayConnectivityCheck(true).Run(ctx, checkCtx).IsFailure()).To(BeTrue())
	})

	It("post install relay connectivity check should succeed", func() {
		checkCtx := buildCheckContext(mockRelayDialer)

		deployment := &v1.Deployment{
			Spec: v1.DeploymentSpec{
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name: "enterprise-agent",
								Args: []string{
									"--metrics-port=9091",
									"--settings-name=settings",
									"--settings-namespace=gloo-mesh",
									"--http-port=9988",
									"--grpc-port=9977",
									"--verbose=true",
									"--relay-address=172.18.0.2:30348",
									"--relay-insecure=false",
									"--relay-authority=enterprise-networking.gloo-mesh",
									"--cluster=remote-cluster",
									"--max-grpc-message-size=4294967295",
									"--relay-client-cert-secret-name=relay-client-tls-secret",
									"--relay-client-cert-secret-namespace=gloo-mesh",
									"--relay-root-cert-secret-name=relay-root-tls-secret",
									"--relay-root-cert-secret-namespace=gloo-mesh",
									"--relay-identity-token-secret-name=relay-identity-token-secret",
									"--relay-identity-token-secret-namespace=gloo-mesh",
									"--relay-identity-token-secret-key=token",
								},
							},
						},
					},
				},
			},
		}

		mockSecretClient.EXPECT().GetSecret(ctx, client.ObjectKey{
			Name:      "relay-client-tls-secret",
			Namespace: "gloo-mesh",
		}).Return(validClientCertSecret, nil)

		mockDeploymentClient.EXPECT().
			GetDeployment(ctx, client.ObjectKey{
				Name:      "enterprise-agent",
				Namespace: "gloo-mesh",
			}).Return(deployment, nil)

		mockRelayDialer.EXPECT().
			DialServer(gomock.Any(), gomock.Any()).
			DoAndReturn(func(
				ctx context.Context,
				relayServerDialOpts grpcutils.DialOpts,
			) {
				Expect(relayServerDialOpts.Address).To(Equal("172.18.0.2:30348"))
				Expect(relayServerDialOpts.Authority).To(Equal("enterprise-networking.gloo-mesh"))
				Expect(relayServerDialOpts.ExtraOptions).To(HaveLen(1))
			}).Return(nil)

		Expect(checks.NewRelayConnectivityCheck(false).Run(ctx, checkCtx).IsSuccess()).To(BeTrue())
	})

	It("post install relay connectivity check with insecure mode should directly dial relay server", func() {
		checkCtx := buildCheckContext(mockRelayDialer)

		deployment := &v1.Deployment{
			Spec: v1.DeploymentSpec{
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name: "enterprise-agent",
								Args: []string{
									"--metrics-port=9091",
									"--settings-name=settings",
									"--settings-namespace=gloo-mesh",
									"--http-port=9988",
									"--grpc-port=9977",
									"--verbose=true",
									"--relay-address=172.18.0.2:30348",
									"--relay-insecure=true",
									"--relay-authority=enterprise-networking.gloo-mesh",
									"--cluster=remote-cluster",
									"--max-grpc-message-size=4294967295",
								},
							},
						},
					},
				},
			},
		}

		mockDeploymentClient.EXPECT().
			GetDeployment(ctx, client.ObjectKey{
				Name:      "enterprise-agent",
				Namespace: "gloo-mesh",
			}).Return(deployment, nil)

		mockRelayDialer.EXPECT().
			DialServer(gomock.Any(), gomock.Any()).
			DoAndReturn(func(
				ctx context.Context,
				relayServerDialOpts grpcutils.DialOpts,
			) {
				Expect(relayServerDialOpts.Address).To(Equal("172.18.0.2:30348"))
				Expect(relayServerDialOpts.Authority).To(Equal("enterprise-networking.gloo-mesh"))
				Expect(relayServerDialOpts.Insecure).To(BeTrue())
			}).Return(nil)

		Expect(checks.NewRelayConnectivityCheck(false).Run(ctx, checkCtx).IsSuccess()).To(BeTrue())
	})
})
