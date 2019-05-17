package containers

import (
	"os"
	"reflect"

	v1alpha1 "github.com/aneeshkp/collectd-operator/pkg/apis/collectd/v1alpha1"
	"github.com/aneeshkp/smartgateway-operator/pkg/constants"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var (
	log = logf.Log.WithName("containers")
)

func containerEnvVarsForCollectd(m *v1alpha1.Collectd) []corev1.EnvVar {
	envVars := []corev1.EnvVar{}
	envVars = append(envVars, corev1.EnvVar{Name: "APPLICATION_NAME", Value: m.Name})
	envVars = append(envVars, corev1.EnvVar{Name: "COLLECTD_CONF", Value: "/etc/collectd/collectd.conf.template"})
	envVars = append(envVars, corev1.EnvVar{Name: "POD_NAMESPACE", ValueFrom: &corev1.EnvVarSource{
		FieldRef: &corev1.ObjectFieldSelector{
			FieldPath: "metadata.namespace",
		},
	},
	})
	envVars = append(envVars, corev1.EnvVar{Name: "POD_IP", ValueFrom: &corev1.EnvVarSource{
		FieldRef: &corev1.ObjectFieldSelector{
			FieldPath: "status.podIP",
		},
	},
	})

	return envVars
}

func CheckCollectdContainer(desired *corev1.Container, actual *corev1.Container) bool {
	if desired.Image != actual.Image {
		return false
	}
	if !reflect.DeepEqual(desired.Env, actual.Env) {
		return false
	}
	if !reflect.DeepEqual(desired.Ports, actual.Ports) {
		return false
	}
	if !reflect.DeepEqual(desired.VolumeMounts, actual.VolumeMounts) {
		return false
	}
	return true
}

func ContainerForCollectd(m *v1alpha1.Collectd) corev1.Container {
	var image string
	if m.Spec.DeploymentPlan.Image != "" {
		image = m.Spec.DeploymentPlan.Image
	} else {
		image = os.Getenv("COLLECTD_IMAGE")
	}
	container := corev1.Container{
		Image: image,
		Name:  m.Name,
		LivenessProbe: &corev1.Probe{
			InitialDelaySeconds: 60,
			Handler: corev1.Handler{
				HTTPGet: &corev1.HTTPGetAction{
					Port: intstr.FromInt(constants.HttpLivenessPort),
				},
			},
		},
		Env:   containerEnvVarsForCollectd(m)
	}
	volumeMounts := []corev1.VolumeMount{}
	volumeMounts = append(volumeMounts, corev1.VolumeMount{
		Name:      m.Name,
		MountPath: "/etc/qpid-dispatch/",
	})

	container.VolumeMounts = volumeMounts
	container.Resources = m.Spec.DeploymentPlan.Resources
	return container
}
