package cmd

import (
	"context"
	"fmt"
	"github.com/gorilla/mux"
	cfg "github.com/marcosQuesada/prometheus-operator/pkg/config"
	"github.com/marcosQuesada/prometheus-operator/pkg/crd"
	crdinformers "github.com/marcosQuesada/prometheus-operator/pkg/crd/generated/informers/externalversions"
	ht "github.com/marcosQuesada/prometheus-operator/pkg/http/handler"
	"github.com/marcosQuesada/prometheus-operator/pkg/operator"
	"github.com/marcosQuesada/prometheus-operator/pkg/service"
	resource2 "github.com/marcosQuesada/prometheus-operator/pkg/service/resource"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// externalCmd represents the external command
var externalCmd = &cobra.Command{
	Use:   "external",
	Short: "prometheus server external controller, useful on development path",
	Long:  `prometheus server external controller development version, useful on development path`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Infof("controller external listening on namespace %s label %s Version %s release date %s http server on port %s", namespace, watchLabel, cfg.Commit, cfg.Date, cfg.HttpPort)

		clientSet := operator.BuildExternalClient()
		pmClientSet := crd.BuildPrometheusServerExternalClient()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		api := operator.BuildAPIExternalClient()
		m := crd.NewManager(api)
		if err := crd.NewBuilder(m).EnsureCRDRegistration(ctx); err != nil {
			log.Fatalf("unable to ensure prometheus server crd registration, error %v", err)
		}

		crdInf := crdinformers.NewSharedInformerFactory(pmClientSet, reSyncInterval)
		shInf := informers.NewSharedInformerFactory(clientSet, 0)

		ps := crdInf.K8slab().V1alpha1().PrometheusServers().Informer()
		cr := shInf.Rbac().V1().ClusterRoles().Informer()
		crb := shInf.Rbac().V1().ClusterRoleBindings().Informer()
		cm := shInf.Core().V1().ConfigMaps().Informer()
		dpl := shInf.Apps().V1().Deployments().Informer()
		svc := shInf.Core().V1().Services().Informer()

		crdInf.Start(ctx.Done())
		shInf.Start(ctx.Done())

		log.Info("Waiting Informer sync")
		if !cache.WaitForCacheSync(ctx.Done(),
			cr.HasSynced,
			crb.HasSynced,
			cm.HasSynced,
			dpl.HasSynced,
			svc.HasSynced) {
			log.Fatal("unable to sync informers")
		}

		r := []service.ResourceEnforcer{
			resource2.NewClusterRole(clientSet, shInf.Rbac().V1().ClusterRoles().Lister()),
			resource2.NewClusterRoleBinding(clientSet, shInf.Rbac().V1().ClusterRoleBindings().Lister()),
			resource2.NewConfigMap(clientSet, shInf.Core().V1().ConfigMaps().Lister()),
			resource2.NewDeployment(clientSet, shInf.Apps().V1().Deployments().Lister()),
			resource2.NewService(clientSet, shInf.Core().V1().Services().Lister()),
		}
		op := service.NewOperator(crdInf.K8slab().V1alpha1().PrometheusServers().Lister(), pmClientSet, r)
		ctl := operator.NewController(op, ps)
		go ctl.Run(ctx, workers)

		router := mux.NewRouter()
		ch := ht.NewChecker(cfg.Commit, cfg.Date)
		ch.Routes(router)
		router.Handle("/metrics", promhttp.Handler())

		srv := &http.Server{
			Addr:         fmt.Sprintf(":%s", cfg.HttpPort),
			Handler:      router,
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
		}

		go func(h *http.Server) {
			log.Infof("starting server on port %s", cfg.HttpPort)
			e := h.ListenAndServe()
			if e != nil && e != http.ErrServerClosed {
				log.Fatalf("Could not Listen and server, error %v", e)
			}
		}(srv)

		sigTerm := make(chan os.Signal, 1)
		signal.Notify(sigTerm, syscall.SIGTERM, syscall.SIGINT)
		<-sigTerm
		if err := srv.Close(); err != nil {
			log.Errorf("unexpected error on http server close %v", err)
		}
		cancel()
		_ = srv.Close()

		log.Info("Stopping controller")
	},
}

func init() {
	rootCmd.AddCommand(externalCmd)
}
