package workload

import (
	"fmt"
	"sort"
	"strings"

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
		Name:      "workload",
		GVK:       discoveryv1.WorkloadGVK,
		Plural:    "workloads",
		Aliases:   []string{"wl"},
		Factory:   Factory{},
		Formatter: Formatter{},
		Applier:   apply.ServerSideApplier(&discoveryv1.WorkloadGVK),
	})
}

// Factory creates new workload objects.
type Factory struct{}

// New returns a new workload.
func (f Factory) New() client.Object {
	return &discoveryv1.Workload{}
}

// NewList returns a new workload list.
func (f Factory) NewList() client.ObjectList {
	return &discoveryv1.WorkloadList{}
}

// Formatter formats workload objects.
type Formatter struct{}

// ToSummary converts a workload to a summary representation.
func (f Formatter) ToSummary(obj runtime.Object) *output.Summary {
	workload := obj.(*discoveryv1.Workload)
	fieldSet := output.FieldSet{}
	fieldSet.AddField("Kubernetes Workload", f.buildKubernetesWorkloadFieldSet(workload))
	return &output.Summary{Meta: workload, Fields: fieldSet}
}

func (f Formatter) buildKubernetesWorkloadFieldSet(workload *discoveryv1.Workload) output.FieldSet {
	fieldSet := output.FieldSet{}
	fieldSet.AddField("Controller", workload.Spec.GetKubernetes().Controller)
	fieldSet.AddField("Pod Labels", workload.Spec.GetKubernetes().PodLabels)
	fieldSet.AddField("Service Account", workload.Spec.GetKubernetes().ServiceAccountName)
	fieldSet.AddField("Service Dependencies", f.buildServideDepsValue(workload))
	fieldSet.AddField("Destination Hostnames", workload.Status.GetServiceDependencies().GetDestinationHostnames())
	return fieldSet
}

func (f Formatter) buildServideDepsValue(workload *discoveryv1.Workload) []ezkube.ResourceId {
	serviceDeps := workload.Status.GetServiceDependencies().GetAppliedServiceDependencies()
	serviceDepRefs := make([]ezkube.ResourceId, len(serviceDeps))
	for i, serviceDep := range serviceDeps {
		serviceDepRefs[i] = serviceDep.GetServiceDependencyRef()
	}
	return serviceDepRefs
}

// ToTable converts a list of meshes into a table
func (f Formatter) ToTable(objs []runtime.Object, includeNS, wide bool) *output.Table {
	return &output.Table{
		Headers: f.makeHeaders(includeNS, wide),
		NextRow: f.makeNextTableRowFunc(objs, includeNS, wide),
	}
}

func (f Formatter) makeHeaders(includeNS, wide bool) []string {
	headers := make([]string, 0, 7)
	if includeNS {
		headers = append(headers, "namespace")
	}
	headers = append(headers, "name", "controller", "pod labels", "age")
	if wide {
		headers = append(headers, "service account", "destination hostnames")
	}

	return headers
}

func (f Formatter) makeNextTableRowFunc(objs []runtime.Object, includeNS, wide bool) func() []string {
	ndx := 0
	return func() []string {
		if ndx == len(objs) {
			return nil
		}

		workload := objs[ndx].(*discoveryv1.Workload)
		ndx++

		data := make([]string, 0, 7)
		if includeNS {
			data = append(data, workload.GetNamespace())
		}
		kubesWorkload := workload.Spec.GetKubernetes()
		data = append(
			data,
			workload.GetName(),
			output.RefToString(kubesWorkload.GetController()),
			f.makePodLablesCell(workload),
			output.FormatAge(workload.GetCreationTimestamp()),
		)
		if wide {
			data = append(
				data,
				kubesWorkload.GetServiceAccountName(),
				strings.Join(workload.Status.GetServiceDependencies().GetDestinationHostnames(), ","),
			)
		}

		return data
	}
}

func (f Formatter) makePodLablesCell(workload *discoveryv1.Workload) string {
	var labels sort.StringSlice
	for label, value := range workload.Spec.GetKubernetes().GetPodLabels() {
		labels = append(labels, fmt.Sprintf("%s=%s", label, value))
	}
	labels.Sort()
	return strings.Join(labels, ",")
}
