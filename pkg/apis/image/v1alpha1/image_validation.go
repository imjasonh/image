package v1alpha1

import (
	"context"
	"fmt"

	"github.com/google/go-containerregistry/pkg/name"
	"knative.dev/pkg/apis"
)

// Validate implements apis.Validatable
func (i *Image) Validate(ctx context.Context) *apis.FieldError {
	return i.Spec.Validate(ctx).ViaField("spec")
}

// Validate implements apis.Validatable
func (is *ImageSpec) Validate(ctx context.Context) *apis.FieldError {
	if _, err := name.ParseReference(is.Ref); err != nil {
		return apis.ErrInvalidValue(fmt.Sprintf("parsing %q: %v", is.Ref, err), ".ref")
	}
	return nil
}
