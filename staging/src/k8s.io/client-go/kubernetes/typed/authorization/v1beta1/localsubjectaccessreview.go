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

package v1beta1

import (
	context "context"

	authorizationv1beta1 "k8s.io/api/authorization/v1beta1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gentype "k8s.io/client-go/gentype"
	scheme "k8s.io/client-go/kubernetes/scheme"
)

// LocalSubjectAccessReviewsGetter has a method to return a LocalSubjectAccessReviewInterface.
// A group's client should implement this interface.
type LocalSubjectAccessReviewsGetter interface {
	LocalSubjectAccessReviews(namespace string) LocalSubjectAccessReviewInterface
}

// LocalSubjectAccessReviewInterface has methods to work with LocalSubjectAccessReview resources.
type LocalSubjectAccessReviewInterface interface {
	Create(ctx context.Context, localSubjectAccessReview *authorizationv1beta1.LocalSubjectAccessReview, opts v1.CreateOptions) (*authorizationv1beta1.LocalSubjectAccessReview, error)
	LocalSubjectAccessReviewExpansion
}

// localSubjectAccessReviews implements LocalSubjectAccessReviewInterface
type localSubjectAccessReviews struct {
	*gentype.Client[*authorizationv1beta1.LocalSubjectAccessReview]
}

// newLocalSubjectAccessReviews returns a LocalSubjectAccessReviews
func newLocalSubjectAccessReviews(c *AuthorizationV1beta1Client, namespace string) *localSubjectAccessReviews {
	return &localSubjectAccessReviews{
		gentype.NewClient[*authorizationv1beta1.LocalSubjectAccessReview](
			"localsubjectaccessreviews",
			c.RESTClient(),
			scheme.ParameterCodec,
			namespace,
			func() *authorizationv1beta1.LocalSubjectAccessReview {
				return &authorizationv1beta1.LocalSubjectAccessReview{}
			}),
	}
}
