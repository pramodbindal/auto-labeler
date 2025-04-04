package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/pramodbindal/auto-labeler/pkg/apis/pramodbindal/v1alpha1"
	myclient "github.com/pramodbindal/auto-labeler/pkg/client/clientset/versioned"
	labelerinformers "github.com/pramodbindal/auto-labeler/pkg/client/informers/externalversions"
	v1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/informers"
	k8s "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"log"
	"reflect"
	"time"
)

func main() {

	// Create k8sConfig Object

	config, err := rest.InClusterConfig()
	handle(err)
	ctx, cancel := context.WithCancel(context.Background())

	client, err := myclient.NewForConfig(config)
	k8sClient, err := k8s.NewForConfig(config)

	defer cancel()

	registerLabelEventHandlers(ctx, client, k8sClient)

	// Wait forever (or until Ctrl+C)
	<-ctx.Done()

}

func labelerAdded(ctx context.Context, labeler *v1alpha1.Labeler, client *k8s.Clientset, factory labelerinformers.SharedInformerFactory) {
	fmt.Printf("LABELER Added: %s\n", labeler.Name)
	targetResource := labeler.Spec.TargetResource
	switch targetResource {
	case "deployment":
		registerDeploymentHandler(ctx, client, factory)
		deploymentFactory := informers.NewSharedInformerFactory(client, 30*time.Second)
		deployments, err := deploymentFactory.Apps().V1().Deployments().Lister().List(labels.Everything())
		handle(err)
		for _, deployment := range deployments {
			updateLabels(deployment, factory, client)
		}

	default:
		log.Printf("labeler %s is not supported", targetResource)
	}

}

func registerLabelEventHandlers(ctx context.Context, client *myclient.Clientset, k8sClient *k8s.Clientset) (labelerinformers.SharedInformerFactory, cache.ResourceEventHandlerFuncs) {
	labelerFactory := labelerinformers.NewSharedInformerFactory(client, 30*time.Second)
	labelEventHandlers := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			labeler := obj.(*v1alpha1.Labeler)
			labelerAdded(ctx, labeler, k8sClient, labelerFactory)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			labeler := newObj.(*v1alpha1.Labeler)
			labelerAdded(ctx, labeler, k8sClient, labelerFactory)
		},
		DeleteFunc: func(obj interface{}) {
			labeler := obj.(*v1alpha1.Labeler)
			fmt.Printf("LABELER Deleted: RE %s\n", labeler.Name)
		},
	}
	addEventHandler(labelerFactory.Pramodbindal().V1alpha1().Labelers().Informer(), labelEventHandlers)
	labelerFactory.Start(ctx.Done()) // Start the informer

	return labelerFactory, labelEventHandlers
}

func addEventHandler(informer cache.SharedIndexInformer, handlers cache.ResourceEventHandlerFuncs) {
	_, err := informer.AddEventHandler(handlers)
	handle(err)
}

func handle(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func updateLabels(deploy *v1.Deployment, labelerFactory labelerinformers.SharedInformerFactory, clientset *k8s.Clientset) {
	log.Printf("Updating Labels: %s\n", deploy.Name)
	labelers, err := labelerFactory.Pramodbindal().V1alpha1().Labelers().Lister().List(labels.Everything())
	handle(err)
	labelsMap := make(map[string]string)
	for k, v := range deploy.Labels {
		labelsMap[k] = v
	}
	for _, labeler := range labelers {
		log.Printf("Updating Labels for labeler : %s\n", labeler.Name)
		for k, v := range labeler.Spec.Labels {
			labelsMap[k] = v
		}
	}
	log.Printf("Labels: %v", labelsMap)
	metadata := deploy.GetObjectMeta()

	updatedMetaData, err := json.Marshal(metadata)

	patch := updatedMetaData
	_, err = clientset.AppsV1().Deployments(deploy.Namespace).Patch(
		context.TODO(), deploy.Name, types.StrategicMergePatchType, patch, metav1.PatchOptions{},
	)

}

func registerDeploymentHandler(ctx context.Context, clientset *k8s.Clientset, labelerfactory labelerinformers.SharedInformerFactory) {
	factory := informers.NewSharedInformerFactory(clientset, 30*time.Second)

	deploymentEventHandlers := cache.ResourceEventHandlerFuncs{

		AddFunc: func(obj interface{}) {
			k8sResource := obj.(*v1.Deployment)
			//fmt.Printf("Deployment Added: %s.%s\n", k8sResource.Namespace, k8sResource.Name)
			updateLabels(k8sResource, labelerfactory, clientset)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			oldDeploy := oldObj.(*v1.Deployment)
			newDeploy := newObj.(*v1.Deployment)
			if !reflect.DeepEqual(oldDeploy.Labels, newDeploy.Labels) {
				updateLabels(newDeploy, labelerfactory, clientset)
			}
			//fmt.Printf("Deployment Updated: %s.%s\n", k8sResource.Namespace, k8sResource.Name)

		},
	}
	addEventHandler(factory.Apps().V1().Deployments().Informer(), deploymentEventHandlers)

	factory.Start(ctx.Done()) // Start the informer
}
