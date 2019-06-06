package configmaps

import (
	v1alpha1 "github.com/aneeshkp/barometer-operator/pkg/apis/collectd/v1alpha1"
	"github.com/aneeshkp/barometer-operator/pkg/utils/configs"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Create NewConfigMapForCR method to create configmap
func NewConfigMapForCR(m *v1alpha1.Collectd) *corev1.ConfigMap {
	config := configs.ConfigForCollectd(m)
	configMap := &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "ConfigMap",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.Spec.DeploymentPlan.ConfigName,
			Namespace: m.Namespace,
		},
		Data: map[string]string{
			"config": config,
		},
	}

	return configMap
}
