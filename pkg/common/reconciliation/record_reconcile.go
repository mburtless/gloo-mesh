package reconciliation

import (
	"context"
	"fmt"
	"strings"

	"github.com/iancoleman/strcase"
	"sigs.k8s.io/controller-runtime/pkg/metrics"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/skv2/pkg/resource"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// the Recorder logs and records metrics for reconciles
type Recorder struct {
	inputName  string
	inputGvks  []schema.GroupVersionKind
	outputName string
	outputGvks []schema.GroupVersionKind

	counter *prometheus.CounterVec
}

// NOTE: Only one Recorder should be initialized for each unique metric.
func NewRecorder(
	counterOpts prometheus.CounterOpts,
	inputName string,
	inputGvks []schema.GroupVersionKind,
	outputName string,
	outputGvks []schema.GroupVersionKind,
) *Recorder {
	labels := makeReconcileLabels(
		inputName,
		inputGvks,
		outputName,
		outputGvks,
	)
	counter := prometheus.NewCounterVec(counterOpts, labels)

	// must happen only once for each unique metric, else will panic
	metrics.Registry.MustRegister(counter)

	return &Recorder{
		inputName:  inputName,
		inputGvks:  inputGvks,
		outputName: outputName,
		outputGvks: outputGvks,
		counter:    counter,
	}
}

// record a reconcile result
func (r *Recorder) RecordReconcileResult(
	ctx context.Context,
	input resource.ClusterSnapshot,
	output resource.ClusterSnapshot,
	success bool,
) {
	// count input gvks
	inputGvkCount := map[schema.GroupVersionKind]int{}
	input.ForEachObject(func(_ string, gvk schema.GroupVersionKind, _ resource.TypedObject) {
		inputGvkCount[gvk]++
	})

	// count output gvks
	outputIstioGvkCount := map[schema.GroupVersionKind]int{}
	output.ForEachObject(func(_ string, gvk schema.GroupVersionKind, _ resource.TypedObject) {
		outputIstioGvkCount[gvk]++
	})

	result := "success"
	if !success {
		result = "failure"
	}
	metricLabelValues := []string{result}
	loggerKeysAndValues := []interface{}{"result", result}

	var inputTotal int
	for _, gvk := range r.inputGvks {
		count := inputGvkCount[gvk]
		metricLabelValues = append(metricLabelValues, fmt.Sprintf("%v", count))
		loggerKeysAndValues = append(loggerKeysAndValues, "input/"+gvk.String(), count)
		inputTotal += count
	}

	metricLabelValues = append(metricLabelValues, fmt.Sprintf("%v", inputTotal))
	loggerKeysAndValues = append(loggerKeysAndValues, "input_total", inputTotal)

	var outputTotal int
	for _, gvk := range r.outputGvks {
		count := outputIstioGvkCount[gvk]
		metricLabelValues = append(metricLabelValues, fmt.Sprintf("%v", count))
		loggerKeysAndValues = append(loggerKeysAndValues, "output/"+gvk.String(), count)
		outputTotal += count
	}

	metricLabelValues = append(metricLabelValues, fmt.Sprintf("%v", outputTotal))
	loggerKeysAndValues = append(loggerKeysAndValues, "output_total", outputTotal)

	contextutils.LoggerFrom(ctx).Debugw("reconcile complete", loggerKeysAndValues...)

	r.counter.WithLabelValues(metricLabelValues...).Inc()
}

func gvkLabel(gvk schema.GroupVersionKind) string {
	snakeKind := strcase.ToSnake(gvk.Kind)
	sanitizedGroup := strings.ReplaceAll(gvk.Group, ".", "")
	return fmt.Sprintf("%v_%v_%v", sanitizedGroup, gvk.Version, snakeKind)
}

// generate the labels for counting reconciles
func makeReconcileLabels(inputName string, inputGvks []schema.GroupVersionKind, outputName string, outputGvks []schema.GroupVersionKind) []string {
	labels := []string{"result"}
	for _, gvk := range inputGvks {
		// count the appearance of each input gvk
		labels = append(labels, fmt.Sprintf("input_%v_%s_count", inputName, gvkLabel(gvk)))
	}
	labels = append(labels, "input_total")
	for _, gvk := range outputGvks {
		// count the appearance of each output gvk
		labels = append(labels, fmt.Sprintf("output_%v_%s_count", outputName, gvkLabel(gvk)))
	}
	labels = append(labels, "output_total")

	return labels
}
