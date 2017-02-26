/*
Copyright 2015 The Kubernetes Authors.

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

package deployment

import (
	"fmt"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/uuid"
	core "k8s.io/client-go/testing"
	"k8s.io/client-go/tools/record"
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/v1"
	extensions "k8s.io/kubernetes/pkg/apis/extensions/v1beta1"
	"k8s.io/kubernetes/pkg/client/clientset_generated/clientset/fake"
	informers "k8s.io/kubernetes/pkg/client/informers/informers_generated/externalversions"
	"k8s.io/kubernetes/pkg/controller"
)

var (
	alwaysReady = func() bool { return true }
	noTimestamp = metav1.Time{}
)

func rs(name string, replicas int, selector map[string]string, timestamp metav1.Time) *extensions.ReplicaSet {
	return &extensions.ReplicaSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:              name,
			CreationTimestamp: timestamp,
			Namespace:         metav1.NamespaceDefault,
		},
		Spec: extensions.ReplicaSetSpec{
			Replicas: func() *int32 { i := int32(replicas); return &i }(),
			Selector: &metav1.LabelSelector{MatchLabels: selector},
			Template: v1.PodTemplateSpec{},
		},
	}
}

func newRSWithStatus(name string, specReplicas, statusReplicas int, selector map[string]string) *extensions.ReplicaSet {
	rs := rs(name, specReplicas, selector, noTimestamp)
	rs.Status = extensions.ReplicaSetStatus{
		Replicas: int32(statusReplicas),
	}
	return rs
}

func newDeployment(name string, replicas int, revisionHistoryLimit *int32, maxSurge, maxUnavailable *intstr.IntOrString, selector map[string]string) *extensions.Deployment {
	d := extensions.Deployment{
		TypeMeta: metav1.TypeMeta{APIVersion: api.Registry.GroupOrDie(extensions.GroupName).GroupVersion.String()},
		ObjectMeta: metav1.ObjectMeta{
			UID:         uuid.NewUUID(),
			Name:        name,
			Namespace:   metav1.NamespaceDefault,
			Annotations: make(map[string]string),
		},
		Spec: extensions.DeploymentSpec{
			Strategy: extensions.DeploymentStrategy{
				Type: extensions.RollingUpdateDeploymentStrategyType,
				RollingUpdate: &extensions.RollingUpdateDeployment{
					MaxUnavailable: func() *intstr.IntOrString { i := intstr.FromInt(0); return &i }(),
					MaxSurge:       func() *intstr.IntOrString { i := intstr.FromInt(0); return &i }(),
				},
			},
			Replicas: func() *int32 { i := int32(replicas); return &i }(),
			Selector: &metav1.LabelSelector{MatchLabels: selector},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: selector,
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Image: "foo/bar",
						},
					},
				},
			},
			RevisionHistoryLimit: revisionHistoryLimit,
		},
	}
	if maxSurge != nil {
		d.Spec.Strategy.RollingUpdate.MaxSurge = maxSurge
	}
	if maxUnavailable != nil {
		d.Spec.Strategy.RollingUpdate.MaxUnavailable = maxUnavailable
	}
	return &d
}

func newReplicaSet(d *extensions.Deployment, name string, replicas int) *extensions.ReplicaSet {
	return &extensions.ReplicaSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:            name,
			UID:             uuid.NewUUID(),
			Namespace:       metav1.NamespaceDefault,
			Labels:          d.Spec.Selector.MatchLabels,
			OwnerReferences: []metav1.OwnerReference{*newControllerRef(d)},
		},
		Spec: extensions.ReplicaSetSpec{
			Selector: d.Spec.Selector,
			Replicas: func() *int32 { i := int32(replicas); return &i }(),
			Template: d.Spec.Template,
		},
	}
}

func getKey(d *extensions.Deployment, t *testing.T) string {
	if key, err := controller.KeyFunc(d); err != nil {
		t.Errorf("Unexpected error getting key for deployment %v: %v", d.Name, err)
		return ""
	} else {
		return key
	}
}

type fixture struct {
	t *testing.T

	client *fake.Clientset
	// Objects to put in the store.
	dLister   []*extensions.Deployment
	rsLister  []*extensions.ReplicaSet
	podLister []*v1.Pod

	// Actions expected to happen on the client. Objects from here are also
	// preloaded into NewSimpleFake.
	actions []core.Action
	objects []runtime.Object
}

func (f *fixture) expectUpdateDeploymentStatusAction(d *extensions.Deployment) {
	action := core.NewUpdateAction(schema.GroupVersionResource{Resource: "deployments"}, d.Namespace, d)
	action.Subresource = "status"
	f.actions = append(f.actions, action)
}

func (f *fixture) expectCreateRSAction(rs *extensions.ReplicaSet) {
	f.actions = append(f.actions, core.NewCreateAction(schema.GroupVersionResource{Resource: "replicasets"}, rs.Namespace, rs))
}

func newFixture(t *testing.T) *fixture {
	f := &fixture{}
	f.t = t
	f.objects = []runtime.Object{}
	return f
}

func (f *fixture) newController() (*DeploymentController, informers.SharedInformerFactory) {
	f.client = fake.NewSimpleClientset(f.objects...)
	informers := informers.NewSharedInformerFactory(f.client, controller.NoResyncPeriodFunc())
	c := NewDeploymentController(informers.Extensions().V1beta1().Deployments(), informers.Extensions().V1beta1().ReplicaSets(), informers.Core().V1().Pods(), f.client)
	c.eventRecorder = &record.FakeRecorder{}
	c.dListerSynced = alwaysReady
	c.rsListerSynced = alwaysReady
	c.podListerSynced = alwaysReady
	for _, d := range f.dLister {
		informers.Extensions().V1beta1().Deployments().Informer().GetIndexer().Add(d)
	}
	for _, rs := range f.rsLister {
		informers.Extensions().V1beta1().ReplicaSets().Informer().GetIndexer().Add(rs)
	}
	for _, pod := range f.podLister {
		informers.Core().V1().Pods().Informer().GetIndexer().Add(pod)
	}
	return c, informers
}

func (f *fixture) run(deploymentName string) {
	c, informers := f.newController()
	stopCh := make(chan struct{})
	defer close(stopCh)
	informers.Start(stopCh)

	err := c.syncDeployment(deploymentName)
	if err != nil {
		f.t.Errorf("error syncing deployment: %v", err)
	}

	actions := filterInformerActions(f.client.Actions())
	for i, action := range actions {
		if len(f.actions) < i+1 {
			f.t.Errorf("%d unexpected actions: %+v", len(actions)-len(f.actions), actions[i:])
			break
		}

		expectedAction := f.actions[i]
		if !expectedAction.Matches(action.GetVerb(), action.GetResource().Resource) {
			f.t.Errorf("Expected\n\t%#v\ngot\n\t%#v", expectedAction, action)
			continue
		}
	}

	if len(f.actions) > len(actions) {
		f.t.Errorf("%d additional expected actions:%+v", len(f.actions)-len(actions), f.actions[len(actions):])
	}
}

func TestSyncDeploymentCreatesReplicaSet(t *testing.T) {
	f := newFixture(t)

	d := newDeployment("foo", 1, nil, nil, nil, map[string]string{"foo": "bar"})
	f.dLister = append(f.dLister, d)
	f.objects = append(f.objects, d)

	rs := newReplicaSet(d, "deploymentrs-4186632231", 1)

	f.expectCreateRSAction(rs)
	f.expectUpdateDeploymentStatusAction(d)
	f.expectUpdateDeploymentStatusAction(d)

	f.run(getKey(d, t))
}

func TestSyncDeploymentDontDoAnythingDuringDeletion(t *testing.T) {
	f := newFixture(t)

	d := newDeployment("foo", 1, nil, nil, nil, map[string]string{"foo": "bar"})
	now := metav1.Now()
	d.DeletionTimestamp = &now
	f.dLister = append(f.dLister, d)
	f.objects = append(f.objects, d)

	f.expectUpdateDeploymentStatusAction(d)
	f.run(getKey(d, t))
}

// issue: https://github.com/kubernetes/kubernetes/issues/23218
func TestDeploymentController_dontSyncDeploymentsWithEmptyPodSelector(t *testing.T) {
	fake := &fake.Clientset{}
	informers := informers.NewSharedInformerFactory(fake, controller.NoResyncPeriodFunc())
	controller := NewDeploymentController(informers.Extensions().V1beta1().Deployments(), informers.Extensions().V1beta1().ReplicaSets(), informers.Core().V1().Pods(), fake)
	controller.eventRecorder = &record.FakeRecorder{}
	controller.dListerSynced = alwaysReady
	controller.rsListerSynced = alwaysReady
	controller.podListerSynced = alwaysReady

	stopCh := make(chan struct{})
	defer close(stopCh)
	informers.Start(stopCh)

	d := newDeployment("foo", 1, nil, nil, nil, map[string]string{"foo": "bar"})
	empty := metav1.LabelSelector{}
	d.Spec.Selector = &empty
	informers.Extensions().V1beta1().Deployments().Informer().GetIndexer().Add(d)
	// We expect the deployment controller to not take action here since it's configuration
	// is invalid, even though no replicasets exist that match it's selector.
	controller.syncDeployment(fmt.Sprintf("%s/%s", d.ObjectMeta.Namespace, d.ObjectMeta.Name))

	filteredActions := filterInformerActions(fake.Actions())
	if len(filteredActions) == 0 {
		return
	}
	for _, action := range filteredActions {
		t.Logf("unexpected action: %#v", action)
	}
	t.Errorf("expected deployment controller to not take action")
}

func filterInformerActions(actions []core.Action) []core.Action {
	ret := []core.Action{}
	for _, action := range actions {
		if len(action.GetNamespace()) == 0 &&
			(action.Matches("list", "pods") ||
				action.Matches("list", "deployments") ||
				action.Matches("list", "replicasets") ||
				action.Matches("watch", "pods") ||
				action.Matches("watch", "deployments") ||
				action.Matches("watch", "replicasets")) {
			continue
		}
		ret = append(ret, action)
	}

	return ret
}

// TestPodDeletionEnqueuesRecreateDeployment ensures that the deletion of a pod
// will requeue a Recreate deployment iff there is no other pod returned from the
// client.
func TestPodDeletionEnqueuesRecreateDeployment(t *testing.T) {
	f := newFixture(t)

	foo := newDeployment("foo", 1, nil, nil, nil, map[string]string{"foo": "bar"})
	foo.Spec.Strategy.Type = extensions.RecreateDeploymentStrategyType
	rs := newReplicaSet(foo, "foo-1", 1)
	pod := generatePodFromRS(rs)

	f.dLister = append(f.dLister, foo)
	f.rsLister = append(f.rsLister, rs)
	f.objects = append(f.objects, foo, rs)

	c, _ := f.newController()
	enqueued := false
	c.enqueueDeployment = func(d *extensions.Deployment) {
		if d.Name == "foo" {
			enqueued = true
		}
	}

	c.deletePod(pod)

	if !enqueued {
		t.Errorf("expected deployment %q to be queued after pod deletion", foo.Name)
	}
}

// TestPodDeletionDoesntEnqueueRecreateDeployment ensures that the deletion of a pod
// will not requeue a Recreate deployment iff there are other pods returned from the
// client.
func TestPodDeletionDoesntEnqueueRecreateDeployment(t *testing.T) {
	f := newFixture(t)

	foo := newDeployment("foo", 1, nil, nil, nil, map[string]string{"foo": "bar"})
	foo.Spec.Strategy.Type = extensions.RecreateDeploymentStrategyType
	rs := newReplicaSet(foo, "foo-1", 1)
	pod := generatePodFromRS(rs)

	f.dLister = append(f.dLister, foo)
	f.rsLister = append(f.rsLister, rs)
	// Let's pretend this is a different pod. The gist is that the pod lister needs to
	// return a non-empty list.
	f.podLister = append(f.podLister, pod)

	c, _ := f.newController()
	enqueued := false
	c.enqueueDeployment = func(d *extensions.Deployment) {
		if d.Name == "foo" {
			enqueued = true
		}
	}

	c.deletePod(pod)

	if enqueued {
		t.Errorf("expected deployment %q not to be queued after pod deletion", foo.Name)
	}
}

func TestGetReplicaSetsForDeployment(t *testing.T) {
	f := newFixture(t)

	// Two Deployments with same labels.
	d1 := newDeployment("foo", 1, nil, nil, nil, map[string]string{"foo": "bar"})
	d2 := newDeployment("bar", 1, nil, nil, nil, map[string]string{"foo": "bar"})

	// Two ReplicaSets that match labels for both Deployments,
	// but have ControllerRefs to make ownership explicit.
	rs1 := newReplicaSet(d1, "rs1", 1)
	rs2 := newReplicaSet(d2, "rs2", 1)

	f.dLister = append(f.dLister, d1, d2)
	f.rsLister = append(f.rsLister, rs1, rs2)
	f.objects = append(f.objects, d1, d2, rs1, rs2)

	// Start the fixture.
	c, informers := f.newController()
	stopCh := make(chan struct{})
	defer close(stopCh)
	informers.Start(stopCh)

	rsList, err := c.getReplicaSetsForDeployment(d1)
	if err != nil {
		t.Fatalf("getReplicaSetsForDeployment() error: %v", err)
	}
	rsNames := []string{}
	for _, rs := range rsList {
		rsNames = append(rsNames, rs.Name)
	}
	if len(rsNames) != 1 || rsNames[0] != rs1.Name {
		t.Errorf("getReplicaSetsForDeployment() = %v, want [%v]", rsNames, rs1.Name)
	}

	rsList, err = c.getReplicaSetsForDeployment(d2)
	if err != nil {
		t.Fatalf("getReplicaSetsForDeployment() error: %v", err)
	}
	rsNames = []string{}
	for _, rs := range rsList {
		rsNames = append(rsNames, rs.Name)
	}
	if len(rsNames) != 1 || rsNames[0] != rs2.Name {
		t.Errorf("getReplicaSetsForDeployment() = %v, want [%v]", rsNames, rs2.Name)
	}
}

func TestGetReplicaSetsForDeploymentAdopt(t *testing.T) {
	f := newFixture(t)

	d := newDeployment("foo", 1, nil, nil, nil, map[string]string{"foo": "bar"})

	// RS with matching labels, but orphaned. Should be adopted and returned.
	rs := newReplicaSet(d, "rs", 1)
	rs.OwnerReferences = nil

	f.dLister = append(f.dLister, d)
	f.rsLister = append(f.rsLister, rs)
	f.objects = append(f.objects, d, rs)

	// Start the fixture.
	c, informers := f.newController()
	stopCh := make(chan struct{})
	defer close(stopCh)
	informers.Start(stopCh)

	rsList, err := c.getReplicaSetsForDeployment(d)
	if err != nil {
		t.Fatalf("getReplicaSetsForDeployment() error: %v", err)
	}
	rsNames := []string{}
	for _, rs := range rsList {
		rsNames = append(rsNames, rs.Name)
	}
	if len(rsNames) != 1 || rsNames[0] != rs.Name {
		t.Errorf("getReplicaSetsForDeployment() = %v, want [%v]", rsNames, rs.Name)
	}
}

func TestGetReplicaSetsForDeploymentRelease(t *testing.T) {
	f := newFixture(t)

	d := newDeployment("foo", 1, nil, nil, nil, map[string]string{"foo": "bar"})

	// RS with matching ControllerRef, but wrong labels. Should be released.
	rs := newReplicaSet(d, "rs", 1)
	rs.Labels = map[string]string{"foo": "notbar"}

	f.dLister = append(f.dLister, d)
	f.rsLister = append(f.rsLister, rs)
	f.objects = append(f.objects, d, rs)

	// Start the fixture.
	c, informers := f.newController()
	stopCh := make(chan struct{})
	defer close(stopCh)
	informers.Start(stopCh)

	rsList, err := c.getReplicaSetsForDeployment(d)
	if err != nil {
		t.Fatalf("getReplicaSetsForDeployment() error: %v", err)
	}
	rsNames := []string{}
	for _, rs := range rsList {
		rsNames = append(rsNames, rs.Name)
	}
	if len(rsNames) != 0 {
		t.Errorf("getReplicaSetsForDeployment() = %v, want []", rsNames)
	}
}

func TestGetPodMapForReplicaSets(t *testing.T) {
	f := newFixture(t)

	d := newDeployment("foo", 1, nil, nil, nil, map[string]string{"foo": "bar"})

	// Two ReplicaSets that match labels for both Deployments,
	// but have ControllerRefs to make ownership explicit.
	rs1 := newReplicaSet(d, "rs1", 1)
	rs2 := newReplicaSet(d, "rs2", 1)

	// Add a Pod for each ReplicaSet.
	pod1 := generatePodFromRS(rs1)
	pod2 := generatePodFromRS(rs2)
	// Add a Pod that has matching labels, but no ControllerRef.
	pod3 := generatePodFromRS(rs1)
	pod3.Name = "pod3"
	pod3.OwnerReferences = nil
	// Add a Pod that has matching labels and ControllerRef, but is inactive.
	pod4 := generatePodFromRS(rs1)
	pod4.Name = "pod4"
	pod4.Status.Phase = v1.PodFailed

	f.dLister = append(f.dLister, d)
	f.rsLister = append(f.rsLister, rs1, rs2)
	f.podLister = append(f.podLister, pod1, pod2, pod3, pod4)
	f.objects = append(f.objects, d, rs1, rs2, pod1, pod2, pod3, pod4)

	// Start the fixture.
	c, informers := f.newController()
	stopCh := make(chan struct{})
	defer close(stopCh)
	informers.Start(stopCh)

	podMap, err := c.getPodMapForReplicaSets(d.Namespace, f.rsLister)
	if err != nil {
		t.Fatalf("getPodMapForReplicaSets() error: %v", err)
	}
	podCount := 0
	for _, podList := range podMap {
		podCount += len(podList.Items)
	}
	if got, want := podCount, 2; got != want {
		t.Errorf("podCount = %v, want %v", got, want)
	}

	if got, want := len(podMap), 2; got != want {
		t.Errorf("len(podMap) = %v, want %v", got, want)
	}
	if got, want := len(podMap[rs1.UID].Items), 1; got != want {
		t.Errorf("len(podMap[rs1]) = %v, want %v", got, want)
	}
	if got, want := podMap[rs1.UID].Items[0].Name, "rs1-pod"; got != want {
		t.Errorf("podMap[rs1] = [%v], want [%v]", got, want)
	}
	if got, want := len(podMap[rs2.UID].Items), 1; got != want {
		t.Errorf("len(podMap[rs2]) = %v, want %v", got, want)
	}
	if got, want := podMap[rs2.UID].Items[0].Name, "rs2-pod"; got != want {
		t.Errorf("podMap[rs2] = [%v], want [%v]", got, want)
	}
}

// generatePodFromRS creates a pod, with the input ReplicaSet's selector and its template
func generatePodFromRS(rs *extensions.ReplicaSet) *v1.Pod {
	trueVar := true
	return &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      rs.Name + "-pod",
			Namespace: rs.Namespace,
			Labels:    rs.Spec.Selector.MatchLabels,
			OwnerReferences: []metav1.OwnerReference{
				{UID: rs.UID, APIVersion: "v1beta1", Kind: "ReplicaSet", Name: rs.Name, Controller: &trueVar},
			},
		},
		Spec: rs.Spec.Template.Spec,
	}
}
