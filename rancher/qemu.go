/*
Copyright © 2022 - 2023 SUSE LLC

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

package rancher

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"

	libvirtxml "libvirt.org/libvirt-go-xml"
)

/**
 * Add a node in libvirt configuration
 * @remarks Add DHCP/DNS configuration in libvirt network config
 * @param file File to modify
 * @param name Node name
 * @param index Index of the node
 * @returns Nothing or an error
 */
func AddNode(file, name string, index int) error {
	// Read live XML configuration
	fileContent, err := exec.Command("sudo", "virsh", "net-dumpxml", "default").Output()
	if err != nil {
		return err
	}

	// Unmarshal fileContent
	netcfg := &libvirtxml.Network{}
	if err := netcfg.Unmarshal(string(fileContent)); err != nil {
		return err
	}

	// Add new host
	// NOTE: we only use one network (IPs[0])
	// TODO: index could be calculated
	host := libvirtxml.NetworkDHCPHost{
		Name: name,
		MAC:  "52:54:00:00:00:" + fmt.Sprintf("%02x", index),
		IP:   "192.168.122." + strconv.Itoa(index+1),
	}
	netcfg.IPs[0].DHCP.Hosts = append(netcfg.IPs[0].DHCP.Hosts, host)

	// Marshal new content
	newFileContent, err := netcfg.Marshal()
	if err != nil {
		return err
	}

	// Re-write XML file
	if err := os.WriteFile(file, []byte(newFileContent), 0644); err != nil {
		return err
	}

	// Update live network configuration
	xmlValue, err := host.Marshal()
	if err != nil {
		return err
	}

	return exec.Command("sudo", "virsh", "net-update",
		"default", "add", "ip-dhcp-host", "--live", "--xml", xmlValue).Run()
}
