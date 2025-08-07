package main

import (
	"log"
	"net/http"
	"os"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func main() {
	// Load in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Fatalf("failed to get cluster config: %v", err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("failed to create clientset: %v", err)
	}

	cfg, err := LoadConfigFromConfigMap(clientset, os.Getenv("POD_NAMESPACE"))
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	log.Println("Starting server on :8443")
	http.HandleFunc("/validate", RouteValidatorHandler(cfg, clientset))
	server := &http.Server{
		Addr: ":8443",
	}
	log.Fatal(server.ListenAndServeTLS("/certs/tls.crt", "/certs/tls.key"))
}