package logctl

import (
	"encoding/json"
	"errors"
	"fmt"
	"eric-odp-cronwrapper/internal/configwatcher"
	"os"
	"time"

	log "eric-odp-cronwrapper/internal/logger"
)

var defaultPollTimeout = 2 * time.Second

var errMissingSchemaJSONFields = errors.New("missing \"container\" or \"severity\" field")

// LogControl The array of log control items.
type LogControl []*LogControlItems

// LogControlItems A log control entry.
type LogControlItems struct {
	// Name of the image for the container producing the log event.
	Container string `json:"container"`

	// Optional list of log events customized filters.
	// eg. to log events generated by a specific POD, class, package, thread, etc.
	CustomFilters []interface{} `json:"customFilters,omitempty"`

	// Log event severity level.
	Severity string `json:"severity"`
}

// unmarshalJSON Unmarshal the logControl file
// https://gerrit.ericsson.se/plugins/gitiles/bssf/adp-log/api/+/refs/heads/master/
// api-logging/src/main/json/logControl.0.json.
func unmarshalJSON(data []byte, logControl *LogControl) error {
	if err := json.Unmarshal(data, &logControl); err != nil {
		return fmt.Errorf("json unmarshal error: %w", err)
	}
	missingField := false
	for _, v := range *logControl {
		if v.Container == "" {
			missingField = true
		}
		if v.Severity == "" {
			missingField = true
		}
	}
	if missingField {
		return errMissingSchemaJSONFields
	}

	return nil
}

// GetLogControl The method unmarshals the json file and returns
// the LogControl data.
func GetLogControl(fileName string) (LogControl, error) {
	data, err := os.ReadFile(fileName)
	if err != nil {
		return nil, fmt.Errorf("get log control read file error: %w", err)
	}
	var logControl LogControl
	err = unmarshalJSON(data, &logControl)
	if err != nil {
		return nil, err
	}
	log.WithFields(log.Fields{"logfile": fileName, "logconfig": logControl}).
		Info("loaded config file")

	return logControl, nil
}

// Watch The method watches for changes to configmap mount path
// and invokes the callback.
func Watch(filePath string, onChangeCallback func() error) {
	notifCh := make(chan configwatcher.Event)
	go func() {
		err := configwatcher.Watch(filePath, notifCh)
		if err != nil {
			log.WithFields(log.Fields{
				"err":       err,
				"logconfig": filePath,
			}).Fatal("Unable to add watch")
		}
	}()

	log.Info("watching for log config updates")
	// throttle, a bit
	tick := time.NewTicker(defaultPollTimeout)
	invokCb := false
	for {
		select {
		case _, ok := <-notifCh:
			if !ok {
				log.Error("notif channel closed")
			}
			log.Debug("config map change event occurred")
			invokCb = true
		case <-tick.C:
			if invokCb {
				if err := onChangeCallback(); err != nil {
					log.WithFields(log.Fields{"err": err}).
						Error("On change call back failed")
				} else {
					invokCb = false
				}
			}
		}
	}
}
