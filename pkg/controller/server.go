package controller

import (
	"io"
	"net/http"
	"time"

	flag "github.com/spf13/pflag"
	"go.uber.org/zap"
	"k8s.io/client-go/kubernetes"
)

var (
	listenAddr   = flag.String("listen-addr", ":8080", "HTTP listen address.")
	readTimeout  = flag.Duration("read-timeout", 2*time.Minute, "HTTP request timeout.")
	writeTimeout = flag.Duration("write-timeout", 2*time.Minute, "HTTP response timeout.")
)

func httpserver(configManager *ConfigManager, clientset *kubernetes.Clientset, log *zap.SugaredLogger) *http.Server {
	mux := http.NewServeMux()

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, err := io.WriteString(w, "ok\n")
		if err != nil {
			log.Fatal(err)
		}
	})

	mux.HandleFunc("/validate", RouteValidatorHandler(configManager, clientset, log))

	server := http.Server{
		Addr:              *listenAddr,
		Handler:           mux,
		ReadTimeout:       *readTimeout,
		ReadHeaderTimeout: *readTimeout,
		WriteTimeout:      *writeTimeout,
	}

	log.Info("HTTP server serving", "addr", server.Addr)

	go func() {
		err := server.ListenAndServeTLS("/certs/tls.crt", "/certs/tls.key")
		log.Error("HTTP server exiting", "error", err)
	}()

	return &server
}
