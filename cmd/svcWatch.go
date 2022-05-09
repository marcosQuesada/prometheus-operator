package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/google/go-cmp/cmp"
	"github.com/marcosQuesada/prometheus-operator/internal/service"
	"github.com/marcosQuesada/prometheus-operator/pkg/operator"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"unicode"

	"github.com/spf13/cobra"
)

// svcWatchCmd represents the svcWatch command
var svcWatchCmd = &cobra.Command{
	Use:   "svcWatch",
	Short: "watch Service resource state events",
	Long:  `watch Service resource state events`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("pmtWatch called")
		clientSet := operator.BuildExternalClient()
		shInf := informers.NewSharedInformerFactory(clientSet, 0)
		svc := shInf.Core().V1().Services().Informer()
		end := shInf.Core().V1().Endpoints().Informer()

		stopper := make(chan struct{})
		svc.AddEventHandler(cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				mObj := obj.(v1.Object)
				if mObj.GetNamespace() != service.MonitoringNamespace {
					return
				}
				log.Printf("New Object Added to Store: %s", mObj.GetName())
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				mObj := newObj.(v1.Object)
				if mObj.GetNamespace() != service.MonitoringNamespace {
					return
				}
				diff := cmp.Diff(oldObj, newObj)
				cleanDiff := strings.TrimFunc(diff, func(r rune) bool {
					return !unicode.IsGraphic(r)
				})
				if len(cleanDiff) > 0 {
					fmt.Println("UPDATE diff: ", cleanDiff)
					return
				}
				log.Printf("Object Updated without changes: %s", mObj.GetName())
			},
			DeleteFunc: func(obj interface{}) {
				mObj := obj.(v1.Object)
				if mObj.GetNamespace() != service.MonitoringNamespace {
					return
				}
				log.Printf("Object Deleted from Store: %s", mObj.GetName())
			},
		})
		end.AddEventHandler(cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				mObj := obj.(v1.Object)
				if mObj.GetNamespace() != service.MonitoringNamespace {
					return
				}
				raw, _ := json.Marshal(mObj)
				log.Printf("New Endpoint Added to Store: %s", mObj.GetName())
				fmt.Println(string(raw))
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				mObj := newObj.(v1.Object)
				if mObj.GetNamespace() != service.MonitoringNamespace {
					return
				}
				diff := cmp.Diff(oldObj, newObj)
				cleanDiff := strings.TrimFunc(diff, func(r rune) bool {
					return !unicode.IsGraphic(r)
				})
				if len(cleanDiff) > 0 {
					fmt.Println("UPDATE diff: ", cleanDiff)
					return
				}

				log.Printf("Endpoint Updated without changes: %s", mObj.GetName())
			},
			DeleteFunc: func(obj interface{}) {
				mObj := obj.(v1.Object)
				if mObj.GetNamespace() != service.MonitoringNamespace {
					return
				}
				log.Printf("Endpoint Deleted from Store: %s", mObj.GetName())
			},
		})

		go shInf.Start(stopper)

		if !cache.WaitForCacheSync(stopper, svc.HasSynced, end.HasSynced) {
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
	rootCmd.AddCommand(svcWatchCmd)
}
