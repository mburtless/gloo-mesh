package checks_test

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/validation"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/validation/checks"
	"github.com/solo-io/skv2/pkg/crdutils"
	appsv1 "k8s.io/api/apps/v1"
	apiextensions_k8s_io_v1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
)

var _ = Describe("Checks", func() {

	var (
		ctx = context.TODO()
	)
	Describe("Server address checks", func() {

		buildCheckContext := func(relayServerAddress string) checks.CheckContext {
			return validation.NewTestCheckContext(
				nil,
				"",
				0,
				0,
				&checks.ServerParams{
					RelayServerAddress: relayServerAddress,
				},
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
				Expect(checks.RunChecks(ctx, checkCtx, checks.Server, checks.PreInstall)).To(BeFalse())
			}

			for _, invalidAddress := range []string{
				invalidDnsName,
				invalidIpv4WithScheme,
				invalidIpv6,
			} {
				checkCtx := buildCheckContext(invalidAddress)
				runChecks := checks.RunChecks(ctx, checkCtx, checks.Server, checks.PreInstall)
				Expect(runChecks).To(BeTrue())
			}
		})
	})

	Describe("CRD Upgrade Checks", func() {

		var (
			cli    client.Client
			scheme *runtime.Scheme
		)
		BeforeEach(func() {
			scheme = runtime.NewScheme()
			clientgoscheme.AddToScheme(scheme)
			apiextensions_k8s_io_v1beta1.AddToScheme(scheme)
			cli = fake.NewFakeClientWithScheme(scheme)
		})
		buildCheckContext := func() checks.CheckContext {
			return validation.NewTestCheckContext(
				cli,
				"gloo-mesh",
				0,
				0,
				nil,
				false,
				nil,
			)
		}
		It("Report CRD needs is up to date", func() {

			cli = fake.NewFakeClientWithScheme(scheme, &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "enterprise-networking",
					Namespace: "gloo-mesh",
					Annotations: map[string]string{
						crdutils.CRDMetadataKey: `{"crds":[{"name":"destinations.discovery.mesh.gloo.solo.io","hash":"7e30f8d386339cbb"}],"version":"1.1.0"}`,
					},
				},
			}, &apiextensions_k8s_io_v1beta1.CustomResourceDefinition{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "destinations.discovery.mesh.gloo.solo.io",
					Labels: map[string]string{"app": "gloo-mesh"},
					Annotations: map[string]string{
						crdutils.CRDVersionKey:  "1.2.3",
						crdutils.CRDSpecHashKey: "7e30f8d386339cbb",
					},
				},
			})
			checkCtx := buildCheckContext()

			crdCheck := checks.NewCrdUpgradeCheck("enterprise-networking")
			result := crdCheck.Run(context.Background(), checkCtx)
			Expect(result.IsSuccess()).To(BeTrue())
		})
		It("Error if CRD is missing", func() {

			cli = fake.NewFakeClientWithScheme(scheme, &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "enterprise-networking",
					Namespace: "gloo-mesh",
					Annotations: map[string]string{
						crdutils.CRDMetadataKey: `{"crds":[{"name":"othercrd","hash":"7e30f8d386339cbb"}],"version":"1.1.0"}`,
					},
				},
			}, &apiextensions_k8s_io_v1beta1.CustomResourceDefinition{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "destinations.discovery.mesh.gloo.solo.io",
					Labels: map[string]string{"app": "gloo-mesh"},
					Annotations: map[string]string{
						crdutils.CRDVersionKey:  "1.2.3",
						crdutils.CRDSpecHashKey: "7e30f8d386339cbb",
					},
				},
			})
			checkCtx := buildCheckContext()

			crdCheck := checks.NewCrdUpgradeCheck("enterprise-networking")
			result := crdCheck.Run(context.Background(), checkCtx)
			Expect(result.IsFailure()).To(BeTrue())
			Expect(result.Errors).To(HaveLen(1))
			Expect(result.Errors[0].Error()).To(ContainSubstring("CRD othercrd not found"))
			Expect(result.Hints).To(HaveLen(1))
			Expect(result.Hints[0].Hint).To(ContainSubstring("One or more required CRD were not found on the cluster. Please verify Gloo-Mesh CRDs are installed"))

		})
		It("Error if CRD needs upgrade", func() {

			cli = fake.NewFakeClientWithScheme(scheme, &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "enterprise-networking",
					Namespace: "gloo-mesh",
					Annotations: map[string]string{
						crdutils.CRDMetadataKey: `{"crds":[{"name":"destinations.discovery.mesh.gloo.solo.io","hash":"differenthash"}],"version":"1.3.4"}`,
					},
				},
			}, &apiextensions_k8s_io_v1beta1.CustomResourceDefinition{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "destinations.discovery.mesh.gloo.solo.io",
					Labels: map[string]string{"app": "gloo-mesh"},
					Annotations: map[string]string{
						crdutils.CRDVersionKey:  "1.2.3",
						crdutils.CRDSpecHashKey: "7e30f8d386339cbb",
					},
				},
			})
			checkCtx := buildCheckContext()

			crdCheck := checks.NewCrdUpgradeCheck("enterprise-networking")
			result := crdCheck.Run(context.Background(), checkCtx)
			Expect(result.IsFailure()).To(BeTrue())
			Expect(result.Errors).To(HaveLen(1))
			Expect(result.Errors[0].Error()).To(ContainSubstring("CRD destinations.discovery.mesh.gloo.solo.io needs to be upgraded"))
			Expect(result.Hints).To(HaveLen(1))
			Expect(result.Hints[0].Hint).To(ContainSubstring("One or more CRD spec has changed. Upgrading your Gloo-Mesh CRDs may be required before continuing."))
		})
	})

})
