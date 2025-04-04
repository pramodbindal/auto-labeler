package labeler

import (
	"context"
	"encoding/json"
	"github.com/pramodbindal/auto-labeler/pkg/apis/pramodbindal/v1alpha1"
	pramodbindalv1alpha1 "github.com/pramodbindal/auto-labeler/pkg/client/listers/pramodbindal/v1alpha1"
	v1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	k8s "k8s.io/client-go/kubernetes"
	v2 "k8s.io/client-go/listers/apps/v1"
	"knative.dev/pkg/logging"
	"knative.dev/pkg/reconciler"
)

type Reconciler struct {
	kubeclient       k8s.Interface
	deploymentLister v2.DeploymentLister
	labelerInformer  pramodbindalv1alpha1.LabelerLister
}

func (r Reconciler) ReconcileKind(ctx context.Context, labeler *v1alpha1.Labeler) reconciler.Event {
	logger := logging.FromContext(ctx)
	logger.Infof("Reconcile Labeler : %s", labeler.Name)
	deployments := r.getDeployments(ctx)
	logger.Infof("Deployments Listed : %v", deployments)
	for _, deployment := range deployments {
		logger.Infof("Updating Labels for Deployment : %s", deployment.Name)
		r.updateDeploymentLabels(ctx, deployment, labeler)
	}
	return nil

}

func (r Reconciler) getDeployments(ctx context.Context) []*v1.Deployment {
	logger := logging.FromContext(ctx)
	deployments, err := r.deploymentLister.List(labels.Everything())
	if err != nil {
		logger.Fatal("Error listing Deployments", err)
	}
	logger.Info("Deployments Returned : ", deployments)
	return deployments

}

func (r Reconciler) updateDeploymentLabels(ctx context.Context, deploy *v1.Deployment, labeler *v1alpha1.Labeler) {
	logger := logging.FromContext(ctx)

	logger.Infof("Updating Labels: %s", deploy.Name)
	////labelsMap := make(map[string]string)
	////for k, v := range deploy.Labels {
	////	labelsMap[k] = v
	////}
	////logger.Infof("Updating Labels for labeler : %s\n", labeler.Name)
	////for k, v := range labeler.Spec.Labels {
	////	labelsMap[k] = v
	////}
	//logger.Infof("Labels: %v", labelsMap)

	patchData, err := json.Marshal(map[string]interface{}{
		"metadata": map[string]interface{}{
			"labels": labeler.Spec.Labels,
		},
	})

	if err != nil {
		logger.Errorf("failed to create patch data: %v", err)
	}

	// Patch the Deployment
	_, err = r.kubeclient.AppsV1().Deployments(deploy.Namespace).Patch(
		context.TODO(),
		deploy.Name,
		types.MergePatchType, // Merge strategy
		patchData,
		metav1.PatchOptions{},
	)
	if err != nil {
		logger.Errorf("failed to patch deployment: %v", err)
	}

	logger.Infof("Successfully patched deployment %s with new labels\n", deploy.Name)
}
