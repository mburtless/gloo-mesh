package output

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/ghodss/yaml"
	"github.com/rotisserie/eris"
	commonv1 "github.com/solo-io/gloo-mesh/pkg/api/common.mesh.gloo.solo.io/v1"
	v1 "github.com/solo-io/skv2/pkg/api/core.skv2.solo.io/v1"
	"github.com/solo-io/skv2/pkg/ezkube"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

//go:generate mockgen -destination mocks/printer.go . Printer

// Printer prints out resources.
type Printer interface {
	// PrintRaw marshals objects into bytes and prints them.
	PrintRaw(obj runtime.Object, fmt Format) error

	// PrintSummary prints the summary for an object
	PrintSummary(summary *Summary) error

	// PrintTable prints a table.
	PrintTable(table *Table) error
}

type printer struct {
	out io.Writer
}

// NewPrinter returns a new printer that writes to the given output.
func NewPrinter(out io.Writer) Printer {
	return &printer{out}
}

// PrintRaw encodes and prints the object to the output.
func (p *printer) PrintRaw(obj runtime.Object, fmt Format) error {
	switch fmt {
	case JSON:
		return p.printJSON(obj)
	case YAML:
		return p.printYAML(obj)
	default:
		return eris.Errorf("unknown output format: %s", fmt)
	}
}

func (p *printer) printJSON(obj runtime.Object) error {
	b, err := json.MarshalIndent(obj, "", " ")
	if err != nil {
		return err
	}

	_, err = p.out.Write(append(b, '\n'))
	return err
}
func (p *printer) printYAML(obj runtime.Object) error {
	b, err := yaml.Marshal(obj)
	if err != nil {
		return err
	}
	_, err = p.out.Write(b)
	return err
}

// PrintSummary prints the summary to the console.
func (p *printer) PrintSummary(summary *Summary) error {
	w := tabwriter.NewWriter(p.out, 0, 8, 2, ' ', 0)
	p.printMeta(w, summary.Meta)
	p.printFieldSet(w, 0, summary.Fields)
	return w.Flush()
}

func (p *printer) printMeta(w io.Writer, meta metav1.Object) {
	fmt.Fprintf(w, "Name:\t%s\n", meta.GetName())
	fmt.Fprintf(w, "Namespace:\t%s\n", meta.GetNamespace())
	fmt.Fprintf(w, "CreationTimestamp:\t%s\n", meta.GetCreationTimestamp().In(time.UTC).Format(time.RFC1123Z))
	fmt.Fprint(w, "Labels:")
	p.printMap(w, meta.GetLabels())
	fmt.Fprint(w, "Annotations:")
	p.printMap(w, meta.GetAnnotations())
}

func (p *printer) printFieldSet(w io.Writer, spaces int, fieldSet FieldSet) {
	for _, field := range fieldSet {
		p.printIndented(w, spaces, "%s:", field.Label)
		switch value := field.Value.(type) {
		case FieldSet:
			fmt.Fprint(w, "\n")
			p.printFieldSet(w, spaces+1, value)
		case *FieldSet:
			fmt.Fprint(w, "\n")
			p.printFieldSet(w, spaces+1, *value)
		case []string:
			p.printListValue(w, spaces, value)
		case map[string]string:
			p.printMap(w, value)
		case string:
			p.printStringValue(w, value)
		case []ezkube.ResourceId:
			p.printResourceIdList(w, spaces+1, value)
		case ezkube.ResourceId:
			p.printResourceId(w, value)
		case []ezkube.ClusterResourceId:
			p.printClusterResourceIdList(w, spaces+1, value)
		case ezkube.ClusterResourceId:
			p.printClusterResourceId(w, value)
		case []*v1.ObjectRef:
			p.printRefList(w, spaces+1, value)
		case []*v1.ClusterObjectRef:
			p.printClusterRefList(w, spaces+1, value)
		case *commonv1.DestinationSelector:
			p.printDestinationSelector(w, spaces+1, value)
		case []*commonv1.DestinationSelector:
			p.printDestinationSelectors(w, spaces+1, value)
		case *commonv1.IdentitySelector:
			p.printIdentitySelector(w, spaces+1, value)
		case []*commonv1.IdentitySelector:
			p.printIdentitySelectors(w, spaces+1, value)
		case *commonv1.WorkloadSelector:
			p.printWorkloadSelector(w, spaces+1, value)
		case []*commonv1.WorkloadSelector:
			p.printWorkloadSelectors(w, spaces+1, value)
		case matcher:
		default:
			fmt.Fprintf(w, "\t%v\n", value)
		}
	}
}

// used for labels and annotations
func (p *printer) printMap(w io.Writer, m map[string]string) {
	if len(m) == 0 {
		p.printNone(w)
		return
	}

	sorted := make(sort.StringSlice, 0, len(m))
	for key, value := range m {
		if key == corev1.LastAppliedConfigAnnotation {
			continue
		}
		sep := "="
		if len(value) > 140 {
			value = value[:137] + "..."
			sep = ":\n\t  "
		}

		sorted = append(sorted, key+sep+value)
	}
	if sorted.Len() == 0 {
		p.printNone(w)
		return
	}

	sorted.Sort()
	for _, item := range sorted {
		fmt.Fprintf(w, "\t%s\n", item)
	}
}

func (p *printer) printListValue(w io.Writer, spaces int, value []string) {
	if len(value) == 0 {
		p.printNone(w)
		return
	}

	fmt.Fprint(w, "\n")
	for _, item := range value {
		p.printIndented(w, spaces+2, item+"\n")
	}
}

func (p *printer) printShortListValue(w io.Writer, value []string) {
	if len(value) == 0 {
		p.printNone(w)
		return
	}

	p.printStringValue(w, strings.Join(value, ","))
}

func (p *printer) printStringValue(w io.Writer, value string) {
	if value == "" {
		p.printNone(w)
		return
	}

	fmt.Fprintf(w, "\t%s\n", value)
}

func (p *printer) printNone(w io.Writer) {
	fmt.Fprintln(w, "\t<none>")
}

func (p *printer) printResourceIdList(w io.Writer, spaces int, ids []ezkube.ResourceId) {
	if len(ids) == 0 {
		p.printNone(w)
		return
	}

	fmt.Fprint(w, "\n")
	for _, id := range ids {
		p.printIndented(w, spaces, "%s.%s\n", id.GetNamespace(), id.GetName())
	}
}

func (p *printer) printRefList(w io.Writer, spaces int, refs []*v1.ObjectRef) {
	if len(refs) == 0 {
		p.printNone(w)
		return
	}

	fmt.Fprint(w, "\n")
	for _, ref := range refs {
		p.printIndented(w, spaces, "%s.%s\n", ref.GetNamespace(), ref.GetName())
	}
}

func (p *printer) printResourceId(w io.Writer, id ezkube.ResourceId) {
	if id.GetName() == "" {
		p.printNone(w)
		return
	}

	fmt.Fprintf(w, "\t%s.%s\n", id.GetNamespace(), id.GetName())
}

func (p *printer) printClusterResourceIdList(w io.Writer, spaces int, ids []ezkube.ClusterResourceId) {
	if len(ids) == 0 {
		p.printNone(w)
		return
	}

	fmt.Fprint(w, "\n")
	for _, id := range ids {
		var sb strings.Builder
		sb.WriteRune('\t')
		if id.GetClusterName() != "" {
			sb.WriteString(id.GetClusterName() + ".")
		}
		sb.WriteString(id.GetNamespace() + "." + id.GetName() + "\n")
		p.printIndented(w, spaces, sb.String())
	}
}

func (p *printer) printClusterRefList(w io.Writer, spaces int, refs []*v1.ClusterObjectRef) {
	if len(refs) == 0 {
		p.printNone(w)
		return
	}

	fmt.Fprint(w, "\n")
	for _, ref := range refs {
		var sb strings.Builder
		sb.WriteRune('\t')
		if ref.GetClusterName() != "" {
			sb.WriteString(ref.GetClusterName() + ".")
		}
		sb.WriteString(ref.GetNamespace() + "." + ref.GetName() + "\n")
		p.printIndented(w, spaces, sb.String())
	}
}

func (p *printer) printClusterResourceId(w io.Writer, id ezkube.ClusterResourceId) {
	if id.GetName() == "" {
		p.printNone(w)
		return
	}

	fmt.Fprint(w, "\t")
	if id.GetClusterName() != "" {
		fmt.Fprint(w, id.GetClusterName()+".")
	}

	fmt.Fprintf(w, "\t%s.%s\n", id.GetNamespace(), id.GetName())
}

type matcher interface {
	GetNamespaces() []string
}

type clusterMatcher interface {
	GetClusters() []string
}

type labeledMatcher interface {
	GetLabels() map[string]string
}

func (p *printer) printMatcher(w io.Writer, spaces int, matcher matcher) {
	fmt.Fprint(w, "\n")
	p.printIndented(w, spaces, "Namespaces: ")
	p.printShortListValue(w, matcher.GetNamespaces())
	if cm, ok := matcher.(clusterMatcher); ok {
		p.printIndented(w, spaces, "Clusters: ")
		p.printShortListValue(w, cm.GetClusters())
	}
	if lm, ok := matcher.(labeledMatcher); ok {
		p.printIndented(w, spaces, "Labels:")
		p.printMap(w, lm.GetLabels())
	}
}

func (p *printer) printObjectSelector(w io.Writer, spaces int, sel *v1.ObjectSelector) {
	p.printMatcher(w, spaces+1, sel)
}

func (p *printer) printObjectSelectors(w io.Writer, spaces int, sels []*v1.ObjectSelector) {
	if len(sels) == 0 {
		p.printNone(w)
		return
	}

	for _, sel := range sels {
		p.printObjectSelector(w, spaces, sel)
	}
}

func (p *printer) printDestinationSelector(w io.Writer, spaces int, sel *commonv1.DestinationSelector) {
	fmt.Fprint(w, "\n")
	p.printIndented(w, spaces, "Matcher:")
	p.printMatcher(w, spaces+1, sel.GetKubeServiceMatcher())
	p.printIndented(w, spaces, "Refs:")
	p.printClusterRefList(w, spaces+1, sel.GetKubeServiceRefs().Services)
}

func (p *printer) printDestinationSelectors(w io.Writer, spaces int, sels []*commonv1.DestinationSelector) {
	if len(sels) == 0 {
		p.printNone(w)
		return
	}

	for _, sel := range sels {
		p.printDestinationSelector(w, spaces, sel)
	}
}

func (p *printer) printIdentitySelector(w io.Writer, spaces int, sel *commonv1.IdentitySelector) {
	fmt.Fprint(w, "\n")
	p.printIndented(w, spaces, "Matcher:")
	p.printMatcher(w, spaces+1, sel.GetKubeIdentityMatcher())
	p.printIndented(w, spaces, "Service Accounts:")
	p.printClusterRefList(w, spaces+1, sel.GetKubeServiceAccountRefs().GetServiceAccounts())
	p.printIndented(w, spaces, "Request Matcher:\n")
	p.printIndented(w, spaces+1, "Principles:")
	p.printShortListValue(w, sel.GetRequestIdentityMatcher().GetRequestPrincipals())
	p.printIndented(w, spaces+1, "Not Principles:")
	p.printShortListValue(w, sel.GetRequestIdentityMatcher().GetRequestPrincipals())
}

func (p *printer) printIdentitySelectors(w io.Writer, spaces int, sels []*commonv1.IdentitySelector) {
	if len(sels) == 0 {
		p.printNone(w)
		return
	}

	for _, sel := range sels {
		p.printIdentitySelector(w, spaces, sel)
	}
}

func (p *printer) printWorkloadSelector(w io.Writer, spaces int, sel *commonv1.WorkloadSelector) {
	fmt.Fprint(w, "\n")
	p.printIndented(w, spaces, "Matcher:")
	p.printMatcher(w, spaces, sel.GetKubeWorkloadMatcher())
}

func (p *printer) printWorkloadSelectors(w io.Writer, spaces int, sels []*commonv1.WorkloadSelector) {
	if len(sels) == 0 {
		p.printNone(w)
		return
	}

	for _, sel := range sels {
		p.printWorkloadSelector(w, spaces, sel)
	}
}

func (p *printer) printIndented(w io.Writer, spaces int, format string, a ...interface{}) {
	buf := bufio.NewWriter(w)
	for i := 0; i < spaces; i++ {
		buf.WriteRune(' ')
	}
	fmt.Fprintf(buf, format, a...)
	buf.Flush()
}

// PrintTable prints the given table to the output.
func (p *printer) PrintTable(table *Table) error {
	w := tabwriter.NewWriter(p.out, 1, 0, 3, ' ', 0)
	_, err := fmt.Fprintln(w, strings.ToUpper(strings.Join(table.Headers, "\t")))
	if err != nil {
		return err
	}
	for table.HasNextRow() {
		row := table.GetNextRow()
		for i, cell := range row {
			if len(cell) > 48 {
				row[i] = cell[:45] + "..."
			} else if len(cell) == 0 {
				row[i] = "<none>"
			}
		}
		if _, err := fmt.Fprintln(w, strings.Join(row, "\t")); err != nil {
			return err
		}
	}

	return w.Flush()
}
