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
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"reflect"

	"github.com/rancher-sandbox/ele-testhelpers/kubectl"
	"github.com/rancher-sandbox/ele-testhelpers/tools"
	"gopkg.in/yaml.v3"
)

const noExist = "'%s' does not exist!"

/**
 * Get cluster informations
 * @remarks Cluster informations are exported to *c
 * @param ns Namespace
 * @param name Cluster name
 * @returns Cluster informations in *c or an error
 */
func (c *Cluster) getCluster(ns, name string) error {
	out, err := kubectl.Run("get",
		"cluster.v1.provisioning.cattle.io",
		"--namespace", ns, name,
		"-o", "yaml")
	if err != nil {
		return err
	}

	// Decode content
	return yaml.Unmarshal([]byte(out), c)
}

/**
 * Set/update cluster configuration
 * @remarks Cluster informations are set/exported to *c
 * @param ns Namespace
 * @returns Cluster informations in *c or an error
 */
func (c *Cluster) setCluster(ns string) error {
	// Encode content
	out, err := yaml.Marshal(&c)
	if err != nil {
		return err
	}

	// Use temporary file
	f, err := os.CreateTemp("", "updatedCluster")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())

	if _, err := f.Write(out); err != nil {
		return err
	}

	if err := f.Close(); err != nil {
		return err
	}

	// Apply new cluster configuration
	return kubectl.Apply(ns, f.Name())
}

/**
 * Set/update cluster configuration
 * @remarks Cluster informations are set/exported to *c
 * @param ns Namespace
 * @param name Cluster name
 * @param pool Pool to increase
 * @param quantity Quantity to set
 * @returns Qantity set or an error
 */
func SetNodeQuantity(ns, name, pool string, quantity int) (int, error) {
	c := &Cluster{}
	quantitySet := 0
	poolFound := false

	// Get cluster configuration
	if err := c.getCluster(ns, name); err != nil {
		return 0, err
	}

	// Try to increase quantity field
	for i := range c.Spec.RkeConfig.MachinePools {
		// Only on selected pool
		if c.Spec.RkeConfig.MachinePools[i].Name == pool {
			// Pool found!
			poolFound = true

			// Increase quantity
			c.Spec.RkeConfig.MachinePools[i].Quantity += quantity
			quantitySet = c.Spec.RkeConfig.MachinePools[i].Quantity

			// Quantity increased, loop can be stopped
			break
		}
	}

	// Throw an error if the pool has not been found
	if !poolFound {
		return 0, errors.New("pool " + fmt.Sprintf(noExist, pool))
	}

	// Save and apply cluster configuration
	if err := c.setCluster(ns); err != nil {
		return 0, err
	}

	return quantitySet, nil
}

/**
 * Set/Unset cluster role
 * @remarks This function (un)set role in a cluster's pool
 * @param ns Namespace
 * @param name Cluster name
 * @param pool Pool to increase
 * @param role Role to set
 * @param value value to set (0 or 1)
 * @returns Qantity set or an error
 * @example err := misc.SetRole(clusterNS, clusterName, "pool-worker-"+clusterName, "ControlPlaneRole", true)
 */
func SetRole(ns, name, pool, role string, value bool) error {
	c := &Cluster{}
	poolFound := false

	// Get cluster configuration
	if err := c.getCluster(ns, name); err != nil {
		return err
	}

	// Try to set value to role
	for i := range c.Spec.RkeConfig.MachinePools {
		// Only on selected pool
		if c.Spec.RkeConfig.MachinePools[i].Name == pool {
			// Pool found!
			poolFound = true

			// Get fields list and check that the role exist
			v := reflect.ValueOf(&c.Spec.RkeConfig.MachinePools[i]).Elem()
			f := v.FieldByName(role)
			if f != (reflect.Value{}) {
				// Yes, set the value accordingly
				v.FieldByName(role).SetBool(value)

				// Role toggled, loop can be stopped
				break
			}

			// No, return an error
			return errors.New("role " + fmt.Sprintf(noExist, role))
		}
	}

	// Throw an error if the pool has not been found
	if !poolFound {
		return errors.New("pool " + fmt.Sprintf(noExist, pool))
	}

	// Save and apply cluster configuration
	return c.setCluster(ns)
}

/**
 * Check daemonset status
 * @remarks This function checks the status of a daemonset based on checkList
 * @param k Already defined Kubectl context
 * @param checkList Array of paired namespaces/labels to check
 * @example CheckDaemonSet(k, [][]string{{"kube-system", "k8s-app=canal"}})
 * @returns Nothing or an error
 */
func CheckDaemonSet(k *kubectl.Kubectl, checkList [][]string) error {
	for _, check := range checkList {
		if err := k.WaitForDaemonSet(check[0], check[1]); err != nil {
			return err
		}
	}
	return nil
}

/**
 * Check pod status
 * @remarks This function checks the status of a pod based on checkList
 * @param k Already defined Kubectl context
 * @param checkList Array of paired namespaces/labels to check
 * @example CheckPod(k, [][]string{{"cattle-elemental-system", "app=elemental-operator"}})
 * @returns Nothing or an error
 */
func CheckPod(k *kubectl.Kubectl, checkList [][]string) error {
	for _, check := range checkList {
		if err := k.WaitForNamespaceWithPod(check[0], check[1]); err != nil {
			return err
		}
	}

	return nil
}

/**
 * Set kubeconfig
 * @remarks This function sets KUBECONFIG env variable to access client cluster
 * @param ns Namesapce
 * @param name Cluster name
 * @returns Filename of created client kubeconfig
 */
func SetClientKubeConfig(ns, name string) (string, error) {
	// Use our internal CreateTemp function!
	kubeConfig, err := tools.CreateTemp("clientKubeConfig")
	if err != nil {
		return "", err
	}

	// Get Kubeconfig of client cluster
	out, err := kubectl.Run("get", "secret",
		"--namespace", ns,
		name+"-kubeconfig", "-o", "jsonpath={.data.value}")
	if err != nil {
		os.Remove(kubeConfig)
		return "", err
	}

	// Decode Kubeconfig data and write into file
	data, err := base64.StdEncoding.DecodeString(out)
	if err != nil {
		os.Remove(kubeConfig)
		return "", err
	}
	err = tools.WriteFile(kubeConfig, data)
	if err != nil {
		os.Remove(kubeConfig)
		return "", err
	}

	// Export KUBECONFIG envar
	err = os.Setenv("KUBECONFIG", kubeConfig)
	if err != nil {
		os.Remove(kubeConfig)
		return "", err
	}

	return kubeConfig, nil
}
