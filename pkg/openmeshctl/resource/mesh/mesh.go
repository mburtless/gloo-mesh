package mesh

import (
	"fmt"

	discoveryv1 "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1"
	"github.com/solo-io/gloo-mesh/pkg/openmeshctl/output"
	"github.com/solo-io/gloo-mesh/pkg/openmeshctl/resource"
	"github.com/solo-io/gloo-mesh/pkg/openmeshctl/resource/apply"
	"github.com/solo-io/gloo-mesh/pkg/openmeshctl/resource/registry"
	"github.com/solo-io/skv2/pkg/ezkube"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func init() {
	registry.RegisterType(resource.Config{
		Name:      "mesh",
		GVK:       discoveryv1.MeshGVK,
		Plural:    "meshes",
		Aliases:   []string{},
		Factory:   Factory{},
		Formatter: Formatter{},
		Applier:   apply.ServerSideApplier(&discoveryv1.MeshGVK),
	})
}

// Factory creates new mesh objects.
type Factory struct{}

// New returns a new mesh.
func (f Factory) New() client.Object {
	return &discoveryv1.Mesh{}
}

// NewList returns a new mesh list.
func (f Factory) NewList() client.ObjectList {
	return &discoveryv1.MeshList{}
}

// Formatter formats mesh objects.
type Formatter struct{}

// ToSummary converts an mesh to a summary representation.
func (f Formatter) ToSummary(obj runtime.Object) *output.Summary {
	mesh := obj.(*discoveryv1.Mesh)
	fieldSet := output.FieldSet{}
	fieldSet.AddField("Installation", f.buildInstallationFieldSet(mesh))
	fieldSet.AddField("VirtualMesh", mesh.Status.GetAppliedVirtualMesh().GetRef())
	virtDests := make([]ezkube.ResourceId, len(mesh.Status.GetAppliedVirtualDestinations()))
	for i, virtDest := range mesh.Status.GetAppliedVirtualDestinations() {
		virtDests[i] = virtDest.GetRef()
	}
	fieldSet.AddField("VirtualDestinations", virtDests)
	return &output.Summary{Meta: mesh, Fields: fieldSet}
}

func (f Formatter) buildInstallationFieldSet(mesh *discoveryv1.Mesh) output.FieldSet {
	fieldSet := output.FieldSet{}
	meshType, meshInstallation := f.meshTypeAndInstallation(mesh)
	fieldSet.AddField("Type", meshType)
	if meshType == "appmesh" {
		appmesh := mesh.Spec.GetAwsAppMesh()
		fieldSet.AddField("Type", "appmesh")
		fieldSet.AddField("Name", appmesh.GetAwsName())
		fieldSet.AddField("Region", appmesh.GetRegion())
		fieldSet.AddField("Account ID", appmesh.GetAwsAccountId())
		fieldSet.AddField("Clusters", appmesh.GetClusters())
	} else if meshInstallation != nil {
		fieldSet.AddField("Version", meshInstallation.GetVersion())
		fieldSet.AddField("Namespace", meshInstallation.GetNamespace())
		fieldSet.AddField("Cluster", meshInstallation.GetCluster())
	}

	return fieldSet
}

// ToTable converts a list of meshes into a table
func (f Formatter) ToTable(objs []runtime.Object, includeNS, wide bool) *output.Table {
	return &output.Table{
		Headers: f.makeHeaders(includeNS, wide),
		NextRow: f.makeNextTableRowFunc(objs, includeNS, wide),
	}
}

func (f Formatter) makeHeaders(includeNS, wide bool) []string {
	headers := make([]string, 0, 8)
	if includeNS {
		headers = append(headers, "namespace")
	}
	headers = append(headers, "name", "type", "virtual mesh", "virtual destinations", "age")
	if wide {
		headers = append(headers, "mesh version", "mesh namespace", "mesh cluster")
	}

	return headers
}

func (f Formatter) makeNextTableRowFunc(objs []runtime.Object, includeNS, wide bool) func() []string {
	ndx := 0
	return func() []string {
		if ndx == len(objs) {
			return nil
		}

		mesh := objs[ndx].(*discoveryv1.Mesh)
		ndx++

		meshType, meshInstallation := f.meshTypeAndInstallation(mesh)
		data := make([]string, 0, 8)
		if includeNS {
			data = append(data, mesh.GetNamespace())
		}
		data = append(
			data,
			mesh.GetName(),
			meshType,
			output.RefToString(mesh.Status.GetAppliedVirtualMesh().GetRef()),
			f.virtualDestinationsValue(mesh),
			output.FormatAge(mesh.GetCreationTimestamp()),
		)
		if wide {
			data = append(
				data,
				meshInstallation.GetVersion(),
				meshInstallation.GetNamespace(),
				meshInstallation.GetCluster(),
			)
		}
		return data
	}
}

func (f Formatter) virtualDestinationsValue(mesh *discoveryv1.Mesh) string {
	virtualDestinations := mesh.Status.GetAppliedVirtualDestinations()
	healthy := 0
	for _, vd := range virtualDestinations {
		if len(vd.Errors) == 0 {
			healthy++
		}
	}

	return fmt.Sprintf("%d/%d", healthy, len(virtualDestinations))
}

func (f Formatter) meshTypeAndInstallation(mesh *discoveryv1.Mesh) (string, *discoveryv1.MeshInstallation) {
	switch mesh.Spec.GetType().(type) {
	case *discoveryv1.MeshSpec_AwsAppMesh_:
		return "appmesh", nil
	case *discoveryv1.MeshSpec_ConsulConnect:
		return "consulconnect", mesh.Spec.GetConsulConnect().GetInstallation()
	case *discoveryv1.MeshSpec_Istio_:
		return "istio", mesh.Spec.GetIstio().GetInstallation()
	case *discoveryv1.MeshSpec_Linkerd:
		return "linkerd", mesh.Spec.GetLinkerd().GetInstallation()
	case *discoveryv1.MeshSpec_Osm:
		return "osm", mesh.Spec.GetOsm().GetInstallation()
	default:
		return "unknown", nil
	}
}
