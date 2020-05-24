/*
© Copyright IBM Corporation 2018, 2019

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package metrics contains code to provide metrics for the queue manager
package metrics

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/ibm-messaging/mq-container/pkg/logger"
	"github.com/ibm-messaging/mq-container/internal/ready"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	defaultPort = "9157"
)

var (
	metricsEnabled = false
	metricsServer  = &http.Server{Addr: ":" + defaultPort}
)

// GatherMetrics gathers metrics for the queue manager
func GatherMetrics(qmName string, log *logger.Logger) {

	// If running in standby mode - wait until the queue manager becomes active
	for {
		active, _ := ready.IsRunningAsActiveQM(qmName)
		if active {
			break
		}
		time.Sleep(requestTimeout * time.Second)
	}

	metricsEnabled = true

	err := startMetricsGathering(qmName, log)
	if err != nil {
		log.Errorf("Metrics Error: %s", err.Error())
		StopMetricsGathering(log)
	}
}

// startMetricsGathering starts gathering metrics for the queue manager
func startMetricsGathering(qmName string, log *logger.Logger) error {

	defer func() {
		if r := recover(); r != nil {
			log.Errorf("Metrics Error: %v", r)
		}
	}()

	log.Println("Starting metrics gathering")

	// Start processing metrics
	go processMetrics(log, qmName)

	// Wait for metrics to be ready before starting the Prometheus handler
	<-startChannel

	// Register metrics
	metricsExporter := newExporter(qmName, log)
	err := prometheus.Register(metricsExporter)
	if err != nil {
		return fmt.Errorf("Failed to register metrics: %v", err)
	}

	// Setup HTTP server to handle requests from Prometheus
	http.Handle("/metrics", prometheus.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		// #nosec G104
		w.Write([]byte("Status: METRICS ACTIVE"))
	})

	go func() {
		err = metricsServer.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			log.Errorf("Metrics Error: Failed to handle metrics request: %v", err)
			StopMetricsGathering(log)
		}
	}()

	return nil
}

// StopMetricsGathering stops gathering metrics for the queue manager
func StopMetricsGathering(log *logger.Logger) {

	if metricsEnabled {

		// Stop processing metrics
		stopChannel <- true

		// Shutdown HTTP server
		timeout, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		err := metricsServer.Shutdown(timeout)
		if err != nil {
			log.Errorf("Failed to shutdown metrics server: %v", err)
		}
	}
}
