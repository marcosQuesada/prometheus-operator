package cmd

import (
	"fmt"
	"os"
	"time"

	cfg "github.com/marcosQuesada/prometheus-operator/pkg/config"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const appID = "prometheus-operator"

var (
	namespace      string
	watchLabel     string
	reSyncInterval time.Duration
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "root",
	Short: "root controller command",
	Long:  `root controller command`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if err := cfg.BuildLogger(appID); err != nil {
		log.Fatalf("unable to build logger, error %v", err)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	cfg.SetCoreFlags(rootCmd, appID)

	var i string
	i = *rootCmd.PersistentFlags().StringP("resync-interval", "r", "5s", "informer resync interval")
	var err error
	if p := os.Getenv("RESYNC_INTERVAL"); p != "" {
		i = p
	}
	reSyncInterval, err = time.ParseDuration(i)
	if err != nil {
		log.Fatalf("Invalid interval duration %s, error %v", i, err)
	}
}
