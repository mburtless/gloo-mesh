package gloo_mesh

import (
	"context"
	"embed"
	"fmt"
	"io"
	"io/fs"
	"time"

	"github.com/solo-io/gloo-mesh/pkg/openmeshctl/commands/install"
	"github.com/solo-io/gloo-mesh/pkg/openmeshctl/commands/register"
	"github.com/solo-io/gloo-mesh/pkg/openmeshctl/commands/uninstall"
	"github.com/solo-io/gloo-mesh/pkg/openmeshctl/runtime"
	gloo_context "github.com/solo-io/gloo-mesh/pkg/test/apps/context"
	"github.com/spf13/pflag"
	"istio.io/istio/pkg/test/framework/components/cluster"
	"istio.io/istio/pkg/test/framework/resource"
	"istio.io/istio/pkg/test/util/retry"
	"istio.io/pkg/log"
	v1 "k8s.io/api/core/v1"
	kubeApiMeta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	//go:embed crds/*.yaml
	crdFiles embed.FS
	_        gloo_context.GlooMeshInstance = &glooMeshInstance{}
	_        io.Closer                     = &glooMeshInstance{}
)

const namespace = "gloo-mesh"

type glooMeshInstance struct {
	id             resource.ID
	instanceConfig InstanceConfig
	ctx            resource.Context
}
type InstanceConfig struct {
	managementPlane               bool
	controlPlaneKubeConfigPath    string
	managementPlaneKubeConfigPath string
	version                       string
	clusterDomain                 string
	cluster                       cluster.Cluster
}

func newInstance(ctx resource.Context, instanceConfig InstanceConfig) (gloo_context.GlooMeshInstance, error) {
	var err error
	i := &glooMeshInstance{
		ctx: ctx,
	}
	i.id = ctx.TrackResource(i)
	i.instanceConfig = instanceConfig
	if i.instanceConfig.managementPlane {
		// deploy enterprise version
		if err := i.deployManagementPlane(); err != nil {
			return nil, err
		}
	} else {
		if err := i.deployControlPlane(); err != nil {
			return nil, err
		}
	}
	return i, err
}

func (i *glooMeshInstance) deployManagementPlane() error {
	flags := pflag.NewFlagSet("test-flags", pflag.PanicOnError)
	ctx := install.NewContext(runtime.DefaultContext(flags))
	ctx.AddToFlags(flags)
	flags.Set("kubeconfig", i.instanceConfig.managementPlaneKubeConfigPath)
	flags.Set("namespace", namespace)
	flags.Set("version", i.instanceConfig.version)
	flags.Set("release-name", fmt.Sprintf("%s-mp", i.instanceConfig.cluster.Name()))
	if err := install.Install(ctx); err != nil {
		return err
	}

	return nil
}

func (i *glooMeshInstance) deployControlPlane() error {
	flags := pflag.NewFlagSet("test-flags", pflag.PanicOnError)
	ctx := register.NewContext(runtime.DefaultContext(flags))
	ctx.AddToFlags(flags)
	flags.Set("kubeconfig", i.instanceConfig.managementPlaneKubeConfigPath)
	flags.Set("namespace", namespace)
	flags.Set("remote-kubeconfig", i.instanceConfig.controlPlaneKubeConfigPath)
	flags.Set("remote-namespace", namespace)
	flags.Set("version", i.instanceConfig.version)
	flags.Set("agent-version", i.instanceConfig.version)
	return register.Register(ctx, i.instanceConfig.cluster.Name(), "")
}

func (i *glooMeshInstance) ID() resource.ID {
	return i.id
}

func (i *glooMeshInstance) GetRelayServerAddress() (string, error) {
	if !i.instanceConfig.managementPlane {
		return "", fmt.Errorf("cluster does not have a management plane")
	}
	svcName := "enterprise-networking"

	svc, err := i.instanceConfig.cluster.CoreV1().Services(namespace).Get(context.TODO(), svcName, kubeApiMeta.GetOptions{})
	if err != nil {
		return "", err
	}

	// This probably wont work in all situations
	return serviceIngressToAddress(svc)
}

func (i *glooMeshInstance) IsManagementPlane() bool {
	return i.instanceConfig.managementPlane
}

func (i *glooMeshInstance) GetKubeConfig() string {
	return i.instanceConfig.managementPlaneKubeConfigPath
}
func (i *glooMeshInstance) GetCluster() cluster.Cluster {
	return i.instanceConfig.cluster
}

// Close implements io.Closer.
func (i *glooMeshInstance) Close() error {
	flags := pflag.NewFlagSet("test-flags", pflag.PanicOnError)
	ctx := uninstall.NewContext(runtime.DefaultContext(flags))
	ctx.AddToFlags(flags)
	flags.Set("kubeconfig", i.instanceConfig.managementPlaneKubeConfigPath)
	flags.Set("namespace", namespace)
	flags.Set("release-name", fmt.Sprintf("%s-mp", i.instanceConfig.cluster.Name()))
	if err := uninstall.Uninstall(ctx); err != nil {
		log.Warn(err)
	}

	// Delete CRDS in cluster
	files, err := crdFiles.ReadDir("crds")
	if err != nil {
		return err
	}
	for _, f := range files {
		file, err := fs.ReadFile(crdFiles, "crds/"+f.Name())
		if err != nil {
			return err
		}
		if err = i.ctx.Config(i.GetCluster()).DeleteYAML("", string(file)); err != nil {
			log.Warn(err)
		}
	}

	return nil
}

func serviceIngressToAddress(svc *v1.Service) (string, error) {

	port := "9900"
	var address string
	ingress := svc.Status.LoadBalancer.Ingress
	if len(ingress) == 0 {
		// Check for user-set external IPs
		externalIPs := svc.Spec.ExternalIPs
		if len(externalIPs) != 0 {
			address = svc.Spec.ExternalIPs[0]
		} else {
			return "", fmt.Errorf("no loadBalancer.ingress status reported for service. Please set an external IP on the service if you are using a non-kubernetes load balancer.")
		}
	} else {
		// If the Ip address is set in the ingress, use that
		if ingress[0].IP != "" {
			address = ingress[0].IP
		} else {
			// Otherwise use the hostname
			address = ingress[0].Hostname
		}
	}
	return fmt.Sprintf("%s:%s", address, port), nil
}

// wait until secrets are created before returning
func (i *glooMeshInstance) waitForSecretsForNamespace(secrets []string, ns string) error {
	for _, s := range secrets {
		if err := retry.UntilSuccess(func() error {

			_, err := i.instanceConfig.cluster.CoreV1().Secrets(ns).Get(context.TODO(), s, kubeApiMeta.GetOptions{})
			if err == nil {
				return nil
			}

			return nil
		}, retry.Timeout(time.Minute)); err != nil {
			return fmt.Errorf("failed to find secret %s %s", s, err.Error())
		}
	}
	return nil
}
