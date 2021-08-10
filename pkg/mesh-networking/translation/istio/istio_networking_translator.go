package istio

import (
	"context"
	"fmt"

	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/istio/mesh"

	discoveryv1 "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1"
	"github.com/solo-io/skv2/contrib/pkg/sets"
	"github.com/solo-io/skv2/pkg/ezkube"

	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/extensions"

	"github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/input"
	"github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/output/istio"
	"github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/output/local"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/reporting"
	istioextensions "github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/istio/extensions"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/istio/internal"
	"github.com/solo-io/go-utils/contextutils"
)

// the istio translator translates an input networking snapshot to an output snapshot of Istio resources
type Translator interface {
	// Translate translates the appropriate resources to apply input configuration resources for all Istio meshes contained in the input snapshot.
	// Output resources will be added to the output.Builder
	// Errors caused by invalid user config will be reported using the Reporter.
	Translate(
		ctx context.Context,
		in input.LocalSnapshot,
		userSupplied input.RemoteSnapshot,
		istioOutputs istio.Builder,
		localOutputs local.Builder,
		reporter reporting.Reporter,
	)
}

type istioTranslator struct {
	totalTranslates int // TODO(ilackarms): metric

	// note: these interfaces are set directly in unit tests, but not exposed in the Translator's constructor
	dependencies internal.DependencyFactory
	extender     istioextensions.IstioExtender

	// we preserve outputs from each translation in order to preserve
	// last known good config when errors occur
	translationOutputsCache *preservedTranslationOutputs
}

func NewIstioTranslator(extensionClients extensions.Clientset) Translator {
	return &istioTranslator{
		dependencies:            internal.NewDependencyFactory(),
		extender:                istioextensions.NewIstioExtender(extensionClients),
		translationOutputsCache: newPreservedTranslationOutputs(),
	}
}

type preservedTranslationOutputs struct {
	// map of mesh id to the outputs for that mesh
	meshOutputs map[string]meshOutputs
}

func newPreservedTranslationOutputs() *preservedTranslationOutputs {
	return &preservedTranslationOutputs{
		meshOutputs: map[string]meshOutputs{},
	}
}

type meshOutputs struct {
	remote istio.Builder
	local  local.Builder
}

func (t *istioTranslator) Translate(
	ctx context.Context,
	in input.LocalSnapshot,
	userSupplied input.RemoteSnapshot,
	istioOutputs istio.Builder,
	localOutputs local.Builder,
	reporter reporting.Reporter,
) {
	ctx = contextutils.WithLogger(ctx, fmt.Sprintf("istio-translator-%v", t.totalTranslates))

	destinationTranslator := t.dependencies.MakeDestinationTranslator(
		ctx,
		userSupplied,
		in.KubernetesClusters(),
		in.Destinations(),
	)

	for _, destination := range in.Destinations().List() {
		destinationTranslator.Translate(in, destination, istioOutputs, reporter)
	}

	meshTranslator := t.dependencies.MakeMeshTranslator(
		ctx,
		in.Secrets(),
		in.Workloads(),
	)

	for _, mesh := range in.Meshes().List() {
		perMeshOutputs, perMeshLocalOutputs := t.translateMesh(
			ctx,
			in,
			mesh,
			meshTranslator,
			reporter,
		)
		// merge per-mesh outputs to translator's whole outputs
		istioOutputs.Merge(perMeshOutputs)
		localOutputs.Merge(perMeshLocalOutputs)
	}

	if err := t.extender.PatchOutputs(ctx, in, istioOutputs); err != nil {
		// TODO(ilackarms): consider providing/checking user option to fail here when the extender server is unavailable.
		// currently we just log the error and continue.
		contextutils.LoggerFrom(ctx).Errorf("failed to apply extension patches: %v", err)
	}

	t.totalTranslates++
}

func (t *istioTranslator) translateMesh(
	ctx context.Context,
	in input.LocalSnapshot,
	mesh *discoveryv1.Mesh,
	meshTranslator mesh.Translator,
	reporter reporting.Reporter,
) (istio.Builder, local.Builder) {
	// create a set of per-mesh outputs
	perMeshOutputs := istio.Builder(istio.NewBuilder(ctx, "mesh-outputs"))
	perMeshLocalOutputs := local.Builder(local.NewBuilder(ctx, "mesh-local-outputs"))

	// intercept reports that the vmesh is invalid
	interceptingReporter := &reportInterceptor{Reporter: reporter}

	// perform translation
	meshTranslator.Translate(in, mesh, perMeshOutputs, perMeshLocalOutputs, interceptingReporter)

	if interceptingReporter.meshInvalid {
		previousOutputs, ok := t.translationOutputsCache.meshOutputs[sets.Key(mesh)]
		if ok {
			// restore last known config if mesh is invalid
			perMeshOutputs = previousOutputs.remote
			perMeshLocalOutputs = previousOutputs.local
		}
	} else {
		// record new config as valid
		t.translationOutputsCache.meshOutputs[sets.Key(mesh)] = meshOutputs{
			remote: perMeshOutputs,
			local:  perMeshLocalOutputs,
		}
	}

	return perMeshOutputs, perMeshLocalOutputs
}

// use this intercepting reporter to catch when a mesh is invalid
type reportInterceptor struct {
	reporting.Reporter
	meshInvalid bool
}

// intercept and record reports to a mesh
func (r *reportInterceptor) ReportVirtualMeshToMesh(mesh *discoveryv1.Mesh, virtualMesh ezkube.ResourceId, err error) {
	r.meshInvalid = true
	r.Reporter.ReportVirtualMeshToMesh(mesh, virtualMesh, err)
}
