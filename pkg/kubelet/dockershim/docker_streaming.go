/*
Copyright 2016 The Kubernetes Authors.

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

package dockershim

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"net"
	"os/exec"
	"strings"
	"time"

	dockertypes "github.com/docker/docker/api/types"

	"github.com/golang/glog"

	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/tools/remotecommand"
	api "k8s.io/kubernetes/pkg/apis/core"
	runtimeapi "k8s.io/kubernetes/pkg/kubelet/apis/cri/v1alpha1/runtime"
	kubecontainer "k8s.io/kubernetes/pkg/kubelet/container"
	"k8s.io/kubernetes/pkg/kubelet/server/streaming"
	"k8s.io/kubernetes/pkg/kubelet/util/ioutils"

	"k8s.io/kubernetes/pkg/kubelet/dockershim/libdocker"
)

type streamingRuntime struct {
	client      libdocker.Interface
	execHandler ExecHandler
}

var _ streaming.Runtime = &streamingRuntime{}

func (r *streamingRuntime) Exec(containerID string, cmd []string, in io.Reader, out, err io.WriteCloser, tty bool, resize <-chan remotecommand.TerminalSize) error {
	return r.exec(containerID, cmd, in, out, err, tty, resize, 0)
}

// Internal version of Exec adds a timeout.
func (r *streamingRuntime) exec(containerID string, cmd []string, in io.Reader, out, errw io.WriteCloser, tty bool, resize <-chan remotecommand.TerminalSize, timeout time.Duration) error {
	container, err := checkContainerStatus(r.client, containerID)
	if err != nil {
		return err
	}
	return r.execHandler.ExecInContainer(r.client, container, cmd, in, out, errw, tty, resize, timeout)
}

func (r *streamingRuntime) Attach(containerID string, in io.Reader, out, errw io.WriteCloser, tty bool, resize <-chan remotecommand.TerminalSize) error {
	_, err := checkContainerStatus(r.client, containerID)
	if err != nil {
		return err
	}

	return attachContainer(r.client, containerID, in, out, errw, tty, resize)
}

//need udp
func (r *streamingRuntime) PortForward(podSandboxID string, protocol string, port int32, stream io.ReadWriteCloser) error {
	glog.V(3).Infof("haha: into func (r *streamingRuntime) PortForward(podSandboxID string, protocol string, port int32, stream io.ReadWriteCloser) error ")

	if port < 0 || port > math.MaxUint16 {
		return fmt.Errorf("invalid port %d", port)
	}
	if protocol != api.PortForwardProtocolTypeTcp4 && protocol != api.PortForwardProtocolTypeUdp4 {
		return fmt.Errorf("invalid or not supported protocol %s", protocol)
	}
	return portForward(r.client, podSandboxID, protocol, port, stream)
}

// ExecSync executes a command in the container, and returns the stdout output.
// If command exits with a non-zero exit code, an error is returned.
func (ds *dockerService) ExecSync(containerID string, cmd []string, timeout time.Duration) (stdout []byte, stderr []byte, err error) {
	var stdoutBuffer, stderrBuffer bytes.Buffer
	err = ds.streamingRuntime.exec(containerID, cmd,
		nil, // in
		ioutils.WriteCloserWrapper(&stdoutBuffer),
		ioutils.WriteCloserWrapper(&stderrBuffer),
		false, // tty
		nil,   // resize
		timeout)
	return stdoutBuffer.Bytes(), stderrBuffer.Bytes(), err
}

// Exec prepares a streaming endpoint to execute a command in the container, and returns the address.
func (ds *dockerService) Exec(req *runtimeapi.ExecRequest) (*runtimeapi.ExecResponse, error) {
	if ds.streamingServer == nil {
		return nil, streaming.ErrorStreamingDisabled("exec")
	}
	_, err := checkContainerStatus(ds.client, req.ContainerId)
	if err != nil {
		return nil, err
	}
	return ds.streamingServer.GetExec(req)
}

// Attach prepares a streaming endpoint to attach to a running container, and returns the address.
func (ds *dockerService) Attach(req *runtimeapi.AttachRequest) (*runtimeapi.AttachResponse, error) {
	if ds.streamingServer == nil {
		return nil, streaming.ErrorStreamingDisabled("attach")
	}
	_, err := checkContainerStatus(ds.client, req.ContainerId)
	if err != nil {
		return nil, err
	}
	return ds.streamingServer.GetAttach(req)
}

// PortForward prepares a streaming endpoint to forward ports from a PodSandbox, and returns the address.
func (ds *dockerService) PortForward(req *runtimeapi.PortForwardRequest) (*runtimeapi.PortForwardResponse, error) {
	if ds.streamingServer == nil {
		return nil, streaming.ErrorStreamingDisabled("port forward")
	}
	_, err := checkContainerStatus(ds.client, req.PodSandboxId)
	if err != nil {
		return nil, err
	}
	// TODO(tallclair): Verify that ports are exposed.
	return ds.streamingServer.GetPortForward(req)
}

func checkContainerStatus(client libdocker.Interface, containerID string) (*dockertypes.ContainerJSON, error) {
	container, err := client.InspectContainer(containerID)
	if err != nil {
		return nil, err
	}
	if !container.State.Running {
		return nil, fmt.Errorf("container not running (%s)", container.ID)
	}
	return container, nil
}

func attachContainer(client libdocker.Interface, containerID string, stdin io.Reader, stdout, stderr io.WriteCloser, tty bool, resize <-chan remotecommand.TerminalSize) error {
	// Have to start this before the call to client.AttachToContainer because client.AttachToContainer is a blocking
	// call :-( Otherwise, resize events don't get processed and the terminal never resizes.
	kubecontainer.HandleResizing(resize, func(size remotecommand.TerminalSize) {
		client.ResizeContainerTTY(containerID, uint(size.Height), uint(size.Width))
	})

	// TODO(random-liu): Do we really use the *Logs* field here?
	opts := dockertypes.ContainerAttachOptions{
		Stream: true,
		Stdin:  stdin != nil,
		Stdout: stdout != nil,
		Stderr: stderr != nil,
	}
	sopts := libdocker.StreamOptions{
		InputStream:  stdin,
		OutputStream: stdout,
		ErrorStream:  stderr,
		RawTerminal:  tty,
	}
	return client.AttachToContainer(containerID, opts, sopts)
}

func protForwardTcp(containerPid int, port int32, stream io.ReadWriteCloser) error {
	socatPath, lookupErr := exec.LookPath("socat")
	if lookupErr != nil {
		return fmt.Errorf("unable to do port forwarding: socat not found.")
	}

	args := []string{"-t", fmt.Sprintf("%d", containerPid), "-n", socatPath, "-", fmt.Sprintf("TCP4:localhost:%d", port)}

	nsenterPath, lookupErr := exec.LookPath("nsenter")
	if lookupErr != nil {
		return fmt.Errorf("unable to do port forwarding: nsenter not found.")
	}

	commandString := fmt.Sprintf("%s %s", nsenterPath, strings.Join(args, " "))
	glog.V(4).Infof("executing port forwarding command: %s", commandString)

	command := exec.Command(nsenterPath, args...)
	command.Stdout = stream

	stderr := new(bytes.Buffer)
	command.Stderr = stderr

	// If we use Stdin, command.Run() won't return until the goroutine that's copying
	// from stream finishes. Unfortunately, if you have a client like telnet connected
	// via port forwarding, as long as the user's telnet client is connected to the user's
	// local listener that port forwarding sets up, the telnet session never exits. This
	// means that even if socat has finished running, command.Run() won't ever return
	// (because the client still has the connection and stream open).
	//
	// The work around is to use StdinPipe(), as Wait() (called by Run()) closes the pipe
	// when the command (socat) exits.
	inPipe, err := command.StdinPipe()
	if err != nil {
		return fmt.Errorf("unable to do port forwarding: error creating stdin pipe: %v", err)
	}
	go func() {
		io.Copy(inPipe, stream)
		inPipe.Close()
	}()

	if err := command.Run(); err != nil {
		return fmt.Errorf("%v: %s", err, stderr.String())
	}

	return nil
}

func portForwardUdp(containerPid int, port int32, stream io.ReadWriteCloser) error {
	unixDomainSocketPath := "/tmp/my.sock"
	unixDomainSocketPath1 := "/tmp/1.sock"
	socatPath, lookupErr := exec.LookPath("socat")
	if lookupErr != nil {
		return fmt.Errorf("unable to do port forwarding: socat not found.")
	}

	args := []string{"-t", fmt.Sprintf("%d", containerPid), "-n", socatPath,
	fmt.Sprintf("UNIX-RECVFROM:%s,fork", unixDomainSocketPath), fmt.Sprintf("UDP:localhost:%d", port)}

	nsenterPath, lookupErr := exec.LookPath("nsenter")
	if lookupErr != nil {
		return fmt.Errorf("unable to do port forwarding: nsenter not found.")
	}

	commandString := fmt.Sprintf("%s %s", nsenterPath, strings.Join(args, " "))
	glog.V(3).Infof("haha: executing port forwarding command: %s", commandString)

	command := exec.Command(nsenterPath, args...)
	command.Stdout = stream

	stderr := new(bytes.Buffer)
	command.Stderr = stderr
	glog.V(3).Infof("haha:  args are %s\n", args)

	go func() error {
		err := command.Run();
		if err != nil {
			glog.V(3).Infof("haha: execute went error \n")
			return fmt.Errorf("%v: %s", err, stderr.String())
		}
		return nil
	}()

	time.Sleep(3*time.Second)

	a, err := net.ResolveUnixAddr("unixgram", unixDomainSocketPath)
	if err != nil {
		return err
	}

	b, err := net.ResolveUnixAddr("unixgram", unixDomainSocketPath1)
	if err != nil {
		return err
	}

	//read stream from spdy, send buf to socat, then receive from socat send back to spdy
	connUDP, err := net.DialUnix("unixgram", b, a)
	if err != nil {
		glog.V(3).Infof("haha: dial went error %s\n", err)
		return err
	}
	glog.V(3).Infof("haha:  dial unixgram success\n")
	portforward.ReadFromStreamAndSendToUDP(stream, connUDP)


	return nil
}

func portForward(client libdocker.Interface, podSandboxID string, protocol string, port int32, stream io.ReadWriteCloser) error {
	container, err := client.InspectContainer(podSandboxID)
	if err != nil {
		return err
	}

	if !container.State.Running {
		return fmt.Errorf("container not running (%s)", container.ID)
	}

	containerPid := container.State.Pid

	glog.V(3).Infof("haha: portForward | the protocol is %s", protocol)
	if protocol == api.PortForwardProtocolTypeUdp4 ||
		protocol == api.PortForwardProtocolTypeUdp6 {
		return portForwardUdp(containerPid, port, stream)
	}

	return protForwardTcp(containerPid, port, stream)
}
