package operator

import (
	"context"
	"fmt"
	"k8s.io/apimachinery/pkg/api/meta"
	"strings"
	"unicode"

	"github.com/google/go-cmp/cmp"
	log "github.com/sirupsen/logrus"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

const maxRetries = 5

// Handler defines final service handler
type Handler interface {
	Update(ctx context.Context, namespace, name string) error
	Delete(ctx context.Context, namespace, name string) error
}

// Controller defines Prometheus Server core base
type Controller struct {
	queue        workqueue.RateLimitingInterface
	informer     cache.SharedIndexInformer
	eventHandler Handler
}

// NewController instantiates PrometheusServer controller
func NewController(eventHandler Handler, informer cache.SharedIndexInformer) *Controller {
	ctl := &Controller{
		informer:     informer,
		queue:        workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter()),
		eventHandler: eventHandler,
	}

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			ctl.enqueuePrometheusServer(obj)
		},
		UpdateFunc: func(old, new interface{}) {
			if log.GetLevel() == log.DebugLevel {
				dumpDifference(old, new)
			}

			na, err := meta.Accessor(new)
			if err != nil {
				log.Errorf("unable to get meta accessor on update new obj, error %v", err)
				return
			}
			oa, err := meta.Accessor(old)
			if err != nil {
				log.Errorf("unable to get meta accessor on update old obj, error %v", err)
				return
			}

			// skip updates without resource version change
			if na.GetResourceVersion() == oa.GetResourceVersion() {
				return
			}

			ctl.enqueuePrometheusServer(new)
		},
		DeleteFunc: func(obj interface{}) {
			ctl.enqueuePrometheusServer(obj)
		},
	})
	return ctl
}

// Run starts controller loop
func (c *Controller) Run(ctx context.Context) {
	defer utilruntime.HandleCrash()

	if !cache.WaitForCacheSync(ctx.Done(), c.informer.HasSynced) {
		return
	}

	log.Debugf("First Cache Synced on version %s", c.informer.LastSyncResourceVersion())

	c.runWorker(ctx)
}

func (c *Controller) runWorker(ctx context.Context) {
	for c.processNextItem(ctx) {
	}
}

func (c *Controller) processNextItem(ctx context.Context) bool {
	key, quit := c.queue.Get()
	if quit {
		log.Info("Queue goes down!")
		return false
	}
	defer c.queue.Done(key)

	err := c.handle(ctx, key)
	if err == nil {
		c.queue.Forget(key)
		return true
	}

	if c.queue.NumRequeues(key) < maxRetries {
		log.Errorf("Error processing key %s, retry. Error: %v", key, err)
		c.queue.AddRateLimited(key)
		return true
	}

	log.Errorf("Error processing key %s Max retries achieved: %v", key, err)
	c.queue.Forget(key)
	utilruntime.HandleError(err)

	return true
}

func (c *Controller) handle(ctx context.Context, k interface{}) error {
	key := k.(string)
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return fmt.Errorf("invalid resource key: %s", key)
	}

	_, exists, err := c.informer.GetIndexer().GetByKey(key)
	if err != nil {
		return fmt.Errorf("unable to fetching object with key %s from store: %v", key, err)
	}

	if !exists {
		log.Debugf("handling deletion on key %s", key)
		return c.eventHandler.Delete(ctx, namespace, name)
	}

	return c.eventHandler.Update(ctx, namespace, name)
}

func (c *Controller) enqueuePrometheusServer(obj interface{}) {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
	if err != nil {
		utilruntime.HandleError(err)
		return
	}

	c.queue.Add(key)
}

func dumpDifference(old, new interface{}) {
	diff := cmp.Diff(old, new)
	cleanDiff := strings.TrimFunc(diff, func(r rune) bool {
		return !unicode.IsGraphic(r)
	})
	if len(cleanDiff) > 0 {
		fmt.Println("UPDATE diff: ", cleanDiff)
	}
}
