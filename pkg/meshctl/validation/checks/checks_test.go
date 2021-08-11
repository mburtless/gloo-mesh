package checks_test

import (
	"context"
	"os"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	mock_v1beta1 "github.com/solo-io/external-apis/pkg/api/k8s/apiextensions.k8s.io/v1beta1/mocks"
	mock_appsv1 "github.com/solo-io/external-apis/pkg/api/k8s/apps/v1/mocks"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/validation"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/validation/checks"
	"github.com/solo-io/skv2/pkg/crdutils"
	appsv1 "k8s.io/api/apps/v1"
	apiextensions_k8s_io_v1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("Checks", func() {

	var (
		ctx = context.TODO()
	)

	BeforeEach(func() {
		// suppress stdout output to make test results easier to read
		os.Stdout, _ = os.Open(os.DevNull)
	})

	Describe("Server address checks", func() {

		var buildCheckContext = func(relayServerAddress string) checks.CheckContext {
			return validation.NewTestCheckContext(
				"",
				0,
				0,
				&checks.AgentParams{
					RelayServerAddress: relayServerAddress,
				},
				nil,
				nil,
				nil,
				nil,
				nil,
				nil,
				false,
				nil,
			)
		}

		It("should verify relay server addresses of different forms", func() {
			ipv4 := "192.0.2.146"
			ipv6 := "2001:db8:1f70::999:de8:7648:6e8"
			ipv4WithPort := "192.0.2.146:9090"
			ipv6WithPort := "[2001:db8:1f70::999:de8:7648:6e8]:9090"
			dnsname := "relay-server"
			dnsNameWithPort := "relay-server:9090"
			invalidDnsName := "asd#$%"
			invalidIpv4WithScheme := "http://192.0.2.146"
			invalidIpv6 := "[2001:db8:1f70::999:de8:7648:6e8]"

			for _, validAddress := range []string{
				ipv4WithPort,
				ipv4,
				ipv6,
				dnsname,
				ipv6WithPort,
				dnsNameWithPort,
			} {
				checkCtx := buildCheckContext(validAddress)
				check := checks.NewAgentParametersCheck()

				Expect(check.Run(ctx, checkCtx).IsSuccess()).To(BeTrue())
			}

			for _, invalidAddress := range []string{
				invalidDnsName,
				invalidIpv4WithScheme,
				invalidIpv6,
			} {
				checkCtx := buildCheckContext(invalidAddress)
				check := checks.NewAgentParametersCheck()

				Expect(check.Run(ctx, checkCtx).IsFailure()).To(BeTrue())
			}
		})
	})

	Describe("CRD Upgrade Checks", func() {

		var (
			ctx  context.Context
			ctrl *gomock.Controller

			gmInstallNamespace = "gloo-mesh"

			mockAppsClientset *mock_appsv1.MockClientset

			mockCrdClient        *mock_v1beta1.MockCustomResourceDefinitionClient
			mockDeploymentClient *mock_appsv1.MockDeploymentClient

			buildCheckContext = func() checks.CheckContext {
				return validation.NewTestCheckContext(gmInstallNamespace, 0, 0, nil, nil, mockAppsClientset, nil, nil, nil, mockCrdClient, false, nil)
			}
		)

		BeforeEach(func() {
			// suppress stdout output to make test results easier to read
			os.Stdout, _ = os.Open(os.DevNull)

			ctrl, ctx = gomock.WithContext(context.Background(), GinkgoT())

			mockAppsClientset = mock_appsv1.NewMockClientset(ctrl)
			mockCrdClient = mock_v1beta1.NewMockCustomResourceDefinitionClient(ctrl)
			mockDeploymentClient = mock_appsv1.NewMockDeploymentClient(ctrl)

			mockAppsClientset.EXPECT().Deployments().Return(mockDeploymentClient)
		})

		It("Report CRD needs is up to date", func() {
			mockCrdClient.EXPECT().
				ListCustomResourceDefinition(ctx).
				Return(&apiextensions_k8s_io_v1beta1.CustomResourceDefinitionList{
					Items: []apiextensions_k8s_io_v1beta1.CustomResourceDefinition{
						{
							ObjectMeta: metav1.ObjectMeta{
								Name:   "destinations.discovery.mesh.gloo.solo.io",
								Labels: map[string]string{"app": "gloo-mesh"},
								Annotations: map[string]string{
									crdutils.CRDVersionKey:  "1.2.3",
									crdutils.CRDSpecHashKey: "7e30f8d386339cbb",
								},
							},
						},
					},
				}, nil)

			mockDeploymentClient.EXPECT().
				GetDeployment(ctx, client.ObjectKey{
					Namespace: gmInstallNamespace,
					Name:      "enterprise-networking",
				}).Return(
				&appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "enterprise-networking",
						Namespace: "gloo-mesh",
						Annotations: map[string]string{
							crdutils.CRDMetadataKey: `{"crds":[{"name":"destinations.discovery.mesh.gloo.solo.io","hash":"7e30f8d386339cbb"}],"version":"1.1.0"}`,
						},
					},
				}, nil,
			)

			checkCtx := buildCheckContext()

			crdCheck := checks.NewCrdUpgradeCheck("enterprise-networking")
			result := crdCheck.Run(ctx, checkCtx)
			Expect(result.IsSuccess()).To(BeTrue())
		})

		It("Warning if CRD is missing", func() {
			mockCrdClient.EXPECT().
				ListCustomResourceDefinition(ctx).
				Return(&apiextensions_k8s_io_v1beta1.CustomResourceDefinitionList{
					Items: []apiextensions_k8s_io_v1beta1.CustomResourceDefinition{
						{
							ObjectMeta: metav1.ObjectMeta{
								Name:   "destinations.discovery.mesh.gloo.solo.io",
								Labels: map[string]string{"app": "gloo-mesh"},
								Annotations: map[string]string{
									crdutils.CRDVersionKey:  "1.2.3",
									crdutils.CRDSpecHashKey: "7e30f8d386339cbb",
								},
							},
						},
					},
				}, nil)

			mockDeploymentClient.EXPECT().
				GetDeployment(ctx, client.ObjectKey{
					Namespace: gmInstallNamespace,
					Name:      "enterprise-networking",
				}).Return(
				&appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "enterprise-networking",
						Namespace: "gloo-mesh",
						Annotations: map[string]string{
							crdutils.CRDMetadataKey: `{"crds":[{"name":"othercrd","hash":"7e30f8d386339cbb"}],"version":"1.1.0"}`,
						},
					},
				}, nil,
			)

			checkCtx := buildCheckContext()

			crdCheck := checks.NewCrdUpgradeCheck("enterprise-networking")
			result := crdCheck.Run(ctx, checkCtx)
			Expect(result.IsWarning()).To(BeTrue())
			Expect(result.Hints).To(HaveLen(1))
			Expect(result.Hints[0].Hint).To(ContainSubstring("CRD othercrd not present on the cluster, ignore this warning if performing a first time install."))

		})
		It("Error if CRD needs upgrade", func() {
			mockCrdClient.EXPECT().
				ListCustomResourceDefinition(ctx).
				Return(&apiextensions_k8s_io_v1beta1.CustomResourceDefinitionList{
					Items: []apiextensions_k8s_io_v1beta1.CustomResourceDefinition{
						{
							ObjectMeta: metav1.ObjectMeta{
								Name:   "destinations.discovery.mesh.gloo.solo.io",
								Labels: map[string]string{"app": "gloo-mesh"},
								Annotations: map[string]string{
									crdutils.CRDVersionKey:  "1.2.3",
									crdutils.CRDSpecHashKey: "7e30f8d386339cbb",
								},
							},
						},
					},
				}, nil)

			mockDeploymentClient.EXPECT().
				GetDeployment(ctx, client.ObjectKey{
					Namespace: gmInstallNamespace,
					Name:      "enterprise-networking",
				}).Return(
				&appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "enterprise-networking",
						Namespace: "gloo-mesh",
						Annotations: map[string]string{
							crdutils.CRDMetadataKey: `{"crds":[{"name":"destinations.discovery.mesh.gloo.solo.io","hash":"differenthash"}],"version":"1.3.4"}`,
						},
					},
				}, nil)

			checkCtx := buildCheckContext()

			crdCheck := checks.NewCrdUpgradeCheck("enterprise-networking")
			result := crdCheck.Run(ctx, checkCtx)
			Expect(result.IsFailure()).To(BeTrue())
			Expect(result.Errors).To(HaveLen(1))
			Expect(result.Errors[0].Error()).To(ContainSubstring("CRD destinations.discovery.mesh.gloo.solo.io needs to be upgraded"))
			Expect(result.Hints).To(HaveLen(1))
			Expect(result.Hints[0].Hint).To(ContainSubstring("One or more CRD spec has changed. Upgrading your Gloo-Mesh CRDs may be required before continuing."))
		})
	})

})
