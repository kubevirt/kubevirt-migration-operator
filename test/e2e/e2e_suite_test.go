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

package e2e

import (
	"flag"
	"fmt"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"kubevirt.io/kubevirt-migration-operator/test/utils"
)

var (
	kubectlPath                = flag.String("kubectl-path", "kubectl", "Path to the kubectl binary")
	kubeConfig                 = flag.String("test-kubeconfig", "", "Path to the kubeconfig file")
	migrationOperatorNamespace = flag.String("migration-operator-namespace", "kubevirt",
		"Namespace of the migration operator")
	kubeURL = flag.String("kubeurl", "", "URL of the kube API server")
	kcs     *kubernetes.Clientset
)

// TestE2E runs the end-to-end (e2e) test suite for the project. These tests execute in an isolated,
// temporary environment to validate project changes with the purposed to be used in CI jobs.
// The default setup requires Kind, builds/loads the Manager Docker image locally, and installs
// CertManager.
func TestE2E(t *testing.T) {
	RegisterFailHandler(Fail)
	_, _ = fmt.Fprintf(GinkgoWriter, "Starting kubevirt-migration-operator integration test suite\n")
	BuildTestSuite()
	RunSpecs(t, "e2e suite")
}

func BuildTestSuite() {
	SynchronizedBeforeSuite(func() {}, func() {
		By("checking if cert manager is installed")
		isCertManagerAlreadyInstalled := utils.IsCertManagerCRDsInstalled(kubectlPath)
		Expect(isCertManagerAlreadyInstalled).To(BeTrue(), "CertManager is not installed")

		_, err := fmt.Fprintf(GinkgoWriter, "Reading parameters\n")
		Expect(err).ToNot(HaveOccurred(), "Unable to write to GinkgoWriter")

		_, err = fmt.Fprintf(GinkgoWriter, "Kubectl path: %s\n", *kubectlPath)
		Expect(err).ToNot(HaveOccurred(), "Unable to write to GinkgoWriter")

		_, err = fmt.Fprintf(GinkgoWriter, "Kubeconfig: %s\n", *kubeConfig)
		Expect(err).ToNot(HaveOccurred(), "Unable to write to GinkgoWriter")

		_, err = fmt.Fprintf(GinkgoWriter, "KubeURL: %s\n", *kubeURL)
		Expect(err).ToNot(HaveOccurred(), "Unable to write to GinkgoWriter")

		_, err = fmt.Fprintf(GinkgoWriter, "Migration controller namespace: %s\n", *migrationOperatorNamespace)
		Expect(err).ToNot(HaveOccurred(), "Unable to write to GinkgoWriter")
		restConfig, err := LoadConfig()
		Expect(err).ToNot(HaveOccurred(), "Unable to load RestConfig")
		// clients
		kcs, err = GetKubeClientFromRESTConfig(restConfig)
		Expect(err).ToNot(HaveOccurred(), "Unable to create K8SClient")
		Expect(kcs).ToNot(BeNil(), "K8SClient is nil")

	})
}

// LoadConfig loads our specified kubeconfig
func LoadConfig() (*rest.Config, error) {
	return clientcmd.BuildConfigFromFlags(*kubeURL, *kubeConfig)
}

// GetKubeClientFromRESTConfig provides a function to get a K8s client using hte REST config
func GetKubeClientFromRESTConfig(config *rest.Config) (*kubernetes.Clientset, error) {
	config.APIPath = "/apis"
	config.ContentType = runtime.ContentTypeJSON
	return kubernetes.NewForConfig(config)
}
