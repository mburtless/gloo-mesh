package accesspolicy

import (
	"fmt"
	"strings"

	commonv1 "github.com/solo-io/gloo-mesh/pkg/api/common.mesh.gloo.solo.io/v1"
	networkingv1 "github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/v1"
	"github.com/solo-io/gloo-mesh/pkg/openmeshctl/output"
	"github.com/solo-io/gloo-mesh/pkg/openmeshctl/resource"
	"github.com/solo-io/gloo-mesh/pkg/openmeshctl/resource/apply"
	"github.com/solo-io/gloo-mesh/pkg/openmeshctl/resource/registry"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func init() {
	registry.RegisterType(resource.Config{
		Name:      "accesspolicy",
		GVK:       networkingv1.AccessPolicyGVK,
		Plural:    "accesspolicies",
		Aliases:   []string{"ap", "aps"},
		Factory:   Factory{},
		Formatter: Formatter{},
		Applier:   apply.ServerSideApplier(&networkingv1.AccessPolicyGVK),
	})
}

// Factory creates new access policy objects.
type Factory struct{}

// New returns a new access policy.
func (f Factory) New() client.Object {
	return &networkingv1.AccessPolicy{}
}

// NewList returns a new access policy list.
func (f Factory) NewList() client.ObjectList {
	return &networkingv1.AccessPolicyList{}
}

// Formatter formats access policy objects.
type Formatter struct{}

// ToSummary converts a access policy to a summary representation.
func (f Formatter) ToSummary(obj runtime.Object) *output.Summary {
	ap := obj.(*networkingv1.AccessPolicy)
	fieldSet := output.FieldSet{}
	fieldSet.AddField("Allowed Methods", strings.Join(ap.Spec.GetAllowedPaths(), ","))
	fieldSet.AddField("Allowed Paths", ap.Spec.GetAllowedPaths())
	ports := make([]string, len(ap.Spec.GetAllowedPorts()))
	for i, port := range ap.Spec.GetAllowedPorts() {
		ports[i] = fmt.Sprint(port)
	}
	fieldSet.AddField("Allowed Ports", strings.Join(ports, ","))
	fieldSet.AddField("Workloads", ap.Status.GetWorkloads())
	dests := make(map[string]string, len(ap.Status.GetDestinations()))
	for name, status := range ap.Status.GetDestinations() {
		dests[name] = status.GetState().String()
	}
	fieldSet.AddField("Destinations", dests)
	return &output.Summary{Meta: ap, Fields: fieldSet}
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
	headers = append(headers, "name", "workloads", "destinations", "age")
	if wide {
		headers = append(headers, "allowed methods", "allowed paths", "allowed ports")
	}

	return headers
}

func (f Formatter) makeNextTableRowFunc(objs []runtime.Object, includeNS, wide bool) func() []string {
	ndx := 0
	return func() []string {
		if ndx == len(objs) {
			return nil
		}

		ap := objs[ndx].(*networkingv1.AccessPolicy)
		ndx++

		data := make([]string, 0, 8)
		if includeNS {
			data = append(data, ap.GetNamespace())
		}
		data = append(
			data,
			ap.GetName(),
			fmt.Sprint(len(ap.Status.GetWorkloads())),
			f.buildDestinationCell(ap),
			output.FormatAge(ap.GetCreationTimestamp()),
		)
		if wide {
			ports := make([]string, len(ap.Spec.GetAllowedPorts()))
			for i, port := range ap.Spec.GetAllowedPorts() {
				ports[i] = fmt.Sprint(port)
			}

			data = append(
				data,
				strings.Join(ap.Spec.AllowedMethods, ","),
				strings.Join(ap.Spec.AllowedPaths, ","),
				strings.Join(ports, ","),
			)
		}

		return data
	}
}

func (f Formatter) buildDestinationCell(ap *networkingv1.AccessPolicy) string {
	dests := ap.Status.GetDestinations()
	healthy := 0
	for _, dest := range dests {
		if dest.GetState() == commonv1.ApprovalState_ACCEPTED {
			healthy++
		}
	}

	return fmt.Sprintf("%d/%d", healthy, len(dests))
}
