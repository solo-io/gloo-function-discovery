package server

import (
	"github.com/pkg/errors"
	clientset "github.com/solo-io/glue/pkg/platform/kube/crd/client/clientset/versioned"
	"github.com/solo-io/glue/pkg/platform/kube/crd/client/clientset/versioned/typed/solo.io/v1"
	"k8s.io/client-go/rest"
)

// UpstreamInterface provides an interfce to talk to Upstreams represented by CRDs in K8S
func UpstreamInterface(cfg *rest.Config, namespace string) (v1.UpstreamInterface, error) {
	glooClient, err := clientset.NewForConfig(cfg)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get Gloo client")
	}
	gloov1 := glooClient.GlueV1()
	return gloov1.Upstreams(namespace), nil
}
