//go:build windows
// +build windows

/*
Copyright 2022 The Kubernetes Authors.

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

package dns

import (
	"fmt"
	"strings"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
	runtimeapi "k8s.io/cri-api/pkg/apis/runtime/v1"

	"k8s.io/klog/v2"
)

const (
	netRegistry       = `System\CurrentControlSet\Services\TCPIP\Parameters`
	netIfacesRegistry = `System\CurrentControlSet\Services\TCPIP\Parameters\Interfaces`
	maxHostnameLen    = 128
	maxDomainNameLen  = 128
	maxScopeIDLen     = 256
)

// FixedInfo information: https://docs.microsoft.com/en-us/windows/win32/api/iptypes/ns-iptypes-fixed_info_w2ksp1
type FixedInfo struct {
	HostName         [maxHostnameLen + 4]byte
	DomainName       [maxDomainNameLen + 4]byte
	CurrentDNSServer *syscall.IpAddrString
	DNSServerList    syscall.IpAddrString
	NodeType         uint32
	ScopeID          [maxScopeIDLen + 4]byte
	EnableRouting    uint32
	EnableProxy      uint32
	EnableDNS        uint32
}

var (
	// GetNetworkParams can be found in iphlpapi.dll
	// see: https://docs.microsoft.com/en-us/windows/win32/api/iphlpapi/nf-iphlpapi-getnetworkparams?redirectedfrom=MSDN
	iphlpapidll          = windows.MustLoadDLL("iphlpapi.dll")
	procGetNetworkParams = iphlpapidll.MustFindProc("GetNetworkParams")
)

func getHostDNSConfig(resolverConfig string) (*runtimeapi.DNSConfig, error) {
	var (
		hostDNS, hostSearch []string
		err                 error
	)
	// Get host DNS settings
	if resolverConfig != "" {
		hostDNS, err = getDNSServerList()
		if err != nil {
			klog.ErrorS(err, "Could not get the host's DNS Server List")
			return nil, err
		}
		hostSearch, err = getDNSSuffixList()
		if err != nil {
			klog.ErrorS(err, "Could not get the host's DNS Suffix List")
			return nil, err
		}
	}
	return &runtimeapi.DNSConfig{
		Servers:  hostDNS,
		Searches: hostSearch,
	}, nil
}

func elemInList(elem string, list []string) bool {
	for _, e := range list {
		if e == elem {
			return true
		}
	}
	return false
}

func getRegistryValue(reg, key string) string {
	regKey, err := registry.OpenKey(registry.LOCAL_MACHINE, reg, registry.QUERY_VALUE)
	if err != nil {
		return ""
	}
	defer regKey.Close()

	regValue, _, err := regKey.GetStringValue(key)
	if err != nil {
		return ""
	}
	return regValue
}

// getDNSSuffixList reads DNS config file and returns the list of configured DNS suffixes
func getDNSSuffixList() ([]string, error) {
	// We start with the general suffix list that apply to all network connections.
	allSuffixes := []string{}
	suffixes := getRegistryValue(netRegistry, "SearchList")
	if suffixes != "" {
		allSuffixes = strings.Split(suffixes, ",")
	}

	// Then we append the network-specific DNS suffix lists.
	regKey, err := registry.OpenKey(registry.LOCAL_MACHINE, netIfacesRegistry, registry.ENUMERATE_SUB_KEYS)
	if err != nil {
		return nil, err
	}
	defer regKey.Close()

	ifaces, err := regKey.ReadSubKeyNames(0)
	if err != nil {
		return nil, err
	}
	for _, iface := range ifaces {
		suffixes := getRegistryValue(fmt.Sprintf("%s\\%s", netIfacesRegistry, iface), "SearchList")
		if suffixes == "" {
			continue
		}
		for _, suffix := range strings.Split(suffixes, ",") {
			if !elemInList(suffix, allSuffixes) {
				allSuffixes = append(allSuffixes, suffix)
			}
		}
	}

	return allSuffixes, nil
}

func getNetworkParams() (*FixedInfo, error) {
	// We don't know how big we should make the byte buffer, but the call will tell us by
	// setting the size afterwards.
	var size int
	buffer := make([]byte, 1)
	procGetNetworkParams.Call(
		uintptr(unsafe.Pointer(&buffer[0])),
		uintptr(unsafe.Pointer(&size)),
	)

	buffer = make([]byte, size)
	ret, _, err := procGetNetworkParams.Call(
		uintptr(unsafe.Pointer(&buffer[0])),
		uintptr(unsafe.Pointer(&size)),
	)
	if ret != 0 {
		return nil, err
	}

	info := (*FixedInfo)(unsafe.Pointer(&buffer[0]))
	return info, nil
}

func getDNSServerList() ([]string, error) {
	dnsServerList := []string{}
	fixedInfo, err := getNetworkParams()
	if err != nil {
		return nil, err
	}

	list := &(fixedInfo.DNSServerList)
	for list != nil {
		dnsServer := strings.TrimRight(string(list.IpAddress.String[:]), "\x00")
		dnsServerList = append(dnsServerList, dnsServer)
		list = list.Next
	}
	return dnsServerList, nil
}
