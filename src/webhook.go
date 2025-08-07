package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	routev1 "github.com/openshift/api/route/v1"
	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
)

func RouteValidatorHandler(cfg *WebhookConfig, client kubernetes.Interface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var admissionReview admissionv1.AdmissionReview
		body, err := ioutil.ReadAll(r.Body)
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
		review.Response = validateRoute(admissionReview.Request, cfg, client)
		review.Response.UID = admissionReview.Request.UID

		respBytes, _ := json.Marshal(review)
		w.Header().Set("Content-Type", "application/json")
		w.Write(respBytes)
	}
}

func validateRoute(req *admissionv1.AdmissionRequest, cfg *WebhookConfig, client kubernetes.Interface) *admissionv1.AdmissionResponse {
	if req.Kind.Kind != "Route" || (req.Operation != admissionv1.Create && req.Operation != admissionv1.Update) {
		return &admissionv1.AdmissionResponse{Allowed: true}
	}

	var route routev1.Route
	if err := json.Unmarshal(req.Object.Raw, &route); err != nil {
		return &admissionv1.AdmissionResponse{
			Allowed: false,
			Result:  &runtime.Status{Message: fmt.Sprintf("could not unmarshal Route object: %v", err)},
		}
	}

	ns, err := client.CoreV1().Namespaces().Get(req.Namespace)
	if err != nil {
		return &admissionv1.AdmissionResponse{
			Allowed: false,
			Result:  &runtime.Status{Message: fmt.Sprintf("could not get namespace: %v", err)},
		}
	}

	labelValue, ok := ns.Labels[cfg.NamespaceLabelKey]
	if !ok || labelValue != cfg.NamespaceLabelValue {
		// Namespace not labeled for enforcement
		return &admissionv1.AdmissionResponse{Allowed: true}
	}

	if !strings.Contains(route.Spec.Host, route.Namespace) {
		return &admissionv1.AdmissionResponse{
			Allowed: false,
			Result: &runtime.Status{
				Message: "route host must include the namespace",
			},
		}
	}

	return &admissionv1.AdmissionResponse{Allowed: true}
}