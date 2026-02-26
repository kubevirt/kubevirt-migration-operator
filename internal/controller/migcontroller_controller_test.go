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
	"fmt"
	"reflect"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	record "k8s.io/client-go/tools/record"

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
		defaultResult := reconcile.Result{RequeueAfter: requeueInterval}

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
				namespace: testNamespace,
				Client:    k8sClient,
				scheme:    k8sClient.Scheme(),
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

			callbackDispatcher := callbacks.NewCallbackDispatcher(log, k8sClient, k8sClient, k8sClient.Scheme(), testNamespace)
			noOpGetCache := func() cache.Cache {
				// no watching, so no cache
				return nil
			}
			controllerReconciler.reconciler = sdkr.NewReconciler(
				controllerReconciler, log, k8sClient,
				callbackDispatcher, k8sClient.Scheme(), noOpGetCache,
				createVersionLabel, updateVersionLabel, LastAppliedConfigAnnotation,
				requeueInterval, finalizerName, true, recorder,
			).
				WithNamespacedCR().
				WithWatching(true)
			controllerReconciler.registerHooks()
		})

		AfterEach(func() {
			// TODO(user): Cleanup logic after each test, like removing the resource instance.
			resource := &migrationsv1alpha1.MigController{}
			err := k8sClient.Get(ctx, typeNamespacedName, resource)
			Expect(err).NotTo(HaveOccurred())

			By("Cleanup the specific resource instance MigController")
			Expect(k8sClient.Delete(ctx, resource)).To(Succeed())
			By("Reconciling the deletion")
			res, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(res).To(Equal(reconcile.Result{}))
			err = k8sClient.Get(ctx, typeNamespacedName, resource)
			Expect(err).To(MatchError(k8serrors.IsNotFound, "k8serrors.IsNotFound"))
			// no garbage collection with envtest, so we need to delete resources manually
			deleteOwnedResources(ctx, k8sClient, resource)
			// Wait for all CRDs to be deleted which is apparently slow on envtest
			Eventually(func(g Gomega) []apiextensionsv1.CustomResourceDefinition {
				crdList := &apiextensionsv1.CustomResourceDefinitionList{}
				err := k8sClient.List(ctx, crdList)
				g.Expect(err).ToNot(HaveOccurred())
				return crdList.Items
			}, time.Second*15, time.Second*1).Should(HaveLen(1), "All CRDs other than the one for MigController should be deleted")
		})

		It("should successfully reconcile the resource", func() {
			resource := &migrationsv1alpha1.MigController{}
			err := k8sClient.Get(ctx, typeNamespacedName, resource)
			Expect(err).NotTo(HaveOccurred())

			Expect(resource.Status.ObservedVersion).To(Equal("0.0.1"))
			By("Reconciling the created resource")
			res, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(res).To(Equal(defaultResult))
			// TODO(user): Add more specific assertions depending on your controller's reconciliation logic.
			// Example: If you expect a certain status condition after reconciliation, verify it here.
		})

		It("should successfully create aggregated cluster role", func() {
			resource := &migrationsv1alpha1.MigController{}
			err := k8sClient.Get(ctx, typeNamespacedName, resource)
			Expect(err).NotTo(HaveOccurred())

			Expect(resource.Status.ObservedVersion).To(Equal("0.0.1"))
			By("Reconciling the created resource")
			res, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(res).To(Equal(defaultResult))

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

// deleteOwnedResources deletes all resources owned by the given owner resource
func deleteOwnedResources(ctx context.Context, k8sClient client.Client, owner client.Object) {
	// Get the owner's UID for comparison
	ownerUID := owner.GetUID()
	ownerNamespace := owner.GetNamespace()

	// List of resource types that might be owned by the CR
	resourceTypes := []client.ObjectList{
		&corev1.ConfigMapList{},
		&corev1.SecretList{},
		&corev1.ServiceList{},
		&corev1.ServiceAccountList{},
		&rbacv1.RoleList{},
		&rbacv1.RoleBindingList{},
		&appsv1.DeploymentList{},
		// Add more resource types as needed
	}

	for _, resourceList := range resourceTypes {
		// List all resources of this type in the namespace
		Expect(k8sClient.List(ctx, resourceList, client.InNamespace(ownerNamespace))).To(Succeed())

		// Use reflection to access the Items field
		items := reflect.ValueOf(resourceList).Elem().FieldByName("Items")
		if !items.IsValid() {
			continue
		}

		// Iterate through items and delete those owned by our CR
		for i := 0; i < items.Len(); i++ {
			item := items.Index(i).Addr().Interface().(client.Object)

			// Check if this resource is owned by our CR
			for _, ownerRef := range item.GetOwnerReferences() {
				if ownerRef.UID == ownerUID {
					By(fmt.Sprintf("Deleting owned %T: %s", item, item.GetName()))
					Expect(k8sClient.Delete(ctx, item)).To(Succeed())
					break
				}
			}
		}
	}
}
