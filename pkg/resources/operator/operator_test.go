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
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestOperator(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Operator Suite")
}

var _ = Describe("Operator Deployment", func() {
	It("should have security context set properly", func() {
		deployment := createOperatorDeployment("0.0.1", "test-ns", "true",
			"operator-image", "controller-image", "2", "Always")
		Expect(deployment.Spec.Template.Spec.Containers).To(HaveLen(1))
		container := deployment.Spec.Template.Spec.Containers[0]
		Expect(container.SecurityContext).NotTo(BeNil())
		Expect(container.SecurityContext.ReadOnlyRootFilesystem).NotTo(BeNil())
		Expect(*container.SecurityContext.ReadOnlyRootFilesystem).To(BeTrue())
		Expect(container.SecurityContext.AllowPrivilegeEscalation).NotTo(BeNil())
		Expect(*container.SecurityContext.AllowPrivilegeEscalation).To(BeFalse())
		Expect(container.SecurityContext.RunAsNonRoot).NotTo(BeNil())
		Expect(*container.SecurityContext.RunAsNonRoot).To(BeTrue())
	})
})
