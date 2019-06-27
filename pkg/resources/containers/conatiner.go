package containers

import (
	"os"
	"reflect"

	v1alpha1 "github.com/aneeshkp/barometer-operator/pkg/apis/collectd/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var (
	log = logf.Log.WithName("Collectd_Containers")
)

//CheckCollectdContainer ...
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

//ContainerForCollectd ...
func ContainerForCollectd(m *v1alpha1.Collectd, cmRevision string) corev1.Container {
	var image string
	if m.Spec.DeploymentPlan.Image != "" {
		image = m.Spec.DeploymentPlan.Image
	} else {
		image = os.Getenv("COLLECTD_IMAGE")
	}

	container := corev1.Container{
		Image: image,
		Name:  m.Name,
	}
	env := []corev1.EnvVar{
		{
			Name:  "CM_REVISION",
			Value: cmRevision,
		},
	}
	container.Env = env
	volumeMounts := []corev1.VolumeMount{}
	volumeMounts = append(volumeMounts, corev1.VolumeMount{
		Name:      m.Name,
		MountPath: "/opt/collectd/etc/",
	})

	container.VolumeMounts = volumeMounts
	return container

}

//DefaultContainerForCollectd  ...
func DefaultContainerForCollectd(m *v1alpha1.Collectd) corev1.Container {
	var image string
	if m.Spec.DeploymentPlan.Image != "" {
		image = m.Spec.DeploymentPlan.Image
	} else {
		image = os.Getenv("COLLECTD_IMAGE")
	}

	container := corev1.Container{
		Image: image,
		Name:  m.Name,
	}
	return container

}
