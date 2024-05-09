/*
Copyright © 2022 - 2024 SUSE LLC

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
	"os"
	"strings"

	"github.com/rancher-sandbox/ele-testhelpers/kubectl"
)

/**
 * Install or upgrade Rancher Manager
 * @remarks Deploy a Rancher Manager instance
 * @param hostname Hostname/URL to use for the deployment
 * @param channel Rancher channel to use (stable, latest, prime, prime-optimus, alpha)
 * @param version Rancher version to install (latest, devel)
 * @param headVersion Rancher head version to install (2.7, 2.8, 2.9)
 * @param ca CA to use (selfsigned, private)
 * @param proxy Define if a a proxy should be configured/used
 * @returns Nothing or an error
 */
// NOTE: AddNode does not have unit test as it is not easy to mock
func DeployRancherManager(hostname, channel, version, headVersion, ca, proxy string) error {
	var password = "rancherpassword"
	if envPW := os.Getenv("RANCHER_PASSWORD"); envPW != "" {
		password = envPW
	}

	channelName := "rancher-" + channel
	var chartRepo string

	switch channel {
	case "prime":
		chartRepo = "https://charts.rancher.com/server-charts/prime"
	case "prime-optimus":
		chartRepo = "https://charts.optimus.rancher.io/server-charts/latest"
	case "alpha":
		chartRepo = "https://releases.rancher.com/server-charts/alpha"
	case "latest":
		chartRepo = "https://releases.rancher.com/server-charts/latest"
	case "stable":
		chartRepo = "https://releases.rancher.com/server-charts/stable"
	}

	// Add Helm repository
	if err := kubectl.RunHelmBinaryWithCustomErr("repo", "add", channelName, chartRepo); err != nil {
		return err
	}

	if err := kubectl.RunHelmBinaryWithCustomErr("repo", "update"); err != nil {
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
		"--set", "replicas=1",
		"--set", "useBundledSystemChart=true",
		"--wait", "--wait-for-jobs",
	}

	// Set specified version if needed
	if version != "" && version != "latest" {
		if version == "devel" {
			flags = append(flags,
				"--devel",
				"--set", "rancherImageTag=v"+headVersion+"-head",
			)
			// Devel image rancher:v2.7-head available only on stgregistry.suse.com
			if headVersion == "2.7" {
				flags = append(flags,
					"--set", "rancherImage=stgregistry.suse.com/rancher/rancher",
					"--set", "extraEnv[1].name=CATTLE_AGENT_IMAGE",
					"--set", "extraEnv[1].value=stgregistry.suse.com/rancher/rancher-agent:v"+headVersion+"-head",
				)
			}
		} else if strings.Contains(version, "-rc") {
			flags = append(flags,
				"--devel",
				"--version", version,
			)
			// For rancher:2.7.x-rc from prime-optimus channel only
			if strings.Contains(version, "2.7.") && channel == "prime-optimus" {
				flags = append(flags,
					// no need to set rancherImageTag as it is already set in the chart
					"--set", "rancherImage=stgregistry.suse.com/rancher/rancher",
					"--set", "extraEnv[1].name=CATTLE_AGENT_IMAGE",
					"--set", "extraEnv[1].value=stgregistry.suse.com/rancher/rancher-agent:v"+version,
				)
			}
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
