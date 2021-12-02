package gloo_mesh

import (
	"os"

	"github.com/solo-io/gloo-mesh/pkg/test/apps/context"
	"istio.io/istio/pkg/test/framework/resource"
)

type Config struct {
	ClusterKubeConfigs                  map[string]string
	DeployControlPlaneToManagementPlane bool
}

const glooMeshVersion = "1.1.0-beta13"

func Deploy(deploymentCtx *context.DeploymentContext, cfg *Config) resource.SetupFn {
	return func(ctx resource.Context) error {
		if deploymentCtx == nil {
			*deploymentCtx = context.DeploymentContext{}
		}
		deploymentCtx.Meshes = []context.GlooMeshInstance{}
		var i context.GlooMeshInstance
		var err error

		version := os.Getenv("GLOO_MESH_VERSION")
		if version == "" {
			version = glooMeshVersion
		}
		// install management plane
		// we need the MP to always be installed and added to the instance list first because
		// istio uninstalls in reverse order meaning control planes unregister first before uninstalling MP
		mpcfg := InstanceConfig{
			managementPlane:               true,
			managementPlaneKubeConfigPath: cfg.ClusterKubeConfigs[ctx.Clusters()[0].Name()],
			version:                       version,
			clusterDomain:                 "",
			cluster:                       ctx.Clusters()[0],
		}
		mpInstance, err := newInstance(ctx, mpcfg)
		if err != nil {
			return err
		}
		deploymentCtx.Meshes = append(deploymentCtx.Meshes, mpInstance)

		// install control planes
		var index = 0
		for n, p := range cfg.ClusterKubeConfigs {
			if n == ctx.Clusters()[index].Name() && !cfg.DeployControlPlaneToManagementPlane {
				// skip deploying CP to MP
				index++
				continue
			}

			cpcfg := InstanceConfig{
				managementPlane:               false,
				controlPlaneKubeConfigPath:    p,
				managementPlaneKubeConfigPath: cfg.ClusterKubeConfigs[ctx.Clusters()[0].Name()],
				version:                       version,
				cluster:                       ctx.Clusters()[index],
				clusterDomain:                 "",
			}
			i, err = newInstance(ctx, cpcfg)
			if err != nil {
				return err
			}
			deploymentCtx.Meshes = append(deploymentCtx.Meshes, i)
			index++
		}
		return nil
	}
}
