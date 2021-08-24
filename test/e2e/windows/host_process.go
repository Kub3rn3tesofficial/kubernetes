/*
Copyright 2021 The Kubernetes Authors.

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

package windows

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	e2epod "k8s.io/kubernetes/test/e2e/framework/pod"

	"github.com/onsi/ginkgo"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/uuid"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/kubernetes/pkg/features"
	"k8s.io/kubernetes/test/e2e/framework"
	e2eskipper "k8s.io/kubernetes/test/e2e/framework/skipper"
	imageutils "k8s.io/kubernetes/test/utils/image"
)

var _ = SIGDescribe("[Feature:WindowsHostProcessContainers] [Excluded:WindowsDocker] [MinimumKubeletVersion:1.22] HostProcess containers", func() {
	ginkgo.BeforeEach(func() {
		e2eskipper.SkipUnlessNodeOSDistroIs("windows")
		SkipUnlessWindowsHostProcessContainersEnabled()
	})

	f := framework.NewDefaultFramework("host-process-test-windows")

	ginkgo.It("should run as a process on the host/node", func() {

		ginkgo.By("selecting a Windows node")
		targetNode, err := findWindowsNode(f)
		framework.ExpectNoError(err, "Error finding Windows node")
		framework.Logf("Using node: %v", targetNode.Name)

		ginkgo.By("scheduling a pod with a container that verifies %COMPUTERNAME% matches selected node name")
		image := imageutils.GetConfig(imageutils.BusyBox)

		trueVar := true
		podName := "host-process-test-pod"
		user := "NT AUTHORITY\\Local service"
		pod := &v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name: podName,
			},
			Spec: v1.PodSpec{
				SecurityContext: &v1.PodSecurityContext{
					WindowsOptions: &v1.WindowsSecurityContextOptions{
						HostProcess:   &trueVar,
						RunAsUserName: &user,
					},
				},
				HostNetwork: true,
				Containers: []v1.Container{
					{
						Image:   image.GetE2EImage(),
						Name:    "computer-name-test",
						Command: []string{"cmd.exe", "/K", "IF", "NOT", "%COMPUTERNAME%", "==", targetNode.Name, "(", "exit", "-1", ")"},
					},
				},
				RestartPolicy: v1.RestartPolicyNever,
				NodeName:      targetNode.Name,
			},
		}

		f.PodClient().Create(pod)

		ginkgo.By("Waiting for pod to run")
		f.PodClient().WaitForFinish(podName, 3*time.Minute)

		ginkgo.By("Then ensuring pod finished running successfully")
		p, err := f.ClientSet.CoreV1().Pods(f.Namespace.Name).Get(
			context.TODO(),
			podName,
			metav1.GetOptions{})

		framework.ExpectNoError(err, "Error retrieving pod")
		framework.ExpectEqual(p.Status.Phase, v1.PodSucceeded)
	})

	ginkgo.It("should support init containers", func() {
		ginkgo.By("scheduling a pod with a container that verifies init container can configure the node")
		trueVar := true
		podName := "host-process-init-pods"
		user := "NT AUTHORITY\\SYSTEM"
		filename := fmt.Sprintf("/testfile%s.txt", string(uuid.NewUUID()))
		pod := &v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name: podName,
			},
			Spec: v1.PodSpec{
				SecurityContext: &v1.PodSecurityContext{
					WindowsOptions: &v1.WindowsSecurityContextOptions{
						HostProcess:   &trueVar,
						RunAsUserName: &user,
					},
				},
				HostNetwork: true,
				InitContainers: []v1.Container{
					{
						Image:   imageutils.GetE2EImage(imageutils.BusyBox),
						Name:    "configure-node",
						Command: []string{"powershell", "-c", "Set-content", "-Path", filename, "-V", "test"},
					},
				},
				Containers: []v1.Container{
					{
						Image:   imageutils.GetE2EImage(imageutils.BusyBox),
						Name:    "read-configuration",
						Command: []string{"powershell", "-c", "ls", filename},
					},
				},
				RestartPolicy: v1.RestartPolicyNever,
				NodeSelector: map[string]string{
					"kubernetes.io/os": "windows",
				},
			},
		}

		f.PodClient().Create(pod)

		ginkgo.By("Waiting for pod to run")
		f.PodClient().WaitForFinish(podName, 3*time.Minute)

		ginkgo.By("Then ensuring pod finished running successfully")
		p, err := f.ClientSet.CoreV1().Pods(f.Namespace.Name).Get(
			context.TODO(),
			podName,
			metav1.GetOptions{})

		framework.ExpectNoError(err, "Error retrieving pod")

		if p.Status.Phase != v1.PodSucceeded {
			logs, err := e2epod.GetPodLogs(f.ClientSet, f.Namespace.Name, podName, "read-configuration")
			if err != nil {
				framework.Logf("Error pulling logs: %v", err)
			}
			framework.Logf("Pod phase: %v\nlogs:\n%s", p.Status.Phase, logs)
		}
		framework.ExpectEqual(p.Status.Phase, v1.PodSucceeded)
	})

	ginkgo.It("should be able to access the api service with token", func() {
		trueVar := true
		podName := "host-process-api-server-test-pod"
		containerName := "inclusterclient"
		user := "NT AUTHORITY\\Local service"
		pod := &v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name: podName,
			},
			Spec: v1.PodSpec{
				SecurityContext: &v1.PodSecurityContext{
					WindowsOptions: &v1.WindowsSecurityContextOptions{
						HostProcess:   &trueVar,
						RunAsUserName: &user,
					},
				},
				HostNetwork: true,
				Containers: []v1.Container{
					{
						Name:  containerName,
						Image: imageutils.GetE2EImage(imageutils.Agnhost),
						Args:  []string{"inclusterclient --poll-interval 5"},
						Command: []string{
							"powershell",
							"-c",
							"start-process",
							"-wait",
							"-nonewwindow",
							"$env:CONTAINER_SANDBOX_MOUNT_POINT\\agnhost",
						},
					},
				},
				RestartPolicy: v1.RestartPolicyNever,
				NodeSelector: map[string]string{
					"kubernetes.io/os": "windows",
				},
			},
		}

		ginkgo.By("Waiting for pod to run")
		f.PodClient().Create(pod)
		if !e2epod.CheckPodsRunningReady(f.ClientSet, f.Namespace.Name, []string{pod.Name}, time.Minute) {
			framework.Failf("pod %q in ns %q never became ready", pod.Name, f.Namespace.Name)
		}

		framework.Logf("pod is ready")

		ginkgo.By("ensure the pod can connect to the API server using the internal ")
		var logs string
		if err := wait.Poll(5*time.Second, 1*time.Minute, func() (done bool, err error) {
			framework.Logf("polling logs")
			logs, err = e2epod.GetPodLogs(f.ClientSet, f.Namespace.Name, podName, containerName)
			if err != nil {
				framework.Logf("Error pulling logs: %v", err)
				return false, nil
			}

			tokenCount, err := ParseInClusterClientLogs(logs)
			if err != nil {
				return false, fmt.Errorf("inclusterclient reported an error: %v", err)
			}
			if tokenCount < 2 {
				framework.Logf("Retrying. Still waiting to see more successful connections: got=%d, want=2", tokenCount)
				return false, nil
			}

			return true, nil
		}); err != nil {
			framework.Failf("Unexpected error: %v\n%s", err, logs)
		}
	})
})

func SkipUnlessWindowsHostProcessContainersEnabled() {
	if !framework.TestContext.FeatureGates[string(features.WindowsHostProcessContainers)] {
		e2eskipper.Skipf("Skipping test because feature 'WindowsHostProcessContainers' is not enabled")
	}
}

// modified from https://github.com/kubernetes/kubernetes/blob/ef754331c453d3b5fdc31edf62da3d90771d5acd/test/e2e/auth/service_accounts.go#L952
// example output from agnhost incluster client:
// 	calling /healthz
// 	authz_header=uoD86lIghNoggHPaVaKcZgtOQVOD1RW7jojjpMULgr4
var reportLogsParser = regexp.MustCompile("([a-zA-Z0-9-_]*)=([a-zA-Z0-9-_]*)$")

func ParseInClusterClientLogs(logs string) (int, error) {
	count := 0

	lines := strings.Split(logs, "\n")
	for _, line := range lines {
		if strings.Contains(line, "err") {
			return 0, fmt.Errorf("saw error in logs: %s", line)
		}
		parts := reportLogsParser.FindStringSubmatch(line)
		if len(parts) != 3 {
			continue
		}

		key, value := parts[1], parts[2]
		switch key {
		case "authz_header":
			if value == "<empty>" {
				return 0, fmt.Errorf("saw empty Authorization header")
			}
			count++
		case "status":
			if value == "failed" {
				return 0, fmt.Errorf("saw status=failed")
			}
		}
	}

	return count, nil
}
