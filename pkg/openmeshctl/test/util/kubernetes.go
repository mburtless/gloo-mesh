package util

import (
	"github.com/solo-io/gloo-mesh/pkg/common/schemes"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/api/meta/testrestmapper"
	"k8s.io/apimachinery/pkg/runtime"
	fake_dynamic "k8s.io/client-go/dynamic/fake"
	"k8s.io/kubectl/pkg/scheme"
	fake_client "sigs.k8s.io/controller-runtime/pkg/client/fake"
)

// NewClientBuilder returns a client builder with the Gloo Mesh scheme
func NewClientBuilder() *fake_client.ClientBuilder {
	return fake_client.NewClientBuilder().WithScheme(buildScheme())
}

// NewRESTMapper returns a test rest mapper that can handle all the Gloo Mesh types.
func NewRESTMapper() meta.RESTMapper {
	return testrestmapper.TestOnlyStaticRESTMapper(buildScheme())
}

// NewDynamicClient returns a fake dynamic client with the schema loaded.
// NOTE: This client does not work with server side apply.
func NewDynamicClient(objs ...runtime.Object) *fake_dynamic.FakeDynamicClient {
	return fake_dynamic.NewSimpleDynamicClient(buildScheme(), objs...)
}

func buildScheme() *runtime.Scheme {
	scheme := scheme.Scheme
	if err := schemes.AddToScheme(scheme); err != nil {
		panic(err)
	}
	if err := apiextensionsv1beta1.AddToScheme(scheme); err != nil {
		panic(err)
	}

	return scheme
}
