package reconciliation

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"go.uber.org/atomic"

	"github.com/iancoleman/strcase"
	"sigs.k8s.io/controller-runtime/pkg/metrics"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/skv2/pkg/resource"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// the level of logging to use when recording custom events
type LogLevel string

const (
	LogLevel_None   = "LogLevel_None"
	LogLevel_Error  = "LogLevel_Error"
	LogLevel_Warn   = "LogLevel_Warn"
	LogLevel_Debug  = "LogLevel_Debug"
	LogLevel_Info   = "LogLevel_Info"
	LogLevel_DPanic = "LogLevel_DPanic"
)

// the Recorder is used to record statistics about each reconcile. Abstracts concerns like logging,
// metrics. (tracing can be added to this as well)
type Recorder interface {
	// RecordReconcileResult records the result of a reconcile
	RecordReconcileResult(
		ctx context.Context,
		input resource.ClusterSnapshot,
		output resource.ClusterSnapshot,
		success bool,
	)

	// Registers a custom counter metric with the recorder.
	// If a counter with the given name has already been registered, this function is a no-op.
	RegisterCustomCounter(counter CustomCounter)

	// Increments the counter with the given name.
	// the number of label values for the counter should match exactly the number of label keys provided when registering the counter.
	IncrementCounter(
		ctx context.Context,
		opts IncrementCustomCounter,
	)
}

// CustomCounter provides options for registering a custom counter with the recorder
type CustomCounter struct {
	// opts for registering the counter. name must be unique
	Counter prometheus.CounterOpts
	// the set of keys for this counter's labels. each label key must have a matching value provided when the counter is incremented
	LabelKeys []string
}

// IncrementCustomCounter provides options for incrementing a custom counter
type IncrementCustomCounter struct {
	// name of the counter to increment. if it's not registered, nothing will happen
	CounterName string
	// the level of logging to use
	Level LogLevel
	// values corresponding to the regstered label keys. must match the label keys in length.
	LabelValues []string
	// an error if it occurred
	Err error
}

// the Recorder logs and records metrics for reconciles
type recorder struct {
	inputName  string
	inputGvks  []schema.GroupVersionKind
	outputName string
	outputGvks []schema.GroupVersionKind

	counter *prometheus.CounterVec
	count   int

	// arbitrary counters registered from anywhere
	customCounters     map[string]*counterVec
	customCountersLock sync.RWMutex
}

type counterVec struct {
	counter   *prometheus.CounterVec
	count     *atomic.Uint32
	labelKeys []string
}

// NOTE: Only one Recorder should be initialized for each unique metric.
func NewRecorder(
	counterOpts prometheus.CounterOpts,
	inputName string,
	inputGvks []schema.GroupVersionKind,
	outputName string,
	outputGvks []schema.GroupVersionKind,
) Recorder {
	labels := makeReconcileLabels(
		inputName,
		inputGvks,
		outputName,
		outputGvks,
	)
	counter := prometheus.NewCounterVec(counterOpts, labels)

	// must happen only once for each unique metric, else will panic
	metrics.Registry.MustRegister(counter)

	return &recorder{
		inputName:      inputName,
		inputGvks:      inputGvks,
		outputName:     outputName,
		outputGvks:     outputGvks,
		counter:        counter,
		customCounters: map[string]*counterVec{},
	}
}

func (r *recorder) RecordReconcileResult(
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

	r.counter.WithLabelValues(metricLabelValues...).Inc()
	r.count++

	loggerKeysAndValues = append(loggerKeysAndValues, "total_reconciles", r.count)

	contextutils.LoggerFrom(ctx).Debugw("reconcile complete", loggerKeysAndValues...)
}

func (r *recorder) RegisterCustomCounter(
	counter CustomCounter,
) {
	r.customCountersLock.RLock()
	_, counterRegistered := r.customCounters[counter.Counter.Name]
	r.customCountersLock.RUnlock()
	if counterRegistered {
		// re-registering causes a panic
		return
	}

	r.customCountersLock.Lock()
	customCounter := prometheus.NewCounterVec(counter.Counter, counter.LabelKeys)
	// must happen only once for each unique metric, else will panic
	metrics.Registry.MustRegister(customCounter)
	r.customCounters[counter.Counter.Name] = &counterVec{
		counter:   customCounter,
		labelKeys: counter.LabelKeys,
		count:     atomic.NewUint32(0),
	}
	r.customCountersLock.Unlock()
}

func (r *recorder) IncrementCounter(
	ctx context.Context,
	opts IncrementCustomCounter,
) {
	counterName := opts.CounterName
	labelValues := opts.LabelValues
	r.customCountersLock.RLock()
	defer r.customCountersLock.RUnlock()
	counter, counterRegistered := r.customCounters[counterName]

	if !counterRegistered {
		// internal error, should never happen
		contextutils.LoggerFrom(ctx).Warnf("no counter registered with name %v", counterName)
		return
	}
	logger := contextutils.LoggerFrom(ctx)
	logFn := func(template string, args ...interface{}) {}
	switch opts.Level {
	case LogLevel_Error:
		logFn = logger.Errorf
	case LogLevel_Warn:
		logFn = logger.Warnf
	case LogLevel_Debug:
		logFn = logger.Debugf
	case LogLevel_Info:
		logFn = logger.Infof
	case LogLevel_DPanic:
		logFn = logger.DPanicf
	}
	labels := map[string]string{}
	if len(counter.labelKeys) == len(labelValues) {
		// only record metric if label keys and values match
		counter.counter.WithLabelValues(labelValues...).Inc()

		for i := range counter.labelKeys {
			labels[counter.labelKeys[i]] = labelValues[i]
		}
	}
	if opts.Err != nil {
		logFn("%v: %v events recorded. labels:<%v>. error:<%v>", counterName, counter.count, labels, opts.Err)
	} else {
		logFn("%v: %v events recorded. labels:<%v>", counterName, counter.count, labels)
	}

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

type ctxKey struct{}

// ContextWithRecorder stores the recorder in the child context
func ContextWithRecorder(ctx context.Context, recorder Recorder) context.Context {
	return context.WithValue(ctx, ctxKey{}, recorder)
}

// RecorderFromContext returns the recorder stored int the context.
// if none is available, a no-op recorder is returned
func RecorderFromContext(ctx context.Context) Recorder {
	if ctx != nil {
		if recorder, ok := ctx.Value(ctxKey{}).(Recorder); ok {
			return recorder
		}
	}
	return noopRecorder{}
}

// does nothing, implements interface
type noopRecorder struct{}

func (r noopRecorder) RecordReconcileResult(ctx context.Context, input resource.ClusterSnapshot, output resource.ClusterSnapshot, success bool) {
}

func (r noopRecorder) RegisterCustomCounter(
	counter CustomCounter,
) {
}
func (r noopRecorder) IncrementCounter(
	ctx context.Context,
	opts IncrementCustomCounter,
) {
}
