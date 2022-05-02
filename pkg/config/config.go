package config

import (
	"fmt"
	logger "github.com/marcosQuesada/prometheus-operator/pkg/log"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
)

var (
	// Commit hash on current version
	Commit string

	// Date on current release build
	Date string

	LogLevel string
	Env      string
	HttpPort string
)

func BuildLogger(appID string) error {
	level, err := log.ParseLevel(LogLevel)
	if err != nil {
		return fmt.Errorf("unexpected error parsing level, error %v", err)
	}
	log.SetLevel(level)
	log.SetReportCaller(true)
	log.SetFormatter(logger.PrettifiedFormatter())
	log.AddHook(logger.NewGlobalFieldHook(appID, Env))

	return nil
}

func SetCoreFlags(cmd *cobra.Command, service string) {
	cmd.PersistentFlags().StringVar(&LogLevel, "log-level", "info", "logging level")
	cmd.PersistentFlags().StringVar(&Env, "env", "dev", "environment where the application is running") // @TODO
	cmd.PersistentFlags().StringVar(&HttpPort, "http-port", "9090", "http server port")
	if p := os.Getenv("HTTP_PORT"); p != "" {
		HttpPort = p
	}
}
