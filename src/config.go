package main

import (
	"fmt"
	"os"

	"k8s.io/client-go/kubernetes"
)

type WebhookConfig struct {
	NamespaceLabelKey   string
	NamespaceLabelValue string
}

func LoadConfigFromConfigMap(client kubernetes.Interface, namespace string) (*WebhookConfig, error) {
	cm, err := client.CoreV1().ConfigMaps(namespace).Get("route-validator-config")
	if err != nil {
		return nil, err
	}

	labelKey := cm.Data["namespaceLabelKey"]
	labelValue := cm.Data["namespaceLabelValue"]

	if labelKey == "" || labelValue == "" {
		return nil, fmt.Errorf("configmap missing keys")
	}

	return &WebhookConfig{
		NamespaceLabelKey:   labelKey,
		NamespaceLabelValue: labelValue,
	}, nil
}