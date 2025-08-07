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

func RouteValidatorHandler(cfg *WebhookConfig, client kubernetes.Interface, log *zap.SugaredLogger) http.HandlerFunc {
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
		return &admissionv1.AdmissionResponse{Allowed: true}
	}

	var route routev1.Route
	if err := json.Unmarshal(req.Object.Raw, &route); err != nil {
		return &admissionv1.AdmissionResponse{
			Allowed: false,
			Result:  &metav1.Status{Message: fmt.Sprintf("could not unmarshal Route object: %v", err)},
		}
	}

	ns, err := client.CoreV1().Namespaces().Get(ctx, req.Namespace, metav1.GetOptions{})
	if err != nil {
		return &admissionv1.AdmissionResponse{
			Allowed: false,
			Result:  &metav1.Status{Message: fmt.Sprintf("could not get namespace: %v", err)},
		}
	}

	selector, err := metav1.LabelSelectorAsSelector(cfg.NamespaceSelector)
	if err != nil {
		// Failed to parse label selector, default to allowing
		log.Errorf("Failed to parse namespaceSelector: %v", err)
		return &admissionv1.AdmissionResponse{Allowed: true}
	}

	log.Debugf("cfg: %+v", cfg)
	log.Debugf("Parsed selector: %v", selector)

	matched := selector.Matches(labels.Set(ns.Labels))
	log.Debugf("Namespace: %s - Matched: %v", ns.Name, matched)
	log.Debugf("Route: %s", req.Name)
	log.Debugf("Labels: %v", ns.Labels)

	if !matched {
		// Namespace labels do not match, skip validation
		return &admissionv1.AdmissionResponse{Allowed: true}
	}

	if !strings.Contains(route.Spec.Host, route.Namespace) {
		return &admissionv1.AdmissionResponse{
			Allowed: false,
			Result: &metav1.Status{
				Message: "route host must include the namespace",
			},
		}
	}

	return &admissionv1.AdmissionResponse{Allowed: true}
}
