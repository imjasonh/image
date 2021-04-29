package image

import (
	"context"

	imageinformer "github.com/imjasonh/image/pkg/client/injection/informers/image/v1alpha1/image"
	imagereconciler "github.com/imjasonh/image/pkg/client/injection/reconciler/image/v1alpha1/image"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/logging"
	"knative.dev/pkg/tracker"
)

// NewController creates a Reconciler and returns the result of NewImpl.
func NewController(
	ctx context.Context,
	cmw configmap.Watcher,
) *controller.Impl {
	logger := logging.FromContext(ctx)

	r := &Reconciler{}
	impl := imagereconciler.NewImpl(ctx, r)
	r.Tracker = tracker.New(impl.EnqueueKey, controller.GetTrackerLease(ctx))

	logger.Info("Setting up event handlers.")

	imageInformer := imageinformer.Get(ctx)
	imageInformer.Informer().AddEventHandler(controller.HandleAll(impl.Enqueue))

	return impl
}
