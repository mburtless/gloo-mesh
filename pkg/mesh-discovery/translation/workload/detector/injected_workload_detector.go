package detector

import (
	v1 "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1"
	v1sets "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1/sets"
	"github.com/solo-io/gloo-mesh/pkg/mesh-discovery/translation/workload/types"
)

//go:generate mockgen -source ./injected_workload_detector.go -destination mocks/injected_workload_detector.go

// a workload detector detects injected Mesh workloads
type InjectedWorkloadDetector interface {
	// returns a ref to a mesh if the provided workload will be injected by that mesh
	DetectMeshForWorkload(workload types.Workload, meshes v1sets.MeshSet) *v1.Mesh
}
