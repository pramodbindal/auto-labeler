package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	SchemeGroupVersion = schema.GroupVersion{Group: "pramodbindal.dev", Version: "v1alpha1"}
	SchemeBuilder      = runtime.NewSchemeBuilder(addToScheme)
	AddToScheme        = SchemeBuilder.AddToScheme
)

func addToScheme(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(SchemeGroupVersion, &Labeler{}, &LabelerList{})
	metav1.AddToGroupVersion(scheme, SchemeGroupVersion)
	return nil
}
// Resource takes an unqualified resource and returns a Group qualified GroupResource
func Resource(resource string) schema.GroupResource {
	return SchemeGroupVersion.WithResource(resource).GroupResource()
}

