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
	context "context"

	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	applyconfigurationsadmissionregistrationv1 "k8s.io/client-go/applyconfigurations/admissionregistration/v1"
	gentype "k8s.io/client-go/gentype"
	scheme "k8s.io/client-go/kubernetes/scheme"
)

// ValidatingWebhookConfigurationsGetter has a method to return a ValidatingWebhookConfigurationInterface.
// A group's client should implement this interface.
type ValidatingWebhookConfigurationsGetter interface {
	ValidatingWebhookConfigurations() ValidatingWebhookConfigurationInterface
}

// ValidatingWebhookConfigurationInterface has methods to work with ValidatingWebhookConfiguration resources.
type ValidatingWebhookConfigurationInterface interface {
	Create(ctx context.Context, validatingWebhookConfiguration *admissionregistrationv1.ValidatingWebhookConfiguration, opts metav1.CreateOptions) (*admissionregistrationv1.ValidatingWebhookConfiguration, error)
	Update(ctx context.Context, validatingWebhookConfiguration *admissionregistrationv1.ValidatingWebhookConfiguration, opts metav1.UpdateOptions) (*admissionregistrationv1.ValidatingWebhookConfiguration, error)
	Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error
	Get(ctx context.Context, name string, opts metav1.GetOptions) (*admissionregistrationv1.ValidatingWebhookConfiguration, error)
	List(ctx context.Context, opts metav1.ListOptions) (*admissionregistrationv1.ValidatingWebhookConfigurationList, error)
	Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (result *admissionregistrationv1.ValidatingWebhookConfiguration, err error)
	Apply(ctx context.Context, validatingWebhookConfiguration *applyconfigurationsadmissionregistrationv1.ValidatingWebhookConfigurationApplyConfiguration, opts metav1.ApplyOptions) (result *admissionregistrationv1.ValidatingWebhookConfiguration, err error)
	ValidatingWebhookConfigurationExpansion
}

// validatingWebhookConfigurations implements ValidatingWebhookConfigurationInterface
type validatingWebhookConfigurations struct {
	*gentype.ClientWithListAndApply[*admissionregistrationv1.ValidatingWebhookConfiguration, *admissionregistrationv1.ValidatingWebhookConfigurationList, *applyconfigurationsadmissionregistrationv1.ValidatingWebhookConfigurationApplyConfiguration]
}

// newValidatingWebhookConfigurations returns a ValidatingWebhookConfigurations
func newValidatingWebhookConfigurations(c *AdmissionregistrationV1Client) *validatingWebhookConfigurations {
	return &validatingWebhookConfigurations{
		gentype.NewClientWithListAndApply[*admissionregistrationv1.ValidatingWebhookConfiguration, *admissionregistrationv1.ValidatingWebhookConfigurationList, *applyconfigurationsadmissionregistrationv1.ValidatingWebhookConfigurationApplyConfiguration](
			"validatingwebhookconfigurations",
			c.RESTClient(),
			scheme.ParameterCodec,
			"",
			func() *admissionregistrationv1.ValidatingWebhookConfiguration {
				return &admissionregistrationv1.ValidatingWebhookConfiguration{}
			},
			func() *admissionregistrationv1.ValidatingWebhookConfigurationList {
				return &admissionregistrationv1.ValidatingWebhookConfigurationList{}
			}),
	}
}
