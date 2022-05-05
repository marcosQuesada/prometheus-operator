package operator

import (
	"context"
	"fmt"
	"strings"
	"time"
	"unicode"

	"github.com/google/go-cmp/cmp"
	"github.com/marcosQuesada/prometheus-operator/pkg/crd/apis/prometheusserver/v1alpha1"
	log "github.com/sirupsen/logrus"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

const maxRetries = 5
const defaultWorkerFrequency = time.Second * 5

// Handler defines final service handler
type Handler interface {
	Update(ctx context.Context, namespace, name string) error
	Delete(ctx context.Context, namespace, name string) error
}

// Controller defines Prometheus Server core base
type Controller struct {
	queue           workqueue.RateLimitingInterface
	informer        cache.SharedIndexInformer
	eventHandler    Handler
	workerFrequency time.Duration
}

// NewController instantiates PrometheusServer controller
func NewController(eventHandler Handler, informer cache.SharedIndexInformer) *Controller {
	ctl := &Controller{
		informer:        informer,
		queue:           workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter()),
		eventHandler:    eventHandler,
		workerFrequency: defaultWorkerFrequency,
	}

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			ctl.enqueuePrometheusServer(obj)
		},
		UpdateFunc: func(old, new interface{}) {
			if log.GetLevel() == log.DebugLevel {
				dumpDifference(old, new)
			}

			newPs, ok := new.(*v1alpha1.PrometheusServer)
			if !ok {
				return
			}
			oldPs, ok := old.(*v1alpha1.PrometheusServer)
			if !ok {
				return
			}
			if newPs.ResourceVersion == oldPs.ResourceVersion {
				return
			}

			ctl.enqueuePrometheusServer(newPs)
		},
		DeleteFunc: func(obj interface{}) {
			ctl.enqueuePrometheusServer(obj)
		},
	})
	return ctl
}

// Run starts controller loop
func (c *Controller) Run(ctx context.Context, workers int) {
	defer utilruntime.HandleCrash()

	if !cache.WaitForCacheSync(ctx.Done(), c.informer.HasSynced) {
		return
	}

	log.Debugf("First Cache Synced on version %s", c.informer.LastSyncResourceVersion())

	for i := 0; i < workers; i++ {
		go wait.UntilWithContext(ctx, c.runWorker, c.workerFrequency)
	}
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
