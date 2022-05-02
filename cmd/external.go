package cmd

import (
	"context"
	"fmt"
	"github.com/gorilla/mux"
	cfg "github.com/marcosQuesada/prometheus-operator/pkg/config"
	"github.com/marcosQuesada/prometheus-operator/pkg/crd"
	"github.com/marcosQuesada/prometheus-operator/pkg/crd/apis/prometheusserver/v1alpha1"
	"github.com/marcosQuesada/prometheus-operator/pkg/operator"
	"github.com/marcosQuesada/prometheus-operator/pkg/operator/resource"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"

	crdinformers "github.com/marcosQuesada/prometheus-operator/pkg/crd/generated/informers/externalversions"

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
	Long:  `prometheus server external controller balance configured keys between swarm peers, useful on development path`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Infof("controller external listening on namespace %s label %s Version %s release date %s http server on port %s", namespace, watchLabel, cfg.Commit, cfg.Date, cfg.HttpPort)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		clientSet := operator.BuildExternalClient()
		pmClientSet := crd.BuildPrometheusServerExternalClient()

		api := operator.BuildAPIExternalClient()
		m := crd.NewManager(api)
		if err := crd.NewBuilder(m).EnsureCRDRegistered(); err != nil {
			log.Fatalf("unable to check swarm crd status, error %v", err)
		}

		crdif := crdinformers.NewSharedInformerFactory(pmClientSet, time.Second*8)
		sif := informers.NewSharedInformerFactory(clientSet, 0)

		ps := crdif.K8slab().V1alpha1().PrometheusServers()
		ns := sif.Core().V1().Namespaces()
		cr := sif.Rbac().V1().ClusterRoles()
		crb := sif.Rbac().V1().ClusterRoleBindings()
		cm := sif.Core().V1().ConfigMaps()
		dpl := sif.Apps().V1().Deployments()
		svc := sif.Core().V1().Services()

		psi := ps.Informer()
		nsi := ns.Informer()
		cli := cr.Informer()
		clbi := crb.Informer()
		cmi := cm.Informer()
		dpli := dpl.Informer()
		svci := svc.Informer()

		crdif.Start(ctx.Done())
		sif.Start(ctx.Done())

		if !cache.WaitForNamedCacheSync(v1alpha1.CrdKind, ctx.Done(), nsi.HasSynced, cli.HasSynced, clbi.HasSynced, cmi.HasSynced, psi.HasSynced, dpli.HasSynced, svci.HasSynced) { // @TODO: CHECK!
			log.Fatal("unable to sync pod informer")
		}

		r := []operator.ResourceEnforcer{
			resource.NewNamespace(clientSet, ns.Lister()),
			resource.NewClusterRole(clientSet, cr.Lister()),
			resource.NewClusterRoleBinding(clientSet, crb.Lister()),
			resource.NewConfigMap(clientSet, cm.Lister()),
			resource.NewDeployment(clientSet, dpl.Lister()),
			resource.NewService(clientSet, svc.Lister()),
		}
		op := operator.NewOperator(pmClientSet, r)

		crdh := crd.NewHandler(op)
		swCtl := operator.New(crdh, ps.Informer(), operator.NewRunner(), v1alpha1.CrdKind)
		go swCtl.Run(ctx)

		router := mux.NewRouter()
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
