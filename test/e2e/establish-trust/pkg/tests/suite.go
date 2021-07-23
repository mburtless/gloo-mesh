package tests

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/solo-io/gloo-mesh/test/e2e/istio/pkg/tests"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/gloo-mesh/test/data"
	. "github.com/solo-io/gloo-mesh/test/e2e"
	"github.com/solo-io/gloo-mesh/test/utils"
	v1 "github.com/solo-io/skv2/pkg/api/core.skv2.solo.io/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	err error
)

const (
	establishTrustTestType = "ESTABLISH_TRUST_TEST_TYPE"
)

// Before running tests, federate the two clusters by creating a VirtualMesh with mTLS enabled.
func SetupClustersAndFederation(customDeployFuc func()) {
	tests.VirtualMeshManifest, err = utils.NewManifest("virtualmesh.yaml")
	Expect(err).NotTo(HaveOccurred())

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Minute)
	defer cancel()

	if customDeployFuc != nil {
		os.Setenv("SKIP_DEPLOY_FROM_SOURCE", "1")
	}
	/* env := */ StartEnvOnce(ctx)

	if customDeployFuc != nil {
		customDeployFuc()
	}

	dynamicClient, err := client.New(GetEnv().Management.Config, client.Options{})
	Expect(err).NotTo(HaveOccurred())

	vm, err := data.SelfSignedVirtualMesh(
		dynamicClient,
		"bookinfo-federation",
		tests.BookinfoNamespace,
		[]*v1.ObjectRef{
			tests.MgmtMesh,
			tests.RemoteMesh,
		},
		false,
	)
	Expect(err).NotTo(HaveOccurred())

	switch os.Getenv(establishTrustTestType) {
	case "provided-ca":
		SetupProvidedSecret(ctx, dynamicClient, vm)
	default:
		Fail(fmt.Sprintf("Must provide a value for %s", establishTrustTestType))
	}

	// wait 5 minutes for Gloo Mesh to initialize and federate traffic across clusters
	tests.FederateClusters(vm, 5)
}

func TeardownFederationAndClusters() {
	err = tests.VirtualMeshManifest.KubeDelete(tests.BookinfoNamespace)
	if err != nil {
		// this is expected to fail in gloo-mesh-enterprise-helm tests as they run the rbac webhook which disables ability to delete this manifest
		log.Printf("warn: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	if os.Getenv("NO_CLEANUP") != "" {
		return
	}
	_ = ClearEnv(ctx)
}

// initialize all tests in suite
// should be called from init() function or top level var
func InitializeTests() bool {
	var (
		_ = Describe("Federation", tests.FederationTest)
	)
	return true
}
