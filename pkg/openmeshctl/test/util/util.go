package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/rest/fake"
)

// MustMarshalJSON marshals an object to and panics on error
func MustMarshalJSON(v interface{}) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}

	return b
}

// GVK extracts the GVK from various Kubernetes objects.
func GVK(obj interface{}) *schema.GroupVersionKind {
	var gvk schema.GroupVersionKind
	switch o := obj.(type) {
	case runtime.Object:
		gvk = o.GetObjectKind().GroupVersionKind()
	case *unstructured.Unstructured:
		gvk = o.GroupVersionKind()
	default:
		panic(fmt.Sprintf("cannot extract GVK from (%T)", obj))
	}

	return &gvk
}

var (
	okText       = http.StatusText(http.StatusOK)
	notFoundText = http.StatusText(http.StatusNotFound)
)

// NewFakeHttpClient returns a new fake HTTP client that returns 200 and the
// data for the endpoints and 404 for endpoints not listed.
func NewFakeHttpClient(endpoints map[string][]byte) *http.Client {
	return fake.CreateHTTPClient(func(r *http.Request) (*http.Response, error) {
		data, ok := endpoints[r.URL.String()]
		if !ok {
			return &http.Response{Status: notFoundText, StatusCode: http.StatusNotFound}, nil
		}

		return &http.Response{
			Status:     okText,
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewReader(data)),
		}, nil
	})
}
