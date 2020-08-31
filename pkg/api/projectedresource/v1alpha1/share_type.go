package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Share is the Schema for the shares API
type Share struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ShareSpec   `json:"spec,omitempty"`
	Status ShareStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ShareList contains a list of Share
type ShareList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Share `json:"items"`
}

// ShareSpec defines the desired state of Share
type ShareSpec struct {
	// BackingResource captures the ConfigMap or Secret that is shared.
	BackingResource `json:"backingResource,omitempty"`

	// Description is a user readable explanation of what the backing resource
	// provides.
	Description string `json:"description,omitempty"`
}

// ShareStatus defines the observed state of Share
type ShareStatus struct {
	// Conditions are the set of k8s Condition instances provided by the associated controller for Shares.
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

type BackingResource struct {
	// Kind is a string value representing the REST resource this object represents.
	// Currently only Secret and ConfigMap are accepted.
	// +optional
	Kind string `json:"kind,omitempty"`

	// APIVersion defines the versioned schema of this representation of an object.
	// +optional
	APIVersion string `json:"apiVersion,omitempty"`

	// Name is the name of the object serving as the backing resource
	// +optional
	Name string `json:"name,omitempty"`

	// Namespace is the namespace of the object serving as the backing resource
	// +optional
	Namespace string `json:"namespace,omitempty"`
}
