package main

import (
	"fmt"
	"os"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

// WebhookConfig holds the unmarshalled structure from the config map
type WebhookConfig struct {
	NamespaceSelector *metav1.LabelSelector `yaml:"namespaceSelector"`
	MatchDomains      []string              `yaml:"matchDomains"`
}

func LoadConfigFromFile(path string) (*WebhookConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg WebhookConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal WebhookConfig: %w", err)
	}

	return &cfg, nil
}
