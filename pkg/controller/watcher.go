package controller

import (
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"go.uber.org/zap"
)

func WatchConfigFile(configFilePath string, cm *ConfigManager, log *zap.SugaredLogger) {
	dir := filepath.Dir(configFilePath)

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatalf("failed to create fsnotify watcher: %v", err)
	}
	defer watcher.Close()

	if err := watcher.Add(dir); err != nil {
		log.Fatalf("failed to watch directory %s: %v", dir, err)
	}

	log.Infof("Watching config dir for changes: %s", dir)

	var debounceTimer *time.Timer
	var timerMu sync.Mutex

	triggerReload := func() {
		if err := cm.LoadFromFile(configFilePath); err != nil {
			log.Errorf("failed to reload config: %v", err)
		} else {
			log.Infof("config reloaded successfully")
		}
	}

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}

			log.Debugf("fsnotify event: %s | ops: %s", event.Name, event.Op.String())

			// Kubernetes updates the file via atomic rename
			if event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Remove|fsnotify.Rename|fsnotify.Chmod) != 0 {
				if filepath.Base(event.Name) == filepath.Base(configFilePath) {
					log.Infof("Config file changed: %s", event)

					timerMu.Lock()
					if debounceTimer != nil {
						debounceTimer.Stop()
					}
					debounceTimer = time.AfterFunc(200*time.Millisecond, func() {
						triggerReload()
					})
					timerMu.Unlock()
				}
			}

		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Errorf("fsnotify error: %v", err)
		}
	}
}
