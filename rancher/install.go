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
	"strings"

	"github.com/rancher-sandbox/ele-testhelpers/kubectl"
)

/**
 * Install or upgrade Rancher Manager
 * @remarks Deploy a Rancher Manager instance
 * @param hostname Hostname/URL to use for the deployment
 * @param channel Rancher channel to use (stable, latest)
 * @param version Rancher version to install (latest, devel)
 * @param headVersion Rancher head version to install (2.7, 2.8)
 * @param ca CA to use (selfsigned, private)
 * @param proxy Define if a a proxy should be configured/used
 * @returns Nothing or an error
 */
// NOTE: AddNode does not have unit test as it is not easy to mock
func DeployRancherManager(hostname, channel, version, headVersion, ca, proxy string) error {
	const password = "rancherpassword"
	channelName := "rancher-" + channel

	// Add Helm repository
	err := kubectl.RunHelmBinaryWithCustomErr("repo", "add", channelName,
		"https://releases.rancher.com/server-charts/"+channel,
	)
	if err != nil {
		return err
	}

	if err = kubectl.RunHelmBinaryWithCustomErr("repo", "update"); err != nil {
		return err
	}

	// Set flags for Rancher Manager installation
	flags := []string{
		"upgrade", "--install", "rancher", channelName + "/rancher",
		"--namespace", "cattle-system",
		"--create-namespace",
		"--set", "hostname=" + hostname,
		"--set", "bootstrapPassword=" + password,
		"--set", "extraEnv[0].name=CATTLE_SERVER_URL",
		"--set", "extraEnv[0].value=https://" + hostname,
		"--set", "extraEnv[1].name=CATTLE_BOOTSTRAP_PASSWORD",
		"--set", "extraEnv[1].value=" + password,
		"--set", "replicas=1",
	}

	// Set specified version if needed
	if version != "" && version != "latest" {
		if version == "devel" {
			flags = append(flags,
				"--devel",
				"--set", "rancherImageTag=v"+headVersion+"-head",
			)
		} else if strings.Contains(version, "-rc") {
			flags = append(flags,
				"--devel",
				"--version", version,
			)
		} else {
			flags = append(flags, "--version", version)
		}
	}

	// For Private CA
	if ca == "private" {
		flags = append(flags,
			"--set", "ingress.tls.source=secret",
			"--set", "privateCA=true",
		)
	}

	// Use Rancher Manager behind proxy
	if proxy == "rancher" {
		flags = append(flags,
			"--set", "proxy=http://172.17.0.1:3128",
			"--set", "noProxy=127.0.0.0/8\\,10.0.0.0/8\\,cattle-system.svc\\,172.16.0.0/12\\,192.168.0.0/16\\,.svc\\,.cluster.local",
		)
	}

	return kubectl.RunHelmBinaryWithCustomErr(flags...)
}
