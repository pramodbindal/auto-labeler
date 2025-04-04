package labeler

import (
	"context"
	labelerinformer "github.com/pramodbindal/auto-labeler/pkg/client/injection/informers/pramodbindal/v1alpha1/labeler"
	labelerreconciler "github.com/pramodbindal/auto-labeler/pkg/client/injection/reconciler/pramodbindal/v1alpha1/labeler"
	kubeclient "knative.dev/pkg/client/injection/kube/client"
	deploymentinformer "knative.dev/pkg/client/injection/kube/informers/apps/v1/deployment"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/logging"
	"log"
)

func NewController(ctx context.Context, cmw configmap.Watcher) *controller.Impl {
	logger := logging.FromContext(ctx)

	labelerInformer := labelerinformer.Get(ctx)
	deploymentInformer := deploymentinformer.Get(ctx)
	//Create Reconciler
	reconciler := &Reconciler{
		// The client will be needed to create/delete Pods via the API.
		kubeclient:       kubeclient.Get(ctx),
		deploymentLister: deploymentInformer.Lister(),
	}
	ctrlOptions := controller.Options{
		Concurrency: 5,
	}

	logger.Info("Setting up event handlers")
	impl := labelerreconciler.NewImpl(ctx, reconciler, func(impl *controller.Impl) controller.Options {
		logger.Infof("Inside Labeler Reconciler optionsFns")
		return ctrlOptions
	})

	logger.Info("Add Event handlers")
	_, err := labelerInformer.Informer().AddEventHandler(controller.HandleAll(impl.Enqueue))
	if err != nil {
		log.Fatalf("error adding labeler informer %v", err)
	}

	return impl

}
