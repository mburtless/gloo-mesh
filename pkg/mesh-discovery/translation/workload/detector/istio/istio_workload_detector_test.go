package istio

import (
	"context"

	"github.com/ghodss/yaml"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1sets "github.com/solo-io/external-apis/pkg/api/k8s/core/v1/sets"
	v1 "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1"
	v1sets "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1/sets"
	"github.com/solo-io/gloo-mesh/pkg/mesh-discovery/translation/workload/types"
	"istio.io/api/annotation"
	"istio.io/istio/galley/pkg/config/analysis/analyzers/injection"
	"istio.io/istio/galley/pkg/config/analysis/analyzers/util"
	"istio.io/istio/pkg/kube/inject"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("IstioWorkloadDetector", func() {
	var (
		clusterName       = "cluster"
		workloadName      = "workload"
		workloadNamespace = "workload-namespace"
		istioNamespace    = "istio-namespace"

		sidecarConfigMap = func(config inject.Config) *corev1.ConfigMap {
			b, err := yaml.Marshal(config)
			Expect(err).NotTo(HaveOccurred())

			return &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:        defaultInjectorConfigMapName,
					Namespace:   istioNamespace,
					ClusterName: clusterName,
				},
				Data: map[string]string{
					injectorConfigMapKey: string(b),
				},
			}
		}

		mesh = &v1.Mesh{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "gloo-mesh",
				Name:      "my-istio",
			},
			Spec: v1.MeshSpec{
				Type: &v1.MeshSpec_Istio_{
					Istio: &v1.MeshSpec_Istio{
						Installation: &v1.MeshInstallation{
							Cluster:   clusterName,
							Namespace: istioNamespace,
						},
					},
				},
			},
		}
	)

	It("detects injected sidecar workloads that are annotated for injection", func() {
		meshes := v1sets.NewMeshSet(mesh)
		namespace := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name:        workloadNamespace,
				ClusterName: clusterName,
			},
		}
		detector := NewWorkloadDetector(
			context.TODO(),
			corev1sets.NewNamespaceSet(namespace),
			corev1sets.NewConfigMapSet(sidecarConfigMap(inject.Config{})),
		)

		workload := &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:   workloadNamespace,
				Name:        workloadName,
				ClusterName: clusterName,
			},
			Spec: appsv1.DeploymentSpec{
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{
							annotation.SidecarInject.Name: "true",
						},
					},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Image: "some-image",
							},
						},
					},
				},
			},
		}

		mesh := detector.DetectMeshForWorkload(types.ToWorkload(workload), meshes)
		Expect(mesh).To(Equal(meshes.List()[0]))
	})

	It("detects injected sidecar workloads that are in a namespace with legacy label for injection", func() {

		meshes := v1sets.NewMeshSet(mesh)
		namespace := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name:        workloadNamespace,
				ClusterName: clusterName,
				Labels: map[string]string{
					util.InjectionLabelName: util.InjectionLabelEnableValue,
				},
			},
		}
		detector := NewWorkloadDetector(
			context.TODO(),
			corev1sets.NewNamespaceSet(namespace),
			corev1sets.NewConfigMapSet(sidecarConfigMap(inject.Config{})),
		)

		workload := &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:   workloadNamespace,
				Name:        workloadName,
				ClusterName: clusterName,
			},
			Spec: appsv1.DeploymentSpec{
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Image: "some-image",
							},
						},
					},
				},
			},
		}

		mesh := detector.DetectMeshForWorkload(types.ToWorkload(workload), meshes)
		Expect(mesh).To(Equal(meshes.List()[0]))
	})
	It("detects injected sidecar workloads that are in a namespace with new label for injection", func() {

		meshes := v1sets.NewMeshSet(mesh)
		namespace := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name:        workloadNamespace,
				ClusterName: clusterName,
				Labels: map[string]string{
					injection.RevisionInjectionLabelName: "revision-1234",
				},
			},
		}
		detector := NewWorkloadDetector(
			context.TODO(),
			corev1sets.NewNamespaceSet(namespace),
			corev1sets.NewConfigMapSet(sidecarConfigMap(inject.Config{})),
		)

		workload := &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:   workloadNamespace,
				Name:        workloadName,
				ClusterName: clusterName,
			},
			Spec: appsv1.DeploymentSpec{
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Image: "some-image",
							},
						},
					},
				},
			},
		}

		mesh := detector.DetectMeshForWorkload(types.ToWorkload(workload), meshes)
		Expect(mesh).To(Equal(meshes.List()[0]))

	})
	It("does not detect injected sidecar workloads that are not labeled for injection", func() {

		meshes := v1sets.NewMeshSet(mesh)
		namespace := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name:        workloadNamespace,
				ClusterName: clusterName,
			},
		}
		detector := NewWorkloadDetector(
			context.TODO(),
			corev1sets.NewNamespaceSet(namespace),
			corev1sets.NewConfigMapSet(sidecarConfigMap(inject.Config{})),
		)

		workload := &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:   workloadNamespace,
				Name:        workloadName,
				ClusterName: clusterName,
			},
			Spec: appsv1.DeploymentSpec{
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Image: "some-image",
							},
						},
					},
				},
			},
		}

		mesh := detector.DetectMeshForWorkload(types.ToWorkload(workload), meshes)
		Expect(mesh).To(BeNil())

	})
	It("does not detect injected sidecar workloads that are in a namespace labeled for injection when the workload is labeled to skip injection", func() {

		meshes := v1sets.NewMeshSet(mesh)
		namespace := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name:        workloadNamespace,
				ClusterName: clusterName,
				Labels: map[string]string{
					injection.RevisionInjectionLabelName: "revision-1234",
				},
			},
		}
		detector := NewWorkloadDetector(
			context.TODO(),
			corev1sets.NewNamespaceSet(namespace),
			corev1sets.NewConfigMapSet(sidecarConfigMap(inject.Config{})),
		)

		workload := &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:   workloadNamespace,
				Name:        workloadName,
				ClusterName: clusterName,
			},
			Spec: appsv1.DeploymentSpec{
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{
							annotation.SidecarInject.Name: "false",
						},
					},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Image: "some-image",
							},
						},
					},
				},
			},
		}

		mesh := detector.DetectMeshForWorkload(types.ToWorkload(workload), meshes)
		Expect(mesh).To(BeNil())

	})
	It("does not detect injected sidecar workloads that are in a namespace with legacy label for injection when the workload is labeled to skip injection", func() {

		meshes := v1sets.NewMeshSet(mesh)
		namespace := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name:        workloadNamespace,
				ClusterName: clusterName,
				Labels: map[string]string{
					util.InjectionLabelName: util.InjectionLabelEnableValue,
				},
			},
		}
		detector := NewWorkloadDetector(
			context.TODO(),
			corev1sets.NewNamespaceSet(namespace),
			corev1sets.NewConfigMapSet(sidecarConfigMap(inject.Config{})),
		)

		workload := &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:   workloadNamespace,
				Name:        workloadName,
				ClusterName: clusterName,
			},
			Spec: appsv1.DeploymentSpec{
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{
							annotation.SidecarInject.Name: "false",
						},
					},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Image: "some-image",
							},
						},
					},
				},
			},
		}

		mesh := detector.DetectMeshForWorkload(types.ToWorkload(workload), meshes)
		Expect(mesh).To(BeNil())

	})
	It("detects workloads with a proxy container already in the pod template", func() {

		meshes := v1sets.NewMeshSet(mesh)
		namespace := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name:        workloadNamespace,
				ClusterName: clusterName,
			},
		}
		detector := NewWorkloadDetector(
			context.TODO(),
			corev1sets.NewNamespaceSet(namespace),
			corev1sets.NewConfigMapSet(sidecarConfigMap(inject.Config{})),
		)

		workload := &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:   workloadNamespace,
				Name:        workloadName,
				ClusterName: clusterName,
			},
			Spec: appsv1.DeploymentSpec{
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name: inject.ProxyContainerName,
							},
						},
					},
				},
			},
		}

		mesh := detector.DetectMeshForWorkload(types.ToWorkload(workload), meshes)
		Expect(mesh).To(Equal(meshes.List()[0]))
	})
})
