package validation

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/rotisserie/eris"
	v1 "github.com/solo-io/external-apis/pkg/api/k8s/apps/v1"
	corev1 "github.com/solo-io/external-apis/pkg/api/k8s/core/v1"
	"github.com/solo-io/gloo-mesh/pkg/common/defaults"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/utils"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/validation/checks"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/validation/consts"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/skv2/pkg/crdutils"
	skutils "github.com/solo-io/skv2/pkg/utils"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type InClusterCheckContext struct {
	checks.CommonContext
	crdMetadata map[string]*crdutils.CRDMetadata
}

type OutOfClusterCheckContext struct {
	checks.CommonContext

	mgmtKubeConfig  string
	mgmtKubeContext string
	localPort       uint32
	remotePort      uint32
	crdMetadata     map[string]*crdutils.CRDMetadata
}

func getCrdMetadataFromEnv() (map[string]*crdutils.CRDMetadata, error) {
	// We expect this variable to be set by the k8s downward api
	deploymentTested := os.Getenv(consts.DeploymentTestedDownwardApiEnvVar)
	if deploymentTested == "" {
		return nil, errors.New(consts.DeploymentTestedDownwardApiEnvVar + " env var not set")
	}
	crdJson := os.Getenv(consts.CrdMetadataDownwardApiEnvVar)
	if crdJson == "" {
		return nil, errors.New(consts.CrdMetadataDownwardApiEnvVar + " env var not set")
	}
	var crdMd crdutils.CRDMetadata
	err := json.Unmarshal([]byte(crdJson), &crdMd)
	if err != nil {
		return nil, err
	}
	return map[string]*crdutils.CRDMetadata{
		deploymentTested: &crdMd,
	}, nil
}

func NewInClusterCheckContext() (checks.CheckContext, error) {
	kubeClient, err := utils.BuildClient("", "")
	if err != nil {
		return nil, err
	}
	ns := os.Getenv("POD_NAMESPACE")
	if ns == "" {
		ns, err = skutils.GetInClusterNamesapce()
		if err != nil {
			return nil, err
		}
	}

	var skipChecks bool
	skipChecksEnv := os.Getenv("SKIP_CHECKS")
	if skipChecksEnv == "1" || strings.ToLower(skipChecksEnv) == "true" {
		skipChecks = true
	}

	// get the crd annotations from the downward api:
	ret := &InClusterCheckContext{
		CommonContext: checks.CommonContext{
			Cli: kubeClient,
			Env: checks.Environment{
				AdminPort: defaults.MetricsPort,
				Namespace: ns,
				InCluster: true,
			},
			ServerParams: nil, // TODO pass in install / upgrade parameters, perhaps through CLI var args?
			SkipChecks:   skipChecks,
		},
	}

	// We expect this variable to be set by the k8s downward api
	crdMd, err := getCrdMetadataFromEnv()
	if err != nil {
		return nil, err
	}
	ret.crdMetadata = crdMd

	return ret, nil
}

// exposed for testing, allows injecting mock k8s client
func NewTestCheckContext(
	client client.Client,
	gmInstallationNamespace string,
	localPort, remotePort uint32,
	serverParams *checks.ServerParams,
	ignoreChecks bool,
	crdMd map[string]*crdutils.CRDMetadata,
) checks.CheckContext {
	return &OutOfClusterCheckContext{
		remotePort:  remotePort,
		localPort:   localPort,
		crdMetadata: crdMd,
		CommonContext: checks.CommonContext{
			Cli: client,
			Env: checks.Environment{
				AdminPort: remotePort,
				Namespace: gmInstallationNamespace,
				InCluster: false,
			},
			ServerParams: serverParams,
			SkipChecks:   ignoreChecks,
		},
	}
}

func NewOutOfClusterCheckContext(
	mgmtKubeConfig string,
	mgmtKubeContext string,
	gmInstallationNamespace string,
	localPort, remotePort uint32,
	serverParams *checks.ServerParams,
	ignoreChecks bool,
	crdMd map[string]*crdutils.CRDMetadata,
) (checks.CheckContext, error) {
	kubeClient, err := utils.BuildClient(mgmtKubeConfig, mgmtKubeContext)
	if err != nil {
		return nil, eris.Wrapf(err, "failed to construct kube client from provided kubeconfig")
	}

	return &OutOfClusterCheckContext{
		remotePort:      remotePort,
		localPort:       localPort,
		mgmtKubeConfig:  mgmtKubeConfig,
		mgmtKubeContext: mgmtKubeContext,
		crdMetadata:     crdMd,
		CommonContext: checks.CommonContext{
			Cli: kubeClient,
			Env: checks.Environment{
				AdminPort: remotePort,
				Namespace: gmInstallationNamespace,
				InCluster: false,
			},
			ServerParams: serverParams,
			SkipChecks:   ignoreChecks,
		},
	}, nil

}

func (c *InClusterCheckContext) Context() checks.CommonContext {
	return c.CommonContext
}

func (c *InClusterCheckContext) CRDMetadata(ctx context.Context, deploymentName string) (*crdutils.CRDMetadata, error) {
	return c.crdMetadata[deploymentName], nil
}

func (c *InClusterCheckContext) AccessAdminPort(ctx context.Context, deployment string, op func(ctx context.Context, adminUrl *url.URL) (error, string)) (error, string) {

	// note: the metrics port is not exposed on the service (it should not be, so this is fine).
	// so we need to find the ip of the deployed pod:
	d, err := v1.NewDeploymentClient(c.Cli).GetDeployment(ctx, client.ObjectKey{
		Namespace: c.Env.Namespace,
		Name:      deployment,
	})
	if err != nil {
		if kerrors.IsNotFound(err) {
			return err, "gloo-mesh enterprise deployment not found. Is gloo-mesh installed in this namespace?"
		}
		return err, ""
	}
	selector, err := metav1.LabelSelectorAsSelector(d.Spec.Selector)
	if err != nil {
		return err, ""
	}
	lo := &client.ListOptions{
		Namespace:     c.Env.Namespace,
		LabelSelector: selector,
		Limit:         1,
	}
	podsList, err := corev1.NewPodClient(c.Cli).ListPod(ctx, lo)
	if err != nil {
		return err, "failed listing deployment pods. is gloo-mesh installed?"
	}
	pods := podsList.Items
	if len(pods) == 0 {
		return err, "no pods are available for deployemt. please check your gloo-mesh installation?"
	}
	if podsList.RemainingItemCount != nil && *podsList.RemainingItemCount != 0 {
		contextutils.LoggerFrom(ctx).Info("You have more than one pod for gloo-mesh deployment. This test may not be accurate.")
	}
	pod := pods[0]
	if pod.Status.PodIP == "" {
		return errors.New("no pod ip"), "gloo-mesh pod doesn't have an IP address. This is usually temporary. please wait or check your gloo-mesh installation?"
	}
	adminUrl := &url.URL{
		Scheme: "http",
		Host:   fmt.Sprintf("%v:%v", pod.Status.PodIP, c.Env.AdminPort),
	}

	return op(ctx, adminUrl)
}

func (c *OutOfClusterCheckContext) Context() checks.CommonContext {
	return c.CommonContext
}

func (c *OutOfClusterCheckContext) AccessAdminPort(ctx context.Context, deployment string, op func(ctx context.Context, adminUrl *url.URL) (error, string)) (error, string) {
	portFwdContext, cancelPtFwd := context.WithCancel(ctx)
	defer cancelPtFwd()

	// start port forward to mgmt server stats port
	localPort, err := utils.PortForwardFromDeployment(
		portFwdContext,
		c.mgmtKubeConfig,
		c.mgmtKubeContext,
		deployment,
		c.Env.Namespace,
		fmt.Sprintf("%v", c.localPort),
		fmt.Sprintf("%v", c.remotePort),
	)
	if err != nil {
		return err, fmt.Sprintf("try verifying that `kubectl port-forward -n %v deployment/%v %v:%v` can be run successfully.", c.Env.Namespace, deployment, c.localPort, c.remotePort)
	}
	// request metrics page from mgmt deployment
	adminUrl := &url.URL{
		Scheme: "http",
		Host:   fmt.Sprintf("localhost:%v", localPort),
	}

	return op(portFwdContext, adminUrl)
}

func (c *OutOfClusterCheckContext) CRDMetadata(ctx context.Context, deploymentName string) (*crdutils.CRDMetadata, error) {
	if c.crdMetadata != nil {
		if md, ok := c.crdMetadata[deploymentName]; ok {
			return md, nil
		}
	}
	// if not provided in construction, fetch from the cluster:
	// read the annotations of the deployment
	d, err := v1.NewDeploymentClient(c.Cli).GetDeployment(ctx, client.ObjectKey{
		Namespace: c.Env.Namespace,
		Name:      deploymentName,
	})
	if err != nil {
		return nil, err
	}

	crdMeta, err := crdutils.ParseCRDMetadataFromAnnotations(d.Annotations)
	if err != nil {
		return nil, err
	}
	if crdMeta == nil {
		return nil, eris.New("No crd information found")
	}

	if c.crdMetadata == nil {
		c.crdMetadata = make(map[string]*crdutils.CRDMetadata)
	}

	c.crdMetadata[deploymentName] = crdMeta
	return crdMeta, nil
}
