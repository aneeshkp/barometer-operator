package deployments

import (
	v1alpha1 "github.com/aneeshkp/barometer-operator/pkg/apis/collectd/v1alpha1"
	"github.com/aneeshkp/barometer-operator/pkg/resources/containers"
	"github.com/aneeshkp/barometer-operator/pkg/utils/selectors"
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

//NewDefaultDaemonSetForCR  Create default
func NewDefaultDaemonSetForCR(m *v1alpha1.Barometer) *appsv1.DaemonSet {
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
					Annotations: map[string]string{
						"cmversion": "",
					},
				},
				Spec: corev1.PodSpec{
					HostNetwork:        true,
					ServiceAccountName: "barometer-operator",
					Containers:         []corev1.Container{containers.DefaultContainerForCollectd(m)},
				},
			},
		},
	}

	return ds
}

//NewDaemonSetForCR Create NewDaemonSetForCR method to create daemonset
func NewDaemonSetForCR(m *v1alpha1.Barometer, cmVersion string) *appsv1.DaemonSet {
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
					Annotations: map[string]string{
						"cmversion": "",
					},
				},
				Spec: corev1.PodSpec{
					HostNetwork:        true,
					ServiceAccountName: "barometer-operator",
					Containers:         []corev1.Container{containers.ContainerForCollectd(m, cmVersion)},
				},
			},
		},
	}
	volumes := []corev1.Volume{
		{
			Name: m.Name,
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: m.Spec.DeploymentPlan.ConfigName,
					},
					Items: []corev1.KeyToPath{
						{
							Key:  "node.barometer.config",
							Path: "collectd.conf",
						},
					},
				},
			},
		},
	}

	ds.Spec.Template.Spec.Volumes = volumes

	return ds
}
