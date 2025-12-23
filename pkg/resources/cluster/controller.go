/*
Copyright 2025 The KubeVirt Authors.

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

package cluster

import (
	"strings"

	rbacv1 "k8s.io/api/rbac/v1"
	extv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	k8syaml "k8s.io/apimachinery/pkg/util/yaml"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"kubevirt.io/kubevirt-migration-operator/pkg/common"
	"kubevirt.io/kubevirt-migration-operator/pkg/resources"
	utils "kubevirt.io/kubevirt-migration-operator/pkg/resources/utils"
)

func createControllerResources(args *FactoryArgs) []client.Object {
	return []client.Object{
		createControllerClusterRole(),
		createControllerClusterRoleBinding(args.Namespace),
	}
}

func createControllerClusterRoleBinding(namespace string) *rbacv1.ClusterRoleBinding {
	return utils.ResourceBuilder.CreateClusterRoleBinding(
		common.ControllerServiceAccountName,
		common.ControllerResourceName,
		common.ControllerServiceAccountName,
		namespace,
	)
}

func getControllerClusterPolicyRules() []rbacv1.PolicyRule {
	// TODO: Figure out a way to read the RBAC rules from the controller config/rbac/role.yaml
	// file in the kubevirt-migration-controller project.
	return []rbacv1.PolicyRule{
		{
			APIGroups: []string{
				"",
			},
			Resources: []string{
				"events",
			},
			Verbs: []string{
				"create",
				"patch",
			},
		},
		{
			APIGroups: []string{
				"",
			},
			Resources: []string{
				"persistentvolumeclaims",
			},
			Verbs: []string{
				"list",
				"watch",
				"update",
			},
		},
		{
			APIGroups: []string{
				"",
			},
			Resources: []string{
				"namespaces",
			},
			Verbs: []string{
				"list",
				"watch",
			},
		},
		{
			APIGroups: []string{
				"",
			},
			Resources: []string{
				"pods",
			},
			Verbs: []string{
				"list",
				"watch",
				"delete",
			},
		},
		{
			APIGroups: []string{
				"kubevirt.io",
			},
			Resources: []string{
				"kubevirts",
			},
			Verbs: []string{
				"list",
				"watch",
			},
		},
		{
			APIGroups: []string{
				"kubevirt.io",
			},
			Resources: []string{
				"virtualmachineinstances",
			},
			Verbs: []string{
				"list",
				"watch",
				"get",
			},
		},
		{
			APIGroups: []string{
				"kubevirt.io",
			},
			Resources: []string{
				"virtualmachines",
			},
			Verbs: []string{
				"list",
				"watch",
				"get",
				"patch",
			},
		},
		{
			APIGroups: []string{
				"kubevirt.io",
			},
			Resources: []string{
				"virtualmachineinstancemigrations",
			},
			Verbs: []string{
				"list",
				"watch",
				"get",
				"delete",
			},
		},
		{
			APIGroups: []string{
				"cdi.kubevirt.io",
			},
			Resources: []string{
				"datavolumes",
			},
			Verbs: []string{
				"list",
				"watch",
				"get",
				"create",
				"update",
				"patch",
				"delete",
			},
		},
		{
			APIGroups: []string{
				"migrations.kubevirt.io",
			},
			Resources: []string{
				"virtualmachinestoragemigrations",
				"virtualmachinestoragemigrationplans",
				"multinamespacevirtualmachinestoragemigrations",
				"multinamespacevirtualmachinestoragemigrationplans",
			},
			Verbs: []string{
				"list",
				"watch",
				"get",
				"patch",
				"update",
				"create",
				"delete",
			},
		},
		{
			APIGroups: []string{
				"migrations.kubevirt.io",
			},
			Resources: []string{
				"virtualmachinestoragemigrations/status",
				"virtualmachinestoragemigrationplans/status",
				"multinamespacevirtualmachinestoragemigrations/status",
				"multinamespacevirtualmachinestoragemigrationplans/status",
			},
			Verbs: []string{
				"get",
				"patch",
				"update",
			},
		},
		{
			APIGroups: []string{
				"storage.k8s.io",
			},
			Resources: []string{
				"storageclasses",
			},
			Verbs: []string{
				"list",
				"watch",
			},
		},
	}
}

func createControllerClusterRole() *rbacv1.ClusterRole {
	return utils.ResourceBuilder.CreateClusterRole(common.ControllerResourceName, getControllerClusterPolicyRules())
}

// createVirtualMachineStorageMigrationCRD creates the migmigration schema
func createVirtualMachineStorageMigrationCRD() *extv1.CustomResourceDefinition {
	crd := extv1.CustomResourceDefinition{}
	_ = k8syaml.NewYAMLToJSONDecoder(strings.NewReader(resources.MigrationControllerCRDs["virtualmachinestoragemigration"])).Decode(&crd) //nolint
	return &crd
}

// createVirtualMachineStorageMigrationPlanCRD creates the migplan schema
func createVirtualMachineStorageMigrationPlanCRD() *extv1.CustomResourceDefinition {
	crd := extv1.CustomResourceDefinition{}
	_ = k8syaml.NewYAMLToJSONDecoder(strings.NewReader(resources.MigrationControllerCRDs["virtualmachinestoragemigrationplan"])).Decode(&crd) //nolint
	return &crd
}

// createVirtualMachineStorageMigrationCRD creates the migmigration schema
func createMultinamespaceVirtualMachineStorageMigrationCRD() *extv1.CustomResourceDefinition {
	crd := extv1.CustomResourceDefinition{}
	_ = k8syaml.NewYAMLToJSONDecoder(strings.NewReader(resources.MigrationControllerCRDs["multinamespacevirtualmachinestoragemigration"])).Decode(&crd) //nolint
	return &crd
}

// createMultinamespaceVirtualMachineStorageMigrationPlanCRD creates the migplan schema
func createMultinamespaceVirtualMachineStorageMigrationPlanCRD() *extv1.CustomResourceDefinition {
	crd := extv1.CustomResourceDefinition{}
	_ = k8syaml.NewYAMLToJSONDecoder(strings.NewReader(resources.MigrationControllerCRDs["multinamespacevirtualmachinestoragemigrationplan"])).Decode(&crd) //nolint
	return &crd
}
