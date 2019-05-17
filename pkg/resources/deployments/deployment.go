package deployments

import (
	v1alpha1 "github.com/aneeshkp/collectd-operator/pkg/apis/collectd/v1alpha1"
	"github.com/aneeshkp/collectd-operator/pkg/resources/containers"
	"github.com/aneeshkp/collectd-operator/pkg/utils/selectors"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// move this to util
// Set labels in a map
func labelsForCollectd(name string) map[string]string {
	return map[string]string{
		selectors.LabelAppKey:      name,
		selectors.LabelResourceKey: name,
	}
}

// Create NewDeploymentForCR method to create deployment
func NewDeploymentForCR(m *v1alpha1.Collectd) *appsv1.Deployment {
	labels := selectors.LabelsForCollectd(m.Name)
	replicas := m.Spec.DeploymentPlan.Size
	affinity := &corev1.Affinity{}
	if m.Spec.DeploymentPlan.Placement == v1alpha1.PlacementAntiAffinity {
		affinity = &corev1.Affinity{
			PodAntiAffinity: &corev1.PodAntiAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{
					{
						LabelSelector: &metav1.LabelSelector{
							MatchExpressions: []metav1.LabelSelectorRequirement{
								{
									Key:      "application",
									Operator: metav1.LabelSelectorOpIn,
									Values:   []string{m.Name},
								},
							},
						},
						TopologyKey: "kubernetes.io/hostname",
					},
				},
			},
		}
	}
	dep := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.Name,
			Namespace: m.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: m.Name,
					Affinity:           affinity,
					Containers:         []corev1.Container{containers.ContainerForCollectd(m)},
				},
			},
		},
	}
	volumes := []corev1.Volume{}
	volumes = append(volumes, corev1.Volume{
		Name: m.Name,
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: m.Name,
				},
			},
		},
	})
	for _, profile := range m.Spec.SslProfiles {
		if len(profile.Credentials) > 0 {
			volumes = append(volumes, corev1.Volume{
				Name: profile.Credentials,
				VolumeSource: corev1.VolumeSource{
					Secret: &corev1.SecretVolumeSource{
						SecretName: profile.Credentials,
					},
				},
			})
		}
		if len(profile.CaCert) > 0 && profile.CaCert != profile.Credentials {
			volumes = append(volumes, corev1.Volume{
				Name: profile.CaCert,
				VolumeSource: corev1.VolumeSource{
					Secret: &corev1.SecretVolumeSource{
						SecretName: profile.CaCert,
					},
				},
			})
		}
	}
	dep.Spec.Template.Spec.Volumes = volumes

	return dep
}

// Create NewDaemonSetForCR method to create daemonset
func NewDaemonSetForCR(m *v1alpha1.Collectd) *appsv1.DaemonSet {
	labels := selectors.LabelsForCollectd(m.Name)

	ds := &appsv1.DaemonSet{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "DaemonSet",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.Name,
			Namespace: m.Namespace,
		},
		Spec: appsv1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: m.Name,
					Containers:         []corev1.Container{containers.ContainerForCollectd(m)},
				},
			},
		},
	}
	volumes := []corev1.Volume{}
	volumes = append(volumes, corev1.Volume{
		Name: m.Name,
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: m.Name,
				},
			},
		},
	})
	for _, profile := range m.Spec.SslProfiles {
		if len(profile.Credentials) > 0 {
			volumes = append(volumes, corev1.Volume{
				Name: profile.Credentials,
				VolumeSource: corev1.VolumeSource{
					Secret: &corev1.SecretVolumeSource{
						SecretName: profile.Credentials,
					},
				},
			})
		}
		if len(profile.CaCert) > 0 && profile.CaCert != profile.Credentials {
			volumes = append(volumes, corev1.Volume{
				Name: profile.CaCert,
				VolumeSource: corev1.VolumeSource{
					Secret: &corev1.SecretVolumeSource{
						SecretName: profile.CaCert,
					},
				},
			})
		}
	}
	ds.Spec.Template.Spec.Volumes = volumes

	return ds
}
