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

package namespaced

import (
	secv1 "github.com/openshift/api/security/v1"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	sdkapi "kubevirt.io/controller-lifecycle-operator-sdk/api"
	"kubevirt.io/kubevirt-migration-operator/pkg/common"
	utils "kubevirt.io/kubevirt-migration-operator/pkg/resources/utils"
)

func createControllerResources(args *FactoryArgs) []client.Object {
	return []client.Object{
		createControllerServiceAccount(),
		createControllerRoleBinding(),
		createControllerRole(),
		createControllerDeployment(
			args.ControllerImage,
			args.Verbosity,
			args.PullPolicy,
			args.PriorityClassName,
			args.InfraNodePlacement),
		createPrometheusService(),
	}
}

func createControllerRoleBinding() *rbacv1.RoleBinding {
	return utils.ResourceBuilder.CreateRoleBinding(
		common.ControllerResourceName,
		common.ControllerResourceName,
		common.ControllerServiceAccountName,
		"",
	)
}

func getControllerNamespacedRules() []rbacv1.PolicyRule {
	return []rbacv1.PolicyRule{
		{
			APIGroups: []string{
				"",
			},
			Resources: []string{
				"configmaps",
			},
			ResourceNames: []string{
				"migration-controller",
			},
			Verbs: []string{
				"get",
			},
		},
		{
			APIGroups: []string{
				"coordination.k8s.io",
			},
			Resources: []string{
				"leases",
			},
			Verbs: []string{
				"get",
				"list",
				"watch",
				"create",
				"update",
				"delete",
				"patch",
			},
		},
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
	}
}

func createControllerRole() *rbacv1.Role {
	return utils.ResourceBuilder.CreateRole(common.ControllerResourceName, getControllerNamespacedRules())
}

func createControllerServiceAccount() *corev1.ServiceAccount {
	return utils.ResourceBuilder.CreateServiceAccount(common.ControllerServiceAccountName)
}

func createControllerDeployment(controllerImage, verbosity, pullPolicy, priorityClassName string,
	infraNodePlacement *sdkapi.NodePlacement) *appsv1.Deployment {
	// The match selector is immutable. that's why we should always use the same labels.
	deployment := utils.CreateDeployment(common.ControllerResourceName,
		common.ComponentLabel,
		common.ControllerResourceName,
		common.ControllerServiceAccountName,
		int32(1),
		infraNodePlacement,
	)
	deployment.ObjectMeta.Labels[common.ComponentLabel] = common.ControllerResourceName
	if priorityClassName != "" {
		deployment.Spec.Template.Spec.PriorityClassName = priorityClassName
	}
	container := utils.CreateContainer(common.ControllerResourceName, controllerImage, verbosity, pullPolicy)
	container.Ports = []corev1.ContainerPort{
		{
			Name:          "metrics",
			ContainerPort: 8443,
			Protocol:      "TCP",
		},
	}
	container.Args = append(container.Args, "--leader-elect", "--health-probe-bind-address=:8081")
	labels := mergeLabels(deployment.Spec.Template.GetLabels(), map[string]string{
		common.PrometheusLabelKey: common.PrometheusLabelValue,
	})
	// Add label for pod affinity
	deployment.SetLabels(labels)
	deployment.Spec.Template.SetLabels(labels)
	if deployment.Spec.Template.Annotations == nil {
		deployment.Spec.Template.Annotations = make(map[string]string)
	}
	deployment.Spec.Template.Annotations[secv1.RequiredSCCAnnotation] = common.RestrictedSCCName
	container.Env = []corev1.EnvVar{
		{
			Name: "CONTROLLER_NAMESPACE",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					APIVersion: "v1",
					FieldPath:  "metadata.namespace",
				},
			},
		},
	}
	container.ReadinessProbe = &corev1.Probe{
		ProbeHandler: corev1.ProbeHandler{
			HTTPGet: &corev1.HTTPGetAction{
				Scheme: corev1.URISchemeHTTP,
				Path:   "/readyz",
				Port: intstr.IntOrString{
					Type:   intstr.Int,
					IntVal: 8081,
				},
			},
		},
		TimeoutSeconds:      10,
		FailureThreshold:    3,
		SuccessThreshold:    1,
		InitialDelaySeconds: 5,
		PeriodSeconds:       10,
	}
	container.LivenessProbe = &corev1.Probe{
		ProbeHandler: corev1.ProbeHandler{
			HTTPGet: &corev1.HTTPGetAction{
				Path:   "/healthz",
				Scheme: corev1.URISchemeHTTP,
				Port: intstr.IntOrString{
					Type:   intstr.Int,
					IntVal: 8081,
				},
			},
		},
		InitialDelaySeconds: 15,
		PeriodSeconds:       20,
		TimeoutSeconds:      10,
		FailureThreshold:    3,
		SuccessThreshold:    1,
	}
	container.Resources = corev1.ResourceRequirements{
		Requests: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("100m"),
			corev1.ResourceMemory: resource.MustParse("64Mi"),
		},
		Limits: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("500m"),
			corev1.ResourceMemory: resource.MustParse("128Mi"),
		},
	}

	deployment.Spec.Template.Spec.Containers = []corev1.Container{container}
	return deployment
}

// mergeLabels copies source labels to destination (overwrites existing labels)
func mergeLabels(src, dest map[string]string) map[string]string {
	if dest == nil {
		dest = map[string]string{}
	}

	for k, v := range src {
		dest[k] = v
	}

	return dest
}

func createPrometheusService() *corev1.Service {
	service := utils.ResourceBuilder.CreateService(common.PrometheusServiceName,
		common.PrometheusLabelKey, common.PrometheusLabelValue, nil)
	service.Spec.Ports = []corev1.ServicePort{
		{
			Name: common.PrometheusServiceName,
			Port: 8443,
			TargetPort: intstr.IntOrString{
				Type:   intstr.String,
				StrVal: "metrics",
			},
			Protocol: corev1.ProtocolTCP,
		},
	}
	return service
}
