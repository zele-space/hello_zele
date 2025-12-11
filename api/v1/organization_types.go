package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// OrganizationSpec defines the desired state of Organization.
type OrganizationSpec struct {
	// Namespace is the target namespace that the organization controls.
	Namespace string `json:"namespace,omitempty"`
	// Labels are propagated to the managed namespace.
	Labels map[string]string `json:"labels,omitempty"`
}

// OrganizationStatus defines the observed state of Organization.
type OrganizationStatus struct {
	// NamespaceCreated indicates whether the namespace exists.
	NamespaceCreated bool `json:"namespaceCreated,omitempty"`
	// ObservedNamespace is the namespace currently managed.
	ObservedNamespace string `json:"observedNamespace,omitempty"`
	// Conditions represent the latest available observations.
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Organization is the Schema for the organizations API.
type Organization struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   OrganizationSpec   `json:"spec,omitempty"`
	Status OrganizationStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// OrganizationList contains a list of Organization.
type OrganizationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Organization `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Organization{}, &OrganizationList{})
}
