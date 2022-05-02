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

	// LogLevel select logging level
	LogLevel string

	// Env defines environment where the app is running
	Env string

	// HttpPort exposed Http Port
	HttpPort string
)

// BuildLogger builds logger with required LogLevel, taints log traces with App ID
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

// SetCoreFlags sets root application flags
func SetCoreFlags(cmd *cobra.Command, service string) {
	cmd.PersistentFlags().StringVar(&LogLevel, "log-level", "info", "logging level")
	cmd.PersistentFlags().StringVar(&Env, "env", "dev", "environment where the application is running")
	cmd.PersistentFlags().StringVar(&HttpPort, "http-port", "9090", "http server port")
	if p := os.Getenv("HTTP_PORT"); p != "" {
		HttpPort = p
	}
}
