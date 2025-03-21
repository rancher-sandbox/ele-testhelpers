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

// Cluster is the definition of a K8s cluster
type Cluster struct {
	APIVersion string        `yaml:"apiVersion"`
	Kind       string        `yaml:"kind,omitempty"`
	Metadata   Metadata      `yaml:"metadata"`
	Spec       ClusterSpec   `yaml:"spec"`
	Status     ClusterStatus `yaml:"status,omitempty"`
}

// Metadata is metadata attached to any object
type Metadata struct {
	Annotations     interface{} `yaml:"annotations"`
	Labels          interface{} `yaml:"labels,omitempty"`
	Finalizers      interface{} `yaml:"finalizers,omitempty"`
	ManagedFields   interface{} `yaml:"managedFields,omitempty"`
	Name            string      `yaml:"name"`
	Namespace       string      `yaml:"namespace"`
	ResourceVersion string      `yaml:"resourceVersion"`
	UID             string      `yaml:"uid"`
}

// ClusterSpec is a description of a cluster
type ClusterSpec struct {
	KubernetesVersion        string      `yaml:"kubernetesVersion"`
	LocalClusterAuthEndpoint interface{} `yaml:"localClusterAuthEndpoint"`
	RkeConfig                RKEConfig   `yaml:"rkeConfig"`
}

// RKEConfig has all RKE/K3s cluster information
type RKEConfig struct {
	Etcd                  interface{}            `yaml:"etcd,omitempty"`
	ChartValues           map[string]interface{} `yaml:"chartValues,omitempty" wrangler:"nullable"`
	MachineGlobalConfig   interface{}            `yaml:"machineGlobalConfig"`
	MachinePools          []MachinePools         `yaml:"machinePools"`
	MachineSelectorConfig interface{}            `yaml:"machineSelectorConfig"`
	UpgradeStrategy       interface{}            `yaml:"upgradeStrategy,omitempty"`
	Registries            interface{}            `yaml:"registries"`
}

// MachinePools has all pools information
type MachinePools struct {
	ControlPlaneRole     bool             `yaml:"controlPlaneRole,omitempty"`
	DrainBeforeDelete    bool             `yaml:"drainBeforeDelete,omitempty"`
	EtcdRole             bool             `yaml:"etcdRole,omitempty"`
	MachineConfigRef     MachineConfigRef `yaml:"machineConfigRef"`
	Name                 string           `yaml:"name"`
	Quantity             int              `yaml:"quantity"`
	UnhealthyNodeTimeout string           `yaml:"unhealthyNodeTimeout"`
	WorkerRole           bool             `yaml:"workerRole,omitempty"`
}

// MachineConfigRef makes the link between the cluster, pool and the Elemental nodes
type MachineConfigRef struct {
	APIVersion string `yaml:"apiVersion"`
	Kind       string `yaml:"kind"`
	Name       string `yaml:"name"`
}

// ClusterStatus has all the cluster status information
type ClusterStatus struct {
	AgentDeployed    bool               `yaml:"agentDeployed,omitempty"`
	ClientSecretName string             `yaml:"clientSecretName"`
	ClusterName      string             `yaml:"clusterName"`
	Conditions       []ClusterCondition `yaml:"conditions,omitempty"`
	Ready            bool               `yaml:"ready,omitempty"`
}

// ClusterCondition is the cluster condition status
type ClusterCondition struct {
	LastUpdateTime string `yaml:"lastUpdateTime"`
	Message        string `yaml:"message,omitempty"`
	Reason         string `yaml:"reason,omitempty"`
	Status         string `yaml:"status"`
	Type           string `yaml:"type"`
}
