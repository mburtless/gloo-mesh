package istio

import (
	"context"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/rotisserie/eris"
	corev1sets "github.com/solo-io/external-apis/pkg/api/k8s/core/v1/sets"
	"github.com/solo-io/gloo-mesh/pkg/common/reconciliation"
	"github.com/solo-io/gloo-mesh/pkg/mesh-discovery/translation/workload/types"
	skv1 "github.com/solo-io/skv2/pkg/api/core.skv2.solo.io/v1"
	"github.com/solo-io/skv2/pkg/ezkube"
	"istio.io/api/annotation"
	"istio.io/istio/galley/pkg/config/analysis/analyzers/injection"
	"istio.io/istio/galley/pkg/config/analysis/analyzers/util"
	"istio.io/istio/pkg/config/constants"
	"istio.io/istio/pkg/config/resource"
	"istio.io/istio/pkg/kube/inject"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"

	v1 "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1"
	v1sets "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1/sets"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/skv2/contrib/pkg/sets"
)

const (
	// defaultInjectorConfigMapName is the default name of the ConfigMap with the injection config
	// The actual name can be different - use getInjectorConfigMapName
	// source: https://github.com/istio/istio/blob/master/pilot/pkg/bootstrap/sidecarinjector.go
	defaultInjectorConfigMapName = "istio-sidecar-injector"

	// the injector config key, hard-coded in istiod:
	// source: https://github.com/istio/istio/blob/master/pilot/pkg/bootstrap/sidecarinjector.go
	injectorConfigMapKey = "config"
)

var (
	// namespaces for which workloads are ignored (never injected)
	ignoredNamespaces = []string{
		constants.KubeSystemNamespace,
		constants.KubePublicNamespace,
		constants.KubeNodeLeaseNamespace,
		constants.LocalPathStorageNamespace,
	}

	// counter for tracking the number of times we failed to parse an injection config for a mesh
	istioInjectionConfigParseFailed = reconciliation.CustomCounter{
		Counter: prometheus.CounterOpts{
			Name: "istio_workload_detector_injector_config_parse_failed",
			Help: "Tracks the number of times the Gloo Mesh Discovery component failed to parse Istio's sidecar injector config",
		},
		LabelKeys: []string{"mesh"},
	}
	recordInjectionConfigParseFailed = func(ctx context.Context, meshId ezkube.ResourceId, err error) {
		reconciliation.RecorderFromContext(ctx).IncrementCounter(ctx, reconciliation.IncrementCustomCounter{
			CounterName: istioInjectionConfigParseFailed.Counter.Name,
			Level:       reconciliation.LogLevel_Error,
			LabelValues: []string{sets.Key(meshId)},
			Err:         err,
		})
	}
)

// TODO(ilackarms): currently we produce a mesh ref that maps directly to the cluster

// detects an istio workload
type workloadDetector struct {
	ctx        context.Context
	configMaps corev1sets.ConfigMapSet

	// map of each istio mesh id to its corresponding injector config, if found
	// we maintain a cache as to not recalculate the injector config multiple times for the same mesh
	cachedInjectorConfigs map[string]*inject.Config
	// map of each namespace to a boolean indicating the namespace has istio injection enabled
	injectedNamespaces map[string]bool
}

func NewWorkloadDetector(
	ctx context.Context,
	namespaces corev1sets.NamespaceSet,
	configMaps corev1sets.ConfigMapSet,
) *workloadDetector {
	ctx = contextutils.WithLogger(ctx, "istio-workload-detector")

	reconciliation.RecorderFromContext(ctx).RegisterCustomCounter(istioInjectionConfigParseFailed)

	injectedNamespaces := map[string]bool{}
	for _, namespace := range namespaces.UnsortedList() {
		if namespaceInjectionEnabled(namespace) {
			injectedNamespaces[namespaceKey(namespace.Name, namespace.ClusterName)] = true
		}
	}

	return &workloadDetector{
		ctx:                   ctx,
		injectedNamespaces:    injectedNamespaces,
		configMaps:            configMaps,
		cachedInjectorConfigs: map[string]*inject.Config{},
	}
}

func namespaceKey(name, cluster string) string {
	return name + "/" + cluster
}

// logic modeled based on https://github.com/istio/istio/blob/master/galley/pkg/config/analysis/analyzers/injection/injection.go
func namespaceInjectionEnabled(namespace *corev1.Namespace) bool {
	name := namespace.Name
	if name == constants.IstioSystemNamespace {
		return false
	}
	if util.IsSystemNamespace(resource.Namespace(name)) {
		return false
	}

	injectionLabel := namespace.Labels[util.InjectionLabelName]

	// TODO(ilackarms): store revision instead of a boolean when supported
	_, okNewInjectionLabel := namespace.Labels[injection.RevisionInjectionLabelName]

	// If legacy label has any value other than the enablement value, they are deliberately not injecting it, so ignore
	return okNewInjectionLabel || injectionLabel == util.InjectionLabelEnableValue
}

func (d workloadDetector) DetectMeshForWorkload(workload types.Workload, meshes v1sets.MeshSet) *v1.Mesh {
	for _, mesh := range meshes.List() {
		istio := mesh.Spec.GetIstio()
		if istio == nil {
			// only care about istio workloads
			continue
		}

		if istio.GetInstallation().GetCluster() != workload.GetClusterName() {
			continue
		}

		// if the workload's pod spec already contains the proxy and is in the mesh's cluster,
		// we consider it to belong to that mesh.
		// TODO(ilackarms): we will need to rely on revision or some other property to correctly attribute workloads to their owning mesh when we support multiple meshes per cluster
		if hasMeshProxyContainer(workload) {
			// the workload is either a gateway or has had its proxy manually injected
			return mesh
		}

		isInjected := d.meshInjectsWorkload(
			d.ctx,
			workload,
			mesh,
		)

		if isInjected {
			return mesh
		}

	}
	return nil
}

// returns true if the given workload contains a proxy container that has been manually added
func hasMeshProxyContainer(
	workload types.Workload,
) bool {
	// TODO(ilackarms): currently we assume one mesh per cluster,
	// and that the control plane for a given proxy is always
	// the mesh
	return containsSidecarContainer(workload.GetPodTemplate().Spec.Containers)
}

// returns true if the given workload is injected by the given mesh (istiod) instance
func (d workloadDetector) meshInjectsWorkload(
	ctx context.Context,
	workload types.Workload,
	mesh *v1.Mesh,
) bool {

	namespaceInjected := d.injectedNamespaces[namespaceKey(workload.GetNamespace(), workload.GetClusterName())]

	cfg, ok := d.cachedInjectorConfigs[sets.Key(mesh)]
	if !ok {
		var err error
		cfg, err = getInjectorConfig(mesh.Spec.GetIstio(), d.configMaps)
		if err != nil {
			recordInjectionConfigParseFailed(ctx, mesh, eris.Wrap(err, "getting injector config for mesh"))
		}
		d.cachedInjectorConfigs[sets.Key(mesh)] = cfg
	}

	return injectRequired(
		ctx,
		ignoredNamespaces,
		namespaceInjected,
		cfg,
		workload,
	)
}

func getInjectorConfig(
	istioMesh *v1.MeshSpec_Istio,
	configMaps corev1sets.ConfigMapSet,
) (*inject.Config, error) {
	injectorCm, err := configMaps.Find(&skv1.ClusterObjectRef{
		Name:        getInjectorConfigMapName(""), // TODO(ilackarms): support mesh revisions here
		Namespace:   istioMesh.GetInstallation().GetNamespace(),
		ClusterName: istioMesh.GetInstallation().GetCluster(),
	})
	if err != nil {
		// TODO(ilackarms): no injector configured for this mesh, explore other options for workload detection
		return nil, nil
	}
	rawCfg, ok := injectorCm.Data[injectorConfigMapKey]
	if !ok {
		return nil, eris.Errorf("injector configmap %v missing 'config' data key", sets.Key(injectorCm))
	}
	cfg, err := inject.UnmarshalConfig([]byte(rawCfg))
	if err != nil {
		return nil, eris.Wrapf(err, "injector configmap %v failed to parse 'config' data key", sets.Key(injectorCm))
	}

	return &cfg, nil
}

// source: https://github.com/istio/istio/blob/master/pilot/pkg/bootstrap/sidecarinjector.go
func getInjectorConfigMapName(revision string) string {
	name := defaultInjectorConfigMapName
	if revision == "" || revision == "default" {
		return name
	}
	return name + "-" + revision
}

// this code copied with modifications from https://github.com/istio/istio/blob/master/pkg/kube/inject/inject.go
func injectRequired(
	ctx context.Context,
	ignored []string,
	namespaceInjected bool,
	config *inject.Config,
	workload types.Workload,
) bool {
	if config == nil {
		// set default to off if no config found
		config = &inject.Config{Policy: inject.InjectionPolicyDisabled}
	}

	podSpec := workload.GetPodTemplate().Spec

	// Skip injection when host networking is enabled. The problem is
	// that the iptables changes are assumed to be within the pod when,
	// in fact, they are changing the routing at the host level. This
	// often results in routing failures within a node which can
	// affect the network provider within the cluster causing
	// additional pod failures.
	if podSpec.HostNetwork {
		return false
	}

	// skip special kubernetes system namespaces
	for _, namespace := range ignored {
		if workload.GetNamespace() == namespace {
			return false
		}
	}

	annos := workload.GetPodTemplate().ObjectMeta.Annotations
	if annos == nil {
		annos = map[string]string{}
	}

	var useDefault bool
	var workloadInjected bool
	switch strings.ToLower(annos[annotation.SidecarInject.Name]) {
	// http://yaml.org/type/bool.html
	case "y", "yes", "true", "on":
		workloadInjected = true
	case "":
		useDefault = true
	}

	// If an annotation is not explicitly given, check the LabelSelectors, starting with NeverInject
	if useDefault {
		for _, neverSelector := range config.NeverInjectSelector {
			selector, err := metav1.LabelSelectorAsSelector(&neverSelector)
			if err != nil {
				contextutils.LoggerFrom(ctx).Debugf("ignoring invalid injection selector for NeverInjectSelector: %v (%v)", neverSelector, err)
			} else if !selector.Empty() && selector.Matches(labels.Set(workload.GetLabels())) {
				// injection explicitly disabled for workload due to pod labels matching NeverInjectSelector config map entry
				workloadInjected = false
				useDefault = false
				break
			}
		}
	}

	// If there's no annotation nor a NeverInjectSelector, check the injection on the namespace, and the config's AlwaysInject selector
	if useDefault {
		if namespaceInjected {
			workloadInjected = true
			useDefault = false
		}
		for _, alwaysSelector := range config.AlwaysInjectSelector {
			selector, err := metav1.LabelSelectorAsSelector(&alwaysSelector)
			if err != nil {
				contextutils.LoggerFrom(ctx).Debugf("ignoring invalid injection selector for AlwaysInjectSelector: %v (%v)", alwaysSelector, err)
			} else if !selector.Empty() && selector.Matches(labels.Set(workload.GetLabels())) {
				// injection explicitly enabled for workload due to pod labels matching NeverInjectSelector config map entry
				workloadInjected = true
				useDefault = false
				break
			}
		}
	}

	var required bool
	switch config.Policy {
	case inject.InjectionPolicyEnabled:
		if useDefault {
			required = true
		} else {
			required = workloadInjected
		}
	default:
		if useDefault {
			required = false
		} else {
			required = workloadInjected
		}
	}

	return required
}
