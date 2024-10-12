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

package v1alpha1

import (
	context "context"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	gentype "k8s.io/client-go/gentype"
	samplecontrollerv1alpha1 "k8s.io/sample-controller/pkg/apis/samplecontroller/v1alpha1"
	scheme "k8s.io/sample-controller/pkg/generated/clientset/versioned/scheme"
)

// FoosGetter has a method to return a FooInterface.
// A group's client should implement this interface.
type FoosGetter interface {
	Foos(namespace string) FooInterface
}

// FooInterface has methods to work with Foo resources.
type FooInterface interface {
	Create(ctx context.Context, foo *samplecontrollerv1alpha1.Foo, opts v1.CreateOptions) (*samplecontrollerv1alpha1.Foo, error)
	Update(ctx context.Context, foo *samplecontrollerv1alpha1.Foo, opts v1.UpdateOptions) (*samplecontrollerv1alpha1.Foo, error)
	// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
	UpdateStatus(ctx context.Context, foo *samplecontrollerv1alpha1.Foo, opts v1.UpdateOptions) (*samplecontrollerv1alpha1.Foo, error)
	Delete(ctx context.Context, name string, opts v1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error
	Get(ctx context.Context, name string, opts v1.GetOptions) (*samplecontrollerv1alpha1.Foo, error)
	List(ctx context.Context, opts v1.ListOptions) (*samplecontrollerv1alpha1.FooList, error)
	Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *samplecontrollerv1alpha1.Foo, err error)
	FooExpansion
}

// foos implements FooInterface
type foos struct {
	*gentype.ClientWithList[*samplecontrollerv1alpha1.Foo, *samplecontrollerv1alpha1.FooList]
}

// newFoos returns a Foos
func newFoos(c *SamplecontrollerV1alpha1Client, namespace string) *foos {
	return &foos{
		gentype.NewClientWithList[*samplecontrollerv1alpha1.Foo, *samplecontrollerv1alpha1.FooList](
			"foos",
			c.RESTClient(),
			scheme.ParameterCodec,
			namespace,
			func() *samplecontrollerv1alpha1.Foo { return &samplecontrollerv1alpha1.Foo{} },
			func() *samplecontrollerv1alpha1.FooList { return &samplecontrollerv1alpha1.FooList{} }),
	}
}
