package handlers

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var authRequestsTotal = promauto.NewCounter(prometheus.CounterOpts{
	Name: "auth_requests_total",
	Help: "Total number of auth requests",
})
