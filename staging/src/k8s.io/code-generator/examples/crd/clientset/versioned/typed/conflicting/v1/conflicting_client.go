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

// Code generated by client-gen. DO NOT EDIT.

package v1

import (
	http "net/http"

	rest "k8s.io/client-go/rest"
	conflictingv1 "k8s.io/code-generator/examples/crd/apis/conflicting/v1"
	scheme "k8s.io/code-generator/examples/crd/clientset/versioned/scheme"
)

type ConflictingExampleV1Interface interface {
	RESTClient() rest.Interface
	TestTypesGetter
}

// ConflictingExampleV1Client is used to interact with features provided by the conflicting.test.crd.code-generator.k8s.io group.
type ConflictingExampleV1Client struct {
	restClient rest.Interface
}

func (c *ConflictingExampleV1Client) TestTypes(namespace string) TestTypeInterface {
	return newTestTypes(c, namespace)
}

// NewForConfig creates a new ConflictingExampleV1Client for the given config.
// NewForConfig is equivalent to NewForConfigAndClient(c, httpClient),
// where httpClient was generated with rest.HTTPClientFor(c).
func NewForConfig(c *rest.Config) (*ConflictingExampleV1Client, error) {
	config := *c
	if err := setConfigDefaults(&config); err != nil {
		return nil, err
	}
	httpClient, err := rest.HTTPClientFor(&config)
	if err != nil {
		return nil, err
	}
	return NewForConfigAndClient(&config, httpClient)
}

// NewForConfigAndClient creates a new ConflictingExampleV1Client for the given config and http client.
// Note the http client provided takes precedence over the configured transport values.
func NewForConfigAndClient(c *rest.Config, h *http.Client) (*ConflictingExampleV1Client, error) {
	config := *c
	if err := setConfigDefaults(&config); err != nil {
		return nil, err
	}
	client, err := rest.RESTClientForConfigAndClient(&config, h)
	if err != nil {
		return nil, err
	}
	return &ConflictingExampleV1Client{client}, nil
}

// NewForConfigOrDie creates a new ConflictingExampleV1Client for the given config and
// panics if there is an error in the config.
func NewForConfigOrDie(c *rest.Config) *ConflictingExampleV1Client {
	client, err := NewForConfig(c)
	if err != nil {
		panic(err)
	}
	return client
}

// New creates a new ConflictingExampleV1Client for the given RESTClient.
func New(c rest.Interface) *ConflictingExampleV1Client {
	return &ConflictingExampleV1Client{c}
}

func setConfigDefaults(config *rest.Config) error {
	gv := conflictingv1.SchemeGroupVersion
	config.GroupVersion = &gv
	config.APIPath = "/apis"
	config.NegotiatedSerializer = scheme.Codecs.WithoutConversion()

	if config.UserAgent == "" {
		config.UserAgent = rest.DefaultKubernetesUserAgent()
	}

	return nil
}

// RESTClient returns a RESTClient that is used to communicate
// with API server by this client implementation.
func (c *ConflictingExampleV1Client) RESTClient() rest.Interface {
	if c == nil {
		return nil
	}
	return c.restClient
}
