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

package operator

import (
	_ "embed"
	"encoding/json"

	"github.com/coreos/go-semver/semver"
	csvv1 "github.com/operator-framework/operator-lifecycle-manager/pkg/api/apis/operators/v1alpha1"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	cluster "kubevirt.io/kubevirt-migration-operator/pkg/resources/cluster"
	namespaced "kubevirt.io/kubevirt-migration-operator/pkg/resources/namespaced"
	utils "kubevirt.io/kubevirt-migration-operator/pkg/resources/utils"
)

const (
	serviceAccountName = "kubevirt-migration-operator"
	roleName           = "kubevirt-migration-operator"
	clusterRoleName    = roleName + "-cluster"
)

// FactoryArgs contains the required parameters to generate all cluster-scoped resources
type FactoryArgs struct {
	NamespacedArgs namespaced.FactoryArgs
	Image          string
}

func getClusterPolicyRules() []rbacv1.PolicyRule {
	rules := []rbacv1.PolicyRule{
		{
			APIGroups: []string{
				"rbac.authorization.k8s.io",
			},
			Resources: []string{
				"clusterrolebindings",
				"clusterroles",
			},
			Verbs: []string{
				"get",
				"list",
				"watch",
				"create",
				"update",
				"delete",
			},
		},
		{
			APIGroups: []string{
				"apiextensions.k8s.io",
			},
			Resources: []string{
				"customresourcedefinitions",
				"customresourcedefinitions/status",
			},
			Verbs: []string{
				"get",
				"list",
				"watch",
				"create",
				"update",
				"delete",
			},
		},
		{
			APIGroups: []string{
				"scheduling.k8s.io",
			},
			Resources: []string{
				"priorityclasses",
			},
			Verbs: []string{
				"get",
				"list",
				"watch",
			},
		},
	}
	rules = append(rules, cluster.GetClusterRolePolicyRules()...)
	return rules
}

// func createClusterRole() *rbacv1.ClusterRole {
// 	return utils.ResourceBuilder.CreateOperatorClusterRole(clusterRoleName, getClusterPolicyRules())
// }

// func createClusterRoleBinding(namespace string) *rbacv1.ClusterRoleBinding {
// 	return utils.ResourceBuilder.CreateOperatorClusterRoleBinding(serviceAccountName,
//  clusterRoleName, serviceAccountName, namespace)
// }

// func createClusterRBAC(args *FactoryArgs) []client.Object {
// 	return []client.Object{
// 		createClusterRole(),
// 		createClusterRoleBinding(args.NamespacedArgs.Namespace),
// 	}
// }

func getNamespacedPolicyRules() []rbacv1.PolicyRule {
	rules := []rbacv1.PolicyRule{
		{
			APIGroups: []string{
				"rbac.authorization.k8s.io",
			},
			Resources: []string{
				"rolebindings",
				"roles",
			},
			Verbs: []string{
				"get",
				"list",
				"watch",
				"create",
				"update",
				"delete",
			},
		},
		{
			APIGroups: []string{
				"",
			},
			Resources: []string{
				"serviceaccounts",
			},
			Verbs: []string{
				"list",
				"watch",
				"create",
				"update",
				"delete",
			},
		},
		{
			APIGroups: []string{
				"apps",
			},
			Resources: []string{
				"deployments",
			},
			Verbs: []string{
				"get",
				"list",
				"watch",
				"create",
				"update",
				"delete",
			},
		},
	}
	rules = append(rules, namespaced.GetRolePolicyRules()...)
	return rules
}

// func createServiceAccount(namespace string) *corev1.ServiceAccount {
// 	return utils.ResourceBuilder.CreateOperatorServiceAccount(serviceAccountName, namespace)
// }

// func createNamespacedRole(namespace string) *rbacv1.Role {
// 	role := utils.ResourceBuilder.CreateRole(roleName, getNamespacedPolicyRules())
// 	role.Namespace = namespace
// 	return role
// }

// func createNamespacedRoleBinding(namespace string) *rbacv1.RoleBinding {
// 	roleBinding := utils.ResourceBuilder.CreateRoleBinding(serviceAccountName, roleName, serviceAccountName, namespace)
// 	roleBinding.Namespace = namespace
// 	return roleBinding
// }

// func createNamespacedRBAC(args *FactoryArgs) []client.Object {
// 	return []client.Object{
// 		createServiceAccount(args.NamespacedArgs.Namespace),
// 		createNamespacedRole(args.NamespacedArgs.Namespace),
// 		createNamespacedRoleBinding(args.NamespacedArgs.Namespace),
// 	}
// }

// func createDeployment(args *FactoryArgs) []client.Object {
// 	return []client.Object{
// 		createOperatorDeployment(args.NamespacedArgs.OperatorVersion,
// 			args.NamespacedArgs.Namespace,
// 			args.NamespacedArgs.DeployClusterResources,
// 			args.Image,
// 			args.NamespacedArgs.ControllerImage,
// 			args.NamespacedArgs.Verbosity,
// 			args.NamespacedArgs.PullPolicy,
// 		),
// 	}
// }

func createOperatorEnvVar(operatorVersion,
	deployClusterResources,
	operatorImage,
	controllerImage,
	verbosity,
	pullPolicy string,
) []corev1.EnvVar {
	return []corev1.EnvVar{
		{
			Name:  "DEPLOY_CLUSTER_RESOURCES",
			Value: deployClusterResources,
		},
		{
			Name:  "OPERATOR_VERSION",
			Value: operatorVersion,
		},
		{
			Name:  "CONTROLLER_IMAGE",
			Value: controllerImage,
		},
		{
			Name:  "VERBOSITY",
			Value: verbosity,
		},
		{
			Name:  "PULL_POLICY",
			Value: pullPolicy,
		},
		{
			Name:  "MONITORING_NAMESPACE",
			Value: "",
		},
		{
			Name:  "OPERATOR_IMAGE",
			Value: operatorImage,
		},
	}
}

func createOperatorDeployment(operatorVersion, namespace, deployClusterResources, operatorImage, controllerImage,
	verbosity, pullPolicy string) *appsv1.Deployment {
	deployment := utils.CreateOperatorDeployment("kubevirt-migration-operator", namespace, "name",
		"kubevirt-migration-operator", serviceAccountName, int32(1))
	container := utils.CreatePortsContainer("operator", operatorImage, pullPolicy, createOperatorPorts())
	container.Resources = corev1.ResourceRequirements{
		Requests: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("100m"),
			corev1.ResourceMemory: resource.MustParse("150Mi"),
		},
	}
	container.LivenessProbe = &corev1.Probe{
		ProbeHandler: corev1.ProbeHandler{
			HTTPGet: &corev1.HTTPGetAction{
				Scheme: corev1.URISchemeHTTP,
				Port: intstr.IntOrString{
					Type:   intstr.Int,
					IntVal: 8081,
				},
				Path: "/healthz",
			},
		},
		InitialDelaySeconds: 15,
		TimeoutSeconds:      20,
	}
	container.ReadinessProbe = &corev1.Probe{
		ProbeHandler: corev1.ProbeHandler{
			HTTPGet: &corev1.HTTPGetAction{
				Scheme: corev1.URISchemeHTTP,
				Port: intstr.IntOrString{
					Type:   intstr.Int,
					IntVal: 8081,
				},
				Path: "/readyz",
			},
		},
		InitialDelaySeconds: 5,
		TimeoutSeconds:      10,
	}
	container.Env = createOperatorEnvVar(operatorVersion, deployClusterResources, operatorImage, controllerImage,
		verbosity, pullPolicy)
	deployment.Spec.Template.Spec.Containers = []corev1.Container{container}
	return deployment
}

func createOperatorPorts() []corev1.ContainerPort {
	return []corev1.ContainerPort{
		{
			Name:          "metrics",
			ContainerPort: 8443,
			Protocol:      "TCP",
		},
		{
			Name:          "health",
			ContainerPort: 8081,
			Protocol:      "TCP",
		},
	}
}

type csvPermissions struct {
	ServiceAccountName string              `json:"serviceAccountName"`
	Rules              []rbacv1.PolicyRule `json:"rules"`
}
type csvDeployments struct {
	Name string                `json:"name"`
	Spec appsv1.DeploymentSpec `json:"spec,omitempty"`
}

type csvStrategySpec struct {
	Permissions        []csvPermissions `json:"permissions"`
	ClusterPermissions []csvPermissions `json:"clusterPermissions"`
	Deployments        []csvDeployments `json:"deployments"`
}

//go:embed migControllerExample.json
var migControllerExample string

// nolint
func createClusterServiceVersion(data *ClusterServiceVersionData) (*csvv1.ClusterServiceVersion, error) {
	description := `
The Kubevirt Migration Controller is an extension that provides extra capabilities capitalizing on kubevirt VM migration methods.
`

	deployment := createOperatorDeployment(
		data.OperatorVersion,
		data.Namespace,
		"true",
		data.OperatorImage,
		data.ControllerImage,
		data.Verbosity,
		data.ImagePullPolicy,
	)

	deployment.Spec.Template.Spec.PriorityClassName = utils.PriorityClassDefault

	strategySpec := csvStrategySpec{
		Permissions: []csvPermissions{
			{
				ServiceAccountName: serviceAccountName,
				Rules:              getNamespacedPolicyRules(),
			},
		},
		ClusterPermissions: []csvPermissions{
			{
				ServiceAccountName: serviceAccountName,
				Rules:              getClusterPolicyRules(),
			},
		},
		Deployments: []csvDeployments{
			{
				Name: "kubevirt-migration-operator",
				Spec: deployment.Spec,
			},
		},
	}

	strategySpecJSONBytes, err := json.Marshal(strategySpec)
	if err != nil {
		return nil, err
	}

	return &csvv1.ClusterServiceVersion{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterServiceVersion",
			APIVersion: "operators.coreos.com/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "kubevirtmigrationoperator." + data.CsvVersion,
			Namespace: data.Namespace,
			Annotations: map[string]string{

				"capabilities": "Full Lifecycle",
				"categories":   "Storage,Virtualization",
				"alm-examples": migControllerExample,
				"description":  "Creates and maintains kubevirt migration controller deployments",
			},
		},

		Spec: csvv1.ClusterServiceVersionSpec{
			DisplayName: "KubeVirt Migration Controller",
			Description: description,
			Keywords:    []string{"Virtualization", "Storage"},
			Version:     *semver.New(data.CsvVersion),
			Maturity:    "alpha",
			Replaces:    data.ReplacesCsvVersion,
			Maintainers: []csvv1.Maintainer{{
				Name:  "KubeVirt project",
				Email: "kubevirt-dev@googlegroups.com",
			}},
			Provider: csvv1.AppLink{
				Name: "KubeVirt/Migration controller project",
			},
			Links: []csvv1.AppLink{
				{
					Name: "Migration Controller",
					URL:  "https://github.com/kubevirt/kubevirt-migration-controller/blob/main/README.md",
				},
				{
					Name: "Source Code",
					URL:  "https://github.com/kubevirt/kubevirt-migration-controller",
				},
			},
			Icon: []csvv1.Icon{{
				Data:      data.IconBase64,
				MediaType: "image/png",
			}},
			Labels: map[string]string{
				"alm-owner-kmc": "kubevirt-migration-operator",
				"operated-by":   "kubevirt-migration-operator",
			},
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"alm-owner-kmc": "kubevirt-migration-operator",
					"operated-by":   "kubevirt-migration-operator",
				},
			},
			InstallModes: []csvv1.InstallMode{
				{
					Type:      csvv1.InstallModeTypeOwnNamespace,
					Supported: true,
				},
				{
					Type:      csvv1.InstallModeTypeSingleNamespace,
					Supported: true,
				},
				{
					Type:      csvv1.InstallModeTypeMultiNamespace,
					Supported: true,
				},
				{
					Type:      csvv1.InstallModeTypeAllNamespaces,
					Supported: true,
				},
			},
			InstallStrategy: csvv1.NamedInstallStrategy{
				StrategyName:    "deployment",
				StrategySpecRaw: json.RawMessage(strategySpecJSONBytes),
			},
			CustomResourceDefinitions: csvv1.CustomResourceDefinitions{

				Owned: []csvv1.CRDDescription{
					{
						Name:        "migcontrollers.migrations.kubevirt.io",
						Version:     "v1alpha1",
						Kind:        "MigController",
						DisplayName: "KubeVirt Migration Controller deployment",
						Description: "Represents a kubevirt migration controller deployment",
						SpecDescriptors: []csvv1.SpecDescriptor{

							{
								Description:  "The ImageRegistry to use for the kubevirt migration controller components.",
								DisplayName:  "ImageRegistry",
								Path:         "imageRegistry",
								XDescriptors: []string{"urn:alm:descriptor:text"},
							},
							{
								Description:  "The ImageTag to use for the kubevirt migration controller components.",
								DisplayName:  "ImageTag",
								Path:         "imageTag",
								XDescriptors: []string{"urn:alm:descriptor:text"},
							},
							{
								Description:  "The ImagePullPolicy to use for the kubevirt migration controller components.",
								DisplayName:  "ImagePullPolicy",
								Path:         "imagePullPolicy",
								XDescriptors: []string{"urn:alm:descriptor:io.kubernetes:imagePullPolicy"},
							},
						},
						StatusDescriptors: []csvv1.StatusDescriptor{
							{
								Description:  "The deployment phase.",
								DisplayName:  "Phase",
								Path:         "phase",
								XDescriptors: []string{"urn:alm:descriptor:io.kubernetes.phase"},
							},
							{
								Description:  "Explanation for the current status of the kubevirt migration controller deployment.",
								DisplayName:  "Conditions",
								Path:         "conditions",
								XDescriptors: []string{"urn:alm:descriptor:io.kubernetes.conditions"},
							},
							{
								Description:  "The observed version of the kubevirt migration controller deployment.",
								DisplayName:  "Observed kubevirt migration controller Version",
								Path:         "observedVersion",
								XDescriptors: []string{"urn:alm:descriptor:text"},
							},
							{
								Description:  "The targeted version of the kubevirt migration controller deployment.",
								DisplayName:  "Target kubevirt migration controller Version",
								Path:         "targetVersion",
								XDescriptors: []string{"urn:alm:descriptor:text"},
							},
							{
								Description:  "The version of the kubevirt migration controller Operator",
								DisplayName:  "kubevirt migration controller Operator Version",
								Path:         "operatorVersion",
								XDescriptors: []string{"urn:alm:descriptor:text"},
							},
						},
					},
				},
			},
		},
	}, nil
}
