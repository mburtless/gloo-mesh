package version

import (
	"encoding/json"
	"fmt"
	"os"

	appsv1 "github.com/solo-io/external-apis/pkg/api/k8s/apps/v1"
	"github.com/solo-io/gloo-mesh/pkg/common/version"
	"github.com/solo-io/gloo-mesh/pkg/mesh-discovery/utils/dockerutils"
	"github.com/solo-io/gloo-mesh/pkg/openmeshctl/runtime"
	"github.com/spf13/cobra"
	v1 "k8s.io/api/apps/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Command returns a new version command to add to the tree.
func Command(ctx runtime.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "version",
		Short:        "Display the version of meshctl and installed Gloo Mesh components",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return PrintVersion(ctx)
		},
	}

	return cmd
}

type versionInfo struct {
	Client clientVersion   `json:"client"`
	Server []ServerVersion `json:"server"`
}
type clientVersion struct {
	Version string `json:"version"`
}
type ServerVersion struct {
	Namespace  string      `json:"Namespace"`
	Components []Component `json:"components"`
}
type Component struct {
	ComponentName string           `json:"componentName"`
	Images        []componentImage `json:"images"`
}
type componentImage struct {
	Name    string `json:"name"`
	Domain  string `json:"domain"`
	Path    string `json:"path"`
	Version string `json:"version"`
}

func PrintVersion(ctx runtime.Context) error {
	serverVersions := MakeServerVersions(ctx)
	versions := versionInfo{
		Client: clientVersion{Version: version.Version},
		Server: serverVersions,
	}

	bytes, err := json.MarshalIndent(versions, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(bytes))
	return nil
}

func MakeServerVersions(ctx runtime.Context) []ServerVersion {
	kubeClient, err := ctx.KubeClient()
	if err != nil {
		return nil
	}
	deploymentClient := appsv1.NewDeploymentClient(kubeClient)
	deployments, err := deploymentClient.ListDeployment(ctx, &client.ListOptions{Namespace: ctx.Namespace()})
	if err != nil {
		return nil
	}

	// map of Namespace to list of components
	componentMap := make(map[string][]Component)
	for _, deployment := range deployments.Items {
		images, err := getImages(&deployment)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to pull image information for %s: %s\n", deployment.Name, err.Error())
			continue
		}
		if len(images) == 0 {
			continue
		}

		namespace := deployment.GetObjectMeta().GetNamespace()
		componentMap[namespace] = append(
			componentMap[namespace],
			Component{
				ComponentName: deployment.GetName(),
				Images:        images,
			},
		)
	}

	// convert to output format
	var serverVersions []ServerVersion
	for namespace, components := range componentMap {
		serverVersions = append(serverVersions, ServerVersion{Namespace: namespace, Components: components})
	}

	return serverVersions
}

func getImages(deployment *v1.Deployment) ([]componentImage, error) {
	images := make([]componentImage, len(deployment.Spec.Template.Spec.Containers))
	for i, container := range deployment.Spec.Template.Spec.Containers {
		parsedImage, err := dockerutils.ParseImageName(container.Image)
		if err != nil {
			return nil, err
		}
		imageVersion := parsedImage.Tag
		if parsedImage.Digest != "" {
			imageVersion = parsedImage.Digest
		}

		images[i] = componentImage{
			Name:    container.Name,
			Domain:  parsedImage.Domain,
			Path:    parsedImage.Path,
			Version: imageVersion,
		}
	}

	return images, nil
}
