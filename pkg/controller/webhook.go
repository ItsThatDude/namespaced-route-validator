package controller

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
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
)

type objectValidator[T any] struct {
	Kind          string
	MatchDomainFn func(*T, []string) bool
	HostnamesFn   func(*T) []string
}

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

		review.Response = validate(r.Context(), admissionReview.Request, cfg, client, log)
		review.Response.UID = admissionReview.Request.UID

		respBytes, _ := json.Marshal(review)
		w.Header().Set("Content-Type", "application/json")
		w.Write(respBytes)
	}
}

func validate(ctx context.Context, req *admissionv1.AdmissionRequest, cfg *WebhookConfig, client kubernetes.Interface, log *zap.SugaredLogger) *admissionv1.AdmissionResponse {
	if req.Operation != admissionv1.Create && req.Operation != admissionv1.Update {
		return allow()
	}

	switch req.Kind.Kind {
	case "Ingress":
		return validateIngress(ctx, req, cfg, client, log)
	case "Route":
		return validateRoute(ctx, req, cfg, client, log)
	default:
		return allow()
	}
}

func validateIngress(ctx context.Context, req *admissionv1.AdmissionRequest, cfg *WebhookConfig, client kubernetes.Interface, log *zap.SugaredLogger) *admissionv1.AdmissionResponse {
	if !isKindAndOp(req, "Ingress", admissionv1.Create, admissionv1.Update) {
		return allow()
	}
	return validateObject(ctx, req, cfg, client, log, objectValidator[networkingv1.Ingress]{
		Kind:          "Ingress",
		MatchDomainFn: ingressMatchesAnyDomain,
		HostnamesFn: func(i *networkingv1.Ingress) []string {
			hosts := []string{}
			for _, rule := range i.Spec.Rules {
				hosts = append(hosts, rule.Host)
			}
			return hosts
		},
	})
}

func validateRoute(ctx context.Context, req *admissionv1.AdmissionRequest, cfg *WebhookConfig, client kubernetes.Interface, log *zap.SugaredLogger) *admissionv1.AdmissionResponse {
	if !isKindAndOp(req, "Route", admissionv1.Create, admissionv1.Update) {
		return allow()
	}
	return validateObject(ctx, req, cfg, client, log, objectValidator[routev1.Route]{
		Kind:          "Route",
		MatchDomainFn: routeMatchesAnyDomain,
		HostnamesFn: func(r *routev1.Route) []string {
			return []string{r.Spec.Host}
		},
	})
}

func validateObject[T any](
	ctx context.Context,
	req *admissionv1.AdmissionRequest,
	cfg *WebhookConfig,
	client kubernetes.Interface,
	log *zap.SugaredLogger,
	v objectValidator[T],
) *admissionv1.AdmissionResponse {
	if !isKindAndOp(req, v.Kind, admissionv1.Create, admissionv1.Update) {
		return allow()
	}

	var obj T
	if err := json.Unmarshal(req.Object.Raw, &obj); err != nil {
		return deny(fmt.Sprintf("could not unmarshal %s object: %v", v.Kind, err))
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

	if !selector.Matches(labels.Set(ns.Labels)) {
		return allow()
	}

	if len(cfg.MatchDomains) > 0 && !v.MatchDomainFn(&obj, cfg.MatchDomains) {
		return allow()
	}

	for _, host := range v.HostnamesFn(&obj) {
		if !validateHostnameSuffix(req.Namespace, host) {
			return deny(fmt.Sprintf("%s host must include the namespace", v.Kind))
		}
	}

	return allow()
}

func isKindAndOp(req *admissionv1.AdmissionRequest, kind string, ops ...admissionv1.Operation) bool {
	if req.Kind.Kind != kind {
		return false
	}
	for _, op := range ops {
		if req.Operation == op {
			return true
		}
	}
	return false
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

func ingressMatchesAnyDomain(ingress *networkingv1.Ingress, matchDomains []string) bool {
	for _, rule := range ingress.Spec.Rules {
		host := rule.Host

		for _, element := range matchDomains {
			domain := element
			if !strings.HasPrefix(element, ".") {
				domain = "." + element
			}
			if strings.HasSuffix(host, domain) {
				return true
			}
		}
	}

	return false
}

func routeMatchesAnyDomain(route *routev1.Route, matchDomains []string) bool {
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

func validateHostnameSuffix(namespace string, hostname string) bool {
	nsSuffix := "-" + namespace + "."

	return strings.Contains(hostname, nsSuffix)
}
