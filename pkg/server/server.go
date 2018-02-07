package server

import (
	"github.com/solo-io/glue/pkg/platform/kube/crd/client/clientset/versioned/typed/solo.io/v1"
)

type UpstreamRepository v1.UpstreamInterface

type Server struct {
	UpstreamRepo UpstreamRepository
	Port         int
}

func (s *Server) Start(stop chan struct{}) {
	ctrlr := newController(s.UpstreamRepo)
	ctrlr.Run(stop)
}
