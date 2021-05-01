package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/apis"
	duckv1 "knative.dev/pkg/apis/duck/v1"
	"knative.dev/pkg/kmeta"
)

// Image is a Knative abstraction that encapsulates the interface by which Knative
// components express a desire to have a particular image cached.
//
// +genclient
// +genreconciler
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type Image struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec holds the desired state of the Image (from the client).
	// +optional
	Spec ImageSpec `json:"spec,omitempty"`

	// Status communicates the observed state of the Image (from the controller).
	// +optional
	Status ImageStatus `json:"status,omitempty"`
}

var (
	// Check that Image can be validated and defaulted.
	_ apis.Validatable   = (*Image)(nil)
	_ apis.Defaultable   = (*Image)(nil)
	_ kmeta.OwnerRefable = (*Image)(nil)
	// Check that the type conforms to the duck Knative Resource shape.
	_ duckv1.KRShaped = (*Image)(nil)
)

// ImageSpec holds the desired state of the Image (from the client).
type ImageSpec struct {
	// Ref is the image reference in the external regsitry.
	Ref string `json:"ref"`
}

const (
	// ImageConditionReady is set when the revision is starting to materialize
	// runtime resources, and becomes true when those resources are ready.
	ImageConditionReady = apis.ConditionReady
)

// ImageStatus communicates the observed state of the Image (from the controller).
type ImageStatus struct {
	duckv1.Status `json:",inline"`

	// +optional
	Digest string `json:"digest,omitempty"`
}

// ImageList is a list of Image resources
//
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type ImageList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []Image `json:"items"`
}

// GetStatus retrieves the status of the resource. Implements the KRShaped interface.
func (i *Image) GetStatus() *duckv1.Status {
	return &i.Status.Status
}
