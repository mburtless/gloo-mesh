package utils

import (
	"context"
	"time"

	"github.com/rotisserie/eris"
	"github.com/solo-io/external-apis/pkg/api/k8s/apiextensions.k8s.io/v1beta1"
	k8s_v1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// block until CRD is in the "established" phase to prevent subsequent race conditions when attempting to create CRs
func WaitUntilCRDsEstablished(ctx context.Context, kubeClient client.Client, timeout time.Duration, crdNames []string) error {
	failed := time.After(timeout)
	notYetEstablished := make(map[string]struct{})
	for {
		select {
		case <-failed:
			return eris.Errorf("timed out waiting for crds to be established: %v", notYetEstablished)
		case <-time.After(time.Second / 2):
			notYetEstablished = make(map[string]struct{})
			for _, crd := range crdNames {
				ready, err := crdEstablished(ctx, kubeClient, crd)
				if err != nil {
					notYetEstablished[crd] = struct{}{}
				}
				if !ready {
					notYetEstablished[crd] = struct{}{}
				}
			}
			if len(notYetEstablished) == 0 {
				return nil
			}
		}
	}
}

func crdEstablished(ctx context.Context, kubeClient client.Client, crdName string) (bool, error) {
	crdClient := v1beta1.NewCustomResourceDefinitionClient(kubeClient)
	existingCrd, err := crdClient.GetCustomResourceDefinition(ctx, crdName)
	if err != nil {
		return false, err
	}
	for _, cond := range existingCrd.Status.Conditions {
		if cond.Type == k8s_v1beta1.Established && cond.Status == k8s_v1beta1.ConditionTrue {
			return true, nil
		}
	}
	return false, nil
}
