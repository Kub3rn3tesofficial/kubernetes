/*
Copyright 2015 The Kubernetes Authors All rights reserved.

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

package proxy

//PortalManager is an interface that abstracts over Managing Portals.
//See IptablesPortalManager for an implementation.
type PortalManager interface {
	Init() error
	OpenPortal(proxier *Proxier, service ServicePortName, info *serviceInfo) error
	ClosePortal(proxier *Proxier, service ServicePortName, info *serviceInfo) error
	DeleteOld()
	Flush() error
}
