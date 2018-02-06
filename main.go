package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apiexts "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	apiv1 "github.com/solo-io/glue/pkg/api/types/v1"
	clientset "github.com/solo-io/glue/pkg/platform/kube/crd/client/clientset/versioned"
	kubev1 "github.com/solo-io/glue/pkg/platform/kube/crd/solo.io/v1"
)

const (
	namespace = "default"
)

func main() {
	fmt.Println("testing glue + upstream client")

	kubeconf := flag.String("kubeconf", "admin.conf", "Path to k8s config. Required for out-of-cluster")
	flag.Parse()

	var restConfig *rest.Config
	if *kubeconf != "" {
		var err error
		restConfig, err = clientcmd.BuildConfigFromFlags("", *kubeconf)
		if err != nil {
			log.Fatalf("Unable to get k8s configuration %q\n", err)
		}
	} else {
		var err error
		restConfig, err = rest.InClusterConfig()
		if err != nil {
			log.Fatalf("Unable to get k8s configuration %q\n", err)
		}
	}
	registerUpstream(restConfig)
	time.Sleep(5 * time.Second) // just wait for registration of crd

	glueClient, err := clientset.NewForConfig(restConfig)
	if err != nil {
		log.Fatalf("Unable to get Glue client %q\n", err)
	}
	gluev1 := glueClient.GlueV1()

	upstreamInterface := gluev1.Upstreams(namespace)

	upstream := apiv1.Upstream{
		Name: "testupstream",
		Type: "aws",
		Spec: map[string]interface{}{
			"CreatedBy": "fds",
			"Region":    "us-east-1",
			"Secret":    "aaah!ican'ttellyou",
		},
		Functions: []apiv1.Function{
			apiv1.Function{
				Name: "testfunc",
				Spec: map[string]interface{}{
					"FunctionName": "testfunc",
					"Qualifier":    "v1-alpha",
				},
			},
		},
	}
	crd := kubev1.UpstreamToCRD(metav1.ObjectMeta{
		Name:   upstream.Name,
		Labels: map[string]string{"CreatedBy": "fds"},
	}, upstream)
	created, err := upstreamInterface.Create(crd)
	if err != nil {
		log.Fatalf("unable to create test upstream %q\n", err)
	} else {
		log.Println("created: ", created)
	}
	upstreamList, err := upstreamInterface.List(metav1.ListOptions{})
	if err != nil {
		log.Fatalf("unable to get list of upstreams %q\n", err)
	} else {
		log.Println("items:\n", upstreamList)
	}
	err = upstreamInterface.Delete(created.Name, &metav1.DeleteOptions{})
	if err != nil {
		log.Fatalf("unable to delete upstream %q\n", err)
	} else {
		log.Println("deleted upstream")
	}
}

func registerUpstream(restConfig *rest.Config) {
	client, err := apiexts.NewForConfig(restConfig)
	if err != nil {
		log.Fatalf("Unable to get a client to k8s to register CRD%q\n", err)
	}
	upstream := &v1beta1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{Name: "upstreams.glue.solo.io"},
		Spec: v1beta1.CustomResourceDefinitionSpec{
			Group:   "glue.solo.io",
			Version: "v1",
			Scope:   v1beta1.NamespaceScoped,
			Names: v1beta1.CustomResourceDefinitionNames{
				Plural: "upstreams",
				Kind:   "Upstream",
			},
		},
	}
	if _, err = client.ApiextensionsV1beta1().CustomResourceDefinitions().Create(upstream); err != nil && !apierrors.IsAlreadyExists(err) {
		log.Fatalf("unable to register upstream %q\n", err)
	}
}
