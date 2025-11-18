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
				"pods",
				"persistentvolumes",
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
				"virtualmachines",
			},
			Verbs: []string{
				"list",
				"watch",
				"get",
			},
		},
		{
			APIGroups: []string{
				"migrations.kubevirt.io",
			},
			Resources: []string{
				"migclusters",
				"migmigrations",
				"migplans",
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
				"migclusters/finalizers",
				"migmigrations/finalizers",
				"migplans/finalizers",
			},
			Verbs: []string{
				"update",
			},
		},
		{
			APIGroups: []string{
				"migrations.kubevirt.io",
			},
			Resources: []string{
				"migclusters/status",
				"migmigrations/status",
				"migplans/status",
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

// createMigMigrationCRD creates the migmigration schema
func createMigMigrationCRD() *extv1.CustomResourceDefinition {
	crd := extv1.CustomResourceDefinition{}
	_ = k8syaml.NewYAMLToJSONDecoder(strings.NewReader(resources.MigrationControllerCRDs["migmigration"])).Decode(&crd)
	return &crd
}

// createMigPlanCRD creates the migplan schema
func createMigPlanCRD() *extv1.CustomResourceDefinition {
	crd := extv1.CustomResourceDefinition{}
	_ = k8syaml.NewYAMLToJSONDecoder(strings.NewReader(resources.MigrationControllerCRDs["migplan"])).Decode(&crd)
	return &crd
}
