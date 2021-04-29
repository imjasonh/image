package main

import (
	// The set of controllers this controller process runs.
	"github.com/imjasonh/image/pkg/reconciler/image"

	// This defines the shared main for injected controllers.
	"knative.dev/pkg/injection/sharedmain"
)

func main() {
	sharedmain.Main("controller",
		image.NewController,
	)
}
