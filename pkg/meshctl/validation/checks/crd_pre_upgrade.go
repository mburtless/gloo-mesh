package checks

import (
	"context"
	"fmt"

	"github.com/solo-io/skv2/pkg/crdutils"
)

type crdUpgradeCheck struct{ deployment string }

func NewCrdUpgradeCheck(deployment string) Check {
	return &crdUpgradeCheck{deployment: deployment}
}

func (c *crdUpgradeCheck) GetDescription() string {
	return "Gloo Mesh CRD Versions"
}

func (c *crdUpgradeCheck) Run(ctx context.Context, checkCtx CheckContext) *Result {
	clusterCrds, err := checkCtx.Context().CrdClient.ListCustomResourceDefinition(ctx)
	if err != nil {
		return new(Result).AddError(err)
	}
	deploymentCrdMetadata, err := checkCtx.CRDMetadata(ctx, c.deployment)
	if err != nil {
		return new(Result).AddError(err)
	}
	if deploymentCrdMetadata == nil {
		return new(Result).AddHint("No CRD metadata present - can't perform upgrade checks. This feature is available on gloo-mesh 1.1 or higher", "")
	}

	var result Result
	errMap := crdutils.DoCrdsNeedUpgrade(*deploymentCrdMetadata, clusterCrds.Items)
	// go over all the errors and see if we have anything interesting:
	var upgradeHintAdded bool
	for name, err := range errMap {
		switch err.(type) {
		case *crdutils.CrdNeedsUpgrade:
			result.AddError(err)
			if !upgradeHintAdded {
				upgradeHintAdded = true
				result.AddHint("One or more CRD spec has changed. Upgrading your Gloo-Mesh CRDs may be required before continuing.", "")
			}
		case *crdutils.CrdNotFound:
			// CRD not found is benign
			result.AddHint(fmt.Sprintf("CRD %s not present on the cluster, ignore this warning if performing a first time install.", name), "")
		default:
			fmt.Printf("Unknown error validating CRD %s: %s\n", name, err.Error())
		}
	}

	return &result
}
