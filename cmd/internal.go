package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gorilla/mux"
	"github.com/marcosQuesada/prometheus-operator/internal/service"
	"github.com/marcosQuesada/prometheus-operator/internal/service/resource"
	"github.com/marcosQuesada/prometheus-operator/internal/service/usecase"
	cfg "github.com/marcosQuesada/prometheus-operator/pkg/config"
	"github.com/marcosQuesada/prometheus-operator/pkg/crd"
	crdinformers "github.com/marcosQuesada/prometheus-operator/pkg/crd/generated/informers/externalversions"
	ht "github.com/marcosQuesada/prometheus-operator/pkg/http/handler"
	"github.com/marcosQuesada/prometheus-operator/pkg/operator"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
)

// internalCmd represents the internal command
var internalCmd = &cobra.Command{
	Use:   "internal",
	Short: "prometheus server internal controller",
	Long:  `prometheus server internal controller handles prometheus-server resources`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Infof("controller internal listening on namespace %s label %s Version %s release date %s http server on port %s", namespace, watchLabel, cfg.Commit, cfg.Date, cfg.HttpPort)

		clientSet := operator.BuildInternalClient()
		pmClientSet := crd.BuildPrometheusServerInternalClient()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		api := operator.BuildAPIInternalClient()
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

		if !cache.WaitForCacheSync(ctx.Done(),
			cr.HasSynced,
			crb.HasSynced,
			cm.HasSynced,
			dpl.HasSynced,
			svc.HasSynced) {
			log.Fatal("unable to sync informers")
		}

		r := []service.ResourceEnforcer{
			resource.NewClusterRole(clientSet, shInf.Rbac().V1().ClusterRoles().Lister()),
			resource.NewClusterRoleBinding(clientSet, shInf.Rbac().V1().ClusterRoleBindings().Lister()),
			resource.NewConfigMap(clientSet, shInf.Core().V1().ConfigMaps().Lister()),
			resource.NewDeployment(clientSet, shInf.Apps().V1().Deployments().Lister()),
			resource.NewService(clientSet, shInf.Core().V1().Services().Lister()),
		}
		re := service.NewResource(r...)
		generationCache := service.NewGenerationCache()
		fnlz := service.NewFinalizer(pmClientSet)
		rec := createRecorder(clientSet, prometheusServerOperatorUserAgent)
		cnlt := service.NewConciliator()
		cnlt.Register(usecase.NewCreator(fnlz, re, rec))
		cnlt.Register(usecase.NewDeleter(fnlz, re, rec))
		cnlt.Register(usecase.NewReloader(generationCache, re, rec))

		op := service.NewOperator(crdInf.K8slab().V1alpha1().PrometheusServers().Lister(), pmClientSet, generationCache, cnlt)
		ctl := operator.NewController(op, ps)
		go ctl.Run(ctx)

		router := mux.NewRouter()
		ch := ht.NewChecker(cfg.Commit, cfg.Date)
		ch.Routes(router)
		router.Handle("/metrics", promhttp.Handler())

		srv := &http.Server{
			Addr:         fmt.Sprintf(":%s", cfg.HttpPort),
			Handler:      router,
			ReadTimeout:  httpReadTimeout,
			WriteTimeout: httpWriteTimeout,
		}

		go func(h *http.Server) {
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
	rootCmd.AddCommand(internalCmd)
}
