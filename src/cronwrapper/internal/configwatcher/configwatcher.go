package configwatcher

import (
	"fmt"
	"path/filepath"

	log "eric-odp-cronwrapper/internal/logger"

	"github.com/fsnotify/fsnotify"
)

// Event Wraps the fsnotify.Event.
type Event struct {
	fsnotify.Event
}

// Watch The method watch for config map changes for the filePath
// The k8s configmap are mounted on the path mentioned by the containers
// volumeMount directory and when the configmap is updated,the entire directory
// is re-created (clone-rename-remove), so watching for ..data link in the mount
// path is good enough for checking for changes.
func Watch(filePath string, notifyCh chan Event) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("creating new watch error: %w", err)
	}
	configDir := filepath.Dir(filePath)
	dataFile := filepath.Join(configDir, "..data")
	if err = watcher.Add(configDir); err != nil {
		return fmt.Errorf("unable to add directory to watch: %w", err)
	}
	ctxLog := log.WithFields(log.Fields{
		"watchdir":   configDir,
		"configdata": dataFile,
	})
	ctxLog.Debug("Added watch")

	createMask := fsnotify.Create
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					ctxLog.WithFields(log.Fields{"err": err}).
						Debug("events channel closed")
					close(notifyCh)

					return
				}
				if filepath.Clean(event.Name) == dataFile &&
					event.Op&createMask != 0 {
					ctxLog.WithFields(log.Fields{"event": event}).
						Debug("configmap updated, recreated ..data")
					notifyCh <- Event{event}
				}
			case werr, ok := <-watcher.Errors:
				if !ok {
					ctxLog.WithFields(log.Fields{"err": werr}).
						Debug("error channel closed")
					close(notifyCh)

					return
				}
				ctxLog.WithFields(log.Fields{"err": werr}).
					Debug("Watch err occurred")
			}
		}
	}()

	return nil
}
