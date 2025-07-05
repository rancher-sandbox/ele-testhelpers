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
	"os"
	"regexp"
	"strings"

	"github.com/rancher-sandbox/ele-testhelpers/kubectl"
)

/** Support function for populating correct helm flags for Devel versions
 * @param flags Helm flags
 * @param headVersion Rancher head version
 * @returns flags with correct values
 */
func appendDevelFlags(flags []string, headVersion string) []string {

	// Regex pattern for 2.12 to 2.99 but not 2.7, 2.8, 2.9, 2.10 and 2.11
	pattern := `^2\.(1[2-9]|[2-9]\d)$`
	re := regexp.MustCompile(pattern)

	switch {
	case headVersion == "head":
		// As of 04/2025 this can be used as "latest/devel/head" to test v2.12-head
		flags = append(flags,
			"--devel",
			"--set", "rancherImageTag=head",
			"--set", "extraEnv[1].name=CATTLE_AGENT_IMAGE",
			"--set", "extraEnv[1].value=rancher/rancher-agent:head",
		)
	case re.MatchString(headVersion):
		// If the version matches the regex, like 2.12 and up.
		flags = append(flags,
			"--devel",
			"--set", "rancherImageTag=v"+headVersion+"-head",
			"--set", "extraEnv[1].name=CATTLE_AGENT_IMAGE",
			"--set", "extraEnv[1].value=rancher/rancher-agent:v"+headVersion+"-head",
		)
	default:
		// Devel images for rancher:v2\.(7|8|9|10|11)-head are available on stgregistry.suse.com
		flags = append(flags,
			"--devel",
			"--set", "rancherImageTag=v"+headVersion+"-head",
			"--set", "rancherImage=stgregistry.suse.com/rancher/rancher",
			"--set", "extraEnv[1].name=CATTLE_AGENT_IMAGE",
			"--set", "extraEnv[1].value=stgregistry.suse.com/rancher/rancher-agent:v"+headVersion+"-head",
		)
	}
	return flags
}

/** Support function for populating correct helm flags for Head versions in "head" channel
 * @param flags Helm flags
 * @returns flags with correct values
 */
func appendHeadFlags(flags []string) []string {

	// For Rancher versions 2.10 and 2.11, head images are available on stgregistry.suse.com
	// For Rancher version 2.12, head images are available on the dockerhub registry
	// For all versions there is no need to provide extra flags, only the --devel flag is needed
	flags = append(flags,
		"--devel",
	)
	return flags
}

/** Support function for populating correct helm flags for RC and Alpha versions
 * @param flags Helm flags
 * @param version Rancher version
 * @param channel Rancher channel
 * @returns flags with correct values
 */
func appendRCAlphaFlags(flags []string, version string, channel string) []string {
	flags = append(flags,
		"--devel",
		"--version", version,
	)
	// For rancher:2.x.y-rc from prime-optimus and prime-optimus-alpha channel only
	if strings.Contains(channel, "prime-optimus") {
		flags = append(flags,
			"--set", "rancherImage=stgregistry.suse.com/rancher/rancher",
			"--set", "extraEnv[1].name=CATTLE_AGENT_IMAGE",
			"--set", "extraEnv[1].value=stgregistry.suse.com/rancher/rancher-agent:v"+version,
		)
	}
	return flags
}

/**
 * Install or upgrade Rancher Manager
 * @remarks Deploy a Rancher Manager instance
 * @param hostname Hostname/URL to use for the deployment
 * @param channel Rancher channel to use (stable, latest, prime, prime-optimus, alpha, prime-optimus-alpha, head)
 * @param version Rancher version to install (latest, devel)
 * @param headVersion Rancher head version to install (2.7, 2.8, 2.9, 2.10, 2.11, 2.12, head)
 * @param ca CA to use (selfsigned, private)
 * @param proxy Define if a a proxy should be configured/used
 * @param extraFlags Optional helm flags for installing Rancher (start from extraEnv[2])
 * @returns Nothing or an error
 */
// NOTE: AddNode does not have unit test as it is not easy to mock
func DeployRancherManager(hostname, channel, version, headVersion, ca, proxy string, extraFlags ...[]string) error {
	var password = "rancherpassword"
	if envPW := os.Getenv("RANCHER_PASSWORD"); envPW != "" {
		password = envPW
	}

	channelName := "rancher-" + channel
	// For "head" channel, append headVersion to channelName
	if channel == "head" {
		// headVersion must be set otherwise use version as a fallback value
		if headVersion == "" {
			headVersion = version
		}
		channelName = channelName + "-" + headVersion
	}

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

	var chartRepo string
	switch channel {
	case "prime":
		chartRepo = "https://charts.rancher.com/server-charts/prime"
	case "prime-optimus":
		chartRepo = "https://charts.optimus.rancher.io/server-charts/latest"
	case "prime-optimus-alpha":
		chartRepo = "https://charts.optimus.rancher.io/server-charts/alpha"
	case "alpha":
		chartRepo = "https://releases.rancher.com/server-charts/alpha"
	case "latest":
		chartRepo = "https://releases.rancher.com/server-charts/latest"
	case "stable":
		chartRepo = "https://releases.rancher.com/server-charts/stable"
	case "head":
		chartRepo = "https://charts.optimus.rancher.io/server-charts/release-" + headVersion
	}

	// Add Helm repository
	if err := kubectl.RunHelmBinaryWithCustomErr("repo", "add", channelName, chartRepo); err != nil {
		return err
	}

	if err := kubectl.RunHelmBinaryWithCustomErr("repo", "update"); err != nil {
		return err
	}

	// Set specified version if needed
	if version != "" && version != "latest" && channel != "head" {
		if version == "devel" {
			flags = appendDevelFlags(flags, headVersion)
		} else if strings.Contains(version, "-rc") || strings.Contains(version, "-alpha") {
			flags = appendRCAlphaFlags(flags, version, channel)
		} else {
			flags = append(flags, "--version", version)
		}
	} else if channel == "head" && headVersion != "" {
		flags = appendHeadFlags(flags)
	}

	// For Private CA
	if ca == "private" {
		flags = append(flags,
			"--set", "ingress.tls.source=secret",
			"--set", "privateCA=true",
		)
	}

	// Use Rancher Manager behind proxy
	// Get the proxyHost if given
	proxyHost := "http://172.17.0.1:3128"
	if proxyHostFromEnv := os.Getenv("PROXY_HOST"); proxyHostFromEnv != "" {
		proxyHost = proxyHostFromEnv
	}
	if proxy == "rancher" {
		flags = append(flags,
			"--set", "proxy="+proxyHost,
			"--set", "noProxy=127.0.0.0/8\\,10.0.0.0/8\\,cattle-system.svc\\,172.16.0.0/12\\,192.168.0.0/16\\,.svc\\,.cluster.local",
		)
	}

	// Append extra flags if any
	for _, extra := range extraFlags {
		flags = append(flags, extra...)
	}

	return kubectl.RunHelmBinaryWithCustomErr(flags...)
}
