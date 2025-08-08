package main

import (
	"fmt"
	"os"
	"sync"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

// WebhookConfig holds the unmarshalled structure from the config map
type WebhookConfig struct {
	NamespaceSelector *metav1.LabelSelector `yaml:"namespaceSelector"`
	MatchDomains      []string              `yaml:"matchDomains"`
}

type ConfigManager struct {
	mu     sync.RWMutex
	config *WebhookConfig
}

func (cm *ConfigManager) Get() *WebhookConfig {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.config
}

func (cm *ConfigManager) LoadFromFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var cfg WebhookConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return fmt.Errorf("failed to unmarshal WebhookConfig: %w", err)
	}

	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.config = &cfg
	return nil
}
