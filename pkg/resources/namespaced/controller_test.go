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
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	sdkapi "kubevirt.io/controller-lifecycle-operator-sdk/api"
)

const (
	testKubevirtCAVolumeName = "kubevirt-ca-configmap"
)

func TestNamespaced(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Namespaced Resources Suite")
}

var _ = Describe("Controller Deployment", func() {
	DescribeTable("should match controller configuration",
		func(priorityClass string, nodePlacement *sdkapi.NodePlacement) {
			deployment := createControllerDeployment(
				"test-image:latest",
				"2",
				"IfNotPresent",
				priorityClass,
				nodePlacement,
			)

			Expect(deployment).NotTo(BeNil())
			Expect(deployment.Spec.Template.Spec.Volumes).NotTo(BeEmpty())
			Expect(deployment.Spec.Template.Spec.Containers).To(HaveLen(1))

			container := deployment.Spec.Template.Spec.Containers[0]

			// Verify kubevirt-ca volume
			var caVolume *corev1.Volume
			for i, volume := range deployment.Spec.Template.Spec.Volumes {
				if volume.Name == testKubevirtCAVolumeName {
					caVolume = &deployment.Spec.Template.Spec.Volumes[i]
					break
				}
			}
			Expect(caVolume).NotTo(BeNil(), "kubevirt-ca-configmap volume should be present")
			Expect(caVolume.VolumeSource.ConfigMap).NotTo(BeNil(), "volume should be backed by a ConfigMap")
			Expect(caVolume.VolumeSource.ConfigMap.Name).To(Equal("kubevirt-ca"), "ConfigMap name should be kubevirt-ca")

			// Verify volumeMount on container
			Expect(container.VolumeMounts).To(HaveLen(1))
			Expect(container.VolumeMounts[0].Name).To(Equal(testKubevirtCAVolumeName))
			Expect(container.VolumeMounts[0].MountPath).To(Equal("/etc/ssl/certs/kubevirt-ca"))
			Expect(container.VolumeMounts[0].ReadOnly).To(BeTrue())

			// Verify resource requests
			Expect(container.Resources.Requests).NotTo(BeNil())
			Expect(container.Resources.Requests.Cpu().String()).To(Equal("100m"))
			Expect(container.Resources.Requests.Memory().String()).To(Equal("150Mi"))

			// Verify terminationGracePeriodSeconds
			Expect(deployment.Spec.Template.Spec.TerminationGracePeriodSeconds).NotTo(BeNil())
			Expect(*deployment.Spec.Template.Spec.TerminationGracePeriodSeconds).To(Equal(int64(10)))
		},
		Entry("with default configuration", "", &sdkapi.NodePlacement{}),
		Entry("with priority class", "system-cluster-critical", &sdkapi.NodePlacement{}),
		Entry("with custom node placement", "", &sdkapi.NodePlacement{
			NodeSelector: map[string]string{
				"custom-label": "custom-value",
			},
		}),
		Entry("with both priority class and node placement", "system-cluster-critical", &sdkapi.NodePlacement{
			NodeSelector: map[string]string{
				"custom-label": "custom-value",
			},
		}),
	)

	Context("kubevirtCAVolume function", func() {
		It("should return a properly configured volume", func() {
			volume := kubevirtCAVolume()

			Expect(volume.Name).To(Equal(testKubevirtCAVolumeName))
			Expect(volume.VolumeSource.ConfigMap).NotTo(BeNil())
			Expect(volume.VolumeSource.ConfigMap.LocalObjectReference.Name).To(Equal("kubevirt-ca"))
		})
	})
})
