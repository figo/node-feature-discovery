/*
Copyright 2017 The Kubernetes Authors.

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

package network

import (
	"bytes"
	"fmt"
	"github.com/golang/glog"
	"io/ioutil"
	"net"
	"strconv"
	"strings"
)

// Source implements FeatureSource.
type Source struct{}

// Name returns an identifier string for this feature source.
func (s Source) Name() string { return "network" }

// Discover returns feature names sriov-configured and sriov if SR-IOV capable NICs are present and/or SR-IOV virtual functions are configured on the node
func (s Source) Discover() ([]string, error) {
	features := []string{}
	netInterfaces, err := net.Interfaces()
	if err != nil {
		return nil, fmt.Errorf("can't obtain the network interfaces details: %s", err.Error())
	}
	// iterating through network interfaces to obtain their respective number of virtual functions
	for _, netInterface := range netInterfaces {
		if strings.Contains(netInterface.Flags.String(), "up") && !strings.Contains(netInterface.Flags.String(), "loopback") {
			totalVfsPath := "/sys/class/net/" + netInterface.Name + "/device/sriov_totalvfs"
			totalBytes, err := ioutil.ReadFile(totalVfsPath)
			if err != nil {
				glog.Errorf("SR-IOV not supported for network interface: %s: %v", netInterface.Name, err)
				continue
			}
			total := bytes.TrimSpace(totalBytes)
			t, err := strconv.Atoi(string(total))
			if err != nil {
				glog.Errorf("Error in obtaining maximum supported number of virtual functions for network interface: %s: %v", netInterface.Name, err)
				continue
			}
			if t > 0 {
				glog.Infof("SR-IOV capability is detected on the network interface: %s", netInterface.Name)
				glog.Infof("%d maximum supported number of virtual functions on network interface: %s", t, netInterface.Name)
				features = append(features, "sriov.capable")
				numVfsPath := "/sys/class/net/" + netInterface.Name + "/device/sriov_numvfs"
				numBytes, err := ioutil.ReadFile(numVfsPath)
				if err != nil {
					glog.Errorf("SR-IOV not configured for network interface: %s: %s", netInterface.Name, err)
					continue
				}
				num := bytes.TrimSpace(numBytes)
				n, err := strconv.Atoi(string(num))
				if err != nil {
					glog.Errorf("Error in obtaining the configured number of virtual functions for network interface: %s: %v", netInterface.Name, err)
					continue
				}
				if n > 0 {
					glog.Infof("%d virtual functions configured on network interface: %s", n, netInterface.Name)
					features = append(features, "sriov.configured")
					break
				} else if n == 0 {
					glog.Errorf("SR-IOV not configured on network interface: %s", netInterface.Name)
				}
			}
		}
	}
	return features, nil
}
