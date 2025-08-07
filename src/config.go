package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// WebhookConfig holds the unmarshalled structure from the config map
type WebhookConfig struct {
	NamespaceSelector *metav1.LabelSelector `yaml:"namespaceSelector"`
}

func LoadConfigFromFile(path string) (*WebhookConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg WebhookConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
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

	fmt.Println("Raw YAML lines:")
	for i, line := range strings.Split(rawConfig, "\n") {
		fmt.Printf("%2d: %q\n", i+1, line)
	}

	var cfg WebhookConfig
	if err := yaml.Unmarshal([]byte(rawConfig), &cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal WebhookConfig: %w", err)
	}

	log.Printf("Parsed NamespaceSelector: %+v", cfg.NamespaceSelector)

	return &cfg, nil
}
