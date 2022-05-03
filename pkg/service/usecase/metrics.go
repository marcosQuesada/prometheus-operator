package usecase

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	emptyProcessed = promauto.NewCounter(prometheus.CounterOpts{
		Name: "prometheus_usecase_creation_empty_total",
		Help: "The total number of processed events on empty state",
	})

	initializingProcessed = promauto.NewCounter(prometheus.CounterOpts{
		Name: "prometheus_usecase_creation_initializing_total",
		Help: "The total number of processed events on initializing state",
	})

	waitingCreationProcessed = promauto.NewCounter(prometheus.CounterOpts{
		Name: "prometheus_usecase_creation_waiting_creation_total",
		Help: "The total number of processed events on waiting creation state",
	})
)

var (
	waitingRemovalProcessed = promauto.NewCounter(prometheus.CounterOpts{
		Name: "prometheus_usecase_deletion_waiting_removal_total",
		Help: "The total number of processed events on waiting removal state",
	})

	terminatingProcessed = promauto.NewCounter(prometheus.CounterOpts{
		Name: "prometheus_usecase_deletion_terminating_total",
		Help: "The total number of processed events on terminating state",
	})
)

var (
	runningProcessed = promauto.NewCounter(prometheus.CounterOpts{
		Name: "prometheus_usecase_running_total",
		Help: "The total number of processed events on running state",
	})

	reloadingProcessed = promauto.NewCounter(prometheus.CounterOpts{
		Name: "prometheus_usecase_reaload_total",
		Help: "The total number of processed events on reloading state",
	})
)
