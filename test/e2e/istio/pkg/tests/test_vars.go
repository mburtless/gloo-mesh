package tests

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/solo-io/gloo-mesh/test/extensions"

	v1 "github.com/solo-io/skv2/pkg/api/core.skv2.solo.io/v1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/gloo-mesh/test/e2e"
)

// Shared test vars
var (
	BookinfoNamespace = "bookinfo"

	MgmtClusterName   = "mgmt-cluster"
	RemoteClusterName = "remote-cluster"

	MgmtMesh = &v1.ObjectRef{
		Name:      "istiod-istio-system-mgmt-cluster",
		Namespace: "gloo-mesh",
	}

	RemoteMesh = func() (*v1.ObjectRef, error) {
		var meshName string
		var out bytes.Buffer

		cmd := exec.Command("bash", "-c", "istioctl version")
		cmd.Stdout = &out
		err := cmd.Run()
		if err != nil {
			return nil, err
		}

		outputString := out.String()
		if strings.Contains(outputString, "1.7") {
			// 1.7 is not installed with a revision since it is unsupported
			meshName = "istiod-istio-system-remote-cluster"
		} else {
			// The revision in this name comes from the revision that we specify when installing Istio
			// to the remote cluster (see setup-kind.sh)
			meshName = "istiod-rev-istio-system-remote-cluster"
		}

		return &v1.ObjectRef{
			Name:      meshName,
			Namespace: "gloo-mesh",
		}, nil
	}

	CurlReviews = func() string {
		return CurlFromProductpage("http://reviews:9080/reviews/1")
	}

	CurlHelloServer = func() string {
		return CurlFromProductpage(fmt.Sprintf("http://%v:%v/", extensions.HelloServerHostname, extensions.HelloServerPort))
	}

	CurlRemoteReviews = func(federatedSuffix string) func() string {
		return func() string {
			return CurlFromProductpage(fmt.Sprintf("http://reviews.%v.svc.%v.%s:9080/reviews/1", BookinfoNamespace, RemoteClusterName, federatedSuffix))
		}
	}

	CurlRatings = func() string {
		return CurlFromProductpage("http://ratings:9080/ratings/1")
	}

	CurlUrl = func(url string) func() string {
		return func() string {
			return CurlFromProductpage(url)
		}
	}

	// Public to be used in enterprise
	CurlFromProductpage = func(url string) string {
		env := e2e.GetEnv()
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute/2)
		defer cancel()
		out := env.Management.GetPod(ctx, BookinfoNamespace, "productpage").Curl(ctx, url, "-v")
		GinkgoWriter.Write([]byte(out))
		return out
	}

	CurlGateway = func(hostname, path, body, method string) string {
		out, err := CurlLocal(hostname, path, body, method, "32000")
		Expect(err).NotTo(HaveOccurred())
		return out
	}

	CurlIngressMgmt = func(hostname, path, body, method string) (string, error) {
		return CurlLocal(hostname, path, body, method, "32011")
	}

	CurlIngressRemote = func(hostname, path, body, method string) (string, error) {
		return CurlLocal(hostname, path, body, method, "32010")
	}

	CurlLocal = func(hostname, path, body, method, port string) (string, error) {
		curlCmd := exec.Command("curl", "--connect-timeout", "1", "--max-time", "5", "-H", "Host: "+hostname, "http://localhost:"+port+path, "-v", "-d", body, "-X", method)
		out, err := curlCmd.CombinedOutput()
		if err == nil {
			GinkgoWriter.Write(out)
		}

		return string(out), err
	}
)
