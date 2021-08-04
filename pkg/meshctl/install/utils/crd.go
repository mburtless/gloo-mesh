package utils

import (
	"bytes"
	"context"

	"github.com/solo-io/gloo-mesh/pkg/meshctl/install/helm"
	"github.com/solo-io/skv2/pkg/crdutils"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	kubeyaml "k8s.io/apimachinery/pkg/util/yaml"
)

func GetCrdMetadataFromManifests(deploy string, manifests []byte) (*crdutils.CRDMetadata, error) {
	decoder := kubeyaml.NewYAMLOrJSONDecoder(bytes.NewBuffer(manifests), 4096)
	for {
		obj := &unstructured.Unstructured{}
		err := decoder.Decode(obj)
		if err != nil {
			return nil, err
		}
		if obj.GetName() != deploy {
			continue
		}
		if obj.GetKind() == "Deployment" {
			annotations := obj.GetAnnotations()
			return crdutils.ParseCRDMetadataFromAnnotations(annotations)
		}
	}
}

// Retrieves the CRD metadata for the chart that's about to be installed.
func GetCrdMetadataFromInstaller(ctx context.Context, deploy string, installer *helm.Installer) (*crdutils.CRDMetadata, error) {
	// take the generated manifest:
	renderedManifests, err := installer.GetRenderedManifests(ctx)
	if err != nil {
		return nil, err
	}
	return GetCrdMetadataFromManifests(deploy, renderedManifests)
}
