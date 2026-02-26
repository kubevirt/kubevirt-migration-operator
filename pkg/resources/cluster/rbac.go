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
	rbacv1 "k8s.io/api/rbac/v1"

	"sigs.k8s.io/controller-runtime/pkg/client"

	utils "kubevirt.io/kubevirt-migration-operator/pkg/resources/utils"
)

func createAggregateClusterRoles(_ *FactoryArgs) []client.Object {
	return []client.Object{
		utils.ResourceBuilder.CreateAggregateClusterRole("migrations.kubevirt.io:admin", "admin", getAdminPolicyRules()),
		utils.ResourceBuilder.CreateAggregateClusterRole("migrations.kubevirt.io:edit", "edit", getEditPolicyRules()),
		utils.ResourceBuilder.CreateAggregateClusterRole("migrations.kubevirt.io:view", "view", getViewPolicyRules()),
	}
}

func getAdminPolicyRules() []rbacv1.PolicyRule {
	// leaving this here so we have optionality in the future for other resources
	// currently we follow the kubevirt model where only admin can migrate
	return []rbacv1.PolicyRule{}
}

func getEditPolicyRules() []rbacv1.PolicyRule {
	// diff between admin and edit ClusterRoles is minimal
	return getAdminPolicyRules()
}

func getViewPolicyRules() []rbacv1.PolicyRule {
	return []rbacv1.PolicyRule{
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
				"get",
				"list",
				"watch",
			},
		},
	}
}
