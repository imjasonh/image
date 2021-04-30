package v1alpha1

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"knative.dev/pkg/apis"
)

var condSet = apis.NewLivingConditionSet()

// GetGroupVersionKind implements kmeta.OwnerRefable
func (*Image) GetGroupVersionKind() schema.GroupVersionKind {
	return SchemeGroupVersion.WithKind("Image")
}

// GetConditionSet retrieves the condition set for this resource. Implements the KRShaped interface.
func (i *Image) GetConditionSet() apis.ConditionSet {
	return condSet
}

// InitializeConditions sets the initial values to the conditions.
func (is *ImageStatus) InitializeConditions() {
	condSet.Manage(is).InitializeConditions()
}

func (is *ImageStatus) MarkServiceUnavailable(name string) {
	condSet.Manage(is).MarkFalse(
		ImageConditionReady,
		"ServiceUnavailable",
		"Service %q wasn't found.", name)
}

func (is *ImageStatus) MarkServiceAvailable() {
	condSet.Manage(is).MarkTrue(ImageConditionReady)
}
