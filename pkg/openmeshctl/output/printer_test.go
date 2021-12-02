package output_test

import (
	"bytes"
	"embed"
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	discoveryv1 "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1"
	"github.com/solo-io/gloo-mesh/pkg/openmeshctl/output"
	"github.com/solo-io/gloo-mesh/pkg/openmeshctl/test/util"
	v1 "github.com/solo-io/skv2/pkg/api/core.skv2.solo.io/v1"
	"github.com/solo-io/skv2/pkg/ezkube"
)

//go:embed testdata
var testData embed.FS

var _ = Describe("Printer", func() {
	var (
		printer output.Printer
		out     = &bytes.Buffer{}
	)

	BeforeEach(func() {
		out.Reset()
		printer = output.NewPrinter(out)
	})

	Describe("Printing encoded objects", func() {
		var dest *discoveryv1.Destination

		BeforeEach(func() {
			dest = util.Destinations()[0]
		})

		type testCase struct {
			format     output.Format
			goldenFile string
		}

		DescribeTable("output formats", func(tt testCase) {
			err := printer.PrintRaw(dest, tt.format)
			Expect(err).ToNot(HaveOccurred())
			Expect(out).To(util.MatchGoldenFile(testData, "testdata/"+tt.goldenFile))
		},
			Entry("YAML", testCase{
				format:     output.YAML,
				goldenFile: "destination.golden.yaml",
			}),
			Entry("JSON", testCase{
				format:     output.JSON,
				goldenFile: "destination.golden.json",
			}),
		)

		It("should return an error when an invalid format is given", func() {
			err := printer.PrintRaw(dest, output.Format("unknown"))
			Expect(err).To(MatchError("unknown output format: unknown"))
		})
	})

	Describe("Printing summaries", func() {
		var summary *output.Summary

		BeforeEach(func() {
			dest := util.Destinations()[0]
			summary = &output.Summary{Meta: dest}
			kubeServiceFields := output.FieldSet{}
			kubeServiceFields.AddField("External Addresses", []string{})
			endpoints := []string{}
			for _, subset := range dest.Spec.GetKubeService().GetEndpointSubsets() {
				for _, endpoint := range subset.Endpoints {
					endpoints = append(endpoints, endpoint.GetIpAddress())
				}
			}
			kubeServiceFields.AddField("IP Addresses", endpoints)
			summary.Fields.AddField("Kube Service", kubeServiceFields)
			summary.Fields.AddField("Traffic Policies", []ezkube.ResourceId{})
			summary.Fields.AddField("Access Policies", []ezkube.ResourceId{
				&v1.ObjectRef{Name: "details", Namespace: "gloo-mesh"},
				&v1.ObjectRef{Name: "details-admin", Namespace: "gloo-mesh"},
			})
			summary.Fields.AddField("Mesh", dest.Spec.GetMesh())
			summary.Fields.AddField("Local FQDN", dest.Status.GetLocalFqdn())
		})

		It("should correctly render the summary", func() {
			err := printer.PrintSummary(summary)
			Expect(err).ToNot(HaveOccurred())
			Expect(out).To(util.MatchGoldenFile(testData, "testdata/summary.golden.txt"))
		})
	})

	Describe("Printing tables", func() {
		var table *output.Table

		BeforeEach(func() {
			dests := util.Destinations()
			ndx := 0
			table = &output.Table{
				Headers: []string{"namespace", "name", "mesh"},
				NextRow: func() []string {
					if ndx == len(dests) {
						return nil
					}

					dest := dests[ndx]
					ndx++
					meshRef := fmt.Sprintf("%s.%s", dest.Spec.Mesh.GetNamespace(), dest.Spec.Mesh.GetName())
					return []string{dest.GetNamespace(), dest.GetName(), meshRef}
				},
			}
		})

		It("should correctly render the table", func() {
			err := printer.PrintTable(table)
			Expect(err).ToNot(HaveOccurred())
			Expect(out).To(util.MatchGoldenFile(testData, "testdata/table.golden.txt"))
		})
	})
})
