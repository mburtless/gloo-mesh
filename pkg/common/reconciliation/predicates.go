package reconciliation

import (
	"github.com/solo-io/skv2/pkg/predicate"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// predicate which filters objects (configmaps) used for leader election
var FilterLeaderElectionObject = predicate.SimplePredicate{
	Filter: predicate.SimpleEventFilterFunc(func(obj client.Object) bool {
		_, isLeaderElectionObj := obj.GetAnnotations()["control-plane.alpha.kubernetes.io/leader"]
		return isLeaderElectionObj
	}),
}

// predicate which filters service account token secrets
var FilterServiceAccountTokenSecret = predicate.SimplePredicate{
	Filter: predicate.SimpleEventFilterFunc(func(obj client.Object) bool {
		sec, ok := obj.(*v1.Secret)
		if !ok {
			return false
		}
		return sec.Type == v1.SecretTypeServiceAccountToken
	}),
}

// predicate which filters configmaps in kube-system
var FilterKubeSystemConfigMap = predicate.SimplePredicate{
	Filter: predicate.SimpleEventFilterFunc(func(obj client.Object) bool {
		_, ok := obj.(*v1.ConfigMap)
		if !ok {
			return false
		}
		return obj.GetNamespace() == "kube-system"
	}),
}
