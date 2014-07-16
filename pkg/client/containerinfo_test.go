/*
Copyright 2014 Google Inc. All rights reserved.

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

package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/google/cadvisor/info"
	itest "github.com/google/cadvisor/info/test"
)

func testHTTPContainerInfoGetter(
	req *info.ContainerInfoRequest,
	cinfo *info.ContainerInfo,
	podID string,
	containerID string,
	status int,
	t *testing.T,
) {
	expectedPath := "/stats"
	if len(podID) > 0 && len(containerID) > 0 {
		expectedPath = path.Join(expectedPath, podID, containerID)
	}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if status != 0 {
			w.WriteHeader(status)
		}
		if strings.TrimRight(r.URL.Path, "/") != strings.TrimRight(expectedPath, "/") {
			t.Fatalf("Received request to an invalid path. Should be %v. got %v",
				expectedPath, r.URL.Path)
		}

		decoder := json.NewDecoder(r.Body)
		var receivedReq info.ContainerInfoRequest
		err := decoder.Decode(&receivedReq)
		if err != nil {
			t.Fatal(err)
		}
		// Note: This will not make a deep copy of req.
		// So changing req after Get*Info would be a race.
		expectedReq := req
		// Fill any empty fields with default value
		expectedReq = expectedReq.FillDefaults()
		if !reflect.DeepEqual(expectedReq, &receivedReq) {
			t.Errorf("received wrong request")
		}
		encoder := json.NewEncoder(w)
		err = encoder.Encode(cinfo)
		if err != nil {
			t.Fatal(err)
		}
	}))
	hostURL, err := url.Parse(ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	parts := strings.Split(hostURL.Host, ":")

	port, err := strconv.Atoi(parts[1])
	if err != nil {
		t.Fatal(err)
	}

	containerInfoGetter := &HTTPContainerInfoGetter{
		Client: http.DefaultClient,
		Port:   port,
	}

	var receivedContainerInfo *info.ContainerInfo
	if len(podID) > 0 && len(containerID) > 0 {
		receivedContainerInfo, err = containerInfoGetter.GetContainerInfo(parts[0], podID, containerID, req)
	} else {
		receivedContainerInfo, err = containerInfoGetter.GetMachineInfo(parts[0], req)
	}
	if status == 0 || status == http.StatusOK {
		if err != nil {
			t.Errorf("received unexpected error: %v", err)
		}

		if !receivedContainerInfo.Eq(cinfo) {
			t.Error("received unexpected container info")
		}
	} else {
		if err == nil {
			t.Error("did not receive expected error.")
		}
	}
}

func TestHTTPContainerInfoGetterGetContainerInfoSuccessfully(t *testing.T) {
	req := &info.ContainerInfoRequest{
		NumStats:   10,
		NumSamples: 10,
	}
	req = req.FillDefaults()
	cinfo := itest.GenerateRandomContainerInfo(
		"dockerIDWhichWillNotBeChecked", // docker ID
		2, // Number of cores
		req,
		1*time.Second,
	)
	testHTTPContainerInfoGetter(req, cinfo, "somePodID", "containerNameInK8S", 0, t)
}

func TestHTTPContainerInfoGetterGetMachineInfoSuccessfully(t *testing.T) {
	req := &info.ContainerInfoRequest{
		NumStats:   10,
		NumSamples: 10,
	}
	req = req.FillDefaults()
	cinfo := itest.GenerateRandomContainerInfo(
		"dockerIDWhichWillNotBeChecked", // docker ID
		2, // Number of cores
		req,
		1*time.Second,
	)
	testHTTPContainerInfoGetter(req, cinfo, "", "", 0, t)
}

func TestHTTPContainerInfoGetterGetContainerInfoWithError(t *testing.T) {
	req := &info.ContainerInfoRequest{
		NumStats:   10,
		NumSamples: 10,
	}
	req = req.FillDefaults()
	cinfo := itest.GenerateRandomContainerInfo(
		"dockerIDWhichWillNotBeChecked", // docker ID
		2, // Number of cores
		req,
		1*time.Second,
	)
	testHTTPContainerInfoGetter(req, cinfo, "somePodID", "containerNameInK8S", http.StatusNotFound, t)
}

func TestHTTPContainerInfoGetterGetMachineInfoWithError(t *testing.T) {
	req := &info.ContainerInfoRequest{
		NumStats:   10,
		NumSamples: 10,
	}
	req = req.FillDefaults()
	cinfo := itest.GenerateRandomContainerInfo(
		"dockerIDWhichWillNotBeChecked", // docker ID
		2, // Number of cores
		req,
		1*time.Second,
	)
	testHTTPContainerInfoGetter(req, cinfo, "", "", http.StatusNotFound, t)
}

func TestHTTPGetMachineSpec(t *testing.T) {
	mspec := &info.MachineInfo{
		NumCores:       4,
		MemoryCapacity: 2048,
	}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		encoder := json.NewEncoder(w)
		err := encoder.Encode(mspec)
		if err != nil {
			t.Fatal(err)
		}
	}))
	hostURL, err := url.Parse(ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	parts := strings.Split(hostURL.Host, ":")

	port, err := strconv.Atoi(parts[1])
	if err != nil {
		t.Fatal(err)
	}

	containerInfoGetter := &HTTPContainerInfoGetter{
		Client: http.DefaultClient,
		Port:   port,
	}

	received, err := containerInfoGetter.GetMachineSpec(parts[0])
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(received, mspec) {
		t.Errorf("received wrong machine spec")
	}
}
