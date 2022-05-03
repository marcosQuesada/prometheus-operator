package service

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	updatesProcessed = promauto.NewCounter(prometheus.CounterOpts{
		Name: "prometheus_operator_updates_total",
		Help: "The total number of operator processed updates",
	})

	deletesProcessed = promauto.NewCounter(prometheus.CounterOpts{
		Name: "prometheus_operator_deletes_total",
		Help: "The total number of operator processed deletes",
	})

	statusUpdatesProcessed = promauto.NewCounter(prometheus.CounterOpts{
		Name: "prometheus_operator_status_updates_total",
		Help: "The total number of operator processed status updates",
	})
)

var (
	conciliationProcessed = promauto.NewCounter(prometheus.CounterOpts{
		Name: "prometheus_conciliation_iteration_total",
		Help: "The total number of conciliation iterations",
	})

	conciliationProcessedErrors = promauto.NewCounter(prometheus.CounterOpts{
		Name: "prometheus_conciliation_iteration_total_errors",
		Help: "The total number of conciliation iterations with error",
	})
)

