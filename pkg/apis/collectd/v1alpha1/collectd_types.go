package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// CollectdSpec defines the desired state of Collectd
// +k8s:openapi-gen=true
type CollectdSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html
	DeploymentPlan DeploymentPlanType `json:"deploymentPlan,omitempty"`
}

// CollectdStatus defines the observed state of Collectd
// +k8s:openapi-gen=true
type CollectdStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html
	//PodNames []string `json:"pods"`
	//Plugins  []Plugin `json:"plugins"`
	condition string `json:"condition,omitempty"`
}

// DeploymentPlanType defines deployment spec
type DeploymentPlanType struct {
	Image string `json:"image,omitempty"`
	Size  int32  `json:"size,omitempty"`
}

// Plugin defines plugin enabled
type Plugin struct {
	Name    string `json:"name,omitempty"`
	Enabled bool   `json:"enabled,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Collectd is the Schema for the collectds API
// +k8s:openapi-gen=true
type Collectd struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CollectdSpec   `json:"spec,omitempty"`
	Status CollectdStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// CollectdList contains a list of Collectd
type CollectdList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Collectd `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Collectd{}, &CollectdList{})
}
