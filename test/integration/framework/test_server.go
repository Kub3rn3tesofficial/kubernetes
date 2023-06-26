/*
Copyright 2018 The Kubernetes Authors.

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

package framework

import (
	"context"
	"net"
	"net/http"
	"os"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/apimachinery/pkg/util/wait"
	genericapiserver "k8s.io/apiserver/pkg/server"
	genericapiserveroptions "k8s.io/apiserver/pkg/server/options"
	client "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/util/cert"
	"k8s.io/kubernetes/cmd/kube-apiserver/app"
	"k8s.io/kubernetes/cmd/kube-apiserver/app/options"
	"k8s.io/kubernetes/pkg/controlplane"
	"k8s.io/kubernetes/test/utils"
	netutils "k8s.io/utils/net"
)

// This key is for testing purposes only and is not considered secure.
const ecdsaPrivateKey = `-----BEGIN EC PRIVATE KEY-----
MHcCAQEEIEZmTmUhuanLjPA2CLquXivuwBDHTt5XYwgIr/kA1LtRoAoGCCqGSM49
AwEHoUQDQgAEH6cuzP8XuD5wal6wf9M6xDljTOPLX2i8uIp/C/ASqiIGUeeKQtX0
/IR3qCXyThP/dbCiHrF3v1cuhBOHY8CLVg==
-----END EC PRIVATE KEY-----`

// TestServerSetup holds configuration information for a kube-apiserver test server.
type TestServerSetup struct {
	ModifyServerRunOptions func(*options.ServerRunOptions)
	ModifyServerConfig     func(*controlplane.Config)
}

type TearDownFunc func()

// StartTestServer runs a kube-apiserver, optionally calling out to the setup.ModifyServerRunOptions and setup.ModifyServerConfig functions
func StartTestServer(ctx context.Context, t testing.TB, setup TestServerSetup) (client.Interface, *rest.Config, TearDownFunc) {
	ctx, cancel := context.WithCancel(ctx)

	certDir, err := os.MkdirTemp("", "test-integration-"+strings.ReplaceAll(t.Name(), "/", "_"))
	if err != nil {
		t.Fatalf("Couldn't create temp dir: %v", err)
	}

	var errCh chan error
	tearDownFn := func() {
		// Calling cancel function is stopping apiserver and cleaning up
		// after itself, including shutting down its storage layer.
		cancel()

		// If the apiserver was started, let's wait for it to
		// shutdown clearly.
		if errCh != nil {
			err, ok := <-errCh
			if ok && err != nil {
				t.Error(err)
			}
		}
		if err := os.RemoveAll(certDir); err != nil {
			t.Log(err)
		}
	}

	_, defaultServiceClusterIPRange, _ := netutils.ParseCIDRSloppy("10.0.0.0/24")
	proxySigningKey, err := utils.NewPrivateKey()
	if err != nil {
		t.Fatal(err)
	}
	proxySigningCert, err := cert.NewSelfSignedCACert(cert.Config{CommonName: "front-proxy-ca"}, proxySigningKey)
	if err != nil {
		t.Fatal(err)
	}
	proxyCACertFile, _ := os.CreateTemp(certDir, "proxy-ca.crt")
	if err := os.WriteFile(proxyCACertFile.Name(), utils.EncodeCertPEM(proxySigningCert), 0644); err != nil {
		t.Fatal(err)
	}
	defer proxyCACertFile.Close()
	clientSigningKey, err := utils.NewPrivateKey()
	if err != nil {
		t.Fatal(err)
	}
	clientSigningCert, err := cert.NewSelfSignedCACert(cert.Config{CommonName: "client-ca"}, clientSigningKey)
	if err != nil {
		t.Fatal(err)
	}
	clientCACertFile, _ := os.CreateTemp(certDir, "client-ca.crt")
	if err := os.WriteFile(clientCACertFile.Name(), utils.EncodeCertPEM(clientSigningCert), 0644); err != nil {
		t.Fatal(err)
	}
	defer clientCACertFile.Close()
	listener, _, err := genericapiserveroptions.CreateListener("tcp", "127.0.0.1:0", net.ListenConfig{})
	if err != nil {
		t.Fatal(err)
	}

	saSigningKeyFile, err := os.CreateTemp("/tmp", "insecure_test_key")
	if err != nil {
		t.Fatalf("create temp file failed: %v", err)
	}
	defer saSigningKeyFile.Close()
	if err = os.WriteFile(saSigningKeyFile.Name(), []byte(ecdsaPrivateKey), 0666); err != nil {
		t.Fatalf("write file %s failed: %v", saSigningKeyFile.Name(), err)
	}

	opts := options.NewServerRunOptions()
	opts.GenericControlPlane.SecureServing.Listener = listener
	opts.GenericControlPlane.SecureServing.BindAddress = netutils.ParseIPSloppy("127.0.0.1")
	opts.GenericControlPlane.SecureServing.ServerCert.CertDirectory = certDir
	opts.GenericControlPlane.ServiceAccountSigningKeyFile = saSigningKeyFile.Name()
	opts.GenericControlPlane.Etcd.StorageConfig.Prefix = path.Join("/", uuid.New().String(), "registry")
	opts.GenericControlPlane.Etcd.StorageConfig.Transport.ServerList = []string{GetEtcdURL()}
	opts.ServiceClusterIPRanges = defaultServiceClusterIPRange.String()
	opts.GenericControlPlane.Authentication.RequestHeader.UsernameHeaders = []string{"X-Remote-User"}
	opts.GenericControlPlane.Authentication.RequestHeader.GroupHeaders = []string{"X-Remote-Group"}
	opts.GenericControlPlane.Authentication.RequestHeader.ExtraHeaderPrefixes = []string{"X-Remote-Extra-"}
	opts.GenericControlPlane.Authentication.RequestHeader.AllowedNames = []string{"kube-aggregator"}
	opts.GenericControlPlane.Authentication.RequestHeader.ClientCAFile = proxyCACertFile.Name()
	opts.GenericControlPlane.Authentication.APIAudiences = []string{"https://foo.bar.example.com"}
	opts.GenericControlPlane.Authentication.ServiceAccounts.Issuers = []string{"https://foo.bar.example.com"}
	opts.GenericControlPlane.Authentication.ServiceAccounts.KeyFiles = []string{saSigningKeyFile.Name()}
	opts.GenericControlPlane.Authentication.ClientCert.ClientCA = clientCACertFile.Name()
	opts.GenericControlPlane.Authorization.Modes = []string{"Node", "RBAC"}

	if setup.ModifyServerRunOptions != nil {
		setup.ModifyServerRunOptions(opts)
	}

	completedOptions, err := opts.Complete()
	if err != nil {
		t.Fatal(err)
	}

	if errs := completedOptions.Validate(); len(errs) != 0 {
		t.Fatalf("failed to validate ServerRunOptions: %v", utilerrors.NewAggregate(errs))
	}

	kubeAPIServerConfig, _, _, err := app.CreateKubeAPIServerConfig(completedOptions)
	if err != nil {
		t.Fatal(err)
	}

	if setup.ModifyServerConfig != nil {
		setup.ModifyServerConfig(kubeAPIServerConfig)
	}
	kubeAPIServer, err := kubeAPIServerConfig.Complete().New(genericapiserver.NewEmptyDelegate())
	if err != nil {
		t.Fatal(err)
	}

	errCh = make(chan error)
	go func() {
		defer close(errCh)
		if err := kubeAPIServer.GenericAPIServer.PrepareRun().Run(ctx.Done()); err != nil {
			errCh <- err
		}
	}()

	// Adjust the loopback config for external use (external server name and CA)
	kubeAPIServerClientConfig := rest.CopyConfig(kubeAPIServerConfig.GenericConfig.LoopbackClientConfig)
	kubeAPIServerClientConfig.CAFile = path.Join(certDir, "apiserver.crt")
	kubeAPIServerClientConfig.CAData = nil
	kubeAPIServerClientConfig.ServerName = ""

	// wait for health
	err = wait.PollImmediate(100*time.Millisecond, 10*time.Second, func() (done bool, err error) {
		select {
		case err := <-errCh:
			return false, err
		default:
		}

		healthzConfig := rest.CopyConfig(kubeAPIServerClientConfig)
		healthzConfig.ContentType = ""
		healthzConfig.AcceptContentTypes = ""
		kubeClient, err := client.NewForConfig(healthzConfig)
		if err != nil {
			// this happens because we race the API server start
			t.Log(err)
			return false, nil
		}

		healthStatus := 0
		kubeClient.Discovery().RESTClient().Get().AbsPath("/healthz").Do(ctx).StatusCode(&healthStatus)
		if healthStatus != http.StatusOK {
			return false, nil
		}

		if _, err := kubeClient.CoreV1().Namespaces().Get(ctx, "default", metav1.GetOptions{}); err != nil {
			return false, nil
		}
		if _, err := kubeClient.CoreV1().Namespaces().Get(ctx, "kube-system", metav1.GetOptions{}); err != nil {
			return false, nil
		}

		return true, nil
	})
	if err != nil {
		t.Fatal(err)
	}

	kubeAPIServerClient, err := client.NewForConfig(kubeAPIServerClientConfig)
	if err != nil {
		t.Fatal(err)
	}

	return kubeAPIServerClient, kubeAPIServerClientConfig, tearDownFn
}
