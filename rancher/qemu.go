/*
Copyright © 2022 - 2025 SUSE LLC

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
	"encoding/xml"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"

	libvirtxml "libvirt.org/libvirt-go-xml"
)

type Host struct {
	XMLName xml.Name `xml:"host"`
	Mac     string   `xml:"mac,attr"`
	Name    string   `xml:"name,attr"`
	IP      string   `xml:"ip,attr"`
}

/**
 * Get information from network/host config file
 * @param regex Regex to use for searching information
 * @param file Name of the file to check
 * @returns Pointer to the Host structure or an error
 */
func GetHostNetConfig(regex, file string) (*Host, error) {
	fileData, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	r := regexp.MustCompile(regex)
	find := r.FindString(string(fileData))

	host := &Host{}
	if err = xml.Unmarshal([]byte(find), host); err != nil {
		return nil, err
	}

	return host, nil
}

/**
 * Add a node in libvirt configuration
 * @remarks Add DHCP/DNS configuration in libvirt network config
 * @param file File to modify
 * @param name Node name
 * @param index Index of the node
 * @returns Nothing or an error
 */
// NOTE: AddNode does not have unit test as it is not easy to mock
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
