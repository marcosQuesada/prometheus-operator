package cmd

import (
	"fmt"
	"github.com/google/go-cmp/cmp"
	"github.com/marcosQuesada/prometheus-operator/pkg/crd"
	crdinformers "github.com/marcosQuesada/prometheus-operator/pkg/crd/generated/informers/externalversions"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/cache"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"unicode"

	"github.com/spf13/cobra"
)

// pmtWatchCmd represents the pmtWatch command
var pmtWatchCmd = &cobra.Command{
	Use:   "pmtWatch",
	Short: "watch Prometheus Server state events",
	Long:  `watch Prometheus Server state events`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("pmtWatch called")
		pmClientSet := crd.BuildPrometheusServerExternalClient()
		crdInf := crdinformers.NewSharedInformerFactory(pmClientSet, reSyncInterval)
		ps := crdInf.K8slab().V1alpha1().PrometheusServers().Informer()

		stopper := make(chan struct{})
		ps.AddEventHandler(cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				mObj := obj.(v1.Object)
				log.Printf("New Object Added to Store: %s", mObj.GetName())
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				diff := cmp.Diff(oldObj, newObj)
				cleanDiff := strings.TrimFunc(diff, func(r rune) bool {
					return !unicode.IsGraphic(r)
				})
				if len(cleanDiff) > 0 {
					fmt.Println("UPDATE diff: ", cleanDiff)
					return
				}
				mObj := newObj.(v1.Object)
				log.Printf("Object Updated without changes: %s", mObj.GetName())
			},
			DeleteFunc: func(obj interface{}) {
				mObj := obj.(v1.Object)
				log.Printf("Object Deleted from Store: %s", mObj.GetName())
			},
		})

		go ps.Run(stopper)
		if !cache.WaitForCacheSync(stopper, ps.HasSynced) {
			runtime.HandleError(fmt.Errorf("timed out waiting for caches to sync"))
			return
		}

		sigTerm := make(chan os.Signal, 1)
		signal.Notify(sigTerm, syscall.SIGTERM, syscall.SIGINT)
		<-sigTerm

		close(stopper)
	},
}

func init() {
	rootCmd.AddCommand(pmtWatchCmd)
}
