/*
Copyright 2019 The Kubernetes Authors.

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

package reconcilers

import (
	"fmt"
	"net"
	"testing"

	"github.com/google/go-cmp/cmp"

	corev1 "k8s.io/api/core/v1"
	discovery "k8s.io/api/discovery/v1"
	apiequality "k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/kubernetes/fake"
	utilnet "k8s.io/utils/net"
)

func TestEndpointsAdapterGet(t *testing.T) {
	endpoints1, epSlice1 := generateEndpointsAndSlice([]int{80, 443}, []string{"10.1.2.3", "10.1.2.4"})
	ips1 := sets.New("10.1.2.3", "10.1.2.4")

	endpoints1v6, epSlice1v6 := generateEndpointsAndSlice([]int{80, 443}, []string{"1234::5678:0000:0000:9abc:def0"})
	epSlice1v6.AddressType = discovery.AddressTypeIPv6
	ipsV6 := sets.New("1234::5678:0000:0000:9abc:def0")

	epSlice1AltName := epSlice1.DeepCopy()
	epSlice1AltName.Name = testServiceName + "-v4"
	epSlice1Secondary := epSlice1v6.DeepCopy()
	epSlice1Secondary.Name = testServiceName + "-v6"

	testCases := map[string]struct {
		initialState       []runtime.Object
		serviceIP          net.IP
		secondaryServiceIP net.IP

		expectedError        error
		expectedIPs          sets.Set[string]
		expectedSecondaryIPs sets.Set[string]
	}{
		"single-existing-endpoints": {
			initialState: []runtime.Object{endpoints1, epSlice1},
			serviceIP:    testServiceIP,

			expectedIPs: ips1,
		},
		"endpoints exists, endpointslice does not": {
			initialState: []runtime.Object{endpoints1},
			serviceIP:    testServiceIP,

			expectedIPs: sets.New[string](),
		},
		"endpointslice exists, endpoints does not": {
			initialState: []runtime.Object{epSlice1},
			serviceIP:    testServiceIP,

			expectedIPs: ips1,
		},
		"endpoints is ignored in favor of endpointslice": {
			initialState: []runtime.Object{endpoints1v6, epSlice1},
			serviceIP:    testServiceIP,

			expectedIPs: ips1,
		},
		"IPv6 endpoints ignored in IPv4 cluster": {
			initialState: []runtime.Object{epSlice1, epSlice1Secondary},
			serviceIP:    testServiceIP,

			expectedIPs: ips1,
		},
		"IPv4 endpoints ignored in IPv6 cluster": {
			initialState: []runtime.Object{epSlice1AltName, epSlice1v6},
			serviceIP:    testServiceIPv6,

			expectedIPs: ipsV6,
		},
		"All endpoints used in dual-stack cluster": {
			initialState:       []runtime.Object{epSlice1, epSlice1Secondary},
			serviceIP:          testServiceIP,
			secondaryServiceIP: testServiceIPv6,

			expectedIPs:          ips1,
			expectedSecondaryIPs: ipsV6,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			client := fake.NewSimpleClientset(testCase.initialState...)
			epAdapter := NewEndpointsAdapter(client.CoreV1(), client.DiscoveryV1(),
				testServiceNamespace, testServiceName,
				testCase.serviceIP, testCase.secondaryServiceIP)

			primaryIPs, secondaryIPs, err := epAdapter.Get()

			if !apiequality.Semantic.DeepEqual(testCase.expectedError, err) {
				t.Errorf("Expected error: %v, got: %v", testCase.expectedError, err)
			}

			if !apiequality.Semantic.DeepEqual(primaryIPs, testCase.expectedIPs) {
				t.Errorf("Wrong primary IPs:\n%s", cmp.Diff(testCase.expectedIPs, primaryIPs))
			}
			if !apiequality.Semantic.DeepEqual(secondaryIPs, testCase.expectedSecondaryIPs) {
				t.Errorf("Wrong secondary IPs:\n%s", cmp.Diff(testCase.expectedSecondaryIPs, secondaryIPs))
			}
		})
	}
}

func TestEndpointsAdapterSync(t *testing.T) {
	endpoints1, epSlice1 := generateEndpointsAndSlice([]int{80}, []string{"10.1.2.3", "10.1.2.4"})
	ips1 := sets.New("10.1.2.3", "10.1.2.4")
	ports1 := endpoints1.Subsets[0].Ports

	endpoints2, epSlice2 := generateEndpointsAndSlice([]int{80, 443}, []string{"10.1.2.3", "10.1.2.4", "10.1.2.5"})
	ips2 := sets.New("10.1.2.3", "10.1.2.4", "10.1.2.5")
	ports2 := endpoints2.Subsets[0].Ports

	endpointsV6, epSliceV6 := generateEndpointsAndSlice([]int{80}, []string{"1234::5678", "1234::abcd"})
	ipsV6 := sets.New("1234::5678", "1234::abcd")

	endpointsDual, _ := generateEndpointsAndSlice([]int{80}, []string{"10.1.2.3", "10.1.2.4", "1234::5678", "1234::abcd"})
	ipsDual := sets.New("10.1.2.3", "10.1.2.4", "1234::5678", "1234::abcd")

	_, epSlice1Deprecated := generateEndpointsAndSlice([]int{80}, []string{"10.1.3", "10.1.2.4"})
	epSlice1Deprecated.AddressType = discovery.AddressType("IP")

	// epSlice1 + epSlice1Secondary represent a hypothetical future dual-stack set of
	// endpoints where epSlice1 has name "kubernetes" and epSlice1Secondary has name
	// "kubernetes-v6" (but both have service-name "kubernetes").
	epSlice1Secondary := epSliceV6.DeepCopy()
	epSlice1Secondary.Name = "kubernetes-v6"

	// epSlice1AltName + epSlice1Secondary represent an alternate hypothetical future
	// dual-stack set of endpoints where both slices have service-name "kubernetes"
	// but neither has name "kubernetes".
	epSlice1AltName := epSlice1.DeepCopy()
	epSlice1AltName.Name = "kubernetes-v4"

	// For testing when some but not all apiservers are dual-stack
	ipsDualPartial := sets.New("10.1.2.3", "10.1.2.4", "1234::5678")
	epSlice1SecondaryEmpty := epSlice1Secondary.DeepCopy()
	epSlice1SecondaryEmpty.Endpoints = []discovery.Endpoint{}

	testCases := map[string]struct {
		serviceIP          net.IP
		secondaryServiceIP net.IP
		initialState       []runtime.Object
		ipsParam           sets.Set[string]
		portsParam         []corev1.EndpointPort

		expectedError error
		expectCreate  []runtime.Object
		expectUpdate  []runtime.Object
		expectDelete  []runtime.Object
	}{
		"single-endpoint": {
			// If the Endpoints/EndpointSlice do not exist, they will be
			// created.
			serviceIP:    testServiceIP,
			initialState: []runtime.Object{},
			ipsParam:     ips1,
			portsParam:   ports1,

			expectCreate: []runtime.Object{endpoints1, epSlice1},
		},
		"ipv4-cluster-no-create-ipv6": {
			// In a single-stack IPv4 cluster, if the Endpoints/EndpointSlice
			// do not exist, and the reconciler erroneously tries to create a
			// dual-stack Endpoints we will ignore the IPv6 endpoints
			serviceIP:    testServiceIP,
			initialState: []runtime.Object{},
			ipsParam:     ipsDual,
			portsParam:   ports1,

			expectCreate: []runtime.Object{endpoints1, epSlice1},
		},
		"ipv4-cluster-no-update-ipv6": {
			// In a single-stack IPv4 cluster, if the existing Endpoints
			// object contains both IPv4 and IPv6 addresses, we should
			// overwrite it without the IPv6 addresses.
			serviceIP:    testServiceIP,
			initialState: []runtime.Object{endpointsDual},
			ipsParam:     ipsDual,
			portsParam:   ports1,

			expectUpdate: []runtime.Object{endpoints1},
			expectCreate: []runtime.Object{epSlice1},
		},
		"single-endpoint-full-ipv6": {
			// In a single-stack IPv6 cluster, if the Endpoints/EndpointSlice
			// do not exist, and the reconciler creates a single-stack IPv6
			// Endpoints, we will create a single-stack IPv6 EndpointSlice.
			serviceIP:    testServiceIPv6,
			initialState: []runtime.Object{},
			ipsParam:     ipsV6,
			portsParam:   ports1,

			expectCreate: []runtime.Object{endpointsV6, epSliceV6},
		},
		"existing-endpointslice-correct": {
			// No error when we need to create the Endpoints but the correct
			// EndpointSlice already exists
			serviceIP:    testServiceIP,
			initialState: []runtime.Object{epSlice1},
			ipsParam:     ips1,
			portsParam:   ports1,

			expectCreate: []runtime.Object{endpoints1},
		},
		"existing-endpointslice-incorrect": {
			// No error when we need to create the Endpoints but an incorrect
			// EndpointSlice already exists
			serviceIP:    testServiceIP,
			initialState: []runtime.Object{epSlice1},
			ipsParam:     ips2,
			portsParam:   ports2,

			expectCreate: []runtime.Object{endpoints2},
			expectUpdate: []runtime.Object{epSlice2},
		},
		"single-existing-endpoints-no-change": {
			// If the Endpoints/EndpointSlice already exist and are correct,
			// then Sync will do nothing
			serviceIP:    testServiceIP,
			initialState: []runtime.Object{endpoints1, epSlice1},
			ipsParam:     ips1,
			portsParam:   ports1,
		},
		"existing-endpointslice-replaced-with-updated-ipv4-address-type": {
			// If an EndpointSlice with deprecated "IP" address type exists,
			// it is deleted and replaced with one that has "IPv4" address
			// type. (This case should never actually happen any more, but it
			// gets handled automatically as a side effect of handling
			// dual-stack rollbacks.)
			serviceIP:    testServiceIP,
			initialState: []runtime.Object{endpoints1, epSlice1Deprecated},
			ipsParam:     ips1,
			portsParam:   ports1,

			expectDelete: []runtime.Object{epSlice1Deprecated},
			expectCreate: []runtime.Object{epSlice1},
		},
		"add-ports-and-ips": {
			// If we add ports/IPs to the Endpoints they will be added to
			// the EndpointSlice.
			serviceIP:    testServiceIP,
			initialState: []runtime.Object{endpoints1, epSlice1},
			ipsParam:     ips2,
			portsParam:   ports2,

			expectUpdate: []runtime.Object{endpoints2, epSlice2},
		},
		"endpoints-correct-endpointslice-wrong": {
			// If the Endpoints is correct and the EndpointSlice is wrong,
			// Sync will update the EndpointSlice.
			serviceIP:    testServiceIP,
			initialState: []runtime.Object{endpoints2, epSlice1},
			ipsParam:     ips2,
			portsParam:   ports2,

			expectUpdate: []runtime.Object{epSlice2},
		},
		"endpointslice-correct-endpoints-wrong": {
			// If the EndpointSlice is correct and the Endpoints is wrong,
			// Sync will update the Endpoints.
			serviceIP:    testServiceIP,
			initialState: []runtime.Object{endpoints1, epSlice2},
			ipsParam:     ips2,
			portsParam:   ports2,

			expectUpdate: []runtime.Object{endpoints2},
		},
		"missing-endpointslice": {
			// No error when we need to update the Endpoints but the
			// EndpointSlice doesn't exist
			serviceIP:    testServiceIP,
			initialState: []runtime.Object{endpoints2},
			ipsParam:     ips1,
			portsParam:   ports1,

			expectUpdate: []runtime.Object{endpoints1},
			expectCreate: []runtime.Object{epSlice1},
		},
		"wrong-family-endpointslice": {
			// If the EndpointSlice has endpoints of the wrong family it will
			// be deleted and recreated correctly.
			serviceIP:    testServiceIP,
			initialState: []runtime.Object{endpoints1, epSliceV6},
			ipsParam:     ips1,
			portsParam:   ports1,

			expectDelete: []runtime.Object{epSliceV6},
			expectCreate: []runtime.Object{epSlice1},
		},
		"dual-stack-partial-rollout": {
			// When the first apiserver in a formerly single-stack IPv4
			// cluster switches to dual-stack, the endpoints will initially
			// stay single-stack.
			serviceIP:          testServiceIP,
			secondaryServiceIP: testServiceIPv6,
			initialState:       []runtime.Object{endpoints1, epSlice1},
			ipsParam:           ipsDualPartial,
			portsParam:         ports1,

			expectCreate: []runtime.Object{epSlice1SecondaryEmpty},
		},
		"dual-stack-finished-rollout": {
			// When the last apiserver in a formerly single-stack IPv4 cluster
			// switches to dual-stack, the endpoints become dual-stack.
			serviceIP:          testServiceIP,
			secondaryServiceIP: testServiceIPv6,
			initialState:       []runtime.Object{endpoints1, epSlice1, epSlice1SecondaryEmpty},
			ipsParam:           ipsDual,
			portsParam:         ports1,

			expectUpdate: []runtime.Object{epSlice1Secondary},
		},
		"dual-stack-partial-rollback": {
			// When the first apiserver in a formerly dual-stack cluster
			// switches to single-stack IPv4, the endpoints immediately become
			// single-stack.
			serviceIP:          testServiceIP,
			secondaryServiceIP: testServiceIPv6,
			initialState:       []runtime.Object{endpoints1, epSlice1, epSlice1Secondary},
			ipsParam:           ipsDualPartial,
			portsParam:         ports1,

			expectUpdate: []runtime.Object{epSlice1SecondaryEmpty},
		},
		"dual-stack-finished-rollback": {
			// When the last apiserver in a formerly dual-stack cluster
			// switches to single-stack IPv4, the empty IPv6 slice is deleted.
			serviceIP:    testServiceIP,
			initialState: []runtime.Object{endpoints1, epSlice1, epSlice1SecondaryEmpty},
			ipsParam:     ips1,
			portsParam:   ports1,
			expectDelete: []runtime.Object{epSlice1SecondaryEmpty},
		},
		"dual-stack-alt-rollback": {
			// If a single-stack IPv4 apiserver starts up in a cluster with
			// weirdly-named dual-stack EndpointSlices, it will delete both
			// slices and create a new IPv4 slice.
			serviceIP:    testServiceIP,
			initialState: []runtime.Object{endpoints1, epSlice1AltName, epSlice1Secondary},
			ipsParam:     ips1,
			portsParam:   ports1,

			expectDelete: []runtime.Object{epSlice1AltName, epSlice1Secondary},
			expectCreate: []runtime.Object{epSlice1},
		},
		"dual-stack-rollback-and-change": {
			// If a single-stack IPv4 apiserver starts up in a cluster with
			// dual-stack EndpointSlices where the IPv4 slice is wrong, it
			// will delete the IPv6 slice and update the IPv4 slice.
			serviceIP:    testServiceIP,
			initialState: []runtime.Object{endpoints1, epSlice1, epSlice1Secondary},
			ipsParam:     ips2,
			portsParam:   ports2,

			expectDelete: []runtime.Object{epSlice1Secondary},
			expectUpdate: []runtime.Object{endpoints2, epSlice2},
		},
		"dual-stack-alt-rollback-and-change": {
			// If a single-stack IPv4 apiserver starts up in a cluster with
			// weirdly-named dual-stack EndpointSlices where the IPv4 slice is
			// wrong, it will delete both slices and create a new IPv4 slice.
			serviceIP:    testServiceIP,
			initialState: []runtime.Object{epSlice1AltName, epSlice1Secondary},
			ipsParam:     ips2,
			portsParam:   ports2,

			expectDelete: []runtime.Object{epSlice1AltName, epSlice1Secondary},
			expectCreate: []runtime.Object{endpoints2, epSlice2},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			client := fake.NewSimpleClientset(testCase.initialState...)
			epAdapter := NewEndpointsAdapter(client.CoreV1(), client.DiscoveryV1(),
				testServiceNamespace, testServiceName,
				testCase.serviceIP, testCase.secondaryServiceIP)

			err := epAdapter.Sync(testCase.ipsParam, testCase.portsParam, true)
			if !apiequality.Semantic.DeepEqual(testCase.expectedError, err) {
				t.Errorf("Expected error: %v, got: %v", testCase.expectedError, err)
			}

			err = verifyActions(client, testCase.expectCreate, testCase.expectUpdate, testCase.expectDelete)
			if err != nil {
				t.Errorf("unexpected error in side effects: %v", err)
			}
		})
	}
}

func generateEndpointsAndSlice(ports []int, addresses []string) (*corev1.Endpoints, *discovery.EndpointSlice) {
	trueBool := true
	addressType := discovery.AddressTypeIPv4
	if len(addresses) > 0 && utilnet.IsIPv6String(addresses[0]) {
		addressType = discovery.AddressTypeIPv6
	}

	epSlice := &discovery.EndpointSlice{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: testServiceNamespace,
			Name:      testServiceName,
		},
		AddressType: addressType,
	}
	epSlice.Labels = map[string]string{discovery.LabelServiceName: testServiceName}
	subset := corev1.EndpointSubset{}

	for i, port := range ports {
		endpointPort := corev1.EndpointPort{
			Name:     fmt.Sprintf("port-%d", i),
			Port:     int32(port),
			Protocol: corev1.ProtocolTCP,
		}
		subset.Ports = append(subset.Ports, endpointPort)
		epSlice.Ports = append(epSlice.Ports, discovery.EndpointPort{
			Name:     &endpointPort.Name,
			Port:     &endpointPort.Port,
			Protocol: &endpointPort.Protocol,
		})
	}

	for _, address := range addresses {
		endpointAddress := corev1.EndpointAddress{
			IP: address,
		}

		subset.Addresses = append(subset.Addresses, endpointAddress)

		epSlice.Endpoints = append(epSlice.Endpoints, discovery.Endpoint{
			Addresses:  []string{endpointAddress.IP},
			Conditions: discovery.EndpointConditions{Ready: &trueBool},
		})
	}

	return &corev1.Endpoints{
		ObjectMeta: metav1.ObjectMeta{
			Name:      testServiceName,
			Namespace: testServiceNamespace,
			Labels: map[string]string{
				discovery.LabelSkipMirror: "true",
			},
		},
		Subsets: []corev1.EndpointSubset{subset},
	}, epSlice
}
