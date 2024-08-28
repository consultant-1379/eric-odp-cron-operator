/*
Copyright 2024.

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

package main

import (
	"crypto/tls"
	"flag"
	"os"

	"eric-odp-cron-operator/internal/controller"
	uzap "go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	cronlabels "k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

var (
	setupLog   = ctrl.Log.WithName("Main: ")
	namespace  = os.Getenv("NAMESPACE")
	odpcronjob = "com.ericsson.odp.cronjob"
)

func main() {

	configureLogging()
	var metricsAddr string
	var probeAddr string
	var secureMetrics bool
	var enableHTTP2 bool
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")

	flag.BoolVar(&secureMetrics, "metrics-secure", false,
		"If set the metrics endpoint is served securely")
	flag.BoolVar(&enableHTTP2, "enable-http2", false,
		"If set, HTTP/2 will be enabled for the metrics and webhook servers")
	opts := zap.Options{
		Development: false,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))
	disableHTTP2 := func(c *tls.Config) {
		setupLog.Info("disabling http/2")
		c.NextProtos = []string{"http/1.1"}
	}

	tlsOpts := []func(*tls.Config){}
	if !enableHTTP2 {
		tlsOpts = append(tlsOpts, disableHTTP2)
	}

	webhookServer := webhook.NewServer(webhook.Options{
		TLSOpts: tlsOpts,
	})

	log := ctrl.Log.WithName("cron-scheduler")
	cronselector, _ := cronlabels.NewRequirement(odpcronjob, selection.Equals, []string{"true"})
	manager, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Cache: cache.Options{
			DefaultNamespaces: map[string]cache.Config{
				namespace: {},
			},
			DefaultLabelSelector: cronlabels.NewSelector().Add(*cronselector),
		},
		Metrics: metricsserver.Options{
			BindAddress:   metricsAddr,
			SecureServing: secureMetrics,
			TLSOpts:       tlsOpts,
		},
		WebhookServer:          webhookServer,
		HealthProbeBindAddress: probeAddr,
	})
	if err != nil {
		log.Error(err, "could not create manager")
		os.Exit(1)
	}
	cl := manager.GetClient()
	err = (&controller.CronJobReconciler{
		Client: cl,
	}).SetupWithManager(manager)
	if err != nil {
		log.Error(err, "could not setup manager")
		os.Exit(1)
	}

	if err := manager.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := manager.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}
	setupLog.Info("starting manager")
	if err := manager.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "Main: problem running manager")
		os.Exit(1)
	}
}

func configureLogging() {

	productionEncoderConfig := uzap.NewProductionEncoderConfig()
	// Overriding EncodeTime to ISO8601TimeEncoder to get Timestamp in human readable format.
	productionEncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	//Overriding TimeKey, LevelKey, MessageKey and NameKey to match ADP log general design rules
	productionEncoderConfig.TimeKey = "timestamp"
	productionEncoderConfig.LevelKey = "severity"
	productionEncoderConfig.MessageKey = "message"
	productionEncoderConfig.NameKey = "service_id"
	productionEncoder := zapcore.NewJSONEncoder(productionEncoderConfig)

	var logLevelReadFromEnv zapcore.Level
	error := logLevelReadFromEnv.UnmarshalText([]byte(os.Getenv("LOG_LEVEL")))
	if error != nil {
		//Failed to read log level from environment variables.Default Log level will be set anyway.
		logLevelReadFromEnv = uzap.InfoLevel
	}

	atomicLogLevel := uzap.NewAtomicLevelAt(logLevelReadFromEnv)
	logger := zap.New(zap.Encoder(productionEncoder), zap.Level(&atomicLogLevel))
	ctrl.SetLogger(logger)
}
