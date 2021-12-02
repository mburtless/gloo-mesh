package trafficpolicy

import (
	"fmt"

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
		Name:      "trafficpolicy",
		GVK:       networkingv1.TrafficPolicyGVK,
		Plural:    "trafficpolicies",
		Aliases:   []string{"tp", "tps"},
		Factory:   Factory{},
		Formatter: Formatter{},
		Applier:   apply.ServerSideApplier(&networkingv1.TrafficPolicyGVK),
	})
}

// Factory creates new traffic policy objects.
type Factory struct{}

// New returns a new traffic policy.
func (f Factory) New() client.Object {
	return &networkingv1.TrafficPolicy{}
}

// NewList returns a new traffic policy list.
func (f Factory) NewList() client.ObjectList {
	return &networkingv1.TrafficPolicyList{}
}

// Formatter formats traffic policy objects.
type Formatter struct{}

// ToSummary converts a traffic policy to a summary representation.
func (f Formatter) ToSummary(obj runtime.Object) *output.Summary {
	tp := obj.(*networkingv1.TrafficPolicy)
	fieldSet := output.FieldSet{}
	fieldSet.AddField("Source Selectors", tp.Spec.GetSourceSelector())
	fieldSet.AddField("Destination Selectors", tp.Spec.GetDestinationSelector())
	// TODO(ryantking): Add policy field here
	fieldSet.AddField("Route Selectors", f.buildRouteSelectorFieldSet(tp))
	fieldSet.AddField("Workloads", tp.Status.GetWorkloads())
	dests := make(map[string]string, len(tp.Status.GetDestinations()))
	for name, state := range tp.Status.GetDestinations() {
		dests[name] = state.String()
	}
	fieldSet.AddField("Destinations", dests)
	return &output.Summary{Meta: tp, Fields: fieldSet}
}

func (f Formatter) buildRouteSelectorFieldSet(tp *networkingv1.TrafficPolicy) *output.FieldSet {
	fieldSet := output.FieldSet{}
	for _, sel := range tp.Spec.GetRouteSelector() {
		fieldSet.AddField("Route Labels", sel.GetRouteLabelMatcher())
	}
	return &fieldSet
}

// ToTable converts a list of traffic policies into a table
func (f Formatter) ToTable(objs []runtime.Object, includeNS, wide bool) *output.Table {
	return &output.Table{
		Headers: f.makeHeaders(includeNS),
		NextRow: f.makeNextTableRowFunc(objs, includeNS),
	}
}

func (f Formatter) makeHeaders(includeNS bool) []string {
	headers := make([]string, 0, 5)
	if includeNS {
		headers = append(headers, "namespace")
	}
	headers = append(headers, "name", "workloads", "destinations", "age")

	return headers
}

func (f Formatter) makeNextTableRowFunc(objs []runtime.Object, includeNS bool) func() []string {
	ndx := 0
	return func() []string {
		if ndx == len(objs) {
			return nil
		}

		tp := objs[ndx].(*networkingv1.TrafficPolicy)
		ndx++

		data := make([]string, 0, 8)
		if includeNS {
			data = append(data, tp.GetNamespace())
		}
		data = append(
			data,
			tp.GetName(),
			fmt.Sprint(len(tp.Status.GetDestinations())),
			fmt.Sprint(len(tp.Status.GetWorkloads())),
			output.FormatAge(tp.GetCreationTimestamp()),
		)

		return data
	}
}
