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

package cache

import (
	"reflect"
	"testing"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
)

func TestAddNode(t *testing.T) {
	tests := []struct {
		name                 string
		node                 *v1.Node
		topologyInfo         NodeTopologyInfo
		expectedTopologyInfo NodeTopologyInfo
	}{
		{
			name: "node with no labels",
			node: &v1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node",
				},
			},
			topologyInfo:         map[TopologyPair]sets.String{},
			expectedTopologyInfo: map[TopologyPair]sets.String{},
		},
		{
			name: "node with labels",
			node: &v1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "node",
					Labels: map[string]string{"k1": "v1", "k2": "v2"},
				},
			},
			topologyInfo: map[TopologyPair]sets.String{},
			expectedTopologyInfo: map[TopologyPair]sets.String{
				{"k1", "v1"}: sets.NewString("node"),
				{"k2", "v2"}: sets.NewString("node"),
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.topologyInfo.AddNode(test.node)
			if !reflect.DeepEqual(test.topologyInfo, test.expectedTopologyInfo) {
				t.Errorf("TestAddNode: Got: %+v, Expected: %+v", test.topologyInfo, test.expectedTopologyInfo)
			}
		})
	}
}

func TestRemoveNode(t *testing.T) {
	tests := []struct {
		name                 string
		node                 *v1.Node
		topologyInfo         NodeTopologyInfo
		expectedTopologyInfo NodeTopologyInfo
	}{
		{
			name: "node with labels",
			node: &v1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "node",
					Labels: map[string]string{"k1": "v1", "k2": "v2"},
				},
			},
			topologyInfo: map[TopologyPair]sets.String{
				{"k1", "v1"}: sets.NewString("node"),
				{"k2", "v2"}: sets.NewString("node"),
			},
			expectedTopologyInfo: map[TopologyPair]sets.String{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.topologyInfo.RemoveNode(test.node)
			if !reflect.DeepEqual(test.topologyInfo, test.expectedTopologyInfo) {
				t.Errorf("TestRemoveNode: Got: %+v, Expected: %+v", test.topologyInfo, test.expectedTopologyInfo)
			}
		})
	}
}

func TestUpdateNode(t *testing.T) {
	tests := []struct {
		name                 string
		oldNode              *v1.Node
		newNode              *v1.Node
		topologyInfo         NodeTopologyInfo
		expectedTopologyInfo NodeTopologyInfo
	}{
		{
			name: "old and new node with the same labels",
			oldNode: &v1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "node",
					Labels: map[string]string{"k1": "v1", "k2": "v2"},
				},
			},
			newNode: &v1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "node",
					Labels: map[string]string{"k1": "v1", "k2": "v2"},
				},
			},
			topologyInfo: map[TopologyPair]sets.String{
				{"k1", "v1"}: sets.NewString("node"),
				{"k2", "v2"}: sets.NewString("node"),
			},
			expectedTopologyInfo: map[TopologyPair]sets.String{
				{"k1", "v1"}: sets.NewString("node"),
				{"k2", "v2"}: sets.NewString("node"),
			},
		},
		{
			name: "old and new node with different labels",
			oldNode: &v1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "node",
					Labels: map[string]string{"k1": "v1", "k2": "v2", "k3": "v3"},
				},
			},
			newNode: &v1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "node",
					Labels: map[string]string{"k1": "v1", "k2": "v2new", "k4": "v4"},
				},
			},
			topologyInfo: map[TopologyPair]sets.String{
				{"k1", "v1"}: sets.NewString("node"),
				{"k2", "v2"}: sets.NewString("node"),
				{"k3", "v3"}: sets.NewString("node"),
			},
			expectedTopologyInfo: map[TopologyPair]sets.String{
				{"k1", "v1"}:    sets.NewString("node"),
				{"k2", "v2new"}: sets.NewString("node"),
				{"k4", "v4"}:    sets.NewString("node"),
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.topologyInfo.UpdateNode(test.oldNode, test.newNode)
			if !reflect.DeepEqual(test.topologyInfo, test.expectedTopologyInfo) {
				t.Errorf("TestUpdateNode: Got: %+v, Expected: %+v", test.topologyInfo, test.expectedTopologyInfo)
			}
		})
	}
}
