package controller

import (
	"net/http"
	"os"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var log *zap.SugaredLogger

func getLogLevelFromEnv() zapcore.Level {
	levelStr := strings.ToLower(os.Getenv("LOG_LEVEL"))

	switch levelStr {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn", "warning":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	case "dpanic":
		return zapcore.DPanicLevel
	case "panic":
		return zapcore.PanicLevel
	case "fatal":
		return zapcore.FatalLevel
	default:
		return zapcore.InfoLevel
	}
}

func setupLogger() *zap.Logger {
	cfg := zap.NewProductionConfig()
	cfg.Level = zap.NewAtomicLevelAt(getLogLevelFromEnv())

	logger, err := cfg.Build()
	if err != nil {
		panic(err)
	}

	return logger
}

func Main() error {
	raw := setupLogger()
	defer raw.Sync()
	log = raw.Sugar()

	// Load in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Fatalf("failed to get cluster config: %v", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("failed to create clientset: %v", err)
	}

	configPath := "/etc/route-validator/config.yaml"

	configManager := &ConfigManager{}
	if err := configManager.LoadFromFile(configPath); err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	go WatchConfigFile(configPath, configManager, log)

	log.Info("Starting server on :8443")
	http.HandleFunc("/validate", RouteValidatorHandler(configManager, clientset, log))
	server := &http.Server{
		Addr: ":8443",
	}
	log.Fatal(server.ListenAndServeTLS("/certs/tls.crt", "/certs/tls.key"))

	return nil
}
