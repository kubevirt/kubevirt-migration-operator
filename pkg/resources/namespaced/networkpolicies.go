/*
Copyright The KubeVirt Authors.

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
	corev1 "k8s.io/api/core/v1"
	networkv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/ptr"
	"kubevirt.io/kubevirt-migration-operator/pkg/common"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	allowIngressToMetrics = "migration-operator-allow-ingress-to-metrics"
)

func createNetworkPolicies(args *FactoryArgs) []client.Object {
	return []client.Object{
		newIngressToMetricsNP(args.Namespace),
	}
}

func newNetworkPolicy(namespace, name string, spec *networkv1.NetworkPolicySpec) *networkv1.NetworkPolicy {
	return &networkv1.NetworkPolicy{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "networking.k8s.io/v1",
			Kind:       "NetworkPolicy",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: *spec,
	}
}

func newIngressToMetricsNP(namespace string) *networkv1.NetworkPolicy {
	return newNetworkPolicy(
		namespace,
		allowIngressToMetrics,
		&networkv1.NetworkPolicySpec{
			PodSelector: metav1.LabelSelector{
				MatchLabels: map[string]string{common.PrometheusLabelKey: common.PrometheusLabelValue},
			},
			PolicyTypes: []networkv1.PolicyType{networkv1.PolicyTypeIngress},
			Ingress: []networkv1.NetworkPolicyIngressRule{
				{
					Ports: []networkv1.NetworkPolicyPort{
						{
							Port:     ptr.To(intstr.FromInt32(8443)),
							Protocol: ptr.To(corev1.ProtocolTCP),
						},
					},
				},
			},
		},
	)
}
