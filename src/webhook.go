package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	routev1 "github.com/openshift/api/route/v1"
	"go.uber.org/zap"
	admissionv1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
)

func RouteValidatorHandler(cfgManager *ConfigManager, client kubernetes.Interface, log *zap.SugaredLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var admissionReview admissionv1.AdmissionReview
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "could not read request", http.StatusBadRequest)
			return
		}

		if err := json.Unmarshal(body, &admissionReview); err != nil {
			http.Error(w, "could not decode admission review", http.StatusBadRequest)
			return
		}

		cfg := cfgManager.Get()

		review := admissionv1.AdmissionReview{
			TypeMeta: admissionReview.TypeMeta,
		}
		review.Response = validateRoute(r.Context(), admissionReview.Request, cfg, client)
		review.Response.UID = admissionReview.Request.UID

		respBytes, _ := json.Marshal(review)
		w.Header().Set("Content-Type", "application/json")
		w.Write(respBytes)
	}
}

func validateRoute(ctx context.Context, req *admissionv1.AdmissionRequest, cfg *WebhookConfig, client kubernetes.Interface) *admissionv1.AdmissionResponse {
	if req.Kind.Kind != "Route" || (req.Operation != admissionv1.Create && req.Operation != admissionv1.Update) {
		return allow()
	}

	var route routev1.Route
	if err := json.Unmarshal(req.Object.Raw, &route); err != nil {
		return deny(fmt.Sprintf("could not unmarshal Route object: %v", err))
	}

	ns, err := client.CoreV1().Namespaces().Get(ctx, req.Namespace, metav1.GetOptions{})
	if err != nil {
		return deny(fmt.Sprintf("could not get namespace: %v", err))
	}

	selector, err := metav1.LabelSelectorAsSelector(cfg.NamespaceSelector)
	if err != nil {
		log.Errorf("Failed to parse namespaceSelector: %v", err)
		return allow()
	}

	log.Debugf("cfg: %+v", cfg)
	log.Debugf("Parsed selector: %v", selector)
	log.Debugf("Namespace: %s - Matched: %v", ns.Name, selector.Matches(labels.Set(ns.Labels)))
	log.Debugf("Route: %s", req.Name)
	log.Debugf("Labels: %v", ns.Labels)

	if !selector.Matches(labels.Set(ns.Labels)) {
		return allow()
	}

	if len(cfg.MatchDomains) > 0 && !matchesAnyDomain(&route, cfg.MatchDomains) {
		return allow()
	}

	if !hasValidHostnameSuffix(&route) {
		return deny("route host must include the namespace")
	}

	return allow()
}

func allow() *admissionv1.AdmissionResponse {
	return &admissionv1.AdmissionResponse{Allowed: true}
}

func deny(message string) *admissionv1.AdmissionResponse {
	return &admissionv1.AdmissionResponse{
		Allowed: false,
		Result:  &metav1.Status{Message: message},
	}
}

func matchesAnyDomain(route *routev1.Route, matchDomains []string) bool {
	for _, element := range matchDomains {
		domain := element
		if !strings.HasPrefix(element, ".") {
			domain = "." + element
		}
		if strings.HasSuffix(route.Spec.Host, domain) {
			return true
		}
	}

	return false
}

func hasValidHostnameSuffix(route *routev1.Route) bool {
	hostname := route.Spec.Host
	nsSuffix := "-" + route.Namespace + "."

	return strings.Contains(hostname, nsSuffix)
}
