/*
Copyright The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1beta1

// This file contains a collection of methods that can be used from go-restful to
// generate Swagger API documentation for its models. Please read this PR for more
// information on the implementation: https://github.com/emicklei/go-restful/pull/215
//
// TODOs are ignored from the parser (e.g. TODO(andronat):... || TODO:...) if and only if
// they are on one line! For multiple line or blocks that you want to ignore use ---.
// Any context after a --- is ignored.
//
// Those methods can be generated by using hack/update-generated-swagger-docs.sh

// AUTO-GENERATED FUNCTIONS START HERE. DO NOT EDIT.
var map_HTTPIngressPath = map[string]string{
	"":         "HTTPIngressPath associates a path with a backend. Incoming urls matching the path are forwarded to the backend.",
	"path":     "Path is matched against the path of an incoming request. Currently it can contain characters disallowed from the conventional \"path\" part of a URL as defined by RFC 3986. Paths must begin with a '/'. When unspecified, all paths from incoming requests are matched.",
	"pathType": "PathType determines the interpretation of the Path matching. PathType can be one of the following values: * Exact: Matches the URL path exactly. * Prefix: Matches based on a URL path prefix split by '/'. Matching is\n  done on a path element by element basis. A path element refers is the\n  list of labels in the path split by the '/' separator. A request is a\n  match for path p if every p is an element-wise prefix of p of the\n  request path. Note that if the last element of the path is a substring\n  of the last element in request path, it is not a match (e.g. /foo/bar\n  matches /foo/bar/baz, but does not match /foo/barbaz).\n* ImplementationSpecific: Interpretation of the Path matching is up to\n  the IngressClass. Implementations can treat this as a separate PathType\n  or treat it identically to Prefix or Exact path types.\nImplementations are required to support all path types. Defaults to ImplementationSpecific.",
	"backend":  "Backend defines the referenced service endpoint to which the traffic will be forwarded to.",
}

func (HTTPIngressPath) SwaggerDoc() map[string]string {
	return map_HTTPIngressPath
}

var map_HTTPIngressRuleValue = map[string]string{
	"":      "HTTPIngressRuleValue is a list of http selectors pointing to backends. In the example: http://<host>/<path>?<searchpart> -> backend where where parts of the url correspond to RFC 3986, this resource will be used to match against everything after the last '/' and before the first '?' or '#'.",
	"paths": "A collection of paths that map requests to backends.",
}

func (HTTPIngressRuleValue) SwaggerDoc() map[string]string {
	return map_HTTPIngressRuleValue
}

var map_Ingress = map[string]string{
	"":         "Ingress is a collection of rules that allow inbound connections to reach the endpoints defined by a backend. An Ingress can be configured to give services externally-reachable urls, load balance traffic, terminate SSL, offer name based virtual hosting etc.",
	"metadata": "Standard object's metadata. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata",
	"spec":     "Spec is the desired state of the Ingress. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#spec-and-status",
	"status":   "Status is the current state of the Ingress. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#spec-and-status",
}

func (Ingress) SwaggerDoc() map[string]string {
	return map_Ingress
}

var map_IngressBackend = map[string]string{
	"":            "IngressBackend describes all endpoints for a given service and port.",
	"serviceName": "Specifies the name of the referenced service.",
	"servicePort": "Specifies the port of the referenced service.",
}

func (IngressBackend) SwaggerDoc() map[string]string {
	return map_IngressBackend
}

var map_IngressList = map[string]string{
	"":         "IngressList is a collection of Ingress.",
	"metadata": "Standard object's metadata. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata",
	"items":    "Items is the list of Ingress.",
}

func (IngressList) SwaggerDoc() map[string]string {
	return map_IngressList
}

var map_IngressRule = map[string]string{
	"":     "IngressRule represents the rules mapping the paths under a specified host to the related backend services. Incoming requests are first evaluated for a host match, then routed to the backend associated with the matching IngressRuleValue.",
	"host": "Host is the fully qualified domain name of a network host, as defined by RFC 3986. Note the following deviations from the \"host\" part of the URI as defined in the RFC: 1. IPs are not allowed. Currently an IngressRuleValue can only apply to the\n\t  IP in the Spec of the parent Ingress.\n2. The `:` delimiter is not respected because ports are not allowed.\n\t  Currently the port of an Ingress is implicitly :80 for http and\n\t  :443 for https.\nBoth these may change in the future. Incoming requests are matched against the host before the IngressRuleValue. If the host is unspecified, the Ingress routes all traffic based on the specified IngressRuleValue.",
}

func (IngressRule) SwaggerDoc() map[string]string {
	return map_IngressRule
}

var map_IngressRuleValue = map[string]string{
	"": "IngressRuleValue represents a rule to apply against incoming requests. If the rule is satisfied, the request is routed to the specified backend. Currently mixing different types of rules in a single Ingress is disallowed, so exactly one of the following must be set.",
}

func (IngressRuleValue) SwaggerDoc() map[string]string {
	return map_IngressRuleValue
}

var map_IngressSpec = map[string]string{
	"":        "IngressSpec describes the Ingress the user wishes to exist.",
	"class":   "Class is the name of the IngressClass cluster resource. The associated IngressClass defines which controller will implement the resource. This replaces the deprecated `kubernetes.io/ingress.class` annotation. For backwards compatibility, when that annotation is set, it must be given precedence over this field. The controller may emit a warning if the field and annotation have different values. Implementations of this API may choose to ignore Ingresses without a class specified. An IngressClass resource may be marked as default, which can be used to set a default value for this field. For more information, refer to the IngressClass documentation.",
	"backend": "A default backend capable of servicing requests that don't match any rule. At least one of 'backend' or 'rules' must be specified. This field is optional to allow the loadbalancer controller or defaulting logic to specify a global default.",
	"tls":     "TLS configuration. Currently the Ingress only supports a single TLS port, 443. If multiple members of this list specify different hosts, they will be multiplexed on the same port according to the hostname specified through the SNI TLS extension, if the ingress controller fulfilling the ingress supports SNI.",
	"rules":   "A list of host rules used to configure the Ingress. If unspecified, or no rule matches, all traffic is sent to the default backend.",
}

func (IngressSpec) SwaggerDoc() map[string]string {
	return map_IngressSpec
}

var map_IngressStatus = map[string]string{
	"":             "IngressStatus describe the current state of the Ingress.",
	"loadBalancer": "LoadBalancer contains the current status of the load-balancer.",
}

func (IngressStatus) SwaggerDoc() map[string]string {
	return map_IngressStatus
}

var map_IngressTLS = map[string]string{
	"":           "IngressTLS describes the transport layer security associated with an Ingress.",
	"hosts":      "Hosts are a list of hosts included in the TLS certificate. The values in this list must match the name/s used in the tlsSecret. Defaults to the wildcard host setting for the loadbalancer controller fulfilling this Ingress, if left unspecified.",
	"secretName": "SecretName is the name of the secret used to terminate SSL traffic on 443. Field is left optional to allow SSL routing based on SNI hostname alone. If the SNI host in a listener conflicts with the \"Host\" header field used by an IngressRule, the SNI host is used for termination and value of the Host header is used for routing.",
}

func (IngressTLS) SwaggerDoc() map[string]string {
	return map_IngressTLS
}

// AUTO-GENERATED FUNCTIONS END HERE
