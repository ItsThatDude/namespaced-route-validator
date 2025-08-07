package main

import (
	"context"
	"fmt"
	"log"

	"gopkg.in/yaml.v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// WebhookConfig holds the unmarshalled structure from the config map
type WebhookConfig struct {
	NamespaceSelector *metav1.LabelSelector `yaml:"namespaceSelector"`
}

// LoadConfigFromConfigMap reads and parses the config from the ConfigMap
func LoadConfigFromConfigMap(client kubernetes.Interface, namespace string) (*WebhookConfig, error) {
	ctx := context.TODO()
	cm, err := client.CoreV1().ConfigMaps(namespace).Get(ctx, "route-validator-config", metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get ConfigMap: %w", err)
	}

	rawConfig, ok := cm.Data["config"]
	if !ok || rawConfig == "" {
		return nil, fmt.Errorf("config not found in ConfigMap data")
	}

	log.Printf("Raw config YAML:\n%s", rawConfig)

	var cfg WebhookConfig
	if err := yaml.Unmarshal([]byte(rawConfig), &cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal WebhookConfig: %w", err)
	}

	log.Printf("Parsed NamespaceSelector: %+v", cfg.NamespaceSelector)

	return &cfg, nil
}
