package destination

import (
	"fmt"
	"strings"

	discoveryv1 "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1"
	"github.com/solo-io/gloo-mesh/pkg/openmeshctl/output"
	"github.com/solo-io/gloo-mesh/pkg/openmeshctl/resource"
	"github.com/solo-io/gloo-mesh/pkg/openmeshctl/resource/apply"
	"github.com/solo-io/gloo-mesh/pkg/openmeshctl/resource/registry"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func init() {
	registry.RegisterType(resource.Config{
		Name:      "destination",
		GVK:       discoveryv1.DestinationGVK,
		Plural:    "destinations",
		Aliases:   []string{"dest", "dests"},
		Factory:   Factory{},
		Formatter: Formatter{},
		Applier:   apply.ServerSideApplier(&discoveryv1.DestinationGVK),
	})
}

// Factory creates new destination objects.
type Factory struct{}

// New returns a new destination.
func (f Factory) New() client.Object {
	return &discoveryv1.Destination{}
}

// NewList returns a new destination list.
func (f Factory) NewList() client.ObjectList {
	return &discoveryv1.DestinationList{}
}

// Formatter formats destination objects.
type Formatter struct{}

// ToSummary converts a destination to a summary representation.
func (f Formatter) ToSummary(obj runtime.Object) *output.Summary {
	dest := obj.(*discoveryv1.Destination)
	fieldSet := output.FieldSet{}
	switch svc := dest.Spec.GetType().(type) {
	case *discoveryv1.DestinationSpec_KubeService_:
		fieldSet.AddField("Kube Service", f.makeKubeServiceFieldSet(svc.KubeService))
	case *discoveryv1.DestinationSpec_ExternalService_:
		fieldSet.AddField("External Service", f.makeExtServiceFieldSet(svc.ExternalService))
	}

	return &output.Summary{Meta: dest, Fields: fieldSet}
}

func (f Formatter) makeKubeServiceFieldSet(svc *discoveryv1.DestinationSpec_KubeService) *output.FieldSet {
	fieldSet := output.FieldSet{}
	fieldSet.AddField("Type", svc.GetServiceType().String())
	fieldSet.AddField("Region", svc.GetRegion())
	fieldSet.AddField("Service", svc.GetRef())
	fieldSet.AddField("Labels", svc.GetLabels())
	fieldSet.AddField("Workload Selector Labels", svc.GetWorkloadSelectorLabels())
	ports := make([]string, len(svc.GetPorts()))
	for i, port := range svc.GetPorts() {
		ports[i] = port.GetName()
	}
	fieldSet.AddField("Ports", strings.Join(ports, ","))
	subsets := make(map[string]string, len(svc.GetSubsets()))
	for name, subset := range svc.GetSubsets() {
		subsets[name] = strings.Join(subset.GetValues(), ",")
	}
	fieldSet.AddField("Subsets", subsets)
	extAddrs := make([]string, 0, len(svc.GetExternalAddresses()))
	for _, extAddr := range svc.GetExternalAddresses() {
		switch a := extAddr.GetExternalAddressType().(type) {
		case *discoveryv1.DestinationSpec_KubeService_ExternalAddress_DnsName:
			extAddrs = append(extAddrs, a.DnsName)
		case *discoveryv1.DestinationSpec_KubeService_ExternalAddress_Ip:
			extAddrs = append(extAddrs, a.Ip)
		}
	}
	fieldSet.AddField("External Addresses", extAddrs)
	return &fieldSet
}

func (f Formatter) makeExtServiceFieldSet(svc *discoveryv1.DestinationSpec_ExternalService) *output.FieldSet {
	fieldSet := output.FieldSet{}
	fieldSet.AddField("Name", svc.GetName())
	fieldSet.AddField("Hosts", svc.GetHosts())
	fieldSet.AddField("Addresses", svc.GetAddresses())
	ports := make([]string, len(svc.GetPorts()))
	for i, port := range svc.GetPorts() {
		ports[i] = port.GetName()
	}
	fieldSet.AddField("Ports", strings.Join(ports, ","))
	endpoints := output.FieldSet{}
	for _, endpoint := range svc.GetEndpoints() {
		ports := make(map[string]string, len(endpoint.GetPorts()))
		for name, number := range endpoint.GetPorts() {
			ports[name] = fmt.Sprint(number)
		}
		endpoints.AddField(endpoint.GetAddress(), ports)
	}
	fieldSet.AddField("Endpoints", endpoints)
	return &fieldSet
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
	headers = append(headers, "name", "type", "service", "external service hosts", "age")
	if wide {
		headers = append(headers, "ports")
	}

	return headers
}

func (f Formatter) makeNextTableRowFunc(objs []runtime.Object, includeNS, wide bool) func() []string {
	ndx := 0
	return func() []string {
		if ndx == len(objs) {
			return nil
		}

		dest := objs[ndx].(*discoveryv1.Destination)
		ndx++

		data := make([]string, 0, 7)
		if includeNS {
			data = append(data, dest.GetNamespace())
		}
		data = append(
			data,
			dest.GetName(),
			f.getTypeCell(dest),
			output.RefToString(dest.Spec.GetKubeService().GetRef()),
			f.getExtSvcHostCell(dest),
			output.FormatAge(dest.GetCreationTimestamp()),
		)
		if wide {
			data = append(data, f.getPortsCell(dest))
		}

		return data
	}
}

func (f Formatter) getTypeCell(dest *discoveryv1.Destination) string {
	switch dest.Spec.GetType().(type) {
	case *discoveryv1.DestinationSpec_KubeService_:
		return "Kubernetes Service"
	case *discoveryv1.DestinationSpec_ExternalService_:
		return "External Service"
	}

	return "Unknown"
}

func (f Formatter) getExtSvcHostCell(dest *discoveryv1.Destination) string {
	svc := dest.Spec.GetExternalService()
	if svc == nil {
		return ""
	}

	return strings.Join(svc.GetHosts(), ",")
}

func (f Formatter) getPortsCell(dest *discoveryv1.Destination) string {
	var ports []string
	switch svc := dest.Spec.GetType().(type) {
	case *discoveryv1.DestinationSpec_KubeService_:
		ports = make([]string, len(svc.KubeService.GetPorts()))
		for i, port := range svc.KubeService.GetPorts() {
			ports[i] = port.GetName()
		}

	case *discoveryv1.DestinationSpec_ExternalService_:
		ports = make([]string, len(svc.ExternalService.GetPorts()))
		for i, port := range svc.ExternalService.GetPorts() {
			ports[i] = port.GetName()
		}
	}

	return strings.Join(ports, ",")
}
