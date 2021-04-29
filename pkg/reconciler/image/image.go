package image

import (
	"context"

	"github.com/imjasonh/image/pkg/apis/image/v1alpha1"
	imagereconciler "github.com/imjasonh/image/pkg/client/injection/reconciler/image/v1alpha1/image"
	"knative.dev/pkg/logging"
	"knative.dev/pkg/reconciler"
	"knative.dev/pkg/tracker"
)

// Reconciler implements imagereconciler.Interface for Image resources.
type Reconciler struct {
	// Tracker builds an index of what resources are watching other resources
	// so that we can immediately react to changes tracked resources.
	Tracker tracker.Interface
}

// Check that our Reconciler implements Interface
var _ imagereconciler.Interface = (*Reconciler)(nil)

// ReconcileKind implements Interface.ReconcileKind.
func (r *Reconciler) ReconcileKind(ctx context.Context, o *v1alpha1.Image) reconciler.Event {
	logger := logging.FromContext(ctx)
	logger.Infof("Reconciling %s/%s", o.Namespace, o.Name)

	o.Status.Digest = "RESOLVED"
	return nil
}
