package apply_test

import (
	"io"
	"testing/fstest"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/gloo-mesh/pkg/openmeshctl/commands/apply"
	mock "github.com/solo-io/gloo-mesh/pkg/openmeshctl/commands/apply/mocks"
	mock_resource "github.com/solo-io/gloo-mesh/pkg/openmeshctl/resource/mocks"
	"github.com/solo-io/gloo-mesh/pkg/openmeshctl/test/util"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

var _ = Describe("Apply Command", func() {
	var (
		ctrl    *gomock.Controller
		ctx     *mock.MockContext
		applier *mock_resource.MockApplier

		filenamesCall *gomock.Call
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		ctx = mock.NewMockContext(ctrl)
		applier = mock_resource.NewMockApplier(ctrl)

		ctx.EXPECT().Out().Return(io.Discard).AnyTimes()

		filenamesCall = ctx.EXPECT().Filenames()
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Describe("Applying resources", func() {
		var files = []struct {
			name string
			data []byte
			obj  *unstructured.Unstructured
		}{
			{"myapp/k8s/access_policy.yaml", util.AccessPolicyRaw(), util.AccessPolicyUnstructured()},
			{"myapp/k8s/traffic_policy.yaml", util.TrafficPolicyRaw(), util.TrafficPolicyUnstructured()},
			{"myapp/k8s/virtual_mesh.yaml", util.VirtualMeshRaw(), util.VirtualMeshUnstructured()},
		}

		When("the files are in the local filesystem", func() {
			BeforeEach(func() {
				fs := make(fstest.MapFS, len(files))
				fnames := make([]string, len(files))
				for i, file := range files {
					fnames[i] = file.name
					fs[file.name] = &fstest.MapFile{Data: file.data}
					ctx.EXPECT().Applier(util.GVK(file.obj)).Return(applier)
					applier.EXPECT().Apply(ctx, util.DiffEq(file.obj))
				}

				ctx.EXPECT().FS().Return(fs).AnyTimes()
				filenamesCall.Return(fnames)
			})

			It("should not error", func() {
				err := apply.Apply(ctx)
				Expect(err).ToNot(HaveOccurred())
			})
		})

		When("when the files are remote", func() {
			BeforeEach(func() {
				const host = "https://raw.githubusercontent.com/"

				endpoints := make(map[string][]byte, len(files))
				fnames := make([]string, len(files))
				for i, file := range files {
					uri := host + file.name
					fnames[i] = uri
					endpoints[uri] = file.data
					ctx.EXPECT().Applier(util.GVK(file.obj)).Return(applier)
					applier.EXPECT().Apply(ctx, util.DiffEq(file.obj))
				}

				ctx.EXPECT().HttpClient().Return(util.NewFakeHttpClient(endpoints)).AnyTimes()
				filenamesCall.Return(fnames)
			})

			It("should not error", func() {
				err := apply.Apply(ctx)
				Expect(err).ToNot(HaveOccurred())
			})
		})
	})
})
