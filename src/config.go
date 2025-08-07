package main

import (
	"context"
	"fmt"

	"gopkg.in/yaml.v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type WebhookConfig struct {
	NamespaceSelector map[string]string `json:"namespaceSelector"`
}

func LoadConfigFromConfigMap(client kubernetes.Interface, namespace string) (*WebhookConfig, error) {
	ctx := context.TODO()
	cm, err := client.CoreV1().ConfigMaps(namespace).Get(ctx, "route-validator-config", metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	nsSelectorYAML, ok := cm.Data["namespaceSelector"]
	if !ok || nsSelectorYAML == "" {
		return nil, fmt.Errorf("namespaceSelector not found in configmap data")
	}

	var selector map[string]string
	if err := yaml.Unmarshal([]byte(nsSelectorYAML), &selector); err != nil {
		return nil, fmt.Errorf("failed to parse namespaceSelector: %w", err)
	}

	return &WebhookConfig{
		NamespaceSelector: selector,
	}, nil
}
