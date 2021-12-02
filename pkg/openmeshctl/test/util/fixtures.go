package util

import (
	"embed"

	"github.com/ghodss/yaml"
	discoveryv1 "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1"
	networkingv1 "github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	serialize_yaml "k8s.io/apimachinery/pkg/runtime/serializer/yaml"
)

//go:embed testdata
var testdata embed.FS

// DestinationList returns test destinations in a list object.
func DestinationList() *discoveryv1.DestinationList {
	var destList discoveryv1.DestinationList
	readYAML("destinations.yaml", &destList)
	return &destList
}

// Destinations returns test destinations.
func Destinations() []*discoveryv1.Destination {
	destList := DestinationList()
	dests := make([]*discoveryv1.Destination, len(destList.Items))
	for i := range destList.Items {
		dests[i] = &destList.Items[i]
	}

	return dests
}

// DestinationObjects returns test destinations as generic objects.
func DestinationObjects() []runtime.Object {
	destList := DestinationList()
	objs := make([]runtime.Object, len(destList.Items))
	for i := range destList.Items {
		objs[i] = &destList.Items[i]
	}

	return objs
}

// WorkloadList returns test workloads in a list object.
func WorkloadList() *discoveryv1.WorkloadList {
	var wlList discoveryv1.WorkloadList
	readYAML("workloads.yaml", &wlList)
	return &wlList
}

// Workloads returns test workloads.
func Workloads() []*discoveryv1.Workload {
	wlList := WorkloadList()
	wls := make([]*discoveryv1.Workload, len(wlList.Items))
	for i := range wlList.Items {
		wls[i] = &wlList.Items[i]
	}

	return wls
}

// WorkloadObjects returns test workloads as generic objects.
func WorkloadObjects() []runtime.Object {
	wlList := WorkloadList()
	objs := make([]runtime.Object, len(wlList.Items))
	for i := range wlList.Items {
		objs[i] = &wlList.Items[i]
	}
	return objs
}

// MeshList returns test meshes in a list object.
func MeshList() *discoveryv1.MeshList {
	var meshList discoveryv1.MeshList
	readYAML("meshes.yaml", &meshList)
	return &meshList
}

// Meshes returns test meshes.
func Meshes() []*discoveryv1.Mesh {
	meshList := MeshList()
	meshs := make([]*discoveryv1.Mesh, len(meshList.Items))
	for i := range meshList.Items {
		meshs[i] = &meshList.Items[i]
	}

	return meshs
}

// MeshObjects returns test meshes as generic objects.
func MeshObjects() []runtime.Object {
	meshList := MeshList()
	objs := make([]runtime.Object, len(meshList.Items))
	for i := range meshList.Items {
		objs[i] = &meshList.Items[i]
	}

	return objs
}

// VirtualMesh returns a test virtual mesh.
func VirtualMesh() *networkingv1.VirtualMesh {
	var vm networkingv1.VirtualMesh
	readYAML("virtual_mesh.yaml", &vm)
	return &vm
}

// VirtualMeshRaw returns a test virtual mesh as raw bytes.
func VirtualMeshRaw() []byte {
	return readFile("virtual_mesh.yaml")
}

// VirtualMeshUnstructured returns a test virtual mesh as an unstructured data type.
func VirtualMeshUnstructured() *unstructured.Unstructured {
	return readUnstructured("virtual_mesh.yaml")
}

// AccessPolicy returns a test access policy.
func AccessPolicy() *networkingv1.AccessPolicy {
	var ap networkingv1.AccessPolicy
	readYAML("access_policy.yaml", &ap)
	return &ap
}

// AccessPolicyRaw returns a test access policy as raw bytes.
func AccessPolicyRaw() []byte {
	return readFile("access_policy.yaml")
}

// AccessPolicyUnstructured returns a test access policy as an unstructured data type.
func AccessPolicyUnstructured() *unstructured.Unstructured {
	return readUnstructured("access_policy.yaml")
}

// TrafficPolicy returns a test traffic policy.
func TrafficPolicy() *networkingv1.TrafficPolicy {
	var tp networkingv1.TrafficPolicy
	readYAML("traffic_policy.yaml", &tp)
	return &tp
}

// TrafficPolicy returns a test traffic policy in an unstructured data type.
func TrafficPolicyUnstructured() *unstructured.Unstructured {
	return readUnstructured("traffic_policy.yaml")
}

// TrafficPolicyRaw returns a test traffic policy as raw bytes.
func TrafficPolicyRaw() []byte {
	return readFile("traffic_policy.yaml")
}

// BookInfo returns the bookinfo objects.
func BookInfo() []runtime.Object {
	var (
		saList  corev1.ServiceAccountList
		depList appsv1.DeploymentList
		svcList corev1.ServiceList
	)

	readYAML("bookinfo/service_accounts.yaml", &saList)
	readYAML("bookinfo/deployments.yaml", &depList)
	readYAML("bookinfo/services.yaml", &svcList)

	objs := make([]runtime.Object, 0, len(saList.Items)+len(depList.Items)+len(svcList.Items))
	for i := range saList.Items {
		objs = append(objs, &saList.Items[i])
	}
	for i := range depList.Items {
		objs = append(objs, &depList.Items[i])
	}
	for i := range svcList.Items {
		objs = append(objs, &svcList.Items[i])
	}

	return objs
}

func readFile(name string) []byte {
	b, err := testdata.ReadFile("testdata/" + name)
	if err != nil {
		panic(err)
	}

	return b
}

func readYAML(name string, into interface{}) {
	if err := yaml.Unmarshal(readFile(name), into); err != nil {
		panic(err)
	}
}

var decoder = serialize_yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)

func readUnstructured(name string) *unstructured.Unstructured {
	var obj unstructured.Unstructured
	if _, _, err := decoder.Decode(readFile(name), nil, &obj); err != nil {
		panic(err)
	}

	return &obj
}
