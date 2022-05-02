package cmd

import (
	"context"
	"fmt"
	"github.com/gorilla/mux"
	cfg "github.com/marcosQuesada/prometheus-operator/pkg/config"
	"github.com/marcosQuesada/prometheus-operator/pkg/crd"
	"github.com/marcosQuesada/prometheus-operator/pkg/crd/apis/prometheusserver/v1alpha1"
	crdinformers "github.com/marcosQuesada/prometheus-operator/pkg/crd/generated/informers/externalversions"
	ht "github.com/marcosQuesada/prometheus-operator/pkg/http/handler"
	"github.com/marcosQuesada/prometheus-operator/pkg/operator"
	"github.com/marcosQuesada/prometheus-operator/pkg/service"
	resource2 "github.com/marcosQuesada/prometheus-operator/pkg/service/resource"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// internalCmd represents the internal command
var internalCmd = &cobra.Command{
	Use:   "internal",
	Short: "prometheus server internal controller",
	Long:  `prometheus server internal controller handles prometheus-server resources`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Infof("controller internal listening on namespace %s label %s Version %s release date %s http server on port %s", namespace, watchLabel, cfg.Commit, cfg.Date, cfg.HttpPort)

		clientSet := operator.BuildInternalClient()
		pmClientSet := crd.BuilBuildPrometheusServerInternalClient()

		api := operator.BuildAPIInternalClient()
		m := crd.NewManager(api)
		if err := crd.NewBuilder(m).EnsureCRDRegistered(); err != nil {
			log.Fatalf("unable to ensure prometheus server crd registration, error %v", err)
		}

		crdif := crdinformers.NewSharedInformerFactory(pmClientSet, time.Second*8)
		sif := informers.NewSharedInformerFactory(clientSet, 0)

		ps := crdif.K8slab().V1alpha1().PrometheusServers()
		cr := sif.Rbac().V1().ClusterRoles()
		crb := sif.Rbac().V1().ClusterRoleBindings()
		cm := sif.Core().V1().ConfigMaps()
		dpl := sif.Apps().V1().Deployments()
		svc := sif.Core().V1().Services()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		crdif.Start(ctx.Done())
		sif.Start(ctx.Done())

		go cr.Informer().Run(ctx.Done())
		go crb.Informer().Run(ctx.Done())
		go cm.Informer().Run(ctx.Done())
		go dpl.Informer().Run(ctx.Done())
		go svc.Informer().Run(ctx.Done())
		go ps.Informer().Run(ctx.Done())

		log.Info("Waiting Cluster Role Informer sync")
		if !cache.WaitForNamedCacheSync("Cluster Role", ctx.Done(), cr.Informer().HasSynced) { // @TODO: CHECK!
			log.Fatal("unable to sync pod informer")
		}

		log.Info("Waiting Cluster Role Binding Informer sync")
		if !cache.WaitForNamedCacheSync("cluster role binding", ctx.Done(), crb.Informer().HasSynced) { // @TODO: CHECK!
			log.Fatal("unable to sync pod informer")
		}

		log.Info("Waiting Configmap informer sync")
		if !cache.WaitForNamedCacheSync("Configmap", ctx.Done(), cm.Informer().HasSynced) { // @TODO: CHECK!
			log.Fatal("unable to sync pod informer")
		}

		log.Info("Waiting Deployments informer sync")
		if !cache.WaitForNamedCacheSync("Deployments", ctx.Done(), dpl.Informer().HasSynced) { // @TODO: CHECK!
			log.Fatal("unable to sync pod informer")
		}

		log.Info("Waiting Services informer sync")
		if !cache.WaitForNamedCacheSync("services", ctx.Done(), svc.Informer().HasSynced) { // @TODO: CHECK!
			log.Fatal("unable to sync pod informer")
		}

		log.Info("Waiting Prometheus server informer sync")
		if !cache.WaitForNamedCacheSync(v1alpha1.CrdKind, ctx.Done(), ps.Informer().HasSynced) { // @TODO: CHECK!
			log.Fatal("unable to sync pod informer")
		}

		r := []service.ResourceEnforcer{
			resource2.NewClusterRole(clientSet, cr.Lister()),
			resource2.NewClusterRoleBinding(clientSet, crb.Lister()),
			resource2.NewConfigMap(clientSet, cm.Lister()),
			resource2.NewDeployment(clientSet, dpl.Lister()),
			resource2.NewService(clientSet, svc.Lister()),
		}
		op := service.NewOperator(pmClientSet, r)

		crdh := crd.NewHandler(op)
		swCtl := operator.New(crdh, ps.Informer(), operator.NewRunner(), v1alpha1.CrdKind)
		go swCtl.Run(ctx)

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
	rootCmd.AddCommand(internalCmd)
}
