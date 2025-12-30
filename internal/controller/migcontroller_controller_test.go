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

package controller

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	record "k8s.io/client-go/tools/record"
	manager "sigs.k8s.io/controller-runtime/pkg/manager"

	"kubevirt.io/controller-lifecycle-operator-sdk/pkg/sdk/callbacks"
	sdkr "kubevirt.io/controller-lifecycle-operator-sdk/pkg/sdk/reconciler"
	migrationsv1alpha1 "kubevirt.io/kubevirt-migration-operator/api/v1alpha1"
	"kubevirt.io/kubevirt-migration-operator/pkg/resources/cluster"
	"kubevirt.io/kubevirt-migration-operator/pkg/resources/namespaced"
)

const (
	testNamespace = "kubevirt"
	resourceName  = "test-resource"
)

var _ = Describe("MigController Controller", func() {
	Context("When reconciling a resource", func() {
		var (
			controllerReconciler *MigControllerReconciler
		)
		ctx := context.Background()

		typeNamespacedName := types.NamespacedName{
			Name:      resourceName,
			Namespace: testNamespace,
		}
		migcontroller := &migrationsv1alpha1.MigController{}

		BeforeEach(func() {
			By("creating the custom resource for the Kind MigController")
			resource := &migrationsv1alpha1.MigController{
				ObjectMeta: metav1.ObjectMeta{
					Name:      resourceName,
					Namespace: testNamespace,
				},
				Spec: migrationsv1alpha1.MigControllerSpec{},
			}
			Expect(k8sClient.Create(ctx, resource)).To(Succeed())
			recorder := record.NewFakeRecorder(100)
			Expect(k8sClient.Get(ctx, typeNamespacedName, migcontroller)).To(Succeed())
			migcontroller.Status.ObservedVersion = "0.0.1"
			Expect(k8sClient.Status().Update(ctx, migcontroller)).To(Succeed())

			controllerReconciler = &MigControllerReconciler{
				Client: k8sClient,
				scheme: k8sClient.Scheme(),
				namespacedArgs: &namespaced.FactoryArgs{
					OperatorVersion: "0.0.1",
					Namespace:       testNamespace,
					ControllerImage: "kubevirt/kubevirt-migration-operator:latest",
				},
				clusterArgs: &cluster.FactoryArgs{
					Namespace: testNamespace,
					Client:    k8sClient,
					Logger:    log,
				},
			}

			mgr, err := manager.New(cfg, manager.Options{
				Scheme: k8sClient.Scheme(),
			})
			Expect(err).NotTo(HaveOccurred())
			callbackDispatcher := callbacks.NewCallbackDispatcher(log, k8sClient, k8sClient, k8sClient.Scheme(), testNamespace)
			controllerReconciler.reconciler = sdkr.NewReconciler(
				controllerReconciler, log, k8sClient,
				callbackDispatcher, k8sClient.Scheme(), mgr.GetCache,
				createVersionLabel, updateVersionLabel, LastAppliedConfigAnnotation,
				requeueInterval, finalizerName, true, recorder,
			).WithNamespacedCR()
			Expect(controllerReconciler.SetupWithManager(mgr)).To(Succeed())
		})

		AfterEach(func() {
			// TODO(user): Cleanup logic after each test, like removing the resource instance.
			resource := &migrationsv1alpha1.MigController{}
			err := k8sClient.Get(ctx, typeNamespacedName, resource)
			Expect(err).NotTo(HaveOccurred())

			By("Cleanup the specific resource instance MigController")
			Expect(k8sClient.Delete(ctx, resource)).To(Succeed())
		})

		It("should successfully reconcile the resource", func() {
			resource := &migrationsv1alpha1.MigController{}
			err := k8sClient.Get(ctx, typeNamespacedName, resource)
			Expect(err).NotTo(HaveOccurred())

			Expect(resource.Status.ObservedVersion).To(Equal("0.0.1"))
			By("Reconciling the created resource")
			_, err = controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())
			// TODO(user): Add more specific assertions depending on your controller's reconciliation logic.
			// Example: If you expect a certain status condition after reconciliation, verify it here.
		})

		It("should successfully create aggregated cluster role", func() {
			resource := &migrationsv1alpha1.MigController{}
			err := k8sClient.Get(ctx, typeNamespacedName, resource)
			Expect(err).NotTo(HaveOccurred())

			Expect(resource.Status.ObservedVersion).To(Equal("0.0.1"))
			By("Reconciling the created resource")
			_, err = controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())

			clusterRole := &rbacv1.ClusterRole{}
			err = k8sClient.Get(ctx, types.NamespacedName{Name: "migrations.kubevirt.io:view"}, clusterRole)
			Expect(err).NotTo(HaveOccurred())
			Expect(clusterRole.Rules).To(ContainElements(
				[]rbacv1.PolicyRule{
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
				},
			))
		})
	})
})
