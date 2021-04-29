package v1alpha1

import (
	"context"

	"knative.dev/pkg/apis"
)

// Validate implements apis.Validatable
func (i *Image) Validate(ctx context.Context) *apis.FieldError {
	return i.Spec.Validate(ctx).ViaField("spec")
}

// Validate implements apis.Validatable
func (is *ImageSpec) Validate(ctx context.Context) *apis.FieldError {
	return nil
}
