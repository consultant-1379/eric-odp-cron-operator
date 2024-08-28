package main

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"eric-odp-cronwrapper/internal/config"
	"eric-odp-cronwrapper/internal/httputils"
	"eric-odp-cronwrapper/internal/logctl"
	"eric-odp-cronwrapper/internal/metric"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	log "eric-odp-cronwrapper/internal/logger"
	openapiclient "odpfactory"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

var (
	appConfig = config.GetConfig()

	//nolint:unused // not unused param
	tracer trace.Tracer
)

func init() {
	log.SetDefaultFields(
		log.Fields{"service_id": ServiceID, "version": Version})
	if err := log.SetLevel(log.InfoLevel); err != nil {
		fmt.Println("Unable to set log level.Exiting")
		os.Exit(1)
	}
	metric.SetupMetric()
	tracer = otel.Tracer(ServiceID)
}

// onChangeCb The callback is invoked by logctl whenever the configmap is
// updated, the callback now get the config related to its service and
// sets the log parameters.
func onChangeCb() error {
	var err error

	logConfig, err := logctl.GetLogControl(appConfig.LogControlFile)
	if err != nil {
		log.WithFields(log.Fields{"err": err}).
			Error("unable to reload log config")

		return fmt.Errorf("unable to get log control file: %w", err)
	}

	// Only setting we have now is logLevel
	for _, entry := range logConfig {
		if entry.Container == ServiceID {
			err = log.SetLevel(entry.Severity)
			if err != nil {
				return fmt.Errorf("unable to set log level: %w", err)
			}
		}
	}

	log.WithFields(log.Fields{"logConfig": logConfig}).
		Info("Updated log configuration, reloaded log config")

	return nil
}

func initMetricsProvider() {
	addr := fmt.Sprintf(":%d", appConfig.MetricsPort)
	mux := http.NewServeMux()

	mux.Handle("/metrics",
		promhttp.HandlerFor(metric.Registry, promhttp.HandlerOpts{}))

	srvConfig := &httputils.ServerConfig{
		Addr:       addr,
		ServerName: "metrics_server",
		Handler:    mux,
	}

	srv := httputils.NewServer(srvConfig)

	log.WithFields(log.Fields{"addr": addr}).Debug("Starting metrics server")
	log.Fatal(srv.ListenAndServe())
}

// It must be possible to start multiple instances of long running cron jobs.
// This method generates a unique pod instanceId.
func generateUniqueInstanceId(uniqueCronId string) string {
	md5Hash := md5.New()
	currentTime := strconv.FormatInt(time.Now().UnixNano(), 10)

	md5Hash.Write([]byte(uniqueCronId + ":" + currentTime))
	return hex.EncodeToString(md5Hash.Sum(nil))
}

// The Command to run is passed as a string to the ODP Factory, so all quotes must be delimited.
func delimitCommand(command string) string {
	delimitedCommand := strings.Replace(command, `"`, `\\\"`, -1)
	log.WithFields(log.Fields{"original Command": command, "escaped Command": delimitedCommand}).Debug("Encoding all double quotes")
	return delimitedCommand
}

// This method reads Environment variables detailing the command to run for the user. It then connects to the ODP Factory and requests an ODP to run the command.
func requestODPFactoryForODP() {
	log.WithFields(log.Fields{"username": appConfig.Username, "command": appConfig.Command, "hostname": appConfig.Hostname, "uniqueCronId": appConfig.UniqueCronId}).Debug("Start CronWrapper.")

	//Generate a unique ID to allow the same overlapping cron command to be executed.
	uniquePodId := generateUniqueInstanceId(appConfig.UniqueCronId)
	log.Debug("Generated uniquePodId: " + uniquePodId)

	// Need to delimit all quotation marks.
	delimitedCommand := delimitCommand(appConfig.Command)

	// Configure the ODP Factory Openapi Client.
	odpPostRequest := *openapiclient.NewOdpPostRequest(appConfig.Username, "cron")
	odpPostRequest.SetInstanceid(uniquePodId)
	data := map[string]interface{}{"command": delimitedCommand, "tokentype": "odptoken"}
	odpPostRequest.SetData(data)

	configuration := openapiclient.NewConfiguration()
	//Configure the ODP Factory URL.
	configuration.Host = appConfig.OdpFactoryHost + ":" + appConfig.OdpFactoryPort
	configuration.Scheme = appConfig.OdpFactorySchema
	//TODO: Configure TLS on a custom httpClient and set in the configuration.
	apiClient := openapiclient.NewAPIClient(configuration)

	log.WithFields(log.Fields{"configuration.Host": configuration.Host, "configuration.Scheme: ": configuration.Scheme, "apiClient": apiClient, "odpPostRequest": odpPostRequest, "hostname": appConfig.Hostname}).Info("Request to ODP factory for Cron ODP")

	sleepTime := appConfig.RetrySleepTime
	maxSleepTime := appConfig.MaxSleepTime
	for {
		// Execute the API call.
		resp, r, err := apiClient.DefaultAPI.OdpPost(context.Background()).OdpPostRequest(odpPostRequest).Execute()
		if err != nil {
			log.WithFields(log.Fields{"error": err}).Info("Fatal Error communicating with ODP factory")
			log.Fatal("Error executing API request: %v", err)
		}
		if resp != nil {
			log.WithFields(log.Fields{"Resultcode": resp.Resultcode, "podname": resp.Podname, "Podips": resp.Podips, "error": resp.Error}).Info("Response returned from the ODP factory")
		}

		if resp.Error != nil && *resp.Error != "" {
			// ODP factory returned an error. Log the error and exit the application.
			log.WithFields(log.Fields{"error": resp.Error}).Error("ODP factory error")
			// This fatal log causes application to exit.
			log.Fatal("Exiting as ODP factory reported error: %v", resp.Error)
		}
		defer func() {
			err := r.Body.Close()
			if err != nil {
				log.Error("Error closing response body: %v", err)
			}
		}()

		// Check if response is ready.
		if resp.Resultcode != nil && *resp.Resultcode == 0 {
			log.WithFields(log.Fields{"Resultcode": resp.Resultcode, "podname": resp.Podname, "Podips": resp.Podips}).Info("ODP Factory successfully created pod")
			break
		}

		duration := time.Duration(sleepTime) * time.Second

		// Sleep for 2 seconds before the next iteration.
		log.WithFields(log.Fields{"sleepTime": sleepTime}).Info("ODP Factory reports the ODP is not ready.Sleeping before next check")
		time.Sleep(duration)

		if sleepTime <= maxSleepTime {
			sleepTime += 2
		} else if sleepTime >= maxSleepTime {
			log.WithFields(log.Fields{"sleepTime": sleepTime, "Resultcode": resp.Resultcode, "podname": resp.Podname, "Podips": resp.Podips}).Error("ODP is not ready within max allowed time.Exiting the current schedule without execution.")
			break
		}
	}

}

func main() {

	_, cancel := context.WithCancel(context.Background())

	if appConfig.LogControlFile != "" {
		go logctl.Watch(appConfig.LogControlFile, onChangeCb)
	}

	go initMetricsProvider()

	requestODPFactoryForODP()
	cancel()

}
