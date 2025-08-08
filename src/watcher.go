package main

import (
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
	"go.uber.org/zap"
)

func WatchConfigFile(path string, cm *ConfigManager, log *zap.SugaredLogger) {
	dir := filepath.Dir(path)

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatalf("failed to create fsnotify watcher: %v", err)
	}
	defer watcher.Close()

	if err := watcher.Add(dir); err != nil {
		log.Fatalf("failed to watch directory %s: %v", dir, err)
	}

	log.Infof("Watching config file for changes: %s", path)

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}

			// Kubernetes updates the file via atomic rename
			if event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Remove|fsnotify.Rename) != 0 {
				if filepath.Base(event.Name) == filepath.Base(path) {
					log.Infof("Config file changed: %s", event)
					// Delay slightly to avoid race with symlink write
					time.Sleep(100 * time.Millisecond)

					if err := cm.LoadFromFile(path); err != nil {
						log.Errorf("Failed to reload config: %v", err)
					} else {
						log.Infof("Successfully reloaded config")
					}
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
