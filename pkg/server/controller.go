package server

import (
	"fmt"
	"time"

	"github.com/pkg/errors"
	solov1 "github.com/solo-io/glue/pkg/platform/kube/crd/solo.io/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilrt "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

const (
	// TODO (ashish) convert these into configurations
	maxRetries  = 5
	resyncDelay = 5 * time.Minute
)

type handler interface {
}

type controller struct {
	clientset    kubernetes.Interface
	queue        workqueue.RateLimitingInterface
	informer     cache.SharedIndexInformer
	eventHandler handler
}

func newController(upstreamRepo UpstreamRepository) *controller {
	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())
	lw := &cache.ListWatch{
		ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
			return upstreamRepo.List(options)
		},
		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
			return upstreamRepo.Watch(options)
		},
	}
	informer := cache.NewSharedIndexInformer(lw, &solov1.Upstream{}, resyncDelay, cache.Indexers{})
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(obj)
			if err == nil {
				queue.Add(key)
			}
		},
		UpdateFunc: func(old, new interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(new)
			if err == nil {
				queue.Add(key)
			}
		},
		DeleteFunc: func(obj interface{}) {
			key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
			if err == nil {
				queue.Add(key)
			}
		},
	})

	return &controller{queue: queue, informer: informer}
}

func (c *controller) Run(stop chan struct{}) {
	defer utilrt.HandleCrash()
	defer c.queue.ShutDown()

	go c.informer.Run(stop)
	// wait for cache sync before starting
	if !cache.WaitForCacheSync(stop, c.informer.HasSynced) {
		utilrt.HandleError(fmt.Errorf("Timed out waiting for caches to sync"))
	}
	wait.Until(c.runWorker, time.Second, stop)
}

func (c *controller) runWorker() {
	for c.processNextItem() {
	}
}

func (c *controller) processNextItem() bool {
	key, quit := c.queue.Get()
	if quit {
		return false
	}
	defer c.queue.Done(key)

	err := c.processItem(key.(string))
	if err == nil {
		c.queue.Forget(key)
	} else if c.queue.NumRequeues(key) < maxRetries {
		c.queue.AddRateLimited(key)
	} else {
		c.queue.Forget(key)
		utilrt.HandleError(err)
	}
	return true
}

func (c *controller) processItem(key string) error {
	obj, exists, err := c.informer.GetIndexer().GetByKey(key)
	if err != nil {
		return errors.Wrapf(err, "error fetching object with key %s from store", key)
	}
	if !exists {
		// remove this upstream from our list of sources
		fmt.Println("removed from fetch list - ", key)
		return nil
	}
	fmt.Println("updated fetch list with - ", key, obj)
	return nil
}
