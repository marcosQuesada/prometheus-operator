package operator

import (
	"context"
	"fmt"
	"github.com/google/go-cmp/cmp"
	"github.com/marcosQuesada/prometheus-operator/pkg/crd/apis/prometheusserver/v1alpha1"
	log "github.com/sirupsen/logrus"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"strings"
	"time"
	"unicode"
)

const maxRetries = 5

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

// NewController instantiates PrometheuServer controller
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
			// @TODO: REMOVE IT, OR ADD IT AS DEBUG
			dumpDifference(old, new)

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

// Run starts controller
func (c *Controller) Run(ctx context.Context, workers int) {
	defer utilruntime.HandleCrash()

	if !cache.WaitForCacheSync(ctx.Done(), c.informer.HasSynced) {
		return
	}

	log.Infof("First Cache Synced on version %s", c.informer.LastSyncResourceVersion())

	for i := 0; i < workers; i++ {
		go wait.UntilWithContext(ctx, c.runWorker, c.workerFrequency)
	}
}

func (c *Controller) runWorker(ctx context.Context) {
	log.Info("Run Worker")
	for c.processNextItem(ctx) {
	}
}

func (c *Controller) processNextItem(ctx context.Context) bool {
	e, quit := c.queue.Get()
	if quit {
		log.Error("Queue goes down!")
		return false
	}
	defer c.queue.Done(e)

	err := c.handle(ctx, e)
	if err == nil {
		c.queue.Forget(e)
		return true
	}

	if c.queue.NumRequeues(e) < maxRetries {
		log.Errorf("Error processing ev %v, retry. Error: %v", e, err)
		c.queue.AddRateLimited(e)
		return true
	}

	log.Errorf("Error processing %v Max retries achieved: %v", e, err)
	c.queue.Forget(e)
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
		log.Infof("handling deletion on key %s", key)
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
		fmt.Println("UPDATE Prometheus Server diff: ", cleanDiff)
	}
}
