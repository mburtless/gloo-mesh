package checks

import (
	"context"
	"fmt"

	extbeta1 "github.com/solo-io/external-apis/pkg/api/k8s/apiextensions.k8s.io/v1beta1"
	"github.com/solo-io/skv2/pkg/crdutils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type crdUpgradeCheck struct{ deployment string }

func NewCrdUpgradeCheck(deployment string) Check {
	return &crdUpgradeCheck{deployment: deployment}
}

func (c *crdUpgradeCheck) GetDescription() string {
	return "Gloo Mesh networking CRDs for are up-to-date for this version"
}

func (c *crdUpgradeCheck) Run(ctx context.Context, checkCtx CheckContext) *Result {
	cli := checkCtx.Client()
	if cli == nil {
		// no k8s client, so can't do anything
		return &Result{}
	}
	crdClient := extbeta1.NewCustomResourceDefinitionClient(cli)
	ls, err := metav1.LabelSelectorAsSelector(&metav1.LabelSelector{
		MatchLabels: map[string]string{"app": "gloo-mesh"},
	})
	if err != nil {
		return new(Result).AddError(err)
	}

	lo := &client.ListOptions{
		LabelSelector: ls,
	}
	gmCrds, err := crdClient.ListCustomResourceDefinition(ctx, lo)
	if err != nil {
		return new(Result).AddError(err)
	}
	md, err := checkCtx.CRDMetadata(ctx, c.deployment)
	if err != nil {
		return new(Result).AddError(err)
	}
	if md == nil {
		return new(Result).AddHint("No CRD metadata present - can't perform upgrade checks. This feature is available on gloo-mesh 1.1 or higher", "")
	}

	var result Result
	errMap := crdutils.DoCrdsNeedUpgrade(*md, gmCrds.Items)
	// go over all the errors and see if we have anything interesting:
	var upgradeHintAdded bool
	var notFoundHintAdded bool
	for name, err := range errMap {
		switch err.(type) {
		case *crdutils.CrdNeedsUpgrade:
			result.AddError(err)
			if !upgradeHintAdded {
				upgradeHintAdded = true
				result.AddHint("One or more CRD spec has changed. Upgrading your Gloo-Mesh CRDs may be required before continuing.", "")
			}

		case *crdutils.CrdNotFound:
			result.AddError(err)
			if !notFoundHintAdded {
				notFoundHintAdded = true
				result.AddHint("One or more required CRD were not found on the cluster. Please verify Gloo-Mesh CRDs are installed.", "")
			}
		default:
			fmt.Printf("Unknown error validating CRD %s: %s\n", name, err.Error())
		}
	}

	return &result
}
