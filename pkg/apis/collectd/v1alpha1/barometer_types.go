package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

//PhaseType ...
type PhaseType string

// Const for phasetype
const (
	CollectdPhaseNone     PhaseType = ""
	CollectdPhaseCreating           = "Creating"
	CollectdPhaseRunning            = "Running"
	CollectdPhaseFailed             = "Failed"
)

// ConditionType ...
type ConditionType string

//Constant for COndition Type
const (
	CollectdConditionProvisioning ConditionType = "Provisioning"
	CollectdConditionDeployed     ConditionType = "Deployed"
	CollectdConditionScalingUp    ConditionType = "ScalingUp"
	CollectdConditionScalingDown  ConditionType = "ScalingDown"
	CollectdConditionUpgrading    ConditionType = "Upgrading"
)

//CollectdCondition ...
type CollectdCondition struct {
	Type           ConditionType `json:"type"`
	TransitionTime metav1.Time   `json:"transitionTime,omitempty"`
	Reason         string        `json:"reason,omitempty"`
}

// DeploymentPlanType defines deployment spec
type DeploymentPlanType struct {
	Image      string `json:"image,omitempty"`
	Size       int32  `json:"size,omitempty"`
	ConfigName string `json:"configname,omitempty"`
}

// BarometerSpec defines the desired state of Barometer
type BarometerSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
	DeploymentPlan DeploymentPlanType `json:"deploymentPlan,omitempty"`
}

// BarometerStatus defines the observed state of Barometer
type BarometerStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
	Phase     PhaseType `json:"phase,omitempty"`
	RevNumber string    `json:"revNumber,omitempty"`
	PodNames  []string  `json:"pods"`

	// Conditions keeps most recent interconnect conditions
	Conditions []CollectdCondition `json:"conditions"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Barometer is the Schema for the barometers API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=barometers,scope=Namespaced
type Barometer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BarometerSpec   `json:"spec,omitempty"`
	Status BarometerStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// BarometerList contains a list of Barometer
type BarometerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Barometer `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Barometer{}, &BarometerList{})
}
