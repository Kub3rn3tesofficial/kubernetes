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

package api_test

import (
	"bytes"
	"fmt"
	"math/rand"
	"reflect"
	"strings"
	"testing"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/testapi"
	apitesting "k8s.io/kubernetes/pkg/api/testing"
	"k8s.io/kubernetes/pkg/api/unversioned"
	"k8s.io/kubernetes/pkg/apimachinery/registered"
	"k8s.io/kubernetes/pkg/util/diff"

	"github.com/google/gofuzz"
)

func TestDeepCopyApiObjects(t *testing.T) {
	for i := 0; i < *fuzzIters; i++ {
		for _, version := range []unversioned.GroupVersion{testapi.Default.InternalGroupVersion(), registered.GroupOrDie(api.GroupName).GroupVersion} {
			f := apitesting.FuzzerFor(t, version, rand.NewSource(rand.Int63()))
			for kind := range api.Scheme.KnownTypes(version) {
				doDeepCopyTest(t, version.WithKind(kind), f)
			}
		}
	}
}

func doDeepCopyTest(t *testing.T, kind unversioned.GroupVersionKind, f *fuzz.Fuzzer) {
	item, err := api.Scheme.New(kind)
	if err != nil {
		t.Fatalf("Could not create a %v: %s", kind, err)
	}
	f.Fuzz(item)
	itemCopy, err := api.Scheme.DeepCopy(item)
	if err != nil {
		t.Errorf("Could not deep copy a %v: %s", kind, err)
		return
	}

	if !reflect.DeepEqual(item, itemCopy) {
		t.Errorf("\nexpected: %#v\n\ngot:      %#v\n\ndiff:      %v", item, itemCopy, diff.ObjectReflectDiff(item, itemCopy))
	}

	prefuzzData := &bytes.Buffer{}
	if err := api.Codecs.LegacyCodec(kind.GroupVersion()).Encode(item, prefuzzData); err != nil {
		t.Error(err)
		return
	}

	// Refuzz the copy, which should have no effect on the original
	f.Fuzz(itemCopy)

	postfuzzData := &bytes.Buffer{}
	if err := api.Codecs.LegacyCodec(kind.GroupVersion()).Encode(item, postfuzzData); err != nil {
		t.Errorf("Could not deep copy a %v: %s", kind, err)
		return
	}

	if bytes.Compare(prefuzzData.Bytes(), postfuzzData.Bytes()) != 0 {
		t.Errorf("Fuzzing copy modified original of %#v", kind)
		return
	}

	// Most runtime.Objects are pointers
	// Make sure calling DeepCopy() with a non-pointer struct either errors, or actually deep copies
	itemValue := reflect.ValueOf(item)
	if itemValue.Kind() == reflect.Ptr {
		itemStructValue := itemValue.Elem()
		fmt.Println(itemStructValue.Type())
		itemStruct := itemStructValue.Interface()
		itemStructCopy, err := api.Scheme.DeepCopy(itemStruct)
		if err == nil || strings.Contains(err.Error(), "must pass addressable value") {
			return
		}
		if err != nil {
			t.Errorf("Could not deep copy struct of %v: %s", kind, err)
			return
		}
		if !reflect.DeepEqual(itemStruct, itemStructCopy) {
			t.Errorf("\nexpected: %#v\n\ngot:      %#v\n\ndiff:      %v", itemStruct, itemStructCopy, diff.ObjectReflectDiff(itemStruct, itemStructCopy))
			return
		}
		// Refuzz the copy, which should have no effect on the original
		f.Fuzz(itemStructCopy)
		if reflect.DeepEqual(itemStruct, itemStructCopy) {
			t.Errorf("Fuzzing struct copy modified original of %#v", kind)
			return
		}
	}
}

func TestDeepCopySingleType(t *testing.T) {
	for i := 0; i < *fuzzIters; i++ {
		for _, version := range []unversioned.GroupVersion{testapi.Default.InternalGroupVersion(), registered.GroupOrDie(api.GroupName).GroupVersion} {
			f := apitesting.FuzzerFor(t, version, rand.NewSource(rand.Int63()))
			doDeepCopyTest(t, version.WithKind("Pod"), f)
		}
	}
}
