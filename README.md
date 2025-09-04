# Namespaced Route Validator

Namespaced Route Validator is a Kubernetes Validating Webhook / Admission Controller that simply enforces the use of the namespace in Kubernetes Ingress and OCP Route hostnames.

## Installation

### Deployment using Helm

```bash
helm repo add itsthatdude https://itsthatdude.github.io/helm-charts
helm repo update
helm upgrade --install namespaced-route-validator itsthatdude/namespaced-route-validator
```
You can customize the values of the helm deployment by using the following Values:

#### Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| deployment.customLivenessProbe | object | `{}` | Specify a custom liveness probe |
| deployment.customReadinessProbe | object | `{}` | Specify a custom readiness probe |
| deployment.customStartupProbe | object | `{}` | Specify a custom startup probe |
| deployment.image.pullPolicy | string | `"IfNotPresent"` | Specify the image pull policy |
| deployment.image.repository | string | `"namespaced-route-validator"` | Specify the repo/image name to pull the image from |
| deployment.image.tag | string | `nil` | Specify the image tag to pull, if blank will pull the version from the chart's AppVersion |
| deployment.livenessProbe | object | N/A | Override the settings for the default liveness probe |
| deployment.logLevel | string | `"info"` | Specify the log level of the controller |
| deployment.readinessProbe | object | N/A | Override the settings for the default readiness probe |
| deployment.replicaCount | int | `1` | Specify the replica count for the deployment |
| deployment.startupProbe | object | N/A | Override the settings for the default startup probe |
| service.annotations | object | `{}` | Specify any additional annotations to add to the service |
| service.name | string | `"namespaced-route-validator-service"` | Specify the name of the service |
| service.port | int | `443` | Specify the port of the service |
| tls.cert | string | `""` | Specify a PEM encoded cert to secure the controller |
| tls.existingSecret | string | `""` | If set, this existing secret will be used to secure the controller |
| tls.key | string | `""` | Specify a PEM encoded key to secure the controller |
| validator.matchDomains | list | `[]` | This specifies which base domains the admission controller applies to |
| validator.namespaceSelector | object | `{"matchLabels":{"enforce-route-check":"true"}}` | This specifies the namespace selector the admission controller applies to |
| validator.subdomainLabel | string | `"route-validator.antware.xyz/subdomain"` | This specifies which label on the namespace to use as the required subdomain.<br /> If a blank string is provided, the validator will use the namespace as the required subdomain. |
| validator.validateIngress | bool | `true` | Validate Ingress objects |
| validator.validateRoutes | bool | `true` | Validate OpenShift/OKD Route objects |
| webhook.annotations | object | `{}` | Specify any additional annotations to add to the webhook |
| webhook.caBundle | string | `""` | Specify the CA Bundle that signs the controller's certificate |
| webhook.failurePolicy | string | `"Fail"` | Specify the failure policy |
| webhook.name | string | `"route-validator.antware.xyz"` | Specify the name of the Webhook |

## Usage

### Annotate the target `namespace`
Add the configured label to the namespace to enable the controller to enforce the hostname validation.  
The default label is `enforce-route-check: "true"`

### Done!
The controller will monitor any namespaces with the configured label.  
If any Ingresses or Routes are created within the namespace, and they don't match the format *-\<namespace\>.\<base-domain\> they will be rejected.