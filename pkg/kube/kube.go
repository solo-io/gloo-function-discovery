package kube

import (
	"github.com/pkg/errors"
	clientset "github.com/solo-io/glue/pkg/platform/kube/crd/client/clientset/versioned"
	"github.com/solo-io/glue/pkg/platform/kube/crd/client/clientset/versioned/typed/solo.io/v1"
	"k8s.io/client-go/rest"
)

func UpstreamInterface(cfg *rest.Config, namespace string) (v1.UpstreamInterface, error) {
	glueClient, err := clientset.NewForConfig(cfg)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get Glue client")
	}
	gluev1 := glueClient.GlueV1()
	return gluev1.Upstreams(namespace), nil
}
